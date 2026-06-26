# Prioritized Recommendations

Scoring is a pragmatic RICE-like synthesis from repository evidence, UX audit severity, synthetic hypotheses, and competitive pattern research.

Impact:

- 5 = directly changes activation/trust/adoption for primary users.
- 4 = materially improves operational or governance workflow.
- 3 = useful but not blocking.

Effort:

- 1 = content/UI-only or small UI.
- 2 = moderate UI changes with existing APIs.
- 3 = new UI model or API extension.
- 4 = cross-cutting backend/frontend/data changes.
- 5 = strategic platform expansion.

Confidence:

- High = strongly code-backed UX gap.
- Medium = code-backed plus inference/pattern.
- Low = mainly hypothesis.

## Top 10

| Rank | Feature | User problem | Target user | Evidence / rationale | Impact | Effort | Priority | Why now |
| --: | --- | --- | --- | --- | --: | --: | --- | --- |
| 1 | Guided First-Run Checklist | Users can create project/probe but may not reach first check/result/alert. | First-time self-host evaluator, SRE | Onboarding opens probe wizard only; README value depends on measurements. | 5 | 2 | Now | Time-to-first-value is the adoption gate. |
| 2 | Netstamp Does / Does Not Do | Product name and prompt language can imply document/certificate stamping. | All users, product team, evaluators | No code evidence for document/cert verification; README says network observability. | 5 | 1 | Now | Prevents wrong UX and roadmap discussion. |
| 3 | Empty States With Operational Next Actions | Empty tables/insight states do not teach next step. | First-time users, NOC | Probes/checks/alerts/status flows rely on empty labels and route navigation. | 4 | 1 | Now | Low effort, large activation improvement. |
| 4 | Error Recovery Playbooks | Probe install, API errors, no data, permissions and read-only failures need action. | SRE, self-host operator, viewer/editor | ApiError/toast exists, but recovery is generic. | 4 | 2 | Now | Reduces support load and failure drop-off. |
| 5 | Role Permission Visibility | Users discover limitations by disabled controls or API errors. | Owner/admin/editor/viewer | Backend role policy exists; UI needs human explanation. | 4 | 2 | Now | Reduces confusion and destructive mistakes. |
| 6 | Result Trust / Proof Page | Charts do not fully explain measurement source, sample count, freshness and failure limits. | SRE, external reviewer, compliance | Results/status/incidents exist; trust layer missing. | 5 | 3 | Next | Netstamp's differentiated value is evidence from controlled probes. |
| 7 | Incident Review Workspace | Incidents are listed/detail-viewed but not full triage/RCA workflows. | SRE, NOC, Support | Incident drawer shows timeline but lacks deep links/actions/copy summary. | 5 | 3 | Next | Alerts without triage flow create operational friction. |
| 8 | Notification Delivery Status Center | Test notification exists but ongoing delivery confidence is not visible. | On-call, admin | Notification test returns success/failure; no history center evidence. | 4 | 3 | Next | Alert trust depends on knowing messages were delivered. |
| 9 | Audit Trail / Event History | Sensitive actions lack visible accountability. | Security/compliance, admin | Roles/secrets/public status operations exist; audit UI not evidenced. | 4 | 4 | Next | Required for regulated/self-host adoption. |
| 10 | Public Status Trust & Redaction Layer | Public pages can expose target/probe details and lack methodology/disclaimer. | Support, external reviewer, security | Public assignment rows show target/probe; generatedAt exists. | 5 | 3 | Next | Public status is the external trust surface. |

## Full Recommendation Table

| ID | Feature name | User problem | Target user | Evidence / rationale | Expected UX impact | Risk reduced | Dependencies | MVP scope | Nice-to-have scope | Impact | Effort | Suggested flow placement | FigJam color | Confidence |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --: | --: | --- | --- | --- |
| F01 | Guided First-Run Checklist | First value is not guaranteed after account/project/probe creation. | First-time user, SRE | Onboarding success opens probe wizard only. | Faster activation and fewer abandoned installs. | Setup confusion. | Existing project/probe/check/result APIs. | Checklist: project, probe heartbeat, check, first result, alert/status optional. | Progress persistence and sample templates. | 5 | 2 | Onboarding success, dashboard empty state | Dark / priority | High |
| F02 | Netstamp Does / Does Not Do | Users may expect document timestamp/certificate verification. | All users | README/network evidence; no document evidence. | Correct product expectation. | Strategy drift and false claims. | Content only. | One panel in onboarding/docs/status. | Interactive glossary. | 5 | 1 | Landing, login/register, docs, app empty states | Pink + dark | High |
| F03 | Result Trust / Proof Page | Users cannot easily explain a result as evidence. | SRE, external reviewer | Results/status/incident evidence exists but fragmented. | Stronger trust and RCA shareability. | Misinterpretation. | Result APIs, status data. | Source, time, sample count, status, raw API link. | Baseline diff and signed snapshot hypothesis. | 5 | 3 | Result/Insight/Incident/Public status | Dark / priority | High |
| F04 | Proof Explanation Panel | Measurement terms and limits are not explicit. | SRE, support, compliance | Traceroute partial/stale/no result states need explanation. | Better comprehension. | Overstated certainty. | Content model. | Inline explanations per state. | Methodology docs linked to APIs. | 4 | 2 | Insight, Public status, Incident detail | Purple | High |
| F05 | Export / Share Report | Teams need customer/RCA evidence snapshot. | Support, SRE, external reviewer | No export evidence; status/incident data exists. | Easier communication. | Ad hoc screenshots. | Result/status APIs. | Copy summary + share link. | PDF/CSV, immutable snapshot. | 4 | 3 | Incident, Result proof, Public status | Purple | Medium |
| F06 | Incident Review Workspace | Incidents lack full operational lifecycle. | SRE, NOC | Incident detail has what happened/timeline/notifications only. | Faster triage and clearer RCA. | Alert fatigue and handoff loss. | Incident API extensions for ack/resolve if implemented. | Deep links, summary copy, related results, timeline. | Ack/resolve/postmortem. | 5 | 3 | Alerts > Incident drawer/page | Dark / priority | High |
| F07 | Notification Delivery Status Center | Users cannot audit real alert delivery. | On-call, admin | Test notification exists; delivery history not evidenced. | Confidence in alerts. | Missed incidents. | Notification/outbox data. | Last test, last incident delivery, failure reason. | Retry policy, provider diagnostics. | 4 | 3 | Alerts > Notifications | Amber/Purple | Medium |
| F08 | Error Recovery Playbooks | Errors do not always explain next step. | All users | ApiError/toasts and blank guard loading states. | Lower support burden. | Abandonment. | Existing ApiError, request IDs. | Common playbooks: no heartbeat, no data, 403, read-only, public 404. | Diagnostic bundle export. | 4 | 2 | Global error/toast/empty states | Purple | High |
| F09 | Empty States With Operational Next Actions | Empty screens do not teach product flow. | First-time users | Empty labels in tables, dashboard lacks setup progress. | More complete activation. | Dead-end first session. | Existing routes/actions. | CTA and explanation per feature. | Context-aware recommendations. | 4 | 1 | Probes, Checks, Insight, Alerts, Status | Purple | High |
| F10 | Search / Filter / History Hub | Larger projects need cross-resource history. | SRE, manager | Existing filters are page-specific. | Faster findability. | Operational delay. | Resource APIs and query params. | Unified recent activity/search on project. | Saved filters, audit fusion. | 3 | 4 | Dashboard/sidebar | Purple | Medium |
| F11 | Audit Trail / Event History | Sensitive actions are not visibly auditable. | Admin, compliance | RBAC/secrets/public status actions exist. | Enterprise trust. | Unaccountable changes. | Backend event model. | Actor/action/resource/time for sensitive events. | Export and retention policy. | 4 | 4 | Project settings, resource details | Dark / priority | Medium |
| F12 | Role Permission Visibility | Users do not know what their role allows. | All project members | Backend `Can` policy exists. | Clear expectations. | 403 frustration and unsafe workarounds. | Role data already available. | Role badges, disabled tooltip, permission matrix. | Request access workflow. | 4 | 2 | Members, toolbar actions, settings | Purple | High |
| F13 | Security / Privacy Messaging | Probes, geocoding, public status, Gravatar and secrets need clearer disclosures. | Security, admin, support | Code uses registration secret, Nominatim, Gravatar, public assignment data. | Adoption confidence. | Privacy/security blockers. | Content + UI surfaces. | Short disclosures at risky moments. | Full data inventory page. | 4 | 2 | Probe wizard, account, public status | Purple | High |
| F14 | Legal / Compliance Disclaimer For Shared Evidence | Public status/result proof can be mistaken for SLA/legal certification. | External reviewer, support | Public status and result evidence exist; legal guarantees not evidenced. | Prevents overclaim. | Legal/compliance misinterpretation. | Content review. | Non-SLA, measurement-limits disclaimer. | Configurable org disclaimer. | 4 | 2 | Public status, export/report | Gray/Purple | Medium |
| F15 | Sample Project / Demo Learning Mode | First-time users need data before deploying agents. | Evaluator | Demo/read-only exists, but sample flow needs clarity. | Faster learning. | Install friction. | Demo fixtures or seeded data. | Sample project read-only walkthrough. | Interactive guided sandbox. | 3 | 3 | Demo login/dashboard | Purple | Medium |
| F16 | Admin Visibility / Controller Health | Self-host operators need to know controller/db/SMTP/OTEL/readiness status. | Self-host operator | Server has readiness/metrics/config; UI not evidenced. | Safer operations. | Misconfiguration. | System/status APIs. | Admin health page. | Upgrade and backup guidance. | 4 | 4 | Project/admin settings or system page | Purple | Medium |
| F17 | In-Product Developer/API Docs Shortcuts | API users need IDs and examples in context. | Developer/SRE | OpenAPI exists. | Faster automation. | Contract drift confusion. | Existing docs/OpenAPI. | Copy ID/curl/View API on resource detail. | Terraform/provider examples. | 3 | 2 | Resource details, docs links | Slate/Purple | High |
| F18 | Accessibility And Mobile Operations Pass | Dense tables/drawers reduce incident usability on small screens. | NOC, mobile on-call | Tables have large min widths; loading states return null. | Better operations under pressure. | Accessibility failures. | UI refactor. | Mobile cards, loading states, focus/motion pass. | Offline/low bandwidth mode. | 4 | 3 | All operational pages | Purple | Medium |

## Impact / Effort Matrix

| Quadrant                      | Features                                                                                                                          |
| ----------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| High impact / Low effort      | F02 Does/Does Not Do, F09 Empty States, F12 Role Permission Visibility, F13 Security/Privacy Messaging                            |
| High impact / Medium effort   | F01 First-Run Checklist, F03 Result Trust Page, F06 Incident Review, F08 Error Recovery, F14 Disclaimer, F18 Accessibility/Mobile |
| High impact / High effort     | F11 Audit Trail, F16 Admin Health                                                                                                 |
| Medium impact / Medium effort | F04 Proof Explanation, F05 Export/Share Report, F07 Delivery Status, F15 Demo Learning, F17 API Shortcuts                         |
| Medium impact / High effort   | F10 Search/History Hub                                                                                                            |

## Now / Next / Later Roadmap

### Now: 0-4 weeks

- F02 Netstamp Does / Does Not Do。
- F01 Guided First-Run Checklist。
- F09 Empty States With Operational Next Actions。
- F08 Error Recovery Playbooks。
- F12 Role Permission Visibility。

Objective: reduce activation failure and product misunderstanding.

### Next: 4-10 weeks

- F03 Result Trust / Proof Page。
- F04 Proof Explanation Panel。
- F06 Incident Review Workspace。
- F07 Notification Delivery Status Center。
- F13 Security / Privacy Messaging。
- F14 Legal / Compliance Disclaimer。
- Public status trust/redaction layer as part of F03/F13/F14。

Objective: turn measurements into explainable, shareable, safer operational evidence.

### Later: 10-16+ weeks

- F05 Export / Share Report。
- F10 Search / Filter / History Hub。
- F11 Audit Trail / Event History。
- F15 Sample Project / Demo Learning Mode。
- F16 Admin Visibility / Controller Health。
- F17 In-Product Developer/API Docs Shortcuts。
- F18 Accessibility And Mobile Operations Pass。

Objective: support scale, governance, enterprise adoption, and automation.

## Validation Needed Before Build

| Recommendation          | What must be validated                                                                                  |
| ----------------------- | ------------------------------------------------------------------------------------------------------- |
| First-run checklist     | Which activation milestone best predicts adoption: heartbeat, first result, first alert, public status? |
| Result Trust page       | What proof fields do SREs/external reviewers need to trust a result?                                    |
| Public status redaction | What target/probe data is acceptable to expose publicly?                                                |
| Incident workspace      | Do users need ack/resolve/manual incident actions in Netstamp, or only RCA/deep links?                  |
| Audit trail             | Which events are compliance-critical?                                                                   |
| Admin health            | Should system health be global admin-only, project owner, or self-host operator via config?             |
