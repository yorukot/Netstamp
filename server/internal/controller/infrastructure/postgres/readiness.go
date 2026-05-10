package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewReadinessCheck(pool *pgxpool.Pool) func(context.Context) error {
	return func(ctx context.Context) error {
		if pool == nil {
			return errors.New("database is not configured")
		}
		return pool.Ping(ctx)
	}
}
