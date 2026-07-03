package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
)

type txContextKey struct{}

type Transactor struct {
	pool *pgxpool.Pool
}

func NewTransactor(pool *pgxpool.Pool) *Transactor {
	return &Transactor{pool: pool}
}

func (t *Transactor) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return t.InTx(ctx, func(ctx context.Context, _ pgx.Tx) error {
		return fn(ctx)
	})
}

func (t *Transactor) InTx(ctx context.Context, fn func(context.Context, pgx.Tx) error) error {
	if tx, ok := TxFromContext(ctx); ok {
		return fn(ctx, tx)
	}

	return pgx.BeginFunc(ctx, t.pool, func(tx pgx.Tx) error {
		return fn(context.WithValue(ctx, txContextKey{}, tx), tx)
	})
}

func TxFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txContextKey{}).(pgx.Tx)
	return tx, ok
}

func Queries(ctx context.Context, queries *sqlc.Queries) *sqlc.Queries {
	if tx, ok := TxFromContext(ctx); ok {
		return queries.WithTx(tx)
	}
	return queries
}
