# Netstamp UX Research Master Report

## Executive Summary

Netstamp 目前的真實產品邊界非常清楚：它是 open-source、self-hosted network observability 產品，讓使用者部署自己控制的 probes，從真實位置監測 latency、packet loss、routes、TCP reachability、probe health，並透過 projects、labels、checks、results、alerts、public status pages 與 OpenAPI 形成可營運的觀測系統。

這份研究最重要的校正是：目前 repository 沒有 document stamping、certificate issuing、file notarization 或 verification portal 的實作 evidence。任何「stamp / certificate / document verification」都必須在 FigJam 上標示為命名誤解風險或 product strategy hypothesis，不能寫成既有功能。

最大的 UX 機會不在新增更多 check type，而是在建立一條可信的 activation-to-evidence loop：

1. 讓第一次使用者理解 `project -> probe -> check -> assignment -> result -> alert/status` 的心智模型。
2. 讓使用者在 30 分鐘內完成 first project、first probe heartbeat、first check、first result、first alert/status proof。
3. 讓每個 result / incident / status page 都能回答「從哪裡量、何時量、樣本夠不夠、失敗原因是什麼、這個狀態能不能對外說」。
4. 把 sensitive operations、RBAC、secret、public redaction、audit trail 做成可見的信任基礎。

## Research Scope

### Included

- Repository code、README、routing、UI components、API TypeSpec、server router、permission model、feature flags、alerts、status pages、probe install flow。
- Existing FigJam board content readback，包含既有 Functions mind map 與 templates。
- Synthetic interviews，明確標示為 simulated，不當作真實使用者研究。
- Competitive/pattern research，聚焦 observability、synthetic monitoring、private probes、status/incident、notification、self-host trust。

### Excluded / No Evidence

- Document stamping、certificate issuing、digital signature verification、file notarization。
- Billing、plans、usage quota、subscription。
- Measurement export/import、CSV/PDF report download as current feature。
- DNS/HTTP check as current feature。
- Global admin/org hierarchy beyond project roles。
- Real user interview data。

## Methodology

| Method                        | Output                                                   | Evidence strength                                 |
| ----------------------------- | -------------------------------------------------------- | ------------------------------------------------- |
| Codebase feature inventory    | `01-feature-inventory.md`                                | High for current-state facts                      |
| Persona/JTBD synthesis        | `02-users-jtbd-scenarios.md`                             | Medium; code-backed inference                     |
| Journey and service blueprint | `03-journey-service-blueprint.md`                        | High for current flow, Medium for emotional state |
| Heuristic UX audit            | `04-heuristic-ux-audit.md`                               | High for code-backed issues                       |
| Competitive pattern research  | `05-competitive-patterns.md`                             | Medium for pattern rationale                      |
| FigJam structure planning     | `06-figjam-structure.md`                                 | High for board organization                       |
| Synthetic interviews          | `07-simulated-interviews.md`, `08-affinity-synthesis.md` | Low as evidence, Medium as hypotheses             |
| Roadmap recommendations       | `09-recommended-features-roadmap.md`                     | Medium-High depending on evidence source          |

## Product Understanding

### What Netstamp Appears To Do

Netstamp is a self-hosted network monitoring and observability system. A team runs a controller, creates projects, registers probes in real networks, defines Ping/TCP/traceroute checks, assigns checks to probes through labels/selectors, collects results, reviews insights, configures alerts, and optionally publishes public status pages.

### Core Value Proposition

「從你控制的 probes 看見網路真相。」Netstamp 的價值不是只知道 service up/down，而是知道不同地點、ISP、lab、edge、private infrastructure 看到的 reachability、latency、packet loss、route behavior 與 probe health。

### Main Workflows

| Workflow                       | Current evidence                                                                | UX status                                                    |
| ------------------------------ | ------------------------------------------------------------------------------- | ------------------------------------------------------------ |
| Account/session                | Login/register/logout/current session; feature flags can disable registration.  | 基本完整；password reset 未見 evidence。                     |
| First project onboarding       | Create project, slug retry, invite viewers, open probe wizard.                  | 能開始，但沒有 first check/result/alert checklist。          |
| Probe registration             | Name, coordinates, one-time secret, install command, heartbeat detection.       | 核心強，但 troubleshooting/secret/privacy 說明不足。         |
| Probe fleet management         | List/map/detail/edit/label/rotate/delete, frequent refetch.                     | 對工程使用者可用；bulk/health diagnostics 不足。             |
| Checks and assignments         | Ping/TCP/traceroute; interval/type config; selector preview.                    | 功能完整但 selector/assignment mental model 複雜。           |
| Results/insight/topology       | Ping/TCP series/insight, traceroute runs/topology, time range.                  | 分析能力強；需要 result detail/proof layer。                 |
| Alerts/incidents/notifications | Rules, incidents, notification CRUD/test, webhook/Slack/Discord/Telegram/email. | 有營運骨架；缺少 incident lifecycle/action/delivery status。 |
| Public status pages            | Slug, title/description, elements, metrics/charts, generatedAt, open incidents. | 是信任介面；需要 redaction/methodology/disclaimer。          |
| Members/RBAC                   | Owner/admin/editor/viewer project roles and invite flow.                        | Backend model 清楚；UI 需更全站可理解。                      |
| API/docs/self-host             | OpenAPI, `/docs`, Docker Compose, install assets.                               | 對工程採用有利；需 in-product API shortcuts。                |

## Existing Feature Inventory Summary

### Core Workflow

- Dashboard overview。
- Probe create/install/list/map/detail/secret rotation/delete。
- Checks create/edit/duplicate/delete/batch delete：Ping/TCP/traceroute。
- Label CRUD and usage。
- Insight/result views for ping/tcp/traceroute。
- Alert rules/incidents/notifications。
- Public status pages and public runtime route。

### User / Account / Auth

- Register/login/logout/session me。
- First project onboarding。
- Account settings: profile/email/password/pending invites。
- Project members/roles/invites。

### Admin / Settings / Deployment

- Project settings：name/slug/delete/leave。
- Project switcher and create project modal。
- Demo/read-only flags。
- API docs/OpenAPI。
- Self-host Docker deployment and observability stack.

### Error / Empty / Permission

- API Problem Details mapped into ApiError。
- Toasts, loading/empty labels, route guards。
- Role-based backend permission model。
- Demo/read-only mutation blocks。

### Integration

- Alert notifications: webhook, Slack, Discord, Telegram, email。
- Agent install scripts/binaries。
- OpenAPI contract。
- Tracking consent providers in docs/app context。

## Important No-Evidence Findings

| Area                                    | Finding                                        | Implication                                                                               |
| --------------------------------------- | ---------------------------------------------- | ----------------------------------------------------------------------------------------- |
| Document/stamp/certificate verification | No code evidence.                              | Treat as naming confusion, not product capability.                                        |
| DNS/HTTP checks                         | Not supported by current check union evidence. | Fix copy or mark planned; do not claim current support.                                   |
| Billing/limits                          | No route/API/schema evidence.                  | Do not add billing UX to current-state map.                                               |
| Export/import                           | No current measurement export/import evidence. | Recommended feature only.                                                                 |
| LINE notification                       | Existing FigJam idea, no code evidence.        | Mark as idea/hypothesis.                                                                  |
| Global admin                            | Project-scoped roles only.                     | Admin/operator persona should be self-host or project owner/admin, not global SaaS admin. |

## User Types And JTBD

| User type | Evidence basis | Primary JTBD | Success criteria | Gaps |
| --- | --- | --- | --- | --- |
| SRE / Platform engineer | README use cases, probes/checks/results/alerts | When latency or packet loss changes, I need to know which probe/check/network path is affected so I can act fast. | Finds affected probe/check/time window; trusts raw evidence; can share incident summary. | Result proof page, deep links, incident workspace. |
| NOC / on-call operator | Alerts/incidents/dashboard/status pages | When an alert fires, I need to triage status and know whether it is target down, probe stale, or no data. | Clear state taxonomy; next action; notification confidence. | State explanation, ack/resolve, delivery status. |
| IT Operations manager | Projects/members/status pages | When managing distributed offices/providers, I need trend and status views by location/team. | Can see health by region/ISP/project; export/share status. | Executive trend reporting, filters, reports. |
| Self-host operator | README quickstart, Docker, config, install assets | When deploying Netstamp, I need safe configuration and controller health visibility. | Confident secrets/HTTPS/SMTP/OTEL are configured. | Admin health page, security checklist. |
| Security / compliance owner | RBAC, secrets, public status exposure | When approving adoption, I need control over roles, secrets, audit trail, public data exposure. | Audit log, redaction, permission visibility, disclaimers. | Audit trail, redaction, privacy messaging. |
| Support/CS lead | Public status pages | When customers ask if service is down, I need a status page I can share without exposing internals. | Plain-language component status and freshness. | Alias/redaction, public explanation, subscriber model. |
| External reviewer / recipient | Public status page and incident evidence | When reviewing an incident/RCA, I need to verify what was measured and when. | GeneratedAt, measurement source, history, limits. | Proof explanation, export snapshot. |
| Developer/API user | OpenAPI/docs | When automating setup, I need stable API docs, IDs, examples and scripts. | Copy IDs/curl; matches OpenAPI. | In-product API shortcuts. |

## Current-State Journey Summary

| Phase | User goal | Current touchpoint | Pain point | Opportunity |
| --- | --- | --- | --- | --- |
| 1. Arrive | Understand what to do next | `/`, login/register redirects | `/` is redirected, little context in app shell for unauth users | Add does/does-not-do and first-run overview |
| 2. Account | Create/log in | AuthPage | Register CTA says Create project; no password recovery evidence | Fix CTA, add recovery/disabled state copy |
| 3. Project | Create workspace | Onboarding | Terminal flow creates project, invites viewer only, no role explanation | Checklist and role-aware invites |
| 4. Probe | Deploy first viewpoint | NewProbeDrawer | Location requirement/geocoding/privacy; install troubleshooting thin | Probe install diagnostics |
| 5. Check | Define measurement | Checks editor | Selector mental model and all-probes default can be risky | Safer templates and assignment impact preview |
| 6. First result | See evidence | Insight/check detail | Result detail/proof explanation not prominent | First-result handoff and proof page |
| 7. Alert | Operationalize | Alerts tabs | Incident lifecycle and delivery status incomplete | Incident review workspace |
| 8. Share | Communicate externally | Public status page | Public data redaction, source/freshness/methodology incomplete | Trust/disclaimer/redaction layer |
| 9. Govern | Manage team/security | Members/settings/project | RBAC and audit trail not surfaced everywhere | Role visibility and event history |

## Future-State Journey

| Phase | Ideal outcome | Required product change |
| --- | --- | --- |
| Orientation | User knows Netstamp is network observability from controlled probes and not document stamping. | Does/Does Not Do module on landing, onboarding, docs, and FigJam product context. |
| Activation | User completes first project, first probe heartbeat, first check, first result, first alert/status in one guided path. | Guided First-Run Checklist and empty states with next actions. |
| Measurement | User can see why a result is successful/failed/partial/no data. | Result Trust / Proof Page and proof explanation panel. |
| Triage | Operator can move from open incident to affected probes/checks/results and copy summary. | Incident Review Workspace with timeline, deep links, ack/resolve if supported. |
| Communication | Support/customer/external reviewer sees a clear public status component with freshness, source and limitations. | Public status trust layer, redaction, generatedAt, non-SLA disclaimer. |
| Governance | Admin/security sees role permissions, sensitive action history, secret lifecycle. | Role permission visibility, audit trail, secret rotation confirmation. |
| Scale | Teams can search, filter, review history and automate setup. | Search/history hub, API shortcuts, admin controller health. |

## Service Blueprint Summary

| Moment | Frontstage | Backstage | Failure modes | Needed improvement |
| --- | --- | --- | --- | --- |
| Session | Login/register/route guard | JWT cookie/session me | Blank loading, disabled registration redirect | Visible loading and disabled reason |
| Project | Onboarding, project switcher | Project service, invites, local project selection | Slug conflict, no assigned project, invite failures | Checklist, request access, invite status |
| Probe | Wizard, secret, command, heartbeat | Probe create, install assets, runtime auth, heartbeat | Host install fails, secret lost, geocode fails | Diagnostics and secret recovery guidance |
| Check | Editor, selector preview | Check CRUD, assignment service | Wrong selector, all-probes blast radius | Impact preview |
| Result | Insight charts/topology | Result query services | Missing samples/stale probe/partial traceroute | Confidence and raw evidence panel |
| Alert | Rules/incidents/notifications | Alert eval, outbox, senders | No channel, test fails, noise, no lifecycle action | Delivery status and rule templates |
| Public status | Public slug page | Public status service/charts/incidents | Disabled/not found ambiguity, data leakage | Redaction, disclaimer, source/freshness |

## Heuristic UX Audit: Highest-Value Issues

| Issue | Severity | Evidence | Recommended fix |
| --- | --- | --- | --- |
| First-run flow stops after probe heartbeat, not first result | High | Onboarding -> new probe; no full checklist | Guided First-Run Checklist |
| Dashboard does not yet support operational triage | High | Dashboard summary vs incidents/results gap | Add recent incidents, failing assignments, stale probes |
| Incident detail lacks lifecycle actions and analysis handoff | High | Incident drawer timeline but no ack/resolve/deep links | Incident Review Workspace |
| Alert scope uses raw Probe ID / Check ID fields | High | Rule editor target fields | Human-readable pickers and natural-language preview |
| Member removal and secret rotation need stronger confirmation | High | Member/probe detail actions | Confirmations with impact summary |
| DNS copy may overpromise unsupported check type | High | Check type evidence only ping/tcp/traceroute | Fix copy or mark planned |
| Public status trust explanation is thin | High | Generated time exists but no methodology/disclaimer/redaction | Public status trust layer |
| Error recovery is mostly generic toast/loading | Medium-High | ApiError/toast and blank guard states | Recovery CTA, field errors, request ID copy |
| Mobile operational tables are dense | Medium | DataTables with large min widths | Mobile summary cards and responsive detail views |
| Privacy messaging for Gravatar/geocoding/public data is weak | Medium | External avatar/geocoding/public assignment data | Privacy controls and disclosures |

## Synthetic Interview Findings

These are simulated, not real interview evidence. They should be used to design real research and stress-test product hypotheses.

| Theme                                                                  | Synthetic signal                                       | Product implication                              |
| ---------------------------------------------------------------------- | ------------------------------------------------------ | ------------------------------------------------ |
| Probe/check/assignment mental model is adoption gate                   | New users need vocabulary and setup order.             | First-run checklist and in-product education.    |
| Trust comes from decomposed status, not red/green                      | Operators need failure/stale/no data/sample reasons.   | Status taxonomy and proof explanation.           |
| Incident flow must support triage, deep dive, and external explanation | SRE/NOC/Support need different views of same incident. | Incident workspace with copyable summary.        |
| Public status is a trust interface                                     | External readers need freshness and redacted wording.  | Public status methodology/disclaimer/redaction.  |
| Self-host adoption needs operational readiness                         | Operators worry about secrets, HTTPS, SMTP, upgrades.  | Admin health and deployment checklist.           |
| Security/RBAC/audit are adoption gates                                 | Compliance users need audit and role clarity.          | Audit log and role visibility.                   |
| Labels are powerful but risky                                          | Selector mistakes can create coverage gaps.            | Selector impact preview and label governance.    |
| Netstamp naming can mislead                                            | Some may infer timestamp/document stamping.            | Explicit positioning and no-evidence board note. |

## Competitive Pattern Takeaways

The competitive research should influence patterns, not overwrite Netstamp's product identity.

| Pattern                       | Source examples                                                              | Netstamp application                                                             |
| ----------------------------- | ---------------------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| Viewpoint-first monitoring    | Grafana Synthetic Monitoring, Checkly private locations                      | Explain probe as viewpoint/location; show measurement source everywhere.         |
| Private agent onboarding      | Grafana private probes, Checkly private locations, Datadog private locations | Key -> command -> heartbeat -> first check -> first result.                      |
| Result drilldown              | Grafana dashboards, Checkly results, Datadog synthetic results               | Summary -> filters -> assignment/run detail -> raw evidence.                     |
| Incident/status communication | Better Stack, Atlassian Statuspage, Upptime                                  | Incident timeline, generatedAt, public components, maintenance/postmortem later. |
| Notification trust            | Grafana contact point test, Checkly alert channels                           | Test result, delivery history, masked secrets, used-by rules.                    |
| Self-host trust               | Uptime Kuma, Prometheus Blackbox Exporter                                    | Screenshots/demo, copyable quickstart, raw API/debug links.                      |

## Trust, Security, And Verification UX

For Netstamp, "verification" should mean verifying a network measurement, not verifying a stamped document. A trustworthy Netstamp result should answer:

- What was measured: check type, target, config, interval.
- Where it was measured from: probe, location, labels, network metadata if safe.
- When it was measured: startedAt, finishedAt/generatedAt, time range, freshness.
- How many samples support it: sample count, missing data, stale status.
- What the status means: successful, timeout, error, partial, stale, no result.
- What changed: compare latest failed run against recent successful baseline.
- What is not guaranteed: non-SLA, not legal proof, partial/traceroute limits, probe-offline limitations.
- What can be shared: redacted status, public alias, report snapshot, API URL.

## Opportunity Areas

| Opportunity           | Why it matters                                  | Candidate features                                     |
| --------------------- | ----------------------------------------------- | ------------------------------------------------------ |
| Onboarding clarity    | Reduces time-to-first-value and support burden. | Guided checklist, sample project, empty states.        |
| Measurement trust     | Makes Netstamp credible for incident/RCA.       | Result proof page, raw evidence, confidence layer.     |
| Error recovery        | Turns failure into next action.                 | Playbooks, request ID, troubleshooting.                |
| Sharing/export        | Supports customer/support/external review.      | Evidence snapshot, public status alias, report export. |
| Alert operations      | Moves from configuration to triage workflow.    | Incident workspace, notification delivery center.      |
| Governance/security   | Supports enterprise/self-host approval.         | RBAC visibility, audit log, secret lifecycle.          |
| IA/content clarity    | Reduces wrong expectations from "stamp" name.   | Does/Does Not Do, terminology cleanup.                 |
| Accessibility/mobile  | Makes ops usable under incident pressure.       | Mobile cards, loading states, focus/motion pass.       |
| Developer/API support | Fits engineering audience.                      | Copy IDs/curl, OpenAPI shortcuts, examples.            |

## Top 10 Recommended Features

| Rank | Feature                                    | Why now                                                                                              |
| ---: | ------------------------------------------ | ---------------------------------------------------------------------------------------------------- |
|    1 | Guided First-Run Checklist                 | Current activation stops before first result; product value is not proven until measurement appears. |
|    2 | Netstamp Does / Does Not Do                | Prevents document/certificate/stamp misunderstanding and DNS/HTTP overclaim.                         |
|    3 | Result Trust / Proof Page                  | Converts charts into shareable, explainable evidence.                                                |
|    4 | Error Recovery Playbooks                   | Probe install, API errors, no data, read-only mode need actionable recovery.                         |
|    5 | Empty States With Operational Next Actions | First-time users need next action from Probes/Checks/Insight/Alerts/Status.                          |
|    6 | Incident Review Workspace                  | Alerts exist, but triage and RCA require timeline, deep links and summary.                           |
|    7 | Notification Delivery Status Center        | Test delivery exists, but ongoing delivery trust is not visible.                                     |
|    8 | Role Permission Visibility                 | RBAC exists in backend; users need to know why actions are disabled/blocked.                         |
|    9 | Audit Trail / Event History                | Sensitive actions require accountability for self-host/enterprise contexts.                          |
|   10 | Public Status Trust & Redaction Layer      | Public pages reveal measurement output; need freshness, source, redaction, disclaimer.               |

## Next Research Plan

### Recruit

- 4 SRE/platform engineers who manage distributed network or edge infrastructure。
- 3 NOC/on-call operators who triage alerts。
- 3 self-host evaluators/open-source operators。
- 2 support/customer communications leads。
- 2 security/compliance reviewers。
- 2 external incident/status page readers。

### Interview Questions

- When do you need network evidence from a specific location rather than generic cloud uptime?
- What would convince you a probe result is trustworthy?
- What does "stale", "partial", "timeout", "no result" mean to you?
- What information can safely appear on a public status page?
- What would stop you from approving a self-hosted probe agent?
- What do you expect a product named Netstamp to do?

### Usability Test Tasks

- Create first project, first probe, first check, and confirm first result。
- Configure an alert rule and notification channel。
- Triage an incident and share a summary。
- Create a public status page without exposing internal details。
- Rotate a probe secret and explain the risk。
- Resolve a no data / stale probe scenario。

### Metrics To Instrument

- Time to first project。
- Time to first heartbeat。
- Time to first successful result。
- Drop-off in probe install step。
- Check selector preview usage and save errors。
- Alert rule creation success/failure。
- Notification test success/failure and retries。
- Public status page publish/disable events。
- Error recovery CTA click-through。
- Search/filter usage in operational pages。

### Open Product Questions

- Should Netstamp intentionally support document timestamping in future, or explicitly avoid that category?
- Should DNS/HTTP checks be roadmap items, or should all copy remove them until implemented?
- What is the minimum proof payload for RCA and external sharing?
- Should public status pages include subscribers/RSS/JSON API in V1 or later?
- What audit events are required for enterprise adoption?
- Should probes expose OS/arch/version/capacity/queue health to the controller UI?
- What role can create public status pages or expose targets?
- Should alert incidents support acknowledge/manual resolve/postmortem?
