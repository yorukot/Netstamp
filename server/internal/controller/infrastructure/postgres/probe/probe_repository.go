package pgprobe

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	pglabel "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/label"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type ProbeRepository struct {
	queries *sqlc.Queries
	tx      *postgres.Transactor
}

func NewProbeRepository(pool *pgxpool.Pool) *ProbeRepository {
	return &ProbeRepository{
		queries: sqlc.New(pool),
		tx:      postgres.NewTransactor(pool),
	}
}

func (r *ProbeRepository) GetActiveProbeCredential(ctx context.Context, probeID string) (domainprobe.Credential, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprobeTracer, "probes", "postgres.probes.get_active_credential", "SELECT", "SELECT active probe credential")
	defer span.End()

	id, err := postgres.ParseUUID(probeID, domainprobe.ErrProbeNotFound)
	if err != nil {
		return domainprobe.Credential{}, err
	}

	row, err := r.queries.GetActiveProbeCredential(ctx, id)
	if err != nil {
		err = mapProbeLookupError(err)
		postgres.RecordDBSpanError(span, err)
		return domainprobe.Credential{}, err
	}

	return mapProbeCredential(row), nil
}

func (r *ProbeRepository) UpdateProbeStatus(ctx context.Context, input domainprobe.Status) (domainprobe.Status, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprobeTracer, "probe_statuses", "postgres.probes.update_status", "UPDATE", "UPDATE probe status")
	defer span.End()

	probeID, err := postgres.ParseUUID(input.ProbeID, domainprobe.ErrProbeNotFound)
	if err != nil {
		return domainprobe.Status{}, err
	}

	row, err := r.queries.UpdateProbeStatus(ctx, sqlc.UpdateProbeStatusParams{
		ProbeID:      probeID,
		Status:       sqlcProbeState(input.State),
		AgentVersion: input.AgentVersion,
		Addrs:        input.Addrs,
	})
	if err != nil {
		err = mapProbeLookupError(err)
		postgres.RecordDBSpanError(span, err)
		return domainprobe.Status{}, err
	}

	return mapProbeStatus(row), nil
}

func (r *ProbeRepository) UpdateProbeIPFamilyCapabilities(ctx context.Context, input domainprobe.IPFamilyCapabilities) (domainprobe.Status, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprobeTracer, "probe_statuses", "postgres.probes.update_ip_family_capability", "UPDATE", "UPDATE probe IP family capability")
	defer span.End()

	probeID, err := postgres.ParseUUID(input.ProbeID, domainprobe.ErrProbeNotFound)
	if err != nil {
		return domainprobe.Status{}, err
	}

	row, err := r.queries.UpdateProbeIPFamilyCapabilities(ctx, sqlc.UpdateProbeIPFamilyCapabilitiesParams{
		UpdateV4: input.UpdateV4,
		PublicV4: input.PublicV4,
		UpdateV6: input.UpdateV6,
		PublicV6: input.PublicV6,
		ProbeID:  probeID,
	})
	if err != nil {
		err = mapProbeLookupError(err)
		postgres.RecordDBSpanError(span, err)
		return domainprobe.Status{}, err
	}

	return mapProbeStatus(row), nil
}

func (r *ProbeRepository) ListAssignments(ctx context.Context, probeID string) ([]domainassignment.Assignment, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprobeTracer, "probe_check_assignments", "postgres.probes.list_assignments", "SELECT", "SELECT active probe assignments")
	defer span.End()

	id, err := postgres.ParseUUID(probeID, domainprobe.ErrProbeNotFound)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListActiveAssignmentsForProbe(ctx, id)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	assignments := make([]domainassignment.Assignment, 0, len(rows))
	for _, row := range rows {
		assignments = append(assignments, mapAssignment(row))
	}

	return assignments, nil
}

func (r *ProbeRepository) ListActiveAssignmentsForProbeChecks(ctx context.Context, probeID string, checkIDValues []string) ([]domainassignment.Assignment, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprobeTracer, "probe_check_assignments", "postgres.probes.list_assignments_for_checks", "SELECT", "SELECT active probe assignments for checks")
	defer span.End()

	id, err := postgres.ParseUUID(probeID, domainprobe.ErrProbeNotFound)
	if err != nil {
		return nil, err
	}
	checkIDs, err := parseCheckIDs(checkIDValues)
	if err != nil {
		return nil, err
	}
	if len(checkIDs) == 0 {
		return nil, nil
	}

	rows, err := r.queries.ListActiveAssignmentsForProbeChecks(ctx, sqlc.ListActiveAssignmentsForProbeChecksParams{
		ProbeID:  id,
		CheckIds: checkIDs,
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	assignments := make([]domainassignment.Assignment, 0, len(rows))
	for _, row := range rows {
		assignments = append(assignments, mapAssignmentForProbeChecks(row))
	}

	return assignments, nil
}

func (r *ProbeRepository) CreateProbe(ctx context.Context, input domainprobe.Probe, secretHash string) (domainprobe.Probe, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprobeTracer, "probes", "postgres.probes.create_with_credentials", "INSERT", "INSERT probe, credential, status, and labels")
	defer span.End()

	projectID, err := postgres.ParseUUID(input.ProjectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return domainprobe.Probe{}, err
	}
	labelIDValues := make([]string, 0, len(input.Labels))
	for _, label := range input.Labels {
		labelIDValues = append(labelIDValues, label.ID)
	}
	labelIDs, err := pglabel.ParseLabelIDs(labelIDValues)
	if err != nil {
		return domainprobe.Probe{}, err
	}

	var created domainprobe.Probe
	err = r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		row, createErr := q.CreateProbe(ctx, sqlc.CreateProbeParams{
			ProjectID:    projectID,
			Name:         input.Name,
			Enabled:      input.Enabled,
			Location:     pointFromCoordinates(input.Longitude, input.Latitude),
			LocationName: input.LocationName,
		})
		if createErr != nil {
			return mapCreateProbeError(createErr)
		}

		if _, credentialErr := q.CreateProbeCredential(ctx, sqlc.CreateProbeCredentialParams{
			ProbeID:    row.ID,
			SecretHash: secretHash,
		}); credentialErr != nil {
			return mapCreateProbeCredentialError(credentialErr)
		}

		statusRow, statusErr := q.CreateProbeStatus(ctx, sqlc.CreateProbeStatusParams{
			ProbeID: row.ID,
			Status:  sqlc.ProbeStateOffline,
		})
		if statusErr != nil {
			return mapCreateProbeStatusError(statusErr)
		}

		for _, labelID := range labelIDs {
			if labelErr := q.CreateProbeLabel(ctx, sqlc.CreateProbeLabelParams{
				ProjectID: projectID,
				ProbeID:   row.ID,
				LabelID:   labelID,
			}); labelErr != nil {
				return mapCreateProbeLabelError(labelErr)
			}
		}

		status := mapProbeStatus(statusRow)
		created = mapCreateProbeRow(row)
		created.Status = &status
		return nil
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainprobe.Probe{}, err
	}

	return created, nil
}

func (r *ProbeRepository) ListProbesForProject(ctx context.Context, projectIDValue string) ([]domainprobe.Probe, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprobeTracer, "probes", "postgres.probes.list", "SELECT", "SELECT active probes for project")
	defer span.End()

	projectID, err := postgres.ParseUUID(projectIDValue, domainproject.ErrProjectNotFound)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListActiveProbesForProject(ctx, projectID)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	return mapListProbeRows(rows), nil
}

func (r *ProbeRepository) GetProbeForProject(ctx context.Context, projectIDValue, probeIDValue string) (domainprobe.Probe, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprobeTracer, "probes", "postgres.probes.select", "SELECT", "SELECT active probe for project")
	defer span.End()

	projectID, probeID, err := parseProjectAndProbeIDs(projectIDValue, probeIDValue)
	if err != nil {
		return domainprobe.Probe{}, err
	}

	rows, err := r.queries.GetActiveProbeRowsForProject(ctx, sqlc.GetActiveProbeRowsForProjectParams{
		ProjectID: projectID,
		ID:        probeID,
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainprobe.Probe{}, err
	}
	probe, ok := mapGetProbeRows(rows)
	if !ok {
		return domainprobe.Probe{}, domainprobe.ErrProbeNotFound
	}

	return probe, nil
}

func (r *ProbeRepository) UpdateProbe(ctx context.Context, input domainprobe.Probe) (domainprobe.Probe, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprobeTracer, "probes", "postgres.probes.update", "UPDATE", "UPDATE probe and labels")
	defer span.End()

	projectID, probeID, err := parseProjectAndProbeIDs(input.ProjectID, input.ID)
	if err != nil {
		return domainprobe.Probe{}, err
	}
	labelIDValues := make([]string, 0, len(input.Labels))
	for _, label := range input.Labels {
		labelIDValues = append(labelIDValues, label.ID)
	}
	labelIDs, err := pglabel.ParseLabelIDs(labelIDValues)
	if err != nil {
		return domainprobe.Probe{}, err
	}

	var updated domainprobe.Probe
	err = r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		if _, updateErr := q.UpdateProbe(ctx, sqlc.UpdateProbeParams{
			ProjectID:    projectID,
			ID:           probeID,
			Name:         input.Name,
			Enabled:      input.Enabled,
			Location:     pointFromCoordinates(input.Longitude, input.Latitude),
			LocationName: input.LocationName,
		}); updateErr != nil {
			if errors.Is(updateErr, pgx.ErrNoRows) {
				return domainprobe.ErrProbeNotFound
			}
			return mapUpdateProbeError(updateErr)
		}

		if deleteErr := q.DeleteProbeLabels(ctx, sqlc.DeleteProbeLabelsParams{
			ProjectID: projectID,
			ProbeID:   probeID,
		}); deleteErr != nil {
			return deleteErr
		}
		for _, labelID := range labelIDs {
			if labelErr := q.CreateProbeLabel(ctx, sqlc.CreateProbeLabelParams{
				ProjectID: projectID,
				ProbeID:   probeID,
				LabelID:   labelID,
			}); labelErr != nil {
				return mapCreateProbeLabelError(labelErr)
			}
		}

		rows, getErr := q.GetActiveProbeRowsForProject(ctx, sqlc.GetActiveProbeRowsForProjectParams{
			ProjectID: projectID,
			ID:        probeID,
		})
		if getErr != nil {
			return getErr
		}
		var ok bool
		updated, ok = mapGetProbeRows(rows)
		if !ok {
			return domainprobe.ErrProbeNotFound
		}

		return nil
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainprobe.Probe{}, err
	}

	return updated, nil
}

func (r *ProbeRepository) SoftDeleteProbe(ctx context.Context, projectIDValue, probeIDValue string) error {
	ctx, span := postgres.StartDBSpan(ctx, pgprobeTracer, "probes", "postgres.probes.soft_delete", "UPDATE", "SOFT DELETE probe")
	defer span.End()

	projectID, probeID, err := parseProjectAndProbeIDs(projectIDValue, probeIDValue)
	if err != nil {
		return err
	}

	_, err = r.queries.SoftDeleteProbe(ctx, sqlc.SoftDeleteProbeParams{
		ProjectID: projectID,
		ID:        probeID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainprobe.ErrProbeNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *ProbeRepository) RotateProbeSecret(ctx context.Context, input domainprobe.Probe, secretHash string) error {
	ctx, span := postgres.StartDBSpan(ctx, pgprobeTracer, "probe_credentials", "postgres.probes.rotate_secret", "UPDATE", "ROTATE probe credential")
	defer span.End()

	projectID, probeID, err := parseProjectAndProbeIDs(input.ProjectID, input.ID)
	if err != nil {
		return err
	}

	if _, err := r.queries.RotateProbeCredential(ctx, sqlc.RotateProbeCredentialParams{
		ProjectID:  projectID,
		ID:         probeID,
		SecretHash: secretHash,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainprobe.ErrProbeNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func pointFromCoordinates(longitude, latitude *float64) pgtype.Point {
	if longitude == nil || latitude == nil {
		return pgtype.Point{}
	}

	return pgtype.Point{
		P:     pgtype.Vec2{X: *longitude, Y: *latitude},
		Valid: true,
	}
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

func parseCheckIDs(values []string) ([]uuid.UUID, error) {
	if len(values) == 0 {
		return nil, nil
	}

	checkIDs := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		checkID, err := postgres.ParseUUID(value, domaincheck.ErrCheckNotFound)
		if err != nil {
			return nil, err
		}
		checkIDs = append(checkIDs, checkID)
	}

	return checkIDs, nil
}
