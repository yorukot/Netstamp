package measurement

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
	domainresult "github.com/yorukot/netstamp/internal/domain/result"
)

type Service struct {
	measurements  Repository
	projectAccess resultshared.ProjectAccess
}

func NewService(measurements Repository, projectAccess resultshared.ProjectAccess) *Service {
	return &Service{
		measurements:  measurements,
		projectAccess: projectAccess,
	}
}

func (s *Service) Query(ctx context.Context, input QueryInput) (Output, error) {
	ctx, span := resultTracer.Start(ctx, "result.measurements.query", trace.WithAttributes(
		attrProjectRef.String(input.ProjectRef),
		attrProbeID.String(input.ProbeID),
		attrCheckID.String(input.CheckID),
	))
	defer span.End()

	normalized, err := normalizeQueryInput(input)
	if err != nil {
		span.SetStatus(codes.Error, "invalid result query input")
		return Output{}, err
	}
	span.SetAttributes(
		attrProjectRef.String(normalized.base.ProjectRef),
		attrProbeID.String(normalized.base.ProbeID),
		attrCheckID.String(normalized.base.CheckID),
	)

	project, err := s.projectAccess.GetProjectForUser(ctx, normalized.base.ProjectRef, normalized.base.CurrentUserID)
	if err != nil {
		span.SetStatus(codes.Error, "project lookup failed")
		span.RecordError(err)
		return Output{}, err
	}
	span.SetAttributes(attrProjectID.String(project.ID))

	if s.measurements == nil {
		configuredErr := errors.New("measurement result repository is not configured")
		span.SetStatus(codes.Error, "measurement repository missing")
		span.RecordError(configuredErr)
		return Output{}, configuredErr
	}

	resultType := domainMeasurementType(normalized.resultType)
	// TODO: Rework measurements around the future result model instead of keeping this cross-type SQL feed.
	result, err := s.measurements.ListMeasurements(ctx, domainresult.MeasurementQuery{
		ProjectID: project.ID,
		ProbeID:   normalized.base.ProbeID,
		CheckID:   normalized.base.CheckID,
		Type:      resultType,
		Status:    normalized.status,
		From:      normalized.base.From,
		To:        normalized.base.To,
		Limit:     normalized.limit,
		Cursor:    normalized.cursor,
	})
	if err != nil {
		span.SetStatus(codes.Error, "measurements query failed")
		span.RecordError(err)
		return Output{}, err
	}

	return Output{
		Measurements: newMeasurements(result.Measurements),
		Query: QueryMetadata{
			FromMs:     normalized.base.From.UnixMilli(),
			ToMs:       normalized.base.To.UnixMilli(),
			Limit:      normalized.limit,
			NextCursor: resultshared.TimePtrMillis(result.NextCursor),
		},
	}, nil
}

func newMeasurements(measurements []domainresult.Measurement) []Measurement {
	values := make([]Measurement, 0, len(measurements))
	for _, measurement := range measurements {
		values = append(values, Measurement{
			Type:         string(measurement.Type),
			StartedAt:    measurement.StartedAt,
			FinishedAt:   measurement.FinishedAt,
			ProbeID:      measurement.ProbeID,
			CheckID:      measurement.CheckID,
			Status:       measurement.Status,
			DurationMs:   measurement.DurationMs,
			LatencyMs:    measurement.LatencyMs,
			LossPercent:  measurement.LossPercent,
			Metadata:     measurement.Metadata,
			ErrorCode:    measurement.ErrorCode,
			ErrorMessage: measurement.ErrorMessage,
		})
	}
	return values
}

func domainMeasurementType(value *string) *domainresult.MeasurementType {
	if value == nil {
		return nil
	}
	resultType := domainresult.MeasurementType(*value)
	return &resultType
}
