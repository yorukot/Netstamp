# 08. Affinity Synthesis：Synthetic 訪談主題整理

> 本文件整理 `07-simulated-interviews.md` 的假設式 / synthetic 訪談模擬。它不是基於真實訪談資料，不可被視為實際使用者研究結論；應作為後續真實訪談、可用性測試與產品假設優先排序的輸入。

## Evidence Baseline

本 synthesis 只採用 repo 內已讀 evidence 支持的產品脈絡：Netstamp 是 self-hosted network observability / monitoring app，核心包含 probes、checks、assignments、results、alerts、incidents、notifications、public status pages、OpenAPI、probe runtime 和 install assets。沒有 evidence 支持 document stamping、certificate stamping 或 timestamp certificate，因此相關內容只作為「命名誤解 / 需釐清」的研究問題。

## Synthetic Persona Coverage

| ID  | Persona                           | 主要視角                         | 對產品成功的判準                                        |
| --- | --------------------------------- | -------------------------------- | ------------------------------------------------------- |
| P1  | Senior SRE                        | 技術深挖、高頻 incident response | 快速定位 affected probe/check/time window，alert 可追溯 |
| P2  | NOC 值班工程師                    | 高頻 triage、交班                | 清楚區分 failure / stale / no data，能複製事件摘要      |
| P3  | IT Operations 經理                | 管理、趨勢、跨區治理             | 可按 office/ISP/region 看趨勢和報表                     |
| P4  | Support Lead                      | 非技術、客戶溝通                 | 內外部狀態說法清楚，不暴露敏感細節                      |
| P5  | Security / Compliance Owner       | 高風險導入、稽核                 | Secret lifecycle、audit log、RBAC、safe defaults        |
| P6  | External Incident Reviewer        | 外部驗證、RCA                    | 事件 evidence 有 timestamp、method、history             |
| P7  | First-time Self-hosting Evaluator | 首次導入、first value            | 30 分鐘內部署 probe、建立 check、看到 result            |
| P8  | Network Architect                 | 方法可信度、測量限制             | Metrics definitions、raw details、confidence layer 透明 |

## Theme 1：Probe / Check / Assignment 心智模型是採用門檻

### Affinity observations

- P1 理解 assignment，但要求 UI 明確顯示 check 跑在哪些 probes。
- P2 不想理解 selector，只要知道「這個 check 目前在哪裡跑」。
- P7 第一次使用時可能卡在 create probe -> install agent -> heartbeat -> create check -> assignment -> insight 的順序。
- P8 要求每個 result 能追溯到 probe、check config、measurement parameters。

### Interpretation

Netstamp 的 domain model 對技術使用者合理，但對第一次使用者和非深度網路角色有明顯認知成本。若 UI 直接暴露 selector / assignment 概念，初次成功率可能下降；但若過度隱藏，又會讓 SRE / architect 不信任結果。

### Product implications

- 使用分層揭露：初次流程以「選全部 probes / 手動選 probes」開始，進階再暴露 label selector / advanced JSON。
- 在 Check detail、Insight、Alert incident 都固定顯示 effective assignments。
- 建立 selector preview / coverage diff，顯示新增、移除、未匹配 probes。
- Empty states 應解釋下一步，而不是只顯示空列表。

### Validation questions

- 使用者能否在沒有文件協助下說出 probe、check、assignment 的差異？
- Selector preview 是否能降低錯配風險？
- First-run checklist 是否能讓新使用者在 30 分鐘內達到 first result？

## Theme 2：可信度來自「狀態拆解」，不是只顯示紅綠燈

### Affinity observations

- P1 要求 alert 顯示 samples、window、probe health、last value。
- P2 擔心 probe stale 和 service failure 看起來一樣紅。
- P6 認為 status page 只有 degraded 文字不足以作為 evidence。
- P8 強調 ping、TCP、traceroute 都有 measurement limitations，不能過度詮釋。

### Interpretation

Netstamp 的核心價值是 controlled viewpoints，但 controlled 不等於 automatically trusted。使用者需要知道紅色狀態背後是 timeout、packet loss、TCP failure、no data、stale probe、insufficient samples，還是 route partial。可信度需要在 UI 裡可見。

### Product implications

- 建立 status taxonomy：successful、partial、timeout、error、stale、no result、insufficient samples、probe offline。
- Alert / incident detail 顯示 evaluation summary：metric、threshold、operator、value、samples、minSamples、window、lastEvaluatedAt、lastTriggeredAt。
- Insight charts 顯示 sample count、staleness、probe health、check config。
- Traceroute / topology 視覺化要標示 partial、timeout、unknown hop 和 confidence。

### Validation questions

- 哪些 metadata 最能提升 SRE 對 alert 的信任？
- 非技術使用者是否能正確理解 stale/no data/failure？
- Network experts 是否接受目前 metrics definitions 和 chart rollup？

## Theme 3：Incident flow 需要同時支援 triage、deep dive、對外說明

### Affinity observations

- P1 要從 incident 直接跳到同一 probe/check/time window 的 insight。
- P2 要 one-click incident summary 和交班資訊。
- P4 要 customer-safe wording。
- P6 要 incident timeline 和可存檔 evidence。

### Interpretation

Alert incident 不是單一頁面問題，而是多角色 workflow hub。SRE 需要 deep evidence，NOC 需要 triage summary，Support 需要對外語言，External reviewer 需要 timeline / generated timestamp / measurement method。

### Product implications

- Incident detail 應包含多層資訊：summary、affected scope、evaluation、timeline、related insight links、notification history。
- 支援 copy incident summary，並可依 internal / customer-safe 模板輸出。
- Public status page 應能掛載 active incidents 和 resolved history。
- Incident 和 Insight 之間要有 stable deep links。

### Validation questions

- Incident summary 最小必要欄位是什麼？
- Support 是否能使用 customer-safe copy 而不需工程師翻譯？
- External stakeholder 需要 PDF、JSON、permalink 或 screenshot evidence 哪一種？

## Theme 4：Public status pages 是信任介面，不只是公開版 dashboard

### Affinity observations

- P2 要內外部狀態分離，避免細節外洩。
- P4 需要 aliases 和非技術狀態定義。
- P5 要最小暴露與 publish checklist。
- P6 要 generated time、measurement methodology、incident history。

### Interpretation

Public status page 同時承擔透明度與風險。顯示太少會像公關文字，顯示太多可能洩漏內部 topology、target、probe location、provider 資訊。設計上需要明確支援 audience、redaction、alias、methodology。

### Product implications

- Status page element 支援 public title / description / alias，與 internal check/probe name 分離。
- Redaction presets：public、partner、internal。
- 顯示 generated timestamp、status definitions、chart range、measurement method summary。
- Publish / enable status page 時提供敏感資訊 checklist。

### Validation questions

- 哪些欄位對外顯示會造成 security concern？
- 客戶需要看到 probe region 到什麼精度才覺得可信？
- Public incident history 是否應可編輯、鎖定或匯出？

## Theme 5：Self-hosted adoption 需要 operational readiness，不只是 quick start

### Affinity observations

- P3 關心備份、升級、資料保留、維運成本。
- P5 關心 deployment hardening、secret storage、audit、network egress。
- P7 需要 installer troubleshooting 和 first value path。
- P1 關心 controller / DB 壓力下 alert evaluation 是否落後。

### Interpretation

Self-hosted 是 Netstamp 的差異化，但也把可靠性、安全、升級、備份責任推給使用者。README quick start 能啟動試用，但導入 production 需要更完整的 readiness path。

### Product implications

- 補齊 production checklist：HTTPS、secrets、backup、retention、upgrade、reverse proxy、DB storage、observability。
- Probe installer 顯示安全說明與 troubleshooting。
- 提供 health / readiness / version / migration status 的 admin view。
- Alert evaluation / notification outbox 需提供 lag / failure observability。

### Validation questions

- 哪些 production hardening items 是導入阻塞？
- 使用者是否願意用 Docker Compose 作為 production 起點？
- Probe agent install 失敗最常見原因是權限、network egress、secret、還是 controller URL？

## Theme 6：Security、RBAC、audit 是 enterprise / regulated 使用者的採購門檻

### Affinity observations

- P5 要 audit log、secret lifecycle、scoped service accounts。
- P3 擔心 labels 和 reports 被不同地區任意改動。
- P2 擔心值班者誤刪 rules / notifications。
- P4 / P6 關注 external view 不暴露內部資訊。

### Interpretation

Netstamp 的現有 roles / project workspaces 是基礎，但 enterprise adoption 可能需要更細緻的權限、稽核與保護。尤其 probe secret、notification destination、public status page、members/roles 是 audit-worthy actions。

### Product implications

- Audit log scope：probes、checks、labels、alerts、notifications、status pages、members、project settings。
- Dangerous action confirmation：delete probe/check/rule/status page、rotate secret、enable public page。
- Scoped API tokens / service accounts 作為 API 整合基礎。
- Secret masking、rotation reminder、last used / created by metadata。

### Validation questions

- 哪些操作必須進 audit log 才能通過合規審查？
- 現有 owner/admin/editor/viewer 是否足夠？
- Notification secrets 和 webhook URLs 需要何種 masking / testing UX？

## Theme 7：Labels 是 powerful primitive，也是治理風險

### Affinity observations

- P1 需要 labels / selectors 精準控制 checks。
- P3 想用 office、ISP、region、criticality 做管理報表。
- P2 不想處理 selector 語法，但需要看到 coverage。
- P3 / P5 擔心 labels 任意改動讓歷史不可比或造成稽核問題。

### Interpretation

Labels 同時支撐 assignment、filter、reporting、status grouping。若沒有治理，短期方便，長期會讓資料難以比較。產品需要把 labels 從 freeform metadata 提升為可管理的資訊架構。

### Product implications

- 建議 default label keys：region、site、provider、network、environment、criticality、owner。
- Label usage view：哪些 probes/checks/status elements/rules 使用某 label。
- Label rename / merge / deprecate workflow。
- Selector impact preview 和 historical caveat。

### Validation questions

- 使用者自然會建立哪些 label keys？
- Label governance 對小團隊是否過重？
- Reporting / saved views 是否能推動 labels 規範化？

## Theme 8：命名可能造成 document / stamp / certificate 誤解

### Affinity observations

- P4 表示 Netstamp 第一次看可能像文件蓋章或憑證 timestamp。
- P6 對 evidence snapshot / verifiable evidence 有興趣，但這不等於 certificate product。
- Repo evidence 明確指向 network observability，未支持 document stamping。

### Interpretation

「Netstamp」命名可保留，但首頁、docs intro、first-run onboarding 必須在第一屏明確說明 network observability from probes you control。任何 evidence snapshot / hash idea 都應被定位為 incident evidence export，而不是文件憑證或法務級 timestamp certificate。

### Product implications

- 首頁 headline / subtitle 固定包含 network observability、probes、self-hosted。
- First-run copy 避免只說 stamp / proof / certificate。
- Docs FAQ 加入「Is Netstamp a certificate or document timestamping tool?」並清楚否定。
- 若未來做 evidence snapshot，命名為 incident evidence snapshot，避免 certificate language。

### Validation questions

- 新訪客看首頁 5 秒後是否能說出 Netstamp 是 network monitoring？
- 搜尋流量是否帶來 document stamping 相關誤入？
- Evidence snapshot 是否會造成產品定位混淆？

## Cross-Theme Priority Matrix

| Opportunity                                      | 影響角色       | User value | Risk reduction | Suggested priority |
| ------------------------------------------------ | -------------- | ---------- | -------------- | ------------------ |
| First-run checklist + empty states               | P7, P1, P3     | 高         | 中             | P0                 |
| Status taxonomy + alert explanation              | P1, P2, P6, P8 | 高         | 高             | P0                 |
| Incident detail deep links + summary copy        | P1, P2, P4, P6 | 高         | 中             | P0                 |
| Public status alias / redaction / generated time | P4, P5, P6     | 高         | 高             | P1                 |
| Probe install troubleshooting + secret hygiene   | P5, P7         | 高         | 高             | P1                 |
| Label governance + selector preview              | P1, P2, P3, P5 | 中         | 高             | P1                 |
| Audit log for sensitive actions                  | P5, P3         | 中         | 高             | P1                 |
| Methodology / confidence layer                   | P1, P6, P8     | 中         | 高             | P2                 |
| Executive trend reports                          | P3             | 中         | 中             | P2                 |
| Evidence snapshot export                         | P6, P5, P1     | 中         | 中             | Research spike     |

## Suggested Research Plan

### Round 1：Concept validation

目標：確認 Netstamp positioning、domain model、first value path 是否清楚。

- 對象：3 位 SRE / DevOps、2 位 first-time self-hosting users、1 位 IT manager。
- 方法：展示 README + clickable prototype / current app，要求受測者說出產品用途並完成 first-run tasks。
- 成功指標：受測者能正確描述 Netstamp 是 network observability；能預測 probe/check/assignment 流程；能在無協助下找到下一步。

### Round 2：Incident workflow test

目標：測試 alert incident detail、insight deep link、summary copy、status page external view。

- 對象：2 位 SRE、2 位 NOC、2 位 Support / external stakeholder。
- 方法：給定 packet loss / TCP failure / stale probe 情境，觀察 triage 和對外說明。
- 成功指標：能正確區分 failure 類型；能產出一致 incident summary；不誤把 stale/no data 當服務故障。

### Round 3：Trust and compliance review

目標：找出 enterprise adoption blockers。

- 對象：2 位 Security / Compliance、2 位 Network Architect。
- 方法：review deployment docs、probe install flow、API / alert / public status model。
- 成功指標：列出必備 audit / RBAC / secret / methodology requirements，並排序導入阻塞程度。

## Open Product Questions

- Probe offline alert 是否應進入 alerting 的下一版？若是，如何避免和 result-submit-driven alert 混淆？
- Public status page 是否要提供 resolved incident history？歷史是否可編輯？
- 是否需要 raw result retention 設定，讓 high-trust users 可查原始資料？
- Labels 是否應加入 required keys 或 schema，還是保持自由？
- API tokens / service accounts 是否是 OpenAPI 整合前提？
- Evidence snapshot 是否值得做？若做，如何避免被理解成 certificate / document timestamping？

## Product Principles Suggested by Synthesis

- **Measured, not magical:** 明確呈現測量方法與限制，不把 active probing 結果包裝成絕對真相。
- **Viewpoint-first:** Probe 是觀測點；UI 應幫使用者理解「從哪裡看」。
- **Explain every alert:** 每個 alert 都應能回答為什麼觸發、依據哪些 samples、何時評估、何時可再通知。
- **Safe external sharing:** 對外狀態頁要先保護敏感資訊，再逐步增加可信度資訊。
- **Self-hosted with guardrails:** 快速啟動之外，必須提供 production、security、backup、upgrade 的導入護欄。
