# Netstamp 推薦功能與 Roadmap

> Subagent 8: Recommended Features and Roadmap  
> 範圍：根據 codebase evidence 與 UX 研究推導，提出目前尚未完整實作但應優先補上的功能。  
> 語言校正：目前沒有 document/stamp/certificate verification 的產品 evidence；以下建議全部改寫為「結果可信度」、「證明頁」、「分享報告」、「incident review」，不把 Netstamp 定位成文件蓋章或憑證驗證工具。

## 1. Codebase Evidence 摘要

### 目前產品邊界

- `README.md` 將 Netstamp 定位為 self-hosted network observability / network monitoring app，核心價值是由使用者控制的 probes 量測 latency、packet loss、routes、TCP reachability 與 probe health。
- `api/main.tsp` 與 `api/services/*.tsp` 顯示公開 API 目前涵蓋 System、Auth、Users、Projects、Labels、Checks、Assignments、Probes、Results、Alerts、Public Status Pages、Probe Runtime、Install。
- `web/src/routes/router.tsx` 顯示 app 主要產品流為 login/register/onboarding、Dashboard、Probes、Checks、Alerts、Status Pages、Labels、Insight、Members、Project Settings、Account Settings，以及公開 `/status/:slug`。
- `server/AGENTS.md` 與 backend packages 顯示 controller 支援 ping、TCP connect、traceroute check definitions、assignment payloads、runtime result ingestion、typed result persistence、result queries、alert incident evaluation、notification outbox worker、Linux probe install assets。
- `server/internal/domain/project/permission.go` 顯示專案角色為 `owner`、`admin`、`editor`、`viewer`，但沒有 global admin、organization hierarchy 或 route-level scope system。
- `api/models/public-status.tsp` 與 `web/src/features/status-pages/components/PublicStatusPage.tsx` 已有公開狀態頁、generated timestamp、active/recent incidents、element metrics、assignment rows 與 charts，是未來「分享結果可信度」的自然延伸點。
- `api/models/result.tsp` 已有 result metadata：raw/aggregate source、resolution、samples、status、error code/message、resolved IP、IP family、traceroute hops。這些欄位足以支撐「proof explanation」與「結果可信度說明」。
- `docs/src/content/docs/reference/configuration.mdx` 與 `web/src/shared/config/features.ts` 顯示 demo/read-only mode 已有設定與 UI flags，但還不是完整的 sample/demo learning flow。
- `docs/src/components/landing/LandingPage.tsx` 與 docs 文案提到 DNS；但 backend/API evidence 目前只有 ping、TCP、traceroute executor。Roadmap 應先修正產品邊界說明，DNS 若保留需標示為未支援或 future check type。

### 沒有 evidence 的方向

- 沒有 evidence 顯示 Netstamp 目前是 document stamping、notary、certificate verification 或檔案簽章產品。
- 因此不建議新增「文件蓋章」、「憑證驗證」、「章戳核驗」等語彙。
- 若要服務「可被他人信任的結果」，應使用 Netstamp 現有脈絡：probe-controlled measurement、result metadata、public status page、generated timestamp、incident history、shareable report、controller/API evidence。

## 2. 功能心智圖

```text
Netstamp Recommended Roadmap
├─ Activation / First Run
│  ├─ Guided first-run checklist
│  ├─ Netstamp does / does not do
│  ├─ Sample project / demo mode
│  └─ Empty states with next action
├─ Measurement Trust
│  ├─ Result trust / proof page
│  ├─ Proof explanation panel
│  ├─ Export / share report
│  └─ Public status confidence context
├─ Operations / Incident Review
│  ├─ Incident review workspace
│  ├─ Notification delivery status center
│  ├─ Error recovery playbooks
│  └─ Search / filter / history
├─ Governance / Security
│  ├─ Audit trail / event history
│  ├─ Role permission visibility
│  ├─ Security and privacy messaging
│  └─ Legal / compliance disclaimer
├─ Admin / Developer Experience
│  ├─ Controller and agent health admin view
│  ├─ In-product API docs shortcuts
│  ├─ API snippets for current project
│  └─ Self-hosting readiness checklist
└─ Quality Layer
   ├─ Accessibility pass
   ├─ Mobile operational views
   └─ Responsive dense data patterns
```

## 3. Roadmap Overview

| Phase           | Timeframe   | Goal                                                       | Recommended features                                                                                   |
| --------------- | ----------- | ---------------------------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| P0 - Foundation | 0-4 weeks   | 讓新使用者理解產品、成功建立第一個可量測結果、避免錯誤期待 | Guided First-Run Checklist, Netstamp Does / Does Not Do, Empty States, Error Recovery, RBAC Visibility |
| P1 - Trust      | 4-8 weeks   | 讓使用者能解釋「這個結果從哪裡來、可信度如何、能不能分享」 | Result Trust / Proof Page, Proof Explanation, Export / Share Report, Legal Disclaimer                  |
| P1 - Operations | 6-10 weeks  | 把 alerts/incidents 從列表提升為 operational review flow   | Incident Review Workspace, Notification Delivery Status Center, Search / Filter / History              |
| P2 - Governance | 8-14 weeks  | 支援團隊操作、審計與安全溝通                               | Audit Trail / Event History, Security/Privacy Messaging, Admin Visibility                              |
| P2 - Scale UX   | 10-16 weeks | 提升大規模資料與跨裝置可用性                               | Sample/Demo Mode, Developer/API Docs Shortcuts, Accessibility/Mobile Pass                              |

## 4. 推薦功能清單

### F01. Guided First-Run Checklist

- **Feature name**: Guided First-Run Checklist / 首次啟用檢查清單
- **User problem**: 新使用者目前 onboarding 主要停在建立 project 與邀請成員，還沒有完整引導到「建立 probe -> 安裝 agent -> 收到 heartbeat -> 建立 check -> 看到第一筆 result -> 設 alert/status page」。
- **Target user**: 第一次 self-host Netstamp 的 operator、SRE、homelab 使用者、社群網路維運者。
- **Evidence/rationale**: `web/src/features/auth/components/OnboardingPage.tsx` 只涵蓋 project name 與 invite；`web/src/features/probes/components/NewProbeDrawer.tsx` 已有 create probe/install/wait heartbeat；`web/src/features/dashboard/components/DashboardPage.tsx` 只顯示 probe/check metrics 與 New Probe action；`README.md` Quick Start 後沒有 app 內完成路徑。
- **Expected UX impact**: 降低從註冊到第一個有效量測結果的時間；讓 Dashboard 空狀態從「空白監控台」變成「下一步作業隊列」。
- **Risk reduced**: 減少 probe 建好了但沒有 check、check 建好了但 selector 沒 match probe、沒有 alert/status page 導致產品價值看不到的風險。
- **Dependencies**: 需要前端聚合 project/probes/checks/assignments/latest results/status pages/alerts 狀態；MVP 可先用既有 API，不一定需要新 backend。
- **MVP scope**: Dashboard 顯示 checklist：Create project、Create probe、Install agent and receive heartbeat、Create first check、View first result in Insight、Create alert rule、Create public status page。每項連到現有 route/drawer。
- **Nice-to-have scope**: 自動偵測卡住的原因；根據缺口顯示短提示；支援 dismiss；支援 workspace-level progress；提供 CLI install command copy telemetry。
- **Priority using RICE**: Reach 5, Impact 5, Confidence 0.90, Effort 2 => Score 11.25，P0。
- **Suggested placement in product flow**: `OnboardingPage` 完成後導到 project dashboard；Dashboard top panel；Probes/Checks/Insight 空狀態也顯示局部 checklist。
- **Suggested FigJam node color**: Blue `#60A5FA`。
- **Confidence**: 高，0.90。

### F02. Netstamp Does / Does Not Do

- **Feature name**: Netstamp Does / Does Not Do / 產品邊界說明
- **User problem**: README/docs/landing 提到 network observability 與 self-hosting，但 app 內沒有清楚說明「Netstamp 做什麼、不做什麼」。若外部研究或 stakeholder 期待 document/stamp/certificate verification，會誤解產品方向。
- **Target user**: 初次評估者、管理者、法務/合規 reviewer、開源部署者、使用 public status page 的外部讀者。
- **Evidence/rationale**: `README.md`、`api/main.tsp`、`server/AGENTS.md` 都指向 network measurement；沒有 document/stamp/certificate verification evidence；`docs/src/components/landing/LandingPage.tsx` 文案提到 DNS，但 backend evidence 目前沒有 DNS executor，需避免過度承諾。
- **Expected UX impact**: 使用者能快速理解 Netstamp 是 probes/checks/results/alerts/status pages，不是文件核驗服務；降低錯誤購買/部署期待。
- **Risk reduced**: 減少產品定位混淆、support 問題、合規誤讀、docs 與 backend 能力不一致的風險。
- **Dependencies**: Docs copy、app onboarding copy、empty state copy；若要保留 DNS 文案，需先新增 DNS check type 或標示「planned」。
- **MVP scope**: 在 onboarding、docs overview、landing feature area、public status footer 加入清楚對照：Does: measure latency/loss/TCP/routes/probe health; Does not: inspect private payloads, notarize documents, certify legal uptime, replace third-party SLA audits。
- **Nice-to-have scope**: 每個 public status/proof report 附「measurement limitations」accordion；admin 可自訂 external-facing disclaimer。
- **Priority using RICE**: Reach 5, Impact 4, Confidence 0.95, Effort 1 => Score 19.0，P0。
- **Suggested placement in product flow**: Onboarding intro、Docs Overview、Landing product surface、Public Status Page footer、Result Proof Page。
- **Suggested FigJam node color**: Slate `#94A3B8`。
- **Confidence**: 高，0.95。

### F03. Result Trust / Proof Page

- **Feature name**: Result Trust / Proof Page / 結果可信度證明頁
- **User problem**: 使用者可以看 Insight charts 與 public status，但缺少一個可分享、可審閱、可解釋的單一結果頁，說明某段時間內某個 probe-check pair 的結果從哪裡來、資料來源、時間窗、樣本數、狀態與限制。
- **Target user**: SRE、網路工程師、support engineer、external stakeholder、incident reviewer。
- **Evidence/rationale**: `api/models/result.tsp` 已有 series metadata、source/resolution、samples、status、resolved IP、IP family、traceroute hops；`api/services/results.tsp` 已有 ping/tcp/traceroute insight/query endpoints；`PublicStatusPage.tsx` 已有 generated timestamp 與 public rendering，但不是 single-result proof。
- **Expected UX impact**: 讓使用者能把「這裡真的從 Tokyo probe 看到 packet loss」轉成一個可被他人閱讀的 artifact。
- **Risk reduced**: 減少截圖失真、chart 無上下文、外部讀者不知道資料可信度與限制的風險。
- **Dependencies**: 新 route 與可能的新 share token/snapshot API；MVP 可先做 authenticated result proof route，後續再公開分享。
- **MVP scope**: `/projects/:projectRef/insight/proof?...` 顯示 probe、check、target、time range、generatedAt、series source/resolution、samples、latest status、error codes、chart、raw API request link、limitations。
- **Nice-to-have scope**: Public signed/share token、snapshot immutability、JSON export、comparison against previous window、route hash diff、public proof landing with status page branding。
- **Priority using RICE**: Reach 4, Impact 5, Confidence 0.85, Effort 3 => Score 5.67，P1。
- **Suggested placement in product flow**: Insight selected pair action「Create proof page」；Alert incident detail「Open proof window」；Public status element「View measurement evidence」。
- **Suggested FigJam node color**: Green `#34D399`。
- **Confidence**: 高，0.85。

### F04. Proof Explanation Panel

- **Feature name**: Proof Explanation Panel / 結果可信度解釋面板
- **User problem**: Metrics 目前對 experienced operators 可讀，但外部讀者或新使用者不一定知道 raw vs aggregate、resolution、samples、stale assignment、partial traceroute、timeout/error 的含義。
- **Target user**: 新 operator、support engineer、public status viewer、incident reviewer。
- **Evidence/rationale**: `api/models/result.tsp` 有 `SeriesSource`、`SeriesResolution`、`LatestResultStatus`、`errorCode`、`errorMessage`；`web/src/features/status-pages/components/PublicStatusPage.tsx` 有 public states 但沒有 measurement explanation；`web/src/features/alerts/components/AlertsPage.tsx` 的 incident detail 有 What happened / Timeline / Notifications，但沒有 proof semantics。
- **Expected UX impact**: 使用者不只看到「紅色/綠色」，也知道結果為何可信、哪裡有限制、是否樣本不足。
- **Risk reduced**: 減少把 no data 誤判成 healthy、把 aggregate 當 raw、把 partial traceroute 當完整路徑的風險。
- **Dependencies**: 前端 copy/design；可使用現有 result metadata；若要顯示更精確 retention/rollup policy，需 API 暴露 policy。
- **MVP scope**: 在 Insight/Proof/Public Status/Incident detail 加入 explanation drawer：data source、sample count、window、probe identity、check config、known limitations。
- **Nice-to-have scope**: Contextual explanations per check type；「why this status」decision trace；link 到 docs/API operation；localized copy。
- **Priority using RICE**: Reach 4, Impact 4, Confidence 0.80, Effort 2 => Score 6.4，P1。
- **Suggested placement in product flow**: Insight panel header、Public status check row、Incident detail drawer、Proof page side panel。
- **Suggested FigJam node color**: Green `#86EFAC`。
- **Confidence**: 中高，0.80。

### F05. Export / Share Report

- **Feature name**: Export / Share Report / 分享報告與匯出
- **User problem**: Public status page 可以公開狀態，但使用者還不能將一次 incident 或一段 measurement window 匯出為可傳給客戶/社群/管理者的報告。
- **Target user**: Support engineer、SRE lead、社群網路維運者、開源專案 maintainer、NOC operator。
- **Evidence/rationale**: `api/services/public-status.tsp` 已有 public status endpoint；`web/src/features/status-pages/components/StatusPagesPage.tsx` 有 Open public page；`api/services/results.tsp` 有 result query endpoints；目前沒有 export/share proof/report endpoints 或 UI。
- **Expected UX impact**: 從「開一個頁面給人看」提升到「給出可保存、可稽核、可比較的 incident/result report」。
- **Risk reduced**: 減少用截圖、手動複製數字、缺少時間窗/樣本資訊造成的溝通錯誤。
- **Dependencies**: Result Proof Page；可先 frontend 產生 print/PDF-friendly HTML；CSV/JSON 需要 backend 或 client export。
- **MVP scope**: Proof page 提供 copy link、print view、download JSON metadata、download CSV series。
- **Nice-to-have scope**: Branded PDF、expiring public share link、redaction controls、compare windows、include incident timeline and notification delivery status。
- **Priority using RICE**: Reach 3, Impact 5, Confidence 0.78, Effort 3 => Score 3.9，P1。
- **Suggested placement in product flow**: Insight proof page action bar；Alert incident detail action bar；Public status management table row action。
- **Suggested FigJam node color**: Green `#22C55E`。
- **Confidence**: 中高，0.78。

### F06. Incident Review Workspace

- **Feature name**: Incident Review Workspace / Incident review 工作區
- **User problem**: Alert incidents 目前可列表、篩選與查看 detail，但沒有明確的 acknowledge/resolve/comment/review flow，也缺少把 incident 連到 result proof、affected assignments、notification attempts 的完整調查視圖。
- **Target user**: On-call operator、SRE、NOC、support engineer、team lead。
- **Evidence/rationale**: `api/models/alert.tsp` 有 `open/acknowledged/resolved` status 與 incident summary；`server/internal/controller/application/alert/service.go` 只有 list/get incidents，沒有 acknowledge/resolve mutation；`AlertsPage.tsx` detail 有 What happened / Timeline / Notifications，但目前偏 read-only。
- **Expected UX impact**: 把 Alerts 從「規則與列表」提升為 incident response console。
- **Risk reduced**: 減少 incident 無 owner、無處理紀錄、無法回顧 root cause、外部報告缺上下文的風險。
- **Dependencies**: 新 backend incident mutation endpoints；DB 可能已有 acknowledged/resolved 欄位但需確認 repository write support；前端 drawer/workspace。
- **MVP scope**: Add acknowledge、resolve、reopen actions；incident detail 顯示 status transition timeline、linked proof window、affected probe/check、last evaluation summary。
- **Nice-to-have scope**: Comments、assignee、severity change history、post-incident review template、export review report、link to status page incident summary。
- **Priority using RICE**: Reach 4, Impact 5, Confidence 0.82, Effort 4 => Score 4.1，P1。
- **Suggested placement in product flow**: Alerts -> Incidents -> Incident detail drawer/full page；Dashboard open incidents card。
- **Suggested FigJam node color**: Red `#F87171`。
- **Confidence**: 中高，0.82。

### F07. Notification Delivery Status Center

- **Feature name**: Notification Delivery Status Center / 通知投遞狀態中心
- **User problem**: Alert notifications 有 test action 與 outbox worker，但使用者看不到正式 incident notification 的 pending/sending/delivered/failed/discarded 歷史與 retry reason。
- **Target user**: SRE、NOC、project admin、self-host operator。
- **Evidence/rationale**: `server/db/migrations/202606140001_add_alerting_beta.sql` 建立 `notification_outbox`，包含 status、attempt_count、next_attempt_at、last_error、dedupe_key；`server/internal/controller/application/notification/worker.go` 實作 retry/discard/failed/delivered；`api/models/alert.tsp` 只暴露 notification test result，未暴露 outbox list。
- **Expected UX impact**: 當 Slack/webhook/email 沒收到時，使用者能在 app 內知道是 destination disabled、sender unavailable、retrying 或 permanent failure。
- **Risk reduced**: 減少 alert 已 firing 但通知未送達而沒被發現的風險。
- **Dependencies**: 新 API: list notification deliveries by project/incident/notification；RBAC 需 owner/admin 或 editor? 建議至少 read for project members、details for admin/owner。
- **MVP scope**: Incident detail Notifications panel 顯示 last delivery events；Notifications tab 顯示最近 50 筆 delivery status；failed row 顯示 sanitized error kind/code/message。
- **Nice-to-have scope**: Manual retry、mute/silence windows、destination health score、per-notification delivery history chart。
- **Priority using RICE**: Reach 3, Impact 5, Confidence 0.88, Effort 3 => Score 4.4，P1。
- **Suggested placement in product flow**: Alerts -> Notifications tab；Incident detail -> Notifications；Project Settings -> Alert delivery。
- **Suggested FigJam node color**: Amber `#FBBF24`。
- **Confidence**: 高，0.88。

### F08. Error Recovery Playbooks

- **Feature name**: Error Recovery Playbooks / 錯誤復原與診斷指引
- **User problem**: Probe install、heartbeat、selector matching、check validation、notification delivery、public status unavailable 都可能失敗；目前多數 UI 顯示 loading/error/toast，但缺少 next diagnostic action。
- **Target user**: 第一次部署者、自架 operator、on-call engineer。
- **Evidence/rationale**: `NewProbeDrawer.tsx` 有 install command 與 heartbeat polling；`ChecksPage.tsx` 有 selector preview 與 validation；`PublicStatusPage.tsx` 有 unavailable/not found fallback；`AlertsPage.tsx` 有 test notification toast；backend runtime 有 probe hello/heartbeat/assignments/results endpoints。
- **Expected UX impact**: 使用者看到失敗時不需要離開 app 猜原因，能直接執行下一個修復步驟。
- **Risk reduced**: 減少 setup abandonment、錯誤設定 selector、agent secret/permission 問題、通知漏送。
- **Dependencies**: 前端 state mapping；若要深度診斷 probe agent，需要 backend 暴露 runtime status detail / last runtime error。
- **MVP scope**: 對常見狀態提供 recovery card：no heartbeat、no assignment matched、no latest result、notification failed、public status empty、query no data。每張卡給 1-3 個具體 action。
- **Nice-to-have scope**: Agent-side diagnostic bundle、copyable `netstamp-agent doctor` command、controller readiness checks、guided remediation wizard。
- **Priority using RICE**: Reach 5, Impact 4, Confidence 0.86, Effort 2 => Score 8.6，P0。
- **Suggested placement in product flow**: New Probe drawer、Probe Detail、Checks selector preview、Insight empty state、Alerts Notifications tab、Public Status management。
- **Suggested FigJam node color**: Amber `#F59E0B`。
- **Confidence**: 高，0.86。

### F09. Empty States With Operational Next Actions

- **Feature name**: Empty States With Operational Next Actions / 有操作性的空狀態
- **User problem**: 多個表格已有 emptyLabel，但空狀態多半只說沒有資料，沒有依據產品階段給具體下一步。
- **Target user**: 新使用者、demo reviewer、低頻 operator。
- **Evidence/rationale**: `ProbeList.tsx` emptyLabel 是 `No probes found`；`StatusPagesPage.tsx` empty detail 是 create status page；`AlertsPage.tsx` 有部分 EmptyAction；`DashboardPage.tsx` 只有 metrics/map，沒有空專案引導。
- **Expected UX impact**: 把空資料變成可行動路徑，縮短學習與設定時間。
- **Risk reduced**: 減少使用者以為系統壞掉、資料遺失或不知道下一步。
- **Dependencies**: 前端 component patterns；最好抽出 shared EmptyState primitive 到 `@netstamp/ui`。
- **MVP scope**: Dashboard/Probes/Checks/Insight/Alerts/Status Pages/Members 各自提供 empty state：原因、下一步 button、docs link、是否受 RBAC/demo mode 限制。
- **Nice-to-have scope**: Empty state 根據 role 調整；可顯示 sample data CTA；用 checklist progress 串接。
- **Priority using RICE**: Reach 5, Impact 3, Confidence 0.90, Effort 2 => Score 6.75，P0。
- **Suggested placement in product flow**: 全部 route-level panels 與 DataTable emptyLabel。
- **Suggested FigJam node color**: Blue `#93C5FD`。
- **Confidence**: 高，0.90。

### F10. Search / Filter / History Hub

- **Feature name**: Search / Filter / History Hub / 搜尋、篩選與歷史查找中心
- **User problem**: Probes、Checks、Alerts、Insight 各自有局部搜尋/篩選，但沒有跨資源查找，也沒有讓使用者從 target/probe/incident 快速追到歷史結果的 unified flow。
- **Target user**: SRE、support、NOC、管理多 probe/check 的 operator。
- **Evidence/rationale**: `ProbesPage.tsx` 有 search/sort；`ChecksPage.tsx` 有 query param search/type；`AlertsPage.tsx` 有 rule search/status/type；`InsightPage.tsx` 有 time range、groupBy、assignment filters；但 route 間缺乏 shared search/result history hub。
- **Expected UX impact**: 使用者能以 target、probe、label、incident ID、public status page slug 找到相關資源與歷史。
- **Risk reduced**: 減少大型 project 中找不到相關 check/result/incident 的風險。
- **Dependencies**: MVP 可在前端聚合現有 list APIs；長期需 backend search endpoint 與 pagination。
- **MVP scope**: App shell command/search palette：搜尋 probes/checks/labels/status pages/alert rules/incidents；結果可跳轉到對應 route。
- **Nice-to-have scope**: Unified history timeline；saved views；advanced filters；global query endpoint；audit/result events in same search。
- **Priority using RICE**: Reach 4, Impact 4, Confidence 0.75, Effort 3 => Score 4.0，P1。
- **Suggested placement in product flow**: AppShell top command/search；route headers；Incident and Proof pages link back to related resources。
- **Suggested FigJam node color**: Blue `#3B82F6`。
- **Confidence**: 中高，0.75。

### F11. Audit Trail / Event History

- **Feature name**: Audit Trail / Event History / 專案事件歷史
- **User problem**: backend 有 application event logging，但使用者無法在產品內查看誰建立/更新/刪除 probe/check/label/alert/status page/member，也無法把這些變更放入 incident review。
- **Target user**: Project owner/admin、security reviewer、incident reviewer、self-host operator。
- **Evidence/rationale**: `server/AGENTS.md` 詳述 auth/project/label/check/probe/proberuntime event recorder；`server/internal/controller/logger` 產生 audit/security-oriented logs；但沒有 `audit_events` table、API 或 web route evidence。
- **Expected UX impact**: 操作可追溯，團隊協作更可信；incident review 可以知道是否剛改了 selector/check/notification。
- **Risk reduced**: 減少配置變更造成 outage 但無法追查、多人操作互相覆蓋、合規審查缺 evidence 的風險。
- **Dependencies**: 新 DB table 或 structured log ingestion；API list events；RBAC policy；PII redaction。
- **MVP scope**: Project Settings -> Activity tab，記錄 project/member/invite/check/label/probe/alert/status page 主要 mutation：actor、action、resource、timestamp、outcome、reason。
- **Nice-to-have scope**: Diff view、export audit log、filter by actor/resource/action、incident auto-correlates recent changes、webhook for audit events。
- **Priority using RICE**: Reach 3, Impact 5, Confidence 0.82, Effort 5 => Score 2.46，P2。
- **Suggested placement in product flow**: Project Settings -> Activity；Incident detail -> Recent changes；Admin visibility -> Security events。
- **Suggested FigJam node color**: Violet `#A78BFA`。
- **Confidence**: 中高，0.82。

### F12. Role Permission Visibility

- **Feature name**: Role Permission Visibility / 角色權限可視化
- **User problem**: Members page 能修改 roles，但使用者不一定知道 owner/admin/editor/viewer 各自能做什麼，也不清楚為何某些 button disabled。
- **Target user**: Project owner/admin、被邀請成員、viewer/editor。
- **Evidence/rationale**: `server/internal/domain/project/permission.go` 定義 canonical action policy；`MembersPage.tsx` 有 role select 與 disabled/protected labels；`server/AGENTS.md` 說 HTTP middleware 只證明 identity，application services 負責 authorization。
- **Expected UX impact**: 減少權限困惑，提高團隊管理信心。
- **Risk reduced**: 降低誤給高權限、低權限成員無法完成工作但不知道原因、support 問題。
- **Dependencies**: 前端 role matrix；可直接從 code policy 手動鏡像，長期可 API 暴露 permission matrix。
- **MVP scope**: Members page 加入 permission matrix：read project、write project、manage members、labels/checks/probes/alerts、delete project；disabled action tooltip 顯示需要的 role。
- **Nice-to-have scope**: Per-resource capability badges；invite flow 預覽 role capability；API returns current user capabilities for project。
- **Priority using RICE**: Reach 4, Impact 3, Confidence 0.90, Effort 2 => Score 5.4，P0。
- **Suggested placement in product flow**: Members page、Project Settings、disabled action tooltips、onboarding invite step。
- **Suggested FigJam node color**: Violet `#C4B5FD`。
- **Confidence**: 高，0.90。

### F13. Security / Privacy Messaging

- **Feature name**: Security / Privacy Messaging / 安全與隱私說明
- **User problem**: Netstamp 自架且處理 probe IP、public IP、AS、targets、notification URLs、email/SMTP 等敏感資訊；產品內沒有集中說明資料收集、保存、分享與不收集內容。
- **Target user**: Self-host admin、security reviewer、project owner、public status viewer。
- **Evidence/rationale**: `api/models/probe.tsp` 暴露 publicV4/publicV6/as/addrs；`api/models/result.tsp` 暴露 resolvedIp/errorMessage；`server/AGENTS.md` 明確禁止 logging secrets/raw personal data；`docs/src/content/docs/reference/configuration.mdx` 有 tracking consent 與 demo controls，但 app 內缺少 security/privacy messaging。
- **Expected UX impact**: 讓部署者更清楚如何安全操作，也讓外部讀者理解 public status/proof report 的資料界線。
- **Risk reduced**: 降低公開敏感 target/IP、誤以為 Netstamp 監看 payload、錯誤暴露 admin API 的風險。
- **Dependencies**: Docs copy、app settings copy、public status/proof disclaimer；可能需要 redaction controls。
- **MVP scope**: Project Settings -> Security & privacy panel；Public Status/Proof footer 顯示「measurements only, no payload inspection」；notification config 顯示 secret handling note。
- **Nice-to-have scope**: Per-status-page redaction settings、hide target/IP options、data retention visualization、privacy export/delete controls。
- **Priority using RICE**: Reach 4, Impact 4, Confidence 0.84, Effort 2 => Score 6.72，P1。
- **Suggested placement in product flow**: Project Settings、Status Page editor、Proof Page、Docs Configuration/Self-hosting。
- **Suggested FigJam node color**: Red `#FB7185`。
- **Confidence**: 高，0.84。

### F14. Legal / Compliance Disclaimer For Shared Evidence

- **Feature name**: Legal / Compliance Disclaimer / 分享證據免責說明
- **User problem**: Public status/proof/report 可能被外部讀者當成 SLA、法律證明或第三方稽核，但 Netstamp 是 self-hosted measurement from controlled probes，需要明確限制說明。
- **Target user**: External stakeholder、customer support recipient、project admin、legal/compliance reviewer。
- **Evidence/rationale**: `README.md` 與 docs 強調 self-hosted probes you control；沒有 certificate/legal verification evidence；Public Status Page 已公開分享 measurement state，但沒有 legal/disclaimer copy。
- **Expected UX impact**: 使用者能安全分享結果，同時避免被解讀為不可挑戰的法律證明。
- **Risk reduced**: 降低 legal misrepresentation、SLA 誤讀、外部報告過度承諾。
- **Dependencies**: Copy/legal review；admin 可自訂 disclaimer 的話需 DB/API field。
- **MVP scope**: Public status/proof/report footer 固定文字：Netstamp reports measurements collected by configured probes; it does not certify documents, inspect payloads, or act as independent legal attestation。
- **Nice-to-have scope**: Project-level custom disclaimer、per-page disclaimer、export report includes disclaimer version and generated timestamp。
- **Priority using RICE**: Reach 3, Impact 4, Confidence 0.88, Effort 1 => Score 10.56，P0/P1。
- **Suggested placement in product flow**: Public Status Page footer、Proof Page footer、Export report cover/metadata、Docs Does/Does Not Do。
- **Suggested FigJam node color**: Slate `#64748B`。
- **Confidence**: 高，0.88。

### F15. Sample Project / Demo Learning Mode

- **Feature name**: Sample Project / Demo Learning Mode / 範例專案與 demo 學習模式
- **User problem**: Demo/read-only mode 已存在，但新使用者若沒有立即可用的 probes/checks/results，很難理解 Insight、Alerts、Status Page 的完整價值。
- **Target user**: Evaluators、contributors、docs readers、sales/demo hosts、first-time self-host users。
- **Evidence/rationale**: `docs/src/content/docs/reference/configuration.mdx` 有 `DEMO_MODE` 與 `VITE_NETSTAMP_DEMO_MODE`；`web/src/shared/config/features.ts` 會停用 writes；`docs/src/components/docs/TopNav.astro` 有 demo link；但沒有 evidence 顯示 app 可從空專案載入 sample dataset。
- **Expected UX impact**: 使用者不必先部署真 probe 就能理解產品流；demo instance 更能展示 end-to-end。
- **Risk reduced**: 降低 trial drop-off、空狀態誤判、開源 contributor 不知道如何驗證 UI 的風險。
- **Dependencies**: Seed data、demo deployment、read-only UX；可先用 static mock/sample mode，不必混入 production DB。
- **MVP scope**: Add 「Load sample project」或 docs demo seed：包含 probes、checks、assignments、results、alert incidents、status page；demo banner 清楚標示 view-only。
- **Nice-to-have scope**: Resettable demo database、scenario switcher（packet loss、route change、TCP failure）、guided demo tour。
- **Priority using RICE**: Reach 4, Impact 4, Confidence 0.78, Effort 4 => Score 3.12，P2。
- **Suggested placement in product flow**: Onboarding empty state、Docs demo link、Dashboard demo banner、Account settings pending invites/sample access。
- **Suggested FigJam node color**: Blue `#38BDF8`。
- **Confidence**: 中高，0.78。

### F16. Admin Visibility / Controller Health

- **Feature name**: Admin Visibility / Controller Health / 管理者可見性與控制器健康
- **User problem**: Self-host operators 需要知道 controller、DB、Timescale rollups、notification worker、alert evaluator、agent install assets、OpenAPI docs、metrics/tracing 是否正常；目前這些散落在 config、healthz、metrics、logs。
- **Target user**: Self-host admin、project owner、operator running Docker Compose。
- **Evidence/rationale**: `server/internal/controller/transport/http/router.go` 有 `/healthz`、`/metrics`、`/api/v1/openapi.json`、`/api/v1/docs`、install assets；`server/internal/controller/app/bootstrap.go` wires tracing/metrics/notification worker/alert evaluator；docs configuration 有 observability stack。
- **Expected UX impact**: 自架者能在 app 內判斷系統狀態，而不是只靠 logs/Grafana。
- **Risk reduced**: 減少部署錯誤、worker 停掉、SMTP 未配置、agent binary missing、OpenAPI mismatch 沒被發現的風險。
- **Dependencies**: 新 admin/system endpoint；RBAC/global admin 目前不存在，需先決定可見性。MVP 可只顯示 read-only controller health to authenticated owners/admins。
- **MVP scope**: Project/Account Settings 加「Self-host status」：controller version、API version、DB readiness、notification worker enabled、alert evaluation enabled、SMTP configured、demo mode、install asset availability。
- **Nice-to-have scope**: System-wide admin role、background job status、rollup freshness、Grafana links、probe-agent version compliance。
- **Priority using RICE**: Reach 3, Impact 4, Confidence 0.76, Effort 4 => Score 2.28，P2。
- **Suggested placement in product flow**: Project Settings or Account Settings -> Admin visibility；Docs deployment checklist。
- **Suggested FigJam node color**: Slate `#475569`。
- **Confidence**: 中高，0.76。

### F17. In-Product Developer / API Docs Shortcuts

- **Feature name**: In-Product Developer / API Docs Shortcuts / 產品內 API 文件入口
- **User problem**: API reference 已存在，但產品內使用者在看 project/probe/check/result 時無法快速跳到對應 OpenAPI operation 或取得帶 project/check/probe context 的 request example。
- **Target user**: Developer、automation engineer、SRE writing scripts、integrator。
- **Evidence/rationale**: `api/main.tsp` 產生 OpenAPI；`docs/src/components/openapi/OpenAPIExplorer.tsx` 是互動 API explorer；`server/internal/controller/transport/http/router.go` exposes `/api/v1/docs` and `/api/v1/openapi.json`；web client 有 generated `openapi.d.ts`。
- **Expected UX impact**: 讓 API 從 docs-only 變成 app workflow 的自然延伸。
- **Risk reduced**: 減少 API 使用錯路徑、忘記 project ref、手動拼 query params 出錯。
- **Dependencies**: Frontend links/snippet generator；若要 API token 支援需新 auth model，目前只有 session cookie/probe auth evidence。
- **MVP scope**: 在 Probes/Checks/Insight/Alerts/Status Pages action menu 加「Open API docs」與「Copy cURL」；cURL 使用 current projectRef/resource IDs。
- **Nice-to-have scope**: Personal API tokens、scoped service accounts、SDK snippets、generated Terraform/import examples。
- **Priority using RICE**: Reach 3, Impact 3, Confidence 0.82, Effort 2 => Score 3.69，P2。
- **Suggested placement in product flow**: Resource action menus、Docs top nav、OpenAPI explorer prefilled route。
- **Suggested FigJam node color**: Slate `#94A3B8`。
- **Confidence**: 中高，0.82。

### F18. Accessibility And Mobile Operations Pass

- **Feature name**: Accessibility And Mobile Operations Pass / 無障礙與行動操作強化
- **User problem**: Netstamp 是密集操作型 dashboard；表格、抽屜、圖表、地圖、public status 在手機與輔助工具上容易出現資訊壓縮、focus trap、圖表不可讀等問題。
- **Target user**: On-call mobile user、keyboard/screen-reader user、public status viewer、operator on small laptop。
- **Evidence/rationale**: `design.md` 明確要求 visible focus、semantic landmarks、not color alone、touch targets、mobile collapse、wide table horizontal scroll；web routes 使用 dense DataTable/EditorDrawer/NetworkMap/ChartPanel；Public status 是外部可讀 surface。
- **Expected UX impact**: 讓 critical status 與 incident review 能在手機上完成最低限度操作，也讓公開頁符合基本可及性期待。
- **Risk reduced**: 降低 on-call 無法在手機確認 incident、鍵盤使用者無法完成設定、public status 無法被輔助工具理解的風險。
- **Dependencies**: UI audit、Playwright/axe 或 manual QA、CSS module updates；可能需 shared component fixes in `@netstamp/ui`。
- **MVP scope**: Audit Dashboard/Probes/Checks/Insight/Alerts/Status/PublicStatus；修正 focus、aria labels、empty/loading state semantics、table overflow、drawer mobile height、chart textual fallback。
- **Nice-to-have scope**: Mobile incident quick view、reduced motion map/chart behavior、screen-reader summaries for charts/topology、accessibility CI check。
- **Priority using RICE**: Reach 5, Impact 4, Confidence 0.80, Effort 4 => Score 4.0，P1/P2。
- **Suggested placement in product flow**: Cross-cutting quality layer；prioritize PublicStatusPage、AlertsPage、NewProbeDrawer、InsightPage。
- **Suggested FigJam node color**: Blue `#0EA5E9`。
- **Confidence**: 中高，0.80。

## 5. Priority Stack

### P0: 先做，立即降低啟用與誤解風險

1. F02 Netstamp Does / Does Not Do
2. F01 Guided First-Run Checklist
3. F08 Error Recovery Playbooks
4. F09 Empty States With Operational Next Actions
5. F12 Role Permission Visibility
6. F14 Legal / Compliance Disclaimer For Shared Evidence

### P1: 建立可信結果與 incident review

1. F03 Result Trust / Proof Page
2. F04 Proof Explanation Panel
3. F05 Export / Share Report
4. F06 Incident Review Workspace
5. F07 Notification Delivery Status Center
6. F13 Security / Privacy Messaging
7. F18 Accessibility And Mobile Operations Pass

### P2: 擴大團隊治理、demo、developer/admin 體驗

1. F10 Search / Filter / History Hub
2. F11 Audit Trail / Event History
3. F15 Sample Project / Demo Learning Mode
4. F16 Admin Visibility / Controller Health
5. F17 In-Product Developer / API Docs Shortcuts

## 6. 建議 Product Flow Placement

```text
Register/Login
└─ Onboarding
   ├─ Does / does not do
   ├─ Create project
   ├─ Invite members with role preview
   └─ Continue to first-run checklist

Dashboard
├─ First-run checklist
├─ Empty operational next action
├─ Controller/admin status summary
└─ Open incidents / recent proof reports

Probes
├─ New probe wizard
├─ Error recovery for no heartbeat
├─ Probe detail -> assignments/results link
└─ API docs/cURL shortcut

Checks
├─ Selector preview
├─ Empty state for no probes/no labels/no assignments
├─ Error recovery for zero matched probes
└─ API docs/cURL shortcut

Insight
├─ Search/filter/history
├─ Result detail
├─ Create proof page
├─ Proof explanation panel
└─ Export/share report

Alerts
├─ Rules
├─ Incidents
│  ├─ Incident review workspace
│  ├─ Acknowledge/resolve
│  ├─ Linked proof window
│  └─ Notification delivery history
└─ Notifications
   ├─ Test destination
   └─ Delivery status center

Status Pages
├─ Public status editor
├─ Public status page
│  ├─ Measurement disclaimer
│  ├─ Proof explanation
│  └─ Evidence link per element
└─ Export/share current status report

Project Settings
├─ Members and role matrix
├─ Activity / audit trail
├─ Security and privacy messaging
├─ Self-host status
└─ Legal disclaimer settings

Docs / OpenAPI
├─ API explorer
├─ Does / does not do
├─ Self-hosting readiness
└─ Demo/sample mode entry
```

## 7. FigJam Node Color Legend

| Category                              | Color            | 用途                                     |
| ------------------------------------- | ---------------- | ---------------------------------------- |
| Activation / onboarding               | Blue `#60A5FA`   | 首次使用、空狀態、學習模式               |
| Measurement trust                     | Green `#34D399`  | 結果可信度、proof、分享報告              |
| Incident / operations risk            | Red `#F87171`    | incident review、critical error、復原    |
| Alert delivery / operational workflow | Amber `#FBBF24`  | notification、worker、retry、狀態更新    |
| Governance / permissions              | Violet `#A78BFA` | RBAC、audit trail、activity              |
| Docs / developer / admin              | Slate `#94A3B8`  | API docs、自架狀態、產品邊界、disclaimer |

## 8. 重要語彙建議

### 建議使用

- 結果可信度
- Measurement evidence
- Proof page / 證明頁
- Shareable report / 分享報告
- Incident review
- Generated at
- Data source: raw / aggregate
- Sample count
- Probe-controlled measurement
- Measurement limitations

### 避免使用

- 文件蓋章
- 文件驗證
- 憑證驗證
- 法律認證
- 第三方 SLA 稽核
- 不可否認證明

## 9. Open Questions

- DNS 是否是近期 roadmap 的 check type？Landing/docs 有 DNS 文案，但 backend/API executor evidence 目前只有 ping、TCP、traceroute。建議短期先修正文案或標示 planned，避免能力不一致。
- Public proof link 是否需要 token、expiration、snapshot immutability？若只是 status page live link，可信度與可追溯性較低。
- Audit trail 要落 DB 還是讀 structured logs？若要產品內可查詢與匯出，建議 DB-backed event history。
- Incident acknowledge/resolve 權限應是 editor 以上還是所有 project members？目前 alert write 是 owner/admin/editor，但 notification write 是 owner/admin；incident response 可以獨立定義。
- 是否需要 global/system admin？目前 codebase 明確沒有 global admin。Admin visibility MVP 應避免假設 system-wide role。

## 10. Recommended Next Step

建議第一批設計/實作切成三個小 vertical slices：

1. **Activation slice**: F02 + F01 + F09 + F12  
   只改前端與 docs copy，多數可用現有 API 完成，立即改善初次使用與產品邊界。

2. **Trust slice**: F03 + F04 + F14  
   先做 authenticated proof route 與 explanation panel，再決定 public share token/export。

3. **Incident operations slice**: F06 + F07 + F08  
   先補 incident review actions 與 notification delivery read model，再把 error recovery 接進 Probes/Checks/Alerts。
