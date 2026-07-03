package pgassignment

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainselector "github.com/yorukot/netstamp/internal/domain/selector"
)

type AssignmentRepository struct {
	queries *sqlc.Queries
	tx      *postgres.Transactor
}

func NewAssignmentRepository(pool *pgxpool.Pool) *AssignmentRepository {
	return &AssignmentRepository{
		queries: sqlc.New(pool),
		tx:      postgres.NewTransactor(pool),
	}
}

func (r *AssignmentRepository) RefreshProbeCheckAssignmentsForProject(ctx context.Context, projectIDValue string) error {
	ctx, span := postgres.StartDBSpan(ctx, pgassignmentTracer, "probe_check_assignments", "postgres.assignments.refresh_for_project", "UPDATE", "REFRESH probe check assignments for project")
	defer span.End()

	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return err
	}

	err = r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		rows, queryErr := q.ListActiveProbesForProject(ctx, projectID)
		if queryErr != nil {
			return queryErr
		}
		for _, probe := range activeProbeLabelsFromProjectRows(rows) {
			if refreshErr := r.refreshProbeCheckAssignmentsForProbe(ctx, q, projectID, probe); refreshErr != nil {
				return refreshErr
			}
		}

		return nil
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *AssignmentRepository) RefreshProbeCheckAssignmentsForProbe(ctx context.Context, projectIDValue, probeIDValue string) error {
	ctx, span := postgres.StartDBSpan(ctx, pgassignmentTracer, "probe_check_assignments", "postgres.assignments.refresh_for_probe", "UPDATE", "REFRESH probe check assignments for probe")
	defer span.End()

	projectID, probeID, err := parseProjectAndProbeIDs(projectIDValue, probeIDValue)
	if err != nil {
		return err
	}

	err = r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		rows, queryErr := q.GetActiveProbeRowsForProject(ctx, sqlc.GetActiveProbeRowsForProjectParams{
			ProjectID: projectID,
			ID:        probeID,
		})
		if queryErr != nil {
			return queryErr
		}
		probe, ok := activeProbeFromRows(rows)
		if !ok {
			return domainprobe.ErrProbeNotFound
		}

		return r.refreshProbeCheckAssignmentsForProbe(ctx, q, projectID, probe)
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *AssignmentRepository) RefreshProbeCheckAssignmentsForCheck(ctx context.Context, projectIDValue, checkIDValue string) error {
	ctx, span := postgres.StartDBSpan(ctx, pgassignmentTracer, "probe_check_assignments", "postgres.assignments.refresh_for_check", "UPDATE", "REFRESH probe check assignments for check")
	defer span.End()

	projectID, checkID, err := parseProjectAndCheckIDs(projectIDValue, checkIDValue)
	if err != nil {
		return err
	}

	err = r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		row, queryErr := q.GetActiveCheckForProject(ctx, sqlc.GetActiveCheckForProjectParams{
			ProjectID: projectID,
			ID:        checkID,
		})
		if errors.Is(queryErr, pgx.ErrNoRows) {
			return domaincheck.ErrCheckNotFound
		}
		if queryErr != nil {
			return queryErr
		}

		return r.refreshProbeCheckAssignmentsForCheck(ctx, q, row)
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *AssignmentRepository) RefreshProbeCheckAssignmentsForLabel(ctx context.Context, projectIDValue, labelIDValue string) error {
	ctx, span := postgres.StartDBSpan(ctx, pgassignmentTracer, "probe_check_assignments", "postgres.assignments.refresh_for_label", "UPDATE", "REFRESH probe check assignments for label")
	defer span.End()

	projectID, labelID, err := parseProjectAndLabelIDs(projectIDValue, labelIDValue)
	if err != nil {
		return err
	}

	err = r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		targets, queryErr := q.ListProbeRefreshTargetsForLabel(ctx, sqlc.ListProbeRefreshTargetsForLabelParams{
			ProjectID: projectID,
			LabelID:   labelID,
		})
		if queryErr != nil {
			return queryErr
		}

		for _, target := range targets {
			probe := activeProbeLabels{
				probeID: target.ID,
				enabled: target.Enabled,
			}
			if refreshErr := r.refreshProbeCheckAssignmentsForProbe(ctx, q, projectID, probe); refreshErr != nil {
				return refreshErr
			}
		}

		return nil
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *AssignmentRepository) DeleteProbeCheckAssignmentsForProbe(ctx context.Context, projectIDValue, probeIDValue string) error {
	ctx, span := postgres.StartDBSpan(ctx, pgassignmentTracer, "probe_check_assignments", "postgres.assignments.delete_for_probe", "DELETE", "DELETE probe check assignments for probe")
	defer span.End()

	projectID, probeID, err := parseProjectAndProbeIDs(projectIDValue, probeIDValue)
	if err != nil {
		return err
	}

	if err := r.queries.DeleteProbeCheckAssignmentsForProbe(ctx, sqlc.DeleteProbeCheckAssignmentsForProbeParams{
		ProjectID: projectID,
		ProbeID:   probeID,
	}); err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *AssignmentRepository) DeleteProbeCheckAssignmentsForCheck(ctx context.Context, projectIDValue, checkIDValue string) error {
	ctx, span := postgres.StartDBSpan(ctx, pgassignmentTracer, "probe_check_assignments", "postgres.assignments.delete_for_check", "DELETE", "DELETE probe check assignments for check")
	defer span.End()

	projectID, checkID, err := parseProjectAndCheckIDs(projectIDValue, checkIDValue)
	if err != nil {
		return err
	}

	if err := r.queries.DeleteProbeCheckAssignmentsForCheck(ctx, sqlc.DeleteProbeCheckAssignmentsForCheckParams{
		ProjectID: projectID,
		CheckID:   checkID,
	}); err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *AssignmentRepository) ListSelectorPreviewProbes(ctx context.Context, projectIDValue string, selector domainselector.Selector) ([]domainprobe.Probe, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgassignmentTracer, "probes", "postgres.assignments.selector_preview", "SELECT", "SELECT probes matching selector")
	defer span.End()

	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListActiveEnabledProbeLabelsForProject(ctx, projectID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	return matchingPreviewProbes(selector, activeProbeLabelsFromRows(rows)), nil
}

func (r *AssignmentRepository) ListProjectAssignments(ctx context.Context, input domainassignment.Query) ([]domainassignment.Assignment, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgassignmentTracer, "probe_check_assignments", "postgres.assignments.list", "SELECT", "SELECT project assignments")
	defer span.End()

	projectID, err := postgres.ParseUUID(input.ProjectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return nil, err
	}
	probeID, err := optionalUUID(input.ProbeID, domainprobe.ErrInvalidInput)
	if err != nil {
		return nil, err
	}
	checkID, err := optionalUUID(input.CheckID, domaincheck.ErrInvalidInput)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListProjectAssignments(ctx, sqlc.ListProjectAssignmentsParams{
		ProjectID: projectID,
		ProbeID:   probeID,
		CheckID:   checkID,
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	return mapProjectAssignments(rows), nil
}

func (r *AssignmentRepository) refreshProbeCheckAssignmentsForProbe(ctx context.Context, queries *sqlc.Queries, projectID uuid.UUID, probe activeProbeLabels) error {
	if !probe.enabled {
		return queries.DeleteProbeCheckAssignmentsForProbe(ctx, sqlc.DeleteProbeCheckAssignmentsForProbeParams{
			ProjectID: projectID,
			ProbeID:   probe.probeID,
		})
	}

	if probe.labels == nil {
		labelRows, err := queries.ListActiveLabelsForProbe(ctx, sqlc.ListActiveLabelsForProbeParams{
			ProjectID: projectID,
			ProbeID:   probe.probeID,
		})
		if err != nil {
			return err
		}
		probe.labels = mapLabels(labelRows)
	}

	checkRows, err := queries.ListActiveChecksForProject(ctx, projectID)
	if err != nil {
		return err
	}

	matchedCheckIDs := make([]uuid.UUID, 0, len(checkRows))
	for _, row := range checkRows {
		selector, selectorRaw, err := listCheckSelector(row)
		if err != nil {
			return err
		}
		if !selector.Matches(probe.labels) {
			continue
		}
		matchedCheckIDs = append(matchedCheckIDs, row.ID)
		if err := queries.UpsertProbeCheckAssignment(ctx, sqlc.UpsertProbeCheckAssignmentParams{
			ProjectID:       projectID,
			ProbeID:         probe.probeID,
			CheckID:         row.ID,
			CheckVersion:    listCheckVersion(row),
			SelectorVersion: domaincheck.SelectorVersion(selectorRaw),
		}); err != nil {
			return mapAssignmentWriteError(err)
		}
	}

	return queries.DeleteStaleProbeCheckAssignmentsForProbe(ctx, sqlc.DeleteStaleProbeCheckAssignmentsForProbeParams{
		ProjectID: projectID,
		ProbeID:   probe.probeID,
		CheckIds:  matchedCheckIDs,
	})
}

func (r *AssignmentRepository) refreshProbeCheckAssignmentsForCheck(ctx context.Context, queries *sqlc.Queries, check sqlc.GetActiveCheckForProjectRow) error {
	selector, selectorRaw, err := checkSelector(check)
	if err != nil {
		return err
	}

	probeRows, err := queries.ListActiveEnabledProbeLabelsForProject(ctx, check.ProjectID)
	if err != nil {
		return err
	}

	matchedProbeIDs := matchingProbeIDs(selector, activeProbeLabelsFromRows(probeRows))
	for _, probeID := range matchedProbeIDs {
		if err := queries.UpsertProbeCheckAssignment(ctx, sqlc.UpsertProbeCheckAssignmentParams{
			ProjectID:       check.ProjectID,
			ProbeID:         probeID,
			CheckID:         check.ID,
			CheckVersion:    checkVersion(check),
			SelectorVersion: domaincheck.SelectorVersion(selectorRaw),
		}); err != nil {
			return mapAssignmentWriteError(err)
		}
	}

	return queries.DeleteStaleProbeCheckAssignments(ctx, sqlc.DeleteStaleProbeCheckAssignmentsParams{
		ProjectID:       check.ProjectID,
		CheckID:         check.ID,
		CheckVersion:    checkVersion(check),
		SelectorVersion: domaincheck.SelectorVersion(selectorRaw),
		ProbeIds:        matchedProbeIDs,
	})
}

func parseProjectAndProbeIDs(projectIDValue, probeIDValue string) (uuid.UUID, uuid.UUID, error) {
	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	probeID, err := postgres.ParseUUID(probeIDValue, domainprobe.ErrProbeNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	return projectID, probeID, nil
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

func parseProjectAndLabelIDs(projectIDValue, labelIDValue string) (uuid.UUID, uuid.UUID, error) {
	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	labelID, err := postgres.ParseUUID(labelIDValue, domainlabel.ErrLabelNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	return projectID, labelID, nil
}

func optionalUUID(value string, invalidErr error) (*uuid.UUID, error) {
	if value == "" {
		return nil, nil //nolint:nilnil // Nil means no optional UUID filter.
	}
	id, err := postgres.ParseUUID(value, invalidErr)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
