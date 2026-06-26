# Subagent 6 - Research Blocks and FigJam Structure

本文定義 Netstamp UX research FigJam board 要補齊的 24 個研究區塊。內容依據 repo evidence 規劃，核心產品範圍鎖定 network observability、probe、check、result、alert、public status 與 agent runtime。凡涉及 document stamp、certificate、timestamp authority、notarization 等方向，目前未在 `api/`、`server/`、`web/`、`docs/`、`packages/` 找到直接產品 evidence，必須標成 assumption 或待驗證，不得放入已實作功能區。

## Board 全域規則

### 畫布網格

- Board 採 4 欄 x 6 列，共 24 個 FigJam frame。
- 每個 frame 建議尺寸：`1080 x 720`。
- Frame 間距：水平 `140`，垂直 `140`。
- 座標原點從左上角開始：C1 `x=0`、C2 `x=1220`、C3 `x=2440`、C4 `x=3660`；R1 `y=0`、R2 `y=860`、R3 `y=1720`、R4 `y=2580`、R5 `y=3440`、R6 `y=4300`。
- 主閱讀流：左至右、上至下。每列尾端以灰色細線連到下一列第一區。
- 跨區高價值連線使用粗線；高風險或最高優先級連線使用深色線或紅色線。

### Sticky note 顏色規範

- 藍：已實作或有 repo evidence 的功能、流程、資料模型、API、UI。
- 綠：使用者需求、JTBD、目標、成功條件。
- 黃：洞察、研究發現、模式。
- 橘：痛點、摩擦、破碎流程、認知負擔。
- 紅：高風險、阻塞、不應誤導的結論。
- 紫：機會、設計方向、產品策略。
- 粉：假設，包含無 evidence 的 document stamp/certificate 類內容。
- 灰：待驗證問題、research backlog、未知狀態。
- 白：說明文字、區塊導讀、方法、legend。
- 深色底或粗框：最高優先級項目，僅用於每區最重要 1-3 張 sticky。

### Evidence tag 規則

每張 sticky 右下角必須有 tag。格式如下：

- 已實作 evidence：`[E:<area>:<slug>]`，並在 sticky body 最後一行標出來源路徑，例如 `src: api/services/checks.tsp; server/internal/domain/check/check.go`。
- 使用者需求或 JTBD：`[J:<persona>:<slug>]`，若是從功能 evidence 推導，要追加 `derived-from: [E:...]`。
- 洞察：`[I:<cluster>:<slug>]`，至少連回 1 張 evidence、訪談筆記或 assumption sticky。
- 痛點：`[P:<journey-step>:<slug>]`，至少連回 current-state journey 的 step。
- 高風險：`[R:<risk>:<slug>]`，必須連到對應 mitigation 或 open question。
- 機會：`[O:<area>:<slug>]`，必須連到 pain point 或 JTBD。
- 假設：`[A:<topic>:<slug>]`，body 需寫 `assumption, no direct repo evidence`。
- 待驗證：`[V:<method>:<slug>]`，需寫建議驗證方式，例如 interview、usability test、code audit、analytics。
- Synthetic interview notes：必須標 `synthetic`，格式為 `[SYN:<persona>:<slug>]`，不得冒充真實訪談。

### 建議 evidence tag 字典

- `[E:API:checks]`：`api/services/checks.tsp`、`api/models/check.tsp`，有 Ping/TCP/Traceroute check CRUD、target、selector、interval、type-specific config。
- `[E:API:probes]`：`api/services/probes.tsp`、`api/models/probe.tsp`，有 probe CRUD、location、coordinates、labels、status、secret rotation。
- `[E:API:runtime]`：`api/services/probe-runtime.tsp`、`server/internal/controller/application/proberuntime/service.go`，有 hello、heartbeat、IP family capability、assignments、result submission。
- `[E:AGENT:runtime]`：`server/internal/agent/runtime/service.go`、`server/internal/agent/scheduling/*`，有 agent hello、heartbeat loop、assignment loop、scheduler、worker pool、result submitter。
- `[E:API:results]`：`api/services/results.tsp`、`api/models/result.tsp`，有 latest、ping/tcp series、ping/tcp insight、traceroute runs/insight/topology。
- `[E:WEB:insight]`：`web/src/features/insight/components/InsightPage.tsx`，有 scope、time range、refresh、assignment selection、Ping/TCP/Traceroute panels。
- `[E:API:alerts]`：`api/services/alerts.tsp`、`api/models/alert.tsp`、`server/internal/domain/alert/alert.go`，有 rules、incidents、notifications、severity、cooldown、evaluation state。
- `[E:SRV:alerteval]`：`server/internal/controller/application/alerteval/service.go`，有 ping/tcp alert evaluation、incident opened/resolved、notification enqueue；traceroute alert evaluation 目前跳過。
- `[E:WEB:alerts]`：`web/src/features/alerts/components/AlertsPage.tsx`，有 Incidents、Rules、Notifications tabs、summary cards、filter、create/test/delete notification。
- `[E:API:status]`：`api/services/public-status.tsp`、`api/models/public-status.tsp`，有 public status pages、elements、assignment groups、metrics、active/recent incidents。
- `[E:WEB:dashboard]`：`web/src/features/dashboard/components/DashboardPage.tsx`，有 Probes Online、Active Checks、NetworkMap。
- `[E:WEB:probes]`：`web/src/features/probes/components/ProbesPage.tsx`、`ProbeDetail.tsx`，有 fleet/map view、probe detail、assignment table、service commands、secret rotation、location editing。
- `[E:WEB:checks]`：`web/src/features/checks/components/ChecksPage.tsx`、`ChecksTable.tsx`，有 check list/editor、type filter、selector builder、preview selector、batch delete、duplicate。
- `[E:WEB:routes]`：`web/src/routes/sidebarItems.ts`、`web/src/routes/router.tsx`，有 Dashboard、Probes、Checks、Alerts、Status、Labels、Insight、Members、Settings。
- `[A:DOCSTAMP:no-evidence]`：document stamp/certificate 類，focused search 無直接產品 evidence，僅能作為假設或待驗證。

## 24 區結構

### 01. Research Overview

- 標題：`Research Overview - Netstamp UX Research Board`
- 內容格式：一張白色說明卡說明本 board 目標、範圍、非範圍；下方放 6 張 summary sticky：`已實作核心`、`主要使用者`、`主要任務`、`主要風險`、`最大機會`、`待驗證假設`。
- Sticky 顏色：白色說明；藍色表示 network observability/probe/check/result/alert evidence；粉色標 document stamp/certificate no-evidence；紅色標高風險；紫色標最大機會。
- 連線方式：藍色核心 evidence 連到 `04 Evidence Map`；粉色 no-evidence 連到 `09 Assumptions & Risks` 與 `23 Open Questions`；紫色機會連到 `18 Opportunity Areas`。
- 排版：上方 1 張橫向白卡；中段 3 x 2 summary sticky；底部放一條色彩 legend。
- 尺寸/位置建議：`x=0, y=0, w=1080, h=720`。白卡 `980 x 120`；summary sticky `280 x 160`；legend 高 `80`。
- Evidence tag 規則：每張 summary sticky 至少 1 個 tag。產品範圍 sticky 必須使用 `[E:API:*]` 或 `[E:WEB:*]`；無 evidence 的 document stamp/certificate 必須用 `[A:DOCSTAMP:no-evidence]`，不得使用藍色。

### 02. Product Context

- 標題：`Product Context - Network Observability System`
- 內容格式：系統 context map。中央放 Netstamp Controller，四周放 Web App、Probe Agent、PostgreSQL/TimescaleDB、OpenAPI/Docs、Notification Targets、Public Status Pages。
- Sticky 顏色：藍色放已實作 component；白色放 context 說明；灰色放外部系統或需要驗證的 deployment reality；粉色放不屬於當前產品 evidence 的 document stamp/certificate。
- 連線方式：Controller 連到 Probe Agent 標 `runtime auth + assignments + results`；Controller 連到 Web App 標 `session auth + project APIs`；Controller 連到 Notification Targets 標 `outbox delivery`；Controller 連到 Public Status Pages 標 `public read model`。
- 排版：中心輻射圖。中央大藍卡，周圍 6 個 component cluster，每 cluster 3-5 張 sticky。
- 尺寸/位置建議：`x=1220, y=0, w=1080, h=720`。中央卡 `260 x 160`；周邊 cluster `260 x 180`。
- Evidence tag 規則：component sticky 使用 `[E:API:runtime]`、`[E:AGENT:runtime]`、`[E:API:status]`、`[E:API:alerts]`、`[E:WEB:routes]`。外部 notification targets 若只來自 config/model，用 `[E:API:alerts]`；不要補不存在的 provider dashboard evidence。

### 03. Methodology

- 標題：`Methodology - Evidence-first Research Workflow`
- 內容格式：研究方法流程：`repo evidence review -> feature inventory -> synthetic interview scaffold -> affinity synthesis -> journey/blueprint -> prioritization -> next research plan`。
- Sticky 顏色：白色方法說明；藍色 evidence review；灰色待做驗證；黃色可產出洞察；粉色假設防線。
- 連線方式：流程箭頭由左至右；每一步向下連到對應區塊，例如 evidence review 連到 `04 Evidence Map`，synthetic interview 連到 `11 Synthetic Interview Notes`。
- 排版：水平 pipeline，上方為方法原則，下方為 deliverable checklist。
- 尺寸/位置建議：`x=2440, y=0, w=1080, h=720`。Pipeline 7 張 sticky，每張 `130 x 130`；deliverable checklist `980 x 220`。
- Evidence tag 規則：方法卡可用白色無 evidence tag，但每個 deliverable 必須標 `[V:method:<slug>]` 或連到 evidence map。Synthetic 類方法必須加 `synthetic, not user interview evidence`。

### 04. Evidence Map

- 標題：`Evidence Map - What We Know From Netstamp`
- 內容格式：依 evidence 類型分成 7 群：API Contract、Backend Domain/Application、Agent Runtime、Web UI、Data/Storage、Docs/Public Site、No Evidence / Assumptions。
- Sticky 顏色：藍色 evidence；灰色間接 evidence 或需確認；粉色 no-evidence assumption；紅色 evidence gap 高風險。
- 連線方式：每個 evidence sticky 連到 `05 Feature Inventory` 對應 feature；no-evidence cluster 連到 `09 Assumptions & Risks`。
- 排版：左 3 欄放已實作 evidence，右側窄欄放 no-evidence 和 gap。
- 尺寸/位置建議：`x=3660, y=0, w=1080, h=720`。Evidence cluster 每組 `300 x 220`；no-evidence column `220 x 620`。
- Evidence tag 規則：所有藍色 sticky 必須包含 repo path。至少覆蓋 `[E:API:checks]`、`[E:API:probes]`、`[E:API:runtime]`、`[E:API:results]`、`[E:API:alerts]`、`[E:WEB:insight]`、`[E:AGENT:runtime]`。Document stamp/certificate 只能放粉色 `[A:DOCSTAMP:no-evidence]`。

### 05. Feature Inventory

- 標題：`Feature Inventory - Implemented, Partial, Missing`
- 內容格式：功能盤點矩陣，欄位為 `Feature`、`Evidence`、`User-visible UI`、`Backend/API`、`Confidence`、`Notes`。
- Sticky 顏色：藍色已實作；灰色部分或需驗證；紫色可延伸；紅色高風險缺口。
- 連線方式：每個 feature 連到 context map 與後續 journey step；缺口連到 `20 Missing/Recommended Features`。
- 排版：以 category 分 lane：Fleet/Probes、Checks/Selectors、Runtime/Agent、Results/Insight、Alerts/Notifications、Public Status、Project/Access、Docs/API。
- 尺寸/位置建議：`x=0, y=860, w=1080, h=720`。8 條 lane，每 lane 高 `70`；右側放 `confidence legend`。
- Evidence tag 規則：Feature sticky 若標已實作，至少要同時有 API 或 backend tag；若只有 UI 或 plan 文件，標灰色而不是藍色。例：Alerts rules/notifications 可藍色；incident acknowledge/resolve mutation 若無 route evidence，標灰或紅。

### 06. User Types / Personas

- 標題：`User Types / Personas - Operational Roles`
- 內容格式：6 張 persona card：`SRE/on-call operator`、`Platform engineer`、`Network engineer`、`Backend/API owner`、`Project admin`、`Public status consumer`。每張 card 包含 goals、tools、decision pressure、Netstamp touchpoints。
- Sticky 顏色：綠色 persona needs；藍色 touchpoint evidence；粉色 persona assumption；灰色待訪談驗證。
- 連線方式：Persona card 連到 `07 JTBD`；touchpoint 連到 feature inventory；待驗證連到 `23 Open Questions`。
- 排版：2 x 3 cards，每 card 內使用 4 小格。
- 尺寸/位置建議：`x=1220, y=860, w=1080, h=720`。Persona card `320 x 200`；每 card 下方放 2-3 張 evidence chips。
- Evidence tag 規則：Persona 本身多半為假設，使用 `[A:persona:<slug>]` 或 `[V:interview:<slug>]`；與已實作 touchpoint 綁定時追加 `[E:WEB:routes]`、`[E:WEB:alerts]`、`[E:WEB:insight]`。

### 07. JTBD

- 標題：`JTBD - Jobs To Be Done`
- 內容格式：以 `When / I want to / So I can` 寫 10-14 張 JTBD sticky，並標 primary/secondary。建議包含 deploy probe、define check、understand latency/loss、triage alert、share status、rotate secret、filter by labels、inspect route topology。
- Sticky 顏色：綠色 JTBD；藍色 evidence linkage；橘色 friction note；紫色 opportunity extension。
- 連線方式：每張 JTBD 連到 persona 與 relevant feature；primary JTBD 以粗框連到 prioritization。
- 排版：3 欄：`Observe`、`Operate`、`Communicate`。
- 尺寸/位置建議：`x=2440, y=860, w=1080, h=720`。JTBD sticky `300 x 120`；每欄最多 5 張。
- Evidence tag 規則：每個 JTBD 需標 `[J:<persona>:<job>]`，並至少連回 1 個 evidence tag 或 1 個 assumption tag。不能把 document stamp/certificate 寫成 JTBD，除非標 `[A:DOCSTAMP:no-evidence]` 並放在低信心區。

### 08. Research Questions

- 標題：`Research Questions - What We Need To Learn`
- 內容格式：research question backlog。每張卡包含 question、why now、target participant、method、decision affected。
- Sticky 顏色：灰色待驗證；紅色高風險問題；綠色需求問題；粉色 assumption 問題。
- 連線方式：問題連到 assumptions、personas、JTBD 與 next research plan；紅色問題連到 roadmap gate。
- 排版：四象限：`User value`、`Operational risk`、`Adoption/onboarding`、`Assumption cleanup`。
- 尺寸/位置建議：`x=3660, y=860, w=1080, h=720`。每象限 `500 x 300`；問題 sticky `220 x 120`。
- Evidence tag 規則：每張 question 用 `[V:<method>:<slug>]`。若 question 由 gap 觸發，需追加 `[R:<risk>:<slug>]` 或 `[A:<topic>:<slug>]`。

### 09. Assumptions & Risks

- 標題：`Assumptions & Risks - Separate Known From Inferred`
- 內容格式：Assumption/Risk register。每列包含 assumption、evidence status、risk if wrong、validation method、owner。
- Sticky 顏色：粉色假設；紅色高風險；灰色驗證；藍色反證或 supporting evidence。
- 連線方式：假設連到 research questions；高風險連到 prioritization matrix；document stamp/certificate 假設連到 open questions。
- 排版：左側 assumptions，右側 risks；中間以線標出 `could invalidate`。
- 尺寸/位置建議：`x=0, y=1720, w=1080, h=720`。Assumption lane `480 x 620`；Risk lane `480 x 620`。
- Evidence tag 規則：所有粉色 sticky 必須明寫 `assumption`。Document stamp/certificate 必須使用 `[A:DOCSTAMP:no-evidence]`，body 寫 `Focused search found no product evidence for document stamp/certificate/notarization`. 高風險需用 `[R:*]` 並連到 mitigation。

### 10. Interview Guide

- 標題：`Interview Guide - Network Observability Discovery`
- 內容格式：訪談 script 模板，分為 opener、context、recent incident、probe deployment、check configuration、alert triage、status communication、closing ranking。
- Sticky 顏色：白色 script section；綠色 objective；灰色 follow-up prompt；紅色敏感問題警示。
- 連線方式：Guide 問題連到 personas 與 research questions；每段末端連到 synthetic notes placeholder。
- 排版：垂直 script flow，左側 participant type，右側 question bank。
- 尺寸/位置建議：`x=1220, y=1720, w=1080, h=720`。8 段 script，每段 `480 x 120`，右側 prompt bank `480 x 620`。
- Evidence tag 規則：每段 objective 使用 `[J:*]` 或 `[V:interview:*]`。不可將 synthetic answer 當作 evidence；訪談前保留灰色 placeholder。

### 11. Synthetic Interview Notes

- 標題：`Synthetic Interview Notes - Labeled Draft Inputs`
- 內容格式：僅作為待替換的 synthetic notes。每位 persona 1 欄，每欄含 4 種 note：goal、workflow、pain、quote-like paraphrase。不得使用直接引號形式偽裝真實訪談。
- Sticky 顏色：粉色 synthetic assumption；黃色 synthetic insight candidate；橘色 possible pain；灰色 verify with real participant。
- 連線方式：每張 synthetic note 連到 affinity map，但線條使用灰色虛線；待真實訪談補齊後才可改為實線。
- 排版：6 欄 persona lanes；每欄 4-6 張 sticky。
- 尺寸/位置建議：`x=2440, y=1720, w=1080, h=720`。Persona lane `160 x 620`。
- Evidence tag 規則：所有 sticky 必須以 `[SYN:<persona>:<slug>]` 開頭，並追加 `[V:interview:<slug>]`。禁止使用 `[E:*]`，除非 sticky 只是註明該 persona 會接觸的已實作功能，且需清楚標為 touchpoint 而非訪談結論。

### 12. Affinity Map

- 標題：`Affinity Map - Patterns From Notes And Evidence`
- 內容格式：聚類板。初始 cluster 建議：`Setup trust`、`Signal clarity`、`Selector mental model`、`Alert actionability`、`Topology explainability`、`Public communication`、`Automation/API`、`Evidence gaps`。
- Sticky 顏色：黃色洞察；橘色痛點；紫色機會；藍色 evidence anchors；粉色假設 anchors。
- 連線方式：notes 以虛線匯入 cluster；cluster 再實線連到 pain points、opportunity areas、feature mind map。
- 排版：8 個 cluster，2 x 4；每 cluster 中心放黃色 insight，周圍貼 evidence/pain/opportunity。
- 尺寸/位置建議：`x=3660, y=1720, w=1080, h=720`。Cluster `240 x 280`。
- Evidence tag 規則：每個 insight sticky 使用 `[I:<cluster>:<slug>]`，至少連到 2 個來源，其中 synthetic 來源不能超過一半；若只有 synthetic 支撐，保留粉色或灰色，不升級為黃色 confirmed insight。

### 13. Current-State Journey

- 標題：`Current-State Journey - From Probe Setup To Incident Triage`
- 內容格式：時間軸 journey：`Create project -> Create probe -> Install agent/service -> Hello/heartbeat -> Create labels -> Define check -> Preview selector -> Runtime pulls assignments -> Execute checks -> Submit results -> View insight/latest -> Alert opens -> Notify team -> Publish/status page`。
- Sticky 顏色：藍色已實作 step；橘色痛點；灰色不確定 step；紅色風險；白色 stage label。
- 連線方式：主流程用粗藍線；痛點從 step 向下垂直連到 `17 Pain Points`；alert/status 連到 `15 Service Blueprint`。
- 排版：水平 timeline，分 5 stage：Setup、Configure、Measure、Investigate、Respond。
- 尺寸/位置建議：`x=0, y=2580, w=1080, h=720`。Timeline 高 `220`；每 step `120 x 90`；下方 pain lane `620 x 300`。
- Evidence tag 規則：每個藍色 step 必須連到 API/backend/web evidence。例：install/agent service 使用 `[E:AGENT:runtime]`；result submission 使用 `[E:API:runtime]`；Insight 使用 `[E:WEB:insight]`；alert notification 使用 `[E:WEB:alerts]` 和 `[E:SRV:alerteval]`。

### 14. Future-State Journey

- 標題：`Future-State Journey - Desired Operator Experience`
- 內容格式：以使用者目標重畫理想流程：`deploy with confidence -> confirm coverage -> define checks safely -> detect change -> explain impact -> notify with context -> communicate public status -> learn and tune`。
- Sticky 顏色：綠色需求/JTBD；紫色機會；黃色情境洞察；紅色不可接受風險；藍色可沿用實作。
- 連線方式：每個 future step 連回 current step；新增或改善項目連到 opportunity areas 和 roadmap。
- 排版：雙層 journey，上層為 user action，下層為 product response。
- 尺寸/位置建議：`x=1220, y=2580, w=1080, h=720`。8 個 step，每 step `120 x 140`；下層 response `120 x 100`。
- Evidence tag 規則：Future sticky 若是改善方向用 `[O:*]`；若沿用現有功能加 `[E:*]`；若沒有 evidence 或來自 speculative domain expansion，用 `[A:*]`。

### 15. Service Blueprint

- 標題：`Service Blueprint - Frontstage, Backstage, Agent, Data`
- 內容格式：Swimlane blueprint。Lanes：User action、Web UI、Controller API、Application service、Agent runtime、Database/Timescale、Notification/Public surfaces、Observability/logging。
- Sticky 顏色：藍色已實作 touchpoint；灰色 unknown/needs trace；橘色 operational friction；紅色 failure risk。
- 連線方式：垂直線表示同一 moment 的跨層互動；runtime/result/alert chain 用加粗線標示。
- 排版：8 swimlanes x 8 moments。Moments 建議對應 current journey 的 stage。
- 尺寸/位置建議：`x=2440, y=2580, w=1080, h=720`。每 lane 高 `70`，每 moment 寬 `120`。
- Evidence tag 規則：Blueprint 中 backend/API/agent/db sticky 必須有精確路徑 tag。若只是推測 deployment 或第三方 delivery behavior，標灰或粉，不標藍。

### 16. UX Audit

- 標題：`UX Audit - Existing Product Surface`
- 內容格式：以 heuristic 分區：Navigation clarity、Information density、Setup/onboarding、Check editor complexity、Insight exploration、Alert triage、Status page publishing、Responsive behavior。
- Sticky 顏色：藍色 UI evidence；黃色 observation；橘色 usability issue；紅色 critical issue；紫色 quick win。
- 連線方式：每個 issue 連到 pain points；每個 quick win 連到 opportunity areas 或 roadmap。
- 排版：左側 evidence screenshots placeholder，右側 audit notes；下方 severity strip。
- 尺寸/位置建議：`x=3660, y=2580, w=1080, h=720`。Evidence placeholder `480 x 460`；audit note grid `520 x 460`；severity strip `980 x 100`。
- Evidence tag 規則：每個 audit note 需標來源 UI component 或 route，例如 `[E:WEB:checks]`、`[E:WEB:probes]`、`[E:WEB:alerts]`。若沒有實際 screenshot，保留白色 placeholder 並標 `[V:usability-test:screenshot-needed]`。

### 17. Pain Points

- 標題：`Pain Points - Where Operators Lose Time Or Trust`
- 內容格式：痛點列表，每張卡包含 pain、affected persona、current step、evidence/assumption、severity、candidate fix。
- Sticky 顏色：橘色痛點；紅色高風險痛點；藍色 supporting evidence；灰色需驗證。
- 連線方式：痛點連回 current-state journey step，向右連到 opportunity areas；紅色高風險連到 prioritization matrix。
- 排版：依 severity 分三列：Critical、High、Medium。每張痛點卡下方附 1 張 evidence chip。
- 尺寸/位置建議：`x=0, y=3440, w=1080, h=720`。Pain card `300 x 130`；Critical lane 高 `190`。
- Evidence tag 規則：痛點使用 `[P:<step>:<slug>]`。建議優先放：selector builder cognitive load `[E:WEB:checks]`、one-time probe secret/rotation complexity `[E:API:probes]`、alert incident action gap `[E:API:alerts]`、traceroute not evaluated for alerts `[E:SRV:alerteval]`、document stamp/certificate no-evidence confusion `[A:DOCSTAMP:no-evidence]`。

### 18. Opportunity Areas

- 標題：`Opportunity Areas - Product Improvements`
- 內容格式：機會區，每張卡包含 opportunity、target JTBD、evidence anchor、expected outcome、confidence。
- Sticky 顏色：紫色機會；綠色 connected JTBD；黃色 insight；藍色 existing leverage；紅色 if-blocking risk。
- 連線方式：每個 opportunity 從 pain point 或 insight 進入，再連到 roadmap suggestions。
- 排版：四象限：`Onboarding`、`Configuration safety`、`Investigation depth`、`Response communication`。
- 尺寸/位置建議：`x=1220, y=3440, w=1080, h=720`。每象限 `500 x 300`。
- Evidence tag 規則：Opportunity 使用 `[O:<area>:<slug>]`，必須連至少 1 個 `[P:*]` 或 `[J:*]`。若涉及新 check type、incident ack workflow、certificate/stamp product direction，標信心為灰或粉，不能藍色。

### 19. Feature Mind Map

- 標題：`Feature Mind Map - Netstamp Product Model`
- 內容格式：中央 `Netstamp` mind map，分支：Fleet、Checks、Runtime、Results/Insight、Alerts、Status Pages、Projects/Access、Docs/API、Assumptions。
- Sticky 顏色：藍色 implemented feature；紫色 opportunity branch；灰色 missing/unknown；粉色 no-evidence assumptions。
- 連線方式：樹狀連線。已實作分支使用藍線；future branch 使用紫線；assumption branch 使用粉色虛線。
- 排版：中心 radial。每個一級分支向外 2 層即可，不要把所有 endpoint 展開到不可讀。
- 尺寸/位置建議：`x=2440, y=3440, w=1080, h=720`。中心節點 `220 x 120`；一級分支 9 個，每個 `160 x 90`。
- Evidence tag 規則：每個藍色分支至少 1 個 tag；Assumptions 分支必須集中在右下角並標 `[A:*]`。Document stamp/certificate 僅能放在 Assumptions 分支。

### 20. Missing / Recommended Features

- 標題：`Missing / Recommended Features - Evidence-aware Backlog`
- 內容格式：分成 `Known gaps`、`Recommended enhancements`、`Assumption-only ideas`、`Do not claim yet` 四欄。
- Sticky 顏色：灰色 known gap；紫色 recommendation；粉色 assumption-only；紅色 do-not-claim/high-risk。
- 連線方式：known gaps 連到 evidence map；recommendations 連到 opportunity areas；assumption-only 連到 open questions。
- 排版：4 欄 kanban，每欄上方有白色 definition card。
- 尺寸/位置建議：`x=3660, y=3440, w=1080, h=720`。欄寬 `240`；卡片 `220 x 120`。
- Evidence tag 規則：建議項必須明確區分 `missing because evidence says not supported` 與 `missing because not found yet`。可放：DNS/HTTP check executors as missing if using backend guidance `[V:code-audit:dns-http-checks]`；incident acknowledge/resolve user action as gap `[V:api-audit:incident-actions]`; traceroute alerting as gap `[E:SRV:alerteval]`; document stamp/certificate as `[A:DOCSTAMP:no-evidence]` and placed under `Do not claim yet`。

### 21. Prioritization Matrix

- 標題：`Prioritization Matrix - Value x Evidence Confidence`
- 內容格式：2 x 2 matrix。X 軸 `Evidence confidence / technical clarity`，Y 軸 `User value / operational impact`。每張 candidate card 包含 opportunity、persona、effort guess、risk、next decision。
- Sticky 顏色：紫色 opportunity candidate；紅色 risk candidate；灰色 validation-needed；深色/粗框為最高優先級。
- 連線方式：候選項從 opportunity/missing features 匯入；最高優先級連到 roadmap `Now`。
- 排版：四象限：`Do now`、`Validate next`、`Defer`、`Do not pursue yet`。
- 尺寸/位置建議：`x=0, y=4300, w=1080, h=720`。Matrix `920 x 560`；legend `120 x 560`。
- Evidence tag 規則：每張 candidate 必須同時有 `[O:*]` 或 `[P:*]`，以及 evidence/assumption tag。最高優先級卡使用深色底或粗框，且不得建立在只有 `[A:*]` 的基礎上。

### 22. Roadmap Suggestions

- 標題：`Roadmap Suggestions - Now / Next / Later`
- 內容格式：Roadmap lanes：`Now`、`Next`、`Later`、`Research-only`。每張卡包含 outcome、scope、evidence confidence、validation gate、owner surface。
- Sticky 顏色：紫色 roadmap item；藍色 existing leverage；綠色 outcome/JTBD；灰色 research gate；粉色 assumption-only。
- 連線方式：從 prioritization matrix 接入；每個 roadmap item 連到 next research plan 或 implementation area。
- 排版：4 條垂直 lane；每 lane 按上到下表示優先順序。
- 尺寸/位置建議：`x=1220, y=4300, w=1080, h=720`。Lane 寬 `250`；卡片 `230 x 130`。
- Evidence tag 規則：`Now` lane 只能放高 confidence evidence-backed items。粉色 assumption-only item 必須放 `Research-only`，包含 document stamp/certificate 方向。

### 23. Open Questions

- 標題：`Open Questions - Decisions Before Design Or Build`
- 內容格式：Open question board。每張卡包含 question、decision needed、blocked artifact、recommended owner、deadline/phase。
- Sticky 顏色：灰色問題；紅色 blocking/high-risk；粉色 assumption cleanup；綠色 user validation need。
- 連線方式：每個 question 連到 assumptions/research questions/roadmap item；blocking 問題以紅線連到 roadmap gate。
- 排版：依 `Product`、`UX`、`Technical`、`Go-to-market/positioning` 分 4 欄。
- 尺寸/位置建議：`x=2440, y=4300, w=1080, h=720`。欄寬 `250`；問題卡 `230 x 120`。
- Evidence tag 規則：問題使用 `[V:*]`。必放問題：`是否要保留或排除 document stamp/certificate 敘事？` 標 `[A:DOCSTAMP:no-evidence]`；`alerts 是否需要 acknowledge/resolve 使用者動作？` 標 `[V:api-audit:incident-actions]`；`traceroute 是否應納入 alert metrics？` 標 `[E:SRV:alerteval]`。

### 24. Next Research Plan

- 標題：`Next Research Plan - Validation Sprint`
- 內容格式：研究計畫表，分 `Week 1 evidence cleanup`、`Week 2 interviews/usability`、`Week 3 synthesis`、`Decision review`。每列包含 activity、participants/data、artifact、success criteria。
- Sticky 顏色：綠色 research objective；灰色 task；黃色 expected insight；紅色 risk to plan；紫色 design output。
- 連線方式：從 open questions 接入；完成後回連 research overview，形成下一輪 board 更新 loop。
- 排版：4-phase timeline；底部放 checklist：招募、script、prototype/screenshot、recording consent、analysis owner。
- 尺寸/位置建議：`x=3660, y=4300, w=1080, h=720`。Phase card `240 x 520`；bottom checklist `980 x 120`。
- Evidence tag 規則：每個 task 使用 `[V:<method>:<slug>]`。若 task 是補 repo evidence，使用 `[V:code-audit:<slug>]`；若 task 是驗證 persona/JTBD，使用 `[V:interview:<slug>]`；若 task 是確認 document stamp/certificate，不得使用 evidence tag，只能使用 `[A:DOCSTAMP:no-evidence]`。

## 建議最高優先級候選卡

以下卡片建議在 `21 Prioritization Matrix` 使用深色底或粗框，但仍需由後續研究確認：

- `Clarify probe setup confidence`：讓使用者建立 probe、拿到 one-time secret、安裝 service、看到 hello/heartbeat 成功的過程更可驗證。Tags: `[O:onboarding:probe-setup-confidence] [E:API:probes] [E:API:runtime] [E:WEB:probes]`
- `Make selector impact understandable`：讓 check selector preview、matched probes、labels 與 assignments 的關係更容易理解。Tags: `[O:config:selector-clarity] [E:API:checks] [E:WEB:checks]`
- `Alert triage with action path`：讓 incident 不只呈現 firing/value，也連到 affected probe/check、chart window、notification history、ack/resolve 問題。Tags: `[O:response:alert-triage] [E:API:alerts] [E:WEB:alerts] [E:SRV:alerteval]`
- `Topology explainability`：讓 traceroute runs/topology 的變化能被 operator 解釋與分享。Tags: `[O:investigation:topology-explainability] [E:API:results] [E:WEB:insight]`
- `Protect product scope from stamp/certificate confusion`：若利害關係人提到 document stamp/certificate，先以 assumption 研究，不要放入已實作產品敘事。Tags: `[R:positioning:scope-confusion] [A:DOCSTAMP:no-evidence]`

## Board 完成檢查表

- 24 個 frame 均存在，且標題與座標符合本文件。
- 每張藍色 sticky 都有 repo path evidence tag。
- 每張粉色 sticky 都明確寫 assumption。
- Document stamp/certificate 只出現在 Assumptions、Missing/Recommended Features、Open Questions、Research Plan，且均標 `[A:DOCSTAMP:no-evidence]`。
- Synthetic interview notes 全部標 `[SYN:*]`，且連線為灰色虛線。
- Pain、Opportunity、Prioritization、Roadmap 之間有可追溯連線。
- 深色/粗框最高優先級卡不超過每區 3 張，且不能只依賴 assumption。
