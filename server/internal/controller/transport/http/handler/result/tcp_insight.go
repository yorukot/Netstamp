package result

import appresult "github.com/yorukot/netstamp/internal/controller/application/result"

type tcpInsightSummaryBody struct {
	AverageConnectMs *float64 `json:"averageConnectMs,omitempty"`
	MaxConnectMs     *float64 `json:"maxConnectMs,omitempty"`
	FailurePercent   *float64 `json:"failurePercent,omitempty"`
	SuccessRate      *float64 `json:"successRate,omitempty"`
	Samples          int64    `json:"samples"`
}

func newQueryTCPInsightBody(output appresult.TCPInsightOutput) queryTCPInsightBody {
	return newQueryInsightBody(newTCPInsightSummaryBody(output.Summary), output.Meta)
}

func newTCPInsightSummaryBody(summary appresult.TCPInsightSummary) tcpInsightSummaryBody {
	return tcpInsightSummaryBody{
		AverageConnectMs: summary.AverageConnectMs,
		MaxConnectMs:     summary.MaxConnectMs,
		FailurePercent:   summary.FailurePercent,
		SuccessRate:      summary.SuccessRate,
		Samples:          summary.Samples,
	}
}
