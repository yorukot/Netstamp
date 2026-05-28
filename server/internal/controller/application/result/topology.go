package result

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

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

type tracerouteHopKey struct {
	probeID  string
	checkID  string
	hopIndex int32
}

func newTracerouteTopology(runs []domaintraceroute.TopologyRun) ([]TracerouteTopologyNode, []TracerouteTopologyEdge) {
	nodeIndex := make(map[string]int)
	edgeIndex := make(map[string]int)
	nodes := make([]tracerouteTopologyNodeAggregate, 0)
	edges := make([]tracerouteTopologyEdgeAggregate, 0)
	knownHopIndex := topologyKnownHopIndex(runs)

	for _, run := range runs {
		probeNode := topologyProbeNode(run)
		probeIndex := upsertTopologyNode(&nodes, nodeIndex, probeNode)
		nodes[probeIndex].node.SeenCount++

		sourceID := probeNode.ID
		for _, hop := range topologyRunHops(run.Hops) {
			var hopNode TracerouteTopologyNode
			if hop.Address == nil {
				if knownHopIndex[tracerouteHopKey{probeID: run.ProbeID, checkID: run.CheckID, hopIndex: hop.HopIndex}] {
					continue
				}
				hopNode = topologyTimeoutNode(run, hop)
			} else {
				hopNode = topologyHopNode(run, hop)
			}

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

func topologyKnownHopIndex(runs []domaintraceroute.TopologyRun) map[tracerouteHopKey]bool {
	values := make(map[tracerouteHopKey]bool)
	for _, run := range runs {
		for _, hop := range run.Hops {
			if hop.Address == nil {
				continue
			}
			values[tracerouteHopKey{probeID: run.ProbeID, checkID: run.CheckID, hopIndex: hop.HopIndex}] = true
		}
	}
	return values
}

func topologyRunHops(hops []domaintraceroute.TopologyHop) []domaintraceroute.TopologyHop {
	hopByIndex := make(map[int32]domaintraceroute.TopologyHop, len(hops))

	for _, hop := range hops {
		existing, ok := hopByIndex[hop.HopIndex]
		if !ok || topologyPreferHop(existing, hop) {
			hopByIndex[hop.HopIndex] = hop
		}
	}

	values := make([]domaintraceroute.TopologyHop, 0, len(hopByIndex))
	for _, hop := range hopByIndex {
		values = append(values, hop)
	}
	slices.SortFunc(values, func(a, b domaintraceroute.TopologyHop) int {
		switch {
		case a.HopIndex < b.HopIndex:
			return -1
		case a.HopIndex > b.HopIndex:
			return 1
		default:
			return 0
		}
	})

	return values
}

func topologyPreferHop(existing, next domaintraceroute.TopologyHop) bool {
	if existing.Address == nil && next.Address != nil {
		return true
	}
	if existing.Address != nil && next.Address == nil {
		return false
	}
	if existing.Hostname == nil && next.Hostname != nil {
		return true
	}
	if existing.RttAvgMs == nil && next.RttAvgMs != nil {
		return true
	}
	return false
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
	kind := "hop"
	if run.ResolvedIP != nil && *run.ResolvedIP == *hop.Address {
		kind = "destination"
	}
	label := hop.Address.String()
	if hop.Hostname != nil {
		label = *hop.Hostname
	}

	return TracerouteTopologyNode{
		ID:       fmt.Sprintf("ip:%d:%s", hop.HopIndex, hop.Address.String()),
		Kind:     kind,
		Label:    label,
		Address:  hop.Address,
		Hostname: hop.Hostname,
		HopIndex: &hopIndex,
	}
}

func topologyTimeoutNode(run domaintraceroute.TopologyRun, hop domaintraceroute.TopologyHop) TracerouteTopologyNode {
	hopIndex := hop.HopIndex
	return TracerouteTopologyNode{
		ID:       fmt.Sprintf("timeout:%s:%s:%d", run.ProbeID, run.CheckID, hop.HopIndex),
		Kind:     "unknown",
		Label:    fmt.Sprintf("hop %d timeout", hop.HopIndex),
		HopIndex: &hopIndex,
	}
}

func upsertTopologyNode(nodes *[]tracerouteTopologyNodeAggregate, index map[string]int, node TracerouteTopologyNode) int {
	if existing, ok := index[node.ID]; ok {
		mergeTopologyNode(&(*nodes)[existing].node, node)
		return existing
	}
	next := len(*nodes)
	index[node.ID] = next
	*nodes = append(*nodes, tracerouteTopologyNodeAggregate{node: node})
	return next
}

func mergeTopologyNode(existing *TracerouteTopologyNode, next TracerouteTopologyNode) {
	if existing.Kind != "destination" && next.Kind == "destination" {
		existing.Kind = next.Kind
	}
	if existing.Hostname == nil && next.Hostname != nil {
		existing.Hostname = next.Hostname
		existing.Label = next.Label
	}
	if existing.Address == nil && next.Address != nil {
		existing.Address = next.Address
	}
	if existing.ProbeID == nil && next.ProbeID != nil {
		existing.ProbeID = next.ProbeID
	}
	if existing.CheckID == nil && next.CheckID != nil {
		existing.CheckID = next.CheckID
	}
	if existing.Target == nil && next.Target != nil {
		existing.Target = next.Target
	}
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
