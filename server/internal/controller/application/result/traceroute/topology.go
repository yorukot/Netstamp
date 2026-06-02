package traceroute

import (
	"fmt"
	"slices"

	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type topologyNodeAggregate struct {
	node      TopologyNode
	rttSum    float64
	rttCount  int32
	lossSum   float64
	lossCount int32
}

type topologyEdgeAggregate struct {
	edge      TopologyEdge
	rttSum    float64
	rttCount  int32
	lossSum   float64
	lossCount int32
}

type hopKey struct {
	probeID  string
	checkID  string
	hopIndex int32
}

func newTopology(runs []domaintraceroute.TopologyRun) ([]TopologyNode, []TopologyEdge) {
	nodeIndex := make(map[string]int)
	edgeIndex := make(map[string]int)
	nodes := make([]topologyNodeAggregate, 0)
	edges := make([]topologyEdgeAggregate, 0)
	knownHopIndex := knownHopIndex(runs)

	for _, run := range runs {
		probeNode := topologyProbeNode(run)
		probeIndex := upsertTopologyNode(&nodes, nodeIndex, probeNode)
		nodes[probeIndex].node.SeenCount++

		sourceID := probeNode.ID
		for _, hop := range topologyRunHops(run.Hops) {
			var hopNode TopologyNode
			if hop.Address == nil {
				if knownHopIndex[hopKey{probeID: run.ProbeID, checkID: run.CheckID, hopIndex: hop.HopIndex}] {
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

	outputNodes := make([]TopologyNode, 0, len(nodes))
	for _, aggregate := range nodes {
		node := aggregate.node
		node.AvgRttMs = resultshared.AveragePtr(aggregate.rttSum, aggregate.rttCount)
		node.LossPercent = resultshared.AveragePtr(aggregate.lossSum, aggregate.lossCount)
		outputNodes = append(outputNodes, node)
	}

	outputEdges := make([]TopologyEdge, 0, len(edges))
	for _, aggregate := range edges {
		edge := aggregate.edge
		edge.AvgRttMs = resultshared.AveragePtr(aggregate.rttSum, aggregate.rttCount)
		edge.LossPercent = resultshared.AveragePtr(aggregate.lossSum, aggregate.lossCount)
		outputEdges = append(outputEdges, edge)
	}

	return outputNodes, outputEdges
}

func knownHopIndex(runs []domaintraceroute.TopologyRun) map[hopKey]bool {
	values := make(map[hopKey]bool)
	for _, run := range runs {
		for _, hop := range run.Hops {
			if hop.Address == nil {
				continue
			}
			values[hopKey{probeID: run.ProbeID, checkID: run.CheckID, hopIndex: hop.HopIndex}] = true
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

func topologyProbeNode(run domaintraceroute.TopologyRun) TopologyNode {
	return TopologyNode{
		ID:      "probe:" + run.ProbeID,
		Kind:    "probe",
		Label:   run.ProbeName,
		ProbeID: resultshared.StringPointer(run.ProbeID),
	}
}

func topologyHopNode(run domaintraceroute.TopologyRun, hop domaintraceroute.TopologyHop) TopologyNode {
	hopIndex := hop.HopIndex
	kind := "hop"
	if run.ResolvedIP != nil && *run.ResolvedIP == *hop.Address {
		kind = "destination"
	}
	label := hop.Address.String()
	if hop.Hostname != nil {
		label = *hop.Hostname
	}

	return TopologyNode{
		ID:       fmt.Sprintf("ip:%d:%s", hop.HopIndex, hop.Address.String()),
		Kind:     kind,
		Label:    label,
		Address:  hop.Address,
		Hostname: hop.Hostname,
		HopIndex: &hopIndex,
	}
}

func topologyTimeoutNode(run domaintraceroute.TopologyRun, hop domaintraceroute.TopologyHop) TopologyNode {
	hopIndex := hop.HopIndex
	return TopologyNode{
		ID:       fmt.Sprintf("timeout:%s:%s:%d", run.ProbeID, run.CheckID, hop.HopIndex),
		Kind:     "unknown",
		Label:    fmt.Sprintf("hop %d timeout", hop.HopIndex),
		HopIndex: &hopIndex,
	}
}

func upsertTopologyNode(nodes *[]topologyNodeAggregate, index map[string]int, node TopologyNode) int {
	if existing, ok := index[node.ID]; ok {
		mergeTopologyNode(&(*nodes)[existing].node, node)
		return existing
	}
	next := len(*nodes)
	index[node.ID] = next
	*nodes = append(*nodes, topologyNodeAggregate{node: node})
	return next
}

func mergeTopologyNode(existing *TopologyNode, next TopologyNode) {
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

func upsertTopologyEdge(edges *[]topologyEdgeAggregate, index map[string]int, source, target string) int {
	id := source + "->" + target
	if existing, ok := index[id]; ok {
		return existing
	}
	next := len(*edges)
	index[id] = next
	*edges = append(*edges, topologyEdgeAggregate{
		edge: TopologyEdge{
			ID:     id,
			Source: source,
			Target: target,
		},
	})
	return next
}

func addTopologyNodeSample(node *topologyNodeAggregate, rttAvgMs, lossPercent *float64) {
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

func addTopologyEdgeSample(edge *topologyEdgeAggregate, rttAvgMs, lossPercent *float64) {
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
