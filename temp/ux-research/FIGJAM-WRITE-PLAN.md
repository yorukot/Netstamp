# FigJam Write Plan

Target board:

`https://www.figma.com/board/lJxnGVfbIM1WmiiHGUmVzx/Netstamp-UX-Research?node-id=0-1&t=BPLSmnfDSL18hTsV-1`

## Write Strategy

- Preserve all existing content.
- Add a new research system to the right of the current board, starting near `x = 28000`, `y = 0`.
- Use 4 columns x 6 rows for 24 sections, matching `06-figjam-structure.md`.
- Each section gets:
  - large header text,
  - short explanation text,
  - concise sticky notes,
  - evidence tags on key claims,
  - consistent color categories.
- Do not paste full markdown into sticky notes. Sticky notes should carry one claim each.
- Keep synthetic interview material explicitly labeled `[Synthetic interview]`.
- Put high-priority recommendations in dark/black cards with white text.

## Color System

| Meaning                         | FigJam color                 |
| ------------------------------- | ---------------------------- |
| Existing feature                | Blue                         |
| User need / JTBD                | Green                        |
| Research insight                | Yellow                       |
| Pain point                      | Orange                       |
| Critical risk                   | Red                          |
| Opportunity                     | Violet                       |
| Assumption / hypothesis         | Pink                         |
| Open question                   | Gray                         |
| Section header / method         | White                        |
| Highest priority recommendation | Dark / black with white text |

## Board Sections

| No. | Section title                  | Purpose                                        | Key content                                                                        |
| --: | ------------------------------ | ---------------------------------------------- | ---------------------------------------------------------------------------------- |
|  01 | Research Overview              | Explain goal, scope, limitations, confidence.  | Goal, scope, known limitations, evidence legend.                                   |
|  02 | Product Context                | Establish true current product.                | Self-hosted network observability; not document stamping.                          |
|  03 | Methodology                    | Show how research was built.                   | Code scan, FigJam readback, synthetic interview, competitive patterns.             |
|  04 | Evidence Map                   | Make claims traceable.                         | README, routes, API, server, UI, FigJam, assumptions.                              |
|  05 | Feature Inventory              | Current implemented feature clusters.          | Auth, projects, probes, checks, results, alerts, status, API.                      |
|  06 | User Types / Personas          | Proto-personas with confidence tags.           | SRE, NOC, IT Ops, self-host, support, security, external reviewer.                 |
|  07 | Jobs To Be Done                | What users need to accomplish.                 | Reachability, latency/loss, route changes, alerts, sharing, governance.            |
|  08 | Research Questions             | What needs real validation.                    | Trust, public data, activation, incident workflow, naming.                         |
|  09 | Assumptions & Risks            | Separate hypothesis from fact.                 | No doc stamping, DNS/HTTP gap, no billing, no export.                              |
|  10 | Interview Guide                | Future research execution.                     | Screening, script, task prompts, metrics.                                          |
|  11 | Synthetic Interview Notes      | Simulated roles and findings.                  | 8 simulated interviewee cards.                                                     |
|  12 | Affinity Map                   | Group recurring needs/pains.                   | Mental model, trust, incident, public status, self-host, security, labels, naming. |
|  13 | Current-State Journey          | Real/inferred current flow from code.          | Arrive, account, project, probe, check, result, alert, status, govern.             |
|  14 | Future-State Journey           | Ideal improved journey.                        | Orientation, activation, measurement, triage, communication, governance.           |
|  15 | Service Blueprint              | Frontstage/backstage dependencies.             | Session, project, probe, check, result, alert, status.                             |
|  16 | UX Audit                       | Heuristic issues and fixes.                    | Top issues with severity.                                                          |
|  17 | Pain Points                    | Concentrated pain list.                        | Activation, trust, selectors, public exposure, incident lifecycle.                 |
|  18 | Opportunity Areas              | Actionable product areas.                      | Onboarding, measurement trust, error recovery, sharing, governance.                |
|  19 | Feature Mind Map               | Structured map of current/recommended/missing. | Netstamp -> Users -> Jobs -> Features -> Pain -> Opportunities -> Features.        |
|  20 | Missing / Recommended Features | Detailed proposed features.                    | F01-F18, highlight top 10.                                                         |
|  21 | Prioritization Matrix          | Impact/Effort.                                 | Top-left quick wins, top-right strategic investments.                              |
|  22 | Roadmap Suggestions            | Now/Next/Later lanes.                          | 0-4, 4-10, 10-16+ weeks.                                                           |
|  23 | Open Questions                 | Product/research unknowns.                     | Document strategy, DNS/HTTP, proof payload, public data, audit.                    |
|  24 | Next Research Plan             | Team-ready research plan.                      | Recruit, interview questions, usability tasks, metrics.                            |

## Required Sticky Notes For FigJam

### Research Overview

- `Research goal: build code-backed UX understanding for Netstamp and convert it into actionable product strategy.`
- `Scope: repository, API, UI, server, FigJam readback, synthetic interviews, competitive patterns.`
- `Known limitation: no real user interviews in this round. [Synthetic interviews only]`
- `Confidence: High = code-backed; Medium = inferred/pattern; Low = assumption/hypothesis.`

### Product Context

- `Netstamp = self-hosted network observability from probes you control. [Evidence: README.md:4,22-24]`
- `Core value: reachability, latency, packet loss, routes, TCP reachability, probe health. [Evidence: README.md:28-33]`
- `Current check types: Ping, TCP, Traceroute. [Evidence: api/models/check.tsp:3]`
- `No current code evidence for document stamping/certificate verification. [No code evidence] [Confidence: High]`

### Feature Inventory

- `Auth/session: register, login, logout, me. [Evidence: api/main.tsp tags; AuthPage]`
- `Projects: workspace, switcher, roles, members, invites. [Evidence: project permission.go]`
- `Probes: create, install command, secret, heartbeat, fleet management. [Evidence: NewProbeDrawer.tsx]`
- `Checks: Ping/TCP/traceroute, selector preview, labels. [Evidence: api/models/check.tsp]`
- `Results: ping/tcp/traceroute series, insight, topology. [Evidence: api/main.tsp Results]`
- `Alerts: rules, incidents, notifications. [Evidence: AlertsPage.tsx]`
- `Public status pages: slug page, metrics, incidents, generatedAt. [Evidence: PublicStatusPage.tsx]`
- `OpenAPI/docs/install assets. [Evidence: router.go]`

### Highest Priority Recommendations

- `F01 Guided First-Run Checklist: do now because value is proven only after first result.`
- `F02 Does / Does Not Do: do now because product name can imply document stamping.`
- `F03 Result Trust Page: do next because Netstamp sells evidence from controlled probes.`
- `F06 Incident Review Workspace: do next because alerts need triage and RCA flow.`
- `F12 Role Permission Visibility: do now because backend RBAC exists but users need explanations.`

## Mind Map Connections

Main connector chain:

```text
Netstamp
-> User Types
-> Jobs
-> Current Features
-> Pain Points
-> Opportunities
-> Recommended Features
```

Risk connectors:

```text
Netstamp name
-> Document stamping expectation
-> No code evidence
-> Does / Does Not Do recommendation
```

Trust connectors:

```text
Results / Incidents / Public Status
-> Trust gaps
-> Proof explanation
-> Result Trust Page
-> Export / Share Report
```

Activation connectors:

```text
Onboarding
-> Probe heartbeat
-> Missing first result handoff
-> Guided First-Run Checklist
```

## Prioritization Matrix Positions

| Feature                    |   X effort |    Y impact |
| -------------------------- | ---------: | ----------: |
| Does / Does Not Do         |        Low |        High |
| Empty States               |        Low |        High |
| Role Permission Visibility | Low-Medium |        High |
| First-Run Checklist        |     Medium |        High |
| Result Trust Page          |     Medium |        High |
| Incident Review Workspace  |     Medium |        High |
| Public Status Trust Layer  |     Medium |        High |
| Audit Trail                |       High |        High |
| Admin Health               |       High | Medium-High |
| Search/History Hub         |       High |      Medium |

## Roadmap Lanes

### Now

- Guided First-Run Checklist。
- Netstamp Does / Does Not Do。
- Empty States With Operational Next Actions。
- Error Recovery Playbooks。
- Role Permission Visibility。

### Next

- Result Trust / Proof Page。
- Proof Explanation Panel。
- Incident Review Workspace。
- Notification Delivery Status Center。
- Public Status Trust and Redaction。
- Security/Privacy Messaging。

### Later

- Export / Share Report。
- Audit Trail / Event History。
- Search / Filter / History Hub。
- Sample Project / Demo Learning Mode。
- Admin Controller Health。
- Developer/API Shortcuts。
- Accessibility/Mobile Pass。

## Validation Checklist After Write

- New sections do not cover existing board content.
- Every section has a header.
- Sticky colors follow the legend.
- Critical claims include `[Evidence: ...]` or `[Assumption]`.
- Synthetic interview block is clearly labeled simulated.
- Feature mind map distinguishes existing vs missing vs recommended.
- Prioritization matrix has impact/effort axes.
- Roadmap has Now / Next / Later lanes.
- Open questions and next research plan are visible.
