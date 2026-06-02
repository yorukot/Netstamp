package traceroute

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type Service struct {
	runs          RunsRepository
	projectAccess resultshared.ProjectAccess
}

func NewService(runs RunsRepository, projectAccess resultshared.ProjectAccess) *Service {
	return &Service{
		runs:          runs,
		projectAccess: projectAccess,
	}
}

func (s *Service) QueryRuns(ctx context.Context, input QueryRunsInput) (RunsOutput, error) {
	ctx, span := resultTracer.Start(ctx, "result.traceroute.runs.query", trace.WithAttributes(
		attrProjectRef.String(input.ProjectRef),
		attrProbeID.String(input.ProbeID),
		attrCheckID.String(input.CheckID),
	))
	defer span.End()

	normalized, err := normalizeQueryRunsInput(input)
	if err != nil {
		span.SetStatus(codes.Error, "invalid result query input")
		return RunsOutput{}, err
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
		return RunsOutput{}, err
	}
	span.SetAttributes(attrProjectID.String(project.ID))

	if s.runs == nil {
		configuredErr := errors.New("traceroute result repository is not configured")
		span.SetStatus(codes.Error, "traceroute repository missing")
		span.RecordError(configuredErr)
		return RunsOutput{}, configuredErr
	}

	result, err := s.runs.ListTracerouteRuns(ctx, domaintraceroute.RunQuery{
		ProjectID: project.ID,
		ProbeID:   normalized.base.ProbeID,
		CheckID:   normalized.base.CheckID,
		From:      normalized.base.From,
		To:        normalized.base.To,
		Limit:     normalized.limit,
		Cursor:    normalized.cursor,
	})
	if err != nil {
		span.SetStatus(codes.Error, "traceroute runs query failed")
		span.RecordError(err)
		return RunsOutput{}, err
	}

	return RunsOutput{
		Runs: newRuns(result.Runs),
		Query: RunsQueryMetadata{
			FromMs:     normalized.base.From.UnixMilli(),
			ToMs:       normalized.base.To.UnixMilli(),
			Limit:      normalized.limit,
			NextCursor: resultshared.TimePtrMillis(result.NextCursor),
		},
	}, nil
}

func (s *Service) QueryInsight(ctx context.Context, input QueryInsightInput) (InsightOutput, error) {
	ctx, span := resultTracer.Start(ctx, "result.traceroute.insight.query", trace.WithAttributes(
		attrProjectRef.String(input.ProjectRef),
		attrProbeID.String(input.ProbeID),
		attrCheckID.String(input.CheckID),
	))
	defer span.End()

	normalized, err := normalizeQueryInsightInput(input)
	if err != nil {
		span.SetStatus(codes.Error, "invalid result query input")
		return InsightOutput{}, err
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
		return InsightOutput{}, err
	}
	span.SetAttributes(attrProjectID.String(project.ID))

	if s.runs == nil {
		configuredErr := errors.New("traceroute result repository is not configured")
		span.SetStatus(codes.Error, "traceroute repository missing")
		span.RecordError(configuredErr)
		return InsightOutput{}, configuredErr
	}

	result, err := s.runs.ListTracerouteInsight(ctx, domaintraceroute.InsightQuery{
		ProjectID:     project.ID,
		ProbeID:       normalized.base.ProbeID,
		CheckID:       normalized.base.CheckID,
		From:          normalized.base.From,
		To:            normalized.base.To,
		MaxDataPoints: normalized.maxDataPoints,
	})
	if err != nil {
		span.SetStatus(codes.Error, "traceroute insight query failed")
		span.RecordError(err)
		return InsightOutput{}, err
	}

	return InsightOutput{
		Points: newInsightPoints(result.Points),
		Query: InsightQueryMetadata{
			FromMs:        normalized.base.From.UnixMilli(),
			ToMs:          normalized.base.To.UnixMilli(),
			MaxDataPoints: normalized.maxDataPoints,
			Resolution:    string(result.Resolution),
			TotalRuns:     result.TotalRuns,
		},
	}, nil
}

func (s *Service) QueryTopology(ctx context.Context, input QueryTopologyInput) (TopologyOutput, error) {
	ctx, span := resultTracer.Start(ctx, "result.traceroute.topology.query", trace.WithAttributes(
		attrProjectRef.String(input.ProjectRef),
		attrProbeID.String(input.ProbeID),
		attrCheckID.String(input.CheckID),
	))
	defer span.End()

	normalized, err := normalizeQueryTopologyInput(input)
	if err != nil {
		span.SetStatus(codes.Error, "invalid result query input")
		return TopologyOutput{}, err
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
		return TopologyOutput{}, err
	}
	span.SetAttributes(attrProjectID.String(project.ID))

	if s.runs == nil {
		configuredErr := errors.New("traceroute result repository is not configured")
		span.SetStatus(codes.Error, "traceroute repository missing")
		span.RecordError(configuredErr)
		return TopologyOutput{}, configuredErr
	}

	result, err := s.runs.ListTracerouteTopologyRuns(ctx, domaintraceroute.TopologyQuery{
		ProjectID: project.ID,
		ProbeID:   normalized.probeID,
		CheckID:   normalized.checkID,
		From:      normalized.from,
		To:        normalized.to,
		Limit:     normalized.limit,
	})
	if err != nil {
		span.SetStatus(codes.Error, "traceroute topology query failed")
		span.RecordError(err)
		return TopologyOutput{}, err
	}

	nodes, edges := newTopology(result.Runs)
	return TopologyOutput{
		Nodes: nodes,
		Edges: edges,
		Query: TopologyQueryMetadata{
			FromMs: normalized.from.UnixMilli(),
			ToMs:   normalized.to.UnixMilli(),
			Limit:  normalized.limit,
		},
	}, nil
}

func newRuns(runs []domaintraceroute.Run) []Run {
	values := make([]Run, 0, len(runs))
	for _, run := range runs {
		values = append(values, Run{
			StartedAt:          run.StartedAt,
			FinishedAt:         run.FinishedAt,
			DurationMs:         run.DurationMs,
			Status:             string(run.Status),
			ResolvedIP:         run.ResolvedIP,
			IPFamily:           resultshared.IPFamilyString(run.IPFamily),
			DestinationReached: run.DestinationReached,
			HopCount:           run.HopCount,
			ErrorCode:          run.ErrorCode,
			ErrorMessage:       run.ErrorMessage,
			Hops:               newHops(run.Hops),
		})
	}
	return values
}

func newInsightPoints(points []domaintraceroute.InsightPoint) []InsightPoint {
	values := make([]InsightPoint, 0, len(points))
	for _, point := range points {
		values = append(values, InsightPoint{
			TimestampMs:        point.Timestamp.UTC().UnixMilli(),
			BucketFromMs:       point.BucketFrom.UTC().UnixMilli(),
			BucketToMs:         point.BucketTo.UTC().UnixMilli(),
			RunStartedAt:       resultshared.TimePtr(point.RunStartedAt),
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

func newHops(hops []domaintraceroute.Hop) []Hop {
	values := make([]Hop, 0, len(hops))
	for _, hop := range hops {
		values = append(values, Hop{
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
