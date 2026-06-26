# 07. Synthetic 使用者訪談模擬

> 本文件為假設式 / synthetic 訪談模擬，沒有進行真實招募、錄音、訪談或逐字稿整理。所有受訪者、情境與引言皆為研究假設，用於探索 Netstamp 可能的使用者需求與風險，不可被引用為真實使用者證據。

## 研究前提

### 產品脈絡證據

本輪快速閱讀了 repo root `README.md`、docs、API TypeSpec 與 web app route / feature code。可確認 Netstamp 的產品脈絡是：

- Self-hosted network observability / monitoring。
- 使用者可部署 controller、PostgreSQL / TimescaleDB 與 probe agents。
- Probes 代表網路觀測點，可放在 region、lab、edge、private server、provider boundary 等位置。
- Checks 包含 ping、TCP connect、traceroute；結果包含 latency、packet loss、connect failure、route / topology 等。
- Web app 有 Dashboard、Probes、Checks、Alerts、Status Pages、Labels、Insight、Members、Settings。
- API contract 包含 projects、members、labels、checks、probes、assignments、results、alerts、public status pages、probe runtime、install assets。
- Alerting / incident / notification 支援 rules、incidents、webhook / Slack / Discord / Telegram / email 等通知方向。
- Public status pages 可對外呈現 status、incidents、assignment metrics 與 chart。

### 命名誤解限制

在已讀 evidence 中，沒有看到 Netstamp 是 document stamping、certificate stamping、timestamp certificate 或 digital seal 產品。因此本文件不把 document / stamp / certificate 當作主要使用情境；只把它列為「產品命名可能造成誤解」與「需要釐清的研究問題」。

### Synthetic 受訪者設計

本輪建立 8 位 synthetic 受訪者，覆蓋以下角色需求：

| ID  | Synthetic 受訪者                       | 主要角色覆蓋              |
| --- | -------------------------------------- | ------------------------- |
| P1  | SRE / 平台工程師                       | 技術型使用者、高頻使用者  |
| P2  | NOC 值班工程師                         | 高頻使用者、事件處理      |
| P3  | IT Operations 經理                     | 管理者                    |
| P4  | Customer Success / Support Lead        | 非技術型使用者            |
| P5  | Enterprise Security / Compliance Owner | 高風險 / 合規敏感使用者   |
| P6  | 外部 Incident Reviewer / 客戶稽核代表  | 驗證方 / 外部 stakeholder |
| P7  | 第一次使用的開源 self-hosting 評估者   | 第一次使用者              |
| P8  | 懷疑產品可信度的 Network Architect     | 懷疑產品可信度使用者      |

## P1. Synthetic 受訪者：平台 SRE / 高頻技術使用者

### Profile

- 角色：中型 SaaS 公司的 Senior SRE。
- 技術背景：熟悉 Kubernetes、Prometheus、Grafana、OpenTelemetry、PagerDuty、Cloudflare / AWS / GCP network primitives。
- 使用頻率假設：每天多次看 dashboard，事件期間高頻切換 probes、checks、insight。
- 主要責任：確認 API latency、regional packet loss、provider route changes 是否影響實際使用者。

### Context

團隊已經有 cloud-based uptime monitoring，但常遇到「cloud monitor 沒報警，特定 ISP / region 使用者卻報障」的情況。SRE 想把 probes 放在 office edge、customer-region VPS、home broadband、partner lab，並用 Netstamp 補足「從我們控制的 viewpoint 看網路」。

### Goal

- 快速建立 probes 和 checks。
- 在 incident 時辨識是 service down、單一 route degraded、特定 ISP loss，還是 probe 自己壞掉。
- 把 Netstamp alerts 接進既有 incident flow。

### Interview transcript summary

P1 會先問 Netstamp 和 Prometheus blackbox exporter、Grafana Synthetic Monitoring、Pingdom / Datadog Synthetics 的差異。他認同「controlled probes」價值，但會要求資料模型能清楚對應 probe、check、assignment 和 result。他覺得 web app 如果能從 Dashboard 點進特定 probe/check pair 的 insight，並保留 time window、query URL，就很適合 incident room 使用。

他最在意結果可信度：一個紅色狀態必須能拆開看 packet loss、sample count、timeout、probe heartbeat、agent version、public IP / AS，以及最近是否有 selector 或 check config change。對 alerts，他偏好 metric threshold + min samples + cooldown 的保守設計，但希望有「為什麼這次 firing」的 explanation。

### Quotes

- **Synthetic quote:**「我不缺另一個 uptime monitor，我缺的是能說明『從東京 VPS 到我們 API 的路徑剛剛變了』的工具。」
- **Synthetic quote:**「如果 alert 只寫 packet loss 12.5%，我還需要知道 samples、window、probe 是否健康，不然很難相信它。」
- **Synthetic quote:**「assignment 這個概念對工程師合理，但 UI 要讓我看得出 check 到底跑在哪些 probes 上。」

### Pain points

- 現有監控多半從 vendor cloud region 看服務，無法代表私有網路或客戶所在 ISP。
- Incident 中需要在多個工具之間比對 latency、packet loss、traceroute、probe health。
- Alert 如果沒有 sample / window / source probe context，容易被視為 noise。
- Selector / labels 很 powerful，但錯配會造成 silent coverage gaps。

### Desired outcomes

- 事件期間能在 30 秒內回答：哪些 probes 受影響、哪些 checks 受影響、是 loss / latency / route / TCP failure 哪一類。
- 每個 alert 都有清楚的 evaluation summary。
- 能分享一個 read-only insight URL 給 incident channel。
- Probe lifecycle 清楚，包括 last seen、agent version、public IP、AS、secret rotation 狀態。

### Feature requests

- Incident drilldown：從 alert incident 直接跳到同一 probe/check/time window 的 insight。
- Assignment coverage preview：建立或修改 selector 時，顯示會影響哪些 probes，並提示 coverage 下降。
- Alert explanation panel：顯示 metric、threshold、value、samples、window、cooldown、suppressed notification count。
- Config change markers：在 charts 上標記 check config、selector、probe metadata 改變時間。
- Export / API：讓 SRE 把 incident evidence 匯出到 postmortem。

### Trust concerns

- Probe agent 是否可能卡住但仍回 heartbeat？
- Clock skew 或 delayed result submission 是否會污染 time window？
- Controller / DB 壓力大時，alert evaluation 是否會落後？
- Traceroute 在不同 OS / privilege / ICMP policy 下結果如何標註可信度？

### Research implications

- 需要深入研究「結果可信度 UI」：使用者需要哪些 metadata 才會相信 alert。
- Probe/check/assignment mental model 是核心理解門檻，需測試 IA 與 onboarding。
- 高頻技術使用者會重視 deep link、filter persistence、API / export、audit markers。

## P2. Synthetic 受訪者：NOC 值班工程師 / 高頻事件處理者

### Profile

- 角色：ISP / hosting provider 的 Network Operations Center 值班工程師。
- 技術背景：懂基本 BGP、traceroute、packet loss、TCP connection，但不負責開發。
- 使用頻率假設：輪班期間常駐 dashboard，告警時依 runbook 處理。
- 主要責任：確認客訴或內部告警是否為網路問題，回報 escalation 所需證據。

### Context

NOC 的工作節奏偏快：告警進來後，需要快速判斷 severity、影響範圍、是否要叫醒 network engineer。P2 不一定會建立新的 checks，但會頻繁查看 probes、public status、incidents 和 latest results。

### Goal

- 用一致的畫面判斷事件狀態。
- 快速複製 incident 摘要給 Slack / ticket。
- 不被太多低信心告警打斷。

### Interview transcript summary

P2 對 Netstamp 的興趣在於「把網路觀測點變成可以輪班交接的儀表板」。他希望 Dashboard 不是只有 probes online / checks active，而是可以看到 open incidents、最近 failing assignments、stale probes。他不想理解太多 selector 語法，只想知道「這個 check 現在在哪裡跑」、「哪幾個 probe 沒資料」。

他對 traceroute topology 有興趣，但擔心資訊過多。值班時他會需要簡化版：affected probes、common failing hop、first bad time、last good time、latest value。對 status page，他希望內部用和外部用分開，避免把太細的網路細節暴露給客戶。

### Quotes

- **Synthetic quote:**「值班時我不會寫 selector，我只想知道現在是不是客戶真的受影響。」
- **Synthetic quote:**「如果 probe stale 跟服務壞掉看起來一樣紅，我會花時間追錯方向。」
- **Synthetic quote:**「我需要一段可以貼到 incident ticket 的摘要，不是叫我截三張圖。」

### Pain points

- Dashboard 的高階指標若太少，無法支持 NOC triage。
- Probe stale、no data、timeout、error 若視覺語言不清楚，會導致錯誤升級。
- Traceroute 資訊對值班者可能太細，需要事件摘要層。
- 外部 status page 和內部 incident view 的資訊需求不同。

### Desired outcomes

- 一眼看到 open / acknowledged / resolved incidents。
- 能區分 no result、stale、timeout、partial、successful。
- 能用 role-based view 保護敏感 details。
- 交班時能查看過去 shift 的 incident timeline。

### Feature requests

- NOC mode dashboard：open incidents、latest failing assignments、stale probes、recent recoveries。
- One-click incident summary copy：含時間、scope、metric、affected probes、latest status。
- Probe stale alert 或至少 stale grouping。
- Internal note / handoff field for incidents。
- Public status page redaction controls：控制是否顯示 probe name、target、AS、chart。

### Trust concerns

- Alert resolve 是否代表服務恢復，還是只是沒有新 sample？
- Public status page 的 generated time 是否足以讓外部知道資料新鮮度？
- 值班者誤刪 rule / notification 的保護機制是否足夠？

### Research implications

- 需要把 incident triage flow 和 SRE deep-dive flow 分開設計。
- 狀態 taxonomy 必須清楚，不可讓 stale / no data / failure 混在一起。
- 外部溝通需要不同資訊密度與 redaction policy。

## P3. Synthetic 受訪者：IT Operations 經理 / 管理者

### Profile

- 角色：跨區辦公室 IT Operations Manager。
- 技術背景：理解網路與 SLA，但不寫程式，不配置低階 probe agent。
- 使用頻率假設：每週查看趨勢，每月做 service review；事件期間查看 status。
- 主要責任：證明跨 office / VPN / SaaS edge 的連線品質，規劃 ISP / provider 改善。

### Context

公司在台北、新加坡、東京、舊金山都有辦公室，常有員工抱怨某些 SaaS 或內部服務慢。IT 團隊想部署 probes 在 office edge 與幾個 cloud regions，長期追 latency / loss，避免只靠 anecdotal complaints。

### Goal

- 用報表或 dashboard 看到長期趨勢。
- 把「哪個辦公室 / ISP 經常不穩」轉成可討論的數據。
- 管理成員權限，讓不同地區 IT 可查看自己的 probes。

### Interview transcript summary

P3 不會關注 API 端點或 TypeSpec，但很在意 Netstamp 是否能「讓非工程主管理解」。他希望 labels 能以 business terms 呈現，例如 office、ISP、region、criticality。對他來說，Probe 不是機器，而是「觀測位置」；如果 UI 太偏 hostname / agent details，管理者不容易使用。

他也擔心 self-hosted 維運成本。若要向主管推導入，必須知道 deployment footprint、資料保存、備份、權限控管、通知渠道、公開狀態頁是否可對 stakeholder 開放。

### Quotes

- **Synthetic quote:**「我不需要每個 hop 的細節，我需要知道東京辦公室這週是不是比上週差。」
- **Synthetic quote:**「Probe 對我來說是觀測點，不是一台機器；如果系統都講 agent，我會需要工程師翻譯。」
- **Synthetic quote:**「自架可以接受，但我要知道維運成本和權限邊界。」

### Pain points

- 技術頁面若缺少 management summary，管理者難以用於決策。
- Labels 如果沒有治理方式，長期會混亂，影響報表可信度。
- Self-hosting 的備份、升級、資料保留與安全責任不清楚。
- 權限只若只有 owner/admin/editor/viewer，可能不足以支援跨區分工。

### Desired outcomes

- 趨勢報表可按 region、office、ISP、service、criticality 分組。
- 每週 / 每月可匯出 summary。
- 簡單管理 probe ownership 和 project members。
- 有清楚 deployment / upgrade / backup guide。

### Feature requests

- Executive summary view：availability、latency p95、loss hot spots、top degraded locations。
- Saved views by label group。
- Scheduled report / CSV export。
- Label governance：建議 key、必填 labels、重複值整理。
- Project-level retention / rollup settings。

### Trust concerns

- 是否能防止某地區 IT 修改 labels 造成歷史報表不可比？
- 趨勢數據是否有 downsampling / retention 解釋？
- Public status page 對外顯示是否會暴露內部拓撲或供應商資訊？

### Research implications

- 管理者需要「位置 / 服務 / 供應商」語言，而不是只看 probe/check technical primitives。
- Netstamp 的 self-hosted 定位需要搭配 operational readiness messaging。
- Labels 是資訊架構基礎，需要研究治理與報表用途。

## P4. Synthetic 受訪者：Customer Support Lead / 非技術型使用者

### Profile

- 角色：B2B SaaS 客服主管。
- 技術背景：能讀懂 up/down、latency、incident，但不懂 traceroute 或 selector。
- 使用頻率假設：客訴或 incident 期間使用；平常查看 status page。
- 主要責任：對客戶說明服務狀態，降低工程團隊被重複打擾。

### Context

Support 常收到「我們這邊連不上」的客訴，但 internal monitoring 顯示服務正常。工程師可能用 Netstamp 查特定 region 或 probe 的網路狀態；Support 希望能自助看到可對客戶解釋的版本。

### Goal

- 判斷是否已有已知 incident。
- 取得可直接給客戶的狀態說法。
- 避免誤讀技術細節。

### Interview transcript summary

P4 對 Netstamp 的 primary UI 不一定有需求，但對 public status page / read-only internal view 非常有需求。他希望畫面能用「服務狀態」而不是「assignment」說話。例如：API from East Asia degraded、Dashboard reachable from most probes、Tokyo broadband path has high packet loss。

他也指出 Netstamp 這個名字可能讓非技術同事誤以為是「文件蓋章」或「憑證 timestamp」。如果行銷頁或 onboarding 沒有一開始說清楚「network observability from probes」，Support 可能不知道該把產品放進哪個 internal tool category。

### Quotes

- **Synthetic quote:**「客戶問的是『你們有沒有壞』，不是問 probe assignment 是什麼。」
- **Synthetic quote:**「我可以看 latency 圖，但不要讓我解釋 traceroute hop。」
- **Synthetic quote:**「Netstamp 這個名字第一次看會像文件或憑證工具，首頁副標要很清楚。」

### Pain points

- 技術術語太多會讓非技術角色害怕誤用。
- Public status page 若沒有 narrative / incident copy，客服仍要找工程師翻譯。
- 外部分享時需要避免暴露 internal target、probe name 或 location precision。
- 名稱可能導致 product category 誤解。

### Desired outcomes

- Read-only status view 用 customer-safe vocabulary。
- Incident summary 有 recommended customer wording。
- Public status page 顯示資料更新時間與狀態定義。
- 命名與 onboarding 清楚區分 network observability，不是 document stamping。

### Feature requests

- Customer-facing incident copy generator。
- Status page element alias：對外顯示「Asia API」而非內部 check target。
- Tooltip / glossary：Ok、Partial、Timeout、Stale、No result 的簡短說明。
- Redaction presets：public、partner、internal。
- Non-technical dashboard mode。

### Trust concerns

- 對外狀態是否和內部 incident 狀態同步？
- 客服是否可能看到不該看到的內部 endpoint / probe location？
- 生成的 status wording 是否會過度承諾 SLA？

### Research implications

- Status page 是 external communication product surface，不只是觀測結果列表。
- 非技術角色需要語言轉換層與權限保護。
- 命名誤解是實際 onboarding 風險，應在首頁、docs、first-run flow 研究。

## P5. Synthetic 受訪者：Enterprise Security / Compliance Owner

### Profile

- 角色：金融科技公司的 Security & Compliance Owner。
- 技術背景：理解安全架構、稽核、資料分類、供應商風險，非日常 SRE。
- 使用頻率假設：導入前審查、季度稽核、重大 incident postmortem。
- 主要責任：確保 self-hosted observability 不違反資料、權限與變更控制要求。

### Context

公司可能在 regulated environment 部署 probes 到 private network、VPC、customer-adjacent environment。P5 對 open-source / self-hosted 有好感，但要求明確的 secret handling、audit log、RBAC、retention、network egress policy。

### Goal

- 評估 Netstamp 是否可部署在高風險網路。
- 確認 probe secret、webhook URL、email / Slack / Telegram token 等敏感資料如何保存。
- 確認操作事件可被稽核。

### Interview transcript summary

P5 一開始會問 threat model：probe agent 能看到什麼？controller 可以下發什麼 assignment？如果 probe 被盜，是否能 rotate secret、disable probe、限制它只能提交自己的 results？他也會問 public status pages 是否可能公開敏感 topology。

他對 API / OpenAPI 合約很正面，因為 enterprise 導入會需要自動化審核和整合。但他會要求所有 unsafe action 有 audit event，例如 create/delete probes、rotate secrets、change alert notifications、enable public status page、edit members / roles。

### Quotes

- **Synthetic quote:**「我喜歡 self-hosted，但 self-hosted 不代表安全；我要看 secret lifecycle 和 audit trail。」
- **Synthetic quote:**「Probe 如果被 compromise，它能拿到哪些 checks？能不能被用來掃內網？」
- **Synthetic quote:**「Public status page 對外很好，但預設不能洩漏內部 hostname 或準確位置。」

### Pain points

- Probe install command 包含 secret，一旦複製、貼到 ticket 或 shell history，可能外洩。
- Alert notifications 涉及 webhook / bot token / email recipients，需要 secret masking 與測試紀錄。
- 若缺 audit log，合規環境很難允許多人操作。
- Public status page 的 external exposure 需要清楚 approval flow。

### Desired outcomes

- Probe secret rotation、disable、last-used metadata 清楚。
- Role permission 有明確矩陣。
- Audit log 可查詢、匯出。
- Sensitive fields masked，且有 safe reveal / rotation flow。
- Status page 預設最小暴露。

### Feature requests

- Audit log：members、probes、checks、labels、alerts、notifications、status pages 的 create/update/delete。
- Secret hygiene：install command 顯示一次、copy warning、secret age、rotation reminder。
- Network policy guidance：probe egress endpoints、minimum privileges、container / systemd hardening。
- Status page publish approval / confirmation checklist。
- API tokens / service accounts with scoped permissions。

### Trust concerns

- Agent binary supply chain、installer script integrity、upgrade path。
- Webhook / notification secret storage 是否加密。
- API session cookie / CSRF / auth hardening。
- Multi-tenant project isolation 是否完整。

### Research implications

- 高風險使用者需要 security readiness content，不能只靠開源透明度。
- Probe install / secret handling 是 first-run 和 compliance 的共同高風險點。
- Public status page 應研究 safe defaults 與 publish checklist。

## P6. Synthetic 受訪者：外部 Incident Reviewer / 客戶稽核代表

### Profile

- 角色：大型企業客戶的 Vendor Management / Incident Review 代表。
- 技術背景：能理解 incident timeline、SLA、affected region、public status，但不會操作 Netstamp。
- 使用頻率假設：重大事件後查看報告；有時看 public status page。
- 主要責任：確認 vendor 對事件的說明是否可信，是否符合合約與 RCA 要求。

### Context

P6 不是 Netstamp 的內部使用者，而是可能透過 public status page、incident report 或 exported evidence 間接接觸 Netstamp 的 external stakeholder。若 Netstamp 的 public output 能清楚顯示 generated time、metrics、incident lifecycle，會提升 vendor 說明可信度。

### Goal

- 確認事件是否真的只影響某 region / path。
- 看懂 incident opened / resolved 時間與判斷依據。
- 取得可存檔的 evidence。

### Interview transcript summary

P6 表示他不需要完整內部細節，但需要「足以佐證」的資訊。他會質疑 vendor 自己提供的 status page 是否只是手動宣告；如果能看到 Netstamp 的 measurement method、probe regions、sample windows、generated timestamp，會比較可信。

他也希望 external status 不要使用內部術語，例如 check ID、probe ID、assignment ID。更好的呈現是：受影響服務、受影響區域、測量方式、目前狀態、最近更新、歷史 chart、incident notes。

### Quotes

- **Synthetic quote:**「我不需要你們內部拓撲，但我要知道這個『已恢復』是根據什麼量測。」
- **Synthetic quote:**「如果狀態頁只有一句 degraded，我會把它當公關文字，不是 evidence。」
- **Synthetic quote:**「Generated time 很重要，因為我需要知道我截圖時資料是不是最新。」

### Pain points

- Public status page 常缺 measurement methodology。
- Incident report 若無 raw-ish evidence 或 timestamp，稽核可信度低。
- External stakeholder 不懂 internal names，需要 customer-safe aliases。
- 如果狀態是 vendor 自己控制，會懷疑有選擇性揭露。

### Desired outcomes

- Public status page 具備 generated timestamp、status definitions、measurement scope。
- Incident lifecycle 可追溯：opened、acknowledged、resolved、last evaluated。
- RCA 可附上 Netstamp export / snapshot。
- External view 使用 service / region / customer-facing labels。

### Feature requests

- Public incident timeline。
- Evidence snapshot export：PDF / JSON / signed permalink 的研究方向。
- Measurement methodology section：probe count、regions、sample window、check type。
- Status page history / uptime by component。
- Optional third-party-verifiable evidence hash，但不應包裝成 certificate 產品。

### Trust concerns

- Vendor 可否事後修改 public incident history？
- Public chart 是否顯示完整時間窗，還是被裁切？
- Probe location 是否足以代表 customer traffic？

### Research implications

- External stakeholder 的 trust 來自透明 method、timestamp、history，而不是更多 raw technical data。
- 可研究「evidence snapshot」但需避免把 Netstamp 定位成文件憑證 / timestamp certificate。
- Status page 應支援 customer-facing alias 與 measurement explanation。

## P7. Synthetic 受訪者：第一次使用的開源 self-hosting 評估者

### Profile

- 角色：小型團隊的 DevOps generalist，第一次評估 Netstamp。
- 技術背景：會 Docker Compose、Linux service、基本 networking；不熟 Netstamp domain model。
- 使用頻率假設：初期一次性 setup，成功後每週查看。
- 主要責任：快速試用，判斷是否值得導入。

### Context

P7 從 GitHub README 看到 Netstamp，可接受用 Docker Compose 起 controller。他想在 30 分鐘內完成：部署、建立帳號 / project、建立 probe、跑第一個 ping check、看到結果、設定一個 alert 或 status page。

### Goal

- 最短路徑完成 first value。
- 不需要先讀完整 architecture。
- 遇到 probe 沒上線時知道怎麼 debug。

### Interview transcript summary

P7 喜歡 README 的 quick start，但真正進 app 後可能卡在 probe 概念：先 create probe，拿 secret，跑 installer，等 heartbeat；再 create check，選 probes / selector；再看 insight。他會希望 onboarding checklist 把這些步驟串起來。

他對 location search / manual coordinates 很有感，因為 map 可以確認 probe viewpoint。但如果 geocoding 失敗或 probe install 後沒有 heartbeat，他需要明確 diagnostic steps：controller URL、probe secret、systemd status、network egress、agent logs。

### Quotes

- **Synthetic quote:**「我想先看到一條 latency 圖，再決定要不要讀完整 docs。」
- **Synthetic quote:**「Create probe 之後如果 agent 沒 heartbeat，我需要知道下一步查哪裡。」
- **Synthetic quote:**「Selector 對第一次使用者太抽象，先讓我選全部 probes 或手動選 probes 就好。」

### Pain points

- First-run path 跨 docs、web app、shell installer、probe heartbeat、check config，容易中斷。
- Probe / check / assignment 的順序不直覺。
- Installer command 失敗時的 troubleshooting 不足會導致放棄。
- Alert / status page 可能太早出現，分散 first value。

### Desired outcomes

- 30 分鐘內看到第一個 result chart。
- App 內 onboarding checklist 清楚顯示剩餘步驟。
- Probe install 失敗有可操作的 diagnostic。
- 初始 check template 合理，例如 ping example.com 或自己的 API。

### Feature requests

- First-run checklist：Create project -> New probe -> Install agent -> Create check -> View insight。
- Empty states with next action：沒有 probes、沒有 checks、沒有 results、probe no heartbeat。
- Probe install troubleshooting drawer。
- Starter check templates：HTTP/TCP 443、ping API、traceroute target。
- Demo dataset / read-only demo mode，引導理解 dashboard。

### Trust concerns

- Installer script 會做什麼？是否可先預覽？
- Agent 需要 root 嗎？traceroute / ICMP 權限如何處理？
- Docker Compose 的 production hardening 是否足夠？

### Research implications

- 第一次使用者成功關鍵不是功能量，而是 first value path。
- Installer transparency 與 troubleshooting 會直接影響 adoption。
- Selector 應分層揭露，先提供 low-friction 選項。

## P8. Synthetic 受訪者：懷疑產品可信度的 Network Architect

### Profile

- 角色：大型網路團隊的 Principal Network Architect。
- 技術背景：深入 BGP、routing policy、measurement methodology、active probing limitations。
- 使用頻率假設：導入評估和重大事件時使用；平常不一定登入。
- 主要責任：判斷 Netstamp 的 measurement 是否足以支持網路工程決策。

### Context

P8 對任何 active measurement 工具都抱持懷疑：ping 可能被 deprioritized、traceroute hop 不一定代表 forwarding path、TCP connect 成功不代表 application health、單一 VPS 不代表 region。他願意試用，但要求 Netstamp 明確標示方法限制，不要過度詮釋結果。

### Goal

- 確認 Netstamp 的 metrics definitions。
- 看出每個 result 的 measurement parameters。
- 避免團隊把 Netstamp 圖表當成絕對真相。

### Interview transcript summary

P8 會逐一追問：ping packet size、count、timeout、IP family；TCP port、timeout；traceroute protocol、max hops、queries per hop、partial status；result aggregation 怎麼算；sample loss 怎麼和 probe offline 區分。若 UI 顯示「route changed」，他會要求看到 hop diff 和 confidence。

他很欣賞 self-hosted + controlled probes 的定位，但希望產品在 messaging 上保守，像「evidence from configured viewpoints」而不是「internet truth」。他也希望 API 提供 raw runs 或足夠 detailed result，而不是只有 rolled-up chart。

### Quotes

- **Synthetic quote:**「Ping loss 是訊號，但不是結論；UI 不該替我過度解讀。」
- **Synthetic quote:**「Traceroute partial 很常見，你要告訴我 partial 的原因，不然 topology 圖會誤導。」
- **Synthetic quote:**「我會相信 Netstamp 作為 measurement notebook，但不會相信沒有方法說明的紅綠燈。」

### Pain points

- Observability UI 常把 noisy active measurements 簡化成過度自信的 status。
- Route / topology 視覺化容易讓非網路專家誤判 causality。
- Aggregation、downsampling、retention 若不透明，長期趨勢可信度下降。
- Probe location / AS / network context 不足會造成錯誤外推。

### Desired outcomes

- 每種 check 都有 method and limitations 說明。
- Chart / insight 顯示 config parameters 和 sample counts。
- Traceroute 能比較 runs，呈現 partial / timeout / hop changes 的限制。
- Public status page 的 claims 保守且可追溯。

### Feature requests

- Measurement methodology panel for each check type。
- Raw result / run detail view，尤其 traceroute run history。
- Confidence indicators：sample count、staleness、probe health、measurement method。
- Route diff view：last known good vs current run。
- API access to raw or semi-raw result rows with retention caveats。

### Trust concerns

- 是否把 ICMP deprioritization 誤判為服務 degraded？
- IPv4 / IPv6 capability 是否明確。
- Time synchronization 與 result timestamp semantics。
- Chart rollup 是否隱藏 outliers。

### Research implications

- 可信度懷疑者不是反對產品，而是反對過度詮釋；產品文案和 UI 需要承認 measurement limitations。
- Insight / topology 需要 confidence layer。
- API 和 raw-detail views 對高端技術評估有價值。

## Cross-Interview Notes

### 共同需求

- 清楚區分 probe health、check failure、stale/no data、service degradation。
- 從 alert / incident 直接進入相同 scope 的 evidence。
- Probe/check/assignment/selector mental model 需要分層解釋。
- Public status pages 需要外部友善的 alias、timestamp、status definitions 和 redaction。
- Self-hosted adoption 需要部署、升級、備份、安全、secret handling 的信任材料。

### 明顯分歧

- SRE / Network Architect 想要 raw details、API、methodology、confidence。
- NOC 想要事件摘要、runbook、低認知負擔。
- Manager 想要趨勢、分組、報表、治理。
- Support / external reviewer 想要對外說法、可存檔 evidence、低敏感度資訊。
- Compliance owner 想要 audit、RBAC、secret lifecycle、safe defaults。

### 需進一步真實研究驗證的問題

- 第一次使用者是否能在 30 分鐘內完成 first value path？
- 使用者如何自然理解 probe、check、assignment、selector 的關係？
- Alert explanation 要顯示哪些欄位才足以建立信任？
- Public status page 應公開多少 measurement detail 才能兼顧可信度與安全？
- Netstamp 命名是否真的造成 document / stamp / certificate 類誤解？若有，首頁與 onboarding 如何修正？
