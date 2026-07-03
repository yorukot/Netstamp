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
)

type AssignmentRepository struct {
	queries *sqlc.Queries
}

func NewAssignmentRepository(pool *pgxpool.Pool) *AssignmentRepository {
	return &AssignmentRepository{
		queries: sqlc.New(pool),
	}
}

func (r *AssignmentRepository) ListProbeRefreshCandidatesForProject(ctx context.Context, projectIDValue string) ([]domainassignment.ProbeAssignmentCandidate, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgassignmentTracer, "probes", "postgres.assignments.list_probe_refresh_candidates_for_project", "SELECT", "SELECT probe refresh candidates for project")
	defer span.End()

	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return nil, err
	}

	rows, err := postgres.Queries(ctx, r.queries).ListActiveProbesForProject(ctx, projectID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	return probeCandidatesFromActive(activeProbeLabelsFromProjectRows(rows)), nil
}

func (r *AssignmentRepository) GetProbeRefreshCandidate(ctx context.Context, projectIDValue, probeIDValue string) (domainassignment.ProbeAssignmentCandidate, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgassignmentTracer, "probes", "postgres.assignments.get_probe_refresh_candidate", "SELECT", "SELECT probe refresh candidate")
	defer span.End()

	projectID, probeID, err := parseProjectAndProbeIDs(projectIDValue, probeIDValue)
	if err != nil {
		return domainassignment.ProbeAssignmentCandidate{}, err
	}

	rows, err := postgres.Queries(ctx, r.queries).GetActiveProbeRowsForProject(ctx, sqlc.GetActiveProbeRowsForProjectParams{
		ProjectID: projectID,
		ID:        probeID,
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainassignment.ProbeAssignmentCandidate{}, err
	}
	probe, ok := activeProbeFromRows(rows)
	if !ok {
		return domainassignment.ProbeAssignmentCandidate{}, domainprobe.ErrProbeNotFound
	}

	return probeCandidateFromActive(probe), nil
}

func (r *AssignmentRepository) ListProbeRefreshCandidatesForLabel(ctx context.Context, projectIDValue, labelIDValue string) ([]domainassignment.ProbeAssignmentCandidate, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgassignmentTracer, "probes", "postgres.assignments.list_probe_refresh_candidates_for_label", "SELECT", "SELECT probe refresh candidates for label")
	defer span.End()

	projectID, labelID, err := parseProjectAndLabelIDs(projectIDValue, labelIDValue)
	if err != nil {
		return nil, err
	}

	q := postgres.Queries(ctx, r.queries)
	targets, err := q.ListProbeRefreshTargetsForLabel(ctx, sqlc.ListProbeRefreshTargetsForLabelParams{
		ProjectID: projectID,
		LabelID:   labelID,
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	probes := make([]domainassignment.ProbeAssignmentCandidate, 0, len(targets))
	for _, target := range targets {
		labelRows, err := q.ListActiveLabelsForProbe(ctx, sqlc.ListActiveLabelsForProbeParams{
			ProjectID: projectID,
			ProbeID:   target.ID,
		})
		if err != nil {
			postgres.RecordDBSpanError(span, err)
			return nil, err
		}
		probes = append(probes, domainassignment.ProbeAssignmentCandidate{
			ProbeID:   target.ID.String(),
			ProjectID: projectID.String(),
			Enabled:   target.Enabled,
			Labels:    mapLabels(labelRows),
		})
	}

	return probes, nil
}

func (r *AssignmentRepository) ListCheckRefreshCandidatesForProject(ctx context.Context, projectIDValue string) ([]domainassignment.CheckAssignmentCandidate, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgassignmentTracer, "checks", "postgres.assignments.list_check_refresh_candidates_for_project", "SELECT", "SELECT check refresh candidates for project")
	defer span.End()

	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return nil, err
	}

	rows, err := postgres.Queries(ctx, r.queries).ListActiveChecksForProject(ctx, projectID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	return listCheckCandidates(rows)
}

func (r *AssignmentRepository) GetCheckRefreshCandidate(ctx context.Context, projectIDValue, checkIDValue string) (domainassignment.CheckAssignmentCandidate, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgassignmentTracer, "checks", "postgres.assignments.get_check_refresh_candidate", "SELECT", "SELECT check refresh candidate")
	defer span.End()

	projectID, checkID, err := parseProjectAndCheckIDs(projectIDValue, checkIDValue)
	if err != nil {
		return domainassignment.CheckAssignmentCandidate{}, err
	}

	row, err := postgres.Queries(ctx, r.queries).GetActiveCheckForProject(ctx, sqlc.GetActiveCheckForProjectParams{
		ProjectID: projectID,
		ID:        checkID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domainassignment.CheckAssignmentCandidate{}, domaincheck.ErrCheckNotFound
	}
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainassignment.CheckAssignmentCandidate{}, err
	}

	return checkCandidate(row)
}

func (r *AssignmentRepository) ListSelectorPreviewCandidates(ctx context.Context, projectIDValue string) ([]domainassignment.ProbeAssignmentCandidate, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgassignmentTracer, "probes", "postgres.assignments.selector_preview_candidates", "SELECT", "SELECT selector preview candidates")
	defer span.End()

	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return nil, err
	}

	rows, err := postgres.Queries(ctx, r.queries).ListActiveEnabledProbeLabelsForProject(ctx, projectID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	return probeCandidatesFromActive(activeProbeLabelsFromRows(rows)), nil
}

func (r *AssignmentRepository) DeleteProbeCheckAssignmentsForProbe(ctx context.Context, projectIDValue, probeIDValue string) error {
	ctx, span := postgres.StartDBSpan(ctx, pgassignmentTracer, "probe_check_assignments", "postgres.assignments.delete_for_probe", "DELETE", "DELETE probe check assignments for probe")
	defer span.End()

	projectID, probeID, err := parseProjectAndProbeIDs(projectIDValue, probeIDValue)
	if err != nil {
		return err
	}

	if err := postgres.Queries(ctx, r.queries).DeleteProbeCheckAssignmentsForProbe(ctx, sqlc.DeleteProbeCheckAssignmentsForProbeParams{
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

	if err := postgres.Queries(ctx, r.queries).DeleteProbeCheckAssignmentsForCheck(ctx, sqlc.DeleteProbeCheckAssignmentsForCheckParams{
		ProjectID: projectID,
		CheckID:   checkID,
	}); err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *AssignmentRepository) UpsertProbeCheckAssignment(ctx context.Context, input domainassignment.AssignmentWrite) error {
	ctx, span := postgres.StartDBSpan(ctx, pgassignmentTracer, "probe_check_assignments", "postgres.assignments.upsert", "INSERT", "UPSERT probe check assignment")
	defer span.End()

	projectID, probeID, checkID, err := parseAssignmentWriteIDs(input.ProjectID, input.ProbeID, input.CheckID)
	if err != nil {
		return err
	}

	if err := postgres.Queries(ctx, r.queries).UpsertProbeCheckAssignment(ctx, sqlc.UpsertProbeCheckAssignmentParams{
		ProjectID:       projectID,
		ProbeID:         probeID,
		CheckID:         checkID,
		CheckVersion:    input.CheckVersion,
		SelectorVersion: input.SelectorVersion,
	}); err != nil {
		postgres.RecordDBSpanError(span, err)
		return mapAssignmentWriteError(err)
	}

	return nil
}

func (r *AssignmentRepository) DeleteStaleAssignmentsForProbe(ctx context.Context, projectIDValue, probeIDValue string, keepCheckIDValues []string) error {
	ctx, span := postgres.StartDBSpan(ctx, pgassignmentTracer, "probe_check_assignments", "postgres.assignments.delete_stale_for_probe", "DELETE", "DELETE stale probe check assignments for probe")
	defer span.End()

	projectID, probeID, err := parseProjectAndProbeIDs(projectIDValue, probeIDValue)
	if err != nil {
		return err
	}
	keepCheckIDs, err := parseUUIDList(keepCheckIDValues, domaincheck.ErrInvalidInput)
	if err != nil {
		return err
	}

	if err := postgres.Queries(ctx, r.queries).DeleteStaleProbeCheckAssignmentsForProbe(ctx, sqlc.DeleteStaleProbeCheckAssignmentsForProbeParams{
		ProjectID: projectID,
		ProbeID:   probeID,
		CheckIds:  keepCheckIDs,
	}); err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *AssignmentRepository) DeleteStaleAssignmentsForCheck(ctx context.Context, projectIDValue, checkIDValue, checkVersion, selectorVersion string, keepProbeIDValues []string) error {
	ctx, span := postgres.StartDBSpan(ctx, pgassignmentTracer, "probe_check_assignments", "postgres.assignments.delete_stale_for_check", "DELETE", "DELETE stale probe check assignments for check")
	defer span.End()

	projectID, checkID, err := parseProjectAndCheckIDs(projectIDValue, checkIDValue)
	if err != nil {
		return err
	}
	keepProbeIDs, err := parseUUIDList(keepProbeIDValues, domainprobe.ErrInvalidInput)
	if err != nil {
		return err
	}

	if err := postgres.Queries(ctx, r.queries).DeleteStaleProbeCheckAssignments(ctx, sqlc.DeleteStaleProbeCheckAssignmentsParams{
		ProjectID:       projectID,
		CheckID:         checkID,
		CheckVersion:    checkVersion,
		SelectorVersion: selectorVersion,
		ProbeIds:        keepProbeIDs,
	}); err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
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

	rows, err := postgres.Queries(ctx, r.queries).ListProjectAssignments(ctx, sqlc.ListProjectAssignmentsParams{
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

func parseAssignmentWriteIDs(projectIDValue, probeIDValue, checkIDValue string) (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	projectID, probeID, err := parseProjectAndProbeIDs(projectIDValue, probeIDValue)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	checkID, err := postgres.ParseUUID(checkIDValue, domaincheck.ErrCheckNotFound)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}

	return projectID, probeID, checkID, nil
}

func parseUUIDList(values []string, invalidErr error) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		id, err := postgres.ParseUUID(value, invalidErr)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
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
