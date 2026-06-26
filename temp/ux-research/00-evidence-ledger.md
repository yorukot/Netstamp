# 00. Evidence Ledger

本 ledger 是主 agent 對 repository、既有 FigJam board、subagent 研究輸出的交叉整理。用途是把所有 UX claim 分成「code-backed fact」、「board-backed observation」、「synthetic/assumption」、「needs validation」，避免把推測寫成事實。

## Confidence Legend

| Confidence | 定義                                                                                                         |
| ---------- | ------------------------------------------------------------------------------------------------------------ |
| High       | 有 repository code、API contract、README、router、server handler、DB/migration 或既有 FigJam node 直接支持。 |
| Medium     | 有多個間接 evidence 或 UI/contract 明確暗示，但仍需要實機流程或產品 owner 確認。                             |
| Low        | 主要來自假設、synthetic interview、競品 pattern 或命名推論；不可當作既有產品事實。                           |

## Evidence Types

| Evidence type           | 說明                                                                             |
| ----------------------- | -------------------------------------------------------------------------------- |
| Code                    | web/server/api/docs/deployment/config/test/source file。                         |
| FigJam                  | 目前 board 上既有 content、template 或 function mind map。                       |
| Research synthesis      | Subagent 01-09 產出的整理，通常仍回指 code evidence。                            |
| Competitive source      | 外部公開文件或產品 pattern，只能支持「可借鑑模式」，不能支持 Netstamp 已有功能。 |
| Assumption / Hypothesis | 需要後續訪談、analytics 或產品策略確認。                                         |

## Ledger

| ID | Claim | Evidence source | Evidence type | Confidence | Related UX artifact | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| E001 | Netstamp 目前定位是 self-hosted network observability，不是文件蓋章或憑證驗證產品。 | `README.md:4`, `README.md:20-24` | Code/docs | High | Product Context, Research Overview | README 直接寫「Self-hosted network observability from probes you control」。 |
| E002 | 主要價值是從使用者控制的 probes/viewpoints 觀測 reachability、latency、packet loss、routes、probe health。 | `README.md:22-33` | Code/docs | High | Value Proposition, Personas, Journey | 支持 SRE/NOC/IT Ops/persona 推導。 |
| E003 | 目前已列出的 core features 包含 controller web app + Go API、probe agents、Ping/TCP/traceroute checks、project roles、labels、dashboards、alerts、notifications、Docker、OpenAPI。 | `README.md:35-46` | Code/docs | High | Feature Inventory | README features 與 API/router 大致一致；README 未列 Slack/email，但 code 有。 |
| E004 | Web app route IA 包含 dashboard、probes、labels、insight、checks、alerts、status pages、members、project settings、account settings。 | `web/src/routes/routePaths.ts:10-22`, `web/src/routes/router.tsx:51-76` | Code | High | IA Map, Feature Inventory | 這是 app 的主要 navigation surface。 |
| E005 | Public status page 是公開 route，不需登入。 | `web/src/routes/router.tsx:16` | Code | High | Public Status Journey, Trust UX | `/status/:slug` 在 protected shell 之外。 |
| E006 | 未登入會導到 login，已登入進 auth pages 會導到 dashboard，無 project 會導到 onboarding。 | `web/src/routes/guards.tsx:19-35`, `web/src/routes/guards.tsx:61-87` | Code | High | Current Journey | 多處 loading 回 `null`，成為 visibility 風險。 |
| E007 | Project route boundary 會把 invalid project ref fallback 到第一個 project 或 onboarding。 | `web/src/routes/guards.tsx:90-126` | Code | High | Error/Permission Journey | 缺少 project-not-found explanation 是 UX 風險。 |
| E008 | API contract 明確列出 System/Auth/Users/Projects/Members/Invites/Labels/Checks/Assignments/Probes/Results/Alerts/Public Status/Probe Runtime/Install。 | `api/main.tsp:33-52` | Code/API | High | Feature Inventory, Evidence Map | 支持產品功能邊界。 |
| E009 | Server router 註冊 install、auth、user、project、alert、assignment、label、check、probe、publicstatus、result、proberuntime handlers。 | `server/internal/controller/transport/http/router.go:125-142` | Code/server | High | Service Blueprint | 對應 frontstage/backstage service map。 |
| E010 | Server 提供 `/openapi.json` 與 `/docs`。 | `server/internal/controller/transport/http/router.go:144-164` | Code/server | High | Developer/API Support | 支持 API-first / developer docs 建議。 |
| E011 | Demo mode 在 server 以 read-only middleware 阻擋 unsafe methods，web 也有 read-only/demo feature flags。 | `server/internal/controller/transport/http/router.go:99-104`, `web/src/shared/config/features.ts:19-37` | Code/config | High | Error Recovery, Demo UX | 需要在 UI action 前更清楚顯示唯讀限制。 |
| E012 | Registration、project creation、credential changes 受 frontend env flags 控制。 | `web/src/shared/config/features.ts:22-25` | Code/config | High | Onboarding, Permissions | 適用 demo/self-host operator 場景。 |
| E013 | Auth register CTA 在 register mode 顯示 `Create project`，但 submit 只建立 account 後導到 onboarding。 | `web/src/features/auth/components/AuthPage.tsx:59-79`, `web/src/features/auth/components/AuthPage.tsx:143-145` | Code/UI | High | UX Audit | 這是具體 content mismatch。 |
| E014 | Onboarding 建立第一個 project、用 slug conflict retry、可邀請 members，但 invite role 固定 viewer。 | `web/src/features/auth/components/OnboardingPage.tsx:195-241`, `web/src/features/auth/components/OnboardingPage.tsx:216-218` | Code/UI | High | Current Journey, Recommended Features | 支持「first-run checklist」與「role selection」建議。 |
| E015 | Onboarding 成功後導向 probe fleet/create probe，沒有完整 checklist 到 first check/result/alert。 | `web/src/features/auth/components/OnboardingPage.tsx:290-296` | Code/UI | High | Activation Journey | 核心啟用斷點。 |
| E016 | Project creation disabled 時 onboarding 只告知 no project assigned，可 logout 或 ask operator。 | `web/src/features/auth/components/OnboardingPage.tsx:248-273` | Code/UI | High | Permission/No Access Journey | 需要 request access / invite status。 |
| E017 | Check API model 只定義 `ping`, `tcp`, `traceroute` 三種 check type。 | `api/models/check.tsp:3-8` | Code/API | High | Product Context, No-evidence Map | DNS/HTTP 不可寫成已實作。 |
| E018 | Web check config 只包含 Ping/TCP/Traceroute 表單與 validation。 | `web/src/features/checks/data/checkConfig.ts:47-120` | Code/UI | High | Feature Inventory | 與 API contract 一致。 |
| E019 | Probe wizard 是兩步：Name 與 Install，需地點/座標，建立 probe 後顯示 registration secret、install command、installer/binary/uninstaller links，輪詢 heartbeat。 | `web/src/features/probes/components/NewProbeDrawer.tsx:25-28`, `web/src/features/probes/components/NewProbeDrawer.tsx:73-83`, `web/src/features/probes/components/NewProbeDrawer.tsx:235-252`, `web/src/features/probes/components/NewProbeDrawer.tsx:340-404` | Code/UI | High | Probe Journey, Service Blueprint | 支持 install troubleshooting、secret hygiene、first heartbeat 建議。 |
| E020 | Probe location search 使用 Nominatim helper 與 manual coordinates。 | `web/src/features/probes/components/NewProbeDrawer.tsx:1-9`, `web/src/features/probes/components/NewProbeDrawer.tsx:197-233`, `web/src/features/probes/components/NewProbeDrawer.tsx:280-321` | Code/UI | High | UX Audit, Privacy Messaging | 需要第三方 geocoding/privacy 說明。 |
| E021 | Alerts UI 有 Incidents、Rules、Notifications 三個 tabs 與 summary cards。 | `web/src/features/alerts/components/alertPageModel.ts:78-82`, `web/src/features/alerts/components/AlertsPage.tsx:465-587` | Code/UI | High | Alert Journey | 支持 incident workflow 盤點。 |
| E022 | Alert rules 支援 ping/TCP metrics；traceroute alert 在 UI 明確 disabled。 | `web/src/features/alerts/components/alertPageModel.ts:119-139`, `web/src/features/alerts/components/AlertsPage.tsx:771-789` | Code/UI | High | Feature Gaps | 不可把 traceroute alert 當已完成。 |
| E023 | Notification channel types 包含 webhook、Slack、Discord、Telegram、Email。 | `web/src/features/alerts/components/alertPageModel.ts:160-168`, `web/src/features/alerts/components/AlertsPage.tsx:891-987` | Code/UI | High | Integrations | 既有 FigJam 提到 LINE，但 code 沒有 LINE channel evidence。 |
| E024 | Notification test 會回報 delivered 或 failed，但沒有看到持久化 delivery status center。 | `web/src/features/alerts/components/AlertsPage.tsx:452-463` | Code/UI | Medium | Recommended Features | 支持「Notification Delivery Status Center」。 |
| E025 | Incident detail 顯示 what happened、timeline、notifications，但沒有 ack/manual resolve/deep-link analysis workspace。 | `web/src/features/alerts/components/AlertsPage.tsx:668-707` | Code/UI | High | UX Audit, Incident Review | 支持 incident review workspace。 |
| E026 | Public status page 顯示 Generated time、open incidents、elements、assignment rows、metrics/charts。 | `web/src/features/status-pages/components/PublicStatusPage.tsx:47-69`, `web/src/features/status-pages/components/PublicStatusPage.tsx:75-160` | Code/UI | High | Trust UX, Public Status Journey | 已有信任基礎，但缺少 methodology/disclaimer/redaction。 |
| E027 | Public status page 的 404 文案把 disabled 與 not exist 合併。 | `web/src/features/status-pages/components/PublicStatusPage.tsx:32-42` | Code/UI | High | Error Recovery | 對外 stakeholder 可能不知狀態。 |
| E028 | Public assignment rows 會顯示 check target、probe name 等資訊。 | `web/src/features/status-pages/components/PublicStatusPage.tsx:170-183` | Code/UI | High | Privacy/Redaction | 需要公開資料紅線與 alias/redaction。 |
| E029 | Project role policy 是 owner/admin/editor/viewer；read all roles，write/manage actions 依 role 限制，delete project owner only。 | `server/internal/domain/project/permission.go:5-30`, `server/internal/domain/project/permission.go:32-55` | Code/server | High | RBAC, Permissions | 支持 role visibility 與 audit log 建議。 |
| E030 | API docs and generated contract flow 是 workspace convention。 | `AGENTS.md`, `api/AGENTS.md`, `api/main.tsp:33-52` | Repo guide/API | High | Developer/API Support | `pnpm generate:openapi` 需保持 contract drift guard。 |
| E031 | Existing FigJam board already has a Functions mind map in Chinese with color legend: red no feature, yellow planned/in progress, black existing backend-oriented. | FigJam file `lJxnGVfbIM1WmiiHGUmVzx`, section `Functions`, node `1:1263` and sticky legend from get_figjam readback | FigJam | High | Existing Board Preservation, Feature Mind Map | New research should preserve and extend, not overwrite. |
| E032 | Existing FigJam board contains persona/journey templates and many placeholder sticky notes, not evidence-backed research synthesis. | FigJam get_figjam readback from board root `0:1` | FigJam | High | FigJam Structure | New board sections should label methodology/evidence. |
| E033 | No code evidence was found for document stamping, certificate issuing, verification portal, signature/notary workflows, billing/subscription/plan limits, import/export of measurements. | `temp/ux-research/01-feature-inventory.md`, scans across `web/src/routes`, `api/services`, `server/internal`, migrations | Research synthesis/code scan | High | Product Boundaries, Assumptions & Risks | Must be labeled no-evidence / assumption. |
| E034 | Docs/landing may mention DNS in places, but API/server executor evidence currently supports Ping/TCP/traceroute only. | `api/models/check.tsp:3`, `temp/ux-research/01-feature-inventory.md`, `temp/ux-research/09-recommended-features-roadmap.md` | Research synthesis/code | High | UX Audit, Copy Risk | Treat DNS as copy/roadmap risk unless backend contract exists. |
| E035 | Competitive pattern sources support private probes, result drilldown, incident/status communication, notification test, probe health, public status components. | `temp/ux-research/05-competitive-patterns.md` with Grafana, Checkly, Datadog, Uptime Kuma, Better Stack, Atlassian Statuspage, Prometheus Blackbox, Upptime, OneUptime, Healthchecks, OpenTimestamps/RFC3161 | Competitive source | Medium | Competitive Patterns, Recommended Features | Supports pattern rationale only, not Netstamp current-state facts. |
| E036 | Synthetic interview outputs are explicitly simulated and cannot be treated as real user evidence. | `temp/ux-research/07-simulated-interviews.md`, `temp/ux-research/08-affinity-synthesis.md` | Synthetic research | Low for facts, Medium for hypotheses | Synthetic Interview Findings, Affinity Map | Useful for hypotheses and interview planning. |
| E037 | Primary users are likely SRE/platform engineer, NOC/operator, IT operations manager; external public status reader/reviewer is secondary. | `README.md:22-33`, routes/API, `temp/ux-research/02-users-jtbd-scenarios.md` | Code-backed inference | Medium | Personas/JTBD | Needs real recruitment validation. |
| E038 | Highest activation risk: user can create account/project/probe but does not have a guided path to first check, first result, first alert, or first status page. | E013-E019 plus `temp/ux-research/03-journey-service-blueprint.md` | Code-backed synthesis | High | Current/Future Journey, Roadmap | Drives top recommendation. |
| E039 | Highest trust risk: results/status/incident pages need clearer source, freshness, sample count, failure reason, and evidence limits. | E025-E028 plus competitive patterns in `05` | Code-backed synthesis + pattern | High | Trust/Security/Verification UX | Network-result verification, not document verification. |
| E040 | Highest governance risk: role permission, sensitive action confirmation, secret rotation, public redaction, audit trail are not consistently surfaced as user-facing safeguards. | E019, E028, E029, `temp/ux-research/04-heuristic-ux-audit.md` | Code-backed synthesis | High | UX Audit, Recommended Features | Especially relevant to enterprise/self-host operators. |

## Evidence Gaps To Preserve In FigJam

| Gap                                                              | Current status                                                        | Label to use                                |
| ---------------------------------------------------------------- | --------------------------------------------------------------------- | ------------------------------------------- |
| Document stamping / certificate verification / file notarization | No implementation evidence found.                                     | `[No code evidence] [Assumption]`           |
| DNS / HTTP checks                                                | Not in current TypeSpec check type union or UI check config evidence. | `[Evidence gap] [Do not claim as existing]` |
| Billing / quota / subscription limits                            | No route/API/schema/migration evidence.                               | `[No code evidence]`                        |
| Measurement export/import / PDF/CSV reports                      | No dedicated route/API evidence.                                      | `[Missing feature]`                         |
| LINE notification channel                                        | Existing FigJam idea, no code evidence.                               | `[FigJam idea] [No code evidence]`          |
| Global admin / org hierarchy                                     | Project-scoped roles only.                                            | `[Assumption]`                              |
| Real user interviews                                             | Only synthetic interviews in this round.                              | `[Synthetic interview] [Needs validation]`  |
