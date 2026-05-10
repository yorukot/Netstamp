package reporter

import (
	"time"

	"github.com/yorukot/netstamp/internal/contracts/probecontrol"
)

type Reporter struct{}

func New() *Reporter {
	return &Reporter{}
}

func (r *Reporter) Batch(probeID string, results []probecontrol.Result) probecontrol.ResultBatch {
	return probecontrol.ResultBatch{
		ProbeID:     probeID,
		CollectedAt: time.Now().UTC(),
		Results:     results,
	}
}
