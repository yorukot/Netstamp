package observability

import "sync/atomic"

type RuntimeCounters struct {
	ScheduledRuns          atomic.Uint64
	CompletedRuns          atomic.Uint64
	SkippedWorkerQueueFull atomic.Uint64
	SkippedLate            atomic.Uint64
	DroppedResults         atomic.Uint64
	AssignmentPullErrors   atomic.Uint64
	ResultSubmitErrors     atomic.Uint64
	HeartbeatErrors        atomic.Uint64
	AuthFailures           atomic.Uint64
}
