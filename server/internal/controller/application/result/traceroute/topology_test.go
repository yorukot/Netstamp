package traceroute

import (
	"fmt"
	"net/netip"
	"testing"
	"time"

	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

const (
	testProbeID = "33333333-3333-3333-3333-333333333333"
	testCheckID = "44444444-4444-4444-4444-444444444444"
)

func TestNewTopologyPreservesBoundedTimeouts(t *testing.T) {
	startedAt := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)
	gateway := netip.MustParseAddr("192.0.2.1")
	destination := netip.MustParseAddr("203.0.113.20")
	rttOne := 1.0
	rttTen := 10.0

	nodes, edges := newTopology([]domaintraceroute.TopologyRun{{
		StartedAt:  startedAt,
		ProbeID:    testProbeID,
		ProbeName:  "fra-bm-02",
		CheckID:    testCheckID,
		CheckName:  "validator-route",
		Target:     "validator.example",
		ResolvedIP: &destination,
		Hops: []domaintraceroute.TopologyHop{
			{HopIndex: 1, Address: &gateway, LossPercent: 0, RttAvgMs: &rttOne},
			{HopIndex: 2, LossPercent: 100},
			{HopIndex: 3, Address: &destination, LossPercent: 0, RttAvgMs: &rttTen},
		},
	}})

	gatewayNodeID := "ip:1:" + gateway.String()
	timeoutNodeID := topologyTimeoutNodeID(2)
	destinationNodeID := "ip:3:" + destination.String()
	if len(nodes) != 4 {
		t.Fatalf("expected probe, gateway, timeout, and destination nodes, got %#v", nodes)
	}
	if len(edges) != 3 {
		t.Fatalf("expected route edges through bounded timeout, got %#v", edges)
	}
	topologyEdgeByID(t, edges, "probe:"+testProbeID+"->"+gatewayNodeID)
	timeoutNode := topologyNodeByID(t, nodes, timeoutNodeID)
	if timeoutNode.Kind != "unknown" || timeoutNode.HopIndex == nil || *timeoutNode.HopIndex != 2 {
		t.Fatalf("unexpected bounded timeout node: %#v", timeoutNode)
	}
	timeoutEdge := topologyEdgeByID(t, edges, gatewayNodeID+"->"+timeoutNodeID)
	if timeoutEdge.LossPercent == nil || *timeoutEdge.LossPercent != 100 {
		t.Fatalf("timeout loss should stay on the timeout edge: %#v", timeoutEdge)
	}
	topologyEdgeByID(t, edges, timeoutNodeID+"->"+destinationNodeID)
	if topologyEdgeExists(edges, gatewayNodeID+"->"+destinationNodeID) {
		t.Fatalf("bounded timeout should not be hidden behind a direct edge: %#v", edges)
	}
}

func TestNewTopologySuppressesTimeoutWhenTracerouteHopHasKnownResponse(t *testing.T) {
	firstStartedAt := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)
	secondStartedAt := firstStartedAt.Add(time.Minute)
	gateway := netip.MustParseAddr("192.0.2.1")
	knownHop := netip.MustParseAddr("198.51.100.10")

	nodes, edges := newTopology([]domaintraceroute.TopologyRun{
		{
			StartedAt: firstStartedAt,
			ProbeID:   testProbeID,
			ProbeName: "fra-bm-02",
			CheckID:   testCheckID,
			CheckName: "validator-route",
			Target:    "validator.example",
			Hops: []domaintraceroute.TopologyHop{
				{HopIndex: 1, Address: &gateway, LossPercent: 0},
				{HopIndex: 2, LossPercent: 100},
			},
		},
		{
			StartedAt: secondStartedAt,
			ProbeID:   testProbeID,
			ProbeName: "fra-bm-02",
			CheckID:   testCheckID,
			CheckName: "validator-route",
			Target:    "validator.example",
			Hops: []domaintraceroute.TopologyHop{
				{HopIndex: 1, Address: &gateway, LossPercent: 0},
				{HopIndex: 2, Address: &knownHop, LossPercent: 0},
			},
		},
	})

	gatewayNodeID := "ip:1:" + gateway.String()
	knownHopNodeID := "ip:2:" + knownHop.String()
	timeoutNodeID := topologyTimeoutNodeID(2)
	if topologyNodeExists(nodes, timeoutNodeID) {
		t.Fatalf("timeout node should be suppressed when the traceroute hop has a known response: %#v", nodes)
	}
	topologyNodeByID(t, nodes, knownHopNodeID)
	topologyEdgeByID(t, edges, gatewayNodeID+"->"+knownHopNodeID)
}

func TestNewTopologyMergesDuplicateTimeoutsForSameRunHop(t *testing.T) {
	startedAt := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)
	gateway := netip.MustParseAddr("192.0.2.1")
	destination := netip.MustParseAddr("203.0.113.20")

	nodes, edges := newTopology([]domaintraceroute.TopologyRun{{
		StartedAt:  startedAt,
		ProbeID:    testProbeID,
		ProbeName:  "fra-bm-02",
		CheckID:    testCheckID,
		CheckName:  "validator-route",
		Target:     "validator.example",
		ResolvedIP: &destination,
		Hops: []domaintraceroute.TopologyHop{
			{HopIndex: 1, Address: &gateway, LossPercent: 0},
			{HopIndex: 2, LossPercent: 100},
			{HopIndex: 2, LossPercent: 100},
			{HopIndex: 3, Address: &destination, LossPercent: 0},
		},
	}})

	timeoutNodeID := topologyTimeoutNodeID(2)
	timeoutNode := topologyNodeByID(t, nodes, timeoutNodeID)
	if timeoutNode.SeenCount != 1 {
		t.Fatalf("expected duplicate same-run timeout hops to aggregate into one observation, got %#v", timeoutNode)
	}
	if topologyEdgeExists(edges, timeoutNodeID+"->"+timeoutNodeID) {
		t.Fatalf("duplicate same-run timeout hops should not create a self edge: %#v", edges)
	}
}

func topologyNodeByID(t *testing.T, nodes []TopologyNode, id string) TopologyNode {
	t.Helper()

	for _, node := range nodes {
		if node.ID == id {
			return node
		}
	}
	t.Fatalf("expected topology node %q in %#v", id, nodes)
	return TopologyNode{}
}

func topologyNodeExists(nodes []TopologyNode, id string) bool {
	for _, node := range nodes {
		if node.ID == id {
			return true
		}
	}
	return false
}

func topologyEdgeByID(t *testing.T, edges []TopologyEdge, id string) TopologyEdge {
	t.Helper()

	for _, edge := range edges {
		if edge.ID == id {
			return edge
		}
	}
	t.Fatalf("expected topology edge %q in %#v", id, edges)
	return TopologyEdge{}
}

func topologyEdgeExists(edges []TopologyEdge, id string) bool {
	for _, edge := range edges {
		if edge.ID == id {
			return true
		}
	}
	return false
}

func topologyTimeoutNodeID(hopIndex int32) string {
	return fmt.Sprintf("timeout:%s:%s:%d", testProbeID, testCheckID, hopIndex)
}
