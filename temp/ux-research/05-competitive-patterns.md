# Subagent 5 - Competitive and Pattern Research

## 研究範圍與定位判讀

Netstamp 的 repo evidence 明確指向 **self-hosted network observability / network monitoring**，核心是 probes、checks、results、alerts、public status pages、self-hosting、API / OpenAPI，而不是文件蓋章、憑證、notary 或 RFC 3161 timestamping 產品。

主要 evidence：

- `README.md:4`：`Self-hosted network observability from probes you control.`
- `README.md:22-24`：Netstamp 是 open-source、self-hosted network monitoring app，讓使用者從自己的 machines、regions、labs、edge nodes、private infrastructure 觀察 reachability、latency、packet loss、routes、probe health。
- `README.md:37-46`：功能包含 self-hosted controller、React web app、Go API、lightweight probe agents、Ping/TCP/traceroute checks、dashboards、alert rules/incidents/notifications、Webhook/Discord/Telegram notifications、Docker Compose、OpenAPI。
- `design.md:3`：產品類別是 `network observability and developer infrastructure`。
- `docs/src/content/docs/guides/probe-operations.mdx:10-27`：probe 被定義為 network viewpoint，checks 與 result flow 都以 latency、DNS、route、API result 為核心。
- `docs/alerting-v1-plan.md:5-22` 與 `docs/alerting-v1-plan.md:943-989`：alerting flow、incident、notification outbox 與 alerts console 規劃已存在。

因此，本研究把競品與 UX pattern 聚焦在：

- Observability / synthetic monitoring
- Uptime monitoring
- Probe agent / private location onboarding
- Status page / incident communication
- Alert integration / notification channel
- Self-hosted monitoring / API-first operations

文件驗證、proof-of-existence、trusted timestamping 僅列為「原始需求提及但 codebase 無證據的相鄰參考」，不可作為 Netstamp 目前產品方向的直接競品。

## 來源使用說明

本文件使用網路查詢，優先採用官方文件、官方 GitHub README、官方產品頁或主要專案網站。若某觀察只來自相鄰產品而非 Netstamp repo evidence，confidence 會降為中或低。

## 競品與相鄰產品觀察

| 競品 / 相鄰產品 | 來源 URL | 觀察 | Netstamp 建議 | Confidence |
| --- | --- | --- | --- | --- |
| Grafana Cloud Synthetic Monitoring | https://grafana.com/docs/grafana-cloud/testing/synthetic-monitoring/ | 將 synthetic monitoring 定義為 black-box monitoring：從全球 probe locations 持續執行 checks，評估 availability、performance、correctness；每次 check execution 對應一個 location 的單次使用者模擬。 | Netstamp 的首頁、docs 與 app 空狀態應強化 `probe = viewpoint`、`check = scheduled network measurement`、`result = evidence from that viewpoint` 的 mental model。 | 高 |
| Grafana private probes | https://grafana.com/docs/grafana-cloud/testing/synthetic-monitoring/set-up/set-up-private-probes/ | Private probes 是 blackbox agents，執行 configured checks、送出 metrics/logs；每個 private probe 有自己的 authentication token，且可用 Docker / Kubernetes 部署。 | Netstamp 新增 probe wizard 應把 token/secret、controller URL、install command、heartbeat confirmation、upgrade/uninstall command 串成單一 activation flow。 | 高 |
| Grafana check types / alerts | https://grafana.com/docs/grafana-cloud/testing/synthetic-monitoring/create-checks/checks/ | Checks 可跑在 public/private probes 上，結果保存為 Prometheus metrics / Loki logs，並可配置 Grafana alerts 與 incident management。 | Netstamp 目前支援 Ping/TCP/traceroute；不要在 UI copy 裡暗示已支援未實作的 HTTP/DNS check。可把 future check types 放在 docs roadmap 或 disabled template。 | 高 |
| Grafana results dashboard | https://grafana.com/docs/grafana-cloud/testing/synthetic-monitoring/analyze-results/ | 結果頁提供 overview、individual check results、trend、filters by region/probe/check type/alert status/labels，並可 drill down 到 individual check dashboard。 | Netstamp Insight / Result page 應提供 project-wide summary -> check/probe filters -> individual assignment/run detail 的 drilldown，不要只給單一 chart。 | 高 |
| Checkly private locations | https://www.checklyhq.com/docs/platform/private-locations/overview/ | Private Location 由使用者部署 lightweight Checkly Agent；create flow 會給 API key、Docker run command、刷新頁面後看到 running agents count，再選擇該 private location 跑 checks。 | Netstamp probe onboarding 應在 install command 後明確顯示 `waiting for heartbeat`、`agent count / last seen`、`next: create check`，並在 10 分鐘無 heartbeat 時給診斷。 | 高 |
| Checkly alert channels | https://www.checklyhq.com/docs/communicate/alerts/channels/ | Alert channels 支援 email、Slack、webhooks、PagerDuty、SMS、自訂整合，並明確描述 failure/degrade/recover notification。 | Netstamp notification editor 應把 channel type、trigger events、test delivery、secret masking、recovery notification 清楚放在同一個模型中。 | 高 |
| Checkly result details | https://www.checklyhq.com/docs/concepts/results/ | Results 不只是 pass/fail，包含 performance metrics、error details、screenshots、network traces 與 telemetry；overview 有 summary、time bars、sidepanel drilldown。 | Netstamp 的 ping/TCP/traceroute result detail 應顯示 status、started/finished、duration、error code/message、probe/check metadata、raw hop/connection evidence；用 time ribbon 做快速定位。 | 高 |
| Datadog private locations | https://docs.datadoghq.com/getting_started/synthetics/private_location/ | Private locations 用於 internal-facing apps/private URLs、自訂 mission-critical locations、release 前比較內外部性能；worker 可以是 Docker container 或 Windows service。 | Netstamp 應把 `private infrastructure / lab / edge / provider boundary` 寫成常見 probe placement templates，降低新手不知道 probe 放哪裡的問題。 | 高 |
| Datadog private location monitoring | https://docs.datadoghq.com/synthetics/platform/private_locations/monitoring/ | Private location tab 顯示 reporting status、monitor status、worker health、metadata、image version；預設 monitors 包含 stopped reporting、underprovisioned、outdated image、poll too long。 | Netstamp Probe detail/fleet 應顯示 heartbeat stale、agent version、assignment polling delay、worker capacity/queue、upgrade needed；probe health 不應只顯示 online/offline。 | 高 |
| Datadog synthetic test results | https://docs.datadoghq.com/synthetics/browser_tests/test_results/ | Test result detail 顯示 latest failed vs recent successful runs，包含 status、duration、location、run type、screenshots/actions、page performance、resources、backend traces。 | Netstamp 沒有 browser checks，但可借用 result detail IA：latest failed run、recent successful baseline、probe location、run type、duration、error evidence、相關 trace/log link。 | 中 |
| Uptime Kuma | https://github.com/louislam/uptime-kuma | 自稱 easy-to-use self-hosted monitoring tool；README 直接展示 live demo、dashboard screenshot、status page screenshot、Docker Compose 安裝；支援 HTTP/TCP/Ping/DNS/Push 等 monitors、多 status pages、ping chart、certificate info。 | Netstamp public site 應展示真實 dashboard/status/alert/insight screenshot 或 demo instance，而不是抽象插圖；self-host quickstart 要短且可複製。 | 高 |
| Uptime Kuma notifications | https://github.com/louislam/uptime-kuma/wiki/Notification-Methods | Uptime Kuma 支援大量通知，透過 native providers 與 Apprise 擴展到 78+ / 90+ channels。 | Netstamp V1 不必追求 channel 數量，但要把 webhook、Discord、Telegram、email 的 test path 做紮實；未來可考慮 Apprise-like generic channel 或 payload template。 | 中 |
| Better Stack uptime | https://betterstack.com/docs/uptime/monitoring-start/ | 新手 flow 是「create first monitor -> basic alerting」；HTTP monitor failure 會建立 incident，先 alert on-call，未回應再 alert team。 | Netstamp onboarding 不應停在建立 probe；應把 first successful result 與 first alert rule template 納入 activation checklist。 | 中高 |
| Better Stack incident detail | https://betterstack.com/docs/uptime/incident-details/ | Incident detail 可包含 downtime screenshots，但也清楚說明 timeout/empty response 時不會有 screenshot，且 screenshot 可能有 delay。 | Netstamp result/incident detail 要誠實呈現 evidence 限制，例如 traceroute partial、probe timeout、no response、stale heartbeat，避免過度確定。 | 高 |
| Better Stack status subscriptions | https://betterstack.com/docs/uptime/subscribing-to-status-updates/ | Status page 支援 email、webhook、RSS、JSON API；訂閱者可選全部 services 或 specific components；新訂閱需要 confirmation。 | Netstamp Public Status Pages 若要加入訂閱，應優先支援 component-scoped subscription 與 JSON/RSS read API；先不必實作 SMS。 | 中 |
| Atlassian Statuspage | https://support.atlassian.com/statuspage/docs/read-the-statuspage-user-guide/ | Statuspage 以 components、component groups、incidents、scheduled maintenance、incident templates、postmortems、system metrics、subscribers 建構透明溝通流程。 | Netstamp status page 應把 probes/check assignment groups 映射成 public components，並提供 open/resolved incidents、maintenance/manual note、system metric/history。 | 高 |
| Atlassian subscriber controls | https://support.atlassian.com/statuspage/docs/enable-subscribers/ | Subscriber notification 有明確事件規則；SMS 只在 incident/maintenance created、resolved、maintenance begins 時發送，email 可發送所有 updates；SMS double opt-in/rate limit。 | Netstamp 未來若做 public subscribers，要先定義哪些事件會通知，避免每次 metric 更新都推送造成噪音。 | 中 |
| Prometheus Blackbox Exporter | https://github.com/prometheus/blackbox_exporter | 支援 HTTP/HTTPS/DNS/TCP/ICMP/gRPC blackbox probing；`/probe?target=...&module=...` 回傳 `probe_success` 與 timing metrics；multi-target exporter pattern 把 module 與 target 分離。 | Netstamp check model 可持續保持 `type/config + target + assigned probes` 分離；對進階使用者提供 raw metrics/API 與 debug output 會增加可驗證性。 | 高 |
| Upptime | https://github.com/upptime/upptime | 用 GitHub Actions、Issues、Pages 做 open-source uptime monitor/status page；response time data 會 commit 到 git，downtime 用 Issues 開關，status page 顯示 uptime、response time、incident history。 | Netstamp self-host 路線可借用 `audit trail / generated public status / historical evidence` 的信任敘事，但不要照搬 GitHub Actions 架構。 | 中 |
| Upptime site | https://upptime.js.org/ | 產品頁用一句話說明 GitHub-powered uptime monitor and status page，強調每 5 分鐘監控、version-controlled response time stats、modern status page、Slack/Telegram/webhook 通知。 | Netstamp 首頁可把價值濃縮為「from probes you control」「network path evidence」「self-hosted controller」「OpenAPI」，並用真實狀態頁示例支撐。 | 中 |
| OneUptime | https://github.com/oneuptime/oneuptime | Open-source observability platform，將 uptime monitoring、status pages、incident management、on-call、logs、APM、error tracking 放在一體化平台中。 | Netstamp 應避免過早追求全平台；更適合把定位收斂為 `network measurements from controlled probes`，只做必要 status/incident/notification。 | 高 |
| Healthchecks.io | https://healthchecks.io/docs/ | Healthchecks.io 是 dead man's switch：每個 check 有 unique ping URL、schedule、integrations，狀態包含 new/up/late/down/paused；明確說它不是用 HTTP probe 做 website uptime 的工具。 | Netstamp 可借用 heartbeat/late/down/paused 狀態與 `grace time` pattern 來表示 probe health；但不要把 cron monitoring 當直接競品。 | 中 |
| Healthchecks badges | https://healthchecks.io/docs/badges/ | Status badges 可嵌入 README、internal dashboard、public status page；badge URL public but hard-to-guess，只揭露 aggregate status，不可反推 ping URL。 | Netstamp 可提供 public/private badge for status page、probe group、check group；badge 必須避免洩漏 probe secret、target inventory 或 internal hostnames。 | 中 |
| OpenTimestamps | https://opentimestamps.org/ | 定義可建立、稍後獨立驗證的 provable timestamps；證明某資料在某時間點前已存在。 | 僅作為「相鄰但非 Netstamp evidence」參考。若有人因 Netstamp 名稱期待文件存證，應用 copy 明確排除或另開 discovery，不要混入現有 UX。 | 低 |
| RFC 3161 Time-Stamp Protocol | https://www.ietf.org/rfc/rfc3161.txt | RFC 3161 描述向 Time Stamping Authority 發 request、回應 TimeStampToken 的格式與 TSA 操作要求。 | 與 Netstamp network observability 不同；除非產品策略改變，不應把 TSA、document hash、signature verification 放進目前 UX research board。 | 低 |

## Common UX Patterns

| Pattern | 來源 URL | 觀察 | Netstamp 建議 | Confidence |
| --- | --- | --- | --- | --- |
| Viewpoint-first monitoring | https://grafana.com/docs/grafana-cloud/testing/synthetic-monitoring/ | Synthetic monitoring 常把 location/probe 視為觀測視角，結果是從某地點對 target 的 evidence。 | 在 Dashboard、Probes、Checks、Insight、Status 裡一致使用 viewpoint 語言：`from probe`、`from location`、`from network`。 | 高 |
| Private agent / private location | https://www.checklyhq.com/docs/platform/private-locations/overview/ | Private location onboarding 通常包含建立 location、顯示 one-time key、copy Docker command、確認 running agents、把 checks 指派到該 location。 | Netstamp `New Probe` wizard 應有 steps：Create metadata -> copy install command -> wait heartbeat -> assign/create check -> view first result。 | 高 |
| Check type + target + interval + probes | https://github.com/prometheus/blackbox_exporter | Blackbox exporter 的 module/target pattern 讓「如何測」與「測誰」分開；Grafana/Checkly 也將 check type、frequency、locations 分開。 | Netstamp check editor 要清楚分區：type-specific config、target、interval、probe selector、assignment preview。 | 高 |
| Label/selector based assignment | https://grafana.com/docs/grafana-cloud/testing/synthetic-monitoring/analyze-results/ | Grafana results 可用 labels、probe、check type 篩選；大規模 monitoring 需要 label mental model。 | Selector builder 應保留 preview matched probes，並在儲存前顯示將新增/移除的 effective assignments。 | 高 |
| Result overview -> run drilldown | https://www.checklyhq.com/docs/concepts/results/ | Checkly 的 result overview 有 summary、chart、sidepanel；使用者可從時間區塊 drill into run detail。 | Insight 不只顯示 chart，還要有 selected time bucket、affected probes、latest failed/successful run、raw evidence panel。 | 高 |
| Built-in alert channel testing | https://grafana.com/docs/grafana/latest/alerting/configure-notifications/manage-contact-points/ | Grafana contact point 可 send test notification；Checkly 也把 alert channels 作為獨立設定。 | Netstamp notifications 應強化 `Test`、最近 test result、delivery failure reason、SMTP configured status。 | 高 |
| Incident state machine | https://support.atlassian.com/statuspage/docs/read-the-statuspage-user-guide/ | Status/incident 產品會明確呈現 investigating、updates、resolved、maintenance、postmortem 等 lifecycle。 | Netstamp incident detail 應顯示 open/acknowledged/resolved、opened/resolved timestamps、last evaluation、manual resolve/ack action、通知紀錄。 | 高 |
| Public status components | https://support.atlassian.com/statuspage/docs/read-the-statuspage-user-guide/ | Statuspage 用 components / component groups 表達系統切面，並可顯示 system metrics。 | Netstamp public status page 可以把 assignment groups/check groups 映射成 components，避免外部讀者看到過多 probe/check technical IDs。 | 高 |
| Subscriber preferences | https://betterstack.com/docs/uptime/subscribing-to-status-updates/ | Better Stack 讓 subscribers 選擇全部服務或 specific components，並提供 email/webhook/RSS/JSON API。 | Netstamp 若實作訂閱，先做 component-scoped email/webhook/RSS，不要讓所有 status update 都無差別推送。 | 中 |
| Self-host trust via short install path | https://github.com/louislam/uptime-kuma | Uptime Kuma README 把 Docker Compose 安裝與 live demo 放在高可見位置。 | Netstamp README/docs/landing 應保持 quickstart 短、copyable，並提供安全暴露/HTTPS/secret checklist。 | 高 |
| API-first discoverability | https://docs.datadoghq.com/api/latest/synthetics/ | Observability 工具通常提供 API docs / generated schemas 供 automation 使用。 | Netstamp 已有 OpenAPI；UX 應在 resource detail 提供 `View API`、copy ID、copy curl、generatedAt/schema version。 | 中高 |

## Trust-building Patterns

| Pattern | 來源 URL | 觀察 | Netstamp 建議 | Confidence |
| --- | --- | --- | --- | --- |
| 明確說明資料從哪裡觀測 | https://grafana.com/docs/grafana-cloud/testing/synthetic-monitoring/set-up/set-up-private-probes/ | Grafana 說 private probes 只對使用者可用，且只把資料寫入自己的 Grafana Cloud account。 | Netstamp 需在 docs 與 onboarding 說清楚 probe 會送哪些資料、送到哪個 controller、哪些 target metadata 會被保存。 | 高 |
| Probe/agent health 不只 online/offline | https://docs.datadoghq.com/synthetics/platform/private_locations/monitoring/ | Datadog 顯示 reporting status、monitor status、worker version、underprovisioned/outdated/poll too long。 | Netstamp Probe fleet 加入 `last seen`、`agent version`、`assignment poll age`、`result submit age`、`stale reason`、`upgrade available`。 | 高 |
| Alert 減噪與可解釋 threshold | https://betterstack.com/docs/uptime/monitoring-start/ | Better Stack first monitor 直接連到 incident/on-call escalation，說明 failure 後誰會被通知。 | Netstamp alert builder 應用 templates 顯示「最近 5 分鐘、samples >= 3、loss >= 20%」會在何時 fire/recover，並預估 noise risk。 | 高 |
| Notification test + secret masking | https://grafana.com/docs/grafana/latest/alerting/configure-notifications/manage-contact-points/ | Contact point test 是信任關鍵；Checkly webhook 也給 payload template control。 | 所有 webhook/Slack/Discord/Telegram URL 必須 masked；test 成功與 incident delivery 成功要分開顯示。 | 高 |
| Evidence limits 透明化 | https://betterstack.com/docs/uptime/incident-details/ | Better Stack 清楚說明 screenshots 不是每種 failure 都有，也可能有 delay。 | Netstamp 對 traceroute partial、timeout、probe stale、missing samples、clock skew、DNS resolve failure 應顯示限制，不要只給紅色 failed。 | 高 |
| Public status 透明溝通 | https://support.atlassian.com/statuspage/docs/read-the-statuspage-user-guide/ | Statuspage 把 incident updates、scheduled maintenance、postmortems、system metrics 放在客戶溝通面。 | Netstamp public status 頁需有 generated timestamp、affected components、open/recent incidents、metric history、last updated by system/manual note。 | 高 |
| Open source / self-host 可審查 | https://github.com/upptime/upptime | Upptime 用 GitHub Actions/Issues/Pages 與 git history 建立低成本可審查性。 | Netstamp 可強調 Apache 2.0、self-host Docker Compose、OpenAPI、DB ownership；也可提供 audit/log export 作為未來能力。 | 中 |
| Badge / embed 不洩漏秘密 | https://healthchecks.io/docs/badges/ | Healthchecks badge URL 只暴露 aggregate status，且不可反推 ping URL。 | Netstamp status badge 不應包含 internal target、probe secret、private hostname；public slug 與 private probe metadata 應分層。 | 中 |

## Onboarding Patterns

| Pattern | 來源 URL | 觀察 | Netstamp 建議 | Confidence |
| --- | --- | --- | --- | --- |
| First useful outcome 比 first object 更重要 | https://betterstack.com/docs/uptime/monitoring-start/ | Better Stack 的 first monitor onboarding 同時建立 basic alerting，直接到 incident readiness。 | Netstamp onboarding 要走到 first successful result，而不只是 project/probe created。Activation checklist：project -> probe online -> check assigned -> first result -> optional alert. | 高 |
| One-time key + copy command + refresh confirmation | https://www.checklyhq.com/docs/platform/private-locations/overview/ | Checkly 顯示 API key、Docker command，刷新後看到 running agents count。 | Probe secret 只顯示一次；copy install command 後 UI 自動 polling；成功後顯示 exact probe name/location/last heartbeat。 | 高 |
| Requirements upfront | https://www.checklyhq.com/docs/platform/private-locations/overview/ | Checkly 在 Private Location 前列出 container runtime、outbound access、network access、permission requirements。 | Netstamp install drawer 應在命令前列出 Linux/systemd/root、controller URL reachable、outbound HTTPS、arch、time sync、firewall notes。 | 高 |
| Private network use cases as templates | https://docs.datadoghq.com/getting_started/synthetics/private_location/ | Datadog 提供 internal apps、mission-critical location、release before production、inside/outside comparison 等使用情境。 | New Probe modal 可提供 placement presets：home ISP、edge VPS、lab network、private server、provider boundary、customer region。 | 中高 |
| CLI / IaC path for engineers | https://www.checklyhq.com/docs/detect/testing/creating-your-first-test/ | Checkly 把 CLI/code configuration 作為進階路徑，讓 monitoring as code 使用者不被 UI 限制。 | Netstamp docs/OpenAPI 可提供 `curl create probe/check/alert`、`docker compose`、`netstamp-agent service install` snippets，並從 UI 複製 resource IDs。 | 中高 |
| Live demo / sample status | https://github.com/louislam/uptime-kuma | Uptime Kuma 有 temporary live demo 與 screenshots。 | Netstamp 若有 demo mode，公開 landing 應連到 read-only demo；demo banner 要說明資料是範例或唯讀。 | 中 |

## Verification / Result Page Patterns

這裡的 verification 指 **network check/result verification**：使用者要驗證某個 probe/check/incident 結論是否可信，不是 document verification。

| Pattern | 來源 URL | 觀察 | Netstamp 建議 | Confidence |
| --- | --- | --- | --- | --- |
| Summary metrics + filters | https://grafana.com/docs/grafana-cloud/testing/synthetic-monitoring/analyze-results/ | Grafana main dashboard 顯示 check status summary、error rate by location、latency comparisons，並可用 region/probes/check types/alert status/custom labels 篩選。 | Insight overview 應提供 availability/reachability、packet loss、average/max latency、TCP failure/connect time、probe count、sample count，並支援 labels/probes/check types filters。 | 高 |
| Time ribbon / bar drilldown | https://www.checklyhq.com/docs/concepts/results/ | Checkly monitoring results chart 以 bar 表示時間段，hover/click 可篩出 right sidepanel 的 runs。 | Netstamp 可在每個 check detail 加入 compact result ribbon：綠/黃/紅/灰 bars；click 後顯示該 bucket 的 probe runs。 | 高 |
| Failed vs successful comparison | https://docs.datadoghq.com/synthetics/browser_tests/test_results/ | Datadog Sample Runs 可比較 latest failed runs 與 recent successful runs。 | Incident detail 應顯示 failure run 與最近 successful baseline：same probe/check/target 的 latency/loss/path diff。 | 中高 |
| Raw evidence panel | https://github.com/prometheus/blackbox_exporter | Blackbox exporter 的 probe endpoint 可加 `debug=true`，並回傳 probe_success 與 timing metrics。 | Netstamp result detail 加入 raw JSON/curl/API link、error code、probe runtime logs excerpt 或 debug payload，提升工程使用者信任。 | 中高 |
| Traceroute hop diagnostics | https://betterstack.com/ | Better Stack 產品頁提及 timeout 會提供 traceroute/MTR outputs 來理解 connection/request timeout。 | Netstamp traceroute result 應將 hop loss、RTT、hostname/IP、ASN/provider、path hash/path diff 分層顯示；unknown hop 不要誤判為 failure。 | 中 |
| Incident timeline | https://betterstack.com/docs/uptime/api/list-of-incident-timeline-events/ | Better Stack incident timeline API 把 incident events 作為完整列表。 | Netstamp incident detail 應有 timeline：opened、evaluated firing、notification queued/delivered/failed、acknowledged、resolved。 | 中高 |
| Public status generatedAt | https://github.com/upptime/upptime | Upptime status page 顯示 uptime、response time、incident history，資料由 GitHub Actions/commits 生成。 | Public status page 應清楚顯示 generatedAt / last updated / data source，避免外部讀者不知道狀態新鮮度。 | 中 |
| Status page components + metrics | https://support.atlassian.com/statuspage/docs/read-the-statuspage-user-guide/ | Statuspage 支援 components、system metrics、incidents、maintenance。 | Netstamp public status 的每個 component 應對應 check group 或 assignment group，並顯示 open incident、latest status、chart mode/range。 | 高 |

## Anti-patterns

| Anti-pattern | 來源 URL | 觀察 | Netstamp 建議 | Confidence |
| --- | --- | --- | --- | --- |
| 產品名暗示文件 timestamping | https://opentimestamps.org/ | OpenTimestamps / RFC 3161 是 proof-of-existence / TSA 領域，與目前 Netstamp network observability evidence 不同。 | 所有外部文案第一屏就說 `network observability from probes you control`，不要只靠 Netstamp 名稱；document stamp 放入 no-evidence/adjacent 區。 | 高 |
| 宣稱未實作 check type | https://grafana.com/docs/grafana-cloud/testing/synthetic-monitoring/create-checks/checks/ | 競品通常清楚列出 check types；Netstamp repo 目前有 Ping/TCP/traceroute evidence，但 docs copy 有 DNS 字樣。 | UI/landing/docs 應以 Ping/TCP/traceroute 為準；DNS/HTTP 若未完成，放 roadmap 或 disabled label，不要作為已支援賣點。 | 高 |
| 把 probe offline 誤報成 target down | https://docs.datadoghq.com/synthetics/platform/private_locations/monitoring/ | Datadog 將 private location worker stopped reporting、underprovisioned、outdated、poll too long 獨立監控。 | Netstamp status/result 應清楚區分 target failure、probe stale、agent offline、insufficient samples、assignment not running。 | 高 |
| 沒有 assignment preview 的 selector editor | https://github.com/prometheus/blackbox_exporter | Multi-target probing 一旦 module/target/selector 配錯會造成覆蓋缺口。 | Selector builder 必須保留 preview，並顯示 matched probes count、excluded probes、why not matched。 | 高 |
| Notification channel 無 test 或不遮 secret | https://grafana.com/docs/grafana/latest/alerting/configure-notifications/manage-contact-points/ | Contact point test 是主流告警 UX；alerting secrets 需要小心處理。 | Webhook/Slack/Discord URL 與 Telegram token 不應完整顯示或記錄；notification editor 要有 test result 與 failure reason。 | 高 |
| Status page 自動公開內部細節 | https://healthchecks.io/docs/badges/ | Healthchecks badge 刻意只揭露 aggregate status，避免反推 secret URL。 | Public status slug、component label、badge 不應包含 internal hostname、private IP、probe IDs、secret、full target inventory。 | 中高 |
| 結果頁只給紅綠燈 | https://www.checklyhq.com/docs/concepts/results/ | Checkly 強調 results 是完整 telemetry，不只是 pass/fail。 | Netstamp result UI 應顯示 sample count、time window、probe/check context、error detail、latency/loss/path evidence。 | 高 |
| Alert builder 只提供自由 JSON | https://grafana.com/docs/grafana/latest/alerting/fundamentals/ | Grafana alert rule 會連到 contact points / notification policies；Netstamp 內部 `docs/alerting-v1-plan.md` 也顯示 alert condition 有 AST/metric threshold。若 UI 暴露太多 JSON，會提高錯配風險。 | Rule editor 應用 metric threshold builder + templates + advanced JSON fallback；儲存前顯示 natural-language summary。 | 高 |
| Public status subscriptions 沒有事件規則 | https://support.atlassian.com/statuspage/docs/enable-subscribers/ | Statuspage 對 SMS/email update events 有清楚規則與 opt-in/rate limit。 | 未來若加入 subscribers，要定義哪些 incident/status events 會通知，以及 component subscription 的選擇範圍。 | 中 |

## Netstamp 可借鑑之處

| Area | 來源 URL | 觀察 | Netstamp 建議 | Confidence |
| --- | --- | --- | --- | --- |
| Product positioning | https://grafana.com/docs/grafana-cloud/testing/synthetic-monitoring/ | Synthetic monitoring 的共同語言是 availability、performance、correctness、probe locations。 | 對外定位：`Self-hosted network observability from probes you control`；副標補 `latency, packet loss, TCP reachability, traceroute path evidence`。 | 高 |
| First-run activation | https://www.checklyhq.com/docs/platform/private-locations/overview/ | Private agent onboarding 的關鍵是 key -> command -> running agent confirmation。 | 建立「first probe checklist」：copy command、heartbeat received、agent version, first assignment, first result；失敗時顯示 diagnostics。 | 高 |
| Probe health surface | https://docs.datadoghq.com/synthetics/platform/private_locations/monitoring/ | Private location 有 health/metadata side panel 與 default monitors。 | Probe detail 增加 Health tab：heartbeat、poll delay、last submit、agent version、OS/arch、active assignments、latest errors、upgrade command。 | 高 |
| Alert templates | https://betterstack.com/docs/uptime/monitoring-start/ | First monitor 直接連 basic alerting 與 incident。 | Alert rule builder 預設 templates：high packet loss、TCP failure percent、high RTT、no samples/stale probe；每個 template 顯示 window/min samples。 | 高 |
| Notification channel model | https://www.checklyhq.com/docs/communicate/alerts/channels/ | Alert channels 是可重用目的地，支援 failure/degrade/recover。 | Notifications tab 可顯示 channel health、last test、last incident delivery、used by rules count、masked config。 | 高 |
| Result explainability | https://www.checklyhq.com/docs/concepts/results/ | Results detail 要回答 what happened、where、how well、why failed。 | 結果頁的最低內容：status、probe/check/target、started/finished、duration、samples、metric values、error code/message、raw API link。 | 高 |
| Traceroute storytelling | https://betterstack.com/ | Traceroute/MTR 是 timeout 排查 evidence。 | Traceroute topology 增加 path diff、new/missing hop、RTT spike、partial/timeout explanation；不要只畫 map。 | 中 |
| Public status communication | https://support.atlassian.com/statuspage/docs/read-the-statuspage-user-guide/ | Components、incidents、maintenance、postmortems、metrics 是 status communication 的骨架。 | Netstamp status page V1 可先做好 components + open/recent incidents + metrics；maintenance/postmortem/subscriber 作後續。 | 高 |
| Open-source trust | https://github.com/louislam/uptime-kuma | Uptime Kuma 用 GitHub README、screenshots、demo、Docker command 快速建立信任。 | Netstamp README/landing 應保持 screenshots、demo、self-host quickstart、license、OpenAPI、security exposure checklist 在高可見位置。 | 高 |
| API / automation | https://github.com/prometheus/blackbox_exporter | Engineering audience 需要 raw metrics/debug/API 來驗證工具本身。 | 每個 probe/check/result/incident detail 提供 copy ID、copy API path、OpenAPI link、curl snippet。 | 中高 |
| Badge / embed | https://healthchecks.io/docs/badges/ | Badge 是低成本信任輸出，但要控制資訊揭露。 | Status Pages 可加 badge endpoint；只顯示 aggregate status 與 label，避免完整 target/probe 曝光。 | 中 |

## 相鄰參考：Document Verification / Timestamping

這一類是原始需求可能提及、但 codebase 無直接 evidence 的相鄰領域。它不應主導 Netstamp 的競品選擇或 UX pattern。

| 類型 | 來源 URL | 觀察 | Netstamp 建議 | Confidence |
| --- | --- | --- | --- | --- |
| OpenTimestamps | https://opentimestamps.org/ | 重點是 proof that some data existed prior to a point in time，並可 later independently verify。 | 可作為「Netstamp 名稱可能造成誤解」的風險例子；目前不建議納入 product UX。 | 低 |
| OpenTimestamps CLI | https://github.com/opentimestamps/opentimestamps-client | CLI 用 `ots stamp` / `ots verify` 產生與驗證 timestamp proofs；verify 需要 Bitcoin Core node 或相關驗證條件。 | 若未來真的探索 document timestamping，要另起 discovery，包含 file hash、proof file、verification result、trust root；不要和 probe/check/result 混在同一 IA。 | 低 |
| RFC 3161 TSA | https://www.ietf.org/rfc/rfc3161.txt | TSA request/response、TimeStampToken、trusted time source、hash imprint、nonce 等是 PKI/TSA 領域。 | 與 Netstamp current codebase 無直接關聯；除非產品策略明確改變，僅保留為 no-evidence assumption。 | 低 |

## 優先級建議

1. **先修正定位與 copy 一致性**  
   聚焦 `network observability from probes you control`、Ping/TCP/traceroute、self-hosted controller、OpenAPI。避免 DNS/HTTP/document timestamping 等無實作 evidence 的暗示。

2. **把 first-run onboarding 做成可完成的 activation flow**  
   Project created 不是成功；`probe online + first check assigned + first result visible` 才是成功。這是競品 private location onboarding 最一致的 pattern。

3. **建立 Probe Health 可信面板**  
   借鏡 Datadog private location monitoring：last heartbeat、worker/agent version、poll delay、result submit age、stale reason、capacity/queue、upgrade needed。

4. **讓結果頁能被工程使用者驗證**  
   借鏡 Grafana/Checkly/Datadog：summary metrics、filters、time ribbon、run detail、raw evidence、failed vs successful comparison。Netstamp 的差異化應是 network path / packet loss / TCP reachability / traceroute evidence。

5. **告警要可測試、可解釋、可減噪**  
   Alert rule editor 用 templates + natural language summary + threshold preview；notification channels 必須有 test、masked secrets、last delivery/failure reason。

6. **Public Status 先做好 component + incident + metric history**  
   不要急著做完整 subscriber/on-call/postmortem。先讓外部讀者能理解哪些 service/component 受影響、資料更新時間、開啟事件與最近歷史。

7. **把 document verification 明確標成非範圍或待驗證假設**  
   若後續 stakeholder 真要求 document timestamping，應另開產品 discovery；現有 Netstamp evidence 不支持把它納入核心競品研究。

## 可放入 FigJam 的高價值 sticky

- `Netstamp is network observability, not document timestamping` - source: `README.md:4`, `design.md:3`
- `Probe = network viewpoint; check = scheduled measurement; result = evidence` - source: Grafana Synthetic Monitoring, Netstamp probe ops docs
- `Activation requires first result, not just first probe` - source: Checkly Private Locations, Better Stack first monitor
- `Probe health needs heartbeat + version + poll delay + result submit age` - source: Datadog Private Location Monitoring
- `Result page must show why, where, and from which probe` - source: Checkly Results, Grafana Analyze Results
- `Alert channel must be testable and secrets must be masked` - source: Grafana Contact Points, Checkly Alert Channels
- `Public status should speak in components, not internal probe IDs` - source: Atlassian Statuspage, Better Stack subscriptions
- `Do not claim DNS/HTTP checks unless contract/executor supports them` - source: Netstamp API/domain evidence
