package result

import (
	"net/netip"
	"testing"
	"time"

	agentworker "github.com/yorukot/netstamp/internal/agent/worker"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

func TestGroupResultsSeparatesPingAndTraceroutePayloads(t *testing.T) {
	startedAt := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	hopAddr := netip.MustParseAddr("192.0.2.1")
	rttMin := 1.5
	groups := groupResults([]agentworker.ResultEnvelope{{
		CheckID: "ping-check",
		Type:    domaincheck.TypePing,
		Ping: domainping.Result{
			StartedAt:     startedAt,
			FinishedAt:    startedAt.Add(time.Second),
			DurationMs:    1000,
			Status:        domainping.StatusSuccessful,
			SentCount:     1,
			ReceivedCount: 1,
			RttSamplesMs:  []float64{10},
		},
	}, {
		CheckID: "trace-check",
		Type:    domaincheck.TypeTraceroute,
		Traceroute: domaintraceroute.Result{
			StartedAt:          startedAt,
			FinishedAt:         startedAt.Add(2 * time.Second),
			DurationMs:         2000,
			Status:             domaintraceroute.StatusPartial,
			DestinationReached: false,
			HopCount:           1,
			Hops: []domaintraceroute.HopResult{{
				HopIndex:      1,
				Address:       &hopAddr,
				SentCount:     2,
				ReceivedCount: 1,
				LossPercent:   50,
				RttMinMs:      &rttMin,
				RttSamplesMs:  []float64{1.5},
			}},
		},
	}})

	if len(groups) != 2 {
		t.Fatalf("expected two groups, got %#v", groups)
	}
	if groups[0].CheckID != "ping-check" || len(groups[0].Ping) != 1 || len(groups[0].Traceroute) != 0 {
		t.Fatalf("unexpected ping group: %#v", groups[0])
	}
	if groups[1].CheckID != "trace-check" || len(groups[1].Traceroute) != 1 || len(groups[1].Ping) != 0 {
		t.Fatalf("unexpected traceroute group: %#v", groups[1])
	}
	trace := groups[1].Traceroute[0]
	if trace.Status != domaintraceroute.StatusPartial || trace.HopCount != 1 || len(trace.Hops) != 1 {
		t.Fatalf("unexpected traceroute payload: %#v", trace)
	}
	if trace.Hops[0].Address == nil || *trace.Hops[0].Address != hopAddr || len(trace.Hops[0].RttSamplesMs) != 1 || trace.Hops[0].RttSamplesMs[0] != 1.5 {
		t.Fatalf("unexpected traceroute hop payload: %#v", trace.Hops[0])
	}
}
