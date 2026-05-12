# Probe Agent Plane 設計筆記

## 1. 背景

Probe agent plane 是 Netstamp controller 與部署在使用者環境中的 probe agent 之間的 runtime 協定。

目前 backend 已經有 probe registry、probe credential、probe status、check、ping check config、probe-check assignment 與 ping result 的資料模型。這份文件以目前 backend 設計為準，避免沿用早期草稿中的自訂 `tasks`、通用 `assignments`、通用 `results`、server-side heap scheduler 等尚未實作模型。

目前 runtime plane 的核心流程是：

1. 使用者在 controller 端建立 probe，controller 產生 probe ID 與一次性 plaintext secret。
2. probe 使用 local env 啟動，包含 controller URL、probe ID、probe secret。
3. probe 以 `Authorization: Probe <secret>` 呼叫 runtime API。
4. controller 驗證 probe secret hash，`hello` 回傳 server time、agent 版本相容下限與 runtime config。
5. probe 透過 `heartbeat` 更新 runtime status，並透過 `assignments` 取得 active checks。
6. probe 依照 assignment 中的 check 設定，在本機排程與執行檢測。
7. probe 以 grouped batch payload 提交 result，controller 驗證 active assignment 與 result type 後寫入 typed result storage。

## 2. 目前目標

目前文件描述的現況能力：

- probe 啟動時呼叫 `hello`，建立 runtime session，但不傳 probe status body。
- probe 定期呼叫 `heartbeat`，更新 `last_seen_at` 與 runtime status。
- probe 呼叫 `assignments`，取得目前指派給自己的 active checks。
- probe 呼叫 `results`，以 check/type 分組批次提交檢測結果。
- controller 以現有 `probe_check_assignments` 作為 assignment materialization。
- probe status 回傳 agent、網路與未來可擴充的硬體/架構資訊，讓 control plane 有 probe runtime 環境概念。

目前非目標：

- 不在本文宣稱已存在通用 task/result schema。
- 不在本文宣稱已存在 assignment lease、running state、expired state 或 controller-side heap scheduler。
- 不把 runtime config 設計成完整 controller-managed config system。

## 3. 目前 Backend 分層

Probe runtime route 位於：

```text
server/internal/controller/transport/http/proberuntime
```

Application service 位於：

```text
server/internal/controller/application/proberuntime
```

目前主要資料模型位於：

```text
server/internal/domain/probe
server/internal/domain/assignment
server/internal/domain/check
server/internal/domain/ping
```

PostgreSQL persistence 使用現有 schema：

```text
probes
probe_credentials
probe_statuses
checks
ping_check_configs
probe_check_assignments
ping_results
```

## 4. Runtime Authentication

所有 probe runtime API 使用 probe secret authentication。

```http
Authorization: Probe <secret>
```

驗證流程：

1. transport 從 path 取得 `{probe_id}`。
2. transport 從 `Authorization` header 取出 `Probe` scheme 的 secret。
3. application service 使用 `GetActiveProbeCredential` 取得該 probe 的 credential record。
4. 若 probe disabled，回傳 `403`。
5. secret verifier 比對 plaintext secret 與 `probe_credentials.secret_hash`。
6. 驗證成功後才執行 hello、heartbeat 或 assignments 行為。

注意：

- controller 不應相信 request body 中的 probe identity。
- plaintext secret 只存在於建立或 rotate probe secret 的 response 以及 probe local env。
- log 與 tracing 不可記錄 plaintext secret 或 secret hash。
- 目前不是 `Authorization: Bearer <token>`，也沒有 JWT subject equals probe ID 的 runtime 模型。

## 5. Runtime API

目前 backend 已註冊的 runtime API：

```text
POST /runtime/probes/{probe_id}/hello
POST /runtime/probes/{probe_id}/heartbeat
GET  /runtime/probes/{probe_id}/assignments
POST /runtime/probes/{probe_id}/results
```

在實際 HTTP server 中，這些 route 會掛在 API version prefix 下，例如：

```text
/api/v1/runtime/probes/{probe_id}/hello
```

### 5.1 Probe Hello

Endpoint：

```http
POST /runtime/probes/{probe_id}/hello
Authorization: Probe <secret>
Content-Type: application/json
```

用途：

- probe 啟動時宣告 runtime session 開始。
- controller 驗證 probe secret。
- controller 不從 `hello` 接收 probe status 或 inventory。
- controller 回傳 server time、agent 版本相容下限與 runtime config。

Request body：

```text
empty
```

`hello` 不應傳 `agentVersion`、`publicV4`、`publicV6`、`as`、`addrs` 或任何 inventory/status 欄位。這些資料應由 `heartbeat` 傳送。

Response body：

```json
{
	"serverTime": "2026-05-13T10:00:01Z",
	"minimumSupportedAgentVersion": "0.1.0",
	"config": {
		"heartbeatIntervalSeconds": 30,
		"assignmentPollIntervalSeconds": 30,
		"maxConcurrentChecks": 16,
		"initialBackoffSeconds": 1,
		"maxBackoffSeconds": 30,
		"maxAttempts": 5
	}
}
```

Response 欄位說明：

| 欄位                                   | 型別      | 說明                                                                                                       |
| -------------------------------------- | --------- | ---------------------------------------------------------------------------------------------------------- |
| `serverTime`                           | timestamp | controller 當下 UTC 時間。                                                                                 |
| `minimumSupportedAgentVersion`         | string    | controller 目前最低支援的 probe agent 版本。probe 版本低於此值時應停止執行或進入明確的 incompatible 狀態。 |
| `config.heartbeatIntervalSeconds`      | integer   | probe 呼叫 heartbeat 的間隔。                                                                              |
| `config.assignmentPollIntervalSeconds` | integer   | probe 拉取 assignments 的間隔。                                                                            |
| `config.maxConcurrentChecks`           | integer   | probe 本機同時執行 checks 的上限。                                                                         |
| `config.initialBackoffSeconds`         | integer   | 未來 result submission retry 的初始 backoff。                                                              |
| `config.maxBackoffSeconds`             | integer   | 未來 result submission retry 的最大 backoff。                                                              |
| `config.maxAttempts`                   | integer   | 未來 result submission retry 的最大嘗試次數。                                                              |

目前 `hello` 不回傳：

- `accepted`
- `probe_id`
- assignment payload
- probe status 或 inventory payload

### 5.2 Probe Heartbeat

Endpoint：

```http
POST /runtime/probes/{probe_id}/heartbeat
Authorization: Probe <secret>
Content-Type: application/json
```

用途：

- probe 定期回報仍在線。
- controller 更新 `probe_statuses.status = online`。
- controller 更新 `last_seen_at = now()`。
- controller 更新 runtime status 與 lightweight inventory 欄位。

Request body 對齊目前 `server/internal/controller/transport/http/proberuntime/runtime_status.go`：

```json
{
	"agentVersion": "netstamp-probe/0.1.0",
	"publicV4": "203.0.113.10",
	"publicV6": "2001:db8::10",
	"as": "AS15169 Google LLC",
	"addrs": ["10.0.0.10", "fd00::10"]
}
```

欄位說明：

| 欄位           | 型別             | 必填 | 說明                                           |
| -------------- | ---------------- | ---- | ---------------------------------------------- |
| `agentVersion` | string           | no   | probe agent 版本。                             |
| `publicV4`     | IP address       | no   | probe 觀測到或偵測到的 public IPv4。           |
| `publicV6`     | IP address       | no   | probe 觀測到或偵測到的 public IPv6。           |
| `as`           | string           | no   | probe 所在或出口網路的 AS 資訊。               |
| `addrs`        | IP address array | no   | probe 本機 interface 或觀測到的 address list。 |

Response body：

```json
{
	"serverTime": "2026-05-13T10:00:31Z"
}
```

目前 `heartbeat` 不回傳 `nextHeartbeatAfterSeconds`，probe 端應使用最新 runtime config 的 `heartbeatIntervalSeconds` 決定下一次 heartbeat 時間。啟動時使用 `hello.config`，之後可由 assignments response 的 `config` refresh。

### 5.3 List Assignments

Endpoint：

```http
GET /runtime/probes/{probe_id}/assignments
Authorization: Probe <secret>
```

用途：

- probe 拉取目前指派給自己的 active checks。
- controller 從 `probe_check_assignments` 查詢 active assignment。
- controller 只回傳 enabled、未刪除 probe 以及未刪除 checks 的 assignment。
- controller 同時回傳最新 probe runtime config，讓 probe 不需要額外打 config endpoint。

Response body：

```json
{
	"serverTime": "2026-05-13T10:00:40Z",
	"config": {
		"heartbeatIntervalSeconds": 30,
		"assignmentPollIntervalSeconds": 30,
		"maxConcurrentChecks": 16,
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

`config` 的 shape 必須和 `hello.config` 相同。即使 `assignments` 為空，controller 也應回傳 `serverTime` 與 `config`。

目前 assignment 沒有：

- `dueAt`
- `leasedUntil`
- `running`
- `expired`
- server-side dispatch state
- query `limit`
- query `capabilities`

probe 端應把 `check.intervalSeconds` 當作本機執行排程的主要依據，並把 assignments response 的 `config` 視為最新 runtime config。

### 5.4 Submit Results

Endpoint：

```http
POST /runtime/probes/{probe_id}/results
Authorization: Probe <secret>
Content-Type: application/json
```

用途：

- probe 將本機執行完成的 checks 以 batch 方式提交給 controller。
- request 以 `checkId` 與 `type` 分組，讓同一個 check 可以一次送多筆同型別結果。
- controller 先驗證 probe runtime auth，再批次查詢這批 `checkId` 是否仍是該 probe 的 active assignments。
- controller 驗證 request `type` 必須和 assignment 裡的 `check.type` 相同。
- controller 依照 type dispatch 到 typed result storage；目前支援 `ping`，未來可加 HTTP/TCP/DNS 等 typed payload。

Request body：

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

Response body：

```json
{
	"accepted": 1,
	"serverTime": "2026-05-13T10:00:02Z"
}
```

Result group 欄位：

| 欄位      | 型別   | 必填 | 說明                                                   |
| --------- | ------ | ---- | ------------------------------------------------------ |
| `checkId` | uuid   | yes  | assignment 中的 check ID。不可由 request 指定 project。 |
| `type`    | string | yes  | result type，必須和 active assignment 的 check type 相同。 |
| `ping`    | array  | when `type=ping` | ping result payload list。                         |

Ping result 欄位：

| 欄位            | 型別        | 必填 | 說明                                      |
| --------------- | ----------- | ---- | ----------------------------------------- |
| `startedAt`     | timestamp   | yes  | check execution start time。              |
| `finishedAt`    | timestamp   | yes  | check execution finish time。需 >= start。 |
| `durationMs`    | integer     | yes  | execution duration，需 >= 0。             |
| `status`        | string      | yes  | `successful`、`timeout` 或 `error`。       |
| `sentCount`     | integer     | yes  | sent packet count，需 >= 0。              |
| `receivedCount` | integer     | yes  | received packet count，需介於 0 和 sent。 |
| `lossPercent`   | number      | yes  | packet loss percentage，0 到 100。         |
| `rtt*Ms`        | number      | no   | RTT aggregate values，需 >= 0 並符合 min/avg/max order。 |
| `rttSamplesMs`  | number[]    | no   | RTT samples，所有值需 >= 0。              |
| `resolvedIp`    | IP address  | no   | 實際使用的 resolved IP。                  |
| `ipFamily`      | string      | no   | `inet` 或 `inet6`。                       |
| `raw`           | JSON object | no   | executor raw details；省略時視為 `{}`。   |
| `errorCode`     | string      | no   | machine-readable error code。             |
| `errorMessage`  | string      | no   | executor error message。                  |

Validation algorithm：

1. runtime auth 驗證成功後，controller normalize batch shape。
2. `results` 必須非空，最多 100 個 result groups。
3. 每個 group 必須有有效 `checkId`、支援的 `type`，且只能帶和 `type` 對應的非空 payload array。
4. 同一 request 不接受重複 `(checkId, type)` group，也不接受同一 check/type 下重複 `startedAt`。
5. controller 收集所有 unique `checkId`，以單一查詢取得該 probe 的 active assignments。
6. 每個 group 的 `checkId` 都必須存在於 active assignment lookup 結果。
7. 每個 group 的 `type` 必須等於 assignment 的 `check.type`。
8. controller 依照 type 執行 typed validation 與 storage mapping；目前 `ping` 會寫入 `ping_results`。

Identity 與 idempotency：

- request body 不接受 `projectId` 或 `probeId`；storage input 使用 authenticated credential 與 active assignment 的 project/probe/check identity。
- ping result retry idempotency 使用既有 unique key：

```text
project_id, probe_id, check_id, started_at
```

同一筆 result 因 retry 重送時，controller 回傳 `200`，DB 透過 `ON CONFLICT DO NOTHING` 去重。

## 6. Probe Status

目前 status 寫入 `probe_statuses`。

現有欄位：

| 欄位            | 來源       | 說明                                                              |
| --------------- | ---------- | ----------------------------------------------------------------- |
| `status`        | controller | 目前 domain 只有 `online`、`offline`。heartbeat 會寫入 `online`。 |
| `last_seen_at`  | controller | `UpdateProbeStatus` 時由 DB 寫入 `now()`。                        |
| `agent_version` | probe      | probe agent version。                                             |
| `public_v4`     | probe      | public IPv4。                                                     |
| `public_v6`     | probe      | public IPv6。                                                     |
| `as`            | probe      | autonomous system 資訊。                                          |
| `addrs`         | probe      | interface/local/observed IP addresses。                           |
| `updated_at`    | controller | status row update time。                                          |

### 6.1 建議擴充的 Probe Inventory

為了讓 control plane 更理解 probe 的執行環境，未來可以擴充 runtime status 或新增 inventory endpoint。這些欄位目前尚未在 backend schema 實作。

建議資訊：

```json
{
	"architecture": {
		"os": "linux",
		"arch": "amd64",
		"kernel": "6.8.0",
		"containerized": true
	},
	"hardware": {
		"cpuCores": 4,
		"memoryBytes": 8589934592
	},
	"network": {
		"hostname": "edge-probe-01",
		"interfaces": [
			{
				"name": "eth0",
				"mac": "02:42:ac:11:00:02",
				"addrs": ["10.0.0.10"]
			}
		],
		"defaultGateway": "10.0.0.1",
		"dnsServers": ["1.1.1.1", "8.8.8.8"]
	}
}
```

建議分界：

- `agentVersion/publicV4/publicV6/as/addrs` 可以繼續留在 lightweight heartbeat。
- OS、arch、CPU、memory、interfaces、DNS 等 inventory 資訊較大，適合低頻更新，避免每次 heartbeat 都送完整硬體資訊。
- 若要持久化這些資訊，需要新增 schema 或 JSONB 欄位，並補 validation、mapper、repository、OpenAPI 與 tests。

## 7. Probe Runtime Config

`ProbeRuntimeConfig` 是 controller 下發給 probe 的共用 runtime config domain model。`hello.config` 與 assignments response 的 `config` 必須使用同一個 shape。

未來 backend 實作時，建議把 Go domain model 放在 `server/internal/domain/probe`，例如 `RuntimeConfig`。application layer 的 `HelloOutput` 與 `ListAssignmentsOutput` 都應引用同一個 domain model，transport layer 再共用同一個 response mapper。

目前 JSON shape：

```json
{
	"heartbeatIntervalSeconds": 30,
	"assignmentPollIntervalSeconds": 30,
	"maxConcurrentChecks": 16,
	"initialBackoffSeconds": 1,
	"maxBackoffSeconds": 30,
	"maxAttempts": 5
}
```

欄位語意：

| 欄位                            | 說明                                             |
| ------------------------------- | ------------------------------------------------ |
| `heartbeatIntervalSeconds`      | probe 呼叫 heartbeat 的間隔。                    |
| `assignmentPollIntervalSeconds` | probe 拉取 assignments 的間隔。                  |
| `maxConcurrentChecks`           | probe 本機同時執行 checks 的上限。               |
| `initialBackoffSeconds`         | 未來 result submission retry 的初始 backoff。    |
| `maxBackoffSeconds`             | 未來 result submission retry 的最大 backoff。    |
| `maxAttempts`                   | 未來 result submission retry 的最大嘗試次數。    |

目前 config 先由 controller 內建 default provider 產生，不需要新增 database schema。未來若要讓 control plane 控制這些值，可以把 provider 接到 project/probe scoped persistence，但 API response shape 應維持不變。

### 7.1 Runtime Config Policy

目前方向是 local env 只負責 probe bootstrap，runtime 行為設定由 controller response 回傳：`hello` 負責啟動初始化，assignments response 負責 polling 時 refresh。

probe local env 只保留：

```text
NETSTAMP_PROBE_CONTROLLER_URL
NETSTAMP_PROBE_ID
NETSTAMP_PROBE_SECRET
```

`hello` 與 assignments response 的 `config` 管理：

- heartbeat interval
- assignment polling interval
- max concurrent checks
- local result retry/backoff

`hello` 也必須回傳 `minimumSupportedAgentVersion`，讓 probe 在啟動時判斷自己的 agent version 是否仍被 controller 支援。

目前設計下，probe 不應透過 local env 覆蓋 controller 下發的 runtime config。若未來要支援 local override，需要另行設計：

- config persistence schema
- per-probe/project default hierarchy
- local env 與 controller config precedence
- config version
- rollout 與 rollback 行為
- probe 對未知 config 欄位的相容策略

## 8. Probe Agent Main Loop

目前建議 probe agent 的 runtime loop：

```text
start
  ↓
load local env
  ↓
POST /runtime/probes/{probe_id}/hello
  ↓
validate agent version compatibility
  ↓
apply hello runtime config
  ↓
start heartbeat loop
  ↓
start assignment polling loop
  ↓
apply assignments runtime config refresh
  ↓
schedule checks locally by check.intervalSeconds
  ↓
execute ping checks
  ↓
submit grouped results to /runtime/probes/{probe_id}/results
  ↓
retry failed result submissions with configured backoff
```

目前 assignment list 是 active check definition，不是 server lease queue。probe 端需要避免本機對同一 assignment 重複排程，並以 check interval 控制執行頻率。assignments response 也會回傳最新 runtime config，probe 應在每次 polling 後套用。

### 8.1 Probe-side Schedule

probe 端應把 `GET /runtime/probes/{probe_id}/assignments` 回傳的 assignments 當作本機 schedule source，並把同一個 response 裡的 `config` 當作 runtime config refresh。controller 負責告訴 probe「有哪些 active checks 要跑」以及「runtime policy 是什麼」，probe 負責決定「下一次何時跑」。

每個 active assignment 在 probe 本機維護一個 schedule entry：

```text
assignment_id
check_id
check_version
selector_version
interval_seconds
next_due_at
running
```

其中：

- `interval_seconds` 來自 `assignment.check.intervalSeconds`。
- `next_due_at` 是 probe 本機計算的下一次執行時間。
- `running` 用來避免同一個 assignment 在本機重疊執行。
- `check_version` 與 `selector_version` 變更時，probe 應更新本機 entry 的 check config。

### 8.2 Min-heap Scheduler

probe 本機 scheduler 建議用 min-heap 管理 `next_due_at`，heap top 永遠是最早到期的 schedule entry。

主迴圈概念：

```text
loop
  ↓
peek heap top
  ↓
if heap empty: wait for assignment update or shutdown
  ↓
if top.next_due_at is in future: wait until timer or assignment update
  ↓
pop all due entries
  ↓
for each due entry:
  if entry is still active and not running:
    execute check
  compute next future due
  push entry back into heap
```

不能只用 `sleep(next_due_at - now)`，因為 assignment polling 可能拉到新增、刪除或 interval 變更的 assignment。scheduler 等待期間必須能被 assignment update 喚醒，重新檢查 heap top。

### 8.3 Assignment Update Handling

assignment polling loop 每次拿到新的 assignment list 後，應和本機 schedule map 做 sync：

- 新 assignment：建立 schedule entry，設定初始 `next_due_at`，push 到 heap，並 wake scheduler。
- 已存在 assignment：若 check config、check version、selector version 或 interval 改變，更新 entry；若新的 `next_due_at` 可能更早，wake scheduler。
- 消失的 assignment：標記 entry inactive；lazy deletion 即可，之後從 heap pop 出來時略過。

heap 裡可能短暫存在舊 entry，因此 scheduler pop 時必須用 schedule map 驗證 entry 是否仍是最新版本。這可以避免每次 assignment 更新都需要在 heap 中間刪除元素，實作較簡單。

assignment polling loop 也應套用 response 中的 runtime config：

- `assignmentPollIntervalSeconds` 改變時，調整下一輪 assignment polling 的 timer。
- `heartbeatIntervalSeconds` 改變時，調整 heartbeat loop 的 timer。
- `maxConcurrentChecks` 改變時，調整 worker pool 或 semaphore 上限。
- retry/backoff 欄位改變時，更新未來 result submission retry policy。

### 8.4 Next Due 計算

probe 不應補跑所有 missed runs。若 probe 忙碌、暫停、重啟或本機時間跳動造成任務過期，下一次執行時間應直接計算到未來，避免 busy loop。

建議規則：

```text
next_due_at = previous_due_at + interval
while next_due_at <= now:
  next_due_at += interval
```

如果是新 assignment，初始 `next_due_at` 可設為 `now` 或 `now + jitter`。初版可用 `now`，讓新 assignment 盡快執行；未來 probe 數量變多時可加入 jitter，避免大量 probe 同時打同一個 target。

### 8.5 Concurrency

最新 runtime config 的 `maxConcurrentChecks` 是 probe 本機執行 check 的並行上限。scheduler pop 到 due entry 後，若 worker pool 已滿，可以延後該 entry 的執行並重新計算一個短暫 retry due time，或保留在 pending queue。

初版建議：

- 同一 assignment 不重疊執行。
- 全 probe process 受 `maxConcurrentChecks` 限制。
- check 執行超過自身 timeout 後由 executor 結束，不讓 scheduler goroutine 被阻塞。

## 9. 現有資料模型

### 9.1 `probes`

probe registry record。包含 project scope、name、enabled、location、subdivision code、soft delete timestamps。

重要語意：

- `enabled = false` 時 runtime auth 會通過 credential lookup 但 application service 回傳 probe disabled。
- `deleted_at IS NOT NULL` 的 probe 不會被 runtime credential lookup 當作 active probe。

### 9.2 `probe_credentials`

每個 probe 一筆 credential record。

```text
probe_id
secret_hash
created_at
last_rotated_at
```

controller 只儲存 hash，不儲存 plaintext secret。

### 9.3 `probe_statuses`

probe runtime status record。

```text
probe_id
status
last_seen_at
agent_version
public_v4
public_v6
as
addrs
updated_at
```

### 9.4 `checks` 與 `ping_check_configs`

目前 check type 只有 `ping`。

`checks.interval_seconds` 是 probe 本機排程的重要依據。

### 9.5 `probe_check_assignments`

目前 assignment materialization table。

```text
id
project_id
probe_id
check_id
check_version
selector_version
created_at
updated_at
deleted_at
```

active assignment 由 partial unique index 保證：

```text
project_id, probe_id, check_id
WHERE deleted_at IS NULL
```

### 9.6 `ping_results`

ping result storage 已存在，probe runtime result submission 會把 `type=ping` 的 payload 寫入此 table。

目前已有 idempotency 相關 unique index：

```text
project_id, probe_id, check_id, started_at
```

這表示 runtime result submission 使用 `(project_id, probe_id, check_id, started_at)` 作為重試去重基礎，而不是沿用舊草稿的 `assignment_id` 通用 result 模型。

## 10. 錯誤模型

Runtime transport 目前使用 Huma error model。

常見 mapping：

| 條件                                 | HTTP status |
| ------------------------------------ | ----------- |
| 缺少 `Authorization: Probe <secret>` | `401`       |
| secret 錯誤                          | `401`       |
| probe disabled                       | `403`       |
| probe 不存在或已刪除                 | `404`       |
| runtime status input invalid         | `422`       |
| result payload invalid               | `422`       |
| result check 不是 active assignment  | `422`       |
| result type 和 assigned check type 不一致 | `422`  |
| unexpected technical failure         | `500`       |

`401` response 會帶：

```http
WWW-Authenticate: Probe
```

## 11. Observability

Probe runtime application event 目前記錄 failure：

- `probe_runtime.hello.failure`
- `probe_runtime.heartbeat.failure`
- `probe_runtime.assignments.list.failure`
- `probe_runtime.results.submit.failure`

成功的 hello、heartbeat、assignment polling、result submission 由 HTTP request logger 覆蓋。

Probe runtime logs 不應包含：

- plaintext probe secret
- secret hash
- raw result body
- raw target
- selector text
- IP address 或 agent version 等可能敏感資訊，除非未來明確調整 privacy policy

Tracing 應維持目前分層：

```text
transport handler
  → application proberuntime service span
  → postgres repository span
```

## 12. Future Work

以下內容保留為未來方向，不屬於目前 backend 現況。

### 12.1 Additional Typed Result Storage

目前 result submission route 已存在，且 v1 支援 `type=ping` 寫入 `ping_results`。未來新增 HTTP/TCP/DNS 等 check type 時，應延續目前 grouped payload 與 type dispatch 模型：

- 新增 typed payload array，例如 `http[]`、`tcp[]` 或 `dns[]`。
- 新增對應 domain validation、repository port、storage schema 與 idempotency key。
- active assignment lookup 與 `type == assignment.check.type` 規則維持不變。

### 12.2 Offline Detection

目前 heartbeat 會寫入 online 與 `last_seen_at`，但本文不宣稱已有 offline detector。

未來可由 background job 或 read-time derived status 判斷：

```text
now - last_seen_at > heartbeat_interval * threshold
```

需要先決定：

- threshold 來源，local env 或 controller config
- 是否寫回 `probe_statuses.status = offline`
- 是否影響 assignment list

### 12.3 Persisted Controller-Managed Runtime Config

目前 `hello` 與 assignments response 已保留 `ProbeRuntimeConfig` 的 wire shape，但 config 值先由 controller default provider 產生。若未來要讓 control plane 持久化管理這些值，需要新增 project/probe scoped config storage、validation 與 rollout 行為。

### 12.4 Assignment Lease

目前 `probe_check_assignments` 是 active check materialization，不是 queue lease model。

若未來要做 server-side dispatch lease，需要新增：

- assignment execution table
- state transition
- lease expiry
- duplicate execution policy
- cleanup/retry behavior

### 12.5 Server-Side Scheduler

目前排程建議先在 probe 端根據 `check.intervalSeconds` 與 min-heap 執行。controller-side heap scheduler、wake channel、due task queue 可留作未來 server-side dispatch scheduler 的參考，但不屬於目前 runtime plane 現況。

## 13. 驗收標準

文件與現有 backend 對齊時，應滿足：

- runtime endpoint 只列出目前 Huma route 已註冊的四個 API：hello、heartbeat、assignments、results。
- `hello` 不帶 request body，response 欄位對齊 `hello.go`。
- `hello` 與 assignments response 都回傳相同 shape 的 `ProbeRuntimeConfig`。
- `heartbeat` request body 承載 `runtime_status.go` 定義的 status 欄位。
- authentication 說明使用 `Authorization: Probe <secret>`。
- probe status 說明對齊 `domain/probe.Status` 與 `probe_statuses`。
- assignment 說明對齊 `domain/assignment.Assignment` 與 `probe_check_assignments`。
- result submission 說明使用 grouped payload，並對齊 `ping_results` typed storage。
- 自創 `tasks`、通用 `assignments`、通用 `results` SQL 不再出現在現況章節。
- lease、offline detector、persisted controller-managed config 都清楚標示為 Future Work。
