package pgping

import (
	"testing"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
)

func TestStorageRTTSamplesReturnsEmptySliceForNil(t *testing.T) {
	got := storageRTTSamples(nil)
	if got == nil {
		t.Fatal("expected nil RTT samples to become an empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("expected empty RTT samples, got %#v", got)
	}
}

func TestStorageRTTSamplesCopiesValues(t *testing.T) {
	input := []float64{10, 20}
	got := storageRTTSamples(input)
	input[0] = 99

	if len(got) != 2 || got[0] != 10 || got[1] != 20 {
		t.Fatalf("expected copied RTT samples, got %#v", got)
	}
}

func TestUsePingRollupOnlyWhenRollupCoversMoreSourcePoints(t *testing.T) {
	tests := []struct {
		name         string
		rawPoints    int64
		rollupPoints int64
		want         bool
	}{
		{name: "raw and rollup equal prefers raw", rawPoints: 3, rollupPoints: 3, want: false},
		{name: "raw has unrefreshed historical data", rawPoints: 3, rollupPoints: 0, want: false},
		{name: "rollup covers retained historical data", rawPoints: 2, rollupPoints: 100, want: true},
		{name: "rollup only", rawPoints: 0, rollupPoints: 12, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := usePingRollup(tt.rawPoints, tt.rollupPoints); got != tt.want {
				t.Fatalf("usePingRollup(%d, %d) = %t, want %t", tt.rawPoints, tt.rollupPoints, got, tt.want)
			}
		})
	}
}

func TestPingRollupInsightSummaryMapsAggregateFields(t *testing.T) {
	got := pingRollupInsightSummary(sqlc.GetPingInsightRollupSummaryRow{
		TotalResults:  3,
		RttValueCount: 3,
		Samples:       11,
		AverageRttMs:  20,
		MaxRttMs:      50,
		LossPercent:   8.333,
		SuccessRate:   66.667,
	})

	if got.AverageRttMs == nil || *got.AverageRttMs != 20 {
		t.Fatalf("expected average RTT from rollup summary, got %#v", got.AverageRttMs)
	}
	if got.MaxRttMs == nil || *got.MaxRttMs != 50 {
		t.Fatalf("expected max RTT from rollup summary, got %#v", got.MaxRttMs)
	}
	if got.LossPercent == nil || *got.LossPercent != 8.333 {
		t.Fatalf("expected loss percent from rollup summary, got %#v", got.LossPercent)
	}
	if got.SuccessRate == nil || *got.SuccessRate != 66.667 {
		t.Fatalf("expected success rate from rollup summary, got %#v", got.SuccessRate)
	}
	if got.Samples != 11 {
		t.Fatalf("expected samples from received count, got %d", got.Samples)
	}
}

func TestPingRollupInsightSummaryOmitsOptionalValuesWhenCountsMissing(t *testing.T) {
	got := pingRollupInsightSummary(sqlc.GetPingInsightRollupSummaryRow{
		Samples: 12,
	})

	if got.AverageRttMs != nil || got.MaxRttMs != nil || got.LossPercent != nil || got.SuccessRate != nil {
		t.Fatalf("expected optional aggregate values to be omitted, got %#v", got)
	}
	if got.Samples != 12 {
		t.Fatalf("expected samples to remain mapped, got %d", got.Samples)
	}
}
