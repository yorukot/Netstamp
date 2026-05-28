package result

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

func (s *Service) QueryTracerouteRuns(ctx context.Context, input QueryTracerouteRunsInput) (TracerouteRunsOutput, error) {
	ctx, span := resultTracer.Start(ctx, "result.traceroute.runs.query", trace.WithAttributes(
		attrProjectRef.String(input.ProjectRef),
		attrProbeID.String(input.ProbeID),
		attrCheckID.String(input.CheckID),
	))
	defer span.End()

	normalized, err := normalizeQueryTracerouteRunsInput(input)
	if err != nil {
		span.SetStatus(codes.Error, "invalid result query input")
		return TracerouteRunsOutput{}, err
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
		return TracerouteRunsOutput{}, err
	}
	span.SetAttributes(attrProjectID.String(project.ID))

	if s.traceroutes == nil {
		configuredErr := errors.New("traceroute result repository is not configured")
		span.SetStatus(codes.Error, "traceroute repository missing")
		span.RecordError(configuredErr)
		return TracerouteRunsOutput{}, configuredErr
	}

	result, err := s.traceroutes.ListTracerouteRuns(ctx, domaintraceroute.RunQuery{
		ProjectID: project.ID,
		ProbeID:   normalized.probeID,
		CheckID:   normalized.checkID,
		From:      normalized.from,
		To:        normalized.to,
		Limit:     normalized.limit,
		Cursor:    normalized.cursor,
	})
	if err != nil {
		span.SetStatus(codes.Error, "traceroute runs query failed")
		span.RecordError(err)
		return TracerouteRunsOutput{}, err
	}

	return TracerouteRunsOutput{
		Runs: newTracerouteRuns(result.Runs),
		Query: TracerouteRunsQueryMetadata{
			FromMs:     normalized.from.UnixMilli(),
			ToMs:       normalized.to.UnixMilli(),
			Limit:      normalized.limit,
			NextCursor: timePtrMillis(result.NextCursor),
		},
	}, nil
}

func (s *Service) QueryTracerouteInsight(ctx context.Context, input QueryTracerouteInsightInput) (TracerouteInsightOutput, error) {
	ctx, span := resultTracer.Start(ctx, "result.traceroute.insight.query", trace.WithAttributes(
		attrProjectRef.String(input.ProjectRef),
		attrProbeID.String(input.ProbeID),
		attrCheckID.String(input.CheckID),
	))
	defer span.End()

	normalized, err := normalizeQueryTracerouteInsightInput(input)
	if err != nil {
		span.SetStatus(codes.Error, "invalid result query input")
		return TracerouteInsightOutput{}, err
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
		return TracerouteInsightOutput{}, err
	}
	span.SetAttributes(attrProjectID.String(project.ID))

	if s.traceroutes == nil {
		configuredErr := errors.New("traceroute result repository is not configured")
		span.SetStatus(codes.Error, "traceroute repository missing")
		span.RecordError(configuredErr)
		return TracerouteInsightOutput{}, configuredErr
	}

	result, err := s.traceroutes.ListTracerouteInsight(ctx, domaintraceroute.InsightQuery{
		ProjectID:     project.ID,
		ProbeID:       normalized.probeID,
		CheckID:       normalized.checkID,
		From:          normalized.from,
		To:            normalized.to,
		MaxDataPoints: normalized.maxDataPoints,
	})
	if err != nil {
		span.SetStatus(codes.Error, "traceroute insight query failed")
		span.RecordError(err)
		return TracerouteInsightOutput{}, err
	}

	return TracerouteInsightOutput{
		Points: newTracerouteInsightPoints(result.Points),
		Query: TracerouteInsightQueryMetadata{
			FromMs:        normalized.from.UnixMilli(),
			ToMs:          normalized.to.UnixMilli(),
			MaxDataPoints: normalized.maxDataPoints,
			Resolution:    string(result.Resolution),
			TotalRuns:     result.TotalRuns,
		},
	}, nil
}

func newTracerouteRuns(runs []domaintraceroute.Run) []TracerouteRun {
	values := make([]TracerouteRun, 0, len(runs))
	for _, run := range runs {
		values = append(values, TracerouteRun{
			StartedAt:          run.StartedAt,
			FinishedAt:         run.FinishedAt,
			DurationMs:         run.DurationMs,
			Status:             string(run.Status),
			ResolvedIP:         run.ResolvedIP,
			IPFamily:           ipFamilyString(run.IPFamily),
			DestinationReached: run.DestinationReached,
			HopCount:           run.HopCount,
			ErrorCode:          run.ErrorCode,
			ErrorMessage:       run.ErrorMessage,
			Hops:               newTracerouteHops(run.Hops),
		})
	}
	return values
}

func newTracerouteInsightPoints(points []domaintraceroute.InsightPoint) []TracerouteInsightPoint {
	values := make([]TracerouteInsightPoint, 0, len(points))
	for _, point := range points {
		values = append(values, TracerouteInsightPoint{
			TimestampMs:        point.Timestamp.UTC().UnixMilli(),
			BucketFromMs:       point.BucketFrom.UTC().UnixMilli(),
			BucketToMs:         point.BucketTo.UTC().UnixMilli(),
			RunStartedAt:       timePtr(point.RunStartedAt),
			ResultCount:        point.ResultCount,
			FinalRttAvgMs:      point.FinalRttAvgMs,
			FinalLossPercent:   point.FinalLossPercent,
			HasLoss:            point.HasLoss,
			HasRouteChange:     point.HasRouteChange,
			DestinationReached: point.DestinationReached,
		})
	}
	return values
}

func newTracerouteHops(hops []domaintraceroute.Hop) []TracerouteHop {
	values := make([]TracerouteHop, 0, len(hops))
	for _, hop := range hops {
		values = append(values, TracerouteHop{
			HopIndex:      hop.HopIndex,
			Address:       hop.Address,
			Hostname:      hop.Hostname,
			SentCount:     hop.SentCount,
			ReceivedCount: hop.ReceivedCount,
			LossPercent:   hop.LossPercent,
			RttMinMs:      hop.RttMinMs,
			RttAvgMs:      hop.RttAvgMs,
			RttMedianMs:   hop.RttMedianMs,
			RttMaxMs:      hop.RttMaxMs,
			RttStddevMs:   hop.RttStddevMs,
			RttSamplesMs:  append([]float64(nil), hop.RttSamplesMs...),
			ErrorCode:     hop.ErrorCode,
			ErrorMessage:  hop.ErrorMessage,
		})
	}
	return values
}
