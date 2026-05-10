package pgprobe

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainselector "github.com/yorukot/netstamp/internal/domain/selector"
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

		if refreshErr := r.refreshEffectiveProbeChecksForProbe(ctx, q, projectID, row.ID, input.Enabled); refreshErr != nil {
			return refreshErr
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

func (r *ProbeRepository) UpdateProbe(ctx context.Context, input domainprobe.UpdateProbeStorageInput) (domainprobe.Probe, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprobeTracer, "probes", "postgres.probes.update", "UPDATE", "UPDATE probe, labels, and effective links")
	defer span.End()

	projectID, probeID, err := parseProjectAndProbeIDs(input.ProjectID, input.ProbeID)
	if err != nil {
		return domainprobe.Probe{}, err
	}
	labelIDs, err := pglabel.ParseLabelIDs(input.LabelIDs)
	if err != nil {
		return domainprobe.Probe{}, err
	}

	var updated domainprobe.Probe
	err = r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		if _, updateErr := q.UpdateProbe(ctx, sqlc.UpdateProbeParams{
			ProjectID: projectID,
			ID:        probeID,
			Name:      input.Name,
			Enabled:   input.Enabled,
			Location:  pointFromCoordinates(input.Longitude, input.Latitude),
			City:      input.City,
		}); updateErr != nil {
			if errors.Is(updateErr, pgx.ErrNoRows) {
				return domainprobe.ErrProbeNotFound
			}
			return mapUpdateProbeError(updateErr)
		}

		if input.ReplaceLabels {
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
		}

		if refreshErr := r.refreshEffectiveProbeChecksForProbe(ctx, q, projectID, probeID, input.Enabled); refreshErr != nil {
			return refreshErr
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
	ctx, span := postgres.StartDBSpan(ctx, pgprobeTracer, "probes", "postgres.probes.soft_delete", "UPDATE", "SOFT DELETE probe and effective links")
	defer span.End()

	projectID, probeID, err := parseProjectAndProbeIDs(projectIDValue, probeIDValue)
	if err != nil {
		return err
	}

	err = r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		if _, deleteErr := q.SoftDeleteProbe(ctx, sqlc.SoftDeleteProbeParams{
			ProjectID: projectID,
			ID:        probeID,
		}); deleteErr != nil {
			if errors.Is(deleteErr, pgx.ErrNoRows) {
				return domainprobe.ErrProbeNotFound
			}
			return deleteErr
		}

		return q.DeleteEffectiveProbeChecksForProbe(ctx, sqlc.DeleteEffectiveProbeChecksForProbeParams{
			ProjectID: projectID,
			ProbeID:   probeID,
		})
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *ProbeRepository) RotateProbeSecret(ctx context.Context, input domainprobe.RotateProbeSecretStorageInput) error {
	ctx, span := postgres.StartDBSpan(ctx, pgprobeTracer, "probe_credentials", "postgres.probes.rotate_secret", "UPDATE", "ROTATE probe credential")
	defer span.End()

	projectID, probeID, err := parseProjectAndProbeIDs(input.ProjectID, input.ProbeID)
	if err != nil {
		return err
	}

	if _, err := r.queries.RotateProbeCredential(ctx, sqlc.RotateProbeCredentialParams{
		ProjectID:  projectID,
		ID:         probeID,
		SecretHash: input.SecretHash,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainprobe.ErrProbeNotFound
		}
		postgres.RecordDBSpanError(span, err)
		return err
	}

	return nil
}

func (r *ProbeRepository) refreshEffectiveProbeChecksForProbe(ctx context.Context, queries *sqlc.Queries, projectID, probeID uuid.UUID, enabled bool) error {
	if !enabled {
		return queries.DeleteEffectiveProbeChecksForProbe(ctx, sqlc.DeleteEffectiveProbeChecksForProbeParams{
			ProjectID: projectID,
			ProbeID:   probeID,
		})
	}

	labelRows, err := queries.ListActiveLabelsForProbe(ctx, sqlc.ListActiveLabelsForProbeParams{
		ProjectID: projectID,
		ProbeID:   probeID,
	})
	if err != nil {
		return err
	}
	labels := mapLabels(labelRows)

	checkRows, err := queries.ListActiveChecksForProject(ctx, projectID)
	if err != nil {
		return err
	}

	matchedCheckIDs := make([]uuid.UUID, 0, len(checkRows))
	for _, row := range checkRows {
		selectorRaw := json.RawMessage(row.Selector)
		selector, parseErr := domainselector.Parse(selectorRaw)
		if parseErr != nil {
			return parseErr
		}
		if !selector.Matches(labels) {
			continue
		}
		matchedCheckIDs = append(matchedCheckIDs, row.ID)
		if linkErr := queries.UpsertEffectiveProbeCheck(ctx, sqlc.UpsertEffectiveProbeCheckParams{
			ProjectID:       projectID,
			ProbeID:         probeID,
			CheckID:         row.ID,
			CheckVersion:    domaincheck.CheckVersion(checkExecutionSpec(row)),
			SelectorVersion: domaincheck.SelectorVersion(selectorRaw),
		}); linkErr != nil {
			return mapUpdateProbeError(linkErr)
		}
	}

	return queries.DeleteStaleEffectiveProbeChecksForProbe(ctx, sqlc.DeleteStaleEffectiveProbeChecksForProbeParams{
		ProjectID: projectID,
		ProbeID:   probeID,
		CheckIds:  matchedCheckIDs,
	})
}

func checkExecutionSpec(row sqlc.ListActiveChecksForProjectRow) domaincheck.ExecutionSpec {
	return domaincheck.ExecutionSpec{
		Type:            domaincheck.Type(row.CheckType),
		Target:          row.Target,
		IntervalSeconds: row.IntervalSeconds,
		PingConfig: domainping.Config{
			PacketCount:     row.PacketCount,
			PacketSizeBytes: row.PacketSizeBytes,
			TimeoutMs:       row.TimeoutMs,
			IPFamily:        mapIPFamily(row.IpFamily),
		},
	}
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
