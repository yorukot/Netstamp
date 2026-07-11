package result

import appresult "github.com/yorukot/netstamp/internal/controller/application/result"

type httpInsightSummaryBody struct {
	AverageTotalMs           *float64 `json:"averageTotalMs,omitempty"`
	MaxTotalMs               *float64 `json:"maxTotalMs,omitempty"`
	AverageTTFBMs            *float64 `json:"averageTtfbMs,omitempty"`
	MaxTTFBMs                *float64 `json:"maxTtfbMs,omitempty"`
	FailurePercent           *float64 `json:"failurePercent,omitempty"`
	SuccessRate              *float64 `json:"successRate,omitempty"`
	CertificateDaysRemaining *float64 `json:"certificateDaysRemaining,omitempty"`
	Samples                  int64    `json:"samples"`
}

func newQueryHTTPInsightBody(output appresult.HTTPInsightOutput) queryHTTPInsightBody {
	s := output.Summary
	return newQueryInsightBody(httpInsightSummaryBody{AverageTotalMs: s.AverageTotalMs, MaxTotalMs: s.MaxTotalMs, AverageTTFBMs: s.AverageTTFBMs, MaxTTFBMs: s.MaxTTFBMs, FailurePercent: s.FailurePercent, SuccessRate: s.SuccessRate, CertificateDaysRemaining: s.CertificateDaysRemaining, Samples: s.Samples}, output.Meta)
}
