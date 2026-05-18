package pgcheck

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	pglabel "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/label"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainselector "github.com/yorukot/netstamp/internal/domain/selector"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type CheckRepository struct {
	queries *sqlc.Queries
	tx      *postgres.Transactor
}

func NewCheckRepository(pool *pgxpool.Pool) *CheckRepository {
	return &CheckRepository{
		queries: sqlc.New(pool),
		tx:      postgres.NewTransactor(pool),
	}
}

func (r *CheckRepository) ListChecks(ctx context.Context, projectIDValue string) ([]domaincheck.Check, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgcheckTracer, "checks", "postgres.checks.list", "SELECT", "SELECT active checks for project")
	defer span.End()

	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListActiveChecksForProject(ctx, projectID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	checks := make([]domaincheck.Check, 0, len(rows))
	for _, row := range rows {
		check := mapListCheck(row)
		check.Labels, err = r.listLabelsForCheck(ctx, r.queries, row.ProjectID, row.ID)
		if err != nil {
			postgres.RecordDBSpanError(span, err)
			return nil, err
		}
		checks = append(checks, check)
	}

	return checks, nil
}

func (r *CheckRepository) GetCheck(ctx context.Context, projectIDValue, checkIDValue string) (domaincheck.Check, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgcheckTracer, "checks", "postgres.checks.select", "SELECT", "SELECT active check for project")
	defer span.End()

	projectID, checkID, err := parseProjectAndCheckIDs(projectIDValue, checkIDValue)
	if err != nil {
		return domaincheck.Check{}, err
	}

	row, err := r.queries.GetActiveCheckForProject(ctx, sqlc.GetActiveCheckForProjectParams{
		ProjectID: projectID,
		ID:        checkID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domaincheck.Check{}, domaincheck.ErrCheckNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return domaincheck.Check{}, err
	}

	check := mapGetCheck(row)
	check.Labels, err = r.listLabelsForCheck(ctx, r.queries, projectID, checkID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domaincheck.Check{}, err
	}

	return check, nil
}

func (r *CheckRepository) CreateCheck(ctx context.Context, input domaincheck.Check, labelIDValues []string) (domaincheck.Check, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgcheckTracer, "checks", "postgres.checks.create", "INSERT", "INSERT check, ping config, and labels")
	defer span.End()

	projectID, err := postgres.ParseUUID(input.ProjectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return domaincheck.Check{}, err
	}
	labelIDs, err := pglabel.ParseLabelIDs(labelIDValues)
	if err != nil {
		return domaincheck.Check{}, err
	}
	if _, parseErr := domainselector.Parse(input.Selector); parseErr != nil {
		return domaincheck.Check{}, domaincheck.ErrInvalidInput
	}
	var created domaincheck.Check
	err = r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		row, createErr := q.CreateCheck(ctx, sqlc.CreateCheckParams{
			ProjectID:       projectID,
			Name:            input.Name,
			CheckType:       sqlcCheckType(input.Type),
			Target:          input.Target,
			Selector:        input.Selector,
			Description:     input.Description,
			IntervalSeconds: input.IntervalSeconds,
		})
		if createErr != nil {
			return mapCheckWriteError(createErr)
		}

		checkWithConfig, configErr := r.createCheckConfig(ctx, q, row, input)
		if configErr != nil {
			return mapCheckWriteError(configErr)
		}

		for _, labelID := range labelIDs {
			if labelErr := q.CreateCheckLabel(ctx, sqlc.CreateCheckLabelParams{
				ProjectID: projectID,
				CheckID:   row.ID,
				LabelID:   labelID,
			}); labelErr != nil {
				return mapCheckWriteError(labelErr)
			}
		}

		created = checkWithConfig
		return nil
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domaincheck.Check{}, err
	}

	return created, nil
}

func (r *CheckRepository) UpdateCheck(ctx context.Context, input domaincheck.Check, replaceLabels bool, labelIDValues []string) (domaincheck.Check, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgcheckTracer, "checks", "postgres.checks.update", "UPDATE", "UPDATE check, ping config, and labels")
	defer span.End()

	projectID, checkID, err := parseProjectAndCheckIDs(input.ProjectID, input.ID)
	if err != nil {
		return domaincheck.Check{}, err
	}
	labelIDs, err := pglabel.ParseLabelIDs(labelIDValues)
	if err != nil {
		return domaincheck.Check{}, err
	}
	if _, parseErr := domainselector.Parse(input.Selector); parseErr != nil {
		return domaincheck.Check{}, domaincheck.ErrInvalidInput
	}
	var updated domaincheck.Check
	err = r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		row, updateErr := q.UpdateCheck(ctx, sqlc.UpdateCheckParams{
			ProjectID:       projectID,
			ID:              checkID,
			Name:            input.Name,
			CheckType:       sqlcCheckType(input.Type),
			Target:          input.Target,
			Selector:        input.Selector,
			Description:     input.Description,
			IntervalSeconds: input.IntervalSeconds,
		})
		if updateErr != nil {
			if errors.Is(updateErr, pgx.ErrNoRows) {
				return domaincheck.ErrCheckNotFound
			}
			return mapCheckWriteError(updateErr)
		}

		checkWithConfig, configErr := r.updateCheckConfig(ctx, q, row, input)
		if configErr != nil {
			if errors.Is(configErr, pgx.ErrNoRows) {
				return domaincheck.ErrCheckNotFound
			}
			return mapCheckWriteError(configErr)
		}

		if replaceLabels {
			if deleteErr := q.DeleteCheckLabels(ctx, sqlc.DeleteCheckLabelsParams{
				ProjectID: projectID,
				CheckID:   checkID,
			}); deleteErr != nil {
				return deleteErr
			}
			for _, labelID := range labelIDs {
				if labelErr := q.CreateCheckLabel(ctx, sqlc.CreateCheckLabelParams{
					ProjectID: projectID,
					CheckID:   checkID,
					LabelID:   labelID,
				}); labelErr != nil {
					return mapCheckWriteError(labelErr)
				}
			}
		}

		updated = checkWithConfig
		updated.Labels, err = r.listLabelsForCheck(ctx, q, projectID, checkID)
		return err
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domaincheck.Check{}, err
	}

	return updated, nil
}

func (r *CheckRepository) SoftDeleteCheck(ctx context.Context, projectIDValue, checkIDValue string) error {
	ctx, span := postgres.StartDBSpan(ctx, pgcheckTracer, "checks", "postgres.checks.soft_delete", "UPDATE", "SOFT DELETE check")
	defer span.End()

	projectID, checkID, err := parseProjectAndCheckIDs(projectIDValue, checkIDValue)
	if err != nil {
		return err
	}

	_, err = r.queries.SoftDeleteCheck(ctx, sqlc.SoftDeleteCheckParams{
		ProjectID: projectID,
		ID:        checkID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domaincheck.ErrCheckNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *CheckRepository) createCheckConfig(ctx context.Context, q *sqlc.Queries, row sqlc.Check, input domaincheck.Check) (domaincheck.Check, error) {
	return r.writeCheckConfig(ctx, q, row, input, checkConfigWriteCreate)
}

func (r *CheckRepository) updateCheckConfig(ctx context.Context, q *sqlc.Queries, row sqlc.Check, input domaincheck.Check) (domaincheck.Check, error) {
	return r.writeCheckConfig(ctx, q, row, input, checkConfigWriteUpdate)
}

type checkConfigWriteMode string

const (
	checkConfigWriteCreate checkConfigWriteMode = "create"
	checkConfigWriteUpdate checkConfigWriteMode = "update"
)

func (r *CheckRepository) writeCheckConfig(ctx context.Context, q *sqlc.Queries, row sqlc.Check, input domaincheck.Check, mode checkConfigWriteMode) (domaincheck.Check, error) {
	switch input.Type {
	case domaincheck.TypePing:
		return r.writePingCheckConfig(ctx, q, row, input.PingConfig, mode)
	case domaincheck.TypeTraceroute:
		return r.writeTracerouteCheckConfig(ctx, q, row, input.TracerouteConfig, mode)
	default:
		return domaincheck.Check{}, domaincheck.ErrInvalidInput
	}
}

func (r *CheckRepository) writePingCheckConfig(ctx context.Context, q *sqlc.Queries, row sqlc.Check, config *domainping.Config, mode checkConfigWriteMode) (domaincheck.Check, error) {
	if config == nil {
		defaultConfig := domainping.DefaultConfig()
		config = &defaultConfig
	}

	switch mode {
	case checkConfigWriteCreate:
		stored, err := q.CreatePingCheckConfig(ctx, sqlc.CreatePingCheckConfigParams{
			CheckID:         row.ID,
			PacketCount:     config.PacketCount,
			PacketSizeBytes: config.PacketSizeBytes,
			TimeoutMs:       config.TimeoutMs,
			IpFamily:        sqlcIPFamily(config.IPFamily),
		})
		if err != nil {
			return domaincheck.Check{}, err
		}
		return mapStoredPingCheck(row, stored), nil
	case checkConfigWriteUpdate:
		stored, err := q.UpdatePingCheckConfig(ctx, sqlc.UpdatePingCheckConfigParams{
			CheckID:         row.ID,
			PacketCount:     config.PacketCount,
			PacketSizeBytes: config.PacketSizeBytes,
			TimeoutMs:       config.TimeoutMs,
			IpFamily:        sqlcIPFamily(config.IPFamily),
		})
		if err != nil {
			return domaincheck.Check{}, err
		}
		return mapStoredPingCheck(row, stored), nil
	default:
		return domaincheck.Check{}, domaincheck.ErrInvalidInput
	}
}

func (r *CheckRepository) writeTracerouteCheckConfig(ctx context.Context, q *sqlc.Queries, row sqlc.Check, config *domaintraceroute.Config, mode checkConfigWriteMode) (domaincheck.Check, error) {
	if config == nil {
		defaultConfig := domaintraceroute.DefaultConfig()
		config = &defaultConfig
	}

	switch mode {
	case checkConfigWriteCreate:
		stored, err := q.CreateTracerouteCheckConfig(ctx, sqlc.CreateTracerouteCheckConfigParams{
			CheckID:         row.ID,
			Protocol:        sqlcTracerouteProtocol(config.Protocol),
			MaxHops:         config.MaxHops,
			TimeoutMs:       config.TimeoutMs,
			QueriesPerHop:   config.QueriesPerHop,
			PacketSizeBytes: config.PacketSizeBytes,
			Port:            config.Port,
			IpFamily:        sqlcIPFamily(config.IPFamily),
		})
		if err != nil {
			return domaincheck.Check{}, err
		}
		return mapStoredTracerouteCheck(row, stored), nil
	case checkConfigWriteUpdate:
		stored, err := q.UpdateTracerouteCheckConfig(ctx, sqlc.UpdateTracerouteCheckConfigParams{
			CheckID:         row.ID,
			Protocol:        sqlcTracerouteProtocol(config.Protocol),
			MaxHops:         config.MaxHops,
			TimeoutMs:       config.TimeoutMs,
			QueriesPerHop:   config.QueriesPerHop,
			PacketSizeBytes: config.PacketSizeBytes,
			Port:            config.Port,
			IpFamily:        sqlcIPFamily(config.IPFamily),
		})
		if err != nil {
			return domaincheck.Check{}, err
		}
		return mapStoredTracerouteCheck(row, stored), nil
	default:
		return domaincheck.Check{}, domaincheck.ErrInvalidInput
	}
}

func (r *CheckRepository) listLabelsForCheck(ctx context.Context, queries *sqlc.Queries, projectID, checkID uuid.UUID) ([]domainlabel.Label, error) {
	rows, err := queries.ListActiveLabelsForCheck(ctx, sqlc.ListActiveLabelsForCheckParams{
		ProjectID: projectID,
		CheckID:   checkID,
	})
	if err != nil {
		return nil, err
	}

	return mapLabels(rows), nil
}

func parseProjectAndCheckIDs(projectIDValue, checkIDValue string) (uuid.UUID, uuid.UUID, error) {
	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	checkID, err := postgres.ParseUUID(checkIDValue, domaincheck.ErrCheckNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	return projectID, checkID, nil
}
