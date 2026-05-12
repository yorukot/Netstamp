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
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainselector "github.com/yorukot/netstamp/internal/domain/selector"
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

func (r *CheckRepository) CreateCheck(ctx context.Context, input domaincheck.CreateCheckStorageInput) (domaincheck.Check, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgcheckTracer, "checks", "postgres.checks.create", "INSERT", "INSERT check, ping config, labels, and assignment links")
	defer span.End()

	projectID, err := postgres.ParseUUID(input.ProjectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return domaincheck.Check{}, err
	}
	labelIDs, err := pglabel.ParseLabelIDs(input.LabelIDs)
	if err != nil {
		return domaincheck.Check{}, err
	}
	parsedSelector, err := domainselector.Parse(input.Selector)
	if err != nil {
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

		config, configErr := q.CreatePingCheckConfig(ctx, sqlc.CreatePingCheckConfigParams{
			CheckID:         row.ID,
			PacketCount:     input.PingConfig.PacketCount,
			PacketSizeBytes: input.PingConfig.PacketSizeBytes,
			TimeoutMs:       input.PingConfig.TimeoutMs,
			IpFamily:        sqlcIPFamily(input.PingConfig.IPFamily),
		})
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

		// probe_check_assignments is a cache of the current probe/check links, so
		// refresh it after the check row, config, and labels have been persisted.
		probes, listProbeErr := r.listActiveEnabledProbeLabels(ctx, q, projectID)
		if listProbeErr != nil {
			return listProbeErr
		}
		for _, probeID := range matchingProbeIDs(parsedSelector, probes) {
			if linkErr := q.CreateProbeCheckAssignment(ctx, sqlc.CreateProbeCheckAssignmentParams{
				ProjectID:       projectID,
				ProbeID:         probeID,
				CheckID:         row.ID,
				CheckVersion:    input.CheckVersion,
				SelectorVersion: input.SelectorVersion,
			}); linkErr != nil {
				return mapCheckWriteError(linkErr)
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
	ctx, span := postgres.StartDBSpan(ctx, pgcheckTracer, "checks", "postgres.checks.update", "UPDATE", "UPDATE check, ping config, labels, and assignment links")
	defer span.End()

	projectID, checkID, err := parseProjectAndCheckIDs(input.ProjectID, input.CheckID)
	if err != nil {
		return domaincheck.Check{}, err
	}
	labelIDs, err := pglabel.ParseLabelIDs(input.LabelIDs)
	if err != nil {
		return domaincheck.Check{}, err
	}
	parsedSelector, err := domainselector.Parse(input.Selector)
	if err != nil {
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

		config, configErr := q.UpdatePingCheckConfig(ctx, sqlc.UpdatePingCheckConfigParams{
			CheckID:         checkID,
			PacketCount:     input.PingConfig.PacketCount,
			PacketSizeBytes: input.PingConfig.PacketSizeBytes,
			TimeoutMs:       input.PingConfig.TimeoutMs,
			IpFamily:        sqlcIPFamily(input.PingConfig.IPFamily),
		})
		if configErr != nil {
			if errors.Is(configErr, pgx.ErrNoRows) {
				return domaincheck.ErrCheckNotFound
			}
			return mapCheckWriteError(configErr)
		}

		if input.ReplaceLabels {
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

		if linkErr := r.refreshProbeCheckAssignments(ctx, q, projectID, checkID, parsedSelector, input); linkErr != nil {
			return linkErr
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

func (r *CheckRepository) refreshProbeCheckAssignments(
	ctx context.Context,
	queries *sqlc.Queries,
	projectID uuid.UUID,
	checkID uuid.UUID,
	selector domainselector.Selector,
	input domaincheck.UpdateCheckStorageInput,
) error {
	probes, err := r.listActiveEnabledProbeLabels(ctx, queries, projectID)
	if err != nil {
		return err
	}

	matchedProbeIDs := matchingProbeIDs(selector, probes)
	for _, probeID := range matchedProbeIDs {
		if linkErr := queries.UpsertProbeCheckAssignment(ctx, sqlc.UpsertProbeCheckAssignmentParams{
			ProjectID:       projectID,
			ProbeID:         probeID,
			CheckID:         checkID,
			CheckVersion:    input.CheckVersion,
			SelectorVersion: input.SelectorVersion,
		}); linkErr != nil {
			return mapCheckWriteError(linkErr)
		}
	}

	return queries.DeleteStaleProbeCheckAssignments(ctx, sqlc.DeleteStaleProbeCheckAssignmentsParams{
		ProjectID:       projectID,
		CheckID:         checkID,
		CheckVersion:    input.CheckVersion,
		SelectorVersion: input.SelectorVersion,
		ProbeIds:        matchedProbeIDs,
	})
}

func (r *CheckRepository) SoftDeleteCheck(ctx context.Context, projectIDValue, checkIDValue string) error {
	ctx, span := postgres.StartDBSpan(ctx, pgcheckTracer, "checks", "postgres.checks.soft_delete", "UPDATE", "SOFT DELETE check and assignment links")
	defer span.End()

	projectID, checkID, err := parseProjectAndCheckIDs(projectIDValue, checkIDValue)
	if err != nil {
		return err
	}

	err = r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		if _, deleteErr := q.SoftDeleteCheck(ctx, sqlc.SoftDeleteCheckParams{
			ProjectID: projectID,
			ID:        checkID,
		}); deleteErr != nil {
			if errors.Is(deleteErr, pgx.ErrNoRows) {
				return domaincheck.ErrCheckNotFound
			}
			return deleteErr
		}

		// Once the check is deleted, no probe should keep a cached link to it.
		return q.DeleteProbeCheckAssignmentsForCheck(ctx, sqlc.DeleteProbeCheckAssignmentsForCheckParams{
			ProjectID: projectID,
			CheckID:   checkID,
		})
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
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

type activeProbeLabels struct {
	probeID uuid.UUID
	labels  []domainlabel.Label
}

func (r *CheckRepository) listActiveEnabledProbeLabels(ctx context.Context, queries *sqlc.Queries, projectID uuid.UUID) ([]activeProbeLabels, error) {
	rows, err := queries.ListActiveEnabledProbeLabelsForProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	probeIndex := make(map[uuid.UUID]int)
	probes := make([]activeProbeLabels, 0)
	for _, row := range rows {
		index, ok := probeIndex[row.ProbeID]
		if !ok {
			index = len(probes)
			probeIndex[row.ProbeID] = index
			probes = append(probes, activeProbeLabels{probeID: row.ProbeID})
		}
		if label, ok := mapProbeLabel(row); ok {
			probes[index].labels = append(probes[index].labels, label)
		}
	}

	return probes, nil
}

func matchingProbeIDs(selector domainselector.Selector, probes []activeProbeLabels) []uuid.UUID {
	probeIDs := make([]uuid.UUID, 0, len(probes))
	for _, probe := range probes {
		if selector.Matches(probe.labels) {
			probeIDs = append(probeIDs, probe.probeID)
		}
	}

	return probeIDs
}

func mapProbeLabel(row sqlc.ListActiveEnabledProbeLabelsForProjectRow) (domainlabel.Label, bool) {
	if row.LabelID == nil || row.LabelProjectID == nil || row.LabelKey == nil || row.LabelValue == nil {
		return domainlabel.Label{}, false
	}

	return domainlabel.Label{
		ID:        row.LabelID.String(),
		ProjectID: row.LabelProjectID.String(),
		Key:       *row.LabelKey,
		Value:     *row.LabelValue,
		CreatedAt: row.LabelCreatedAt.Time,
		UpdatedAt: row.LabelUpdatedAt.Time,
		DeletedAt: timePtr(row.LabelDeletedAt),
	}, true
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
