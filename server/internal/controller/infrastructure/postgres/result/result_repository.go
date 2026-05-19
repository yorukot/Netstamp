package pgresult

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainresult "github.com/yorukot/netstamp/internal/domain/result"
)

type ResultRepository struct {
	queries *sqlc.Queries
}

func NewResultRepository(pool *pgxpool.Pool) *ResultRepository {
	return &ResultRepository{queries: sqlc.New(pool)}
}

func (r *ResultRepository) ListMeasurements(ctx context.Context, input domainresult.MeasurementQuery) (domainresult.MeasurementResult, error) {
	ctx, span := postgres.StartDBSpan(ctx, pgresultTracer, "results", "postgres.results.measurements", "SELECT", "SELECT project measurements")
	defer span.End()

	projectID, err := postgres.ParseUUID(input.ProjectID, domainproject.ErrProjectNotFound)
	if err != nil {
		return domainresult.MeasurementResult{}, err
	}
	probeID, err := optionalUUID(input.ProbeID, domainprobe.ErrInvalidInput)
	if err != nil {
		return domainresult.MeasurementResult{}, err
	}
	checkID, err := optionalUUID(input.CheckID, domaincheck.ErrInvalidInput)
	if err != nil {
		return domainresult.MeasurementResult{}, err
	}

	rows, err := r.queries.ListProjectMeasurements(ctx, sqlc.ListProjectMeasurementsParams{
		ProjectID:       projectID,
		ProbeID:         probeID,
		CheckID:         checkID,
		MeasurementType: measurementTypeParam(input.Type),
		Status:          input.Status,
		StartedAtFrom:   timestamptz(input.From),
		StartedAtTo:     timestamptz(input.To),
		CursorStartedAt: optionalTimestamptz(input.Cursor),
		LimitCount:      input.Limit + 1,
	})
	if err != nil {
		postgres.RecordDBSpanError(span, err)
		return domainresult.MeasurementResult{}, err
	}

	return mapMeasurements(rows, input.Limit)
}

func optionalUUID(value string, invalidErr error) (*uuid.UUID, error) {
	if value == "" {
		return nil, nil //nolint:nilnil // Nil means no optional UUID filter.
	}
	id, err := postgres.ParseUUID(value, invalidErr)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func measurementTypeParam(value *domainresult.MeasurementType) *string {
	if value == nil {
		return nil
	}
	output := string(*value)
	return &output
}

func timestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func optionalTimestamptz(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return timestamptz(*value)
}

func mapMeasurements(rows []sqlc.ListProjectMeasurementsRow, limit int32) (domainresult.MeasurementResult, error) {
	measurements := make([]domainresult.Measurement, 0, len(rows))
	for index, row := range rows {
		if int32(index) >= limit {
			nextCursor := row.StartedAt.Time
			return domainresult.MeasurementResult{
				Measurements: measurements,
				NextCursor:   &nextCursor,
			}, nil
		}
		measurements = append(measurements, domainresult.Measurement{
			Type:         domainresult.MeasurementType(row.MeasurementType),
			StartedAt:    row.StartedAt.Time,
			FinishedAt:   row.FinishedAt.Time,
			ProbeID:      row.ProbeID.String(),
			CheckID:      row.CheckID.String(),
			Status:       row.Status,
			DurationMs:   row.DurationMs,
			LatencyMs:    row.LatencyMs,
			LossPercent:  measurementLossPercent(row),
			Metadata:     stringPtr(row.Metadata),
			ErrorCode:    row.ErrorCode,
			ErrorMessage: row.ErrorMessage,
		})
	}

	return domainresult.MeasurementResult{Measurements: measurements}, nil
}

func measurementLossPercent(row sqlc.ListProjectMeasurementsRow) *float64 {
	if row.MeasurementType != string(domainresult.MeasurementTypePing) {
		return nil
	}
	return &row.LossPercent
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
