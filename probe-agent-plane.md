# Probe Agent Runtime Plan

## 1. Purpose

This document describes the Netstamp probe agent runtime plan.

The focus is the probe process that runs in a user environment, pulls active assignments from the controller, schedules local measurement occurrences, executes checks with bounded concurrency, and submits typed results back to the controller.

This document intentionally keeps the controller/control-plane section short. The current backend already owns probe registry, credentials, assignment materialization, status persistence, and typed ping result storage. The probe agent should use the existing runtime API and should not introduce a new server-side task queue, assignment lease model, generic result table, or controller-side scheduler.

Core direction:

```text
Scheduling can be dense.
Same-assignment overlap is allowed.
Multiple ping jobs on the same probe are allowed.
Execution is bounded by NETSTAMP_PROBE_MAX_WORKERS.
```

## 2. Current Controller Contract

Runtime routes are registered under the API version prefix:

```text
/api/v1/runtime/probes/{probe_id}/hello
/api/v1/runtime/probes/{probe_id}/heartbeat
/api/v1/runtime/probes/{probe_id}/assignments
/api/v1/runtime/probes/{probe_id}/results
```

All probe runtime requests use probe-secret authentication:

```http
Authorization: Probe <secret>
```

The controller identifies the probe from the path `{probe_id}` and verifies the plaintext secret against the active probe credential hash. The probe must not send probe identity in request bodies, and the controller must not trust request body identity if it appears.

### 2.1 Hello

```http
POST /runtime/probes/{probe_id}/hello
Authorization: Probe <secret>
```

Request body:

```text
empty
```

Response body:

```json
{
	"serverTime": "2026-05-13T10:00:01Z",
	"minimumSupportedAgentVersion": "0.1.0",
	"config": {
		"heartbeatIntervalSeconds": 30,
		"assignmentPollIntervalSeconds": 30,
		"maxConcurrentWorkers": 16,
		"initialBackoffSeconds": 1,
		"maxBackoffSeconds": 30,
		"maxAttempts": 5
	}
}
```

Probe behavior:

- `hello` is the first runtime request.
- `hello` must succeed before heartbeat, assignment polling, scheduling, workers, or result submission start.
- The probe validates `minimumSupportedAgentVersion` before entering runtime.
- The probe applies `config` for control-loop intervals and result retry policy.
- `hello` does not send status, inventory, worker counters, capabilities, `boot_id`, or runtime session metadata in this version.

### 2.2 Heartbeat

```http
POST /runtime/probes/{probe_id}/heartbeat
Authorization: Probe <secret>
Content-Type: application/json
```

Request body:

```json
{
	"agentVersion": "netstamp-probe/0.1.0",
	"publicV4": "203.0.113.10",
	"publicV6": "2001:db8::10",
	"as": "AS15169 Google LLC",
	"addrs": ["10.0.0.10", "fd00::10"]
}
```

Response body:

```json
{
	"serverTime": "2026-05-13T10:00:31Z"
}
```

Heartbeat remains lightweight in this version. Worker health counters, assignment cache status, queue depth, and runtime session IDs are local observability only, not heartbeat API fields.

### 2.3 Assignments

```http
GET /runtime/probes/{probe_id}/assignments
Authorization: Probe <secret>
```

Response body:

```json
{
	"serverTime": "2026-05-13T10:00:40Z",
	"config": {
		"heartbeatIntervalSeconds": 30,
		"assignmentPollIntervalSeconds": 30,
		"maxConcurrentWorkers": 16,
		"initialBackoffSeconds": 1,
		"maxBackoffSeconds": 30,
		"maxAttempts": 5
	},
	"assignments": [
		{
			"id": "f67a3d88-3b2a-4f65-ae5f-4e8c9b3e5971",
			"projectId": "0fd60e87-f2a2-4ca9-84c7-0ce91f6a8ae2",
			"probeId": "a95f6e58-6c2d-4e90-9a54-b7aa7c99e4c0",
			"checkId": "c7a0aa56-7729-4140-a3ef-9fe3d6a7cf5f",
			"checkVersion": "v1",
			"selectorVersion": "v1",
			"check": {
				"id": "c7a0aa56-7729-4140-a3ef-9fe3d6a7cf5f",
				"projectId": "0fd60e87-f2a2-4ca9-84c7-0ce91f6a8ae2",
				"name": "example ping",
				"type": "ping",
				"target": "example.com",
				"selector": {},
				"intervalSeconds": 30,
				"pingConfig": {
					"packetCount": 4,
					"packetSizeBytes": 56,
					"timeoutMs": 3000,
					"ipFamily": null
				},
				"labels": []
			}
		}
	]
}
```

Assignment response semantics:

- The response is a full active assignment snapshot.
- It is not a server-side lease queue.
- It does not include controller-computed phase, due time, lease state, running state, or execution token.
- The probe calculates local phase and next due time.
- `assignments.config` has the same shape as `hello.config` and refreshes control-loop timing and result retry policy.

### 2.4 Results

```http
POST /runtime/probes/{probe_id}/results
Authorization: Probe <secret>
Content-Type: application/json
```

Request body:

```json
{
	"results": [
		{
			"checkId": "44444444-4444-4444-4444-444444444444",
			"type": "ping",
			"ping": [
				{
					"startedAt": "2026-05-13T10:00:00Z",
					"finishedAt": "2026-05-13T10:00:01Z",
					"durationMs": 1000,
					"status": "successful",
					"sentCount": 4,
					"receivedCount": 4,
					"lossPercent": 0,
					"rttMinMs": 10.1,
					"rttAvgMs": 12.3,
					"rttMedianMs": 12.0,
					"rttMaxMs": 15.6,
					"rttStddevMs": 1.7,
					"rttSamplesMs": [10.1, 11.5, 12.0, 15.6],
					"resolvedIp": "1.1.1.1",
					"ipFamily": "inet",
					"raw": {}
				}
			]
		}
	]
}
```

Response body:

```json
{
	"accepted": 1,
	"serverTime": "2026-05-13T10:00:02Z"
}
```

Result submission semantics:

- Payload is grouped by `checkId` and `type`.
- Current supported type is `ping`.
- The probe does not send `projectId`, `probeId`, or `assignmentId`.
- The controller verifies the check is still an active assignment for the authenticated probe.
- The controller verifies result `type` matches assignment check type.
- Current storage idempotency is based on `(project_id, probe_id, check_id, started_at)`.

## 3. Probe Runtime Principles

The probe runtime follows these rules:

- Probe is pull-based.
- Controller tells probe which active checks exist.
- Probe computes local schedule.
- Probe may overlap occurrences of the same assignment.
- Probe may run multiple ping jobs concurrently.
- Future HTTP, DNS, TCP, and ping jobs share the same worker pool.
- `NETSTAMP_PROBE_MAX_WORKERS` is the only hard measurement concurrency limit.
- Do not add `NETSTAMP_PROBE_MAX_PING_CONCURRENCY`.
- Runtime config from controller controls loop intervals and retry policy, not local worker capacity.
- Scheduler creates measurement occurrences.
- Worker pool executes measurement occurrences.
- Result submitter batches completed results.
- Missed occurrences are not backfilled.
- Worker queue full means skip the occurrence.
- Scheduler must not block waiting for worker capacity.

Responsibility split:

| Component         | Responsibility                                        |
| ----------------- | ----------------------------------------------------- |
| Runtime client    | Calls current controller runtime API.                 |
| Hello manager     | Performs startup handshake and version check.         |
| Heartbeat loop    | Sends lightweight probe status.                       |
| Assignment puller | Pulls full active assignment snapshot.                |
| Reconciler        | Converts remote assignment snapshot into local tasks. |
| Scheduler         | Computes due occurrences from interval and phase.     |
| Worker queue      | Buffers a small number of pending occurrences.        |
| Worker pool       | Enforces global measurement concurrency.              |
| Ping executor     | Runs a ping measurement for one occurrence.           |
| Result queue      | Buffers completed measurement results.                |
| Result submitter  | Groups and submits results.                           |

## 4. Local Configuration

Required environment variables:

```text
NETSTAMP_PROBE_CONTROLLER_URL
NETSTAMP_PROBE_ID
NETSTAMP_PROBE_SECRET
```

Optional environment variables:

```text
NETSTAMP_PROBE_HTTP_TIMEOUT
NETSTAMP_PROBE_MAX_WORKERS
NETSTAMP_PROBE_RESULT_QUEUE_SIZE
NETSTAMP_PROBE_RESULT_BATCH_SIZE
NETSTAMP_PROBE_RESULT_FLUSH_INTERVAL
NETSTAMP_PROBE_ASSIGNMENT_TTL
NETSTAMP_PROBE_SHUTDOWN_TIMEOUT
NETSTAMP_PROBE_HEARTBEAT_INTERVAL
NETSTAMP_PROBE_ASSIGNMENT_POLL_INTERVAL
NETSTAMP_PROBE_INITIAL_BACKOFF
NETSTAMP_PROBE_MAX_BACKOFF
NETSTAMP_PROBE_MAX_ATTEMPTS
```

Recommended defaults:

```text
NETSTAMP_PROBE_HTTP_TIMEOUT=10s
NETSTAMP_PROBE_MAX_WORKERS=4
NETSTAMP_PROBE_RESULT_QUEUE_SIZE=10000
NETSTAMP_PROBE_RESULT_BATCH_SIZE=100
NETSTAMP_PROBE_RESULT_FLUSH_INTERVAL=5s
NETSTAMP_PROBE_ASSIGNMENT_TTL=10m
NETSTAMP_PROBE_SHUTDOWN_TIMEOUT=10s
```

Runtime config override defaults, when not set, come from the controller runtime config after `hello` succeeds:

```text
NETSTAMP_PROBE_HEARTBEAT_INTERVAL=30s
NETSTAMP_PROBE_ASSIGNMENT_POLL_INTERVAL=30s
NETSTAMP_PROBE_INITIAL_BACKOFF=1s
NETSTAMP_PROBE_MAX_BACKOFF=30s
NETSTAMP_PROBE_MAX_ATTEMPTS=5
```

Derived values:

```text
worker_queue_capacity = max(1, NETSTAMP_PROBE_MAX_WORKERS * 2)
```

Precedence:

- `NETSTAMP_PROBE_MAX_WORKERS` wins for measurement concurrency.
- Runtime config env overrides win per field when explicitly set.
- Controller runtime config wins for heartbeat, assignment poll, and retry fields that were not explicitly set through env.
- Controller `config.maxConcurrentWorkers` is not a hard worker limit in this version.
- The probe may record `config.maxConcurrentWorkers` for diagnostics, but it should not resize the worker pool from it.
- The worker pool size is fixed after startup for this version.

Suggested local config type:

```go
type ProbeConfig struct {
	ControllerURL       string
	ProbeID             string
	ProbeSecret         string
	HTTPTimeout         time.Duration
	MaxWorkers          int
	ResultQueueSize     int
	ResultBatchSize     int
	ResultFlushInterval time.Duration
	AssignmentTTL       time.Duration
	ShutdownTimeout     time.Duration
	RuntimeOverrides    RuntimeConfigOverrides
}

type RuntimeConfigOverrides struct {
	HeartbeatInterval      ConfigValue[time.Duration]
	AssignmentPollInterval ConfigValue[time.Duration]
	InitialBackoff         ConfigValue[time.Duration]
	MaxBackoff             ConfigValue[time.Duration]
	MaxAttempts            ConfigValue[int]
}

type ConfigValue[T any] struct {
	Value   T
	Defined bool
}
```

Suggested runtime config type after applying controller config:

```go
type RuntimeConfig struct {
	HeartbeatInterval      time.Duration
	AssignmentPollInterval time.Duration
	InitialBackoff         time.Duration
	MaxBackoff             time.Duration
	MaxAttempts            int
}
```

## 5. Lifecycle

Probe runtime lifecycle:

```text
STARTING
  -> HELLO
  -> BOOTSTRAPPING
  -> RUNNING
  -> DEGRADED
  -> AUTH_FAILED
  -> DRAINING
  -> STOPPED
```

### 5.1 STARTING

Startup work:

- Load local environment config.
- Validate controller URL.
- Validate probe ID.
- Validate probe secret is non-empty.
- If `NETSTAMP_PROBE_MAX_WORKERS` is set, validate it is greater than `0`.
- If `NETSTAMP_PROBE_MAX_WORKERS` is unset, use the default max workers value.
- Create runtime HTTP client.
- Create assignment store.
- Create scheduler.
- Create bounded worker queue.
- Create worker pool.
- Create bounded result queue.
- Create result submitter.

Measurements must not run in `STARTING`.

### 5.2 HELLO

First runtime request:

```http
POST /runtime/probes/{probe_id}/hello
```

Rules:

- Request body is empty.
- Success returns server time, minimum supported agent version, and runtime config.
- Probe must check version compatibility.
- Probe applies runtime config before starting loops.
- `401` or `403` enters `AUTH_FAILED`.
- Transient network or `5xx` failure should retry with backoff.

No heartbeat, assignment pull, scheduling, worker execution, or result submission starts before `hello` succeeds.

### 5.3 BOOTSTRAPPING

After successful `hello`:

- Start heartbeat loop.
- Start result submitter.
- Start worker pool.
- Pull assignments immediately.
- Reconcile first assignment snapshot.
- Start scheduler after first successful assignment pull.
- Enter `RUNNING`.

The scheduler may start with an empty assignment set if the first successful assignment response is empty.

### 5.4 RUNNING

Runtime loops:

```text
heartbeat loop
assignment pull loop
scheduler loop
worker pool
result submit loop
```

Example timing:

```text
heartbeat: every config.heartbeatIntervalSeconds
assignments: every config.assignmentPollIntervalSeconds
results: every NETSTAMP_PROBE_RESULT_FLUSH_INTERVAL or batch full
measurements: based on assignment check.intervalSeconds plus local phase
```

### 5.5 DEGRADED

The probe is degraded when part of the runtime is failing but it can still make useful progress.

Assignment pull failure:

- Heartbeat continues.
- Scheduler continues using cached assignments until assignment TTL expires.
- Result submission continues.
- Assignment pull keeps retrying.

Assignment TTL expired:

- Scheduler stops producing new occurrences.
- Worker pool finishes already queued or running jobs.
- Result submitter continues.
- Assignment pull keeps retrying.

Result submit failure:

- Workers continue.
- Results stay queued if capacity allows.
- Retries use runtime config backoff.
- Queue overflow drops oldest results.

Heartbeat failure:

- Scheduler and workers continue.
- Heartbeat retries.
- Local counters/logs record the failure.

### 5.6 AUTH_FAILED

Any runtime endpoint returning `401` or `403` means the probe should enter `AUTH_FAILED`.

Behavior:

- Stop assignment polling.
- Stop scheduler.
- Stop dispatching new jobs.
- Stop result submission after best-effort final handling.
- Exit non-zero or wait for operator restart.
- Do not keep measuring with invalid credentials or a disabled probe.

### 5.7 DRAINING

Triggered by SIGTERM or SIGINT.

Behavior:

- Stop assignment polling.
- Stop scheduler.
- Close worker queue.
- Do not create new `RunRequest` values.
- Wait for running workers until `NETSTAMP_PROBE_SHUTDOWN_TIMEOUT`.
- Flush result queue best-effort.
- Exit.

The current backend heartbeat API does not define a draining state, so draining is local runtime behavior only.

## 6. Runtime Components

### 6.1 Runtime Client

The runtime client owns:

- Base controller URL.
- API version prefix.
- Probe ID.
- Probe secret.
- HTTP timeout.
- `Authorization: Probe <secret>` header.
- JSON encoding and decoding.
- Error mapping for `401`, `403`, `404`, `422`, network errors, and `5xx`.

Runtime client methods:

```go
Hello(ctx context.Context) (HelloResponse, error)
Heartbeat(ctx context.Context, input HeartbeatInput) (HeartbeatResponse, error)
ListAssignments(ctx context.Context) (AssignmentsResponse, error)
SubmitResults(ctx context.Context, input SubmitResultsInput) (SubmitResultsResponse, error)
```

### 6.2 Assignment Store

The assignment store holds local scheduling state keyed by assignment ID.

Suggested task state:

```go
type TaskState struct {
	AssignmentID    string
	ProjectID       string
	ProbeID         string
	CheckID         string
	CheckType       string
	CheckVersion    string
	SelectorVersion string

	Interval time.Duration
	Phase    time.Duration
	NextDue  time.Time

	PingConfig *PingConfig
	HTTPConfig *HTTPConfig
	DNSConfig  *DNSConfig
	TCPConfig  *TCPConfig

	Enabled      bool
	Removed      bool
	Generation   uint64
	LastPulledAt time.Time
}
```

Important:

```text
TaskState does not contain Running bool.
```

The same assignment may have multiple overlapping occurrences. Running status is not assignment state.

Typed config rule:

- All check-type-specific config fields are nullable.
- `PingConfig` is set only for ping assignments.
- Future check types should add their own nullable typed config fields, such as `HTTPConfig`, `DNSConfig`, or `TCPConfig`.
- A task must have the typed config required by its `CheckType`; otherwise the assignment is invalid for local execution and should be skipped.
- Probe-side typed config models should be reused from the domain layer whenever possible. Do not create parallel probe-only models when an existing domain model already represents the same data cleanly.

### 6.3 RunRequest

The scheduler dispatches occurrence snapshots, not mutable task pointers.

Suggested run request:

```go
type RunRequest struct {
	AssignmentID    string
	ProjectID       string
	ProbeID         string
	CheckID         string
	CheckType       string
	CheckVersion    string
	SelectorVersion string

	ScheduledAt time.Time
	CreatedAt   time.Time

	PingConfig *PingConfig
	HTTPConfig *HTTPConfig
	DNSConfig  *DNSConfig
	TCPConfig  *TCPConfig
}
```

Snapshot rule:

- A queued or running `RunRequest` uses the assignment config from the time it was created.
- If the assignment changes later, future occurrences use new config.
- Already queued or running occurrences are not canceled just because assignment config changed.
- `RunRequest` carries nullable typed config snapshots for all supported check types, and exactly one typed config should be non-nil for a valid measurement occurrence.

### 6.4 Result Envelope

Current backend result submission does not accept `assignmentId`, `scheduledAt`, or `actualStartedAt`.

Suggested local envelope:

```go
type PingResultEnvelope struct {
	CheckID    string
	Type       string
	StartedAt  time.Time
	FinishedAt time.Time
	DurationMs int32
	Ping       PingResult
}
```

Mapping rule:

```text
backend startedAt = RunRequest.ScheduledAt
backend finishedAt = actual finish time
```

This preserves controller idempotency for overlapping occurrences because each scheduled occurrence has a distinct `startedAt`.

The implementation may record actual worker start time in local logs or metrics, but it is not sent to the current backend API.

## 7. Assignment Reconcile

Assignment polling returns a full active assignment snapshot.

Reconcile process:

```text
remote snapshot
  -> remote map by assignment.id
  -> compare with local task map
  -> add new tasks
  -> update changed tasks
  -> mark missing tasks removed
  -> wake scheduler
```

### 7.1 Add

For a new assignment:

- Validate assignment has a supported check type.
- Validate check config for that type exists.
- Convert check interval to `time.Duration`.
- Compute deterministic phase.
- Compute next future due time.
- Insert task into assignment store.
- Push scheduler heap item.
- Wake scheduler.

### 7.2 Update

Update local task if any of these change:

- `checkVersion`
- `selectorVersion`
- check type
- target
- interval
- ping config

If interval changes:

- Recompute phase.
- Recompute next due.
- Push a new scheduler heap item.
- Wake scheduler.

Already queued or running occurrences keep their old `RunRequest` snapshot.

### 7.3 Remove

If a local assignment no longer appears in the remote snapshot:

- Mark local task as removed.
- Wake scheduler.
- Future occurrences must not be generated.
- Already queued or running occurrences may finish.
- The controller decides whether submitted results are still accepted.

### 7.4 Cached Assignments And TTL

Track:

```text
last_successful_assignment_pull_at
```

If assignment pull fails:

- Keep cached assignments until TTL expires.
- Keep scheduling cached tasks.
- Keep retrying assignment pull.

If TTL expires:

- Stop producing new occurrences.
- Keep heartbeat loop running.
- Keep result submitter running.
- Let already running workers finish.

## 8. Bounded-Overlap Scheduler

The scheduler is a min-heap scheduler keyed by `NextDue`.

The scheduler does not prevent same-assignment overlap.

Each due time creates one independent occurrence:

```text
10:00:01 assignment A occurrence #1
10:00:02 assignment A occurrence #2
10:00:03 assignment A occurrence #3
```

If worker pool capacity exists, these may run concurrently.

### 8.1 Phase

Phase spreads load. It is not required for correctness.

Calculation:

```text
phase = hash(probe_id + ":" + assignment_id) % interval
```

Rules:

- Same probe ID and assignment ID produce the same phase after restart.
- Different assignments are likely to spread across the interval.
- Phase is less than interval.
- If interval is `1s`, phase is `0`.
- Controller does not store or return phase.
- Recompute phase when interval changes.

Suggested implementation:

```go
func computePhase(probeID, assignmentID string, interval time.Duration) time.Duration {
	if interval <= time.Second {
		return 0
	}

	seconds := uint64(interval / time.Second)
	hash := fnv64(probeID + ":" + assignmentID)

	return time.Duration(hash%seconds) * time.Second
}
```

### 8.2 Next Due

Initial next due:

```go
func computeNextDue(now time.Time, interval, phase time.Duration) time.Time {
	base := now.Truncate(interval)
	due := base.Add(phase)
	if !due.After(now) {
		due = due.Add(interval)
	}
	return due
}
```

After dispatch or skip:

```go
func computeNextFutureDue(previousDue, now time.Time, interval time.Duration) time.Time {
	next := previousDue.Add(interval)
	for !next.After(now) {
		next = next.Add(interval)
	}
	return next
}
```

No catch-up rule:

```text
If the scheduler was delayed by 10 seconds for a 1 second interval,
do not enqueue 10 missed occurrences.
Compute the next future occurrence instead.
```

### 8.3 Scheduler Loop

Conceptual loop:

```text
loop:
  task = heap.peek()

  if task is nil:
    wait for assignment update or shutdown
    continue

  if task.NextDue is in the future:
    wait until timer, assignment update, or shutdown
    continue

  dueTasks = pop all due tasks

  for each task:
    validate task is current and active
    create RunRequest snapshot
    non-blocking enqueue to worker queue
    skip if worker queue is full
    compute next future due
    push updated task
```

Scheduler must be wakeable by assignment reconcile. It must not sleep uninterruptibly until a stale due time.

### 8.4 Currentness Check

Heap items may be stale because updates push new entries instead of removing old entries from the middle of the heap.

When popping a task:

- Look up current task by assignment ID.
- Confirm generation/version matches.
- Skip if removed.
- Skip if disabled.
- Skip if stale heap item.

This avoids complex heap deletion logic.

### 8.5 Late Occurrences

Late skip policy:

```go
func isTooLate(scheduledAt, now time.Time, interval time.Duration) bool {
	return now.Sub(scheduledAt) > interval
}
```

If too late:

- Do not enqueue.
- Increment `skipped_late`.
- Compute next future due.
- Push task back if still active.

## 9. Worker Pool And Backpressure

Worker pool semantics:

```text
A worker is one concurrent measurement execution slot.
```

One worker slot is used by:

- one ping occurrence
- one future HTTP occurrence
- one future DNS occurrence
- one future TCP occurrence

Worker slots are not used by:

- hello request
- heartbeat request
- assignment pull request
- result submit request

### 9.1 Worker Pool Size

Worker pool size:

```text
max_workers = NETSTAMP_PROBE_MAX_WORKERS
```

Initialization:

```go
workerQueue := make(chan RunRequest, max(1, maxWorkers*2))

for i := 0; i < maxWorkers; i++ {
	go worker.Run(ctx, workerQueue)
}
```

### 9.2 Worker Queue

Worker queue capacity:

```text
worker_queue_capacity = max(1, NETSTAMP_PROBE_MAX_WORKERS * 2)
```

The queue absorbs short bursts when multiple assignments are due at the same time.

The queue must not become unbounded.

### 9.3 Queue Full

Scheduler dispatch is non-blocking:

```go
select {
case workerQueue <- req:
	counters.ScheduledRuns.Add(1)
default:
	counters.SkippedWorkerQueueFull.Add(1)
}
```

Queue full behavior:

- Skip this occurrence.
- Do not block scheduler.
- Do not wait for a worker.
- Do not retry the same occurrence later.
- Compute next future due.

Reasoning:

```text
Measurements are time-sensitive.
If an occurrence cannot run near its scheduled time, it should be skipped
instead of being executed much later as stale data.
```

### 9.4 Worker Loop

Worker loop:

```go
func (w *Worker) Run(ctx context.Context, q <-chan RunRequest) {
	for {
		select {
		case <-ctx.Done():
			return
		case req, ok := <-q:
			if !ok {
				return
			}
			w.runOne(ctx, req)
		}
	}
}
```

Each worker handles exactly one measurement job at a time.

## 10. Ping Measurement Execution

Initial probe agent implementation should support ping checks only.

Ping input comes from assignment `check.pingConfig`:

- target
- packet count
- packet size bytes
- timeout ms
- IP family

Ping execution responsibilities:

- Resolve hostname targets.
- Respect IPv4 or IPv6 preference when provided.
- Send configured packet count.
- Respect packet size.
- Respect timeout.
- Collect RTT samples.
- Compute sent count.
- Compute received count.
- Compute loss percent.
- Compute RTT min, average, median, max, and stddev.
- Capture resolved IP when available.
- Capture error code and message for timeout or executor failures.

Statuses:

```text
successful
timeout
error
```

Timing:

- `RunRequest.ScheduledAt` becomes backend `startedAt`.
- Actual finish time becomes backend `finishedAt`.
- `durationMs` should be `finishedAt - startedAt`.

This makes overlapping occurrences of the same assignment distinct for backend idempotency.

## 11. Result Queue And Submission

Workers do not submit directly to the controller.

Flow:

```text
worker
  -> result queue
  -> result submitter
  -> POST /runtime/probes/{probe_id}/results
```

### 11.1 Result Queue

Result queue:

- Bounded.
- Default capacity is `10000`.
- Completed worker results are enqueued.
- Queue overflow drops the oldest result.
- Overflow increments `dropped_results`.

Drop-oldest behavior keeps newer measurement data when the probe is under pressure.

### 11.2 Result Submitter

Submitter flush conditions:

- Flush every `NETSTAMP_PROBE_RESULT_FLUSH_INTERVAL`.
- Flush early when batch reaches `NETSTAMP_PROBE_RESULT_BATCH_SIZE`.
- Flush best-effort on shutdown.

Batching:

- Group by `checkId`.
- Group by `type`.
- Put ping payloads in `results[].ping`.

Retry:

- Use controller runtime config:
  - `initialBackoffSeconds`
  - `maxBackoffSeconds`
  - `maxAttempts`
- Network errors and `5xx` are retryable.
- `401` or `403` enters `AUTH_FAILED`.
- `422` means the controller rejected the batch; drop it and record an error.

## 12. Failure Modes

### 12.1 Assignment Pull Failed

Behavior:

- Increment `assignment_pull_errors`.
- Keep cached assignments if TTL has not expired.
- Keep scheduling from cache.
- Keep retrying assignment pull.

### 12.2 Assignment TTL Expired

Behavior:

- Stop scheduling new occurrences.
- Continue heartbeat.
- Continue result submission.
- Let already queued or running workers finish.
- Keep retrying assignment pull.

### 12.3 Heartbeat Failed

Behavior:

- Increment `heartbeat_errors`.
- Continue scheduler.
- Continue workers.
- Continue assignment polling.
- Retry heartbeat.

### 12.4 Result Submit Failed

Behavior:

- Increment `result_submit_errors`.
- Keep results queued if capacity allows.
- Retry according to runtime config.
- Continue workers.

### 12.5 Worker Queue Full

Behavior:

- Skip occurrence.
- Increment `skipped_worker_queue_full`.
- Do not block scheduler.
- Do not catch up later.

### 12.6 Late Occurrence

Behavior:

- Skip occurrence.
- Increment `skipped_late`.
- Compute next future due.

### 12.7 Result Queue Full

Behavior:

- Drop oldest result.
- Increment `dropped_results`.
- Keep accepting new worker results.

### 12.8 Unsupported Check Type

Behavior:

- Ignore unsupported assignment locally.
- Log sanitized warning.
- Continue other assignments.
- Do not fail whole probe runtime.

### 12.9 Auth Failed

Behavior:

- Enter `AUTH_FAILED`.
- Stop measuring.
- Stop polling assignments.
- Stop scheduler.
- Exit non-zero or wait for operator restart.

## 13. Shutdown

Shutdown starts on SIGTERM or SIGINT.

Sequence:

```text
enter DRAINING
stop assignment puller
stop scheduler
close worker queue
wait for active workers until shutdown timeout
flush result queue best-effort
exit
```

Rules:

- No new measurements after draining starts.
- Running measurements can complete until shutdown timeout.
- Result flush is best-effort.
- Current backend does not have a draining heartbeat state.

## 14. Observability

This version keeps worker health and scheduler counters local. They are not sent in heartbeat request bodies.

### 14.1 Logs

Probe agent logging should use Go standard structured logging:

```go
log/slog
```

Use structured attributes for stable runtime dimensions such as probe ID, assignment ID, check ID, check type, state, duration, queue depth, and counter values. Do not use ad-hoc formatted strings for fields that operators may need to filter.

Log:

- startup config summary without secret
- hello success/failure
- version incompatibility
- heartbeat failure
- assignment pull failure
- assignment reconcile summary
- skipped worker queue full
- skipped late occurrences
- worker execution errors
- result submit failures
- auth failure
- shutdown drain summary

Debug log:

- runtime config applied from `hello`
- runtime config refreshed from assignments
- assignment reconcile add/update/remove details
- scheduler due-time decisions
- computed assignment phase
- skipped occurrence details
- worker queue depth transitions
- result batch grouping and flush decisions
- retry/backoff decisions

Do not log:

- plaintext probe secret
- secret hash
- raw result payloads by default
- high-cardinality check targets in high-volume logs

Default log level should be info. Debug logs should be opt-in through local configuration, for example `NETSTAMP_PROBE_LOG_LEVEL=debug`.

### 14.2 Counters

Suggested counters:

```go
type RuntimeCounters struct {
	ScheduledRuns          uint64
	CompletedRuns          uint64
	SkippedWorkerQueueFull uint64
	SkippedLate            uint64
	DroppedResults         uint64
	AssignmentPullErrors   uint64
	ResultSubmitErrors     uint64
	HeartbeatErrors        uint64
	AuthFailures           uint64
}
```

### 14.3 Gauges

Suggested gauges:

- active workers
- max workers
- worker queue depth
- worker queue capacity
- result queue depth
- assignment count
- seconds since last successful assignment pull
- seconds since last successful result submit

## 15. Implementation Phases

This section is an implementation sequence for the future probe agent.

### Phase 1: Config And Runtime Client

Implement:

- env loading
- config validation
- default values
- runtime HTTP client
- auth header
- hello request and response
- heartbeat request
- assignment response DTOs
- result submit DTOs

Acceptance:

- missing required env fails
- invalid env fails
- hello sends empty body
- runtime client uses `Authorization: Probe <secret>`

### Phase 2: Assignment Pull And Reconcile

Implement:

- assignment polling loop
- full snapshot parsing
- local task map
- add/update/remove reconcile
- assignment TTL
- scheduler wake on changes

Acceptance:

- new assignment creates task
- updated assignment refreshes task
- removed assignment stops future scheduling
- cached assignments continue until TTL

### Phase 3: Scheduler

Implement:

- deterministic phase
- min-heap scheduler
- no overlap prevention
- no catch-up burst
- late skip
- non-blocking worker queue dispatch

Acceptance:

- same assignment can create overlapping run requests
- scheduler skips when worker queue is full
- scheduler skips late occurrences
- scheduler does not enqueue missed backlog

### Phase 4: Worker Pool

Implement:

- `NETSTAMP_PROBE_MAX_WORKERS`
- bounded worker queue
- worker goroutines
- active worker accounting
- queue depth metrics

Acceptance:

- active workers never exceed max workers
- multiple ping jobs can run concurrently
- no ping-specific concurrency setting exists

### Phase 5: Ping Executor

Implement:

- DNS resolution
- IPv4 and IPv6 handling
- packet count
- packet size
- timeout
- RTT samples
- loss percentage
- error mapping

Acceptance:

- successful ping maps result fields
- timeout maps to `timeout`
- resolver failure maps to `error`
- context cancellation stops execution

### Phase 6: Result Queue And Submitter

Implement:

- bounded result queue
- drop-oldest overflow
- batch grouping
- flush interval
- batch size flush
- retry and backoff

Acceptance:

- worker result enters queue
- submitter sends grouped payload
- transient failures retry
- auth failure stops runtime
- queue overflow increments dropped counter

### Phase 7: Runtime Failure And Shutdown

Implement:

- degraded mode
- assignment TTL expiry behavior
- auth failed behavior
- draining behavior
- best-effort result flush

Acceptance:

- assignment pull failure keeps cached tasks until TTL
- TTL expiry stops new occurrences
- SIGTERM stops scheduler and drains workers
- runtime exits cleanly

## 16. Test Plan

### 16.1 Config Tests

- Missing `NETSTAMP_PROBE_CONTROLLER_URL` fails.
- Missing `NETSTAMP_PROBE_ID` fails.
- Missing `NETSTAMP_PROBE_SECRET` fails.
- Invalid controller URL fails.
- Invalid probe ID fails.
- If set, `NETSTAMP_PROBE_MAX_WORKERS <= 0` fails.
- If unset, the default max workers value is used.
- Defaults are applied.
- No `NETSTAMP_PROBE_MAX_PING_CONCURRENCY` exists.

### 16.2 Runtime Client Tests

- `hello` sends empty body.
- Runtime auth header is `Authorization: Probe <secret>`.
- Assignments parse `serverTime`, `config`, and assignment list.
- Results are grouped by `checkId` and `type`.
- `401` and `403` map to auth failed.

### 16.3 Phase Tests

- Same probe ID and assignment ID produce same phase.
- Different assignment IDs likely produce different phases.
- Phase is less than interval.
- `interval=1s` gives zero phase.
- Interval change recalculates phase.

### 16.4 Reconcile Tests

- New assignment creates task.
- Changed check version updates task.
- Changed selector version updates task.
- Changed ping config updates task.
- Removed assignment marks task removed.
- Queued or running request keeps old snapshot.
- TTL expiry stops new scheduling.

### 16.5 Scheduler Tests

- Same assignment can overlap.
- Scheduler does not use `Running`.
- Global active workers never exceed max workers.
- Worker queue full increments skip counter.
- Late occurrence increments skip counter.
- Scheduler does not catch up missed occurrences.
- Assignment update wakes scheduler.

### 16.6 Worker Pool Tests

- Starts exactly max worker count.
- Each job occupies one worker slot.
- Multiple ping jobs can run concurrently.
- Control-plane requests do not consume worker slots.
- Context cancellation stops workers.

### 16.7 Ping Executor Tests

- Successful ping maps sent, received, loss, and RTT fields.
- Timeout maps to `timeout`.
- Resolver failure maps to `error`.
- IPv4 and IPv6 selection is respected.
- Context timeout stops execution.

### 16.8 Result Queue Tests

- Worker result enters queue.
- Batch full flush works.
- Timer flush works.
- Queue overflow drops oldest.
- Submit transient failure retries.
- Auth failure stops runtime.

### 16.9 Shutdown Tests

- SIGTERM stops scheduler.
- No new jobs after draining.
- Running workers get shutdown timeout.
- Result queue flushes best-effort.
- Runtime exits cleanly.

### 16.10 Documentation Acceptance Checks

- No section says the same assignment must not overlap.
- No section introduces `NETSTAMP_PROBE_MAX_PING_CONCURRENCY`.
- No section uses old `/probes/{probe_id}/runtime/*` endpoint order.
- No section documents snake_case runtime API fields as current backend contract.
- No section requires backend support for `boot_id`, `runtime_session_id`, `scheduledAt`, or `actualStartedAt`.
- Current backend API and future work are clearly separated.

## 17. Future Work

Future work that is intentionally not part of the current probe runtime contract:

- Rich heartbeat health payload with worker counters and queue depths.
- Runtime session IDs and boot IDs.
- Sending `scheduledAt` and `actualStartedAt` to the controller.
- HTTP, DNS, TCP, and traceroute typed executors.
- Persisted controller-managed runtime config.
- Controller-enforced `maxConcurrentWorkers` rollout policy.
- Server-side assignment leases.
- Server-side dispatch scheduler.
- Offline detector based on heartbeat age.
- Low-frequency inventory endpoint for OS, arch, interfaces, hardware, and DNS metadata.

## 18. Current Repo Notes

The backend runtime API exists under:

```text
server/internal/controller/transport/http/proberuntime
server/internal/controller/application/proberuntime
```

The current repository still contains references to a probe binary in `server/AGENTS.md`, `Justfile`, and `server/Dockerfile`, but the actual `server/cmd/probe` and `server/internal/probe` implementation should be treated as future implementation work unless present in the working tree.
