# Netstamp Product & Technical Document (English)

> A distributed network observability platform — measure reachability, latency, packet loss, and route paths from probes you control, placed at real network locations.
>
> Chinese version: [PRODUCT.zh-TW.md](./PRODUCT.zh-TW.md)

---

## Table of Contents

1. [Product Overview](#1-product-overview)
2. [What Problem It Solves](#2-what-problem-it-solves)
3. [Target Users & Use Cases](#3-target-users--use-cases)
4. [Core Concepts & Domain Model](#4-core-concepts--domain-model)
5. [System Architecture](#5-system-architecture)
6. [How It Works (Implementation Principles)](#6-how-it-works-implementation-principles)
7. [Data Model & Time-Series Storage](#7-data-model--time-series-storage)
8. [API Design](#8-api-design)
9. [Tech Stack](#9-tech-stack)
10. [Deployment Architecture](#10-deployment-architecture)
11. [Security Design](#11-security-design)
12. [Observability](#12-observability)
13. [Scalability & Performance](#13-scalability--performance)
14. [Key Design Decisions & Trade-offs](#14-key-design-decisions--trade-offs)
15. [Glossary](#15-glossary)

---

## 1. Product Overview

**Netstamp** is an open-source, self-hostable **distributed network observability platform**. It lets teams deploy measurement probes to any network location — cloud regions, ISP edges, labs, private infrastructure, or geographic nodes — and continuously measure how target services behave from each of those vantage points.

One-line positioning:

> **"See the network before it fails you" — measure latency, packet loss, and route paths from probes you own.**

The platform consists of five parts:

| Component | Role | Technology |
| --- | --- | --- |
| **Controller** | Backend API: auth, authorization, projects, probes, checks, assignments, result storage | Go + chi + PostgreSQL/TimescaleDB |
| **Probe Agent** | Runs near the network under test; executes measurements and reports results | Go (installable as a Linux systemd service) |
| **Web App** | Authenticated operator console | React 19 + Vite |
| **Docs Site** | Public docs, API Explorer, landing page | Astro + MDX |
| **`@netstamp/ui`** | Cross-surface shared React components and design tokens | React + Storybook |

Core mental model: **A project owns probes, labels, checks, and members. A probe authenticates to the controller, polls for "assignments," executes checks, and submits time-series measurement results for analysis.**

---

## 2. What Problem It Solves

Traditional "single-vantage" monitoring (e.g., pinging a target from one machine or a single cloud region) cannot answer the questions distributed systems actually care about:

- **"How does my service look from different regions and ISPs?"** — A single vantage point hides geographic and path differences.
- **"Users report slowness — is it my service, or some hop in between?"** — Without traceroute topology, localization is hard.
- **"Where exactly is the packet loss and latency jitter happening?"** — Requires multi-vantage, long-horizon time-series data.
- **"When did the route path change, and was it correlated with an incident?"** — Requires historical path-hash comparison.
- **"Can I fully own my measurement points without depending on a third-party SaaS black box?"** — Requires a self-hostable, open-source, data-owned solution.

Netstamp's answers:

1. **Multi-vantage measurement** — Deploy probes anywhere and measure from real network edges, not a single central point.
2. **Three measurement types** — ICMP Ping (reachability / latency / loss), TCP Connect (port reachability + handshake latency), Traceroute (per-hop route topology).
3. **Label-driven assignment** — Selector expressions automatically decide "which probes run which checks," without manual one-by-one binding.
4. **Time-series analysis** — High-frequency measurements stored in TimescaleDB, with continuous aggregates delivering sub-second dashboard queries.
5. **Fully owned, open source** — Probes, controller, and database all self-hosted; data never leaves your infrastructure.
6. **Public status pages** — Optionally expose the health of selected checks to the outside world.

---

## 3. Target Users & Use Cases

**Primary users**

- **Platform / SRE teams** who need to watch their service's external reachability from multiple geographic and network locations.
- **Network engineers** who need traceroute topology and route-change detection to localize cross-network issues.
- **Infrastructure owners** who want a self-hosted, data-owned observability solution with no external SaaS dependency.

**Typical scenarios**

- Place one probe each in northern Taiwan, Japan, and the US West Coast, continuously pinging your API endpoint and comparing latency and loss across the three.
- Set TCP checks against critical third-party dependencies (payment gateways, DNS, CDN edges) to track handshake-latency changes.
- Enable traceroute on core targets, get alerted the moment the route path hash changes, and see on the topology map which hop degraded.
- Publish a public status page to show customers the health of key services.

---

## 4. Core Concepts & Domain Model

Understanding Netstamp comes down to the relationships among these domain objects:

```text
User ──< ProjectMember >── Project
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
      Label                 Probe                  Check
        │  (key/value)        │ (vantage point)      │ (what to measure)
        │                     │                     │
        ├── probe_labels ─────┤                     ├── ping/tcp/traceroute config
        └── check_labels ─────┼─────────────────────┤
                              │                     │
                              │   selector match    │
                              ▼                     ▼
                       ProbeCheckAssignment
                              │  (check_version, selector_version)
                              ▼
                   Probe executes → PingResult / TCPResult / TracerouteResult (time-series)
```

| Concept | Description |
| --- | --- |
| **User** | An account. Passwords stored as Argon2id hashes. |
| **Project** | A workspace and the boundary for all resources. Referenced externally by `slug`. Supports soft delete. |
| **ProjectMember** | A user's role in a project: `owner` / `admin` / `editor` / `viewer`. |
| **ProjectInvite** | Invitation to join a project, with status `pending` / `accepted` / `rejected`. |
| **Label** | A project-scoped key/value pair. Attached to probes and checks; the basis for selector matching. |
| **Probe** | A measurement vantage point. Has a name, enabled state, geo coordinates, location name, labels, and runtime status. Uses a UUID externally and a compact `internal_id` for time-series data. |
| **Probe Credential** | A probe's secret hash. Plaintext is returned only once, at creation or rotation. |
| **Probe Status** | Online/offline, last heartbeat, agent version, public IPs, AS number, etc. |
| **Check** | Defines "what to measure." Type is `ping` / `tcp` / `traceroute`, with a target, interval seconds, selector, and type-specific config. |
| **Selector** | A JSON expression describing "which probes match." Supports `all` / `any` / `not` / `label` (`eq` / `in` / `exists`). An empty selector matches every probe. |
| **Assignment** | The fact "this probe runs this check," computed from selector × probe labels, carrying `check_version` and `selector_version` hashes. |
| **Result** | Time-series measurement data returned by a probe: ping/tcp/traceroute each have a structure, stored in TimescaleDB hypertables. |
| **Public Page** | An externally visible status page; checks to display can be organized in a folder tree. |

---

## 5. System Architecture

### 5.1 Four Runtime Surfaces

```text
                        ┌──────────────────────────────┐
   Operator browser ───▶│  Web App (React 19 + Vite)    │
                        │  authenticated operator console│
                        └──────────────┬───────────────┘
                                       │  cookie session (JWT)
                                       │  /api/v1/*
                                       ▼
   Public visitor ──▶ Docs Site ──▶ ┌──────────────────────────────┐      ┌──────────────────────┐
   (Astro + MDX)     API Explorer   │  Controller API (Go + chi)   │◀────▶│ PostgreSQL/TimescaleDB │
                                    │  transport→application→domain │      │ relational + time-series│
                                    │  →infrastructure→observability│      └──────────────────────┘
                                    └──────────────┬───────────────┘
                                       ▲            │
              Authorization: Probe ... │            │ assignments, result validation
                                       │            ▼
                        ┌──────────────────────────────┐
   Near network ────────│  Probe Agent (Go)             │
   under test           │  hello/heartbeat/assignments/ │
                        │  results; runs ping/tcp/trace  │
                        └──────────────────────────────┘
```

Two main request paths:

```text
Operator: Browser → React App → /api/v1/* → chi → application service → PostgreSQL/TimescaleDB
Probe:    Probe  → /api/v1/runtime/probes/{probe_id}/* → runtime service → assignments + results
```

> Note: the actual probe-runtime endpoint paths are `/api/v1/runtime/probes/{probe_id}/{hello|heartbeat|assignments|results}`.

### 5.2 Layered Backend Architecture (Controller)

The controller uses a strict layered (hexagonal-flavored) architecture with a **one-way downward** dependency direction:

```text
┌─────────────────────────────────────────────────────────┐
│ transport/http                                          │
│   route registration, middleware stack, request/response │
│   DTOs, HTTP error mapping                               │
├─────────────────────────────────────────────────────────┤
│ application/*                                           │
│   use cases, authorization decisions, orchestration,     │
│   event semantics, input validation                     │
│   per domain: service.go / flow.go / ports.go /          │
│              validate.go / dto.go / errors.go / trace.go │
├─────────────────────────────────────────────────────────┤
│ domain/*                                                │
│   stable domain models, permission policy, selector      │
│   parsing, version hashes, validate-normalize (VN*)      │
├─────────────────────────────────────────────────────────┤
│ infrastructure/*                                        │
│   PostgreSQL repositories (sqlc-generated), JWT,         │
│   Argon2id, probe secrets                                │
├─────────────────────────────────────────────────────────┤
│ platform/observability                                  │
│   metrics (Prometheus), tracing (OTLP), HTTP trace help  │
└─────────────────────────────────────────────────────────┘
```

**Dependency direction**: `transport → application → domain ← infrastructure`, with `platform` providing cross-cutting observability. Authorization decisions live in the application layer; HTTP middleware only proves identity, and role policy is enforced inside application services via domain policy functions.

### 5.3 The Standard Application-Layer File Pattern

Every domain (auth, project, label, check, probe, proberuntime, assignment, result, publicpage, user) follows the same file split — the project's most distinctive convention:

| File | Responsibility |
| --- | --- |
| `service.go` | Core business logic; public method entry points |
| `flow.go` | Execution-flow management: create OpenTelemetry spans, record application events, distinguish "business failure" from "technical failure" |
| `ports.go` | Interface definitions (repository, event recorder) and event constants |
| `validate.go` | Input normalization and validation (the normalize + validate pattern) |
| `dto.go` | Data transfer objects |
| `errors.go` | Domain error definitions |
| `trace.go` | Tracer and span attributes |

The **business vs technical failure** distinction is key: `flow.businessFailure()` is for known domain errors (e.g., "user already exists," "forbidden") and does not mark the span as an error; `flow.technicalFailure()` is for systemic errors (e.g., a database outage) and sets the span to error status. This lets tracing and alerting correctly distinguish "expected rejections" from "real failures."

---

## 6. How It Works (Implementation Principles)

### 6.1 Startup & Lifecycle

The controller starts in distinct phases (`server/cmd/controller/main.go` → `app/bootstrap.go` → `app/lifecycle.go`):

1. **Signal context**: build a context that receives SIGINT/SIGTERM via `signal.NotifyContext`.
2. **Dependency assembly** (bootstrap): load config → init Zap logger → init metrics/tracing → create the pgx pool → instantiate security (Argon2id hasher, JWT issuer), repositories, and application services → assemble the HTTP router.
3. **Start the HTTP server**: use `errgroup` to run "HTTP serve" and "watch for context cancellation" concurrently.
4. **Graceful shutdown**: shut down in order — HTTP server → DB pool → metrics provider → tracing provider (flush pending spans).

### 6.2 Authentication & Authorization

**User authentication**

- Passwords are hashed with **Argon2id** (64 MiB memory, 3 iterations, parallelism 4, 16-byte salt, 32-byte hash); comparison uses `subtle.ConstantTimeCompare` to resist timing attacks.
- On login a **JWT (HS256)** is issued, signed with `AUTH_JWT_SECRET`, with claims `sub` (user id), `email`, `iss`/`aud` (both `netstamp`), and `iat`/`nbf`/`exp`.
- The token lives in an **HTTP-only `netstamp_session` cookie**, forced `Secure` outside local. The frontend sends it automatically via `credentials: "include"`.
- The `RequireAuth` middleware reads and verifies the JWT from the cookie and injects the claims into the request context.

**Project authorization (RBAC)**

- The domain layer exposes a decision function (`Can(role, action)`):
  - `viewer`: read-only;
  - `editor`: create/update labels, checks, probes;
  - `admin`: project writes + member administration (except owner-level role changes);
  - `owner`: full control, including project deletion.
- Before each operation, the application service looks up the user's role in the project, then calls the domain decision function; on rejection it returns `ErrForbidden`, which transport maps to 403.

### 6.3 Selector Engine & Assignment Computation

This is Netstamp's core automation mechanism. Instead of binding each check to each probe manually, the user writes a selector expression describing "which probes I want":

```json
{
  "all": [
    { "label": { "key": "region", "op": "eq", "value": "tw-north" } },
    { "label": { "key": "network", "op": "in", "values": ["fiber", "ix"] } }
  ]
}
```

Supported nodes: `all` (all must hold), `any` (any holds), `not` (negation), `label` (apply `eq` / `in` / `exists` to a label). An empty selector matches all probes.

**How assignments are produced and maintained:**

- When a **probe**, **check**, or **label** is updated, the application's assignment service recomputes assignments for the affected scope: for each relevant (probe, check) pair it evaluates whether the selector matches; if so it upserts a `probe_check_assignments` row, otherwise it soft-deletes.
- Each assignment stores two version hashes:
  - `check_version` = SHA256 of the check's content (target, interval, type config); **name and description do not affect the version** (they don't affect execution).
  - `selector_version` = SHA256 of the selector JSON.
- The controller also offers a **selector preview** endpoint so users can see, while authoring a check, exactly which probes would be matched and how many.

### 6.4 Probe Agent Runtime Lifecycle

The probe is a standalone Go program, runnable via `netstamp-agent run` or installable as a systemd service. Its flow:

```text
Start → Hello (authenticate + get server time / minimum version / config)
        │ (exponential backoff retry; auth failure terminates immediately)
        ▼
  errgroup runs five concurrent loops:
  ├─ heartbeatLoop   report status every HeartbeatInterval (agent version, local IPs)
  ├─ assignmentLoop  poll assignments every PollInterval → Reconcile old vs new
  ├─ Scheduler.Run   schedule each check by due time using a min-heap
  ├─ Workers.Run     worker pool executes checks concurrently
  └─ Submitter.Run   batch-upload results
        │
        ▼
  context cancelled → wait ShutdownTimeout, best-effort flush remaining results → exit
```

**Key designs:**

- **Version sync (resync)**: the probe tags each task locally with a `generation` number. When `check_version` or `selector_version` changes, generation is incremented and the task is rescheduled; tasks absent from the new list are marked removed. If the controller receives results for a stale assignment version, it still accepts valid results and asks the probe to resync.
- **Phase jitter**: an offset of 0–59 seconds is derived from `FNV-1a(probeID + assignmentID)` so probes don't all hit the same target on the minute boundary, spreading measurement load.
- **Assignment TTL**: if the last successful assignment poll is older than `AssignmentTTL`, scheduling stops — avoiding execution of stale tasks while disconnected from the controller.
- **Result queue full policy**: when the result queue (default capacity 10,000) is full, the **oldest** entry is dropped, prioritizing the newest measurements.
- **Retry**: hello/heartbeat/results all use exponential backoff (initial → ×2 → max backoff, up to MaxAttempts); auth failure or a permanent 4xx terminates immediately.

### 6.5 Executing the Three Measurement Types

| Type | How it runs | Output metrics | Privilege |
| --- | --- | --- | --- |
| **Ping (ICMP)** | Open an ICMP raw socket, send N echo requests per config, record sequence numbers and timestamps, receive replies concurrently | duration, sent/received count, loss%, RTT min/avg/median/max/stddev, RTT sample array, resolved IP, IP family, status | requires `CAP_NET_RAW` |
| **TCP** | Use `net.Dialer` to TCP-connect to `host:port`, record handshake duration, then close immediately | duration, connect duration, resolved IP, IP family, status | no privilege |
| **Traceroute** | ICMP or UDP, incrementing TTL per hop, N queries per hop | destination reached, hop count, per-hop address/hostname/loss/RTT stats and samples, overall status (successful/partial/timeout) | ICMP mode requires `CAP_NET_RAW` |

### 6.6 Result Submission, Validation & Idempotency

Probes send results in **batches** to `POST /runtime/probes/{probe_id}/results`. The controller's validation chain:

1. **Authenticate the probe** (`Authorization: Probe <secret>`).
2. **Normalize and structurally validate**: results must be non-empty; no duplicate group per (checkId, type); no duplicate `startedAt` within a check.
3. **Assignment reconciliation**: every checkId must have a matching active assignment, with matching type.
4. **Timing and value validation**: time ordering, loss% range, RTT ordering, resolved IP, IP family, raw JSON payload.
5. **Type-specific writes** to the ping/tcp/traceroute result tables.

The **idempotency key** is the combination (project, probe, check, startedAt) — result tables use `(probe_id, check_id, started_at)` as the primary key. Even if a probe resubmits identical results, no duplicates are created; if a stale assignment version is submitted, the controller accepts valid results and responds asking the probe to resync.

### 6.7 Frontend Data Flow

The Web App uses a **feature-based** architecture (`web/src/features/*`), with a shared layer under `web/src/shared/*`:

- **Type-safe API**: the API contract is authored in TypeSpec → emitted to OpenAPI → TS types generated via `openapi-typescript` → a strongly typed client built with `openapi-fetch` (`credentials: "include"` carries the cookie).
- **Data fetching**: centered on **TanStack Query (React Query)** — queries deduplicate and cache automatically (measurement staleTime ~30s, the Insight page auto-refetches every 15s), and mutations `invalidateQueries` and push a toast on success.
- **URL state**: the Insight page encodes time range and filters into the query string, supporting sharing and browser back/forward.
- **Visualization**: **ECharts** for time-series charts (with drag-to-select time range), **MapLibre GL** for a geographic probe-distribution map, and a custom SVG engine for traceroute route topology.
- **Auth flow**: `SessionProvider` calls `GET /auth/me` to sync login state; `ProtectedAppShell` guards protected routes and redirects unauthenticated users to `/login`; first-time registration redirects to `/onboarding` to create an initial project.
- **Project switching**: `useCurrentProject` remembers the selected project in localStorage (like a workspace switcher).
- UI always prefers `@netstamp/ui` shared components (Button, DataTable, Panel, MetricCard, Badge, …), maintaining the consistent "network operations console" visual language across surfaces.

---

## 7. Data Model & Time-Series Storage

Netstamp stores both **relational state** and **time-series results** in PostgreSQL (with TimescaleDB enabled).

### 7.1 Table Categories (~38 tables / materialized views)

| Layer | Representative tables | TimescaleDB | Retention |
| --- | --- | --- | --- |
| **Business** | `users`, `projects`, `project_members`, `project_invites`, `labels`, `probes`, `probe_credentials`, `probe_statuses`, `probe_labels`, `checks`, `ping/tcp/traceroute_check_configs`, `check_labels`, `probe_check_assignments` | No | Forever (soft delete) |
| **Results** | `ping_results`, `tcp_results`, `traceroute_results`, `traceroute_result_hops` | Yes (hypertable) | raw 3 days |
| **Observations** | `ping_rtt_sample_observations`, `traceroute_hop_observations`, `traceroute_edge_observations` | Yes (hypertable) | 3 days |
| **Aggregates** | Continuous aggregates per result type (1m/10m/15m/30m/1h) | Yes (continuous aggregate) | 30–180 days |
| **Public pages** | `public_pages`, `public_page_folders`, `public_page_folder_checks` | No | Forever (soft delete) |

### 7.2 Key Design: UUID Externally, internal_id Internally

`probes` and `checks` carry both `id` (UUID, for the external API) and `internal_id` (bigint identity, used as the foreign key in time-series result tables). Why: time-series tables routinely hold hundreds of millions of rows, so a compact bigint as the key dramatically cuts storage and speeds up queries.

### 7.3 Hypertables & Retention

- `*_results` and the observation tables are all hypertables, partitioned by `started_at`.
- **chunk interval = 1 day** (reduced from the 7-day default for finer-grained compression and queries).
- **raw results retained 3 days**: keep high-precision raw data recently and auto-drop old chunks; long-term trends are served by continuous aggregates.

### 7.4 Continuous Aggregates

For each of ping/tcp/traceroute, **five time scales — 1m, 10m, 15m, 30m, 1h** — are built as continuous aggregates, pre-downsampled with `time_bucket`. Each aggregate includes counts (success/timeout/error), duration stats, loss, and RTT stats (min/avg/max/stddev as sum + count).

- **Refresh policy**: `start_offset` ~3 days (refresh only the recent window), `end_offset` slightly behind real time (tolerate a few minutes of lag), `schedule_interval` matching the bucket width.
- **Retention**: 1m → 30 days, 10m/15m → 90 days, 30m → 180 days, 1h longest.

Purpose: **when dashboards and the Insight page query long time ranges, they read pre-aggregated materialized results directly instead of scanning huge raw row sets**, achieving sub-second response.

### 7.5 Why the Observation Tables Exist

- `ping_rtt_sample_observations`: "explodes" the `ping_results.rtt_samples_ms[]` array into one row per sample, making it easy to compute p50/p95/p99 with `percentile_cont` and stddev, and to produce an RTT-distribution histogram (latency heatmap).
- `traceroute_hop_observations` / `traceroute_edge_observations`: a trigger automatically derives "node" and "edge-between-adjacent-hops" observations from traceroute hops, precomputing the node health and edge quality needed for the topology graph — avoiding expensive JOINs at query time.

### 7.6 sqlc Workflow

SQL queries live in `server/db/query/*.sql` and are compiled by **sqlc** into type-safe Go code under `server/internal/controller/infrastructure/postgres/sqlc/`. Migrations are managed by **Goose** in `server/db/migrations/`. Complex time-series queries (e.g., `GetPingInsightSummary` using CTEs + `percentile_cont`; traceroute using `string_agg` + the `lag()` window function to detect path changes) are expressed as raw SQL in the query files.

---

## 8. API Design

### 8.1 Contract-First

The API contract is authored in **TypeSpec** (the `api/` directory, split by models and services), emitted to OpenAPI at `docs/public/openapi.json` via `pnpm generate:openapi`, and served by the controller in the docs-site API Explorer (`/openapi/`). The frontend's TS types are generated from the same OpenAPI, keeping front and back ends contractually aligned.

### 8.2 Two Authentication Schemes

- **User routes**: HTTP-only `netstamp_session` cookie (containing a signed JWT).
- **Probe runtime routes**: `Authorization: Probe <secret>` header, fully separate from user JWTs.

### 8.3 Main Endpoints (all mounted under `/api/{API_VERSION}`, default `/api/v1`)

**System**
- `GET /`, `GET /healthz` (checks DB connectivity), `GET /metrics` (Prometheus, mounted at `/metrics` without the version prefix)

**Authentication**
- `POST /auth/register`, `POST /auth/login`, `GET /auth/me`

**Projects & members**
- `GET|POST /projects`, `GET|PATCH|DELETE /projects/{ref}`
- `GET|POST /projects/{ref}/members`, `PATCH|DELETE /projects/{ref}/members/{user_id}`
- Project invite endpoints (pending/accept/reject)

**Labels**
- `GET|POST /projects/{ref}/labels`, `PATCH|DELETE /projects/{ref}/labels/{label_id}`

**Checks**
- `GET|POST /projects/{ref}/checks`, `GET|PATCH|DELETE /projects/{ref}/checks/{check_id}`
- Selector preview endpoint

**Probes**
- `GET|POST /projects/{ref}/probes`, `GET|PATCH|DELETE /projects/{ref}/probes/{probe_id}`
- `POST /projects/{ref}/probes/{probe_id}/secret-rotations` (secret rotation; plaintext returned only once)

**Results & measurements**
- Project-level measurement and per-type insight query endpoints

**Probe runtime** (probe auth)
- `POST /runtime/probes/{probe_id}/hello`
- `POST /runtime/probes/{probe_id}/heartbeat`
- `GET  /runtime/probes/{probe_id}/assignments`
- `POST /runtime/probes/{probe_id}/results`

**Public pages** (some require no user auth)
- Project-level management endpoints + an endpoint to fetch a public page by slug

**Install assets**
- `GET /api/v1/install/agent.sh`, `/uninstall-agent.sh`, `/netstamp-agent-linux-amd64`

### 8.4 Error Format

Uses **RFC 7807 Problem Details** (`application/problem+json`) with `type`/`title`/`status`/`detail`/`instance` plus an `errors[]` array for validation errors (each with `message`/`location`/`value`), and returns `X-Request-ID` in the header. Public responses stay conservative; technical details go only to logs and traces.

---

## 9. Tech Stack

**Backend (Controller / Agent)**
- Go, chi (HTTP routing), pgx (PostgreSQL driver), sqlc (type-safe queries), Goose (migrations)
- PostgreSQL + TimescaleDB, Viper (config), Zap (structured logging), OpenTelemetry (traces), Prometheus (metrics)
- Cobra (agent CLI)

**Frontend & Docs**
- pnpm workspace, React 19, React Router, Vite, TypeScript
- TanStack Query, openapi-fetch / openapi-typescript
- ECharts (time-series charts), MapLibre GL (maps)
- Astro + MDX (docs site), Storybook (`@netstamp/ui`)

**Deployment & Observability**
- Docker / Docker Compose, Nginx
- VictoriaMetrics (metrics), VictoriaTraces (traces), VictoriaLogs + Vector (logs), Grafana (dashboards)

---

## 10. Deployment Architecture

### 10.1 Development

```bash
pnpm install
cp server/.env.example server/.env
cp server/probe.env.example server/probe.env
docker compose -f deployments/docker/compose.backend.dev.yaml up -d postgres victoria-traces victoria-metrics grafana
just backend-migrate-up   # apply Goose migrations
just backend-dev          # start the controller (Air hot reload)
just web-dev              # start the Web App
just docs-dev             # start the docs site
just backend-probe server/probe.env   # run a probe
```

### 10.2 Production

`deployments/docker/compose.yaml` builds:

- The controller image (`server/Dockerfile`)
- A migration job (same image, applies Goose migrations before the controller starts)
- The Linux amd64 probe agent binary (served by the controller's install endpoints)
- TimescaleDB
- An Nginx image (serves the web and docs static assets)

```bash
docker compose -f deployments/docker/compose.yaml up -d --build
```

Required production environment values: `DATABASE_PASSWORD`, `AUTH_JWT_SECRET`, `LOG_PSEUDONYM_KEY` (and `GF_SECURITY_ADMIN_PASSWORD` when running the observability stack).

### 10.3 Installing a Probe on a Linux Host

```bash
curl -fsSL https://example.com/api/v1/install/agent.sh | sudo sh
sudo netstamp-agent service install \
  --url https://example.com \
  --probe-id <probe-id> \
  --probe-secret <probe-secret>
```

`service install` creates a `netstamp` system user, writes `/etc/netstamp/probe.env`, and enables `netstamp-agent.service`. The systemd unit runs as non-root, granting only `CAP_NET_RAW` (needed for ICMP) and enabling sandbox options like `NoNewPrivileges`, `PrivateTmp`, `ProtectHome`, and `ProtectSystem`. Add `--purge` on uninstall to also remove config and the system user.

---

## 11. Security Design

- **Passwords**: Argon2id hashes, constant-time comparison.
- **User sessions**: HTTP-only `netstamp_session` cookie with an HS256 JWT; `Secure` forced outside local.
- **Probe authentication**: per-probe secrets, stored only as hashes; plaintext returned only at creation/rotation. Fully separate from user JWTs.
- **Authorization**: enforced in the application layer via domain role policy (owner/admin/editor/viewer).
- **Soft-delete isolation**: soft-deleted projects, labels, checks, and probes are excluded from normal access paths.
- **Error information**: conservative externally (no internal detail leakage); technical detail goes only to logs and traces.
- **Privacy**: logs use `LOG_PSEUDONYM_KEY` for privacy-preserving pseudonyms.
- **Secret management**: never commit production `.env`, JWT secrets, DB passwords, probe secrets, or telemetry endpoints with credentials.
- **Frontend tracking consent**: supports `regional` / `always` / `never` consent gating, deciding whether to request consent based on visitor country (resolved from multiple sources), meeting EEA/UK/Switzerland requirements.

---

## 12. Observability

The controller has three built-in observability pillars:

- **Metrics**: Prometheus-compatible, exposed at `/metrics`.
- **Traces**: OpenTelemetry — automatic (HTTP middleware spans) plus manual (application flow spans), exportable via `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`.
- **Logs**: Zap structured logs recording application events for auth, projects, labels, checks, probes, and runtime workflows, leveled by status code (5xx → Error, 4xx → Warn, 2xx → Info).

Local observability stack (dev compose): VictoriaMetrics (`:8428`), VictoriaTraces (`:10428`), Grafana (`:3000`, default `admin`/`admin`). Grafana provisions the `Netstamp Controller Status` dashboard and sets it as the home dashboard.

---

## 13. Scalability & Performance

- **Horizontal multi-vantage scaling**: probes are stateless measurement units that can be added at any location; the controller handles coordination and storage.
- **Phase jitter** spreads measurement load, avoiding on-the-minute spikes.
- **Tiered time-series storage**: high-precision raw with short retention (3 days) + continuous aggregates with long retention (30–180 days), balancing storage cost against query precision.
- **Query performance**: dashboards/Insight read directly from pre-aggregated materialized results; hypertable chunk exclusion accelerates time-range queries; observation tables pre-flatten samples and topology to avoid expensive on-the-fly computation.
- **Batching & queuing**: probe-side batch upload, drop-oldest-keep-newest under queue pressure, ensuring the freshest data is prioritized after bursts or reconnection.
- **Idempotent writes**: keyed by (probe, check, startedAt), so resubmission never duplicates.

---

## 14. Key Design Decisions & Trade-offs

| Decision | Rationale | Trade-off |
| --- | --- | --- |
| **Strict layered backend + authorization in application** | Centralized, testable domain logic and authorization; swappable storage | More boilerplate (7 files per domain) |
| **Selector-expression-driven assignment** | Label-driven automation; new probes auto-join matching checks | Needs version hashes and reconcile to stay consistent |
| **UUID externally + internal_id internally** | Compact bigint in time-series tables saves storage and speeds queries | Dual keys add a little complexity |
| **Short raw retention + long aggregate retention** | Sub-second dashboards; controlled storage cost | Raw per-sample data unavailable beyond 3 days |
| **Probe queue drop-oldest-keep-newest** | Prioritize the latest state after reconnection | Some historical points lost under extreme congestion |
| **TypeSpec contract-first** | Single source of truth for types, docs, and the Explorer | Route changes require regenerating OpenAPI |
| **Dual track: user JWT vs probe secret** | Separate security models for two principal types (human/machine) | Two auth systems to maintain |
| **Lightweight default TimescaleDB image** | Only core hypertables and `time_bucket` needed; no Toolkit | Downsampling uses `time_bucket` rather than `lttb` |

---

## 15. Glossary

- **Probe**: an agent deployed at a network location that runs measurements and reports back.
- **Check**: the config defining "what to measure" (ping/tcp/traceroute + target + interval + selector).
- **Label**: a project-scoped key/value attached to probes and checks.
- **Selector**: a JSON expression describing "which probes match" (`all`/`any`/`not`/`label`).
- **Assignment**: the computed "this probe runs this check," carrying version hashes.
- **Controller**: the backend API coordinating auth, assignment, and result storage.
- **Hypertable**: a TimescaleDB time-partitioned table used for high-frequency time-series results.
- **Continuous Aggregate**: a pre-downsampled materialized view for fast long-range queries.
- **Path hash**: a fingerprint of a traceroute route path, used to detect route changes.
- **Heartbeat**: a probe's periodic "I'm still online" signal.

---

*This document is written against the actual codebase, covering the controller, probe agent, frontend, database, and deployment layers. If architecture, configuration, or commands change, please update this document and the nearest `AGENTS.md` together.*
