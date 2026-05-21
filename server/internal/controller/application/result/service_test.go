package result

import (
	"context"
	"errors"
	"net/netip"
	"slices"
	"testing"
	"time"

	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainresult "github.com/yorukot/netstamp/internal/domain/result"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

const (
	testUserID    = "11111111-1111-1111-1111-111111111111"
	testProjectID = "22222222-2222-2222-2222-222222222222"
	testProbeID   = "33333333-3333-3333-3333-333333333333"
	testCheckID   = "44444444-4444-4444-4444-444444444444"
)

func TestQueryPingSeriesUsesDefaultsAndMapsPoints(t *testing.T) {
	now := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)
	pointTime := now.Add(-time.Hour)
	pings := &recordingPingSeriesRepository{
		result: domainping.SeriesResult{
			Points: []domainping.SeriesPoint{{
				Timestamp: pointTime,
				Value:     42.5,
			}},
			Resolution:  domainping.SeriesResolutionRaw,
			TotalPoints: 1,
		},
	}
	service := NewService(pings, &recordingTracerouteRunsRepository{}, &recordingMeasurementRepository{}, staticProjectAccess{})

	output, err := service.QueryPingSeries(context.Background(), QueryPingSeriesInput{
		CurrentUserID: testUserID,
		ProjectRef:    "vector-ix",
		ProbeID:       testProbeID,
		CheckID:       testCheckID,
		Now:           now,
	})
	if err != nil {
		t.Fatalf("expected query to succeed: %v", err)
	}

	if pings.got.ProjectID != testProjectID || pings.got.ProbeID != testProbeID || pings.got.CheckID != testCheckID {
		t.Fatalf("unexpected repository identity input: %#v", pings.got)
	}
	if !pings.got.From.Equal(now.Add(-24*time.Hour)) || !pings.got.To.Equal(now) {
		t.Fatalf("unexpected default range: from=%s to=%s", pings.got.From, pings.got.To)
	}
	if pings.got.Metric != string(PingMetricRTTAvgMS) {
		t.Fatalf("expected default metric %q, got %q", PingMetricRTTAvgMS, pings.got.Metric)
	}
	if pings.got.MaxDataPoints != defaultMaxDataPoint {
		t.Fatalf("expected default max data points %d, got %d", defaultMaxDataPoint, pings.got.MaxDataPoints)
	}
	if output.Query.FromMs != now.Add(-24*time.Hour).UnixMilli() || output.Query.ToMs != now.UnixMilli() {
		t.Fatalf("unexpected output query metadata: %#v", output.Query)
	}
	if len(output.Series) != 1 || len(output.Series[0].Points) != 1 {
		t.Fatalf("expected one series with one point, got %#v", output.Series)
	}
	if got := output.Series[0].Points[0]; got.TimestampMs != pointTime.UnixMilli() || got.Value != 42.5 {
		t.Fatalf("unexpected mapped point: %#v", got)
	}
	if output.Query.Resolution != string(domainping.SeriesResolutionRaw) || output.Query.TotalPoints != 1 {
		t.Fatalf("unexpected query sampling metadata: %#v", output.Query)
	}
}

func TestQueryPingSeriesRejectsInvalidMetric(t *testing.T) {
	service := NewService(&recordingPingSeriesRepository{}, &recordingTracerouteRunsRepository{}, &recordingMeasurementRepository{}, staticProjectAccess{})

	_, err := service.QueryPingSeries(context.Background(), QueryPingSeriesInput{
		CurrentUserID: testUserID,
		ProjectRef:    "vector-ix",
		ProbeID:       testProbeID,
		CheckID:       testCheckID,
		Metric:        "median",
		Now:           time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestQueryTracerouteRunsUsesDefaultsAndMapsRuns(t *testing.T) {
	now := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)
	startedAt := now.Add(-time.Hour)
	finishedAt := startedAt.Add(4 * time.Second)
	nextCursor := startedAt
	resolved := netip.MustParseAddr("93.184.216.34")
	hopAddr := netip.MustParseAddr("192.0.2.1")
	ipFamily := domainnetwork.IPFamilyInet
	hostname := "gateway.local"
	rttMin := 1.5
	rttAvg := 1.7
	rttMedian := 1.7
	rttMax := 1.9
	rttStddev := 0.2
	traceroutes := &recordingTracerouteRunsRepository{
		result: domaintraceroute.RunResult{
			Runs: []domaintraceroute.Run{{
				StartedAt:          startedAt,
				FinishedAt:         finishedAt,
				DurationMs:         4000,
				Status:             domaintraceroute.StatusPartial,
				ResolvedIP:         &resolved,
				IPFamily:           &ipFamily,
				DestinationReached: false,
				HopCount:           1,
				Hops: []domaintraceroute.Hop{{
					HopIndex:      1,
					Address:       &hopAddr,
					Hostname:      &hostname,
					SentCount:     3,
					ReceivedCount: 3,
					LossPercent:   0,
					RttMinMs:      &rttMin,
					RttAvgMs:      &rttAvg,
					RttMedianMs:   &rttMedian,
					RttMaxMs:      &rttMax,
					RttStddevMs:   &rttStddev,
					RttSamplesMs:  []float64{1.5, 1.7, 1.9},
				}},
			}},
			NextCursor: &nextCursor,
		},
	}
	service := NewService(&recordingPingSeriesRepository{}, traceroutes, &recordingMeasurementRepository{}, staticProjectAccess{})

	output, err := service.QueryTracerouteRuns(context.Background(), QueryTracerouteRunsInput{
		CurrentUserID: testUserID,
		ProjectRef:    "vector-ix",
		ProbeID:       testProbeID,
		CheckID:       testCheckID,
		Now:           now,
	})
	if err != nil {
		t.Fatalf("expected query to succeed: %v", err)
	}

	if traceroutes.got.ProjectID != testProjectID || traceroutes.got.ProbeID != testProbeID || traceroutes.got.CheckID != testCheckID {
		t.Fatalf("unexpected repository identity input: %#v", traceroutes.got)
	}
	if !traceroutes.got.From.Equal(now.Add(-24*time.Hour)) || !traceroutes.got.To.Equal(now) {
		t.Fatalf("unexpected default range: from=%s to=%s", traceroutes.got.From, traceroutes.got.To)
	}
	if traceroutes.got.Limit != defaultRunLimit || traceroutes.got.Cursor != nil {
		t.Fatalf("unexpected pagination query: %#v", traceroutes.got)
	}
	if output.Query.FromMs != now.Add(-24*time.Hour).UnixMilli() || output.Query.ToMs != now.UnixMilli() {
		t.Fatalf("unexpected output query metadata: %#v", output.Query)
	}
	if output.Query.Limit != defaultRunLimit || output.Query.NextCursor == nil || *output.Query.NextCursor != nextCursor.UnixMilli() {
		t.Fatalf("unexpected pagination metadata: %#v", output.Query)
	}
	if len(output.Runs) != 1 || len(output.Runs[0].Hops) != 1 {
		t.Fatalf("expected one run with one hop, got %#v", output.Runs)
	}
	run := output.Runs[0]
	if !run.StartedAt.Equal(startedAt) || run.Status != string(domaintraceroute.StatusPartial) || run.IPFamily == nil || *run.IPFamily != string(domainnetwork.IPFamilyInet) {
		t.Fatalf("unexpected mapped run: %#v", run)
	}
	if run.ResolvedIP == nil || *run.ResolvedIP != resolved {
		t.Fatalf("unexpected resolved ip: %#v", run.ResolvedIP)
	}
	hop := run.Hops[0]
	if hop.HopIndex != 1 || hop.Address == nil || *hop.Address != hopAddr || !slices.Equal(hop.RttSamplesMs, []float64{1.5, 1.7, 1.9}) {
		t.Fatalf("unexpected mapped hop: %#v", hop)
	}
}

func TestQueryTracerouteTopologyAggregatesRuns(t *testing.T) {
	now := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)
	gateway := netip.MustParseAddr("192.0.2.1")
	destination := netip.MustParseAddr("203.0.113.20")
	gatewayName := "gateway.local"
	rttOne := 1.0
	rttThree := 3.0
	rttTen := 10.0
	traceroutes := &recordingTracerouteRunsRepository{
		topologyResult: domaintraceroute.TopologyRunResult{
			Runs: []domaintraceroute.TopologyRun{
				{
					StartedAt:  now.Add(-time.Minute),
					ProbeID:    testProbeID,
					ProbeName:  "fra-bm-02",
					CheckID:    testCheckID,
					CheckName:  "validator-route",
					Target:     "validator.example",
					ResolvedIP: &destination,
					Hops: []domaintraceroute.TopologyHop{
						{HopIndex: 1, Address: &gateway, Hostname: &gatewayName, LossPercent: 0, RttAvgMs: &rttOne},
						{HopIndex: 2, Address: &destination, LossPercent: 0, RttAvgMs: &rttTen},
					},
				},
				{
					StartedAt:  now.Add(-2 * time.Minute),
					ProbeID:    testProbeID,
					ProbeName:  "fra-bm-02",
					CheckID:    testCheckID,
					CheckName:  "validator-route",
					Target:     "validator.example",
					ResolvedIP: &destination,
					Hops: []domaintraceroute.TopologyHop{
						{HopIndex: 1, Address: &gateway, Hostname: &gatewayName, LossPercent: 10, RttAvgMs: &rttThree},
						{HopIndex: 2, LossPercent: 100},
					},
				},
			},
		},
	}
	service := NewService(&recordingPingSeriesRepository{}, traceroutes, &recordingMeasurementRepository{}, staticProjectAccess{})

	output, err := service.QueryTracerouteTopology(context.Background(), QueryTracerouteTopologyInput{
		CurrentUserID: testUserID,
		ProjectRef:    "vector-ix",
		Now:           now,
	})
	if err != nil {
		t.Fatalf("expected query to succeed: %v", err)
	}

	if traceroutes.gotTopology.ProjectID != testProjectID || traceroutes.gotTopology.ProbeID != "" || traceroutes.gotTopology.CheckID != "" {
		t.Fatalf("unexpected repository topology input: %#v", traceroutes.gotTopology)
	}
	if traceroutes.gotTopology.Limit != defaultRunLimit {
		t.Fatalf("expected default topology limit %d, got %d", defaultRunLimit, traceroutes.gotTopology.Limit)
	}
	if output.Query.FromMs != now.Add(-24*time.Hour).UnixMilli() || output.Query.ToMs != now.UnixMilli() {
		t.Fatalf("unexpected output query metadata: %#v", output.Query)
	}

	probeNode := topologyNodeByID(t, output.Nodes, "probe:"+testProbeID)
	if probeNode.Kind != "probe" || probeNode.SeenCount != 2 || probeNode.ProbeID == nil || *probeNode.ProbeID != testProbeID {
		t.Fatalf("unexpected probe node: %#v", probeNode)
	}
	gatewayNode := topologyNodeByID(t, output.Nodes, "ip:"+gateway.String())
	if gatewayNode.SeenCount != 2 || gatewayNode.Hostname == nil || *gatewayNode.Hostname != gatewayName {
		t.Fatalf("unexpected gateway node: %#v", gatewayNode)
	}
	if gatewayNode.AvgRttMs == nil || *gatewayNode.AvgRttMs != 2 || gatewayNode.LossPercent == nil || *gatewayNode.LossPercent != 5 {
		t.Fatalf("unexpected gateway aggregate metrics: %#v", gatewayNode)
	}
	destinationNode := topologyNodeByID(t, output.Nodes, "ip:"+destination.String())
	if destinationNode.Kind != "destination" || destinationNode.SeenCount != 1 {
		t.Fatalf("unexpected destination node: %#v", destinationNode)
	}
	unknownNode := topologyNodeByID(t, output.Nodes, "unknown:2")
	if unknownNode.Kind != "unknown" || unknownNode.SeenCount != 1 || unknownNode.LossPercent == nil || *unknownNode.LossPercent != 100 {
		t.Fatalf("unexpected unknown node: %#v", unknownNode)
	}

	probeGatewayEdge := topologyEdgeByID(t, output.Edges, "probe:"+testProbeID+"->ip:"+gateway.String())
	if probeGatewayEdge.SeenCount != 2 || probeGatewayEdge.AvgRttMs == nil || *probeGatewayEdge.AvgRttMs != 2 {
		t.Fatalf("unexpected probe gateway edge: %#v", probeGatewayEdge)
	}
	if len(output.Edges) != 3 {
		t.Fatalf("expected three topology edges, got %#v", output.Edges)
	}
}

type staticProjectAccess struct{}

func (staticProjectAccess) GetProjectForUser(_ context.Context, projectRef, userID string) (domainproject.Project, error) {
	if projectRef != "vector-ix" || userID != testUserID {
		return domainproject.Project{}, domainproject.ErrProjectNotFound
	}
	return domainproject.Project{ID: testProjectID, Slug: "vector-ix"}, nil
}

type recordingPingSeriesRepository struct {
	got    domainping.SeriesQuery
	result domainping.SeriesResult
}

func (r *recordingPingSeriesRepository) ListPingSeries(_ context.Context, input domainping.SeriesQuery) (domainping.SeriesResult, error) {
	r.got = input
	return r.result, nil
}

type recordingTracerouteRunsRepository struct {
	got            domaintraceroute.RunQuery
	gotTopology    domaintraceroute.TopologyQuery
	result         domaintraceroute.RunResult
	topologyResult domaintraceroute.TopologyRunResult
}

func (r *recordingTracerouteRunsRepository) ListTracerouteRuns(_ context.Context, input domaintraceroute.RunQuery) (domaintraceroute.RunResult, error) {
	r.got = input
	return r.result, nil
}

func (r *recordingTracerouteRunsRepository) ListTracerouteTopologyRuns(_ context.Context, input domaintraceroute.TopologyQuery) (domaintraceroute.TopologyRunResult, error) {
	r.gotTopology = input
	return r.topologyResult, nil
}

type recordingMeasurementRepository struct {
	got    domainresult.MeasurementQuery
	result domainresult.MeasurementResult
}

func (r *recordingMeasurementRepository) ListMeasurements(_ context.Context, input domainresult.MeasurementQuery) (domainresult.MeasurementResult, error) {
	r.got = input
	return r.result, nil
}

func topologyNodeByID(t *testing.T, nodes []TracerouteTopologyNode, id string) TracerouteTopologyNode {
	t.Helper()

	for _, node := range nodes {
		if node.ID == id {
			return node
		}
	}
	t.Fatalf("expected topology node %q in %#v", id, nodes)
	return TracerouteTopologyNode{}
}

func topologyEdgeByID(t *testing.T, edges []TracerouteTopologyEdge, id string) TracerouteTopologyEdge {
	t.Helper()

	for _, edge := range edges {
		if edge.ID == id {
			return edge
		}
	}
	t.Fatalf("expected topology edge %q in %#v", id, edges)
	return TracerouteTopologyEdge{}
}
