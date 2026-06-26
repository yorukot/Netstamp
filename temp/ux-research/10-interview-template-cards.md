# 10. FigJam 使用者訪談模板卡片

> 本文件整理自 synthetic 使用者訪談與 affinity synthesis。以下內容皆為假設式研究素材，未經真實招募、訪談、錄音或逐字稿驗證；不可當作真實使用者證據引用。

## 使用方式

- 每一位受訪者區塊可直接拆成 FigJam 使用者訪談模板的一張卡片或一組 sticky notes。
- 每位受訪者皆標示 `[Synthetic interview] [Assumption]`。
- 每則引言皆標示 `[Synthetic quote]`。
- 「Affinity themes 簡短版」適合放在訪談卡片旁邊，作為主題群組與設計機會摘要。

---

## 受訪者 1：陳柏任 / 平台 SRE（Senior SRE）

**標籤：** [Synthetic interview] [Assumption]

**個人檔案**

- 中型 SaaS 公司的 Senior SRE。
- 熟悉 Kubernetes、Prometheus、Grafana、OpenTelemetry、PagerDuty、Cloudflare / AWS / GCP network primitives。
- 每天多次查看 dashboard，事件期間會高頻切換 probes、checks、insight。
- 主要責任是確認 API latency、regional packet loss、provider route changes 是否影響實際使用者。

**脈絡**

- 團隊已有 cloud-based uptime monitoring，但常遇到 vendor monitor 沒報警、特定 ISP 或 region 使用者卻報障的情況。
- 想把 probes 放在 office edge、customer-region VPS、home broadband、partner lab，用 Netstamp 補足「從自己控制的 viewpoint 看網路」。
- 會把 Netstamp 放進既有 incident room、on-call 與 postmortem 流程中。

**目標**

- 快速建立 probes 和 checks。
- 在 incident 時辨識問題是 service down、單一路由 degraded、特定 ISP loss，還是 probe 自己壞掉。
- 將 Netstamp alerts 接進既有 incident flow。
- 能快速分享 scoped insight 給 incident channel。

**訪談逐字稿摘要**

- 會先比較 Netstamp 與 Prometheus blackbox exporter、Grafana Synthetic Monitoring、Pingdom / Datadog Synthetics 的差異。
- 認同 controlled probes 的價值，但要求資料模型能清楚對應 probe、check、assignment、result。
- 希望從 Dashboard 或 incident 直接點進特定 probe/check pair 的 insight，並保留 time window 與 query URL。
- 最在意結果可信度；紅色狀態必須能拆開看 packet loss、sample count、timeout、probe heartbeat、agent version、public IP / AS，以及最近是否有 selector 或 check config change。
- 對 alerts 偏好 metric threshold、min samples、cooldown 的保守設計，但每次 firing 都需要 explanation。

**合成引言**

- [Synthetic quote]「我不缺另一個 uptime monitor，我缺的是能說明『從東京 VPS 到我們 API 的路徑剛剛變了』的工具。」
- [Synthetic quote]「如果 alert 只寫 packet loss 12.5%，我還需要知道 samples、window、probe 是否健康，不然很難相信它。」
- [Synthetic quote]「assignment 這個概念對工程師合理，但 UI 要讓我看得出 check 到底跑在哪些 probes 上。」

**痛點**

- 現有監控多半從 vendor cloud region 看服務，無法代表私有網路或客戶所在 ISP。
- Incident 中需要在多個工具之間比對 latency、packet loss、traceroute、probe health。
- Alert 如果沒有 sample、window、source probe context，容易被視為 noise。
- Selector / labels 很 powerful，但錯配會造成 silent coverage gaps。

**期望成果**

- 事件期間能在 30 秒內回答：哪些 probes 受影響、哪些 checks 受影響、是 loss / latency / route / TCP failure 哪一類。
- 每個 alert 都有清楚的 evaluation summary。
- 能分享 read-only insight URL 給 incident channel。
- Probe lifecycle 清楚，包括 last seen、agent version、public IP、AS、secret rotation 狀態。

**功能需求**

- Incident drilldown：從 alert incident 直接跳到同一 probe/check/time window 的 insight。
- Assignment coverage preview：建立或修改 selector 時顯示會影響哪些 probes，並提示 coverage 下降。
- Alert explanation panel：顯示 metric、threshold、value、samples、window、cooldown、suppressed notification count。
- Config change markers：在 charts 上標記 check config、selector、probe metadata 改變時間。
- Export / API：讓 SRE 把 incident evidence 匯出到 postmortem。

**信任疑慮**

- Probe agent 是否可能卡住但仍回 heartbeat。
- Clock skew 或 delayed result submission 是否會污染 time window。
- Controller / DB 壓力大時，alert evaluation 是否會落後。
- Traceroute 在不同 OS / privilege / ICMP policy 下，結果如何標註可信度。

**研究啟示**

- 結果可信度 UI 是核心；需要測試哪些 metadata 最能讓 SRE 信任 alert。
- Probe / check / assignment 心智模型是採用門檻，IA 與 onboarding 需要清楚分層。
- 高頻技術使用者會重視 deep link、filter persistence、API / export、audit markers。

---

## 受訪者 2：林佳穎 / NOC 值班工程師

**標籤：** [Synthetic interview] [Assumption]

**個人檔案**

- ISP / hosting provider 的 Network Operations Center 值班工程師。
- 理解基本 BGP、traceroute、packet loss、TCP connection，但不負責開發。
- 輪班期間常駐 dashboard，告警時依 runbook 處理。
- 主要責任是確認客訴或內部告警是否為網路問題，並回報 escalation 所需證據。

**脈絡**

- NOC 工作節奏快；告警進來後，需要快速判斷 severity、影響範圍、是否要叫醒 network engineer。
- 不一定會建立新的 checks，但會頻繁查看 probes、public status、incidents、latest results。
- 需要把網路觀測點轉成可以輪班交接的儀表板與事件摘要。

**目標**

- 用一致的畫面判斷事件狀態。
- 快速複製 incident 摘要給 Slack 或 ticket。
- 降低低信心告警造成的中斷。
- 在交班時能回顧 incident timeline。

**訪談逐字稿摘要**

- 對 Netstamp 的興趣是把多個 probe viewpoint 變成可操作的值班介面。
- Dashboard 不能只有 probes online / checks active；也需要 open incidents、最近 failing assignments、stale probes、recent recoveries。
- 不想理解太多 selector 語法，只想知道「這個 check 現在在哪裡跑」與「哪幾個 probe 沒資料」。
- 對 traceroute topology 有興趣，但值班時需要簡化成 affected probes、common failing hop、first bad time、last good time、latest value。
- 希望內部 incident view 與外部 status page 分開，避免把太細的網路細節暴露給客戶。

**合成引言**

- [Synthetic quote]「值班時我不會寫 selector，我只想知道現在是不是客戶真的受影響。」
- [Synthetic quote]「如果 probe stale 跟服務壞掉看起來一樣紅，我會花時間追錯方向。」
- [Synthetic quote]「我需要一段可以貼到 incident ticket 的摘要，不是叫我截三張圖。」

**痛點**

- Dashboard 的高階指標若太少，無法支持 NOC triage。
- Probe stale、no data、timeout、error 若視覺語言不清楚，會導致錯誤升級。
- Traceroute 資訊對值班者可能太細，需要事件摘要層。
- 外部 status page 和內部 incident view 的資訊需求不同。

**期望成果**

- 一眼看到 open / acknowledged / resolved incidents。
- 能區分 no result、stale、timeout、partial、successful。
- 能用 role-based view 保護敏感 details。
- 能查看過去 shift 的 incident timeline，支援交班與追蹤。

**功能需求**

- NOC mode dashboard：open incidents、latest failing assignments、stale probes、recent recoveries。
- One-click incident summary copy：含時間、scope、metric、affected probes、latest status。
- Probe stale alert 或 stale grouping。
- Internal note / handoff field for incidents。
- Public status page redaction controls：控制是否顯示 probe name、target、AS、chart。

**信任疑慮**

- Alert resolve 是否代表服務恢復，還是只是沒有新 sample。
- Public status page 的 generated time 是否足以讓外部知道資料新鮮度。
- 值班者誤刪 rule / notification 的保護機制是否足夠。

**研究啟示**

- Incident triage flow 和 SRE deep-dive flow 需要分開設計。
- 狀態 taxonomy 必須清楚，避免 stale / no data / failure 混在一起。
- 外部溝通需要不同資訊密度與 redaction policy。

---

## 受訪者 3：王明哲 / IT Operations 經理

**標籤：** [Synthetic interview] [Assumption]

**個人檔案**

- 跨區辦公室 IT Operations Manager。
- 理解網路與 SLA，但不寫程式，也不配置低階 probe agent。
- 每週查看趨勢，每月做 service review；事件期間查看 status。
- 主要責任是證明跨 office / VPN / SaaS edge 的連線品質，並規劃 ISP / provider 改善。

**脈絡**

- 公司在台北、新加坡、東京、舊金山都有辦公室，員工常抱怨某些 SaaS 或內部服務慢。
- IT 團隊想部署 probes 在 office edge 與幾個 cloud regions，長期追蹤 latency / loss。
- 希望用數據取代 anecdotal complaints，支援跨區 IT 與供應商討論。

**目標**

- 用報表或 dashboard 看到長期趨勢。
- 把「哪個辦公室 / ISP 經常不穩」轉成可討論的數據。
- 管理成員權限，讓不同地區 IT 可查看自己的 probes。
- 向主管說明 self-hosted 維運成本與責任邊界。

**訪談逐字稿摘要**

- 不關注 API 端點或 TypeSpec，但很在意 Netstamp 是否能讓非工程主管理解。
- 希望 labels 能以 business terms 呈現，例如 office、ISP、region、criticality。
- 對他來說，Probe 是「觀測位置」，不是一台機器；若 UI 太偏 hostname / agent details，管理者需要工程師翻譯。
- 擔心 self-hosted 維運成本；導入前需要知道 deployment footprint、資料保存、備份、權限控管、通知渠道、公開狀態頁是否可對 stakeholder 開放。

**合成引言**

- [Synthetic quote]「我不需要每個 hop 的細節，我需要知道東京辦公室這週是不是比上週差。」
- [Synthetic quote]「Probe 對我來說是觀測點，不是一台機器；如果系統都講 agent，我會需要工程師翻譯。」
- [Synthetic quote]「自架可以接受，但我要知道維運成本和權限邊界。」

**痛點**

- 技術頁面若缺少 management summary，管理者難以用於決策。
- Labels 如果沒有治理方式，長期會混亂，影響報表可信度。
- Self-hosting 的備份、升級、資料保留與安全責任不清楚。
- 若權限只有 owner/admin/editor/viewer，可能不足以支援跨區分工。

**期望成果**

- 趨勢報表可按 region、office、ISP、service、criticality 分組。
- 每週 / 每月可匯出 summary。
- 能簡單管理 probe ownership 和 project members。
- 有清楚 deployment / upgrade / backup guide。

**功能需求**

- Executive summary view：availability、latency p95、loss hot spots、top degraded locations。
- Saved views by label group。
- Scheduled report / CSV export。
- Label governance：建議 key、必填 labels、重複值整理。
- Project-level retention / rollup settings。

**信任疑慮**

- 是否能防止某地區 IT 修改 labels，造成歷史報表不可比。
- 趨勢數據是否有 downsampling / retention 解釋。
- Public status page 對外顯示是否會暴露內部拓撲或供應商資訊。

**研究啟示**

- 管理者需要「位置 / 服務 / 供應商」語言，而不是只看 probe/check technical primitives。
- Netstamp 的 self-hosted 定位需要搭配 operational readiness messaging。
- Labels 是資訊架構基礎，需要研究治理與報表用途。

---

## 受訪者 4：周怡君 / Customer Support Lead

**標籤：** [Synthetic interview] [Assumption]

**個人檔案**

- B2B SaaS 客服主管。
- 能讀懂 up/down、latency、incident，但不懂 traceroute 或 selector。
- 客訴或 incident 期間使用；平常查看 status page。
- 主要責任是對客戶說明服務狀態，降低工程團隊被重複打擾。

**脈絡**

- Support 常收到「我們這邊連不上」的客訴，但 internal monitoring 顯示服務正常。
- 工程師可能用 Netstamp 查特定 region 或 probe 的網路狀態；Support 希望能自助看到可對客戶解釋的版本。
- 對 public status page / read-only internal view 的需求高於 primary UI。

**目標**

- 判斷是否已有已知 incident。
- 取得可直接給客戶的狀態說法。
- 避免誤讀技術細節。
- 讓對外溝通與內部 incident 狀態同步。

**訪談逐字稿摘要**

- 需要畫面用「服務狀態」而不是「assignment」說話。
- 希望看到像 API from East Asia degraded、Dashboard reachable from most probes、Tokyo broadband path has high packet loss 這類 customer-safe 敘述。
- 指出 Netstamp 這個名字可能讓非技術同事誤以為是文件蓋章或憑證 timestamp。
- 若首頁或 onboarding 沒有說清楚 network observability from probes，Support 可能不知道該把產品放進哪個 internal tool category。

**合成引言**

- [Synthetic quote]「客戶問的是『你們有沒有壞』，不是問 probe assignment 是什麼。」
- [Synthetic quote]「我可以看 latency 圖，但不要讓我解釋 traceroute hop。」
- [Synthetic quote]「Netstamp 這個名字第一次看會像文件或憑證工具，首頁副標要很清楚。」

**痛點**

- 技術術語太多會讓非技術角色害怕誤用。
- Public status page 若沒有 narrative / incident copy，客服仍要找工程師翻譯。
- 外部分享時需要避免暴露 internal target、probe name 或 location precision。
- 名稱可能導致 product category 誤解。

**期望成果**

- Read-only status view 使用 customer-safe vocabulary。
- Incident summary 有 recommended customer wording。
- Public status page 顯示資料更新時間與狀態定義。
- 命名與 onboarding 清楚區分 network observability，不是 document stamping。

**功能需求**

- Customer-facing incident copy generator。
- Status page element alias：對外顯示「Asia API」而非內部 check target。
- Tooltip / glossary：Ok、Partial、Timeout、Stale、No result 的簡短說明。
- Redaction presets：public、partner、internal。
- Non-technical dashboard mode。

**信任疑慮**

- 對外狀態是否和內部 incident 狀態同步。
- 客服是否可能看到不該看到的內部 endpoint / probe location。
- 生成的 status wording 是否會過度承諾 SLA。

**研究啟示**

- Status page 是 external communication product surface，不只是觀測結果列表。
- 非技術角色需要語言轉換層與權限保護。
- 命名誤解是 onboarding 風險，應在首頁、docs、first-run flow 中驗證。

---

## 受訪者 5：許德安 / Enterprise Security & Compliance Owner

**標籤：** [Synthetic interview] [Assumption]

**個人檔案**

- 金融科技公司的 Security & Compliance Owner。
- 理解安全架構、稽核、資料分類、供應商風險，但非日常 SRE。
- 導入前審查、季度稽核、重大 incident postmortem 時使用。
- 主要責任是確保 self-hosted observability 不違反資料、權限與變更控制要求。

**脈絡**

- 公司可能在 regulated environment 部署 probes 到 private network、VPC、customer-adjacent environment。
- 對 open-source / self-hosted 有好感，但需要明確的 secret handling、audit log、RBAC、retention、network egress policy。
- 會從 threat model、agent 行為、public exposure、安全預設值檢視導入風險。

**目標**

- 評估 Netstamp 是否可部署在高風險網路。
- 確認 probe secret、webhook URL、email / Slack / Telegram token 等敏感資料如何保存。
- 確認操作事件可被稽核。
- 確認 public status page 預設不洩漏敏感 topology。

**訪談逐字稿摘要**

- 一開始會問 probe agent 能看到什麼、controller 可以下發什麼 assignment、probe 被盜時可否 rotate secret / disable probe。
- 會確認被 compromise 的 probe 是否只能提交自己的 results，不能被用來掃內網。
- 對 API / OpenAPI 合約正面，因為 enterprise 導入需要自動化審核和整合。
- 要求所有 unsafe action 有 audit event，例如 create/delete probes、rotate secrets、change alert notifications、enable public status page、edit members / roles。

**合成引言**

- [Synthetic quote]「我喜歡 self-hosted，但 self-hosted 不代表安全；我要看 secret lifecycle 和 audit trail。」
- [Synthetic quote]「Probe 如果被 compromise，它能拿到哪些 checks？能不能被用來掃內網？」
- [Synthetic quote]「Public status page 對外很好，但預設不能洩漏內部 hostname 或準確位置。」

**痛點**

- Probe install command 包含 secret，一旦複製、貼到 ticket 或 shell history，可能外洩。
- Alert notifications 涉及 webhook / bot token / email recipients，需要 secret masking 與測試紀錄。
- 若缺 audit log，合規環境很難允許多人操作。
- Public status page 的 external exposure 需要清楚 approval flow。

**期望成果**

- Probe secret rotation、disable、last-used metadata 清楚。
- Role permission 有明確矩陣。
- Audit log 可查詢、匯出。
- Sensitive fields masked，且有 safe reveal / rotation flow。
- Status page 預設最小暴露。

**功能需求**

- Audit log：members、probes、checks、labels、alerts、notifications、status pages 的 create/update/delete。
- Secret hygiene：install command 顯示一次、copy warning、secret age、rotation reminder。
- Network policy guidance：probe egress endpoints、minimum privileges、container / systemd hardening。
- Status page publish approval / confirmation checklist。
- API tokens / service accounts with scoped permissions。

**信任疑慮**

- Agent binary supply chain、installer script integrity、upgrade path。
- Webhook / notification secret storage 是否加密。
- API session cookie / CSRF / auth hardening。
- Multi-tenant project isolation 是否完整。

**研究啟示**

- 高風險使用者需要 security readiness content，不能只靠開源透明度。
- Probe install / secret handling 是 first-run 和 compliance 的共同高風險點。
- Public status page 應研究 safe defaults 與 publish checklist。

---

## 受訪者 6：李雅婷 / 外部 Incident Reviewer 與客戶稽核代表

**標籤：** [Synthetic interview] [Assumption]

**個人檔案**

- 大型企業客戶的 Vendor Management / Incident Review 代表。
- 能理解 incident timeline、SLA、affected region、public status，但不會操作 Netstamp。
- 重大事件後查看報告；有時查看 public status page。
- 主要責任是確認 vendor 對事件的說明是否可信，是否符合合約與 RCA 要求。

**脈絡**

- 不是 Netstamp 內部使用者，而是透過 public status page、incident report 或 exported evidence 間接接觸 Netstamp。
- 若 public output 能清楚顯示 generated time、metrics、incident lifecycle，會提升 vendor 說明可信度。
- 不需要完整內部細節，但需要足以佐證的資訊。

**目標**

- 確認事件是否真的只影響某 region / path。
- 看懂 incident opened / resolved 時間與判斷依據。
- 取得可存檔的 evidence。
- 判斷 status page 是否只是手動宣告，或有 measurement method 支撐。

**訪談逐字稿摘要**

- 會質疑 vendor 自己提供的 status page 是否只是公關文字。
- 若能看到 measurement method、probe regions、sample windows、generated timestamp，會比較可信。
- 不希望看到 internal terms，例如 check ID、probe ID、assignment ID。
- 更好的呈現是受影響服務、受影響區域、測量方式、目前狀態、最近更新、歷史 chart、incident notes。

**合成引言**

- [Synthetic quote]「我不需要你們內部拓撲，但我要知道這個『已恢復』是根據什麼量測。」
- [Synthetic quote]「如果狀態頁只有一句 degraded，我會把它當公關文字，不是 evidence。」
- [Synthetic quote]「Generated time 很重要，因為我需要知道我截圖時資料是不是最新。」

**痛點**

- Public status page 常缺 measurement methodology。
- Incident report 若無 raw-ish evidence 或 timestamp，稽核可信度低。
- External stakeholder 不懂 internal names，需要 customer-safe aliases。
- 如果狀態是 vendor 自己控制，會懷疑有選擇性揭露。

**期望成果**

- Public status page 具備 generated timestamp、status definitions、measurement scope。
- Incident lifecycle 可追溯：opened、acknowledged、resolved、last evaluated。
- RCA 可附上 Netstamp export / snapshot。
- External view 使用 service / region / customer-facing labels。

**功能需求**

- Public incident timeline。
- Evidence snapshot export：PDF / JSON / signed permalink 的研究方向。
- Measurement methodology section：probe count、regions、sample window、check type。
- Status page history / uptime by component。
- Optional third-party-verifiable evidence hash，但不應包裝成 certificate 產品。

**信任疑慮**

- Vendor 可否事後修改 public incident history。
- Public chart 是否顯示完整時間窗，還是被裁切。
- Probe location 是否足以代表 customer traffic。

**研究啟示**

- External stakeholder 的 trust 來自透明 method、timestamp、history，而不是更多 raw technical data。
- 可研究 evidence snapshot，但需避免把 Netstamp 定位成文件憑證 / timestamp certificate。
- Status page 應支援 customer-facing alias 與 measurement explanation。

---

## 受訪者 7：張凱翔 / 第一次使用的開源 Self-hosting 評估者

**標籤：** [Synthetic interview] [Assumption]

**個人檔案**

- 小型團隊的 DevOps generalist，第一次評估 Netstamp。
- 會 Docker Compose、Linux service、基本 networking；不熟 Netstamp domain model。
- 初期一次性 setup，成功後每週查看。
- 主要責任是快速試用，判斷是否值得導入。

**脈絡**

- 從 GitHub README 看到 Netstamp，可接受用 Docker Compose 啟動 controller。
- 想在 30 分鐘內完成部署、建立帳號 / project、建立 probe、跑第一個 ping check、看到結果、設定一個 alert 或 status page。
- 對 location search / manual coordinates 有感，因為 map 能確認 probe viewpoint。

**目標**

- 最短路徑完成 first value。
- 不需要先讀完整 architecture。
- 遇到 probe 沒上線時知道怎麼 debug。
- 先看到一條 latency 圖，再決定是否深入導入。

**訪談逐字稿摘要**

- 喜歡 README quick start，但進 app 後可能卡在 probe 概念。
- First value path 跨 create probe、拿 secret、跑 installer、等 heartbeat、create check、選 probes / selector、看 insight。
- 希望 onboarding checklist 把步驟串起來。
- 若 geocoding 失敗或 probe install 後沒有 heartbeat，需要明確 diagnostic steps：controller URL、probe secret、systemd status、network egress、agent logs。

**合成引言**

- [Synthetic quote]「我想先看到一條 latency 圖，再決定要不要讀完整 docs。」
- [Synthetic quote]「Create probe 之後如果 agent 沒 heartbeat，我需要知道下一步查哪裡。」
- [Synthetic quote]「Selector 對第一次使用者太抽象，先讓我選全部 probes 或手動選 probes 就好。」

**痛點**

- First-run path 跨 docs、web app、shell installer、probe heartbeat、check config，容易中斷。
- Probe / check / assignment 的順序不直覺。
- Installer command 失敗時 troubleshooting 不足會導致放棄。
- Alert / status page 可能太早出現，分散 first value。

**期望成果**

- 30 分鐘內看到第一個 result chart。
- App 內 onboarding checklist 清楚顯示剩餘步驟。
- Probe install 失敗有可操作的 diagnostic。
- 初始 check template 合理，例如 ping example.com 或自己的 API。

**功能需求**

- First-run checklist：Create project -> New probe -> Install agent -> Create check -> View insight。
- Empty states with next action：沒有 probes、沒有 checks、沒有 results、probe no heartbeat。
- Probe install troubleshooting drawer。
- Starter check templates：HTTP/TCP 443、ping API、traceroute target。
- Demo dataset / read-only demo mode，引導理解 dashboard。

**信任疑慮**

- Installer script 會做什麼，是否可先預覽。
- Agent 是否需要 root；traceroute / ICMP 權限如何處理。
- Docker Compose 的 production hardening 是否足夠。

**研究啟示**

- 第一次使用者成功關鍵不是功能量，而是 first value path。
- Installer transparency 與 troubleshooting 會直接影響 adoption。
- Selector 應分層揭露，先提供 low-friction 選項。

---

## 受訪者 8：鄭維倫 / 懷疑產品可信度的 Network Architect

**標籤：** [Synthetic interview] [Assumption]

**個人檔案**

- 大型網路團隊的 Principal Network Architect。
- 深入 BGP、routing policy、measurement methodology、active probing limitations。
- 導入評估和重大事件時使用；平常不一定登入。
- 主要責任是判斷 Netstamp 的 measurement 是否足以支持網路工程決策。

**脈絡**

- 對任何 active measurement 工具都抱持懷疑。
- 認為 ping 可能被 deprioritized、traceroute hop 不一定代表 forwarding path、TCP connect 成功不代表 application health、單一 VPS 不代表 region。
- 願意試用，但要求 Netstamp 明確標示方法限制，不要過度詮釋結果。

**目標**

- 確認 Netstamp 的 metrics definitions。
- 看出每個 result 的 measurement parameters。
- 避免團隊把 Netstamp 圖表當成絕對真相。
- 讓 UI 與 public claims 保守且可追溯。

**訪談逐字稿摘要**

- 會追問 ping packet size、count、timeout、IP family；TCP port、timeout；traceroute protocol、max hops、queries per hop、partial status。
- 會檢查 result aggregation 怎麼算，sample loss 如何和 probe offline 區分。
- 若 UI 顯示 route changed，會要求看到 hop diff 和 confidence。
- 欣賞 self-hosted + controlled probes 的定位，但希望 messaging 是「evidence from configured viewpoints」，不是「internet truth」。
- 希望 API 提供 raw runs 或足夠 detailed result，而不是只有 rolled-up chart。

**合成引言**

- [Synthetic quote]「Ping loss 是訊號，但不是結論；UI 不該替我過度解讀。」
- [Synthetic quote]「Traceroute partial 很常見，你要告訴我 partial 的原因，不然 topology 圖會誤導。」
- [Synthetic quote]「我會相信 Netstamp 作為 measurement notebook，但不會相信沒有方法說明的紅綠燈。」

**痛點**

- Observability UI 常把 noisy active measurements 簡化成過度自信的 status。
- Route / topology 視覺化容易讓非網路專家誤判 causality。
- Aggregation、downsampling、retention 若不透明，長期趨勢可信度下降。
- Probe location / AS / network context 不足會造成錯誤外推。

**期望成果**

- 每種 check 都有 method and limitations 說明。
- Chart / insight 顯示 config parameters 和 sample counts。
- Traceroute 能比較 runs，呈現 partial / timeout / hop changes 的限制。
- Public status page 的 claims 保守且可追溯。

**功能需求**

- Measurement methodology panel for each check type。
- Raw result / run detail view，尤其 traceroute run history。
- Confidence indicators：sample count、staleness、probe health、measurement method。
- Route diff view：last known good vs current run。
- API access to raw or semi-raw result rows with retention caveats。

**信任疑慮**

- 是否把 ICMP deprioritization 誤判為服務 degraded。
- IPv4 / IPv6 capability 是否明確。
- Time synchronization 與 result timestamp semantics。
- Chart rollup 是否隱藏 outliers。

**研究啟示**

- 可信度懷疑者不是反對產品，而是反對過度詮釋；產品文案和 UI 需要承認 measurement limitations。
- Insight / topology 需要 confidence layer。
- API 和 raw-detail views 對高端技術評估有價值。

---

## Affinity themes 簡短版

### Theme 1：Probe / Check / Assignment 心智模型是採用門檻

- 使用者需要知道「從哪裡量測、量測什麼、目前跑在哪些 probes 上」。
- 第一次使用者與 NOC 不想先理解 selector；SRE 與 Architect 又需要可追溯的 assignment。
- 設計機會：first-run checklist、effective assignments 固定顯示、selector preview / coverage diff、低門檻手動選 probes。

### Theme 2：可信度來自狀態拆解，不是紅綠燈

- 紅色狀態必須拆成 timeout、packet loss、TCP failure、no data、stale probe、insufficient samples、partial traceroute 等狀態。
- SRE、NOC、External Reviewer、Network Architect 都要求 sample、window、probe health、measurement method。
- 設計機會：status taxonomy、alert explanation panel、confidence indicators、chart sample count / staleness 標示。

### Theme 3：Incident flow 需要同時支援 triage、deep dive、對外說明

- NOC 需要快速摘要，SRE 需要 deep link 與 evidence，Support 需要 customer-safe wording，External Reviewer 需要 timeline / timestamp。
- 單一 incident detail 應成為多角色 workflow hub。
- 設計機會：incident summary copy、internal/customer-safe 模板、related insight links、notification history、handoff notes。

### Theme 4：Public status page 是信任介面，不只是公開 dashboard

- 對外顯示太少會像公關文字，顯示太多可能洩漏 internal target、probe name、位置、provider。
- Support、Compliance、External Reviewer 都需要 alias、redaction、generated time、methodology。
- 設計機會：status page aliases、redaction presets、publish checklist、generated timestamp、methodology section。

### Theme 5：Self-hosted adoption 需要 operational readiness

- Self-hosted 是差異化，也帶來部署、備份、升級、secret、retention、alert worker reliability 責任。
- First-time evaluator 想快速看到 first result；manager / compliance owner 要 production readiness。
- 設計機會：production checklist、probe install troubleshooting、admin health view、upgrade / backup / retention guidance。

### Theme 6：Security、RBAC、Audit 是 enterprise 採用門檻

- Probe secret、notification secret、members / roles、public status page、unsafe actions 都是 audit-worthy。
- Compliance owner 擔心 supply chain、secret storage、scoped permissions、public exposure。
- 設計機會：audit log、dangerous action confirmation、scoped API tokens / service accounts、secret masking / rotation metadata。

### Theme 7：Labels 是 powerful primitive，也是治理風險

- Labels 支撐 assignment、filter、reporting、status grouping，但自由命名會破壞長期比較。
- SRE 需要精準 selector；Manager 需要 office / ISP / region / criticality；NOC 需要 coverage 可見。
- 設計機會：default label keys、label usage view、rename / merge / deprecate workflow、selector impact preview。

### Theme 8：命名可能造成 document / stamp / certificate 誤解

- Netstamp 可能讓非技術角色誤以為是文件蓋章、憑證或 timestamp 工具。
- Evidence snapshot 若未來要做，需定位為 incident evidence，而不是 certificate 產品。
- 設計機會：首頁與 onboarding 第一屏明確寫 network observability from probes you control；FAQ 清楚否定 document stamping 定位。

### 跨主題優先機會

- P0：First-run checklist + empty states。
- P0：Status taxonomy + alert explanation。
- P0：Incident detail deep links + summary copy。
- P1：Public status alias / redaction / generated time。
- P1：Probe install troubleshooting + secret hygiene。
- P1：Label governance + selector preview。
- P1：Audit log for sensitive actions。
- P2：Methodology / confidence layer。
- P2：Executive trend reports。
- Research spike：Evidence snapshot export。
