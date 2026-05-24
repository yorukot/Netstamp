package pgtcp

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
)

type TCPRepository struct {
	queries *sqlc.Queries
	tx      *postgres.Transactor
}

func NewTCPRepository(pool *pgxpool.Pool) *TCPRepository {
	return &TCPRepository{
		queries: sqlc.New(pool),
		tx:      postgres.NewTransactor(pool),
	}
}

func (r *TCPRepository) CreateTCPResults(ctx context.Context, inputs []domaintcp.ResultStorageInput) error {
	ctx, span := postgres.StartDBSpan(ctx, pgtcpTracer, "tcp_results", "postgres.tcp_results.create_batch", "INSERT", "INSERT tcp result batch")
	defer span.End()

	err := r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := r.queries.WithTx(tx)
		for _, input := range inputs {
			createErr := q.CreateTCPResult(ctx, sqlc.CreateTCPResultParams{
				ProbeStorageID:    input.ProbeStorageID,
				CheckStorageID:    input.CheckStorageID,
				StartedAt:         timestamptz(input.StartedAt),
				FinishedAt:        timestamptz(input.FinishedAt),
				DurationMs:        input.DurationMs,
				Status:            sqlcTCPStatus(input.Status),
				ConnectDurationMs: input.ConnectDurationMs,
				ResolvedIp:        input.ResolvedIP,
				IpFamily:          sqlcIPFamily(input.IPFamily),
				ErrorCode:         input.ErrorCode,
				ErrorMessage:      input.ErrorMessage,
			})
			if createErr != nil {
				return mapTCPResultWriteError(createErr)
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
