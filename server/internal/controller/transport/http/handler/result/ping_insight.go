package result

import appresult "github.com/yorukot/netstamp/internal/controller/application/result"

type pingInsightSummaryBody struct {
	AverageRttMs *float64 `json:"averageRttMs,omitempty"`
	MaxRttMs     *float64 `json:"maxRttMs,omitempty"`
	LossPercent  *float64 `json:"lossPercent,omitempty"`
	SuccessRate  *float64 `json:"successRate,omitempty"`
	Samples      int64    `json:"samples"`
}

func newQueryPingInsightBody(output appresult.PingInsightOutput) queryPingInsightBody {
	return newQueryInsightBody(newPingInsightSummaryBody(output.Summary), output.Meta)
}

func newPingInsightSummaryBody(summary appresult.PingInsightSummary) pingInsightSummaryBody {
	return pingInsightSummaryBody{
		AverageRttMs: summary.AverageRttMs,
		MaxRttMs:     summary.MaxRttMs,
		LossPercent:  summary.LossPercent,
		SuccessRate:  summary.SuccessRate,
		Samples:      summary.Samples,
	}
}
