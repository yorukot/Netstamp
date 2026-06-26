# Netstamp MVP UX Design 文件

建立日期：2026-06-18

狀態：MVP UX/UI 改版規劃草稿

適用範圍：目前 React Web App、公開首頁、Docs、OpenAPI Explorer、共用 UI 元件使用情境

## 1. 文件目的

這份文件用來整理 Netstamp 目前 MVP 的 UX 現況，並把「現在可以做什麼」、「使用者需要什麼」、「流程卡在哪裡」、「下一版應該先改什麼」整理成可執行的產品設計文件。

本文件不取代 `design.md`。`design.md` 是視覺系統、色彩、版型、元件風格的 source of truth；本文件聚焦在需求分析、功能分類、使用者歷程、現況問題、MVP 範圍與改版優先級。

Netstamp 的核心產品承諾是：

> 使用者可以部署自己控制的 probes，設定重複執行的 checks，觀察 latency、packet loss、TCP 連線、traceroute route 行為，並把 incidents 通知給團隊。

MVP 必須先證明一條完整的操作閉環：

1. 建立或進入 project。
2. 註冊 probe。
3. 確認 probe heartbeat。
4. 建立 check。
5. 看到 measurement result。
6. 設定 alert rule 和 notification。
7. 邀請另一位 operator，或分享可還原的分析 URL。

## 2. 產品定位

### 2.1 產品類型

Netstamp 是自架型 network observability 工具，位置介於 uptime monitoring、synthetic checks、route diagnostics、self-hosted infrastructure tooling 之間。

它不是一般 SaaS 監控後台，也不是完整 NOC 平台。MVP 最重要的價值不是畫很多圖，而是讓使用者從自己控制的網路位置得到可信的測量證據。

### 2.2 MVP 價值主張

Netstamp 需要讓技術使用者快速回答這些問題：

- 我的 probes 是否在線、是否還有 heartbeat？
- 不同 probes 是否能連到我在意的 targets？
- latency、packet loss、TCP reachability 或 routes 是否有變化？
- 出問題的是哪個 probe/check assignment？
- 誰需要被通知？
- 這些資料是否來自我信任、我控制的網路位置？

### 2.3 MVP 不應承諾的範圍

目前 MVP 不應包裝成完整 incident management、billing product、generic metrics platform 或完整 NOC。若 DNS checks 或 HTTP checks 尚未在 controller、probe runtime、API 和 UI 端完整實作，也不應在 MVP UI 或主要文案中當成可執行能力宣傳。

## 3. 需求分析

### 3.1 北極星指標

新使用者應該能在 10 分鐘內完成：

> 建立 project → probe online → 建立第一個 check → 看到第一筆可信 measurement。

### 3.2 商業與產品需求

- 降低第一次部署和第一次看到資料的成本。
- 讓公開網站、文件、OpenAPI 和 authenticated app 共同支持「可自架、可驗證、可操作」的信任感。
- 讓 SRE、network operator、infra admin 在重複操作中能快速掃描狀態。
- 讓 labels、selectors、checks、assignments 這套核心模型變得容易理解。
- 讓 alerts 從設定到 incident triage 有完整路徑。
- 讓團隊管理、角色、邀請、project/account settings 的邏輯清楚。

### 3.3 使用者類型

| 使用者                         | 使用情境                         | 核心需求                                     | UX 風險                        |
| ------------------------------ | -------------------------------- | -------------------------------------------- | ------------------------------ |
| Self-hosted evaluator          | 本機或 VPS 試跑 Netstamp         | 快速知道產品有沒有價值                       | 初次設定太長就放棄             |
| SRE / network operator         | 維護跨地區、跨 ISP、跨環境的服務 | 比較 reachability、latency、loss、route 變化 | 需要 triage，不需要裝飾型 UI   |
| Infrastructure admin           | 管理 projects、probes、alerts    | 安全設定、避免誤刪、管理權限                 | 權限與後果不清楚會造成操作風險 |
| Viewer / stakeholder           | 查看 dashboard 和 incidents      | 快速理解目前網路健康狀態                     | 原始設定資訊太多會阻礙理解     |
| Open-source contributor        | 閱讀 docs、API、程式碼           | 理解架構與可擴充點                           | 文案與實作不一致會降低信任     |
| Developer / automation builder | 使用 OpenAPI 或 scripts          | 穩定 API contract 與可預測的錯誤             | UI 和 API 名詞不一致會增加成本 |

### 3.4 使用者 Jobs To Be Done

#### 評估產品

當我部署一個自架型 observability 工具時，我想快速讓 probe online 並看到真實 measurement，這樣我才能判斷是否值得繼續使用。

#### 維護 Probe Fleet

當我管理不同網路位置的 probes 時，我想同時看到健康狀態、location、labels、assigned checks、安裝與維護命令，這樣我才能信任每個 vantage point。

#### 建立測量工作

當我需要重複測量某個 target 時，我想建立 Ping、TCP 或 Traceroute check，並用 labels 自動指派 probes，這樣新 probes 加入後也能自動收到正確工作。

#### 分析結果

當 latency、loss、TCP connect 或 route 發生變化時，我想用 check、probe、assignment、type、time window 進行範圍篩選，判斷問題是 target-specific、region-specific、route-specific 還是 global。

#### 設定告警

當某個 threshold 被突破時，我想讓 rule 建立 incident 並通知團隊，這樣我不用一直盯著 dashboard。

#### 管理團隊

當多位 operator 使用同一個 project 時，我想清楚管理 invites、roles、members 和 account settings，這樣權限變動是安全且可預期的。

## 4. 現有資訊架構

### 4.1 Public Surfaces

- Public landing page
- Documentation
- OpenAPI Explorer
- Storybook for shared UI primitives

### 4.2 Authenticated App

- Dashboard
- Probes
- Checks
- Alerts
- Labels
- Insight
- Members
- Project settings
- Account settings through user menu

### 4.3 建議的導覽模型

目前 IA 大致成立，但命名和心智模型需要修正。

| 目前標籤           | 建議標籤          | 原因                                                                        |
| ------------------ | ----------------- | --------------------------------------------------------------------------- |
| Dashboard          | Dashboard         | 保留，作為健康總覽與 next action hub                                        |
| Probes             | Probes 或 Fleet   | Fleet 更像營運概念，但 MVP 階段 Probes 更直覺                               |
| Checks             | Checks            | 保留                                                                        |
| Alerts             | Alerts            | 保留                                                                        |
| Labels             | Labels            | 保留，但需要靠近 Checks 的心智模型，因為 labels 驅動 selectors              |
| Insight            | Insight           | 保留，但必須從 Alerts、Checks、Probes 深連過去                              |
| Members            | Members           | 保留                                                                        |
| Settings           | Project           | sidebar 的 Settings 實際上是 project settings，容易和 account settings 混淆 |
| User menu Settings | Account           | user menu 內的設定應命名為 Account                                          |
| Invites            | Account / Invites | 邀請接受是個人帳號流程，不是 project settings                               |

## 5. 完整功能盤點與分類

### 5.1 Public、Docs、Developer Support

| 功能                        | 現況 | 分類                | UX 備註                       |
| --------------------------- | ---- | ------------------- | ----------------------------- |
| Public landing page         | 已有 | acquisition         | 需避免宣傳尚未實作的 DNS 能力 |
| Dashboard screenshot assets | 已有 | trust building      | 有助於產品可信度              |
| Documentation shell         | 已有 | support             | 文件是初次部署的重要輔助      |
| OpenAPI Explorer            | 已有 | developer workflow  | 支援 automation 與 API 信任   |
| Storybook static output     | 已有 | design system       | 支援元件一致性                |
| Tracking consent            | 已有 | privacy / analytics | 需保持簡潔且不干擾操作        |

### 5.2 Auth、Onboarding、Project Entry

| 功能                      | 現況         | 分類       | UX 備註                           |
| ------------------------- | ------------ | ---------- | --------------------------------- |
| Login                     | 已有         | core       | 表單方向清楚                      |
| Registration              | feature flag | core       | submit label 需要更準確           |
| Demo credentials          | feature flag | evaluation | 有助 demo mode                    |
| First project onboarding  | 已有         | activation | 風格和主 app 不一致，流程可更直接 |
| Onboarding invite members | 已有         | activation | 需要角色與邀請結果回饋            |
| No-project state          | 已有         | recovery   | 可再加入可操作下一步              |
| Project switcher          | 已有         | navigation | 支援 project scoped routes        |
| Create project modal      | 已有         | workspace  | 需和 onboarding 行為保持一致      |

### 5.3 Dashboard

| 功能                 | 現況 | 分類       | UX 備註                        |
| -------------------- | ---- | ---------- | ------------------------------ |
| Probes online metric | 已有 | fleet      | 數字不足以支援 triage          |
| Active checks metric | 已有 | checks     | 缺少 failed/degraded 摘要      |
| Fleet map            | 已有 | overview   | 有視覺價值，但不能取代狀態摘要 |
| New probe CTA        | 已有 | activation | 適合初次使用者                 |

Dashboard 目前比較像 fleet map，而不是 operational dashboard。下一版需要承擔 setup progress、open incidents、stale probes、recent degraded assignments 的入口角色。

### 5.4 Probe Fleet

| 功能                         | 現況          | 分類         | UX 備註                              |
| ---------------------------- | ------------- | ------------ | ------------------------------------ |
| Probe list                   | 已有          | core         | 支援 search/sort                     |
| Probe map view               | 已有          | spatial      | 可用於理解分布                       |
| Probe detail drawer          | 已有          | edit/inspect | 資訊密度高                           |
| Probe labels                 | 已有          | taxonomy     | 不能 inline 建立 label               |
| Probe location editing       | 已有          | metadata     | 需要 search/manual fallback          |
| New probe wizard             | 已有          | activation   | 是 MVP 最重要 activation flow        |
| Location search              | 已有          | convenience  | 外部 geocode 失敗時需有明確 fallback |
| Manual coordinates           | 已有          | fallback     | 對技術使用者可接受                   |
| Install command generation   | 已有          | activation   | 需要 troubleshooting 補強            |
| Heartbeat polling            | 已有          | activation   | 是很好的信任回饋                     |
| Rotate secret                | 已有          | security     | 需要更清楚說明後果                   |
| Reinstall / upgrade commands | 已有          | maintenance  | 建議改為更明確 command cards         |
| Delete probe                 | 已有確認      | destructive  | 符合預期                             |
| Enable / disable / drain     | UI 未清楚暴露 | gap          | 若 API 支援，MVP 應提供              |

### 5.5 Checks And Assignments

| 功能                   | 現況           | 分類          | UX 備註                  |
| ---------------------- | -------------- | ------------- | ------------------------ |
| Check list             | 已有           | core          | 支援 operational table   |
| Search and type filter | 已有 URL state | repeated work | 方向正確                 |
| Create check           | 已有           | core          | 表單資訊量較大           |
| Edit check             | 已有           | maintenance   | 需要 dirty state warning |
| Duplicate check        | 已有           | efficiency    | 很適合無法改 type 的情境 |
| Delete check           | 已有確認       | destructive   | 符合預期                 |
| Batch delete checks    | 已有確認       | bulk          | 符合預期                 |
| Ping config            | 已有           | measurement   | MVP core                 |
| TCP config             | 已有           | measurement   | MVP core                 |
| Traceroute config      | 已有           | measurement   | MVP core                 |
| Interval validation    | 已有           | form quality  | 方向正確                 |
| Label selector builder | 已有           | assignment    | 功能強，但學習成本高     |
| Advanced selector JSON | 已有           | power user    | 需要明確標示 advanced    |
| Selector preview       | 已有           | confidence    | 應顯示 match reason      |
| Assignments            | 已有           | system model  | 需要被更直覺地解釋       |

### 5.6 Labels

| 功能                     | 現況     | 分類           | UX 備註                             |
| ------------------------ | -------- | -------------- | ----------------------------------- |
| Label table              | 已有     | taxonomy       | 是 selector 心智模型的入口          |
| Search and key filter    | 已有     | operations     | 可用                                |
| Create label             | 已有     | setup          | 需能從 probe/check 流程 inline 進入 |
| Edit label               | 已有     | maintenance    | 可用                                |
| Delete label             | 已有確認 | destructive    | 符合預期                            |
| Probe/check usage counts | 已有     | impact preview | 很重要，應保留                      |

### 5.7 Insight And Results

| 功能                     | 現況           | 分類          | UX 備註                        |
| ------------------------ | -------------- | ------------- | ------------------------------ |
| Time range control       | 已有 URL state | analysis      | 支援分享分析狀態               |
| Relative/absolute window | 已有           | analysis      | 方向正確                       |
| Refresh interval         | 已有           | monitoring    | 適合 operational workflow      |
| Type filter              | 已有           | filtering     | Ping/TCP/Traceroute            |
| Group by check/probe     | 已有           | scoping       | 需要更清楚文案                 |
| Scope select             | 已有           | scoping       | 可用                           |
| Assignment multi-select  | 已有           | comparison    | 功能強，需降低理解成本         |
| Focus chips              | 已有           | state clarity | 對 deep link 很有幫助          |
| Ping insight/series      | 已有           | analysis      | MVP core                       |
| TCP insight/series       | 已有           | analysis      | MVP core                       |
| Traceroute insight/runs  | 已有           | route         | MVP core                       |
| Group topology           | 已有           | topology      | 應明確顯示資料範圍與 freshness |
| Multi-series comparison  | 已有           | comparison    | 需說明正在比較 assignments     |

### 5.8 Alerts

| 功能                                 | 現況           | 分類        | UX 備註                        |
| ------------------------------------ | -------------- | ----------- | ------------------------------ |
| Alert summary cards                  | 已有           | overview    | 可再接到 dashboard             |
| Incidents tab                        | 已有           | triage      | 缺少 lifecycle action          |
| Incident status filter               | 已有 URL state | triage      | 可用                           |
| Incident detail drawer               | 已有           | triage      | 缺少 Insight/Probe/Check links |
| Rules tab                            | 已有           | config      | 可用                           |
| Rule search/filter                   | 已有 URL state | operations  | 可用                           |
| Create/edit/delete rule              | 已有           | config      | rule scope UX 需要改善         |
| Ping/TCP metric threshold            | 已有           | alert logic | MVP core                       |
| Traceroute alert rules               | UI 禁用        | limitation  | 限制有被說明，方向正確         |
| Notifications tab                    | 已有           | routing     | 可用                           |
| Webhook/Slack/Discord/Telegram/Email | 已有           | routing     | 通知能力完整度高               |
| Test notification                    | 已有           | confidence  | 應保持 prominent               |
| Acknowledge/resolve incident         | 未見明確 UI    | gap         | triage loop 不完整             |

### 5.9 Collaboration And Account

| 功能                  | 現況         | 分類        | UX 備註                |
| --------------------- | ------------ | ----------- | ---------------------- |
| Member list           | 已有         | access      | 可用                   |
| Invite member         | owner/admin  | access      | 需要角色說明           |
| Pending invites       | 已有         | access      | 可用                   |
| Role update           | 已有         | permission  | 需要 role meaning      |
| Remove member         | 已有         | destructive | 目前缺少一致的確認流程 |
| Accept/reject invite  | 已有         | account     | 放在 Account 合理      |
| Profile display name  | 已有         | account     | 可用                   |
| Change email/password | feature flag | account     | 可用                   |
| Gravatar preview      | 已有         | identity    | 可用                   |
| Leave project         | 已有確認     | destructive | 符合預期               |
| Delete project        | 已有確認     | destructive | 符合預期               |

## 6. 使用者歷程

### 6.1 第一次使用：從註冊到第一筆結果

目標：讓使用者完成 project、probe、check、result 的 activation loop。

目前流程：

1. 使用者從 public site 或 app 進入。
2. 使用者註冊或登入。
3. 若沒有 project，進入 onboarding。
4. Onboarding 要求 project name 和 optional member emails。
5. 成功後導向 probe fleet / create probe。
6. 使用者輸入 probe name。
7. 使用者必須搜尋 location 或輸入 coordinates。
8. 系統建立 probe 並顯示 install command。
9. 使用者在 host 上執行 installer。
10. Wizard 等待 heartbeat。
11. Heartbeat 成功後使用者回到 Probes 或 Dashboard。
12. 使用者需要自行發現 Checks，建立第一個 check。
13. 使用者需要自行發現 Insight，查看 result。

目前問題：

- 流程在 heartbeat 後中斷，沒有直接引導到 first check。
- 建立 probe 前需要 location，可能阻礙只想快速驗證 agent 的使用者。
- Registration 頁面的 submit label 和實際動作不完全一致。
- Onboarding 的 terminal 風格和主要 dashboard 風格不一致。
- Dashboard 沒有 first-run checklist。

建議流程：

1. 註冊或登入。
2. 建立 project。
3. 建立第一個 probe，location 可略過或之後補。
4. 顯示 install command、heartbeat 狀態、troubleshooting。
5. Heartbeat 成功後，primary CTA 是「Create first check」。
6. Check wizard 可預填 Ping check，讓使用者輸入 target。
7. 第一筆 result 產生後，直接開啟 Insight 並選好 assignment。
8. Dashboard 從 setup checklist 切換成 operational overview。

### 6.2 回訪 Operator：事故 triage

目標：快速知道現在哪裡壞了，並找到原因。

目前流程：

1. Operator 打開 Dashboard。
2. Dashboard 顯示 probes online、active checks、map。
3. Operator 需要手動進入 Alerts。
4. 選擇 incident row。
5. Incident detail 顯示 what happened、timeline、notifications。
6. 若要分析原因，使用者需要自己去 Insight、Probes 或 Checks。

目前問題：

- Dashboard 沒有 open incidents、degraded checks、recent failures、stale probes。
- Incident detail 沒有「Open in Insight」、「Open check」、「Open probe」。
- 有 incident status filters，但 UI 沒有明確 acknowledge/resolve 操作。
- Alerts 與 Insight 之間沒有形成 triage loop。

建議流程：

1. Dashboard 直接顯示 open incidents、degraded assignments、stale probes。
2. Incident row 開啟 detail。
3. Detail 提供 direct actions：Inspect in Insight、Open check、Open probe。
4. 若 backend 支援，提供 Acknowledge / Resolve。
5. Insight 以 incident 的 probe、check、time window 預先篩好。

### 6.3 Probe Fleet Maintenance

目標：維持 probes 可靠、正確標籤化、可維護。

目前流程：

1. 開啟 Probes。
2. 使用 search/sort 或 map view。
3. 開啟 probe detail drawer。
4. 編輯 name、labels、location。
5. 儲存、rotate secret、copy service command 或 delete。

目前問題：

- Probe detail 不能 inline 建立 labels。
- Reinstall、upgrade、rotate secret 同時出現時資訊密度偏高。
- 沒有明確 enable/disable/drain 操作。
- Probe list 的 search/sort/view state 沒有 URL state。

建議流程：

1. 保留 list 和 map。
2. 增加 stale heartbeat、offline reason、last result recency。
3. Labels 區提供 inline create label。
4. 若 API 支援，提供 enable/disable control。
5. Service commands 改成清楚的 command cards。
6. Probes view/search/sort/selected probe 寫進 URL。

### 6.4 Measurement Setup

目標：建立 check，並確定它會指派到正確 probes。

目前流程：

1. 開啟 Checks。
2. New check。
3. 輸入 name、target、type、interval。
4. 設定 type-specific config。
5. 使用 all probes、label rules 或 advanced JSON 建立 selector。
6. Preview selector。
7. Save。

目前問題：

- Selector 很強，但 labels 和 assignments 的心智模型需要先建立。
- Labels 為空時，使用者不知道 selector 該怎麼用。
- Matched probes 只顯示 flat badges，沒有說明 match reason。
- Drawer 關閉時沒有 unsaved changes warning。
- Check type 建立後不能改，UI 需要說明原因與替代方案。

建議流程：

1. Labels 空時顯示「Use all probes」和「Create labels」兩條路。
2. Selector preview 顯示 count、probe、matched reason。
3. Dirty drawer 關閉前確認。
4. Check type lock 提示「Duplicate check to change type」。
5. Save 後提供「View results in Insight」。

### 6.5 Alert Routing

目標：設定 rule，測試 notification，並能從 incident 進入分析。

目前流程：

1. 建立 notification target。
2. Test notification。
3. 建立 alert rule。
4. Rule 可選 check type、metric、operator、threshold、timing。
5. Scope 使用 optional raw Probe ID / Check ID。
6. Threshold breach 後產生 incident。

目前問題：

- Rule scope 要輸入 raw UUID，不適合一般 operator。
- 沒有 notification 時，rule builder 沒有足夠強的引導。
- Incident lifecycle 不明確。
- Traceroute alerts 被禁用是合理的，但產品文案要同步。

建議流程：

1. 若沒有 notification，Alerts 應先引導建立並 test。
2. Rule scope 改用 searchable Probe / Check selects。
3. Rule preview 要包含 scope、condition、notification、cooldown。
4. Incident detail 提供 Insight deep link 和 lifecycle action 或清楚說明 system-managed。

### 6.6 Team Access

目標：安全邀請、調整角色、移除成員。

目前流程：

1. Owner/admin 開啟 Members。
2. 輸入 email、選 role、send invite。
3. Pending invites 出現在 table。
4. Role 可 inline 調整。
5. Member 可移除。
6. 被邀請者在 Account settings 接受或拒絕。

目前問題：

- Role 意義沒有在選擇時說明。
- Remove member 缺少和其他 destructive actions 一致的 confirmation。
- Project settings 與 Account settings 命名容易混淆。

建議流程：

1. Invite form 和 RoleSelect 顯示 owner/admin/editor/viewer 的權限摘要。
2. Remove member 使用 confirmation modal。
3. Sidebar 的 Settings 改 Project。
4. User menu 的 Settings 改 Account。

## 7. 現在的問題

### 7.1 P0：會阻礙 MVP 可用性的問題

| 問題                               | 影響                       | 建議                                         |
| ---------------------------------- | -------------------------- | -------------------------------------------- |
| First-run flow 在 heartbeat 後中斷 | 使用者可能看不到第一筆結果 | 加 setup checklist 和 Create first check CTA |
| Dashboard 太薄                     | 回訪 operator 無法 triage  | 加 incidents、stale probes、degraded checks  |
| Incident detail 沒有分析入口       | Triage 需要手動跳頁        | 加 Insight / Check / Probe deep links        |
| Alert rule scope 使用 raw IDs      | 容易設定錯                 | 改 searchable selects                        |
| Member removal 沒有 confirmation   | 可能誤移除成員             | 套用既有 confirm pattern                     |
| Probe creation 被 location 擋住    | 初次測試成本上升           | 允許 skip location 或之後補                  |

### 7.2 P1：心智模型混亂

| 問題                                  | 影響               | 建議                                           |
| ------------------------------------- | ------------------ | ---------------------------------------------- |
| Settings 同時代表 project/account     | 使用者容易進錯頁面 | Project 與 Account 分開命名                    |
| Labels 是 selector 前提，但太晚被理解 | Checks setup 抽象  | 在 Checks 和 Probes 中加入 contextual guidance |
| Advanced JSON selector 太早露出       | 初學者有壓力       | 保留但視覺上降權，標示為 Advanced              |
| Check type lock 缺少說明              | 使用者以為 UI 壞掉 | 說明建立後不可改，改用 Duplicate               |
| Onboarding 風格和 app 不一致          | 第一印象分裂       | 改成 dashboard-aligned setup wizard            |

### 7.3 P1：回饋與復原不足

| 問題                        | 影響                   | 建議                                  |
| --------------------------- | ---------------------- | ------------------------------------- |
| 某些 route loading 會空白   | 慢網路下像壞掉         | 使用 compact loading state            |
| Drawer 表單可無提示丟失修改 | 造成資料遺失           | dirty state + close confirmation      |
| Empty state 不一定有下一步  | 使用者不知道要做什麼   | 每個 empty state 加 primary action    |
| Copy/loading 狀態文案不一致 | 可近用性與狀態感不足   | 使用 Saving… / Copying… 和 aria-live  |
| 日期時間使用瀏覽器預設      | locale/timezone 不一致 | 建立一致的 Intl.DateTimeFormat helper |

### 7.4 P2：產品一致性問題

| 問題                                            | 影響                          | 建議                                     |
| ----------------------------------------------- | ----------------------------- | ---------------------------------------- |
| 文案提到 DNS，但目前 UI/API 未見完整 DNS check  | 信任風險                      | 移除 MVP DNS claim 或完整實作            |
| Probe enable/disable 沒有清楚 UI                | fleet maintenance 不完整      | 若 API 支援，加入 lifecycle control      |
| Traceroute topology 在不同 context 下呈現不一致 | route intelligence 感覺不完整 | 明確顯示 data scope 和 availability      |
| Probes/Labels list state 沒 URL state           | 難以分享與還原                | 將 search/filter/view/selection 寫進 URL |

## 8. 下一版 UX Requirements

### 8.1 Activation Requirements

- 新使用者在第一筆 result 出現前，永遠應該看得到 next action。
- Probe creation 不應要求完整 location metadata。
- Install step 必須包含 copy command、asset links、heartbeat state、troubleshooting。
- Heartbeat 成功後，primary CTA 應是 Create first check。
- First result 出現後，primary CTA 應是 View in Insight。

### 8.2 Dashboard Requirements

Dashboard 應回答：

- 目前 probes online、stale、offline 各有多少？
- active checks 有多少？
- 是否有 open incidents？
- 哪些 assignments 正在 failing 或 degraded？
- 最近 latency/loss/connectivity 是否有變化？
- 使用者下一步應該做什麼？

建議模組：

- Fleet summary cards。
- Open incidents compact table。
- Recent degraded assignments。
- Probe map。
- Setup checklist。
- Quick actions：New probe、New check、Add notification。

### 8.3 Forms And Drawers

- 所有欄位需要 label；會提交的 form controls 需要 meaningful name。
- validation error 必須靠近欄位，並說明如何修正。
- Dirty drawer 關閉前要確認。
- Destructive actions 需要 confirmation 或 undo。
- Advanced settings 不應成為第一眼主內容。
- Pending state 只在 request 開始後啟用。

### 8.4 Tables And Lists

- Table row 必須可用鍵盤操作。
- Wide operational data 可 horizontal scroll，不應壓縮到不可讀。
- Selection state 要明顯。
- Batch actions 要靠近 selected summary。
- Empty table 要有下一步。
- 重複使用的 filters、tabs、selected detail 應寫入 URL。

### 8.5 Insight

- Scope 必須永遠清楚：time、type、group、probe/check/assignment。
- Invalid shared URL 要說明哪個 object 不存在，並提供 clear action。
- Incident、Check、Probe 都應可以 deep-link 到 Insight。
- Multi-series comparison 要說明正在比較 assignments。
- Traceroute topology 要說明資料時間範圍與 freshness。

### 8.6 Alerts

- Incident lifecycle 要清楚：若是 system-managed，就明說；若可操作，就提供 acknowledge/resolve。
- Rule builder 不應要求一般使用者輸入 raw UUID。
- Notification setup 要突出 test delivery。
- Rule row 要能一眼看出 scope、condition、notification、cooldown、enabled state。
- Unsupported alert type 要明確禁用並說明原因。

### 8.7 Accessibility

- Icon-only buttons 需要 `aria-label`。
- Focus state 必須清楚，使用 `:focus-visible`。
- Loading、copy、validation 狀態要能被 screen reader 理解。
- 表單必須能 keyboard-only 完成。
- Onboarding/wizard 動畫需尊重 reduced motion。
- 日期、數字、時間格式應使用 `Intl.*` helper，而不是散落的 browser default。

## 9. MVP 範圍建議

### 9.1 保留在 MVP

- Auth、registration、login、demo mode。
- Project creation、project switcher。
- New probe wizard 和 heartbeat detection。
- Probe list、map、detail。
- Ping、TCP、Traceroute check definitions。
- Label-based selector builder 和 preview。
- Insight for Ping、TCP、Traceroute。
- Multi-assignment comparison。
- Alerts for Ping/TCP metrics。
- Webhook、Slack、Discord、Telegram、Email notifications。
- Members、roles、invites。
- Account settings。
- Docs 和 OpenAPI Explorer。

### 9.2 MVP 前必須修正

- First-run checklist。
- Heartbeat 成功後接到 first check。
- Dashboard operational summary。
- Incident to Insight deep links。
- Alert rule searchable scope selectors。
- Member removal confirmation。
- Dirty drawer protection。
- Project / Account 命名修正。
- Actionable empty states。

### 9.3 MVP 後再做

- DNS checks，除非完成 end-to-end 實作。
- 完整 NOC incident workflow。
- 複雜 RBAC policy editor。
- Billing、SSO、audit log UI。
- Custom dashboard builder。
- 進階 route analytics beyond current traceroute output。

## 10. 改版 Roadmap

### Phase 1：Activation Loop

- 將 onboarding 改成 dashboard-aligned setup wizard。
- Dashboard 加 setup checklist。
- Probe location 可 skip。
- Heartbeat 成功後導向 Create first check。
- First result 出現後導向 Insight。

### Phase 2：Operational Dashboard

- Dashboard 加 open incidents。
- 加 stale/offline probes。
- 加 degraded assignments / recent failures。
- 加 recent measurement trend cards。
- 保留 map，但不要讓 map 成為唯一核心內容。

### Phase 3：Triage Loop

- Incident detail 加 Insight、Check、Probe links。
- Insight 支援 incident prefilled scope。
- 若 backend 支援，加入 acknowledge/resolve。
- Alert rule builder 改 searchable selects。

### Phase 4：Configuration Confidence

- Drawer dirty state。
- Inline label creation。
- Selector match reasons。
- Role descriptions。
- Member removal confirmation。
- Settings 命名整理。

### Phase 5：Polish And Trust

- 統一日期時間格式。
- 統一 loading/copy 文案與 aria-live。
- Onboarding 視覺與 `design.md` 對齊。
- 移除或註明尚未支援的 DNS 能力。
- 做 keyboard 和 screen reader review。

## 11. 驗收標準

MVP UX 可以被視為可用，當：

- 新使用者不看外部文件，也能完成 project、probe、check、first result。
- Dashboard 在 setup 前顯示進度，在 setup 後顯示 network health。
- Operator 可以從 open incident 一鍵進入對應 Insight scope。
- 所有 destructive actions 都有 confirmation 或 undo。
- Alert rules 可以不用手動複製 raw UUID 完成 scope。
- Probe、Check、Alert、Insight detail URLs 可以分享且能處理 object missing。
- Empty states 都告訴使用者下一步。
- Onboarding 和 authenticated app 使用同一套 dashboard visual language。
- 核心流程可用 keyboard 操作。

## 12. UX Research Plan

建議找 5 至 7 位技術使用者做 moderated usability sessions，使用有 realistic data 的 demo environment。

測試任務：

1. 建立帳號和第一個 project。
2. 新增 probe，並說明什麼時候代表 online。
3. 建立 Ping check 並指派到 probe。
4. 找到該 probe/check 的最新結果。
5. 建立 notification target 並 test。
6. 建立 alert rule。
7. 調查 open incident，說明受影響的 probe/check。
8. 邀請 viewer，並說明 viewer 可以做什麼。

觀察重點：

- 使用者在哪裡停住。
- 哪些詞不理解。
- 是否看得到下一步。
- 是否信任 status labels。
- validation error 能否修正。
- 使用一次後能否解釋 labels 和 selectors。

## 13. Open Questions

- Probe location 應該 required、optional，還是可之後補？
- Backend 是否支援 probe enable/disable/drain？若支援，MVP 是否要暴露？
- Incidents 是 user-actionable，還是 system-managed records？
- DNS 是否先從 MVP 文案移除？
- owner/admin/editor/viewer 的實際權限矩陣是什麼？
- First check target 應由使用者輸入、產品建議，或提供範例？
- Labels 是否長期保留 top-level navigation，或未來併入 Checks/Probes？
- Docs 是否需要一頁「MVP limitations」降低預期落差？

## 14. 設計原則

- 介面要貼近工作本身：probes、checks、assignments、results、incidents。
- App 內優先使用密集、可掃描的 panels，不使用 marketing-style sections。
- Orange 用於 primary action 和 active state；blue 用於 workspace 和 secondary support；status colors 只表示語意狀態。
- 使用 `@netstamp/ui` 的 square dashboard primitives。
- 不要讓關鍵 operational state 藏在裝飾視覺後面。
- 不要求使用者記 raw IDs，能選物件就使用 searchable object picker。
- 每個頁面都要回答三件事：現在狀態是什麼、為什麼重要、下一步能做什麼。
