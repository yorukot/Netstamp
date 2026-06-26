# UX Flow, Journey Map 與 Service Blueprint

> 範圍：此文件只描述 codebase 目前可驗證的 current-state UX flow、web routes、API/TypeSpec contract、Go backend routes 與 permission/service behavior。Netstamp 在程式碼與文件中呈現為「由使用者部署 probes 的 network observability 產品」，而不是文件蓋章、憑證簽發或驗證產品。
>
> Evidence 格式：`path:line-line`。若某功能只有 UI 或 contract evidence、沒有完整 backend 行為 evidence，會明確標為限制或假設。

## 1. Current-State Flow Map

### 1.1 主要入口與 route guard

- 使用者進入 `/` 後，React router 直接導向 project dashboard；未登入使用者會被 `ProtectedAppShell` 導到 `/login`。
- 已登入但沒有可用 project 的使用者會被導向 `/onboarding`。
- app 的主要 authenticated routes 都是 project-scoped：`/projects/:projectRef/dashboard`、`/probes`、`/probes/new`、`/checks`、`/alerts`、`/status-pages`、`/labels`、`/insight`、`/members`、`/settings`。
- legacy routes `/dashboard`、`/probes`、`/checks` 等會 redirect 到目前選取的 project route。
- 公開 status page 是 `/status/:slug`，不在 authenticated shell 裡。

Evidence:

- `web/src/routes/router.tsx:12-79` 定義 `/`、auth、onboarding、public status、project routes、legacy redirect 與 catch-all redirect。
- `web/src/routes/guards.tsx:19-33` 定義登入/註冊頁對已登入者的 redirect 與 registration-disabled behavior。
- `web/src/routes/guards.tsx:39-52` 定義 onboarding 必須有 session。
- `web/src/routes/guards.tsx:63-85` 定義未登入導到 `/login`，無 project 導到 `/onboarding`。
- `web/src/routes/guards.tsx:91-125` 定義 project ref 不存在時 fallback 到第一個 project 或 onboarding。
- `web/src/routes/routePaths.ts:4-20`、`web/src/routes/routePaths.ts:24-34`、`web/src/routes/routePaths.ts:62-103` 定義 public、legacy、project-scoped route paths 與 detail route builder。
- `web/src/routes/sidebarItems.ts:11-19` 定義 app sidebar 的主要 IA：Dashboard、Probes、Checks、Alerts、Status、Labels、Insight、Members、Settings。

### 1.2 API 與 session model

- Web client 預設呼叫 `/api/v1`，request 帶 `credentials: "include"`，代表 session 主要依賴 cookie。
- Backend versioned API base path 是 `/api/{version}`；auth、project、probe、check、assignment、label、alert、result、public status、probe runtime route 都在同一 router 註冊。
- 使用者 session auth 透過 `netstamp_session` cookie；probe runtime 不用使用者 cookie，而是 `Authorization: Probe <secret>`。
- Demo/read-only mode 會在 backend middleware 限制 mutating API；frontend 也會關閉 registration、project creation、credential changes。

Evidence:

- `web/src/shared/api/client.ts:18-23` 使用 `/api/v1` 或 env base URL，且 include credentials。
- `web/src/shared/api/client.ts:52-89` 解析 RFC/problem-like API error 並丟出 `ApiError`。
- `server/internal/controller/transport/http/router.go:85-105` 建立 API router、demo read-only middleware 與 health routes。
- `server/internal/controller/transport/http/router.go:125-142` 註冊 install、auth、user、project、alert、assignment、label、check、probe、public status、result、probe runtime routes。
- `server/internal/controller/transport/http/router.go:167-169` 建立 `/api/{version}` base path。
- `server/internal/controller/transport/http/middleware/auth.go:11-35` 使用 `netstamp_session` cookie 並在缺少/無效 session 時回 401。
- `server/internal/controller/transport/http/handler/proberuntime/context.go:19-60` 解析 `Authorization: Probe <secret>`。
- `web/src/shared/config/features.ts:19-25` 用 demo/read-only mode 控制 registration、project creation、credential changes。

## 2. Current-State User Journey Map

| Phase | User goal | User action | System action | Touchpoint | Emotion | Pain point | Opportunity | Related feature | Evidence |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 1. 第一次進入與 session 判斷 | 找到下一步：登入、註冊、或進入既有 project | 開啟 `/` 或收到舊路徑連結 | Router 導向 dashboard；未登入導向 login；無 project 導向 onboarding；project ref 錯誤時 fallback | `/`、`/login`、`/onboarding`、`/projects/:projectRef/dashboard` | 初期不確定，若 redirect 正確會快速進入工作狀態 | `/` 沒有產品脈絡；project 不存在時直接跳轉，使用者不一定知道原因 | 加入 first-run checklist、route fallback toast、project-not-found explanation | Routing、Auth guard、Project selection | `web/src/routes/router.tsx:12-79`; `web/src/routes/guards.tsx:63-125`; `web/src/shared/api/useCurrentProject.ts:5-73` |
| 2. 註冊 | 建立帳號並開始設定 workspace | 填 email、display name、password、confirm password | Frontend 驗證密碼長度與確認密碼；呼叫 register；成功後導到 onboarding；若 registration disabled 則 redirect login | `/register`、Auth form | 期待快速開始 | Register submit button 在註冊模式顯示 `Create project`，實際是建立帳號；read-only/demo 時註冊不可用但主要以 redirect 表現 | 將 CTA 改為 `Create account`；在 disabled 狀態顯示原因與替代 demo login | Auth registration | `web/src/features/auth/components/AuthPage.tsx:36-78`; `web/src/features/auth/components/AuthPage.tsx:109-150`; `web/src/routes/guards.tsx:19-33`; `web/src/shared/config/features.ts:22-25` |
| 3. 登入 | 回到既有 observability workspace | 輸入 email/password，或使用 demo credentials | 呼叫 login，session provider 重新取得 `/auth/me`，成功後導到 dashboard | `/login`、demo credential helper | 熟悉、務實 | 沒看到 forgot password/recovery flow；web session user role 被映射成固定 `Admin`，可能與 project role 模型產生語意落差 | 加入 password reset；把 session-level identity 與 project role 分開呈現 | Auth login/session | `web/src/features/auth/components/AuthPage.tsx:39-48`; `web/src/features/auth/components/AuthPage.tsx:82-85`; `web/src/features/auth/session/SessionProvider.tsx:13-53`; `web/src/features/auth/services/authService.ts:25-44`; `server/internal/controller/transport/http/handler/auth/handler.go:29-37` |
| 4. First project onboarding | 建立第一個 project，讓後續 probes/checks 有歸屬 | 輸入 project name；可加入 invite emails；提交後進入下一步 | Frontend slugify project slug；遇 conflict 追加序號重試；建立 project；可邀請 viewer；選取 project；成功頁導向 create probe | `/onboarding` | 有被引導，但仍需要理解 project/probe 概念 | 邀請角色固定 viewer；project creation disabled 時只能登出或請 operator 指派；成功後只有 probe CTA，沒有完整 setup checklist | Onboarding 增加 role selection、project purpose、setup checklist、sample topology preview | Project onboarding、Invite bootstrap | `web/src/features/auth/components/OnboardingPage.tsx:30-58`; `web/src/features/auth/components/OnboardingPage.tsx:76-82`; `web/src/features/auth/components/OnboardingPage.tsx:195-241`; `web/src/features/auth/components/OnboardingPage.tsx:248-296`; `web/src/features/auth/components/OnboardingPage.tsx:307-367` |
| 5. Project 切換與管理 | 在多個 network environment / tenant 間切換與維護基本資料 | 用 project switcher 切換；建立新 project；在 Settings 更新 name/slug；刪除或離開 project | 選取 project ref 存到 localStorage；切換時保留目前 route segment；settings mutation 更新/刪除/leave | Sidebar project switcher、Project settings page | 掌控感；刪除時緊張 | Project delete 文案提到會 revoke probe registration tokens，但 UI 對 assignments/checks/status pages 的影響預覽不明顯；localStorage selected project 在多人/多裝置上不一致 | 加入 project impact preview、recent projects、server-side last project preference | Project management | `web/src/layouts/components/ProjectSwitcher.tsx:36-79`; `web/src/layouts/components/ProjectSwitcher.tsx:106-125`; `web/src/features/project/components/CreateProjectModal.tsx:32-154`; `web/src/features/project/components/ProjectPage.tsx:60-153`; `web/src/shared/api/useCurrentProject.ts:22-73` |
| 6. Members、invites 與權限 | 讓隊友共同管理或查看 observability data | Owner/Admin 在 Members invite email + role；調整角色；移除 member；被邀請者在 account settings 接受/拒絕 | Backend permission model 判斷 read/update/member/check/probe/alert/project delete；invite 狀態 pending/accepted/rejected | `/members`、`/settings` account tab | 協作感；面對權限限制時可能困惑 | 可管理權限主要在 Members page 有 UI guard，其他頁面可能仍由 API 403/toast 才阻止；Admin 不能升級 owner/admin 等細節需要更明確 | 全站 role-aware controls；permission tooltip；invite email deliverability/status | Project members、RBAC | `web/src/features/project/components/MembersPage.tsx:70-132`; `web/src/features/project/components/MembersPage.tsx:165-179`; `web/src/features/project/components/MembersPage.tsx:293-329`; `web/src/features/settings/components/SettingsPage.tsx:98-123`; `server/internal/domain/project/permission.go:9-55`; `server/internal/domain/project/project.go:12-27` |
| 7. Labels 與 selector 準備 | 用 metadata 組織 probe fleet，讓 checks 可以套到正確 probes | 建立 label key/value；在 probe detail 指派 labels；check editor 使用 selector rules 或 advanced JSON | Label CRUD 成功後 invalidates labels/probes/checks/assignments；selector preview 可試算 matched probes | `/labels`、Probe detail label editor、Check selector builder | 有控制感，但需要理解 selector 語意 | Label 變更會重新影響 assignment，但使用者需要自己推斷 blast radius；advanced JSON 對非工程角色較硬 | Label usage impact preview；selector simulator；常用 labels template | Labels、Assignments selector | `web/src/features/labels/components/LabelsPage.tsx:88-161`; `web/src/features/labels/components/LabelsPage.tsx:249-378`; `web/src/shared/api/mutations/labels.ts:9-54`; `web/src/features/probes/components/ProbeDetail.tsx:151-184`; `web/src/features/checks/components/selectorState.ts:3-38`; `web/src/features/checks/components/selectorState.ts:125-220` |
| 8. 建立與安裝 Probe | 部署第一個觀測點，讓系統開始收資料 | 在 New Probe wizard 輸入名稱/位置；建立 probe；複製 one-time secret 與 install command；在 Linux host 執行；等待 heartbeat | Backend 建 probe 並回 plaintext secret；frontend 顯示 registration token、install command、installer/binary/uninstaller links；poll probe detail 直到 heartbeat；runtime 以 Probe secret 上報 hello/heartbeat/results | `/probes/new` drawer、shell install command、probe runtime API | 開始有成就感；安裝失敗時焦慮 | 安裝流程假設 Linux + sudo + systemd；secret 只顯示一次但 recovery 需 rotate；Finish 依 heartbeat gating，缺少 inline troubleshooting；位置搜尋依賴外部 Nominatim | OS/package detection；copy-safe secret vault；preflight diagnostics；heartbeat troubleshooting checklist；manual install alternatives | Probe create/install/runtime | `web/src/features/probes/components/NewProbeDrawer.tsx:25-100`; `web/src/features/probes/components/NewProbeDrawer.tsx:197-252`; `web/src/features/probes/components/NewProbeDrawer.tsx:254-405`; `web/src/shared/api/installAssets.ts:4-74`; `server/internal/controller/transport/http/handler/probe/handler.go:26-36`; `server/internal/controller/transport/http/handler/proberuntime/handler.go:20-29`; `server/internal/controller/transport/http/handler/proberuntime/context.go:19-84` |
| 9. 管理 Probe fleet | 查看每個觀測點狀態、位置、版本、labels，並做維護 | 切換 grid/map；搜尋/排序；開 detail drawer；更新名稱、labels、位置；rotate secret；reinstall/upgrade/delete | Frontend 2 秒 refetch probes；status 由 heartbeat 映射 online/draining/offline；secret rotation 回新 secret 與 update command | `/probes`、Probe detail drawer、map/table | 對 fleet 狀況有掌握；維護時小心 | Secret rotation 後仍需手動更新 host；delete 對 checks/status/alerts 影響不夠視覺化；缺少 bulk labels/maintenance mode | Fleet bulk operations；probe health diagnostics；delete impact graph；agent update campaign | Probe management | `web/src/features/probes/components/ProbesPage.tsx:21-96`; `web/src/features/probes/components/ProbeList.tsx:84-129`; `web/src/features/probes/components/ProbeDetail.tsx:96-184`; `web/src/features/probes/components/ProbeDetail.tsx:381-565`; `web/src/features/probes/api/probeAdapters.ts:35-88`; `web/src/shared/api/mutations/probes.ts:7-81` |
| 10. 建立與管理 Checks | 定義要從哪些 probes 監測哪些 targets | 建立 Ping/TCP/Traceroute check；設定 interval、target、type config；選 all probes 或 selector；preview matched probes；duplicate/edit/delete/batch delete | Check CRUD 後 invalidates checks/assignments/results；selector preview 由 API 驗證；check type edit 時 locked | `/checks`、Check drawer、selector preview | 有力量但容易擔心配置錯誤 | New check 預設 all probes，可能不小心產生大量 assignments；advanced selector JSON 門檻高；目前 UI/contract 僅支援 Ping/TCP/Traceroute，沒有 HTTP/DNS check evidence | Check templates；safe default selector；assignment cost preview；HTTP/DNS check 若 product 需要應先補 backend executor/contract | Checks、Assignments | `web/src/features/checks/components/ChecksPage.tsx:65-180`; `web/src/features/checks/components/ChecksPage.tsx:234-335`; `web/src/features/checks/components/ChecksPage.tsx:374-521`; `web/src/features/checks/components/ChecksPage.tsx:524-725`; `web/src/features/checks/components/CheckConfigFields.tsx:21-211`; `web/src/features/checks/data/checkConfig.ts:47-120`; `web/src/shared/api/mutations/checks.ts:19-123`; `server/internal/controller/transport/http/handler/check/handler.go:26-35` |
| 11. Dashboard overview | 快速知道目前 probes/checks 是否活著 | 開啟 dashboard，查看 online probes、active checks、network map，點 New Probe | Dashboard 讀 probes/checks，計算 summary，渲染 NetworkMap | `/dashboard` | 快速掌握大局 | Dashboard 偏總覽，尚未看到 incident triage、最近失敗、onboarding progress；新使用者沒有資料時下一步主要靠 New Probe | 加入 empty-state setup progress、recent incidents、latest failing assignments、topology shortcut | Dashboard | `web/src/features/dashboard/components/DashboardPage.tsx:21-59`; `web/src/shared/components/NetworkMap.tsx` |
| 12. Results、Insights、Topology | 從 probe/check 結果診斷 latency、loss、connect failure、路由變化 | 在 Insight 選時間範圍、refresh、check type、group/scope、assignment；查看 charts、metrics、traceroute hops/topology | Frontend 查 ping/tcp series + insight、traceroute runs/insight/topology；URL state 保存 scope；auto-refresh invalidates project detail | `/insight`、charts、route topology map、focus chips | 進入分析模式；若沒有資料容易卡住 | Multi-series 不支援 traceroute；invalid shared scope/no active paths 需要更強 recovery；結果視圖和 check/probe detail 的上下文跳轉可更完整 | First-result guided handoff；traceroute diff/compare；from incident/check/probe deep link with context | Results、Insight、Topology | `web/src/features/insight/components/InsightPage.tsx:101-218`; `web/src/features/insight/components/InsightPage.tsx:309-335`; `web/src/features/insight/components/InsightPage.tsx:517-647`; `web/src/features/insight/components/PingInsightPanel.tsx:23-71`; `web/src/features/insight/components/TcpInsightPanel.tsx:22-71`; `web/src/features/insight/components/TracerouteInsightPanel.tsx:89-164`; `web/src/features/insight/components/MultiSeriesInsightPanel.tsx:102-188`; `api/services/results.tsp:166-265`; `server/internal/controller/transport/http/handler/result/handler.go:26-38` |
| 13. Alerts 與 integrations | 在指標異常時收到通知，並追蹤 incident | 建立 notification channel；test notification；建立 alert rule；選 scope/metric/operator/threshold/window/cooldown；看 incidents/timeline/notifications | Alert rules/incidents/notifications 30 秒 refetch；backend 只對 Ping/TCP changed assignments evaluate；firing 建 incident 並 enqueue notification；resolved 時 enqueue resolved notification | `/alerts` Incidents/Rules/Notifications tabs、notification editor、rule editor、incident drawer | 對生產狀態更放心；設定 integrations 時偏工程化 | Traceroute alert 在 UI 明確 disabled，因 controller API 只 exposes ping/TCP metrics；notification write 需要 owner/admin，比 manage alerts 更嚴；Email 需 SMTP configured；webhook/slack/discord/telegram/email config 都需要使用者知道外部服務參數 | Alert setup wizard；rule recommendation from historical baseline；notification health/status；integration-specific validation docs | Alerts、Incidents、Notifications | `web/src/features/alerts/components/AlertsPage.tsx:113-164`; `web/src/features/alerts/components/AlertsPage.tsx:465-587`; `web/src/features/alerts/components/AlertsPage.tsx:668-855`; `web/src/features/alerts/components/AlertsPage.tsx:891-987`; `web/src/features/alerts/components/alertPageModel.ts:78-139`; `web/src/features/alerts/components/alertPageModel.ts:581-619`; `api/models/alert.tsp:12-22`; `api/services/alerts.tsp:9-136`; `server/internal/controller/application/alert/service.go:27-165`; `server/internal/controller/application/alert/service.go:198-243`; `server/internal/controller/application/alerteval/service.go:39-184`; `server/internal/controller/infrastructure/notify/sender.go:38-67` |
| 14. Public status pages | 對外公開服務健康狀態與 charts | 建立 status page；設定 slug/title/description/chart mode/range；加入 folder 或 assignment group element；開 public link | Authenticated API 管理 pages/elements；public endpoint 依 slug 回傳 page status、open incidents、elements、assignment metrics、charts；disabled/not found 回 404 UI | `/status-pages`、`/status/:slug` | 分享狀態透明度；公開前需要信心 | Element editor 只允許 Ping/TCP assignment，Traceroute 被排除；public 404 文案同時涵蓋 disabled/not exist；沒有看到 custom domain/subscriber/maintenance window evidence | Publish review checklist；status page preview；maintenance windows；custom domain/subscribers 若產品需要另建 | Public Status | `web/src/features/status-pages/components/StatusPagesPage.tsx:71-202`; `web/src/features/status-pages/components/StatusPagesPage.tsx:204-300`; `web/src/features/status-pages/components/StatusPageEditorDrawer.tsx:89-123`; `web/src/features/status-pages/components/StatusElementEditorDrawer.tsx:45-100`; `web/src/features/status-pages/components/StatusElementEditorDrawer.tsx:204-312`; `web/src/features/status-pages/components/PublicStatusPage.tsx:14-69`; `web/src/features/status-pages/components/PublicStatusPage.tsx:75-265`; `web/src/shared/api/mutations/statusPages.ts:12-40`; `server/internal/controller/transport/http/handler/publicstatus/handler.go:21-35` |
| 15. Account settings | 維護個人資料、接受 project invite、更新登入憑證 | 在 Settings 接受/拒絕邀請；更新 profile、email、password | 接受 invite 後選取 project 並導到 dashboard；credential changes 受 feature flag 控制 | `/settings` account page | 掌控個人狀態 | 密碼更新需要 current password，但沒有 reset flow；demo/read-only 會停用 credential changes | 加入 password reset、security sessions、API tokens、credential-disabled explanation | Account/User settings | `web/src/features/settings/components/SettingsPage.tsx:45-55`; `web/src/features/settings/components/SettingsPage.tsx:98-123`; `web/src/features/settings/components/SettingsPage.tsx:171-238`; `api/services/users.tsp:35`; `web/src/shared/config/features.ts:22-25` |
| 16. 錯誤、空狀態與無權限 | 從錯誤中理解原因並恢復 | 遇到未登入、project 不存在、no project、API 400/401/403/404/422、read-only mode、probe credential invalid、public status 404 | Backend 回 problem+json；frontend ApiError 轉 toast/empty/loading states；route guard redirect；runtime auth 回 `WWW-Authenticate: Probe` | Toasts、LoadingState、empty states、redirect、public status 404 panel | 可能焦慮或疑惑 | 錯誤多為通用 message/toast；role/permission 的前置說明不完整；read-only mutation failure 可能在 action 後才感知 | 建立 role-aware disabled controls、error recovery CTA、request ID copy、diagnostic panel | Error handling、Permissions | `server/internal/controller/transport/http/httpx/httpx.go:30-69`; `server/internal/controller/transport/http/httpx/httpx.go:75-133`; `server/internal/controller/transport/http/not_found.go:10-29`; `web/src/shared/api/client.ts:52-89`; `web/src/shared/api/mutations/shared.ts:14-29`; `server/internal/controller/transport/http/handler/proberuntime/context.go:62-84`; `web/src/features/status-pages/components/PublicStatusPage.tsx:22-43` |

## 3. Service Blueprint

| Service moment | Frontstage user experience | Web/API touchpoint | Backstage service action | Support process / data dependency | Failure / permission mode | Evidence |
| --- | --- | --- | --- | --- | --- | --- |
| Session bootstrap | 使用者看到 login/register、或直接進 app | `/auth/login`、`/auth/register`、`/auth/me`; React guards | Auth service 設定/讀取 `netstamp_session`; session provider 將 API user 映射為 web session | Cookie、JWT verifier、read-only/demo feature flags | Missing/invalid cookie -> 401；registration disabled -> login；demo/read-only mutating API 受限 | `web/src/features/auth/session/SessionProvider.tsx:13-53`; `server/internal/controller/transport/http/handler/auth/handler.go:29-37`; `server/internal/controller/transport/http/middleware/auth.go:11-35`; `server/internal/controller/transport/http/router.go:99-103` |
| Project bootstrap | 使用者建立 project、邀請初始 viewer、選取 project | `/projects`; `/projects/{ref}/members/invites`; localStorage selected ref | Project service 建立 slug/name；invite service 建 pending invites；web fallback first project | Project slug validation、role model、query cache invalidation | Project creation disabled；slug conflict；no assigned project -> onboarding disabled state | `web/src/features/auth/components/OnboardingPage.tsx:195-241`; `web/src/shared/api/mutations/projects.ts:12-25`; `server/internal/domain/project/project.go:64-100`; `web/src/routes/guards.tsx:76-85` |
| Project access control | 使用者切換 project、管理 member/role | `/projects/{ref}`, `/members`, `/invites`, `/users/me/invites` | Backend permission model 以 owner/admin/editor/viewer 控制 read/update/member/probe/check/label/alert/delete | Project membership table、invite status、role transition rules | Viewer 可讀但不能寫；delete project owner only；notification write owner/admin only | `server/internal/domain/project/permission.go:9-55`; `server/internal/domain/project/project.go:12-27`; `server/internal/controller/application/alert/service.go:198-207`; `web/src/features/project/components/MembersPage.tsx:126-132` |
| Probe registration | 使用者建立 probe 並複製 install command | `/projects/{ref}/probes`; `/install/agent.sh`; `/install/netstamp-agent-linux-*` | Backend create probe 回 one-time plaintext secret；install script 下載 binary；agent 後續用 probe secret 與 runtime API 通訊 | Probe secret generation/verification、Linux binary hosting、public IP/heartbeat metadata | Secret 遺失需 rotate；install host 無 sudo/systemd 或無法連 controller 時無 heartbeat | `web/src/features/probes/components/NewProbeDrawer.tsx:235-405`; `web/src/shared/api/installAssets.ts:4-74`; `server/internal/controller/transport/http/handler/probe/handler.go:26-36`; `server/AGENTS.md:31` |
| Probe runtime operations | 已安裝 agent heartbeat、拉 assignments、送 results | `/runtime/probes/{probe_id}/hello`; `/heartbeat`; `/assignments`; `/results`; `Authorization: Probe <secret>` | Runtime service 驗證 probe credential、更新 status/IP family capabilities、接收 ping/tcp/traceroute results | Agent service、network reachability、probe enabled state | Missing/invalid secret -> 401；disabled probe -> 403；invalid runtime input -> 422 | `server/internal/controller/transport/http/handler/proberuntime/handler.go:20-75`; `server/internal/controller/transport/http/handler/proberuntime/context.go:19-113` |
| Check assignment | 使用者建立 check 並選 selector | `/projects/{ref}/checks`; `/assignments`; `/assignment-selector-preview` | Backend 根據 selector 與 labels/probes 建立 active assignment pairs；results 與 alerts 依 assignments 運作 | Probe labels、selector grammar、check config validation | Selector invalid -> 422；all-probes default 可能產生大範圍 assignment | `web/src/features/checks/components/ChecksPage.tsx:234-335`; `web/src/features/checks/components/ChecksPage.tsx:477-521`; `server/internal/controller/transport/http/handler/check/handler.go:26-35`; `server/internal/controller/transport/http/handler/assignment/handler.go:26-32` |
| Result ingestion and insight | 使用者看 latest、series、insight、topology | `/projects/{ref}/results/latest`; `/ping/series`; `/ping/insight`; `/tcp/series`; `/tcp/insight`; `/traceroute/runs`; `/traceroute/insight`; `/traceroute/topology` | Result service 聚合 time series、latest status、traceroute hop/topology data；frontend auto-refresh | Timeseries data、assignment IDs、time range/max data points | 無資料、invalid shared scope、topology unavailable -> empty/loading states | `web/src/features/insight/components/InsightPage.tsx:191-218`; `api/services/results.tsp:166-265`; `server/internal/controller/transport/http/handler/result/handler.go:26-38` |
| Alerts and notifications | 使用者設定 rule、notification、看 incidents | `/projects/{ref}/alert-rules`; `/alert-incidents`; `/notifications`; `/notifications/{id}/test` | Alert evaluator 在 changed assignments 後只評估 ping/tcp；建立/resolve incidents；enqueue notification jobs；sender 支援 webhook/slack/discord/telegram/email | Notification secrets/config、SMTP config、backend base URL、cooldown/window/min samples | Traceroute unsupported；notification write owner/admin；delivery failure/retry由 notification service 承接 | `api/services/alerts.tsp:9-136`; `server/internal/controller/application/alerteval/service.go:39-184`; `server/internal/controller/infrastructure/notify/sender.go:38-67`; `web/src/features/alerts/components/AlertsPage.tsx:727-855` |
| Public status publishing | 使用者把部分 checks 對外公開 | `/projects/{ref}/status-pages`; `/public/status-pages/{slug}`; browser `/status/:slug` | Authenticated service 管理 pages/elements；public service 組合 page status、active incidents、assignment metrics、charts | Status page slug、enabled flag、assignment visibility、chart data | Disabled/not found -> public 404 panel；element assignment options 只含 ping/tcp | `web/src/features/status-pages/components/StatusPagesPage.tsx:204-300`; `web/src/features/status-pages/components/PublicStatusPage.tsx:14-69`; `server/internal/controller/transport/http/handler/publicstatus/handler.go:21-35` |
| Error handling | 使用者看到 toast、empty state、loading/error panel | `ApiError`、problem+json、route guards、ConfirmProvider | Backend 統一 `WriteProblem`；frontend `requestErrorMessage` 與 mutation `onError` 顯示 toast | HTTP status、field errors、request validation、permission errors | 多數 recovery CTA 不具體；權限與 read-only 狀態可能晚於 action 才揭露 | `server/internal/controller/transport/http/httpx/httpx.go:30-133`; `web/src/shared/api/client.ts:52-89`; `web/src/shared/api/mutations/shared.ts:14-29`; `web/src/features/status-pages/components/PublicStatusPage.tsx:32-43` |

## 4. 權限與角色現況

| Role / policy | Current behavior | UX implication | Evidence |
| --- | --- | --- | --- |
| Owner | 可做 project update、members、labels、checks、probes、alerts、delete project；可管理 notification | 最高權限；刪除 project 是 owner-only，需要高風險確認與 impact preview | `server/internal/domain/project/permission.go:17-29`; `server/internal/controller/application/alert/service.go:198-207` |
| Admin | 可做大多數 project 管理與 content mutation；不能 delete project；member management 有 owner/admin 限制 | UI 應避免讓 admin 誤以為可以刪 project 或管理 owner/admin 邊界 | `server/internal/domain/project/permission.go:17-55`; `web/src/features/project/components/MembersPage.tsx:126-132` |
| Editor | 可管理 labels/checks/probes/alerts，但 notification write backend 要 owner/admin | Alert rule 與 notification 的權限差異可能讓 editor 卡在「可以建 rule，但不能建 channel」 | `server/internal/domain/project/permission.go:21-28`; `server/internal/controller/application/alert/service.go:198-207` |
| Viewer | 可讀 project，但不可寫 | 應在 UI 提前 disabled mutation controls，而不是等 API 403 | `server/internal/domain/project/permission.go:17-29`; `web/src/shared/api/mutations/shared.ts:26-29` |
| Demo/read-only | Frontend feature flags 停用 registration/project creation/credential changes；backend middleware 限制 mutating API | Demo 使用者需要明確知道哪些操作只能看不能改 | `web/src/shared/config/features.ts:19-25`; `server/internal/controller/transport/http/router.go:99-103` |

## 5. 不適用、限制與 Hypotheses

### 5.1 Document / stamp / certificate / verification

- Current state：不適用。未找到文件蓋章、certificate issuance、certificate verification、document signing、verification portal 等產品功能的 web route、API service、backend handler 或 domain model。
- 搜尋 evidence：針對 `web/src`、`server/internal`、`api/services`、`api/models`、`packages`、`docs/src/content` 使用 `certificate|certificates|document|documents|verification|verifications` 精準搜尋，結果主要是 DOM `document`、一般 docs 文案、以及使用者變更密碼時的「verify existing password」。沒有 product-level document/stamp/certificate/verification workflow。
- Hypothesis：若「stamp」是品牌名 Netstamp 的一部分，不能推論成文件蓋章功能。若未來產品方向需要「觀測結果證明」或「SLA certificate」，應新增明確 domain model、API、UI route 與 audit trail。

### 5.2 Check type 覆蓋

- Current state 支援 Ping、TCP、Traceroute。
- 沒有可驗證的 HTTP/DNS check executor evidence；若要新增，需從 TypeSpec contract、backend domain、runtime agent executor、results model、UI config、alerts/status pages 一起補齊。
- Alerts 目前只支援 Ping/TCP metrics；Traceroute alert 在 UI 被 disabled，backend evaluator 對非 Ping/TCP 直接忽略。

Evidence:

- `web/src/features/checks/components/CheckConfigFields.tsx:21-211` 只顯示 Traceroute/TCP/Ping config fields。
- `web/src/features/checks/data/checkConfig.ts:47-68` 定義 ping/tcp/traceroute defaults。
- `web/src/features/alerts/components/alertPageModel.ts:119-139` 顯示 traceroute disabled 且 metrics 只含 ping/tcp。
- `server/internal/controller/application/alerteval/service.go:52-59` 非 Ping/TCP 不評估。
- `server/internal/controller/application/alert/service.go:226-243` rule normalization 拒絕 traceroute。

### 5.3 Probe install 平台

- Current state install assets 明確是 Linux amd64/arm64 binary 與 shell install/uninstall scripts。
- 沒有 macOS/Windows package、Kubernetes manifest、Docker sidecar、Terraform module 的 UI evidence。

Evidence:

- `web/src/shared/api/installAssets.ts:4-11` 定義 agent shell scripts 與 Linux binaries。
- `server/internal/controller/transport/http/router.go:125-142` 註冊 install route。

### 5.4 Public status 能力

- Current state 有 public status pages，但 status elements 的 assignment options 只納入 Ping/TCP，排除 Traceroute。
- 沒有 custom domain、subscriber notification、maintenance window、incident manual postmortem 的 code evidence。

Evidence:

- `web/src/features/status-pages/components/StatusElementEditorDrawer.tsx:83-100` public assignment options 過濾 Ping/TCP。
- `web/src/features/status-pages/components/PublicStatusPage.tsx:75-265` public page 顯示 active incidents、checks、assignment metrics、charts。

## 6. Future-State Journey

| Future phase | Ideal flow | Needed feature / change | Why it matters | Current-state basis / gap |
| --- | --- | --- | --- | --- |
| 1. Entry and product orientation | 第一次進入時，使用者能清楚知道自己要登入、註冊、使用 demo，或等待 operator 指派 project | `/` 或 login 前增加 authenticated-aware intro；demo/read-only banner；route fallback message | 減少被 redirect 的困惑，尤其 self-host/demo 環境 | Current router 直接 redirect，registration disabled 時導 login；`web/src/routes/router.tsx:12-17`; `web/src/routes/guards.tsx:19-33` |
| 2. Account + project bootstrap | 註冊後一個流程完成 project、初始角色、第一個 probe/check/status goal | Onboarding checklist；invite role selection；project template（global edge、internal DC、API uptime） | 把「建立 project」從表單變成可理解的 observability setup | Current onboarding 可建 project + viewer invites，但缺 role/template/checklist；`web/src/features/auth/components/OnboardingPage.tsx:195-367` |
| 3. First probe deployment | Wizard 根據 OS/architecture/環境產生 install path，並在 UI 顯示 preflight、heartbeat、agent version、network reachability | OS/package/Kubernetes options；preflight command；copy command audit；heartbeat troubleshooting | Probe 是產品 activation 的核心，安裝失敗會阻斷所有後續價值 | Current install 偏 Linux shell + manual secret；`web/src/shared/api/installAssets.ts:4-74`; `web/src/features/probes/components/NewProbeDrawer.tsx:340-405` |
| 4. Fleet organization | 安裝後自動建議 labels，使用者可批次套用並看到 selector impact | Bulk label operations；auto labels from location/IP/agent metadata；selector impact preview | 降低 checks 錯配或 all-probes 誤用風險 | Current labels 有 usage，但 impact 與 bulk flow 不完整；`web/src/features/labels/components/LabelsPage.tsx:88-161`; `web/src/features/checks/components/selectorState.ts:125-220` |
| 5. Check creation | 使用者從 template 建 Ping/TCP/Traceroute checks，預覽 matched probes、run cost、first result ETA | Safe default selector；target templates；assignment count/cost preview；first-run status | Checks 是資料來源，錯誤配置會造成無資料或過大 blast radius | Current new check default all probes；selector preview exists but safety guidance 可增強；`web/src/features/checks/components/ChecksPage.tsx:174-180`; `web/src/features/checks/components/ChecksPage.tsx:629-705` |
| 6. First result handoff | 第一筆 heartbeat/result 到達後，自動把使用者帶到 Insight，並標出「這筆資料從哪個 probe/check 來」 | First-result notification；deep link to insight with selected assignment/time range；empty-state ETA | 把安裝成功轉成產品價值瞬間 | Current Insight 需要使用者自行選 scope；`web/src/features/insight/components/InsightPage.tsx:517-647` |
| 7. Diagnosis and topology | Incident/check/probe/detail 都能一鍵進入對應 time range 的 chart/topology，Traceroute 支援 diff/route change explanation | Cross-feature deep links；route diff；hop annotation；saved views | 網路 observability 的核心是從異常快速定位路徑/節點 | Current traceroute topology exists，但 multi-series traceroute 不支援；`web/src/features/insight/components/TracerouteInsightPanel.tsx:89-164`; `web/src/features/insight/components/MultiSeriesInsightPanel.tsx:102-188` |
| 8. Alert setup | 使用者從 check 或 insight 建 rule，系統基於歷史 baseline 推 threshold，並完成 notification end-to-end test | Alert recommendation；integration-specific validation；delivery history；role-aware notification permissions | Alert 設定若太抽象，容易過度告警或完全沒告警 | Current alert editor 功能完整但 technical；traceroute unsupported；`web/src/features/alerts/components/AlertsPage.tsx:711-855`; `server/internal/controller/application/alerteval/service.go:52-59` |
| 9. Status publishing | 使用者可從 check group 一鍵建立 public status page，先 preview，再 publish，並看到公開 URL 健康狀態 | Publish checklist；preview mode；custom domain/subscriber/maintenance windows if product requires | 對外透明度需要安全審查與預覽 | Current status pages 可建立與公開，但沒有 publish review/custom domain evidence；`web/src/features/status-pages/components/StatusPagesPage.tsx:204-300` |
| 10. Collaboration and permissions | 全站 mutation controls 依 project role 顯示能力；無權限時給出「誰可以幫忙」與 request access path | Role-aware UI primitives；permission explanation; request access/invite flow | 減少 403 toast 造成的挫折 | Current backend permission 清楚，但 UI guard 分散；`server/internal/domain/project/permission.go:9-55`; `web/src/features/project/components/MembersPage.tsx:126-132` |
| 11. Error recovery | 每種錯誤都有下一步：重新登入、選 project、請 admin、rotate probe secret、查看 install diagnostics | Error taxonomy；request ID；diagnostic copy block；probe troubleshooting | Observability tool 的信任感來自清楚可恢復 | Current problem+json 和 ApiError 已具備基礎，但 recovery CTA 不完整；`server/internal/controller/transport/http/httpx/httpx.go:30-133`; `web/src/shared/api/client.ts:52-89` |
| 12. Certificate / verification as optional future | 若未來要做「觀測證明」或「SLA attestation」，應從 result snapshot、signature、audit log、public verification route 設計，不應混同現有 Netstamp 品牌名 | New domain models：attestation/certificate/snapshot/signature；new routes：issue/verify；UI：evidence export | 避免把沒有 evidence 的文件蓋章功能誤寫進 current product | Current code 沒有此產品 flow；見 5.1 |

## 7. Prioritized Opportunities

### High impact

1. First-run setup checklist：串起 project -> probe -> check -> first result -> alert/status page，並在 dashboard/onboarding 顯示進度。
2. Probe install diagnostics：在 New Probe wizard 加入 preflight command、常見錯誤、heartbeat debug、secret rotation recovery。
3. Safe check creation：把 all-probes default 改成明確選擇，並在儲存前顯示 matched probe count、selector impact、first result ETA。
4. Role-aware UI：根據 project role 與 read-only mode disabled controls，減少 action 後才收到 403/toast。
5. Alert wizard：從 check/insight 建 rule，提供 threshold suggestion、notification test status、traceroute unsupported explanation。

### Medium impact

1. Project deletion/leave impact preview：列出 probes、checks、assignments、alerts、status pages 受影響項目。
2. Label blast-radius preview：更動 label 時顯示哪些 checks/assignments 會改變。
3. Insight deep links：從 dashboard/probe/check/incident 跳到帶 time range、assignment、type 的 insight URL。
4. Public status publish review：公開前預覽、確認 exposed checks、顯示 disabled/not-found 差異。
5. Account recovery：加入 forgot password/session management；demo/read-only 時顯示清楚限制。

### Low / hypothesis

1. Custom domain、subscribers、maintenance windows：目前無 code evidence，但可作為 status page 成熟化方向。
2. HTTP/DNS check types：需完整 backend/agent/API/UI 工作，不應只做 UI option。
3. Result certificate / verification：目前不適用；若產品策略需要，應作為新 capability 設計。
