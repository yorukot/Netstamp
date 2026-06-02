package pingquery

import (
	"testing"

	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

func TestSelectReadPlan(t *testing.T) {
	tests := []struct {
		name          string
		counts        domainping.SeriesPointCounts
		maxDataPoints int32
		want          domainping.SeriesReadPlan
	}{
		{
			name:          "raw fits requested density",
			counts:        domainping.SeriesPointCounts{Raw: 3, Rollup: 3},
			maxDataPoints: 3,
			want: domainping.SeriesReadPlan{
				Mode:        domainping.SeriesReadModeRaw,
				Source:      domainping.SeriesSourceRaw,
				Resolution:  domainping.SeriesResolutionRaw,
				TotalPoints: 3,
			},
		},
		{
			name:          "bucket raw points above requested density",
			counts:        domainping.SeriesPointCounts{Raw: 4, Rollup: 0},
			maxDataPoints: 3,
			want: domainping.SeriesReadPlan{
				Mode:        domainping.SeriesReadModeBucket,
				Source:      domainping.SeriesSourceRaw,
				Resolution:  domainping.SeriesResolutionBucket,
				TotalPoints: 4,
			},
		},
		{
			name:          "rollup covers retained historical data",
			counts:        domainping.SeriesPointCounts{Raw: 2, Rollup: 100},
			maxDataPoints: 50,
			want: domainping.SeriesReadPlan{
				Mode:        domainping.SeriesReadModeRollup,
				Source:      domainping.SeriesSourceAggregate,
				Resolution:  domainping.SeriesResolutionOneMinute,
				TotalPoints: 100,
			},
		},
		{
			name:          "raw and rollup equal prefers raw",
			counts:        domainping.SeriesPointCounts{Raw: 3, Rollup: 3},
			maxDataPoints: 10,
			want: domainping.SeriesReadPlan{
				Mode:        domainping.SeriesReadModeRaw,
				Source:      domainping.SeriesSourceRaw,
				Resolution:  domainping.SeriesResolutionRaw,
				TotalPoints: 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SelectReadPlan(tt.counts, tt.maxDataPoints); got != tt.want {
				t.Fatalf("unexpected ping read plan: got %#v want %#v", got, tt.want)
			}
		})
	}
}
