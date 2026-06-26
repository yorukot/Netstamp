# Subagent 4：Heuristic Evaluation and UX Audit

日期：2026-06-24  
範圍：Netstamp 目前 web UI、docs、`ux-design.md`、API error handling、auth/permission flow。  
評估框架：Nielsen heuristics、accessibility basics、IA clarity、onboarding clarity、error recovery、trust/security/privacy/legal communication、mobile responsiveness、content design，並參考 Vercel Web Interface Guidelines。

## 前提與限制

- 本次沒有在 codebase 中找到 `design.md`，只有 `ux-design.md`。`ux-design.md` 自述 `design.md` 應是視覺系統 source of truth，因此本報告以 `ux-design.md`、現有 UI 與 docs 為評估基準。
- 沒有在目前讀到的範圍內看到「document stamp / certificate verification」產品流程；依任務要求，trust/authenticity/security 的評估改以 network monitoring trust 為主：使用者是否理解 probe、check、result、alert、incident、status page 的意義、可信度、失敗狀態、下一步、歷史證據，以及是否可能誤解為 SLA、法律或安全保證。
- 這是 heuristic audit，不是完整 usability test。Severity 以 MVP 啟用、誤操作風險、信任風險與修復成本綜合判斷。

## 已確認的正向基礎

- API 已有結構化 `ProblemDetails` 與 field-level validation 模型：`api/models/common.tsp:63-89`、`server/internal/controller/application/validation/error.go:5-90`。
- HTTP 層會回傳 `X-Request-ID`，有助於客服與除錯追蹤：`server/internal/controller/transport/http/httpx/httpx.go:107-133`。
- Session cookie 使用 `HttpOnly`、`SameSite=Lax`，並依 runtime 設定 `Secure`：`server/internal/controller/transport/http/handler/auth/cookie.go:12-22`。
- 基本 security headers 已存在：`server/internal/controller/transport/http/middleware/security_headers.go:5-12`。
- 後端 project/member permission 有明確 role policy 與 last-owner 保護：`server/internal/domain/project/permission.go:17-55`、`server/internal/controller/application/project/service.go:174-421`。
- UI 元件庫已有一些可及性基礎，例如 `DataTable` row keyboard activation、`IconButton` required aria-label、`Field` label/description/error 關聯、toast live region、confirm dialog：`packages/ui/src/components/DataTable/DataTable.tsx:155-229`、`packages/ui/src/components/IconButton/IconButton.tsx:5-8`、`packages/ui/src/components/Field/Field.tsx:354-405`、`packages/ui/src/components/Toast/ToastProvider.tsx:16-18`、`packages/ui/src/components/Confirm/ConfirmProvider.tsx:181-240`。

## Severity 摘要

| Severity | Count | 主要風險                                                                                                                                                         |
| -------- | ----: | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Critical |     0 | 未發現足以判定為 Critical 的單點問題；但多個 High 會共同削弱 MVP trust loop。                                                                                    |
| High     |     8 | 初次啟用中斷、Dashboard 無法 triage、incident/alert/member/probe 操作容易誤解或誤設、DNS 文案過度承諾、public status trust 說明不足。                            |
| Medium   |    10 | error recovery、loading visibility、permission clarity、selector clarity、mobile density、mutating API explorer、audit/time context 等會造成操作成本或信任折損。 |
| Low      |     2 | motion/content/privacy polish。                                                                                                                                  |

## Issues

### 1. 初次啟用流程在 heartbeat 後中斷，沒有引導到第一個 check/result

- Issue：使用者建立 project 與 probe 後，流程主要停在 heartbeat 成功與 probe fleet，沒有把使用者帶到「建立第一個 check -> 看第一筆 result/insight」的完整 activation loop。
- Severity: High
- Affected user：首次使用者、self-host operator、POC 評估者。
- Affected flow：Register -> Onboarding -> New Probe -> Heartbeat -> First Check -> First Result。
- Evidence：`ux-design.md:287-322` 明確定義 project/probe/check/result activation loop，也列出 heartbeat 後中斷與缺少 first check CTA；`web/src/features/auth/components/OnboardingPage.tsx:290-296` 成功文案只導向 probe fleet/create probe；`web/src/features/probes/components/NewProbeDrawer.tsx:386-404` heartbeat 成功後主要是 Finish/close drawer。
- Why it matters：Netstamp 的可信價值不是「probe 上線」本身，而是從使用者控制的網路位置產生可解釋的測量證據。停在 heartbeat 會讓新使用者不知道下一步、也無法快速感受到產品價值。
- Recommended fix：新增 first-run checklist，狀態包含 Project created、Probe heartbeat、Create first check、View first result、Configure first alert。Heartbeat 成功後提供主要 CTA「Create first check」，並預選剛建立的 probe；建立第一個 check 後導向 insight/result。
- Effort estimate：M
- Confidence：High

### 2. Dashboard 不是 operational dashboard，無法支援回訪 triage

- Issue：Dashboard 目前只呈現 probes online、active checks 與 map，缺少 open incidents、degraded checks、recent failures、stale probes、setup progress 與 next action。
- Severity: High
- Affected user：值班 operator、團隊 owner/admin、回訪使用者。
- Affected flow：Daily monitoring、incident triage、first-run setup progress。
- Evidence：`web/src/features/dashboard/components/DashboardPage.tsx:21-59` 僅查 probes/checks 並渲染兩個 metric card 與 `NetworkMap`；`ux-design.md:170-179`、`ux-design.md:330-346`、`ux-design.md:465-470` 都指出 Dashboard 太薄，應成為 setup progress 與 operational summary 入口。
- Why it matters：使用者回到監控產品時，第一個問題是「現在是否有事、哪裡壞、下一步做什麼」。目前 Dashboard 會讓使用者自行切換 Alerts、Checks、Insights 才能拼出狀態，降低信任與效率。
- Recommended fix：在 Dashboard 增加 open incidents、stale probes、recent degraded assignments、latest check failures、last evaluation time、setup checklist/next action。每個摘要應連到對應 incident/check/probe/insight。
- Effort estimate：M-L
- Confidence：High

### 3. Incident detail 缺少分析入口與 lifecycle 說明

- Issue：Incident detail drawer 顯示 what happened、timeline、notifications，但沒有 deep link 到 Insight、Check、Probe、Rule，也沒有 ack/resolve action 或「system-managed lifecycle」說明。
- Severity: High
- Affected user：operator、on-call、team admin。
- Affected flow：Alert -> Incident detail -> Diagnose -> Recover。
- Evidence：`web/src/features/alerts/components/AlertsPage.tsx:668-707` 只顯示 incident 基本資訊與通知；`ux-design.md:256`、`ux-design.md:339-340`、`ux-design.md:435`、`ux-design.md:471` 都指出缺少 Insight/Probe/Check links；UI filter 有 `acknowledged` 狀態 `web/src/features/alerts/components/alertPageModel.ts:84-89`，但 API 只有 list/get incidents `api/services/alerts.tsp:63-79`、handler 也只有 incident list/get routes `server/internal/controller/transport/http/handler/alert/handler.go:25-39`。
- Why it matters：事件細節頁應讓使用者快速確認原因、影響範圍與下一步。現在使用者看到 incident 後仍需手動找 check/probe/insight；acknowledged filter 又暗示可以手動處理 lifecycle，容易造成不一致期待。
- Recommended fix：新增「Open insight」、「Open check」、「Open probe」、「Open rule」連結；若 lifecycle 是系統自動管理，明確標示「Incidents are opened/resolved by rule evaluation」並移除或隱藏無法操作的 acknowledged filter。若產品需要 ack，補 API 與 audit trail。
- Effort estimate：M
- Confidence：High

### 4. Alert rule scope 使用 raw UUID，容易設定錯誤且難以建立信任

- Issue：Alert rule scope 讓使用者輸入 Probe ID / Check ID UUID，而不是從可理解的 probe/check 名稱選取。
- Severity: High
- Affected user：admin、editor、設定告警的人。
- Affected flow：Create/edit alert rule、scope a rule to a probe/check。
- Evidence：`web/src/features/alerts/components/AlertsPage.tsx:761-768` 使用 `TextField label="Probe ID"`、`Check ID`；API scope 模型是 raw `probeId` / `checkId` `api/models/alert.tsp:59-64`；`ux-design.md:472` 已列為 P0。
- Why it matters：告警是 trust-sensitive flow。raw ID 容易貼錯、難以檢查、也無法讓使用者直覺知道「這條規則會監控哪些東西」。錯誤 scope 會導致漏報或誤報。
- Recommended fix：改成 searchable selects，顯示 probe/check name、target、type、status 與 ID suffix；建立前顯示 rule preview：「applies to 3 TCP checks on 2 probes」。保留 advanced/raw ID 只給 power users。
- Effort estimate：M
- Confidence：High

### 5. Member removal 沒有 confirmation，與破壞性操作風險不相稱

- Issue：Members table 的移除成員 action 會直接呼叫 delete mutation，沒有確認對話框，也沒有顯示被移除者與角色。
- Severity: High
- Affected user：project owner/admin、被誤移除的成員。
- Affected flow：Project settings -> Members -> Remove member。
- Evidence：`web/src/features/members/components/MembersPage.tsx:264-267` 直接 `removeMemberMutation.mutate(row.userId)`；相較之下 project delete/leave 已有 confirm pattern `web/src/features/project/components/ProjectPage.tsx:65-105`；`ux-design.md:473` 已列為 P0。
- Why it matters：移除成員會立即影響存取權限與營運責任歸屬。雖然後端有 last owner guard，但 UI 缺少確認仍容易造成誤操作與信任受損。
- Recommended fix：套用既有 confirm dialog，文案包含 member name/email、role、影響範圍，並要求確認「Remove member」。成功 toast 應包含可追蹤資訊；若有 audit log，寫入 actor/action/target。
- Effort estimate：S
- Confidence：High

### 6. Probe creation 被 location/geocoding 擋住，且第三方查詢與隱私溝通不足

- Issue：建立 probe 需要選定 coordinates 才能產生 install command；UI 預設使用 Nominatim 查詢位置，但沒有清楚說明可否跳過、為何需要、資料會送到第三方、或如何稍後補上。
- Severity: High
- Affected user：首次使用者、隱私敏感組織、只想快速測試的 operator。
- Affected flow：New Probe wizard -> Name/location -> Install command。
- Evidence：`web/src/features/probes/components/NewProbeDrawer.tsx:84-95` `canCreate` 需要 `selectedCoordinates`；`web/src/features/probes/components/NewProbeDrawer.tsx:197-228` 依賴 location search/manual coordinates；Nominatim endpoint 在 `web/src/features/probes/data/probeLocation.ts:14`、錯誤訊息在 `web/src/features/probes/data/probeLocation.ts:53-72`；`ux-design.md:474`、`ux-design.md:510` 指出 probe creation 不應被 location 擋住。
- Why it matters：位置有助於 map 與營運脈絡，但不應阻擋核心 trust loop。第三方 geocoding 也會把使用者輸入的地點送出產品邊界，對 privacy/security 評估敏感。
- Recommended fix：允許「Skip location for now」並使用未定位狀態；把 location 作為可稍後補的 metadata。使用地點搜尋前加入簡短 privacy notice：「Searches are sent to OpenStreetMap Nominatim」；手動座標模式提供 validation 與範例。
- Effort estimate：M
- Confidence：High

### 7. Landing/docs 宣稱 DNS 能力，但產品 UI/API 目前核心流程未完整支援 DNS check

- Issue：公開文案與 docs 多處提到 DNS checks/data，但目前 web UI 的 check type 主要是 Ping/TCP/Traceroute，`ux-design.md` 也提醒 DNS 不應在 MVP 當成可執行能力宣傳。
- Severity: High
- Affected user：評估產品的 buyer/operator、依 docs 規劃部署的人。
- Affected flow：Landing page -> Docs -> Product trial -> Create check。
- Evidence：Landing section 寫「Latency, loss, DNS, and routes」與「measure latency, packet loss, DNS, and routes」：`docs/src/components/landing/LandingPage.tsx:16-28`、`docs/src/components/landing/LandingPage.tsx:72-73`；probe operations docs 提到 historical DNS data 與 DNS checks：`docs/src/content/docs/guides/probe-operations.mdx:10-22`；UI check type options 是 ping/tcp/traceroute：`web/src/features/checks/components/ChecksPage.tsx:596-600`；`ux-design.md:48-50`、`ux-design.md:500`、`ux-design.md:656` 明確標示 DNS claim 是 trust risk。
- Why it matters：監控產品的 trust 建立在「說到的能力確實可用」。文案過度承諾會讓使用者誤以為可以監控 DNS，進而影響採購、合規或事件分析期待。
- Recommended fix：在 MVP 前移除主要文案中的 DNS claim，或標示為 roadmap/coming soon。若要保留，需完成 controller、probe runtime、API、UI、docs 的 end-to-end DNS check 並提供 limitations。
- Effort estimate：S for copy removal / L for full implementation
- Confidence：High

### 8. Public status page 沒有充分說明測量來源、可信度、stale/unknown 與非 SLA 限制

- Issue：Public status page 顯示 summary、incidents、elements 與 generated timestamp，但缺少 measurement provenance、aggregation window、stale threshold、sample count、probe coverage、status semantics、非 SLA/非安全保證說明。
- Severity: High
- Affected user：外部 stakeholder、客戶、legal/compliance reviewer、incident subscriber。
- Affected flow：Public status page interpretation、incident communication、customer trust review。
- Evidence：`web/src/features/status-pages/components/PublicStatusPage.tsx:47-69` 顯示 header/generated/incident/elements；`web/src/features/status-pages/components/PublicStatusPage.tsx:145-183` 顯示 element type、target、latest time 與 assignment status；status 映射在 `web/src/features/status-pages/components/statusPageAdapters.ts:13-23`。目前沒有看到 how-to-interpret、measurement source、SLA disclaimer 或 proof/audit explanation。
- Why it matters：public status 是 high-trust surface。外部使用者可能把「operational/degraded/down」理解成 SLA、法律聲明或完整安全保證；也可能不知道 unknown/stale 是沒有資料、probe offline、或目標真的恢復。
- Recommended fix：在 public status page 加「How this status is measured」區塊：probe 數量/位置、check type、最近樣本時間、aggregation window、stale definition、incident lifecycle、資料延遲。增加短 disclaimer：「This page reports Netstamp probe observations and is not an SLA or security certification unless separately stated.」
- Effort estimate：M
- Confidence：High

### 9. Error recovery 沒有充分利用 API field errors 與 request ID

- Issue：後端已有 structured problem 與 field errors，但前端多數情境只顯示 detail/title 或合併後的錯誤訊息，缺少欄位名稱、可操作修復步驟與 request ID；部分 handler 也回 generic validation error。
- Severity: Medium
- Affected user：所有表單使用者、支援/客服、API 使用者。
- Affected flow：Create/edit probe/check/alert/status page/member、API failure recovery。
- Evidence：`server/internal/controller/transport/http/httpx/httpx.go:107-133` 支援 `ProblemDetails` 與 `X-Request-ID`；`web/src/shared/api/client.ts:52-67` 主要從 `problem.detail || problem.title` 建立 message；`web/src/shared/api/requestErrorMessage.ts:7-22` 只 join field message，沒有清楚保留 field/location；alert handler 對 validation 回 `invalid alert input`：`server/internal/controller/transport/http/handler/alert/context.go:37-55`，相較 check handler 有 field mapping：`server/internal/controller/transport/http/handler/check/context.go:27-75`。
- Why it matters：監控設定表單通常有數值、threshold、selector、scope。錯誤訊息若不能指向具體欄位，使用者會反覆試錯，也較難把 request ID 提供給支援。
- Recommended fix：建立統一 form error adapter：field path -> label -> inline error；toast 保留摘要，detail 面板顯示 request ID。讓 alert/probe/status handlers 盡量回 field-level `errors`。對 403/401/demo read-only 提供下一步：switch project、login again、contact owner、duplicate as draft。
- Effort estimate：M
- Confidence：High

### 10. Loading 與 guard 狀態使用 blank screen，違反系統狀態可見性

- Issue：多個 route guard 與 route suspense 在 loading 或 project mismatch 時 return `null`，lazy route fallback 也是 `null`。
- Severity: Medium
- Affected user：慢網路使用者、第一次載入者、登入/session refresh 使用者。
- Affected flow：App load、project switch、protected route navigation。
- Evidence：`web/src/routes/guards.tsx:23-25`、`web/src/routes/guards.tsx:43-45`、`web/src/routes/guards.tsx:65-67`、`web/src/routes/guards.tsx:79-81`、`web/src/routes/guards.tsx:107-109`、`web/src/routes/guards.tsx:122-123` 多處 return `null`；`web/src/routes/routeSuspense.tsx:1-5` fallback 是 `null`。
- Why it matters：Blank screen 會讓使用者以為 app 壞掉、session 過期或權限被移除。這對監控工具尤其敏感，因為使用者通常是在有問題時打開產品。
- Recommended fix：提供 app-level loading shell、route skeleton、project switching indicator、auth refresh message。若超過短時間，顯示可重試與 request/session context。
- Effort estimate：S-M
- Confidence：High

### 11. Auth/session 與 permission flow 的人類可理解說明不足

- Issue：後端 permission policy 清楚，但 UI 在角色、disabled action、401/403 情境中缺少「你目前是什麼角色、為什麼不能做、誰可以處理」的說明。
- Severity: Medium
- Affected user：viewer/editor/admin、被邀請者、session expired 使用者。
- Affected flow：Members、project settings、protected route, create/edit/delete actions。
- Evidence：role options 只顯示角色名稱 `web/src/features/members/components/RoleSelect.tsx:13-39`；members page 使用 local permission predicates `web/src/features/members/components/MembersPage.tsx:70-83`，但邀請/角色選擇區缺少 role capability descriptions `web/src/features/members/components/MembersPage.tsx:293-310`；auth middleware 401 detail 是 technical message `server/internal/controller/transport/http/middleware/auth.go:11-32`。
- Why it matters：權限與身份是 trust/security communication 的核心。使用者需要知道目前權限邊界，不應透過失敗或 disabled button 猜測。
- Recommended fix：在 Members/Settings 加 role description matrix 或 inline helper；disabled destructive actions 加 tooltip/inline reason；401/403 前端文案轉成「Your session expired」「Only owners/admins can invite members」等可操作訊息，並附 login/contact owner CTA。
- Effort estimate：M
- Confidence：Medium

### 12. Check selector builder 的 mental model 不夠清楚，Advanced JSON 過早暴露

- Issue：Check selector 有 match mode、labels、advanced JSON、preview count，但使用者不容易理解規則如何匹配 probe、沒有 match reason、沒有 label 空狀態下一步；Advanced JSON 以一般選項出現，增加誤用成本。
- Severity: Medium
- Affected user：editor/admin、設定 checks 的 operator。
- Affected flow：Create/edit check -> Select probes。
- Evidence：`web/src/features/checks/components/ChecksPage.tsx:629-706` 顯示 selector builder 與 preview；`web/src/features/checks/components/selectorState.ts:27-32` 把 Advanced JSON 作為正常模式；`web/src/features/checks/components/selectorState.ts:59-68` 在沒有 labels 時仍建立空 key/value rule；`web/src/features/probes/components/ProbeDetail.tsx:385-399` 無 labels 時只說「No project labels available」。
- Why it matters：selector 是 check coverage 的核心。若使用者誤以為 check 覆蓋所有 probe，實際卻只匹配少數或零個，會造成監控盲點。
- Recommended fix：預設提供「All active probes」「Specific probes」「By labels」三種人類語意模式；preview 顯示 matched/unmatched 與原因；無 labels 時提供 create label/link to probe labels；Advanced JSON 收到 collapsible expert mode 並加 schema validation。
- Effort estimate：M-L
- Confidence：High

### 13. Drawer/editor 關閉時可能丟失未儲存修改

- Issue：Checks editor 等 drawer 透過 close/reset 關閉，未看到 dirty-state confirmation。使用者點 backdrop、escape 或 close 可能失去正在編輯的設定。
- Severity: Medium
- Affected user：admin/editor、正在建立複雜 check/alert/status page 的使用者。
- Affected flow：Create/edit check、alert rule、status page element、probe detail edits。
- Evidence：`web/src/features/checks/components/ChecksPage.tsx:96-108` 使用 local editor state；`web/src/features/checks/components/ChecksPage.tsx:168-172` `closeEditor` 重設 draft 並 navigate；`packages/ui/src/components/Drawer/Drawer.tsx:39-55` 提供一般 dialog close/back 行為。
- Why it matters：監控設定常含多步驟與精確數值。未提示就丟失會造成挫折，也會讓使用者不敢探索設定。
- Recommended fix：在 shared editor/drawer 層支援 `isDirty` 與 confirm before close；儲存成功後才 reset；對 create flow 可保留 draft in memory/sessionStorage。
- Effort estimate：M
- Confidence：Medium

### 14. Project Settings / Account Settings / sidebar Settings 的 IA 命名容易混淆

- Issue：sidebar 的 Settings 指向 project settings；user menu 也有 Settings 指向 account page；頁面標題分別是 Settings、Account，但使用者不一定知道目前在改 project 還是個人帳號。
- Severity: Medium
- Affected user：所有登入使用者，特別是多 project/team 使用者。
- Affected flow：Navigate settings、manage project、manage account、invite acceptance。
- Evidence：`web/src/routes/sidebarItems.ts:10-19` sidebar `Settings` route 是 `projectSettings`；`web/src/features/project/components/ProjectPage.tsx:108-157` 標題是 `Settings` 但內容是 project info/dangerous project actions；`web/src/layouts/components/UserMenu.tsx:133-143` user menu `Settings` 連到 account settings；account page title 是 `Account`：`web/src/features/settings/components/SettingsPage.tsx:167-170`。
- Why it matters：設定資訊架構不清楚會導致使用者在錯誤範圍做修改，例如以為在改個人設定卻改到 project，或找不到邀請/帳號設定。
- Recommended fix：sidebar label 改為 `Project settings`；user menu label 改為 `Account settings`；頁面 header 明確顯示 scope，例如「Project settings: <project name>」。Breadcrumb 或 subnav 分開 Project / Account / Members。
- Effort estimate：S
- Confidence：High

### 15. Operational tables 在 mobile 上過度依賴寬表格與水平捲動

- Issue：多個核心頁面使用 52-72rem 以上的 table min-width；手機上雖可水平捲動，但 triage/context/action 會被切開，操作密度高。
- Severity: Medium
- Affected user：手機上查看 incident/status 的 operator、on-call、外部 status page 管理者。
- Affected flow：Probes list、Checks table、Alerts incidents/rules/notifications、Members。
- Evidence：Probes table `web/src/features/probes/components/ProbeList.tsx:121-129` minWidth 約 62rem；Checks table `web/src/features/checks/components/ChecksTable.tsx:121-157` minWidth 約 72rem；Alerts tables `web/src/features/alerts/components/AlertsPage.tsx:491-595` 包含多個寬表格；Members invite/list 表格也使用寬 table layout `web/src/features/members/components/MembersPage.tsx:176-320`。
- Why it matters：監控產品常在手機上被用來快速確認「是否有事」。寬表格使重要欄位、status、action 分散，降低可掃描性與反應速度。
- Recommended fix：為小螢幕提供 card/list summary：status + name + target + latest + primary action；table 保留 desktop。Filters 改成 collapsible sheet，row action 使用 overflow menu，incident/probe/check detail 用 full-screen drawer。
- Effort estimate：M-L
- Confidence：Medium

### 16. Docs OpenAPI Explorer 可用登入 cookie 發送 mutating request，但缺少足夠警示與確認

- Issue：Docs 的 OpenAPI Explorer 支援直接送 authenticated request，且 `credentials: include`，對 POST/PATCH/DELETE 這類 mutating endpoint 缺少明顯危險提示或二次確認。
- Severity: Medium
- Affected user：開發者、admin、正在瀏覽 docs 的已登入使用者。
- Affected flow：Docs -> API Explorer -> Test Request。
- Evidence：`docs/src/components/openapi/OpenAPIExplorer.tsx:638-682` 使用 `fetch` 並 `credentials: include`；request console 有 `Test Request`：`docs/src/components/openapi/OpenAPIExplorer.tsx:898-952`；docs 只提醒 authenticated routes 需要 valid session `docs/src/content/docs/guides/api-explorer.mdx:26-29`。
- Why it matters：文件中的「測試」容易被理解成 sandbox。若實際使用 production session cookie 發出刪除或修改請求，會造成意外資料變更。
- Recommended fix：對 non-GET 顯示 persistent warning；DELETE/PATCH/POST 需 confirm，顯示 target environment 與 account/project；提供 read-only demo mode 或 require explicit base URL/session selection。預設不帶 credentials，除非使用者開啟 authenticated mode。
- Effort estimate：M
- Confidence：High

### 17. Probe secret rotation 是高風險操作，但操作前說明與確認不足

- Issue：Probe detail 提供 Rotate secret action，按下後立即 mutation；成功後才顯示 rotated secret 與 service command，操作前沒有清楚說明舊 secret 失效、需要重裝/更新 service、可能造成 probe offline。
- Severity: Medium
- Affected user：operator、admin/editor。
- Affected flow：Probe detail -> Rotate secret -> Update installed probe。
- Evidence：`web/src/features/probes/components/ProbeDetail.tsx:472-502` action row 有 Rotate secret；`web/src/features/probes/components/ProbeDetail.tsx:530-554` 成功後顯示 secret/command 並說明 rewrite systemd service environment。
- Why it matters：credential rotation 是 security-sensitive。使用者需要在操作前理解後果，否則可能讓 production probe 掉線或遺失新 secret。
- Recommended fix：Rotate 前加 confirm dialog，內容包含 probe name、舊 secret 將失效、必須立即更新 service、secret 只顯示一次。成功後加「copy command」「mark as updated」「view heartbeat」步驟與 audit event。
- Effort estimate：S-M
- Confidence：High

### 18. 時間、歷史與 proof explanation 不足，削弱 audit trail 感

- Issue：多處時間顯示依 browser locale `toLocaleString`，常缺 timezone/year/absolute+relative pair；incident/status/result 的 proof context 不夠明確。
- Severity: Medium
- Affected user：operator、incident reviewer、compliance/security reviewer、外部 status reader。
- Affected flow：Incident review、probe heartbeat review、public status interpretation、postmortem。
- Evidence：Probe list format helper 使用 `new Date(timestamp).toLocaleString()` 與 relative heartbeat：`web/src/features/probes/components/ProbeList.tsx:73-81`；status adapter `formatDateTime` 只顯示 month/day/hour/min：`web/src/features/status-pages/components/statusPageAdapters.ts:87-103`；public status page 顯示 generated/latest time，但沒有 timezone/audit/proof explanation：`web/src/features/status-pages/components/PublicStatusPage.tsx:47-183`。
- Why it matters：監控證據的可信度需要時間脈絡。缺少 timezone/year 會讓跨地區團隊或事後分析無法精準判讀，尤其是 public incident communication。
- Recommended fix：統一顯示 absolute + relative time，例如 `2026-06-24 14:03 UTC+8 · 5m ago`；status page 顯示 generated at、latest sample at、evaluation window。Incident timeline 加 evaluation/run ID、rule ID、probe/check links，方便 audit trail。
- Effort estimate：M
- Confidence：High

### 19. Onboarding terminal animation 與 auto-focus 可能造成可及性與風格不一致

- Issue：Onboarding 使用 typewriter terminal 動畫與自動 focus 流程，未看到 reduced-motion handling；視覺風格也與主 app operational UI 有落差。
- Severity: Low
- Affected user：motion-sensitive 使用者、screen reader/keyboard 使用者、首次使用者。
- Affected flow：Register -> Onboarding。
- Evidence：`web/src/features/auth/components/OnboardingPage.tsx:76-113` 有 script/typewriter timeouts；`web/src/features/auth/components/OnboardingPage.tsx:115-131` 自動 focus input；`ux-design.md:164` 指出 onboarding 風格與主 app 不一致。
- Why it matters：首次啟用應該降低認知負擔。過強的 terminal metaphor 可能有趣，但若阻礙快速完成 project/probe/check/result，會與 MVP 目標衝突。
- Recommended fix：尊重 `prefers-reduced-motion`；提供直接表單模式；移除非必要 delay；讓 copy 更接近主 app 的 operational language。
- Effort estimate：S-M
- Confidence：Medium

### 20. Gravatar avatar 使用外部服務，privacy 說明與控制不足

- Issue：Account settings 說明 avatar 來自 Gravatar，但沒有清楚說明 email hash/external request/privacy tradeoff，也沒有 opt-out 或本地預設控制。
- Severity: Low
- Affected user：隱私敏感使用者、企業/合規環境。
- Affected flow：Account settings、profile display。
- Evidence：`web/src/features/settings/components/SettingsPage.tsx:198-207` 顯示 Gravatar preview 與簡短說明；`web/src/features/settings/data/gravatar.ts:16-30` 以 email SHA-256 建立 Gravatar URL。
- Why it matters：即使 email 是 hash，載入外部 avatar 仍會向第三方產生請求。企業使用者可能需要避免外部 profile lookup。
- Recommended fix：加入 privacy note：「Avatar is loaded from Gravatar using a hashed email」；提供 disable external avatar / use initials option；self-host mode 可設定禁用 Gravatar。
- Effort estimate：S
- Confidence：High

## Cross-cutting 建議

1. 把 MVP trust loop 做成產品主軸：Project -> Probe heartbeat -> Check -> Result/Insight -> Alert -> Incident -> Public status，每一步都要有狀態、下一步與 proof context。
2. 把 network monitoring trust 說清楚：probe 是誰控制、check 在測什麼、result 代表哪個時間窗、alert rule 怎麼觸發、incident 何時開啟/解除、public status 不是 SLA/法律/安全認證。
3. 優先修正 `ux-design.md` 已列 P0 且本次 code review 也確認存在的項目：first-run loop、Dashboard summary、incident deep links、alert scope selects、member removal confirm、probe location optional、DNS claim 移除或補齊。
4. 建立一套共用 recovery pattern：loading skeleton、inline field error、request ID、retry、permission reason、destructive confirmation、dirty close guard。
5. Mobile 不只做 responsive table；對 operator/on-call flow 應提供可掃描的 mobile-first summary。
