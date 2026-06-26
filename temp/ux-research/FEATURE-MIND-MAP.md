# Netstamp Feature Mind Map

Legend:

- `[Existing]` code-backed current feature, blue in FigJam.
- `[Recommended]` research-backed proposed feature, purple or dark priority card in FigJam.
- `[Assumption]` hypothesis requiring product/user validation, pink in FigJam.
- `[No code evidence]` explicitly not found in repository, red/gray in FigJam depending context.

```text
Netstamp
├─ Product Definition
│  ├─ [Existing] Self-hosted network observability
│  ├─ [Existing] Probes from networks users control
│  ├─ [Existing] Ping / TCP / traceroute checks
│  ├─ [Existing] Results, alerts, public status, OpenAPI
│  ├─ [No code evidence] Document stamping
│  ├─ [No code evidence] Certificate issuing
│  ├─ [No code evidence] File verification / notary portal
│  ├─ [No code evidence] Billing / plan / usage quota
│  └─ [Recommended] Does / Does Not Do product boundary
│
├─ User Types
│  ├─ [Evidence-based inference] SRE / platform engineer
│  ├─ [Evidence-based inference] NOC / on-call operator
│  ├─ [Evidence-based inference] IT operations manager
│  ├─ [Evidence-based inference] Self-host operator
│  ├─ [Evidence-based inference] Support / customer comms lead
│  ├─ [Evidence-based inference] Security / compliance owner
│  ├─ [Evidence-based inference] External status / incident reviewer
│  └─ [Evidence-based inference] Developer / API user
│
├─ Jobs To Be Done
│  ├─ Know if a service is reachable from real locations
│  ├─ Compare latency / loss by region, ISP, provider, lab, edge
│  ├─ Detect route changes and unstable paths
│  ├─ Track probe health and reporting freshness
│  ├─ Get alerts in existing team channels
│  ├─ Share status without exposing internal infrastructure
│  ├─ Explain incident evidence for RCA / customer communication
│  ├─ Govern access, secrets, public data and audit history
│  └─ Automate via OpenAPI
│
├─ Current Features
│  ├─ Identity & Access
│  │  ├─ [Existing] Register / login / logout / session me
│  │  ├─ [Existing] Account settings: profile, email, password
│  │  ├─ [Existing] Pending project invites
│  │  ├─ [Existing] Demo/read-only mode
│  │  ├─ [Missing] Password reset
│  │  └─ [Recommended] Security session / credential explanation
│  │
│  ├─ Project Workspace
│  │  ├─ [Existing] Project create/update/delete/leave
│  │  ├─ [Existing] Project switcher
│  │  ├─ [Existing] Owner/Admin/Editor/Viewer roles
│  │  ├─ [Existing] Members and invites
│  │  ├─ [Recommended] Role permission visibility
│  │  └─ [Recommended] Audit trail / event history
│  │
│  ├─ Probe Fleet
│  │  ├─ [Existing] New probe wizard
│  │  ├─ [Existing] Location search / manual coordinates
│  │  ├─ [Existing] Registration token and install command
│  │  ├─ [Existing] Heartbeat detection
│  │  ├─ [Existing] Probe list / map / detail / edit / delete
│  │  ├─ [Existing] Secret rotation
│  │  ├─ [Recommended] Probe install troubleshooting
│  │  ├─ [Recommended] Probe health diagnostics
│  │  └─ [Recommended] Secret hygiene and rotation impact warning
│  │
│  ├─ Labels & Assignment
│  │  ├─ [Existing] Label CRUD
│  │  ├─ [Existing] Label usage
│  │  ├─ [Existing] Selector preview
│  │  ├─ [Existing] Advanced selector state
│  │  ├─ [Recommended] Assignment impact preview
│  │  └─ [Recommended] Label governance templates
│  │
│  ├─ Checks
│  │  ├─ [Existing] Ping config
│  │  ├─ [Existing] TCP config
│  │  ├─ [Existing] Traceroute config
│  │  ├─ [Existing] Create / edit / duplicate / delete / batch delete
│  │  ├─ [No code evidence] DNS check
│  │  ├─ [No code evidence] HTTP check
│  │  └─ [Recommended] Check templates and safer defaults
│  │
│  ├─ Results & Insights
│  │  ├─ [Existing] Ping series / insight
│  │  ├─ [Existing] TCP series / insight
│  │  ├─ [Existing] Traceroute runs / topology
│  │  ├─ [Existing] Time range and refresh controls
│  │  ├─ [Recommended] Result trust / proof page
│  │  ├─ [Recommended] Raw evidence panel
│  │  ├─ [Recommended] Failed vs successful comparison
│  │  └─ [Recommended] Export / share report
│  │
│  ├─ Alerts & Incidents
│  │  ├─ [Existing] Alert rules
│  │  ├─ [Existing] Incidents
│  │  ├─ [Existing] Notifications: webhook, Slack, Discord, Telegram, email
│  │  ├─ [Existing] Test notification
│  │  ├─ [Partial] Traceroute alerts disabled
│  │  ├─ [Recommended] Incident review workspace
│  │  ├─ [Recommended] Notification delivery status center
│  │  ├─ [Recommended] Alert templates
│  │  └─ [Recommended] Acknowledge / manual resolve / postmortem hypothesis
│  │
│  ├─ Public Status
│  │  ├─ [Existing] Public status pages by slug
│  │  ├─ [Existing] Page elements / folders / assignment groups
│  │  ├─ [Existing] Generated timestamp
│  │  ├─ [Existing] Metrics/charts/open incidents
│  │  ├─ [Recommended] Public data redaction / aliases
│  │  ├─ [Recommended] Methodology and non-SLA disclaimer
│  │  ├─ [Recommended] Public proof snapshot
│  │  └─ [Assumption] Subscribers / RSS / JSON API
│  │
│  └─ Developer / Self-host
│     ├─ [Existing] OpenAPI contract
│     ├─ [Existing] `/docs` Scalar UI
│     ├─ [Existing] Docker Compose
│     ├─ [Existing] Install assets
│     ├─ [Existing] Metrics/OTLP/logging config
│     ├─ [Recommended] Copy ID / copy curl / View API
│     ├─ [Recommended] Controller health / readiness page
│     └─ [Recommended] Security exposure checklist
│
├─ Pain Points
│  ├─ Activation stops before first result
│  ├─ Product name can imply document timestamping
│  ├─ Probe/check/assignment vocabulary is dense
│  ├─ Public status may expose technical details
│  ├─ Incident flow lacks operational lifecycle
│  ├─ Error states often lack recovery CTA
│  ├─ Role limitations not always clear before action
│  ├─ Secret/geocoding/public data privacy explanation is thin
│  ├─ DNS/HTTP copy risks overclaiming
│  └─ Mobile operational tables are dense
│
├─ Opportunities
│  ├─ Onboarding clarity
│  ├─ Measurement trust
│  ├─ Incident triage and RCA
│  ├─ Public communication
│  ├─ Governance/security
│  ├─ Error recovery
│  ├─ API/developer workflow
│  ├─ Demo/sample learning
│  └─ Accessibility/mobile operations
│
└─ Recommended Features
   ├─ P0 / Now
   │  ├─ Guided First-Run Checklist
   │  ├─ Netstamp Does / Does Not Do
   │  ├─ Empty States With Operational Next Actions
   │  ├─ Error Recovery Playbooks
   │  └─ Role Permission Visibility
   ├─ P1 / Next
   │  ├─ Result Trust / Proof Page
   │  ├─ Proof Explanation Panel
   │  ├─ Incident Review Workspace
   │  ├─ Notification Delivery Status Center
   │  ├─ Public Status Trust & Redaction Layer
   │  ├─ Export / Share Report
   │  └─ Audit Trail / Event History
   └─ P2 / Later
      ├─ Sample Project / Demo Learning Mode
      ├─ Admin Visibility / Controller Health
      ├─ Developer/API Docs Shortcuts
      ├─ Accessibility and Mobile Operations Pass
      ├─ Legal / Compliance Disclaimer
      ├─ Search / Filter / History Hub
      └─ Subscribers / RSS / JSON API hypothesis
```

## FigJam Mind Map Guidance

Center node: `Netstamp = self-hosted network observability`.

Primary branches:

1. User Types: green。
2. Jobs: green。
3. Current Features: blue。
4. Pain Points: orange/red。
5. Opportunities: purple。
6. Recommended Features: purple/dark priority cards。
7. No-Evidence / Assumptions: pink/gray。

Required high-risk red nodes:

- `No document/certificate verification in current code`
- `DNS/HTTP current support not evidenced`
- `Activation stops before first result`
- `Public status can expose technical target/probe details`
- `Result trust needs source/freshness/sample/error explanation`
