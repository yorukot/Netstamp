package pgping

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres/sqlc"
)

type PingRepository struct {
	queries *sqlc.Queries
	tx      *postgres.Transactor
}

func NewPingRepository(pool *pgxpool.Pool) *PingRepository {
	return &PingRepository{
		queries: sqlc.New(pool),
		tx:      postgres.NewTransactor(pool),
	}
}

func (r *PingRepository) CreatePingResults(ctx context.Context, inputs []domainping.ResultStorageInput) error {
	ctx, span := postgres.StartDBSpan(ctx, pgpingTracer, "ping_results", "postgres.ping_results.create_batch", "INSERT", "INSERT ping result batch")
	defer span.End()

	err := r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)
		for _, input := range inputs {
			projectID, err := postgres.ParseUUID(input.ProjectID, domainping.ErrInvalidResult)
			if err != nil {
				return err
			}
			probeID, err := postgres.ParseUUID(input.ProbeID, domainping.ErrInvalidResult)
			if err != nil {
				return err
			}
			checkID, err := postgres.ParseUUID(input.CheckID, domainping.ErrInvalidResult)
			if err != nil {
				return err
			}

			createErr := q.CreatePingResult(ctx, sqlc.CreatePingResultParams{
				ProjectID:     projectID,
				CheckID:       checkID,
				ProbeID:       probeID,
				StartedAt:     pgtype.Timestamptz{Time: input.StartedAt.UTC(), Valid: true},
				FinishedAt:    pgtype.Timestamptz{Time: input.FinishedAt.UTC(), Valid: true},
				DurationMs:    input.DurationMs,
				Status:        sqlcPingStatus(input.Status),
				SentCount:     input.SentCount,
				ReceivedCount: input.ReceivedCount,
				LossPercent:   input.LossPercent,
				RttMinMs:      input.RttMinMs,
				RttAvgMs:      input.RttAvgMs,
				RttMedianMs:   input.RttMedianMs,
				RttMaxMs:      input.RttMaxMs,
				RttStddevMs:   input.RttStddevMs,
				RttSamplesMs:  input.RttSamplesMs,
				ResolvedIp:    input.ResolvedIP,
				IpFamily:      sqlcIPFamily(input.IPFamily),
				Raw:           input.Raw,
				ErrorCode:     input.ErrorCode,
				ErrorMessage:  input.ErrorMessage,
			})
			if createErr != nil {
				return mapPingResultWriteError(createErr)
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
