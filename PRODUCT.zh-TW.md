# Netstamp 產品與技術文件（中文版）

> 分散式網路可觀測性平台 — 從你所掌控的探針（probe），測量真實世界中的網路可達性、延遲、封包遺失與路由路徑。
>
> 對應英文版本：[PRODUCT.en.md](./PRODUCT.en.md)

---

## 目錄

1. [產品概述](#1-產品概述)
2. [它解決什麼問題](#2-它解決什麼問題)
3. [目標使用者與使用情境](#3-目標使用者與使用情境)
4. [核心概念與領域模型](#4-核心概念與領域模型)
5. [系統架構](#5-系統架構)
6. [實現原理](#6-實現原理)
7. [資料模型與時序儲存](#7-資料模型與時序儲存)
8. [API 設計](#8-api-設計)
9. [技術棧](#9-技術棧)
10. [部署架構](#10-部署架構)
11. [安全性設計](#11-安全性設計)
12. [可觀測性](#12-可觀測性)
13. [擴展性與效能](#13-擴展性與效能)
14. [關鍵設計決策與權衡](#14-關鍵設計決策與權衡)
15. [詞彙表](#15-詞彙表)

---

## 1. 產品概述

**Netstamp** 是一套開源、可自架的**分散式網路可觀測性平台**。它讓團隊把測量探針（probe）部署到任何網路位置——雲端機房、ISP 邊緣節點、實驗室、私有基礎設施或不同地理區域——並從這些「視角」持續測量目標服務的網路行為。

一句話定位：

> **「在網路出問題之前看見它」——從你自己掌控的探針，量測延遲、封包遺失與路由路徑。**

平台由五個部分組成：

| 元件 | 角色 | 技術 |
| --- | --- | --- |
| **Controller（控制器）** | 後端 API：認證、授權、專案、探針、檢查、指派、結果儲存 | Go + chi + PostgreSQL/TimescaleDB |
| **Probe Agent（探針代理）** | 部署在被測網路附近，執行測量並回報結果 | Go（可裝成 Linux systemd 服務） |
| **Web App（操作介面）** | 已認證的操作者主控台 | React 19 + Vite |
| **Docs Site（文件站）** | 公開文件、API Explorer、Landing Page | Astro + MDX |
| **`@netstamp/ui`** | 跨介面共享的 React 元件與設計 token | React + Storybook |

核心心智模型：**一個專案（project）擁有探針、標籤、檢查與成員；探針認證後向控制器領取「指派（assignment）」，執行檢查，並回傳時序測量結果供分析。**

---

## 2. 它解決什麼問題

傳統的「單點」監控（例如只從一台機器或單一雲端區域 ping 目標）無法回答現代分散式系統真正關心的問題：

- **「我的服務從不同地區、不同 ISP 看起來如何？」** — 單點探測看不到地理與網路路徑差異。
- **「使用者回報變慢，是我的服務、還是中間的某一跳（hop）出問題？」** — 沒有 traceroute 拓撲就難以定位。
- **「封包遺失與延遲抖動發生在哪個網段？」** — 需要多視角、長時間的時序資料。
- **「路由路徑什麼時候改變了？是否與事故相關？」** — 需要路徑雜湊（path hash）的歷史比對。
- **「我能不能完全掌控探測點、不依賴第三方 SaaS 黑盒？」** — 需要可自架、開源、資料自有。

Netstamp 的回答：

1. **多視角測量** — 把探針部署到任意位置，從真實網路邊緣量測，而非單一中心點。
2. **三種測量類型** — ICMP Ping（可達性／延遲／遺失）、TCP Connect（連接埠連通性與握手延遲）、Traceroute（逐跳路由拓撲）。
3. **標籤驅動的指派** — 用 selector 表達式自動決定「哪些探針執行哪些檢查」，不必手動逐一綁定。
4. **時序分析** — 以 TimescaleDB 儲存高頻測量，並透過連續聚合（continuous aggregate）提供秒級的儀表板查詢。
5. **完全自有、開源** — 探針、控制器、資料庫全部自架，資料不離開你的基礎設施。
6. **公開狀態頁（Public Pages）** — 可選擇性對外公開部分檢查的健康狀態。

---

## 3. 目標使用者與使用情境

**主要使用者**

- **平台／SRE 團隊**：需要從多個地理與網路位置監看自家服務的對外可達性。
- **網路工程師**：需要 traceroute 拓撲與路由變化偵測來定位跨網段問題。
- **基礎設施擁有者**：希望自架、資料自有、不依賴外部 SaaS 的觀測方案。

**典型情境**

- 在台灣北部、日本、美西各放一台探針，持續 ping 你的 API 端點，比較三地延遲與遺失率。
- 對關鍵第三方依賴（支付閘道、DNS、CDN 邊緣）設定 TCP 檢查，量測握手延遲變化。
- 對核心目標啟用 traceroute，當路由路徑雜湊改變時即時察覺，並在拓撲圖上看到是哪一跳劣化。
- 建立公開狀態頁，對客戶展示關鍵服務的健康度。

---

## 4. 核心概念與領域模型

理解 Netstamp 的關鍵在於這幾個領域物件之間的關係：

```text
User ──< ProjectMember >── Project
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
      Label                 Probe                  Check
        │  (key/value)        │ (測量視角)            │ (要測什麼)
        │                     │                     │
        ├── probe_labels ─────┤                     ├── ping/tcp/traceroute config
        └── check_labels ─────┼─────────────────────┤
                              │                     │
                              │   selector 匹配      │
                              ▼                     ▼
                       ProbeCheckAssignment（指派）
                              │  (check_version, selector_version)
                              ▼
                   Probe 執行 → PingResult / TCPResult / TracerouteResult（時序）
```

| 概念 | 說明 |
| --- | --- |
| **User（使用者）** | 帳號。密碼以 Argon2id 雜湊儲存。 |
| **Project（專案）** | 工作區，是所有資源的邊界。以 `slug` 作為對外參照。支援軟刪除。 |
| **ProjectMember（成員）** | 使用者在專案中的角色：`owner`／`admin`／`editor`／`viewer`。 |
| **ProjectInvite（邀請）** | 邀請使用者加入專案，狀態 `pending`／`accepted`／`rejected`。 |
| **Label（標籤）** | 專案範圍內的 key/value 對。掛在探針與檢查上，是 selector 匹配的基礎。 |
| **Probe（探針）** | 一個測量視角。含名稱、啟用狀態、地理座標、位置名稱、標籤、執行時狀態。對外用 UUID，內部時序資料用緊湊的 `internal_id`。 |
| **Probe Credential（憑證）** | 探針的密鑰雜湊。明文僅在建立或輪替時回傳一次。 |
| **Probe Status（狀態）** | 探針線上／離線、最後心跳時間、agent 版本、對外 IP、AS 號等。 |
| **Check（檢查）** | 定義「要測什麼」。型別為 `ping`／`tcp`／`traceroute`，含目標、間隔秒數、selector 與型別專屬設定。 |
| **Selector（選擇器）** | JSON 表達式，描述「哪些探針符合」。支援 `all`／`any`／`not`／`label`（`eq`／`in`／`exists`）。空 selector 匹配全部探針。 |
| **Assignment（指派）** | 由 selector × probe labels 計算得出的「某探針要跑某檢查」的事實，帶有 `check_version` 與 `selector_version` 版本雜湊。 |
| **Result（結果）** | 探針回傳的時序測量資料：ping/tcp/traceroute 各有結構，存於 TimescaleDB hypertable。 |
| **Public Page（公開頁）** | 對外公開的狀態頁，可用樹狀資料夾組織要展示的檢查。 |

---

## 5. 系統架構

### 5.1 四個執行面（Runtime Surfaces）

```text
                        ┌──────────────────────────────┐
   操作者瀏覽器  ───────▶ │  Web App (React 19 + Vite)    │
                        │  已認證操作主控台              │
                        └──────────────┬───────────────┘
                                       │  cookie session (JWT)
                                       │  /api/v1/*
                                       ▼
   公開訪客 ──────▶ Docs Site ──▶ ┌──────────────────────────────┐      ┌──────────────────────┐
   (Astro + MDX)   API Explorer   │  Controller API (Go + chi)   │◀────▶│ PostgreSQL/TimescaleDB │
                                  │  transport→application→domain │      │ 關聯狀態 + 時序結果     │
                                  │  →infrastructure→observability│      └──────────────────────┘
                                  └──────────────┬───────────────┘
                                       ▲          │
              Authorization: Probe ... │          │ 指派、結果驗證
                                       │          ▼
                        ┌──────────────────────────────┐
   被測網路附近 ─────────│  Probe Agent (Go)             │
                        │  hello/heartbeat/assignments/ │
                        │  results；執行 ping/tcp/trace  │
                        └──────────────────────────────┘
```

兩條主要請求路徑：

```text
操作者：Browser → React App → /api/v1/* → chi → application service → PostgreSQL/TimescaleDB
探針：  Probe  → /api/v1/runtime/probes/{probe_id}/* → runtime service → assignments + results
```

> 注意：探針執行時端點的實際路徑為 `/api/v1/runtime/probes/{probe_id}/{hello|heartbeat|assignments|results}`。

### 5.2 後端分層架構（Controller）

控制器採用嚴格的分層式（layered / hexagonal 風格）架構，依賴方向**單向向下**：

```text
┌─────────────────────────────────────────────────────────┐
│ transport/http                                          │
│   路由註冊、middleware 堆疊、請求/回應 DTO、HTTP 錯誤映射  │
├─────────────────────────────────────────────────────────┤
│ application/*                                           │
│   use case、授權決策、流程編排、事件語義、輸入驗證        │
│   每個領域：service.go / flow.go / ports.go /            │
│            validate.go / dto.go / errors.go / trace.go  │
├─────────────────────────────────────────────────────────┤
│ domain/*                                                │
│   穩定領域模型、權限政策、selector 解析、版本雜湊、       │
│   驗證-正規化（VN*）輔助函式                              │
├─────────────────────────────────────────────────────────┤
│ infrastructure/*                                        │
│   PostgreSQL repository（sqlc 生成）、JWT、Argon2id、     │
│   探針密鑰                                                │
├─────────────────────────────────────────────────────────┤
│ platform/observability                                  │
│   metrics（Prometheus）、tracing（OTLP）、HTTP trace 輔助 │
└─────────────────────────────────────────────────────────┘
```

**依賴方向**：`transport → application → domain ← infrastructure`，`platform` 提供橫切的可觀測性能力。授權決策由 application 層負責；HTTP middleware 只證明身分，角色政策在 application 服務中以 domain 政策函式強制執行。

### 5.3 Application 層的標準分檔模式

每個領域（auth、project、label、check、probe、proberuntime、assignment、result、publicpage、user）都遵循同一套檔案分工，這是本專案最具特色的設計慣例：

| 檔案 | 職責 |
| --- | --- |
| `service.go` | 核心業務邏輯，公開方法入口 |
| `flow.go` | 執行流程管理：建立 OpenTelemetry span、記錄應用事件、區分「業務失敗」與「技術失敗」 |
| `ports.go` | 介面定義（repository、event recorder）與事件常數 |
| `validate.go` | 輸入正規化與驗證（normalize + validate 模式） |
| `dto.go` | 資料傳輸物件 |
| `errors.go` | 領域錯誤定義 |
| `trace.go` | tracer 與 span 屬性 |

**業務失敗 vs 技術失敗** 的區分很關鍵：`flow.businessFailure()` 用於已知的領域錯誤（如「使用者已存在」「無權限」），不會把 span 標成 error；`flow.technicalFailure()` 用於系統性錯誤（如資料庫斷線），會把 span 標成 error 狀態。這讓追蹤與告警能正確分辨「預期內的拒絕」與「真正的故障」。

---

## 6. 實現原理

### 6.1 啟動與生命週期

控制器啟動分為五個階段（`server/cmd/controller/main.go` → `app/bootstrap.go` → `app/lifecycle.go`）：

1. **訊號上下文**：以 `signal.NotifyContext` 建立可接收 SIGINT/SIGTERM 的 context。
2. **依賴組裝**（bootstrap）：載入設定 → 初始化 Zap logger → 初始化 metrics/tracing → 建立 pgx 連線池 → 實例化 security（Argon2id hasher、JWT issuer）、repository、application service → 組裝 HTTP router。
3. **啟動 HTTP server**：用 `errgroup` 同時跑「HTTP serve」與「監聽 context 取消」兩個 goroutine。
4. **優雅關閉**：依序關閉 HTTP server → DB 連線池 → metrics provider → tracing provider（flush 待送 span）。

### 6.2 認證與授權

**使用者認證**

- 密碼以 **Argon2id** 雜湊（記憶體 64 MiB、迭代 3、平行度 4，16-byte salt、32-byte hash），比對時用 `subtle.ConstantTimeCompare` 防時序攻擊。
- 登入後簽發 **JWT（HS256）**，以 `AUTH_JWT_SECRET` 簽章，claims 含 `sub`（user id）、`email`、`iss`/`aud`（皆為 `netstamp`）、`iat`/`nbf`/`exp`。
- Token 放在 **HTTP-only `netstamp_session` cookie**，非 local 環境強制 `Secure`。前端透過 `credentials: "include"` 自動帶上。
- middleware `RequireAuth` 從 cookie 取出並驗證 JWT，將 claims 注入 request context。

**專案授權（RBAC）**

- domain 層提供決策函式（`Can(role, action)`）：
  - `viewer`：唯讀；
  - `editor`：可建立/更新標籤、檢查、探針；
  - `admin`：專案寫入 + 成員管理（除 owner 級角色變更）；
  - `owner`：完全控制，含刪除專案。
- application service 在每個操作前，先查詢使用者在該專案的角色，再呼叫 domain 決策函式；拒絕時回 `ErrForbidden`，由 transport 映射為 403。

### 6.3 Selector 引擎與指派計算

這是 Netstamp 的核心「自動化」機制。使用者不必逐一把檢查綁到探針上，而是寫一段 selector 表達式描述「想要哪種探針」：

```json
{
  "all": [
    { "label": { "key": "region", "op": "eq", "value": "tw-north" } },
    { "label": { "key": "network", "op": "in", "values": ["fiber", "ix"] } }
  ]
}
```

支援的節點：`all`（全部成立）、`any`（任一成立）、`not`（反向）、`label`（針對標籤做 `eq`／`in`／`exists`）。空 selector 匹配所有探針。

**指派（assignment）如何產生與維護**：

- 當**探針更新**、**檢查更新**或**標籤更新**時，application 的 assignment 服務會重新計算受影響範圍的指派：對每個相關 (probe, check) 配對評估 selector 是否匹配，匹配則 upsert 一筆 `probe_check_assignments`，不匹配則軟刪除。
- 每筆指派儲存兩個版本雜湊：
  - `check_version` = 檢查內容（target、interval、型別設定）的 SHA256；**名稱與描述不影響版本**（因為不影響執行）。
  - `selector_version` = selector JSON 的 SHA256。
- 控制器另提供 **selector 預覽**端點，讓使用者在建立檢查時即時看到「會匹配到哪些探針、共幾個」。

### 6.4 探針執行時生命週期（Probe Agent）

探針是一個獨立的 Go 程式，可用 `netstamp-agent run` 執行或裝成 systemd 服務。其執行流程：

```text
啟動 → Hello（認證 + 取得伺服器時間/最小版本/設定）
        │（指數退避重試；認證失敗直接終止）
        ▼
  errgroup 並發跑五個迴圈：
  ├─ heartbeatLoop      每隔 HeartbeatInterval 回報狀態（agent 版本、本地 IP）
  ├─ assignmentLoop     每隔 PollInterval 拉取指派 → Reconcile 比對新舊
  ├─ Scheduler.Run      以最小堆（min-heap）按到期時間排程各檢查
  ├─ Workers.Run        worker pool 並發執行檢查
  └─ Submitter.Run      批次上傳結果
        │
        ▼
  context 取消 → 等待 ShutdownTimeout，盡力上傳剩餘結果 → 結束
```

**關鍵設計**：

- **版本同步（resync）**：探針本地以 `generation` 編號每筆任務。當 `check_version` 或 `selector_version` 改變時 generation++，重新排程；不在新列表中的任務標記為 removed。控制器若收到過時版本的結果，仍接受有效結果，並請探針 resync。
- **相位抖動（phase jitter）**：用 `FNV-1a(probeID + assignmentID)` 算出 0–59 秒的偏移，避免所有探針在整點同時打同一目標，分散測量負載。
- **任務 TTL**：若距上次成功拉取指派超過 `AssignmentTTL`，停止排程，避免在與控制器失聯時繼續執行過期任務。
- **結果佇列滿載策略**：結果佇列（預設容量 10000）滿時**丟棄最舊的**，優先保留最新測量資料。
- **重試**：hello/heartbeat/results 皆有指數退避重試（initial → ×2 → max backoff，最多 MaxAttempts 次）；認證失敗或永久性 4xx 直接終止。

### 6.5 三種測量類型的執行

| 類型 | 如何執行 | 輸出指標 | 權限 |
| --- | --- | --- | --- |
| **Ping（ICMP）** | 開 ICMP raw socket，依設定發送 N 個 echo request、記錄序號與時間、並行收 reply | duration、sent/received count、loss%、RTT min/avg/median/max/stddev、RTT 樣本陣列、resolved IP、IP family、狀態 | 需 `CAP_NET_RAW` |
| **TCP** | 用 `net.Dialer` 對 `host:port` 做 TCP connect，記錄握手耗時後立即關閉 | duration、connect duration、resolved IP、IP family、狀態 | 不需特權 |
| **Traceroute** | ICMP 或 UDP，逐跳遞增 TTL，每跳重複 N 次 | destination reached、hop count、每跳的 address/hostname/loss/RTT 統計與樣本、整體狀態（successful/partial/timeout） | ICMP 模式需 `CAP_NET_RAW` |

### 6.6 結果提交、驗證與冪等性

探針把結果**批次**送到 `POST /runtime/probes/{probe_id}/results`。控制器的驗證鏈：

1. **認證探針**（`Authorization: Probe <secret>`）。
2. **正規化與結構驗證**：結果不可為空；同一 (checkId, type) 不可有重複群組；同一檢查內不可有重複 `startedAt`。
3. **指派核對**：每個 checkId 必須有對應 active assignment，且型別相符。
4. **時序與數值驗證**：時間順序、loss% 範圍、RTT 排序、resolved IP、IP family、原始 JSON payload。
5. **分型別寫入** ping/tcp/traceroute 結果表。

**冪等性鍵**為 (project, probe, check, startedAt) 的組合（結果表以 `(probe_id, check_id, started_at)` 為主鍵）。即使探針重送相同結果也不會產生重複；若送來過時版本的指派，控制器接受有效結果並回應請探針 resync。

### 6.7 前端資料流

Web App 採 **feature-based** 架構（`web/src/features/*`），共享層在 `web/src/shared/*`：

- **型別安全 API**：API 合約以 TypeSpec 撰寫 → 產生 OpenAPI → 用 `openapi-typescript` 生成 TS 型別 → `openapi-fetch` 建立強型別 client（`credentials: "include"` 帶 cookie）。
- **資料抓取**：以 **TanStack Query（React Query）** 為中心，query 自動去重與快取（測量類 staleTime ~30s、Insight 頁自動每 15s refetch），mutation 成功後 `invalidateQueries` 並推 toast。
- **URL 狀態**：Insight 頁把時間範圍與篩選編碼進 query string，支援分享與瀏覽器前後退。
- **視覺化**：**ECharts** 畫時序圖（可拖曳選時間範圍）、**MapLibre GL** 畫探針地理分佈地圖、自製 SVG 引擎畫 traceroute 路由拓撲。
- **認證流程**：`SessionProvider` 呼叫 `GET /auth/me` 同步登入狀態；`ProtectedAppShell` 守衛受保護路由，未登入導向 `/login`；首次註冊導向 `/onboarding` 建立初始專案。
- **專案切換**：`useCurrentProject` 以 localStorage 記住選定專案（類似 workspace switcher）。
- UI 一律優先用 `@netstamp/ui` 共享元件（Button、DataTable、Panel、MetricCard、Badge…），維持跨介面一致的「網路操作主控台」視覺語言。

---

## 7. 資料模型與時序儲存

Netstamp 在 PostgreSQL（啟用 TimescaleDB）中同時儲存**關聯狀態**與**時序結果**。

### 7.1 表的分類（約 38 張表/物化視圖）

| 層級 | 代表表 | TimescaleDB | 保留 |
| --- | --- | --- | --- |
| **業務層** | `users`、`projects`、`project_members`、`project_invites`、`labels`、`probes`、`probe_credentials`、`probe_statuses`、`probe_labels`、`checks`、`ping/tcp/traceroute_check_configs`、`check_labels`、`probe_check_assignments` | 否 | 永久（軟刪除） |
| **結果層** | `ping_results`、`tcp_results`、`traceroute_results`、`traceroute_result_hops` | 是（hypertable） | raw 3 天 |
| **觀測層** | `ping_rtt_sample_observations`、`traceroute_hop_observations`、`traceroute_edge_observations` | 是（hypertable） | 3 天 |
| **聚合層** | 各結果的連續聚合（1m/10m/15m/30m/1h） | 是（continuous aggregate） | 30–180 天 |
| **公開頁** | `public_pages`、`public_page_folders`、`public_page_folder_checks` | 否 | 永久（軟刪除） |

### 7.2 關鍵設計：UUID 對外、internal_id 對內

`probes` 與 `checks` 同時擁有 `id`（UUID，對外 API 用）與 `internal_id`（bigint identity，時序結果表外鍵用）。原因：時序表動輒上億列，用緊湊的 bigint 作外鍵與主鍵能大幅節省儲存並加速查詢。

### 7.3 Hypertable 與保留

- `*_results` 與觀測表皆為 hypertable，依 `started_at` 分區。
- **chunk interval = 1 天**（從預設 7 天調小，利於更細粒度壓縮與查詢）。
- **raw 結果保留 3 天**：近期保留高精度原始資料，超過自動刪除舊 chunk。長期趨勢由連續聚合承擔。

### 7.4 連續聚合（Continuous Aggregates）

針對 ping/tcp/traceroute 各建立 **1m、10m、15m、30m、1h** 五個時間尺度的連續聚合，用 `time_bucket` 預先降採樣。每個聚合包含計數（成功/逾時/錯誤）、duration 統計、loss、RTT 統計（min/avg/max/stddev 的 sum+count）等。

- **刷新策略**：`start_offset` 約 3 天（只刷新近期窗口）、`end_offset` 略小於即時（容許數分鐘延遲）、`schedule_interval` 對應 bucket 寬度。
- **保留**：1m→30 天、10m/15m→90 天、30m→180 天、1h 最久。

目的：**儀表板與 Insight 頁查長時間範圍時，直接讀預先聚合的物化結果，避免掃描海量 raw 列**，達到秒級回應。

### 7.5 觀測表的用途

- `ping_rtt_sample_observations`：把 `ping_results.rtt_samples_ms[]` 陣列「攤平」成每行一個樣本，方便用 `percentile_cont` 算 p50/p95/p99 與 stddev，並產生 RTT 分佈直方圖（latency heatmap）。
- `traceroute_hop_observations` / `traceroute_edge_observations`：由 trigger 從 traceroute hop 自動產生「節點」與「相鄰跳之間的邊」觀測，預先計算拓撲圖所需的節點健康度與邊品質，避免查詢時即時 JOIN。

### 7.6 sqlc 工作流

SQL 查詢寫在 `server/db/query/*.sql`，由 **sqlc** 生成型別安全的 Go 程式碼至 `server/internal/controller/infrastructure/postgres/sqlc/`。migration 由 **Goose** 管理，放在 `server/db/migrations/`。複雜的時序查詢（如 `GetPingInsightSummary` 用 CTE + `percentile_cont`、traceroute 用 `string_agg` + `lag()` window function 偵測路徑變化）都在 query 檔中以原生 SQL 表達。

---

## 8. API 設計

### 8.1 合約優先（Contract-First）

API 合約以 **TypeSpec** 撰寫（`api/` 目錄，依 models 與 services 分檔），透過 `pnpm generate:openapi` 產生 OpenAPI 至 `docs/public/openapi.json`，並由控制器在 docs 站的 API Explorer（`/openapi/`）提供。前端的 TS 型別也由同一份 OpenAPI 生成，確保前後端契約一致。

### 8.2 兩套認證機制

- **使用者路由**：HTTP-only `netstamp_session` cookie（內含簽章 JWT）。
- **探針執行時路由**：`Authorization: Probe <secret>` 標頭，與使用者 JWT 完全分離。

### 8.3 主要端點（皆掛在 `/api/{API_VERSION}`，預設 `/api/v1`）

**系統**
- `GET /`、`GET /healthz`（檢查 DB 連線）、`GET /metrics`（Prometheus，掛在 `/metrics` 不帶版本前綴）

**認證**
- `POST /auth/register`、`POST /auth/login`、`GET /auth/me`

**專案與成員**
- `GET|POST /projects`、`GET|PATCH|DELETE /projects/{ref}`
- `GET|POST /projects/{ref}/members`、`PATCH|DELETE /projects/{ref}/members/{user_id}`
- 專案邀請相關端點（pending/accept/reject）

**標籤**
- `GET|POST /projects/{ref}/labels`、`PATCH|DELETE /projects/{ref}/labels/{label_id}`

**檢查**
- `GET|POST /projects/{ref}/checks`、`GET|PATCH|DELETE /projects/{ref}/checks/{check_id}`
- selector 預覽端點

**探針**
- `GET|POST /projects/{ref}/probes`、`GET|PATCH|DELETE /projects/{ref}/probes/{probe_id}`
- `POST /projects/{ref}/probes/{probe_id}/secret-rotations`（密鑰輪替，明文僅回傳一次）

**結果與測量**
- 專案層的 measurements 與各型別 insight 查詢端點

**探針執行時**（Probe 認證）
- `POST /runtime/probes/{probe_id}/hello`
- `POST /runtime/probes/{probe_id}/heartbeat`
- `GET  /runtime/probes/{probe_id}/assignments`
- `POST /runtime/probes/{probe_id}/results`

**公開頁**（部分無需使用者認證）
- 專案層管理端點 + 以 slug 取得公開頁的端點

**安裝資產**
- `GET /api/v1/install/agent.sh`、`/uninstall-agent.sh`、`/netstamp-agent-linux-amd64`

### 8.4 錯誤格式

採 **RFC 7807 Problem Details**（`application/problem+json`），含 `type`/`title`/`status`/`detail`/`instance` 與驗證錯誤的 `errors[]`（每項含 `message`/`location`/`value`），並在標頭回傳 `X-Request-ID`。對外回應保守，技術細節只進日誌與 trace。

---

## 9. 技術棧

**後端（Controller / Agent）**
- Go、chi（HTTP 路由）、pgx（PostgreSQL driver）、sqlc（型別安全查詢）、Goose（migration）
- PostgreSQL + TimescaleDB、Viper（設定）、Zap（結構化日誌）、OpenTelemetry（trace）、Prometheus（metrics）
- Cobra（agent CLI）

**前端與文件**
- pnpm workspace、React 19、React Router、Vite、TypeScript
- TanStack Query、openapi-fetch / openapi-typescript
- ECharts（時序圖）、MapLibre GL（地圖）
- Astro + MDX（docs 站）、Storybook（`@netstamp/ui`）

**部署與可觀測性**
- Docker / Docker Compose、Nginx
- VictoriaMetrics（metrics）、VictoriaTraces（traces）、VictoriaLogs + Vector（logs）、Grafana（儀表板）

---

## 10. 部署架構

### 10.1 開發環境

```bash
pnpm install
cp server/.env.example server/.env
cp server/probe.env.example server/probe.env
docker compose -f deployments/docker/compose.backend.dev.yaml up -d postgres victoria-traces victoria-metrics grafana
just backend-migrate-up   # 套用 Goose migration
just backend-dev          # 啟動控制器（Air 熱重載）
just web-dev              # 啟動 Web App
just docs-dev             # 啟動 docs 站
just backend-probe server/probe.env   # 跑一個探針
```

### 10.2 生產環境

`deployments/docker/compose.yaml` 會建置：

- 控制器映像（`server/Dockerfile`）
- migration job（用同一映像在控制器啟動前套用 Goose migration）
- Linux amd64 探針 agent 二進位（由控制器的 install 端點提供下載）
- TimescaleDB
- Nginx（服務 web 與 docs 靜態資產）

```bash
docker compose -f deployments/docker/compose.yaml up -d --build
```

必要的生產環境變數：`DATABASE_PASSWORD`、`AUTH_JWT_SECRET`、`LOG_PSEUDONYM_KEY`（以及跑可觀測性堆疊時的 `GF_SECURITY_ADMIN_PASSWORD`）。

### 10.3 在 Linux 主機安裝探針

```bash
curl -fsSL https://example.com/api/v1/install/agent.sh | sudo sh
sudo netstamp-agent service install \
  --url https://example.com \
  --probe-id <probe-id> \
  --probe-secret <probe-secret>
```

`service install` 會建立 `netstamp` 系統使用者、寫入 `/etc/netstamp/probe.env`，並啟用 `netstamp-agent.service`。systemd unit 以非 root 執行，僅授予 `CAP_NET_RAW`（ICMP 所需），並啟用 `NoNewPrivileges`、`PrivateTmp`、`ProtectHome`、`ProtectSystem` 等沙盒選項。卸載可加 `--purge` 連設定與系統使用者一併移除。

---

## 11. 安全性設計

- **密碼**：Argon2id 雜湊，常數時間比對。
- **使用者工作階段**：HTTP-only `netstamp_session` cookie，內含 HS256 JWT；非 local 環境強制 `Secure`。
- **探針認證**：每探針獨立密鑰，僅存雜湊；明文僅在建立／輪替時回傳一次。與使用者 JWT 完全分離。
- **授權**：以 domain 角色政策在 application 層強制（owner/admin/editor/viewer）。
- **軟刪除隔離**：已軟刪除的專案、標籤、檢查、探針排除於正常存取路徑。
- **錯誤資訊**：對外保守（不洩漏內部細節），技術細節只進日誌與 trace。
- **隱私**：日誌使用 `LOG_PSEUDONYM_KEY` 做隱私保護的化名。
- **祕密管理**：切勿提交生產 `.env`、JWT secret、DB 密碼、探針密鑰或帶憑證的 telemetry 端點。
- **前端追蹤同意**：支援 `regional`/`always`/`never` 同意門檻，依訪客國別（多重來源解析）決定是否徵求同意，符合 EEA/UK/瑞士等地區要求。

---

## 12. 可觀測性

控制器內建三大可觀測性支柱：

- **Metrics**：Prometheus 相容，於 `/metrics` 暴露。
- **Traces**：OpenTelemetry，自動（HTTP middleware span）+ 手動（application flow span），可經 `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` 匯出。
- **Logs**：Zap 結構化日誌，記錄 auth、專案、標籤、檢查、探針、執行時工作流的應用事件，並依狀態碼分級（5xx→Error、4xx→Warn、2xx→Info）。

本機可觀測性堆疊（dev compose）：VictoriaMetrics（`:8428`）、VictoriaTraces（`:10428`）、Grafana（`:3000`，預設 `admin`/`admin`）。Grafana 預先佈建 `Netstamp Controller Status` 儀表板並設為首頁。

---

## 13. 擴展性與效能

- **多視角水平擴展**：探針是無狀態的測量單元，可任意增加部署位置；控制器負責協調與儲存。
- **相位抖動**分散測量負載，避免整點尖峰。
- **時序分層儲存**：raw 高精度短期保留（3 天）+ 連續聚合長期保留（30–180 天），在儲存成本與查詢精度間取得平衡。
- **查詢效能**：儀表板/Insight 直接讀連續聚合的物化結果；hypertable chunk 排除加速時間範圍查詢；觀測表預先攤平樣本與拓撲，避免昂貴的即時運算。
- **批次與佇列**：探針端結果批次上傳、佇列滿載丟舊保新，確保在突發或失聯後恢復時仍以最新資料為優先。
- **冪等寫入**：以 (probe, check, startedAt) 為鍵，重送不重複。

---

## 14. 關鍵設計決策與權衡

| 決策 | 理由 | 權衡 |
| --- | --- | --- |
| **後端嚴格分層 + application 持有授權** | 領域邏輯與授權集中、可測試、可替換儲存 | 樣板較多（每領域 7 個檔案） |
| **selector 表達式驅動指派** | 標籤驅動自動化，新增探針即自動納入符合的檢查 | 需要版本雜湊與 reconcile 機制維持一致 |
| **UUID 對外 + internal_id 對內** | 時序表用緊湊 bigint 大幅省儲存、加速 | 雙鍵增加少量複雜度 |
| **raw 短保留 + 連續聚合長保留** | 儀表板秒級回應、控制儲存成本 | 超過 3 天無法看 raw 逐筆樣本 |
| **探針佇列滿載丟舊保新** | 失聯恢復後優先呈現最新狀態 | 極端壅塞下會遺失部分歷史點 |
| **TypeSpec 合約優先** | 前後端型別、docs、Explorer 單一真實來源 | 改路由需重新生成 OpenAPI |
| **使用者 JWT 與探針密鑰雙軌** | 兩類主體（人/機器）安全模型分離 | 兩套認證需各自維護 |
| **預設輕量 TimescaleDB 映像** | 僅用核心 hypertable 與 `time_bucket`，免裝 Toolkit | 降採樣用 `time_bucket` 而非 `lttb` |

---

## 15. 詞彙表

- **Probe（探針）**：部署在某網路位置、執行測量並回報的 agent。
- **Check（檢查）**：定義「要測什麼」的設定（ping/tcp/traceroute + 目標 + 間隔 + selector）。
- **Label（標籤）**：專案範圍的 key/value，掛在探針與檢查上。
- **Selector（選擇器）**：描述「哪些探針符合」的 JSON 表達式（`all`/`any`/`not`/`label`）。
- **Assignment（指派）**：由 selector 計算出的「某探針要跑某檢查」，帶版本雜湊。
- **Controller（控制器）**：後端 API，協調認證、指派與結果儲存。
- **Hypertable**：TimescaleDB 依時間分區的表，用於高頻時序結果。
- **Continuous Aggregate（連續聚合）**：預先降採樣的物化視圖，供長時間範圍快速查詢。
- **Path hash（路徑雜湊）**：traceroute 路由路徑的指紋，用於偵測路由變化。
- **Heartbeat（心跳）**：探針定期回報「我還在線」的訊號。

---

*本文件依據程式碼庫實況撰寫，涵蓋控制器、探針 agent、前端、資料庫與部署各層。若架構、設定或命令有變更，請同步更新本文件與最接近的 `AGENTS.md`。*
