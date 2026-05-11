package reporter

import (
	"time"

	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

const MaxBatchSize = 500

type Reporter struct{}

func New() *Reporter {
	return &Reporter{}
}

func (r *Reporter) Batches(probeID string, results []domainprobe.Result) []domainprobe.ResultBatch {
	if len(results) == 0 {
		return nil
	}

	batches := make([]domainprobe.ResultBatch, 0, (len(results)+MaxBatchSize-1)/MaxBatchSize)
	for start := 0; start < len(results); start += MaxBatchSize {
		end := start + MaxBatchSize
		if end > len(results) {
			end = len(results)
		}
		batches = append(batches, domainprobe.ResultBatch{
			ProbeID:     probeID,
			CollectedAt: time.Now().UTC(),
			Results:     append([]domainprobe.Result(nil), results[start:end]...),
		})
	}

	return batches
}
