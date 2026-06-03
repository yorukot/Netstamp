package tcpquery

import (
	"testing"
	"time"

	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
)

func TestSelectReadPlan(t *testing.T) {
	now := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		rawPoints     int64
		from          time.Time
		maxDataPoints int32
		want          domaintcp.SeriesReadPlan
	}{
		{
			name:          "raw fits requested density",
			rawPoints:     3,
			from:          now.Add(-time.Hour),
			maxDataPoints: 3,
			want: domaintcp.SeriesReadPlan{
				Mode:        domaintcp.SeriesReadModeRaw,
				Source:      domaintcp.SeriesSourceRaw,
				Resolution:  domaintcp.SeriesResolutionRaw,
				TotalPoints: 3,
			},
		},
		{
			name:          "bucket raw points above requested density",
			rawPoints:     4,
			from:          now.Add(-time.Hour),
			maxDataPoints: 3,
			want: domaintcp.SeriesReadPlan{
				Mode:        domaintcp.SeriesReadModeBucket,
				Source:      domaintcp.SeriesSourceRaw,
				Resolution:  domaintcp.SeriesResolutionBucket,
				TotalPoints: 4,
			},
		},
		{
			name:          "rollup when query starts before raw retention",
			rawPoints:     2,
			from:          now.Add(-RawRetentionWindow - time.Millisecond),
			maxDataPoints: 50,
			want: domaintcp.SeriesReadPlan{
				Mode:        domaintcp.SeriesReadModeRollup,
				Source:      domaintcp.SeriesSourceAggregate,
				Resolution:  domaintcp.SeriesResolutionOneMinute,
				TotalPoints: 2,
			},
		},
		{
			name:          "raw at exact retention boundary",
			rawPoints:     4,
			from:          now.Add(-RawRetentionWindow),
			maxDataPoints: 10,
			want: domaintcp.SeriesReadPlan{
				Mode:        domaintcp.SeriesReadModeRaw,
				Source:      domaintcp.SeriesSourceRaw,
				Resolution:  domaintcp.SeriesResolutionRaw,
				TotalPoints: 4,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SelectReadPlan(tt.rawPoints, tt.from, now, tt.maxDataPoints); got != tt.want {
				t.Fatalf("unexpected tcp read plan: got %#v want %#v", got, tt.want)
			}
		})
	}
}
