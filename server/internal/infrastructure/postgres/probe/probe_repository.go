package pgprobe

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres"
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

func (r *ProbeRepository) CreateProbe(ctx context.Context, input domainprobe.CreateProbeStorageInput) (domainprobe.Probe, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgprobeTracer, "probes", "postgres.probes.create_with_credentials", "INSERT", "INSERT probe, credential, status, and labels")
	defer span.End()

	projectID, err := postgres.ParseUUID(input.ProjectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return domainprobe.Probe{}, err
	}
	labelIDs, err := parseLabelIDs(input.LabelIDs)
	if err != nil {
		return domainprobe.Probe{}, err
	}

	var created domainprobe.Probe
	err = r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		labels, err := r.getActiveLabels(ctx, q, projectID, labelIDs)
		if err != nil {
			return err
		}

		row, err := q.CreateProbe(ctx, sqlc.CreateProbeParams{
			ProjectID: projectID,
			Name:      input.Name,
			Enabled:   input.Enabled,
			Location:  pointFromCoordinates(input.Longitude, input.Latitude),
			City:      input.City,
		})
		if err != nil {
			return mapCreateProbeError(err)
		}

		if _, err := q.CreateProbeCredential(ctx, sqlc.CreateProbeCredentialParams{
			ProbeID:    row.ID,
			SecretHash: input.SecretHash,
		}); err != nil {
			return mapCreateProbeCredentialError(err)
		}

		if _, err := q.CreateProbeStatus(ctx, sqlc.CreateProbeStatusParams{
			ProbeID: row.ID,
			Status:  sqlc.ProbeStateOffline,
		}); err != nil {
			return mapCreateProbeStatusError(err)
		}

		for _, labelID := range labelIDs {
			if err := q.CreateProbeLabel(ctx, sqlc.CreateProbeLabelParams{
				ProjectID: projectID,
				ProbeID:   row.ID,
				LabelID:   labelID,
			}); err != nil {
				return mapCreateProbeLabelError(err)
			}
		}

		created = mapProbe(row, labels)
		return nil
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainprobe.Probe{}, err
	}

	return created, nil
}

func (r *ProbeRepository) getActiveLabels(ctx context.Context, q *sqlc.Queries, projectID uuid.UUID, labelIDs []uuid.UUID) ([]domainprobe.Label, error) {
	if len(labelIDs) == 0 {
		return []domainprobe.Label{}, nil
	}

	rows, err := q.GetActiveLabelsByIDsForProject(ctx, sqlc.GetActiveLabelsByIDsForProjectParams{
		ProjectID: projectID,
		LabelIds:  labelIDs,
	})
	if err != nil {
		return nil, err
	}
	if len(rows) != len(labelIDs) {
		return nil, domainprobe.ErrLabelNotFound
	}

	return mapLabels(rows), nil
}

func parseLabelIDs(values []string) ([]uuid.UUID, error) {
	if len(values) == 0 {
		return nil, nil
	}

	labelIDs := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		labelID, err := postgres.ParseUUID(value, domainprobe.ErrInvalidInput)
		if err != nil {
			return nil, err
		}
		labelIDs = append(labelIDs, labelID)
	}

	return labelIDs, nil
}

func pointFromCoordinates(longitude *float64, latitude *float64) pgtype.Point {
	if longitude == nil || latitude == nil {
		return pgtype.Point{}
	}

	return pgtype.Point{
		P:     pgtype.Vec2{X: *longitude, Y: *latitude},
		Valid: true,
	}
}
