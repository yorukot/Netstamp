# Netstamp Alerting V1 Plan

## Summary

第一版 alerting 會做成一個新的 backend 子系統：

```text
alert rule -> incident evaluation -> notification outbox -> webhook/email delivery
```

核心目標是支援時間窗門檻，例如：

- 最近 5 分鐘 ping packet loss >= 20%，且 samples >= 3。
- 最近 5 分鐘 TCP failure percent >= 10%，且 samples >= 3。

設計原則：

- Alert condition 採用新的 `domain/alertcondition` AST，不直接擴充現有 `domain/selector`。
- Rule scope 與 condition 分開：scope 決定套用到哪些 check/probe，condition 決定什麼狀態算觸發。
- Result submit 後同步做「評估與 incident 狀態更新」，但外部通知走 `notification_outbox` 非同步 worker。
- Webhook 和 Email 第一版都支援，但都只由 outbox worker 發送，不影響 probe result ingestion。
- Probe offline 類 alert 不放在第一版，因為那需要 scheduler 掃 heartbeat，不適合只靠 result submit 觸發。

## Architecture

新增 backend packages：

### `internal/domain/alertcondition`

負責 alert condition JSON AST 的 domain 邏輯。

- Parse raw JSON。
- Validate node shape、operator、metric name、window、sample constraints。
- Canonicalize JSON，讓 semantically equivalent 的 condition 能以穩定格式存進 DB。
- Evaluate 已查好的 metric summary。
- 不直接查 DB。
- 不 import application、transport、infrastructure。

V1 支援 node：

```text
all
any
not
metric
```

### `internal/domain/alert`

負責 alerting domain models 與 validators。

包含：

- `Rule`
- `RuleScope`
- `Incident`
- `Severity`
- `IncidentStatus`
- `Notification`
- `NotificationOutboxJob`

Domain validators：

- rule name
- rule description
- severity
- condition JSON size
- notification name
- webhook URL
- email recipients
- window seconds
- min samples

### `internal/controller/application/alert`

Authenticated user-facing alert API use cases。

負責：

- Create/update/delete/list alert rules。
- List/get/acknowledge/resolve incidents。
- Create/update/delete/list notifications。
- Validate project access and project role permissions。
- Translate domain/application errors through sentinel errors。
- Record application events for audit-worthy actions。

Permission design：

- 新增 project action：`write:project_alerts`。
- `owner`、`admin`、`editor` 可建立/修改/delete rules and notifications。
- `viewer` 可讀 rules、notifications、incidents，但不能修改。
- Incident acknowledgement/resolution first version 允許 `owner/admin/editor`。

### `internal/controller/application/alerteval`

Internal-only evaluator，給 result ingestion path 呼叫。

Input 是本次 submit 影響到的 assignment pairs：

```text
project_id
probe_id
check_id
check_type
probe_storage_id
check_storage_id
```

負責：

- 查 active matching alert rules。
- 查每條 rule 需要的 time-window metric summary。
- Evaluate alert condition AST。
- Open incident。
- Update active incident。
- Resolve active incident。
- Enqueue notification outbox jobs。

這個 package 不提供 HTTP endpoint。

### `internal/controller/application/notification`

負責 notification outbox worker orchestration。

行為：

- Pull pending jobs。
- Claim jobs with DB locking。
- Dispatch 到 webhook/email sender。
- Retry with backoff。
- Mark delivered/failed/discarded。

### `internal/controller/infrastructure/postgres/alert`

負責 PostgreSQL persistence。

SQL files：

```text
server/db/query/alerts.sql
server/db/query/notification_outbox.sql
```

Repository responsibilities：

- alert rules CRUD
- incident state transitions
- notifications CRUD
- outbox enqueue/claim/update
- query metric summaries for alert evaluation

### `internal/controller/infrastructure/notify`

負責外部通知 delivery adapters。

V1 senders：

- `WebhookSender`
- `EmailSender`

Webhook 用 standard `net/http`。

Email 建議先做 SMTP adapter，因為 repo 目前沒有第三方 email SDK。未來如果要接 Resend、SendGrid、SES，可以新增 provider-specific implementation，但不要讓 application package 依賴 provider SDK。

### `internal/controller/transport/http/handler/alert`

負責 HTTP binding、response DTO、error mapping。

Routes 掛在：

```text
/api/{version}/projects/{ref}/alerts/*
```

### Composition

在 `internal/controller/app/bootstrap.go` wire：

- alert repository
- alert service
- alert evaluator
- notification service
- webhook sender
- email sender
- notification worker

Extend `proberuntime.Service` with an optional `AlertEvaluator` port：

```go
type AlertEvaluator interface {
    EvaluateChangedAssignments(ctx context.Context, inputs []ChangedAssignmentInput) error
}
```

After successful result writes：

```text
write ping/tcp/traceroute results
collect changed assignment pairs
evaluate alerts
return SubmitResults response
```

Important behavior：

- Result persistence 成功後才 evaluate。
- Alert evaluation failure 會 log/trace，但不要讓 probe result ingestion 失敗。
- 外部通知永遠不在 `SubmitResults` request path 直接送。

## Public API

新增 TypeSpec：

```text
api/models/alert.tsp
api/services/alerts.tsp
```

並在 `api/main.tsp` import。

### Rule Routes

```text
GET    /projects/{ref}/alerts/rules
POST   /projects/{ref}/alerts/rules
GET    /projects/{ref}/alerts/rules/{rule_id}
PATCH  /projects/{ref}/alerts/rules/{rule_id}
DELETE /projects/{ref}/alerts/rules/{rule_id}
```

### Incident Routes

```text
GET    /projects/{ref}/alerts/incidents
GET    /projects/{ref}/alerts/incidents/{incident_id}
POST   /projects/{ref}/alerts/incidents/{incident_id}/acknowledgements
POST   /projects/{ref}/alerts/incidents/{incident_id}/resolutions
```

### Notification Routes

```text
GET    /projects/{ref}/alerts/notifications
POST   /projects/{ref}/alerts/notifications
PATCH  /projects/{ref}/alerts/notifications/{notification_id}
DELETE /projects/{ref}/alerts/notifications/{notification_id}
POST   /projects/{ref}/alerts/notifications/{notification_id}/test
```

### Create Rule Request

```json
{
	"name": "High packet loss",
	"description": "Alert when Tokyo probes see elevated packet loss.",
	"enabled": true,
	"severity": "warning",
	"scope": {
		"checkType": "ping",
		"probeSelector": {
			"label": {
				"key": "region",
				"op": "eq",
				"value": "tokyo"
			}
		}
	},
	"condition": {
		"metric": {
			"name": "ping.loss_percent",
			"op": "gte",
			"value": 20,
			"windowSeconds": 300,
			"minSamples": 3
		}
	},
	"notificationIds": ["77777777-7777-7777-7777-777777777777"]
}
```

### Alert Condition AST V1

Supported top-level node shape follows the same pattern as current selector AST：每個 node object 只能有一個 operator。

#### `metric`

```json
{
	"metric": {
		"name": "ping.loss_percent",
		"op": "gte",
		"value": 20,
		"windowSeconds": 300,
		"minSamples": 3
	}
}
```

Fields：

```text
name: supported metric name
op: gt | gte | lt | lte | eq
value: number
windowSeconds: 60..86400
minSamples: 1..10000
```

#### `all`

```json
{
	"all": [
		{
			"metric": {
				"name": "ping.loss_percent",
				"op": "gte",
				"value": 20,
				"windowSeconds": 300,
				"minSamples": 3
			}
		},
		{
			"metric": {
				"name": "ping.success_rate",
				"op": "lt",
				"value": 95,
				"windowSeconds": 300,
				"minSamples": 3
			}
		}
	]
}
```

#### `any`

```json
{
	"any": [
		{
			"metric": {
				"name": "ping.loss_percent",
				"op": "gte",
				"value": 20,
				"windowSeconds": 300,
				"minSamples": 3
			}
		},
		{
			"metric": {
				"name": "ping.max_rtt_ms",
				"op": "gte",
				"value": 1000,
				"windowSeconds": 300,
				"minSamples": 3
			}
		}
	]
}
```

#### `not`

```json
{
	"not": {
		"metric": {
			"name": "ping.success_rate",
			"op": "gte",
			"value": 99,
			"windowSeconds": 300,
			"minSamples": 3
		}
	}
}
```

### Supported V1 Metrics

Ping：

```text
ping.loss_percent
ping.average_rtt_ms
ping.max_rtt_ms
ping.success_rate
```

TCP：

```text
tcp.failure_percent
tcp.average_connect_ms
tcp.max_connect_ms
tcp.success_rate
```

Excluded from V1：

```text
traceroute.*
probe.offline
custom arithmetic expressions
percentile metrics
multi-window burn rate
```

### Scope V1

Scope shape：

```json
{
	"checkType": "ping",
	"probeId": "33333333-3333-3333-3333-333333333333",
	"checkId": "44444444-4444-4444-4444-444444444444",
	"probeSelector": {
		"label": {
			"key": "region",
			"op": "eq",
			"value": "tokyo"
		}
	}
}
```

Rules：

- `checkType` is required。
- `probeId` optional。
- `checkId` optional。
- `probeSelector` optional，並重用現有 `domain/selector`。
- `probeSelector` 只做 probe label matching。
- `condition` 只做 telemetry metric evaluation。
- `condition.metric.name` 必須跟 `scope.checkType` 相容。

Do not mix label selectors into alert condition AST. 這樣可以避免 metadata matching 與 time-window metric evaluation 混在同一套 evaluator 裡。

## Database Tables

新增 Goose migration。

### Enum Types

```sql
CREATE TYPE alert_severity AS ENUM ('info', 'warning', 'critical');
CREATE TYPE alert_rule_status AS ENUM ('enabled', 'disabled');
CREATE TYPE alert_incident_status AS ENUM ('open', 'acknowledged', 'resolved');
CREATE TYPE notification_type AS ENUM ('webhook', 'email');
CREATE TYPE notification_outbox_status AS ENUM ('pending', 'sending', 'delivered', 'failed', 'discarded');
```

### `alert_rules`

Stores configured alert rules。

```sql
CREATE TABLE alert_rules (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id),
    name text NOT NULL,
    description text,
    status alert_rule_status NOT NULL DEFAULT 'enabled',
    severity alert_severity NOT NULL,
    check_type check_type NOT NULL,
    probe_id uuid,
    check_id uuid,
    probe_selector jsonb NOT NULL DEFAULT '{}'::jsonb,
    condition jsonb NOT NULL,
    condition_version text NOT NULL,
    created_by_user_id uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT alert_rules_name_not_empty CHECK (length(btrim(name)) > 0),
    CONSTRAINT alert_rules_description_not_empty CHECK (description IS NULL OR length(btrim(description)) > 0),
    CONSTRAINT alert_rules_probe_selector_is_object CHECK (jsonb_typeof(probe_selector) = 'object'),
    CONSTRAINT alert_rules_condition_is_object CHECK (jsonb_typeof(condition) = 'object'),
    CONSTRAINT alert_rules_deleted_at_after_created_at CHECK (deleted_at IS NULL OR deleted_at >= created_at),
    CONSTRAINT fk_alert_rules_project_probe
        FOREIGN KEY (project_id, probe_id) REFERENCES probes(project_id, id),
    CONSTRAINT fk_alert_rules_project_check
        FOREIGN KEY (project_id, check_id) REFERENCES checks(project_id, id)
);
```

Indexes：

```sql
CREATE INDEX idx_alert_rules_project_active
    ON alert_rules (project_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_alert_rules_project_status_active
    ON alert_rules (project_id, status)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_alert_rules_project_check_type_status_active
    ON alert_rules (project_id, check_type, status)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_alert_rules_project_probe_active
    ON alert_rules (project_id, probe_id)
    WHERE deleted_at IS NULL AND probe_id IS NOT NULL;

CREATE INDEX idx_alert_rules_project_check_active
    ON alert_rules (project_id, check_id)
    WHERE deleted_at IS NULL AND check_id IS NOT NULL;
```

### `alert_incidents`

Tracks open/resolved alert instances。

```sql
CREATE TABLE alert_incidents (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id),
    rule_id uuid NOT NULL REFERENCES alert_rules(id),
    probe_id uuid NOT NULL,
    check_id uuid NOT NULL,
    check_type check_type NOT NULL,
    status alert_incident_status NOT NULL,
    severity alert_severity NOT NULL,
    opened_at timestamptz NOT NULL,
    acknowledged_at timestamptz,
    acknowledged_by_user_id uuid REFERENCES users(id),
    resolved_at timestamptz,
    resolved_by_user_id uuid REFERENCES users(id),
    last_evaluated_at timestamptz NOT NULL,
    last_triggered_at timestamptz NOT NULL,
    last_value double precision,
    last_summary jsonb NOT NULL DEFAULT '{}'::jsonb,
    notification_fingerprint text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT alert_incidents_last_summary_is_object CHECK (jsonb_typeof(last_summary) = 'object'),
    CONSTRAINT alert_incidents_acknowledged_consistency CHECK (
        (status = 'acknowledged' AND acknowledged_at IS NOT NULL)
        OR status <> 'acknowledged'
    ),
    CONSTRAINT alert_incidents_resolved_consistency CHECK (
        (status = 'resolved' AND resolved_at IS NOT NULL)
        OR status <> 'resolved'
    ),
    CONSTRAINT fk_alert_incidents_project_probe
        FOREIGN KEY (project_id, probe_id) REFERENCES probes(project_id, id),
    CONSTRAINT fk_alert_incidents_project_check
        FOREIGN KEY (project_id, check_id) REFERENCES checks(project_id, id)
);
```

Indexes：

```sql
CREATE INDEX idx_alert_incidents_project_status_opened
    ON alert_incidents (project_id, status, opened_at DESC);

CREATE INDEX idx_alert_incidents_project_rule_status
    ON alert_incidents (project_id, rule_id, status);

CREATE INDEX idx_alert_incidents_project_probe_check_status
    ON alert_incidents (project_id, probe_id, check_id, status);

CREATE UNIQUE INDEX uq_alert_incidents_active_rule_probe_check
    ON alert_incidents (rule_id, probe_id, check_id)
    WHERE status IN ('open', 'acknowledged');
```

### `notifications`

Stores webhook/email notification config。

```sql
CREATE TABLE notifications (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id),
    name text NOT NULL,
    type notification_type NOT NULL,
    enabled boolean NOT NULL DEFAULT true,
    config jsonb NOT NULL,
    created_by_user_id uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT notifications_name_not_empty CHECK (length(btrim(name)) > 0),
    CONSTRAINT notifications_config_is_object CHECK (jsonb_typeof(config) = 'object'),
    CONSTRAINT notifications_deleted_at_after_created_at CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);
```

Indexes：

```sql
CREATE INDEX idx_notifications_project_active
    ON notifications (project_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_notifications_project_enabled_active
    ON notifications (project_id, enabled)
    WHERE deleted_at IS NULL;
```

Config shapes：

Webhook：

```json
{
	"url": "https://example.com/netstamp-alerts"
}
```

Email：

```json
{
	"to": ["ops@example.com"]
}
```

Security defaults：

- V1 webhook has no custom secret header unless added later。
- SMTP credentials come from env，不存在 notification config。
- Do not store email provider secrets in DB。

### `alert_notifications`

Many-to-many mapping between rules and notification destinations。

```sql
CREATE TABLE alert_notifications (
    project_id uuid NOT NULL REFERENCES projects(id),
    rule_id uuid NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    notification_id uuid NOT NULL REFERENCES notifications(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (rule_id, notification_id),
    CONSTRAINT fk_alert_notifications_project_rule
        FOREIGN KEY (project_id, rule_id) REFERENCES alert_rules(project_id, id),
    CONSTRAINT fk_alert_notifications_project_notification
        FOREIGN KEY (project_id, notification_id) REFERENCES notifications(project_id, id)
);
```

### `notification_outbox`

Reliable delivery queue。

```sql
CREATE TABLE notification_outbox (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id),
    incident_id uuid NOT NULL REFERENCES alert_incidents(id),
    rule_id uuid NOT NULL REFERENCES alert_rules(id),
    notification_id uuid NOT NULL REFERENCES notifications(id),
    notification_type notification_type NOT NULL,
    event_type text NOT NULL,
    status notification_outbox_status NOT NULL DEFAULT 'pending',
    payload jsonb NOT NULL,
    attempt_count integer NOT NULL DEFAULT 0,
    max_attempts integer NOT NULL DEFAULT 5,
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    last_attempt_at timestamptz,
    delivered_at timestamptz,
    last_error text,
    dedupe_key text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT notification_outbox_event_type_not_empty CHECK (length(btrim(event_type)) > 0),
    CONSTRAINT notification_outbox_payload_is_object CHECK (jsonb_typeof(payload) = 'object'),
    CONSTRAINT notification_outbox_attempt_count_non_negative CHECK (attempt_count >= 0),
    CONSTRAINT notification_outbox_max_attempts_positive CHECK (max_attempts > 0),
    CONSTRAINT notification_outbox_last_error_not_empty CHECK (last_error IS NULL OR length(btrim(last_error)) > 0),
    CONSTRAINT notification_outbox_dedupe_key_not_empty CHECK (length(btrim(dedupe_key)) > 0)
);
```

Indexes：

```sql
CREATE INDEX idx_notification_outbox_status_next_attempt
    ON notification_outbox (status, next_attempt_at);

CREATE INDEX idx_notification_outbox_project_created
    ON notification_outbox (project_id, created_at DESC);

CREATE UNIQUE INDEX uq_notification_outbox_dedupe_key
    ON notification_outbox (dedupe_key);
```

Dedupe key：

```text
incident_id + notification_id + event_type + incident_status_transition_time
```

## Evaluation Behavior

Evaluation triggered after successful result persistence in `SubmitResults`。

Flow：

```text
normalize result payload
resolve active assignments
write result rows
collect changed assignment pairs
AlertEvaluator.EvaluateChangedAssignments(ctx, pairs)
return SubmitResults response
```

Evaluator behavior：

1. Load enabled alert rules matching project/check type。
2. Apply optional `probe_id` filter。
3. Apply optional `check_id` filter。
4. Apply `probeSelector` against probe labels when present。
5. Parse/canonicalize condition if needed，or use already persisted canonical condition。
6. Collect metric summary requirements from condition AST。
7. Deduplicate metric summary queries by:

```text
probe_id
check_id
metric family
window_seconds
```

8. Query recent summary for each requirement。
9. Evaluate condition AST。
10. If condition is true and no active incident exists:
    - create `open` incident
    - enqueue `incident.opened` notifications
11. If condition is true and active incident exists:
    - update `last_evaluated_at`
    - update `last_triggered_at`
    - update `last_value`
    - update `last_summary`
12. If condition is false and active incident exists:
    - mark incident `resolved`
    - enqueue `incident.resolved` notifications
13. If there are not enough samples:
    - condition evaluates false
    - summary should record insufficient sample reason

Performance guardrails：

- Batch rule lookup by project/check type。
- Do not evaluate all project rules on every submit。
- Do not send notification during evaluation。
- Alert evaluation failure should be logged and traced but should not fail result ingestion。

### Incident Status

```text
open
acknowledged
resolved
```

Status behavior：

- `open`: condition triggered and no user acknowledgement。
- `acknowledged`: user acknowledged but condition may still be active。
- `resolved`: condition cleared or user manually resolved。
- Manual resolve is allowed，but next true evaluation may open a new incident。
- Acknowledged incident should stay acknowledged while condition remains true。
- Resolved incident should not be reused; next trigger creates a new incident。

## Metric Summary Design

Alert evaluator should not reuse frontend-facing result DTOs directly. It should use smaller repository methods that return alert-specific summaries。

Suggested domain summary shape：

```go
type MetricSummary struct {
    Metric       string
    WindowStart  time.Time
    WindowEnd    time.Time
    Samples      int64
    Value        float64
    Values       map[string]float64
    Source       string
}
```

Ping summary query can reuse logic from existing ping insight:

- `loss_percent`
- `average_rtt_ms`
- `max_rtt_ms`
- `success_rate`
- `samples`

TCP summary query can reuse logic from existing TCP insight:

- `failure_percent`
- `average_connect_ms`
- `max_connect_ms`
- `success_rate`
- `samples`

Read source:

- For recent windows inside raw retention, read raw result tables。
- If a future alert allows long windows past raw retention, use rollup tables。
- V1 should default to windows short enough to fit raw retention, but repository can still support rollups where available。

## Notification Delivery

Notification delivery uses outbox pattern。

### Outbox Worker

Runs inside controller process in V1。

Loop behavior：

```text
sleep interval
claim pending jobs
deliver each job
mark delivered or retry
repeat until shutdown
```

Claim query：

- Select rows where `status = 'pending'` and `next_attempt_at <= now()`。
- Use `FOR UPDATE SKIP LOCKED`。
- Mark selected rows `sending` before delivery。
- On process crash, a recovery query should return stale `sending` rows back to `pending` after a timeout。

Failure handling：

- Increment `attempt_count`。
- Store conservative `last_error`。
- Schedule exponential backoff。
- Mark `failed` after `max_attempts`。

Backoff default：

```text
attempt 1: 30s
attempt 2: 2m
attempt 3: 5m
attempt 4: 15m
attempt 5: failed
```

### Webhook

HTTP behavior：

- `POST` JSON。
- Timeout default 10s。
- Treat 2xx as success。
- Treat non-2xx and network errors as retryable until max attempts。

Payload shape：

```json
{
	"eventType": "incident.opened",
	"project": {
		"id": "22222222-2222-2222-2222-222222222222",
		"name": "Production"
	},
	"rule": {
		"id": "55555555-5555-5555-5555-555555555555",
		"name": "High packet loss",
		"severity": "warning"
	},
	"incident": {
		"id": "88888888-8888-8888-8888-888888888888",
		"status": "open",
		"openedAt": "2026-06-14T10:00:00Z"
	},
	"target": {
		"probeId": "33333333-3333-3333-3333-333333333333",
		"checkId": "44444444-4444-4444-4444-444444444444",
		"checkType": "ping"
	},
	"summary": {
		"metric": "ping.loss_percent",
		"value": 23.5,
		"threshold": 20,
		"operator": "gte",
		"windowSeconds": 300,
		"samples": 5
	}
}
```

### Email

SMTP config from env。

Notification config only stores recipients：

```json
{
	"to": ["ops@example.com"]
}
```

Email V1 behavior：

- Send plain text and simple HTML。
- Subject includes severity、rule name、status。
- Body includes project、probe/check IDs、metric value、threshold、opened/resolved time。
- No attachments。
- No per-notification SMTP credentials。

## Configuration

Add config fields in `internal/controller/config` and mirror in `server/.env.example`。

```text
ALERT_EVALUATION_ENABLED=true
NOTIFICATION_WORKER_ENABLED=true
NOTIFICATION_WORKER_INTERVAL=5s
NOTIFICATION_WORKER_BATCH_SIZE=25
NOTIFICATION_HTTP_TIMEOUT=10s
SMTP_HOST=
SMTP_PORT=
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=
```

Validation：

- Worker interval > 0。
- Worker batch size > 0。
- HTTP timeout > 0。
- If any enabled email notification exists and notification worker is enabled, SMTP config must be valid before sending email jobs。
- Missing SMTP config should fail email delivery jobs with clear retryable/non-retryable error classification。

## Frontend

Add alerts console after API exists。

Routes：

```text
/projects/:projectRef/alerts
/projects/:projectRef/alerts/rules/:ruleId
```

UI V1：

- Alerts nav item in app shell。
- Rules table:
  - name
  - severity
  - scope
  - condition summary
  - notifications
  - enabled state
- Incident table:
  - status
  - severity
  - rule
  - probe
  - check
  - opened time
  - resolved time
  - last value
- Rule editor:
  - name
  - description
  - severity
  - enabled
  - scope fields
  - metric threshold builder
  - notification multi-select
- Notification editor:
  - webhook URL
  - email recipients
- Incident detail:
  - current status
  - latest summary
  - acknowledgement action
  - manual resolve action

Important frontend constraint：

- Do not build full arbitrary AST UI in V1。
- UI can generate simple `metric` nodes。
- Backend API can still accept logical AST for future automation/import use cases。

## Logging, Tracing, And Metrics

### Application Events

Add alert event recorder following existing auth/project/label/check/probe patterns。

Events：

```text
alert.rule.create.success
alert.rule.create.failure
alert.rule.update.success
alert.rule.update.failure
alert.rule.delete.success
alert.rule.delete.failure
alert.incident.open.success
alert.incident.open.failure
alert.incident.acknowledge.success
alert.incident.acknowledge.failure
alert.incident.resolve.success
alert.incident.resolve.failure
notification.delivery.success
notification.delivery.failure
```

Privacy：

- Do not log webhook URLs in full。
- Do not log email recipient list。
- Do not log raw notification payloads。
- Do not log raw result error messages from probe payloads。

### Tracing

Add spans：

```text
alert.evaluate_changed_assignments
alert.evaluate_rule
alert.rule.create
alert.rule.update
alert.incident.acknowledge
alert.incident.resolve
notification.deliver
```

DB spans should use existing postgres helper style。

### Metrics

Add Prometheus metrics：

```text
netstamp_alert_evaluations_total
netstamp_alert_evaluation_failures_total
netstamp_alert_incidents_open_total
netstamp_alert_incidents_resolved_total
netstamp_notification_delivery_attempts_total
netstamp_notification_delivery_failures_total
netstamp_notification_outbox_pending
```

## Testing Plan

### Backend Unit Tests

`domain/alertcondition`：

- Parses valid `metric` node。
- Parses valid `all`、`any`、`not`。
- Rejects invalid JSON。
- Rejects array root。
- Rejects multiple operators in one node。
- Rejects unknown node。
- Rejects unknown metric。
- Rejects unsupported operator。
- Rejects empty logical children。
- Rejects non-number metric value。
- Rejects invalid window seconds。
- Rejects invalid min samples。
- Canonicalizes stable JSON。
- Enforces metric/check type compatibility。
- Evaluates `gt/gte/lt/lte/eq` correctly。
- Evaluates insufficient samples as false。

`application/alerteval`：

- Opens incident when condition becomes true。
- Does not duplicate active incidents。
- Updates active incident when condition remains true。
- Resolves active incident when condition clears。
- Preserves acknowledged incident while condition remains true。
- Opens a new incident after a previous incident was resolved。
- Enqueues `incident.opened` notification once。
- Enqueues `incident.resolved` notification once。
- Ignores unrelated project/probe/check pairs。
- Handles repository failures without panicking。

`application/notification`：

- Delivers pending webhook jobs。
- Delivers pending email jobs。
- Retries failed jobs with backoff。
- Marks failed after max attempts。
- Does not retry delivered jobs。
- Dispatches by notification type。
- Handles disabled/deleted notification safely。

### Repository Tests

- Create/list/get/update/delete alert rules。
- Soft-deleted rules excluded from active lookup。
- Active rule lookup filters by project/check type/probe/check。
- Create active incident。
- Unique active incident constraint prevents duplicates。
- Resolve active incident。
- Acknowledge active incident。
- Enqueue outbox row with dedupe key。
- Duplicate dedupe key does not create duplicate notification。
- Claim pending outbox rows with `FOR UPDATE SKIP LOCKED`。
- Mark delivered。
- Mark retry。
- Mark failed。

### HTTP Handler Tests

- Viewer can list rules/incidents/notifications。
- Viewer cannot create/update/delete rules/notifications。
- Owner/admin/editor can create/update/delete rules/notifications。
- Invalid rule condition maps to `422` with field detail。
- Invalid probe selector maps to `422`。
- Missing project maps to `404`。
- Missing rule maps to `404`。
- Missing incident maps to `404`。
- Notification test endpoint validates config and returns result。

### Integration Tests

- Submit ping results crossing loss threshold creates incident。
- Submit later healthy ping window resolves incident。
- Submit TCP failures crossing threshold creates incident。
- Webhook outbox row is created on incident open。
- Webhook delivery succeeds against local test HTTP server。
- Email delivery uses fake sender in tests。

### Frontend Checks

- Regenerated OpenAPI types compile。
- Rule form serializes expected condition JSON。
- Rule list renders enabled/disabled states。
- Incident list renders open/acknowledged/resolved states。
- Notification editor validates webhook URL。
- Notification editor validates email recipients。

Validation commands：

```text
pnpm generate:openapi
cd server && go test ./...
pnpm --filter @netstamp/web typecheck
pnpm --filter @netstamp/web lint
```

## Rollout Plan

Recommended implementation order：

1. Add `domain/alertcondition` and tests。
2. Add database migration and SQLC queries。
3. Add alert repositories。
4. Add alert application service and HTTP APIs。
5. Add alert evaluator and wire it after result writes。
6. Add notification outbox tables and repository。
7. Add webhook/email senders。
8. Add notification worker and app lifecycle wiring。
9. Add TypeSpec contract and regenerate OpenAPI/web types。
10. Add frontend alerts pages。
11. Add integration tests。

Feature flags：

```text
ALERT_EVALUATION_ENABLED
NOTIFICATION_WORKER_ENABLED
```

This allows shipping rule/incident API before enabling external delivery。

## Explicit Non-Goals For V1

V1 does not include：

- Probe offline alerts。
- Traceroute alerts。
- Slack/Discord/PagerDuty。
- Escalation policies。
- Silence windows。
- Maintenance windows。
- Alert grouping across multiple probes/checks。
- SLO burn-rate rules。
- Percentile latency metrics。
- Arbitrary expression language。
- User-specific notification preferences。
- Multi-tenant organization-level routing。

## Assumptions And Defaults

- V1 supports ping and TCP metric threshold alerts only。
- V1 supports webhook and email delivery。
- V1 does not support arbitrary arithmetic expressions。
- `condition` and `probeSelector` are both JSON ASTs, but live in separate fields and separate domain packages。
- Result ingestion should not fail because alert evaluation or notification delivery fails。
- Notification worker runs in the controller process for V1。
- Email uses SMTP env config unless a provider-specific integration is chosen later。
- The UI only builds simple metric threshold rules in V1。
- Advanced logical AST can still be supported at API level for future automation。
