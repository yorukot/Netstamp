package pgprobe

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres"
	pglabel "github.com/yorukot/netstamp/internal/infrastructure/postgres/label"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres/sqlc"
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

func (r *ProbeRepository) UpdateProbeStatus(ctx context.Context, input domainprobe.UpdateStatusInput) (domainprobe.Status, error) {
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
		PublicV4:     input.PublicV4,
		PublicV6:     input.PublicV6,
		Addrs:        input.Addrs,
	})
	if err != nil {
		err = mapProbeLookupError(err)
		postgres.RecordDBSpanError(span, err)
		return domainprobe.Status{}, err
	}

	return mapProbeStatus(row), nil
}

func (r *ProbeRepository) ListAssignments(ctx context.Context, probeID string) ([]domaincheck.Assignment, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprobeTracer, "effective_probe_checks", "postgres.probes.list_assignments", "SELECT", "SELECT active probe assignments")
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

	assignments := make([]domaincheck.Assignment, 0, len(rows))
	for _, row := range rows {
		assignments = append(assignments, mapAssignment(row))
	}

	return assignments, nil
}

func (r *ProbeRepository) ListActiveAssignedCheckIDs(ctx context.Context, probeID string) ([]string, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprobeTracer, "effective_probe_checks", "postgres.probes.list_assigned_check_ids", "SELECT", "SELECT active assigned check ids")
	defer span.End()

	id, err := postgres.ParseUUID(probeID, domainprobe.ErrProbeNotFound)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListActiveAssignedCheckIDsForProbe(ctx, id)
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return nil, err
	}

	checkIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		checkIDs = append(checkIDs, row.String())
	}

	return checkIDs, nil
}

func (r *ProbeRepository) CreateProbe(ctx context.Context, input domainprobe.CreateProbeStorageInput) (domainprobe.Probe, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprobeTracer, "probes", "postgres.probes.create_with_credentials", "INSERT", "INSERT probe, credential, status, and labels")
	defer span.End()

	projectID, err := postgres.ParseUUID(input.ProjectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return domainprobe.Probe{}, err
	}
	labelIDs, err := pglabel.ParseLabelIDs(input.LabelIDs)
	if err != nil {
		return domainprobe.Probe{}, err
	}

	var created domainprobe.Probe
	err = r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		row, createErr := q.CreateProbe(ctx, sqlc.CreateProbeParams{
			ProjectID: projectID,
			Name:      input.Name,
			Enabled:   input.Enabled,
			Location:  pointFromCoordinates(input.Longitude, input.Latitude),
			City:      input.City,
		})
		if createErr != nil {
			return mapCreateProbeError(createErr)
		}

		if _, credentialErr := q.CreateProbeCredential(ctx, sqlc.CreateProbeCredentialParams{
			ProbeID:    row.ID,
			SecretHash: input.SecretHash,
		}); credentialErr != nil {
			return mapCreateProbeCredentialError(credentialErr)
		}

		if _, statusErr := q.CreateProbeStatus(ctx, sqlc.CreateProbeStatusParams{
			ProbeID: row.ID,
			Status:  sqlc.ProbeStateOffline,
		}); statusErr != nil {
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

		created = mapProbe(row)
		return nil
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainprobe.Probe{}, err
	}

	return created, nil
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
