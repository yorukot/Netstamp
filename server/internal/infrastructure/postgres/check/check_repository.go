package pgcheck

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres"
	pglabel "github.com/yorukot/netstamp/internal/infrastructure/postgres/label"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres/sqlc"
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

func (r *CheckRepository) GetCheck(ctx context.Context, projectIDValue string, checkIDValue string) (domaincheck.Check, error) {
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

func (r *CheckRepository) CreateCheck(ctx context.Context, input domaincheck.CreateCheckStorageInput) (domaincheck.Check, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgcheckTracer, "checks", "postgres.checks.create", "INSERT", "INSERT check, ping config, and labels")
	defer span.End()

	projectID, err := postgres.ParseUUID(input.ProjectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return domaincheck.Check{}, err
	}
	labelIDs, err := pglabel.ParseLabelIDs(input.LabelIDs)
	if err != nil {
		return domaincheck.Check{}, err
	}

	var created domaincheck.Check
	err = r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		row, err := q.CreateCheck(ctx, sqlc.CreateCheckParams{
			ProjectID:       projectID,
			Name:            input.Name,
			CheckType:       sqlcCheckType(input.Type),
			Target:          input.Target,
			Selector:        input.Selector,
			Description:     input.Description,
			IntervalSeconds: int32(input.IntervalSeconds),
		})
		if err != nil {
			return mapCheckWriteError(err)
		}

		config, err := q.CreatePingCheckConfig(ctx, sqlc.CreatePingCheckConfigParams{
			CheckID:         row.ID,
			PacketCount:     int32(input.PingConfig.PacketCount),
			PacketSizeBytes: int32(input.PingConfig.PacketSizeBytes),
			TimeoutMs:       int32(input.PingConfig.TimeoutMs),
			IpFamily:        sqlcIPFamily(input.PingConfig.IPFamily),
		})
		if err != nil {
			return mapCheckWriteError(err)
		}

		for _, labelID := range labelIDs {
			if err := q.CreateCheckLabel(ctx, sqlc.CreateCheckLabelParams{
				ProjectID: projectID,
				CheckID:   row.ID,
				LabelID:   labelID,
			}); err != nil {
				return mapCheckWriteError(err)
			}
		}

		created = mapStoredCheck(row, config)
		return nil
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domaincheck.Check{}, err
	}

	return created, nil
}

func (r *CheckRepository) UpdateCheck(ctx context.Context, input domaincheck.UpdateCheckStorageInput) (domaincheck.Check, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgcheckTracer, "checks", "postgres.checks.update", "UPDATE", "UPDATE check, ping config, and labels")
	defer span.End()

	projectID, checkID, err := parseProjectAndCheckIDs(input.ProjectID, input.CheckID)
	if err != nil {
		return domaincheck.Check{}, err
	}
	labelIDs, err := pglabel.ParseLabelIDs(input.LabelIDs)
	if err != nil {
		return domaincheck.Check{}, err
	}

	var updated domaincheck.Check
	err = r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		row, err := q.UpdateCheck(ctx, sqlc.UpdateCheckParams{
			ProjectID:       projectID,
			ID:              checkID,
			Name:            input.Name,
			CheckType:       sqlcCheckType(input.Type),
			Target:          input.Target,
			Selector:        input.Selector,
			Description:     input.Description,
			IntervalSeconds: int32(input.IntervalSeconds),
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domaincheck.ErrCheckNotFound
			}
			return mapCheckWriteError(err)
		}

		config, err := q.UpdatePingCheckConfig(ctx, sqlc.UpdatePingCheckConfigParams{
			CheckID:         checkID,
			PacketCount:     int32(input.PingConfig.PacketCount),
			PacketSizeBytes: int32(input.PingConfig.PacketSizeBytes),
			TimeoutMs:       int32(input.PingConfig.TimeoutMs),
			IpFamily:        sqlcIPFamily(input.PingConfig.IPFamily),
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domaincheck.ErrCheckNotFound
			}
			return mapCheckWriteError(err)
		}

		if input.ReplaceLabels {
			if err := q.DeleteCheckLabels(ctx, sqlc.DeleteCheckLabelsParams{
				ProjectID: projectID,
				CheckID:   checkID,
			}); err != nil {
				return err
			}
			for _, labelID := range labelIDs {
				if err := q.CreateCheckLabel(ctx, sqlc.CreateCheckLabelParams{
					ProjectID: projectID,
					CheckID:   checkID,
					LabelID:   labelID,
				}); err != nil {
					return mapCheckWriteError(err)
				}
			}
		}

		updated = mapStoredCheck(row, config)
		updated.Labels, err = r.listLabelsForCheck(ctx, q, projectID, checkID)
		return err
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domaincheck.Check{}, err
	}

	return updated, nil
}

func (r *CheckRepository) SoftDeleteCheck(ctx context.Context, projectIDValue string, checkIDValue string) error {
	ctx, span := postgres.StartDBSpan(ctx, pgcheckTracer, "checks", "postgres.checks.soft_delete", "UPDATE", "SOFT DELETE check")
	defer span.End()

	projectID, checkID, err := parseProjectAndCheckIDs(projectIDValue, checkIDValue)
	if err != nil {
		return err
	}

	if _, err := r.queries.SoftDeleteCheck(ctx, sqlc.SoftDeleteCheckParams{
		ProjectID: projectID,
		ID:        checkID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domaincheck.ErrCheckNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *CheckRepository) listLabelsForCheck(ctx context.Context, queries *sqlc.Queries, projectID uuid.UUID, checkID uuid.UUID) ([]domainlabel.Label, error) {
	rows, err := queries.ListActiveLabelsForCheck(ctx, sqlc.ListActiveLabelsForCheckParams{
		ProjectID: projectID,
		CheckID:   checkID,
	})
	if err != nil {
		return nil, err
	}

	return mapLabels(rows), nil
}

func parseProjectAndCheckIDs(projectIDValue string, checkIDValue string) (uuid.UUID, uuid.UUID, error) {
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
