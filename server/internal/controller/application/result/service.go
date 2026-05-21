package result

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainresult "github.com/yorukot/netstamp/internal/domain/result"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type Service struct {
	pings         PingSeriesRepository
	traceroutes   TracerouteRunsRepository
	measurements  MeasurementRepository
	projectAccess ProjectAccess
}

func NewService(pings PingSeriesRepository, traceroutes TracerouteRunsRepository, measurements MeasurementRepository, projectAccess ProjectAccess) *Service {
	return &Service{
		pings:         pings,
		traceroutes:   traceroutes,
		measurements:  measurements,
		projectAccess: projectAccess,
	}
}

func (s *Service) QueryPingSeries(ctx context.Context, input QueryPingSeriesInput) (PingSeriesOutput, error) {
	ctx, span := resultTracer.Start(ctx, "result.ping.series.query", trace.WithAttributes(
		attrProjectRef.String(input.ProjectRef),
		attrProbeID.String(input.ProbeID),
		attrCheckID.String(input.CheckID),
	))
	defer span.End()

	normalized, err := normalizeQueryPingSeriesInput(input)
	if err != nil {
		span.SetStatus(codes.Error, "invalid result query input")
		return PingSeriesOutput{}, err
	}
	span.SetAttributes(
		attrProjectRef.String(normalized.projectRef),
		attrProbeID.String(normalized.probeID),
		attrCheckID.String(normalized.checkID),
		attrMetric.String(string(normalized.metric)),
	)

	project, err := s.projectAccess.GetProjectForUser(ctx, normalized.projectRef, normalized.currentUserID)
	if err != nil {
		span.SetStatus(codes.Error, "project lookup failed")
		span.RecordError(err)
		return PingSeriesOutput{}, err
	}
	span.SetAttributes(attrProjectID.String(project.ID))

	if s.pings == nil {
		configuredErr := errors.New("ping result repository is not configured")
		span.SetStatus(codes.Error, "ping repository missing")
		span.RecordError(configuredErr)
		return PingSeriesOutput{}, configuredErr
	}

	result, err := s.pings.ListPingSeries(ctx, domainping.SeriesQuery{
		ProjectID:     project.ID,
		ProbeID:       normalized.probeID,
		CheckID:       normalized.checkID,
		From:          normalized.from,
		To:            normalized.to,
		Metric:        string(normalized.metric),
		MaxDataPoints: normalized.maxDataPoints,
	})
	if err != nil {
		span.SetStatus(codes.Error, "ping series query failed")
		span.RecordError(err)
		return PingSeriesOutput{}, err
	}

	return PingSeriesOutput{
		Series: []Series{{
			Name: string(normalized.metric),
			Labels: SeriesLabels{
				ProbeID:   normalized.probeID,
				CheckID:   normalized.checkID,
				CheckType: "ping",
			},
			Unit:   unitForMetric(normalized.metric),
			Points: newSeriesPoints(result.Points),
		}},
		Query: QueryMetadata{
			FromMs:        normalized.from.UnixMilli(),
			ToMs:          normalized.to.UnixMilli(),
			MaxDataPoints: normalized.maxDataPoints,
			Resolution:    string(result.Resolution),
			TotalPoints:   result.TotalPoints,
		},
	}, nil
}

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

func (s *Service) QueryTracerouteTopology(ctx context.Context, input QueryTracerouteTopologyInput) (TracerouteTopologyOutput, error) {
	ctx, span := resultTracer.Start(ctx, "result.traceroute.topology.query", trace.WithAttributes(
		attrProjectRef.String(input.ProjectRef),
		attrProbeID.String(input.ProbeID),
		attrCheckID.String(input.CheckID),
	))
	defer span.End()

	normalized, err := normalizeQueryTracerouteTopologyInput(input)
	if err != nil {
		span.SetStatus(codes.Error, "invalid result query input")
		return TracerouteTopologyOutput{}, err
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
		return TracerouteTopologyOutput{}, err
	}
	span.SetAttributes(attrProjectID.String(project.ID))

	if s.traceroutes == nil {
		configuredErr := errors.New("traceroute result repository is not configured")
		span.SetStatus(codes.Error, "traceroute repository missing")
		span.RecordError(configuredErr)
		return TracerouteTopologyOutput{}, configuredErr
	}

	result, err := s.traceroutes.ListTracerouteTopologyRuns(ctx, domaintraceroute.TopologyQuery{
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
		return TracerouteTopologyOutput{}, err
	}

	nodes, edges := newTracerouteTopology(result.Runs)
	return TracerouteTopologyOutput{
		Nodes: nodes,
		Edges: edges,
		Query: TracerouteTopologyQueryMetadata{
			FromMs: normalized.from.UnixMilli(),
			ToMs:   normalized.to.UnixMilli(),
			Limit:  normalized.limit,
		},
	}, nil
}

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

func newSeriesPoints(points []domainping.SeriesPoint) []SeriesPoint {
	values := make([]SeriesPoint, 0, len(points))
	for _, point := range points {
		values = append(values, SeriesPoint{
			TimestampMs: point.Timestamp.UTC().UnixMilli(),
			Value:       point.Value,
		})
	}
	return values
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

type tracerouteTopologyNodeAggregate struct {
	node      TracerouteTopologyNode
	rttSum    float64
	rttCount  int32
	lossSum   float64
	lossCount int32
}

type tracerouteTopologyEdgeAggregate struct {
	edge      TracerouteTopologyEdge
	rttSum    float64
	rttCount  int32
	lossSum   float64
	lossCount int32
}

func newTracerouteTopology(runs []domaintraceroute.TopologyRun) ([]TracerouteTopologyNode, []TracerouteTopologyEdge) {
	nodeIndex := make(map[string]int)
	edgeIndex := make(map[string]int)
	nodes := make([]tracerouteTopologyNodeAggregate, 0)
	edges := make([]tracerouteTopologyEdgeAggregate, 0)

	for _, run := range runs {
		probeNode := topologyProbeNode(run)
		probeIndex := upsertTopologyNode(&nodes, nodeIndex, probeNode)
		nodes[probeIndex].node.SeenCount++

		sourceID := probeNode.ID
		for _, hop := range run.Hops {
			hopNode := topologyHopNode(run, hop)
			hopIndex := upsertTopologyNode(&nodes, nodeIndex, hopNode)
			addTopologyNodeSample(&nodes[hopIndex], hop.RttAvgMs, &hop.LossPercent)

			edgePosition := upsertTopologyEdge(&edges, edgeIndex, sourceID, hopNode.ID)
			addTopologyEdgeSample(&edges[edgePosition], hop.RttAvgMs, &hop.LossPercent)
			sourceID = hopNode.ID
		}
	}

	outputNodes := make([]TracerouteTopologyNode, 0, len(nodes))
	for _, aggregate := range nodes {
		node := aggregate.node
		node.AvgRttMs = averagePtr(aggregate.rttSum, aggregate.rttCount)
		node.LossPercent = averagePtr(aggregate.lossSum, aggregate.lossCount)
		outputNodes = append(outputNodes, node)
	}

	outputEdges := make([]TracerouteTopologyEdge, 0, len(edges))
	for _, aggregate := range edges {
		edge := aggregate.edge
		edge.AvgRttMs = averagePtr(aggregate.rttSum, aggregate.rttCount)
		edge.LossPercent = averagePtr(aggregate.lossSum, aggregate.lossCount)
		outputEdges = append(outputEdges, edge)
	}

	return outputNodes, outputEdges
}

func topologyProbeNode(run domaintraceroute.TopologyRun) TracerouteTopologyNode {
	return TracerouteTopologyNode{
		ID:      "probe:" + run.ProbeID,
		Kind:    "probe",
		Label:   run.ProbeName,
		ProbeID: stringPointer(run.ProbeID),
	}
}

func topologyHopNode(run domaintraceroute.TopologyRun, hop domaintraceroute.TopologyHop) TracerouteTopologyNode {
	hopIndex := hop.HopIndex
	if hop.Address == nil {
		return TracerouteTopologyNode{
			ID:       fmt.Sprintf("unknown:%d", hop.HopIndex),
			Kind:     "unknown",
			Label:    fmt.Sprintf("hop %d timeout", hop.HopIndex),
			HopIndex: &hopIndex,
		}
	}

	kind := "hop"
	if run.ResolvedIP != nil && *run.ResolvedIP == *hop.Address {
		kind = "destination"
	}
	label := hop.Address.String()
	if hop.Hostname != nil {
		label = *hop.Hostname
	}

	return TracerouteTopologyNode{
		ID:       "ip:" + hop.Address.String(),
		Kind:     kind,
		Label:    label,
		Address:  hop.Address,
		Hostname: hop.Hostname,
		HopIndex: &hopIndex,
	}
}

func upsertTopologyNode(nodes *[]tracerouteTopologyNodeAggregate, index map[string]int, node TracerouteTopologyNode) int {
	if existing, ok := index[node.ID]; ok {
		return existing
	}
	next := len(*nodes)
	index[node.ID] = next
	*nodes = append(*nodes, tracerouteTopologyNodeAggregate{node: node})
	return next
}

func upsertTopologyEdge(edges *[]tracerouteTopologyEdgeAggregate, index map[string]int, source, target string) int {
	id := source + "->" + target
	if existing, ok := index[id]; ok {
		return existing
	}
	next := len(*edges)
	index[id] = next
	*edges = append(*edges, tracerouteTopologyEdgeAggregate{
		edge: TracerouteTopologyEdge{
			ID:     id,
			Source: source,
			Target: target,
		},
	})
	return next
}

func addTopologyNodeSample(node *tracerouteTopologyNodeAggregate, rttAvgMs, lossPercent *float64) {
	node.node.SeenCount++
	if rttAvgMs != nil {
		node.rttSum += *rttAvgMs
		node.rttCount++
	}
	if lossPercent != nil {
		node.lossSum += *lossPercent
		node.lossCount++
	}
}

func addTopologyEdgeSample(edge *tracerouteTopologyEdgeAggregate, rttAvgMs, lossPercent *float64) {
	edge.edge.SeenCount++
	if rttAvgMs != nil {
		edge.rttSum += *rttAvgMs
		edge.rttCount++
	}
	if lossPercent != nil {
		edge.lossSum += *lossPercent
		edge.lossCount++
	}
}

func ipFamilyString(value *domainnetwork.IPFamily) *string {
	if value == nil {
		return nil
	}
	copied := string(*value)
	return &copied
}

func stringPointer(value string) *string {
	copied := value
	return &copied
}

func averagePtr(sum float64, count int32) *float64 {
	if count == 0 {
		return nil
	}
	average := sum / float64(count)
	return &average
}

func timePtrMillis(value *time.Time) *int64 {
	if value == nil {
		return nil
	}
	millis := value.UTC().UnixMilli()
	return &millis
}

func unitForMetric(metric PingMetric) string {
	switch metric {
	case PingMetricRTTAvgMS:
		return "ms"
	case PingMetricLossPercent, PingMetricSuccessRate:
		return "percent"
	default:
		return ""
	}
}
