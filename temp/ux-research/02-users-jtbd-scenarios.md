# Netstamp 使用者類型、JTBD 與情境

## 研究範圍

本文件根據目前 codebase、README、docs、React routing、TypeSpec API contract、server/agent 架構與部署檔案推導 Netstamp 的使用者類型。以下內容區分：

- **Evidence-based**：repo 內有明確文字、路由、API、domain、UI 或部署設定支持。
- **Assumption**：由產品相鄰脈絡合理推導，但尚未看到 repo 內明確使用者研究、文案或功能證據。
- **Gap / research needed**：目前產品或證據不足，需要後續訪談、使用紀錄或需求驗證。

## Evidence Inventory

### 明確 evidence

- `README.md` 將 Netstamp 定位為「Self-hosted network observability from probes you control」，用於從自己的 machines、regions、labs、edge nodes、private infrastructure 觀察 reachability、latency、packet loss、routes、probe health。
- `README.md` 列出 self-hosted controller、React web app、Go API、lightweight probe agents、Ping/TCP/traceroute checks、project workspaces、roles、labels、dashboards、alerts、notifications、Docker Compose、OpenAPI contract。
- `server/AGENTS.md` 說明 controller、probe agent、runtime routes、install assets、PostgreSQL/TimescaleDB、result ingestion、alert evaluation、notification outbox，以及目前支援 Ping、TCP connect、traceroute；也明確說目前沒有 global admin role、organization/project hierarchy、GraphQL、message queue、payment、object storage、functional DNS/HTTP probe executors。
- `api/main.tsp` 的 tags 包含 Auth、Users、Projects、Project Members、Project Invites、Labels、Checks、Assignments、Probes、Results、Alerts、Public Status Pages、Probe Runtime、Install。
- `api/models/check.tsp` 的 `CheckType` 只有 `ping | tcp | traceroute`。DNS 在部分 landing/docs copy 出現，但目前 API contract 與 server guidance 沒有 DNS check/executor evidence。
- `api/models/project.tsp` 與 `server/internal/domain/project/permission.go` 顯示 project role 為 owner/admin/editor/viewer，權限是 project-scoped，不是全域管理者模型。
- `web/src/routes/router.tsx` 與 `web/src/routes/sidebarItems.ts` 顯示主要產品路由：Dashboard、Probes、Checks、Alerts、Status Pages、Labels、Insight、Members、Project Settings、Account Settings，以及公開 `/status/:slug`。
- `web/src/features/probes/components/NewProbeDrawer.tsx` 顯示建立 probe 後產生 registration secret/install command，使用者需在 host 上安裝 agent，UI 會等待 controller 收到 heartbeat。
- `web/src/features/checks/components/ChecksPage.tsx` 顯示 check creation/editing、Ping/TCP/Traceroute config、selector preview、assignment matching、duplicate/delete/batch delete。
- `web/src/features/insight/components/InsightPage.tsx` 顯示使用者可依 probe/check/assignment、時間範圍與 refresh interval 查詢 Ping/TCP series 與 insight，以及 Traceroute runs/topology。
- `web/src/features/alerts/components/AlertsPage.tsx` 與 `api/models/alert.tsp` 顯示 alert rules、incidents、notifications；通知類型包含 webhook、Slack、Discord、Telegram、email。前端明確指出 traceroute alert metrics 尚未可用。
- `api/models/public-status.tsp` 與 `web/src/features/status-pages/components/PublicStatusPage.tsx` 顯示公開 status page 可呈現 status、open incidents、assignment metrics、latest status、charts 與 generated timestamp。
- `deployments/docker/compose.yaml` 與 docs deployment/configuration 顯示 self-host runtime 包含 Netstamp app image、PostgreSQL/TimescaleDB、migrations、controller/web/API；observability stack 另含 Grafana/VictoriaMetrics/VictoriaTraces/VictoriaLogs/Vector。

### 明確排除或限制

- **沒有 evidence 顯示 Netstamp 是文件 stamp、certificate、notary、timestamping 或簽章產品**。目前 evidence 全部指向 network observability/probe/check/result/alert/agent/API/docs。
- **DNS check 是不確定項**：landing/docs 文案提到 DNS，但 TypeSpec `CheckType` 與 server guidance 目前只確認 Ping、TCP connect、Traceroute。本文不把 DNS 當成已支援事實。
- **沒有 global admin evidence**：目前管理邊界是 project owner/admin/editor/viewer，加上 self-host infra operator；沒有全域 admin、organization hierarchy、billing/admin console evidence。
- **外部 stakeholder 的 evidence 主要來自 public status pages 與 notification recipient**；沒有 evidence 顯示外部者可登入、留言、訂閱、ack incident 或驗證法律/合規文件。

## 使用者類型總覽

| 類型 | Evidence-based 定義 | Assumption / 待驗證 |
| --- | --- | --- |
| **Primary user** | 需要從受控 probe viewpoint 監看 latency、packet loss、TCP reachability、routes、probe health，並處理 incidents 的網路/平台/SRE 使用者。 | 可能是 SRE、network engineer、platform engineer、DevOps 或 homelab/community network operator；repo 沒有直接 persona 名稱。 |
| **Secondary user** | 透過 project access 查看 dashboard/insight/status、建立或維護 checks/labels、使用 OpenAPI/docs 的技術團隊成員。 | 可能是 service owner、backend engineer、support engineer、technical analyst 或 automation integrator；需驗證他們是否每天使用或只在 incident 時使用。 |
| **Admin/operator** | project owner/admin 與 self-host runtime operator：管理 projects/members/invites、部署 controller、DB/migrations、agent install/update/secret rotation、notifications、SMTP/observability。 | 同一人可能同時是 Primary user；較大團隊可能分成 infra operator 與 project owner。 |
| **Verifier / recipient / external stakeholder** | 可透過 public status page 查看公開狀態、incidents、metrics；或透過 Slack/Discord/Telegram/email/webhook 收到 alert。 | 「verifier」若指文件或憑證驗證者則無 evidence；本文僅作 network status verifier/incident recipient 的相鄰概念假設。 |

## Persona 1：Primary User - 網路可靠性 / SRE 操作者

### Evidence-based

- README 描述目標是從自己的 networks、regions、labs、edge nodes、private infrastructure 觀察 reachability、latency、packet loss、routes、probe health。
- Web routes 與 API 支援 Dashboard、Probes、Checks、Insight、Alerts、Results。
- Result APIs 支援 Ping latency/loss、TCP connect/failure、Traceroute runs/topology。
- Alerts 支援 incidents、rules、notifications，且 UI 每 30 秒 refetch incidents。

### Assumption

此人最可能是 SRE、network engineer、platform engineer、DevOps、NOC operator，或管理多地 VPS/edge/lab/private network 的技術操作者。repo 沒有直接寫出職稱，因此職稱是推論。

### 背景

Primary user 對「服務從特定真實網路看起來是否正常」負責。他們不只看雲端監控或單一 uptime check，而是需要知道不同 probe viewpoint 到目標服務的 latency、loss、TCP connect、route path 是否異常。

### 主要目標

- 確認服務是否從多個真實 locations/network boundaries 可達。
- 比較不同 region、ISP、hosting provider、lab/private network 的 latency/loss/route。
- 找出 route changes、hop movement、packet loss、TCP failure 等使用者感知問題的網路證據。
- 在 incidents 發生前或發生時取得可行的 probe/check/result context。
- 將 alerts 發送到團隊已使用的通知工具。

### 觸發情境

- 使用者回報「某地區連不上」或「延遲變高」。
- 某個 probe/check 的 latest status 變成 timeout/error/partial。
- Alert incident 開啟，Slack/Discord/Telegram/email/webhook 通知送出。
- 新服務、region、provider、edge node 上線，需要建立 baseline。
- 路由或 ISP/provider 變更後，需要比較前後 network path。

### 成功標準

- 能快速辨識影響範圍：哪個 probe、哪個 check、哪個 target、哪個時間窗。
- 能看到 latency、loss、success rate、connect time、route/hop/topology 變化。
- 能區分「目標服務故障」、「特定地區/ISP 路由異常」、「probe 離線或 stale」。
- 能用 alerts/incidents 與 public status page 對團隊或外部 stakeholder 溝通狀態。
- 能保留歷史 measurement，作為 incident review 或供應商溝通依據。

### 顧慮與風險

- Probe viewpoint 若命名、location 或 labels 不穩定，歷史比較會失真。
- Probe offline/stale 可能被誤解為服務異常。
- Selector 設錯會造成 checks 沒有跑在預期 probes 上。
- Alert threshold、window、min samples、cooldown 設錯會造成 noisy alerts 或 missed incidents。
- Traceroute alerts 目前 UI 顯示未支援，Primary user 不能假設 route change 可自動觸發 alert。
- DNS support 在 copy 中出現但 contract 未確認，若 Primary user 期待 DNS monitoring 會有落差。

### 目前產品支援度

**高**：Dashboard、Probes、Checks、Insight、Alerts、Results、labels/selectors、project-scoped access 都直接服務此 persona。

已支援：

- Probe fleet map、online count、heartbeat/status。
- Probe detail、assigned checks、labels、location、secret rotation/reinstall/upgrade commands。
- Ping/TCP/Traceroute checks 與 interval/config。
- Selector builder/advanced JSON/preview matched probes。
- Insight time range、refresh、probe/check/assignment scoping。
- Ping/TCP summary + series；Traceroute run timeline、hop diagnostics、route topology。
- Alert rules/incidents/notifications。

### 缺口

- Route/traceroute alerting 尚未支援或 UI 明確 disabled。
- DNS monitoring 目前無 contract/executor evidence。
- Incident acknowledgement/resolution actions 在 API status enum 中存在，但目前 contract/UI 未看到更新 incident 狀態的 operation。
- 沒看到 SLO/SLA、maintenance window、mute/silence、escalation policy、on-call rotation。
- 沒看到 built-in runbook、annotation、incident notes、postmortem export。

### 需要驗證的研究問題

- Primary user 最常用的 scope 是 probe、check、assignment，還是 project-wide overview？
- Alert threshold 預設值是否符合真實網路噪音？min samples/cooldown 是否容易理解？
- Traceroute topology 對 incident triage 是否足夠，還是需要 path diff、ASN/provider names、hop ownership、baseline comparison？
- 使用者如何命名 probes、labels、checks 才能長期維持可比較性？
- DNS check 是真需求、文案殘留，還是 roadmap？需要先釐清。

### JTBD

| 情境                                   | 我想要                                                | 以便                                               |
| -------------------------------------- | ----------------------------------------------------- | -------------------------------------------------- |
| 當某區使用者回報服務變慢或連不上       | 從該區附近 probe 查看 latency/loss/TCP/route evidence | 判斷是服務端、ISP、route、還是 probe 本身問題      |
| 當我新增一個 region/provider/edge host | 建立 probe 並讓它自動取得 checks assignments          | 立刻把新 viewpoint 納入既有監控                    |
| 當 alert incident 觸發                 | 打開 incident detail 與 Insight 時間窗                | 找到造成 alert 的 metric、probe、check、時間與變化 |
| 當網路路徑改變                         | 比較 traceroute runs/topology                         | 用可觀測證據與 provider 或團隊討論                 |

## Persona 2：Secondary User - 服務擁有者 / 技術分析與 API 使用者

### Evidence-based

- Project roles 包含 editor/viewer；read project 對所有 valid roles 開放，write labels/checks/probes/alerts 對 owner/admin/editor 開放。
- Docs/OpenAPI explorer 讓使用者 inspect routes、request bodies、response schemas；TypeSpec contract 產生 docs/public/openapi.json 與 web API types。
- Web 有 Checks、Labels、Insight、Status Pages、Members、Account Settings；route guards 讓登入使用者依 project access 使用。
- Assignments API 明確說 authenticated console users 可以 list effective probe-check assignments；probe runtime polling 走另一條 runtime endpoint。

### Assumption

Secondary user 可能是 application/service owner、backend engineer、support engineer、technical customer support、technical PM，或使用 API 建 automation 的 integrator。他們不一定部署 controller 或 probe，但會消費結果、維護少量 checks，或把 Netstamp data 接入其他流程。

### 背景

Secondary user 負責某些服務或使用 Netstamp 的資料做決策。他們需要可靠地知道「自己的 target 在哪些 networks 下的觀測結果」，但未必負責整個 self-host runtime 或 agent fleet。

### 主要目標

- 查看自己服務相關 checks 與 result history。
- 建立或調整 checks/labels/selectors，使 service monitoring 覆蓋正確 probes。
- 用 Insight 分析特定 target/probe 的時間窗。
- 使用 OpenAPI 了解可用 endpoints，建立 dashboard、report、automation 或 internal tooling。
- 接受 project invite，依 role 參與 project。

### 觸發情境

- 新服務、port 或 endpoint 需要納入 monitoring。
- incident 後需要查詢特定時間窗資料。
- 團隊要求把 result status 顯示在別的內部系統。
- 收到 project invite，需要進入 workspace 查看。
- 需要調整 labels/selectors 讓 checks 指派到正確 probe group。

### 成功標準

- 不需要理解所有 infra 細節，也能找到與自己服務相關的 checks/results。
- 能清楚看到 role 允許哪些操作，避免誤改 project-wide resources。
- API contract 與 UI 行為一致，automation 不會因文件漂移失效。
- Selector preview 能在儲存前避免錯誤指派。

### 顧慮與風險

- Project role/permission 若 UI 不夠清楚，使用者會不知道為何某些操作被拒絕。
- API docs 需要有效 session cookie，OpenAPI explorer 不能繞過 auth，這可能讓 API 初學者困惑。
- Advanced selector JSON 對非熟悉使用者可能有門檻。
- Checks/labels/probes 的關聯如果只靠 IDs 或短名，可能讓跨團隊使用者難以辨識 scope。

### 目前產品支援度

**中高**：結果查詢、OpenAPI、checks/labels/selectors、project invites、role-based access 都存在；但 role explanation、API onboarding、使用者任務導向 docs 仍偏薄。

已支援：

- OpenAPI explorer 與 generated contract。
- Account settings 與 pending project invites accept/reject。
- Checks/Labels UI 與 selector preview。
- Insight query by time/probe/check/assignment。
- Viewer 可讀 project；editor 可管理 labels/checks/probes/alerts。

### 缺口

- 缺少角色權限矩陣的產品內說明或 docs guide。
- 缺少 task-based API examples，例如「查某服務過去 24h latency/loss」。
- 未看到 saved views、shared Insight URLs 的正式設計說明，雖然 URL state 已存在。
- 未看到 per-service ownership、favorite checks、dashboard customization。

### 需要驗證的研究問題

- Secondary user 主要透過 UI 還是 API 消費資料？
- Editor 是否應能管理 probes？目前 domain 允許 editor `write:project_probes`，需驗證是否符合實際團隊分工。
- Viewer 是否需要匯出資料、複製圖表、分享 scope？
- OpenAPI explorer 對 authenticated routes 的操作流程是否足夠清楚？

### JTBD

| 情境                            | 我想要                                                  | 以便                                         |
| ------------------------------- | ------------------------------------------------------- | -------------------------------------------- |
| 當我負責的 service 上線         | 建立 Ping/TCP/Traceroute checks 並用 labels 指派 probes | 確認服務從重要 networks 的可達性與品質       |
| 當我收到 project invite         | 接受邀請並進入對應 project                              | 快速查看 team 已建立的 probes/checks/results |
| 當我需要做內部報表或 automation | 查 OpenAPI contract 與 result endpoints                 | 以穩定 schema 整合 Netstamp data             |
| 當 selector 邏輯變複雜          | 先 preview matched probes                               | 避免 checks 錯跑或漏跑                       |

## Persona 3：Admin / Operator - Self-host 與 Project 管理者

### Evidence-based

- Docker Compose self-host runtime 包含 PostgreSQL/TimescaleDB、migrate、netstamp app。
- Observability compose 包含 controller、nginx、demo、Grafana、VictoriaMetrics/VictoriaTraces/VictoriaLogs、Vector、postgres exporter。
- Server config 包含 APP_ENV、JWT secret、DB settings、demo mode、registration flag、alert evaluation、notification worker、SMTP、OTLP traces。
- Probe install assets 與 frontend install command 使用 Linux installer、systemd service install、probe ID/secret、controller URL。
- Project owner/admin 可管理 members/invites；owner 可 delete project，admin/owner 可 manage members。
- Server guidance 明確說沒有 global admin role；project access 是 membership roles。

### Assumption

Admin/operator 可以是同一個 primary SRE，也可以是獨立的 self-host maintainer、platform administrator、homelab operator 或 community infrastructure maintainer。repo 沒有明確區分「infra admin」和「project admin」兩個 persona，但 evidence 顯示兩種工作都存在。

### 背景

Admin/operator 負責讓 Netstamp 本身可靠運作：部署 controller/web/API、資料庫與 migrations、設定 secrets、HTTPS/reverse proxy、agent binaries、notification transports、observability stack，以及 project/member access。

### 主要目標

- 快速部署並升級 self-host Netstamp。
- 確保 controller、DB、migrations、web/API routing、healthz 正常。
- 保護 JWT secret、DB password、probe secret、SMTP credentials、notification credentials。
- 管理 project lifecycle、members、invites、roles。
- 建立、更新、重裝或輪替 probe agent secrets。
- 設定 alert evaluation/notification worker 與 SMTP/webhook targets。

### 觸發情境

- 首次 self-host deployment。
- 新 team/project 建立或成員變更。
- 新 probe host 要加入 fleet。
- Agent secret 外洩或 host 重建，需要 rotate/reinstall。
- Demo/public instance 需要 read-only mode。
- Alerts 沒有送出，需要檢查 notification worker/SMTP/webhook settings。
- Controller/API/docs/schema 有變更，需要 regenerate OpenAPI 或 deploy。

### 成功標準

- `docker compose up -d` 後 app、API、DB migration、healthz 正常。
- Probe agent 安裝後有 heartbeat，assignment polling 與 result submission 正常。
- 成員能依角色進入 project，owner/admin/editor/viewer 權限符合預期。
- Secrets 不被 log、UI 或 API 長期暴露；probe plaintext secret 只在 creation/rotation 回傳。
- Observability/Grafana 能協助排查 controller、DB、agent metrics/logs/traces。

### 顧慮與風險

- Self-host 部署若暴露 public，需要 operator 自行設定 HTTPS/reverse proxy；README 提醒 before exposing publicly 要改 `.env` 與放在 HTTPS 後。
- No global admin role：若 operator 期待跨 project admin console，目前沒有 evidence。
- Owner 不能 self-leave；last owner 保護可能需要清楚 UI。
- Probe secret 只在 creation/rotation 顯示，遺失後需 rotate；這是安全設計但會影響操作。
- SMTP/env/secrets/notification credentials 錯誤會導致 alert delivery failure。
- Demo mode backend 與 frontend flags 需一致，否則 UI 和 API 行為可能不一致。

### 目前產品支援度

**中高**：self-host path、env examples、compose、install assets、agent service commands、member management 都有 evidence；但 admin observability、deployment hardening、role education 還需要更完整 docs/product flows。

已支援：

- Docker Compose self-host runtime。
- Backend health endpoints。
- Probe Linux installer、binary、uninstaller、service install/update/reinstall commands。
- Project/member/invite/role operations。
- Account/project settings。
- Demo/read-only flags。
- Notification worker and SMTP/env configuration。

### 缺口

- 缺少完整 production hardening checklist，例如 TLS、proxy headers、backup/restore、DB retention、upgrade/rollback、secret rotation policy。
- 缺少 global operator console：跨 project、all users、all probes、system health 的管理 UI 無 evidence。
- 缺少 UI 內的 role/permission explanation。
- 缺少 agent install troubleshooting guide，例如 raw socket/CAP_NET_RAW、systemd logs、firewall/proxy、controller URL。
- 缺少 notification delivery diagnostics UI；目前有 test notification，但 worker/outbox 狀態未看到完整 UI。

### 需要驗證的研究問題

- Self-host operator 與 project owner 是否通常同一人？若不同，哪些權限應分開？
- Operator 最需要 CLI、docs 還是 UI 來管理 agents？
- Probe secret rotation/reinstall/upgrade command 是否符合實際 Linux fleet 管理流程？
- 是否需要 backup/restore、multi-project admin、audit logs、SSO 或 invite-less user provisioning？

### JTBD

| 情境                        | 我想要                                                     | 以便                                          |
| --------------------------- | ---------------------------------------------------------- | --------------------------------------------- |
| 當我首次 self-host Netstamp | 用 Compose、env、migrations 快速啟動 controller/web/API/DB | 建立可被團隊使用的 network observability 平台 |
| 當我新增 probe host         | 產生 install command 並確認 heartbeat                      | 確保新 viewpoint 真正上線並可收 assignments   |
| 當團隊成員變動              | 邀請、調整 role、移除 member                               | 控制誰能讀取或改動 project resources          |
| 當 secret 或 host 有風險    | rotate probe secret、reinstall/upgrade agent               | 恢復受信任的 result ingestion                 |

## Persona 4：Verifier / Recipient / External Stakeholder - 公開狀態查看者與通知接收者

### Evidence-based

- Public app route `/status/:slug` 不在 protected shell 下。
- Public Status API `/public/status-pages/{slug}` 不使用 session cookie auth。
- Public status response 包含 page summary、elements、public assignments、metrics、charts、active/recent incidents、generatedAt。
- Alerts notifications 支援 webhook、Slack、Discord、Telegram、email；notification configs 可被 rules 使用，且有 test notification operation。

### Assumption

此 persona 不是文件/certificate verifier。它是「network/service status verifier」或「incident notification recipient」：customer、partner、internal stakeholder、support team、on-call channel、community member 或 manager。repo 沒有明確外部 persona 文案，所以 stakeholder identity 需要研究驗證。

### 背景

這類使用者不一定有 Netstamp 帳號，也不一定能看到完整 project。他們只需要知道某些公開服務或 network paths 現在是否 operational、是否 degraded/down、是否有 open incident，以及最近 metrics 是否支持該狀態。

### 主要目標

- 從公開 status page 驗證服務/路徑目前狀態。
- 在 incident 發生時透過通知工具收到可行摘要。
- 不登入 console 也能理解 high-level status、open incidents、latest checks、latency/loss/connect/failure。
- 對外或對內溝通「目前問題是否仍存在」。

### 觸發情境

- 客戶或內部使用者覺得服務連線異常。
- On-call channel 收到 critical/warning alert。
- Support 要回覆「是否平台端異常」。
- Team 想對外分享部分 network status，而非完整 project console。

### 成功標準

- status page 能快速顯示 operational/degraded/down/unknown。
- 能看到 active incidents 與 generated timestamp。
- 能看到相關 assignment/check/probe metrics，但不暴露不該公開的 internal project details。
- 通知能送到既有渠道，且 test notification 可確認 destination。

### 顧慮與風險

- Public status page 可能揭露 probe names、probe locations、targets、metrics；需要驗證使用者是否需要 masking/alias。
- 外部 viewer 無法 ack、comment、subscribe 或 request update；目前沒有這些 evidence。
- Notifications 是 delivery target，不代表 recipient 在 Netstamp 內有工作流。
- 如果 status page elements 沒有配置，外部 viewer 只看到 empty state。

### 目前產品支援度

**中**：公開 status page 與 notifications 有明確支援；但外部 stakeholder 的完整溝通流程仍有限。

已支援：

- Public status pages with slugs。
- Page elements/folders/assignment groups。
- Public metrics/charts/incidents。
- Alert notification destinations and test。

### 缺口

- 無 evidence 顯示 public subscriber management、RSS/webhook subscriptions、incident update posts、maintenance windows。
- 無 evidence 顯示 external-friendly copy customization、branding/custom domain、privacy controls。
- 無 evidence 顯示 notification recipients 能點擊到 scoped incident view 或 public status incident detail。

### 需要驗證的研究問題

- Public status page 的主要受眾是 customers、internal teams、community，還是 incident responders？
- 外部 stakeholder 需要看到 probe/check/target 真名，還是需要 alias/redaction？
- 通知內容應包含多少 detail？是否需要 link 到 public page 或 internal console？
- 是否需要 maintenance/incident updates、subscriptions、history、SLA reporting？

### JTBD

| 情境                       | 我想要                  | 以便                                      |
| -------------------------- | ----------------------- | ----------------------------------------- |
| 當我懷疑服務或網路路徑異常 | 打開公開 status page    | 不登入也能確認目前狀態與最新 metrics      |
| 當 alert 發生              | 在既有 channel 收到通知 | 立即知道要不要介入或通知使用者            |
| 當我需要向外部說明狀態     | 分享 status page        | 用可讀、受控的資料取代完整 console access |

## Cross-Persona JTBD Table

| # | User type | Evidence level | Situation | Job-to-be-done | Success outcome | Current support | Gaps / risks |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | Primary | Evidence-based | 需要知道服務從真實 networks 看起來如何 | 佈署 probes 到 VPS、lab、edge、private infra | Dashboard 顯示 online/stale 狀態與位置 | Probes page、map、heartbeat、install wizard | Probe 命名/位置/labels 不穩定會破壞歷史比較 |
| 2 | Primary / Admin | Evidence-based | 新 probe host 要加入 | 建立 probe、取得 secret、安裝 agent、等待 heartbeat | Controller 收到 heartbeat，probe online | NewProbeDrawer、install assets、runtime heartbeat | Linux privilege、firewall、controller URL、secret handling docs 仍需強化 |
| 3 | Primary / Secondary | Evidence-based | 新 target 要被監控 | 建立 Ping/TCP/Traceroute check，設定 interval/config/selector | assignments 覆蓋正確 probes | ChecksPage、selector builder、preview API | DNS copy 與 contract 不一致；selector 複雜度可能高 |
| 4 | Primary / Secondary | Evidence-based | incident 或 performance regression | 在 Insight 查 probe/check/assignment time window | 找到 latency/loss/connect/route evidence | Insight page、result APIs、charts/topology | 缺少 annotation、baseline diff、postmortem export |
| 5 | Primary | Evidence-based | 需要自動通知 | 建立 alert rule 與 notification target | Incident firing 時通知正確 channel | AlertsPage、rules/incidents/notifications、test notification | Traceroute alerts unavailable；缺少 mute/escalation/ack update operation evidence |
| 6 | Admin | Evidence-based | 團隊成員加入/離開 | 管理 project members、invites、roles | 正確授權且不移除 last owner | MembersPage、Settings invites、project permissions | 無 global admin；role guidance 不足 |
| 7 | Admin | Evidence-based | self-host deployment | 啟動 controller/web/API/DB/migrations，設定 secrets | App 可用、healthz ok、DB migrated | Docker Compose、env examples、docs | Production hardening/backup/restore/upgrade checklist 不完整 |
| 8 | Secondary | Evidence-based | 需要 automation | 查 OpenAPI contract 與 endpoints | 產生 aligned integration/script | TypeSpec、OpenAPI explorer、generated web types | Authenticated route try-out、task examples 需要補 |
| 9 | External stakeholder | Evidence-based + assumption | 需要公開查狀態 | 查看 `/status/:slug` | 看到 operational/degraded/down、incidents、metrics | PublicStatusPage、public API | 受眾、privacy、subscription、incident updates 需研究 |
| 10 | External recipient | Evidence-based + assumption | Alert 發生 | 接收 Slack/Discord/Telegram/email/webhook | 及時知道 incident 摘要 | Notification config + test | Recipient 不一定有 console access；通知內容/links 未驗證 |

## Scenarios

### Scenario 1：首次建立 project 並部署第一個 probe

- **User type**：Admin/operator + Primary user。
- **Evidence level**：Evidence-based。
- **Trigger**：新 self-host instance 或新 team workspace 沒有 projects/probes。
- **Preconditions**：使用者已註冊/登入；project creation enabled；controller/API 可用。
- **Flow**：
  1. App route guard 發現沒有 `projectRef`，導向 onboarding。
  2. 使用者輸入 project name，系統建立 project slug。
  3. 可選填 member emails，送出 project invites。
  4. Onboarding 完成後導向 probe fleet / create probe。
  5. 使用者輸入 probe identity、location/coordinates。
  6. 系統建立 probe，回傳 plaintext secret 與 install command。
  7. 使用者在 Linux host 執行 installer/service install。
  8. UI poll probe detail，直到 controller 收到 heartbeat。
- **Success criteria**：Project created；first probe online；probe 有 location；可進入 Dashboard/Probes。
- **Risks / gaps**：如果 agent host 沒有權限、controller URL 不通、secret 遺失，使用者需要更明確 troubleshooting；目前 DNS expectation 也需避免在 onboarding 中誤導。
- **Research questions**：使用者是否會在 onboarding 就邀請 members？第一個 probe 需要 location 必填嗎？是否需要 copy/paste command 之外的 package manager/container install path？

### Scenario 2：用 labels/selectors 建立 region-specific checks

- **User type**：Primary user + Secondary user。
- **Evidence level**：Evidence-based。
- **Trigger**：要監控 `api.example.com` 從 Tokyo probes 的 latency/loss，或只對特定 provider/region 跑 TCP check。
- **Preconditions**：Project 已有 probes 與 labels；使用者 role 可 write checks/labels。
- **Flow**：
  1. 在 Labels 建立 `region:tokyo`、`provider:vps` 等 labels。
  2. 在 Probe detail 勾選相關 labels。
  3. 在 Checks 建立 check，選 Ping/TCP/Traceroute、target、interval、config。
  4. 使用 selector builder 或 advanced JSON 設定 matching rule。
  5. 點 `Preview selector` 查看 matched probes。
  6. 儲存 check，查看 assignments 或 Insight。
- **Success criteria**：Matched probes 符合預期；assignments 建立；result 開始進入 latest/series。
- **Risks / gaps**：Selector mode/advanced JSON 對非 expert 可能太抽象；label delete 會 refresh assignments，需讓使用者理解 blast radius。
- **Research questions**：使用者最常用的 label keys 是 region/provider/network/role 嗎？需要 selector templates 嗎？

### Scenario 3：Alert incident 後定位影響範圍

- **User type**：Primary user。
- **Evidence level**：Evidence-based。
- **Trigger**：Open incident 出現在 Alerts summary，或外部 channel 收到 critical notification。
- **Preconditions**：Alert rule enabled；notification 已設定；result ingestion 正常。
- **Flow**：
  1. 使用者打開 Alerts，查看 open incidents。
  2. 點 incident detail，確認 severity、probe、check、target、last value、metric threshold、timeline。
  3. 依 incident 的 probe/check 到 Insight，選相同或更窄 time window。
  4. 對 Ping 看 latency/loss/success rate；對 TCP 看 connect/failure；對 Traceroute 看 runs/hops/topology。
  5. 判斷是否為 route issue、service issue、probe issue 或 transient noise。
  6. 必要時更新 alert rule threshold/window/cooldown 或 notification target。
- **Success criteria**：使用者能在幾分鐘內拿到可解釋 evidence，並決定 incident response。
- **Risks / gaps**：目前沒看到 incident ack/resolution mutation；traceroute alert metrics disabled；缺少 runbook/notes。
- **Research questions**：Incident detail 是否需要直接 link 到 scoped Insight？是否需要「compare before/after」和「copy incident summary」？

### Scenario 4：Self-host operator 做 production hardening

- **User type**：Admin/operator。
- **Evidence level**：Evidence-based + assumption。
- **Trigger**：要把 Netstamp 從 local/demo 改為 publicly exposed production。
- **Preconditions**：Operator 有 server/Docker/DB/reverse proxy access。
- **Flow**：
  1. 下載 compose/example env，設定 `DATABASE_PASSWORD`、`AUTH_JWT_SECRET`、`APP_ENV=production`、固定 version tag。
  2. 透過 reverse proxy 提供 HTTPS。
  3. 啟動 migrations 與 app container。
  4. 設定 SMTP 或 webhook notification targets。
  5. 可選啟動 observability compose/Grafana。
  6. 檢查 `/api/v1/healthz`、登入、project/probe heartbeat。
- **Success criteria**：Controller/web/API/DB 可用；secrets 不用預設值；health OK；notifications 可 test；probe results 正常寫入。
- **Risks / gaps**：README 提醒要 HTTPS 與改 secrets，但缺少完整 production checklist、backup/restore、retention tuning、upgrade rollback。
- **Research questions**：Operator 需要 Helm/Kubernetes、systemd package、Terraform、or Cloudflare Tunnel docs 嗎？備份與升級是 MVP 前需求嗎？

### Scenario 5：建立 public status page 給外部 stakeholder

- **User type**：Admin/operator + Verifier/external stakeholder。
- **Evidence level**：Evidence-based。
- **Trigger**：團隊想公開部分服務或 network measurements，不提供完整 console access。
- **Preconditions**：Project 有 assignments/results；使用者 role 可 manage public status pages。
- **Flow**：
  1. 在 Status Pages 建立 page：slug、title、description、enabled、chart mode/range。
  2. 新增 folder 或 assignment group elements。
  3. 選 `all_check` 或 selected assignments。
  4. 開啟 `/status/:slug` public page。
  5. 外部 viewer 查看 overall status、open incidents、metrics/charts、latest timestamp。
- **Success criteria**：外部 viewer 不登入即可理解 operational/degraded/down 與目前 incidents。
- **Risks / gaps**：可能暴露 internal probe/check/target 命名；缺少 branding/privacy/alias/subscription/maintenance update evidence。
- **Research questions**：Public status page 應面向客戶還是內部 stakeholder？哪些欄位需要隱藏或 alias？

### Scenario 6：用 OpenAPI 建立內部 automation

- **User type**：Secondary user。
- **Evidence level**：Evidence-based。
- **Trigger**：團隊要把 Netstamp result summary 放入內部 dashboard，或用 script 批量查 latest results。
- **Preconditions**：使用者有 session/auth；懂 API；OpenAPI contract 是最新。
- **Flow**：
  1. 打開 docs `/openapi/` 或 backend non-production `/api/v1/docs`。
  2. 查看 projects/checks/probes/results endpoints 與 schemas。
  3. 使用 session-authenticated endpoints 查 project refs、assignments、result series/insight。
  4. 把回應轉入內部 dashboard/report。
- **Success criteria**：Script 能穩定查資料；schema 與 web types 一致；route changes 經 TypeSpec regenerate 後保持同步。
- **Risks / gaps**：Explorer 不繞過 auth；缺少 token/API key auth evidence；session cookie 對 automation 不一定理想。
- **Research questions**：Integrators 需要 API keys/service accounts 嗎？哪些 result query 是最常用 automation path？

## 重要產品假設清單

- **Assumption A**：Primary user 是 network/SRE/platform persona，而不是 general manager。原因是產品語言、routes、metrics、agent runtime 都高度技術化。
- **Assumption B**：Secondary user 可能需要 read-only 或 editor access 來服務特定 service/team。原因是 project roles 存在 viewer/editor，但 repo 未明確說明職能。
- **Assumption C**：Admin/operator 可分為 project admin 與 self-host infra operator。原因是 project role 與 deployment/agent/systemd/env work 是兩種不同工作，但可能由同一人完成。
- **Assumption D**：Verifier/recipient 是 network status viewer/notification recipient，不是法務或證書驗證者。原因是 public status pages/notifications 有 evidence，document stamp/certificate 無 evidence。

## 後續研究優先問題

1. 主要購買/採用動機是 incident triage、multi-region visibility、homelab/community transparency，還是 self-host privacy/control？
2. Primary users 的最小有用 probe fleet 是幾個？常見 placement patterns 是 home ISP、VPS region、office/lab、edge POP 還是 customer network？
3. Ping/TCP/Traceroute 的 priority 如何排序？DNS 是否為真需求或文案殘留？
4. Alerting 是否足夠：是否需要 ack/resolution actions、silence/maintenance windows、route-change alerts、escalation policy？
5. Public status pages 要服務誰？是否需要 custom branding、privacy controls、subscriptions、incident updates？
6. API consumers 是否需要 API keys/service accounts，而不只是 browser session cookie？
7. Project role model 是否符合 team workflow？Editor 是否應該能管理 probes/alerts，Viewer 是否需要 export/share？
8. Self-host operators 最需要哪些 deployment targets、backup/restore、upgrade、observability docs？
