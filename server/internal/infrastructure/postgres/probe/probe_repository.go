package pgprobe

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

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
