package tx

import "context"

type Transactor interface {
	WithinTx(ctx context.Context, fn func(context.Context) error) error
}

type NoopTransactor struct{}

func (NoopTransactor) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}
