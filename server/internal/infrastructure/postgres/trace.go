package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func StartDBSpan(ctx context.Context, tracer trace.Tracer, collection, name, operation, summary string) (context.Context, trace.Span) {
	return tracer.Start(ctx, name, trace.WithAttributes(
		attribute.String("db.system.name", "postgresql"),
		attribute.String("db.operation.name", operation),
		attribute.String("db.collection.name", collection),
		attribute.String("db.query.summary", summary),
	))
}

func StartUserDBSpan(ctx context.Context, tracer trace.Tracer, name, operation, summary string) (context.Context, trace.Span) {
	return StartDBSpan(ctx, tracer, "users", name, operation, summary)
}

func RecordDBSpanError(span trace.Span, err error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, "database query failed")

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		span.SetAttributes(
			attribute.String("db.response.status_code", pgErr.Code),
			attribute.String("error.type", pgErr.Code),
		)
	}
}
