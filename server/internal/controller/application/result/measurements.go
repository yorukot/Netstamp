package result

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	domainresult "github.com/yorukot/netstamp/internal/domain/result"
)

func (s *Service) QueryMeasurements(ctx context.Context, input QueryMeasurementsInput) (MeasurementsOutput, error) {
	ctx, span := resultTracer.Start(ctx, "result.measurements.query", trace.WithAttributes(
		attrProjectRef.String(input.ProjectRef),
		attrProbeID.String(input.ProbeID),
		attrCheckID.String(input.CheckID),
	))
	defer span.End()

	normalized, err := normalizeQueryMeasurementsInput(input)
	if err != nil {
		span.SetStatus(codes.Error, "invalid result query input")
		return MeasurementsOutput{}, err
	}
	span.SetAttributes(
		attrProjectRef.String(normalized.projectRef),
		attrProbeID.String(normalized.probeID),
		attrCheckID.String(normalized.checkID),
	)

	project, err := s.projectAccess.GetProjectForUser(ctx, normalized.projectRef, normalized.currentUserID)
	if err != nil {
		span.SetStatus(codes.Error, "project lookup failed")
		span.RecordError(err)
		return MeasurementsOutput{}, err
	}
	span.SetAttributes(attrProjectID.String(project.ID))

	if s.measurements == nil {
		configuredErr := errors.New("measurement result repository is not configured")
		span.SetStatus(codes.Error, "measurement repository missing")
		span.RecordError(configuredErr)
		return MeasurementsOutput{}, configuredErr
	}

	resultType := domainMeasurementType(normalized.resultType)
	result, err := s.measurements.ListMeasurements(ctx, domainresult.MeasurementQuery{
		ProjectID: project.ID,
		ProbeID:   normalized.probeID,
		CheckID:   normalized.checkID,
		Type:      resultType,
		Status:    normalized.status,
		From:      normalized.from,
		To:        normalized.to,
		Limit:     normalized.limit,
		Cursor:    normalized.cursor,
	})
	if err != nil {
		span.SetStatus(codes.Error, "measurements query failed")
		span.RecordError(err)
		return MeasurementsOutput{}, err
	}

	return MeasurementsOutput{
		Measurements: newMeasurements(result.Measurements),
		Query: MeasurementQueryMetadata{
			FromMs:     normalized.from.UnixMilli(),
			ToMs:       normalized.to.UnixMilli(),
			Limit:      normalized.limit,
			NextCursor: timePtrMillis(result.NextCursor),
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
