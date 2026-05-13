# Netstamp

Netstamp is a distributed network observability platform for measuring reachability, latency, packet loss, and probe health from multiple network viewpoints. It combines a Go controller API, a React service app, a static Astro documentation site, a shared React UI package, and a probe runtime that reports measurements back to the controller.

The product is designed for teams that need to understand how network services behave from real locations, provider edges, labs, private infrastructure, or regional nodes. A project owns probes, labels, checks, team membership, and measurement results. Probes authenticate to the controller, receive assignments, execute checks, and submit time-series results for analysis.

## Highlights

- Project-based workspaces with owner, admin, editor, and viewer roles.
- User registration, login, JWT sessions, and authenticated project APIs.
- Probe inventory with labels, location metadata, status reporting, and secret rotation.
- Label-driven check assignment through selector expressions.
- Ping checks with packet count, packet size, timeout, and IP family configuration.
- Probe runtime API for hello, heartbeat, assignment polling, and result submission.
- PostgreSQL plus TimescaleDB storage for relational state and time-series ping results.
- OpenAPI generation from Huma route definitions.
- React product app for dashboard, probes, checks, insights, alerts, team, and settings.
- Astro documentation site with API explorer and Storybook output for shared UI components.
- Structured logging, Prometheus-compatible metrics, and optional OpenTelemetry trace export.

## Repository Layout

```text
.
├── server/               Go backend controller, migrations, sqlc queries, and probe runtime commands
├── web/                  React 19 + Vite authenticated service app
├── docs/                 Astro public site, documentation, OpenAPI explorer, and Storybook publishing
├── packages/ui/          Shared React components and design tokens exported as @netstamp/ui
├── packages/brand/       Shared Netstamp brand assets
├── deployments/          Docker Compose, Nginx, Grafana, VictoriaMetrics, VictoriaTraces, and Vector config
├── Justfile              Canonical local task runner
└── package.json          pnpm workspace scripts
```

## Architecture

Netstamp has four main runtime surfaces:

1. The controller API in `server/` owns authentication, authorization, projects, labels, checks, probes, assignments, probe runtime traffic, persistence, logging, metrics, and traces.
2. The probe runtime runs near the network being measured. It authenticates with `Authorization: Probe <secret>`, polls assignments, executes checks, sends heartbeats, and submits result batches.
3. The web app in `web/` is the authenticated operator interface.
4. The docs app in `docs/` is the public documentation surface and API reference.

The normal request path is:

```text
Browser -> React app -> /api/v1/* -> chi + Huma -> application service -> PostgreSQL/TimescaleDB
```

The probe path is:

```text
Probe -> /api/v1/probes/{probe_id}/runtime/* -> runtime service -> assignments + ping results
```

The backend uses a layered Go structure:

- `transport/http`: route registration, Huma DTOs, middleware, and HTTP error mapping.
- `application/*`: use cases, authorization, orchestration, event semantics, and feature validation.
- `domain/*`: stable domain models, permissions, selector parsing, and validation-normalization helpers.
- `infrastructure/*`: PostgreSQL repositories, sqlc integration, JWT, Argon2id password hashing, and probe secrets.
- `platform/observability`: metrics, tracing, and HTTP trace helpers.

## Backend Capabilities

All HTTP API routes except `/metrics` are mounted under `/api/{API_VERSION}`, which defaults to `/api/v1`.

### System

- `GET /`
- `GET /healthz`
- `GET /metrics`

The health endpoint checks database connectivity. Metrics are exposed separately at `/metrics` for Prometheus-compatible scraping.

### Authentication

- `POST /auth/register`
- `POST /auth/login`
- `GET /auth/me`

User passwords are hashed with Argon2id. Access tokens are JWT HS256 tokens signed with `AUTH_JWT_SECRET`.

### Projects And Members

- `GET /projects`
- `POST /projects`
- `GET /projects/{ref}`
- `PATCH /projects/{ref}`
- `DELETE /projects/{ref}`
- `GET /projects/{ref}/members`
- `POST /projects/{ref}/members`
- `PATCH /projects/{ref}/members/{user_id}`
- `DELETE /projects/{ref}/members/{user_id}`

`{ref}` is a project slug. Project membership is role-based:

- `owner`: full project control, project deletion, and member administration.
- `admin`: project write access and member administration, except owner-level role changes.
- `editor`: create and update labels, checks, and probes.
- `viewer`: read-only project access.

The application layer owns authorization decisions. HTTP middleware proves identity, then project services enforce role policy.

### Labels

- `GET /projects/{ref}/labels`
- `POST /projects/{ref}/labels`
- `PATCH /projects/{ref}/labels/{label_id}`
- `DELETE /projects/{ref}/labels/{label_id}`

Labels are key/value pairs scoped to a project. They are attached to probes and checks, then used by selectors to decide which probes should receive which checks.

### Checks

- `GET /projects/{ref}/checks`
- `POST /projects/{ref}/checks`
- `GET /projects/{ref}/checks/{check_id}`
- `PATCH /projects/{ref}/checks/{check_id}`
- `DELETE /projects/{ref}/checks/{check_id}`

The completed check model supports ping checks. A ping check includes:

- target domain or IP address
- interval in seconds
- optional description
- selector JSON
- attached labels
- packet count
- packet size
- timeout
- optional IPv4 or IPv6 preference

Selectors are JSON expressions that match probe labels. Empty selectors match every probe. Supported selector nodes include `all`, `any`, `not`, and `label` with `eq`, `in`, and `exists` operations.

Example selector:

```json
{
	"all": [{ "label": { "key": "region", "op": "eq", "value": "tw-north" } }, { "label": { "key": "network", "op": "in", "values": ["fiber", "ix"] } }]
}
```

### Probes

- `GET /projects/{ref}/probes`
- `POST /projects/{ref}/probes`
- `GET /projects/{ref}/probes/{probe_id}`
- `PATCH /projects/{ref}/probes/{probe_id}`
- `DELETE /projects/{ref}/probes/{probe_id}`
- `POST /projects/{ref}/probes/{probe_id}/secret-rotations`

Probes represent measurement viewpoints. A probe can include a name, enabled state, subdivision code, latitude, longitude, labels, and runtime status. Probe creation and secret rotation return plaintext credentials once; only secret hashes are stored.

### Probe Runtime

- `POST /probes/{probe_id}/runtime/hello`
- `POST /probes/{probe_id}/runtime/heartbeat`
- `GET /probes/{probe_id}/runtime/assignments`
- `POST /probes/{probe_id}/runtime/results`

Probe runtime endpoints use the `Authorization: Probe <secret>` header instead of user JWT auth.

The runtime flow is:

1. `hello`: authenticate, report probe status, and receive controller timing hints.
2. `heartbeat`: keep the probe online and refresh status metadata.
3. `assignments`: retrieve active check assignments produced from check selectors and probe labels.
4. `results`: submit ping result batches with idempotency on project, probe, check, and start time.

Result submission validates assignment identity, check type, check version, selector version, timing, loss percentage, RTT ordering, resolved IP, IP family, and raw JSON payloads. If a probe submits stale assignment versions, the controller accepts valid results and asks the probe to resync.

## Data Model

Netstamp stores relational state and time-series results in PostgreSQL with TimescaleDB enabled.

Core tables include:

- `users`
- `projects`
- `project_members`
- `labels`
- `checks`
- `ping_check_configs`
- `probes`
- `probe_credentials`
- `probe_statuses`
- `probe_labels`
- `check_labels`
- `probe_check_assignments`
- `ping_results`

`ping_results` is a TimescaleDB hypertable partitioned by `started_at`. It stores duration, sent and received packet counts, loss percentage, RTT metrics, RTT samples, resolved IP, IP family, raw result data, and optional error details.

The Docker Compose stacks default to the lightweight `timescale/timescaledb:latest-pg16` image through `TIMESCALEDB_IMAGE`. Migrations enable core TimescaleDB only; result series downsampling uses `time_bucket` instead of TimescaleDB Toolkit `lttb`.

Database migrations live in `server/db/migrations/`. SQL queries for sqlc live in `server/db/query/`, and generated Go code lives under `server/internal/controller/infrastructure/postgres/sqlc/`.

## Tech Stack

### Backend

- Go
- chi
- Huma
- pgx
- sqlc
- Goose
- PostgreSQL
- TimescaleDB
- Viper
- Zap
- OpenTelemetry
- Prometheus metrics

### Frontend And Docs

- pnpm workspace
- React 19
- React Router
- Vite
- TypeScript
- Astro
- MDX
- Storybook
- ECharts
- MapLibre GL
- `@netstamp/ui`

### Deployment And Observability

- Docker and Docker Compose
- Nginx
- VictoriaMetrics
- VictoriaTraces
- VictoriaLogs through Vector
- Grafana

## Requirements

- Node.js 22.12 or newer
- pnpm 11
- Go version matching `server/go.mod`
- Just
- Docker and Docker Compose
- PostgreSQL/TimescaleDB for backend development
- Air for backend hot reload
- golangci-lint for backend formatting and linting

## Getting Started

Install workspace dependencies:

```bash
pnpm install
```

Create local backend configuration:

```bash
cp server/.env.example server/.env
cp server/probe.env.example server/probe.env
```

Start local backend dependencies:

```bash
docker compose -f deployments/docker/compose.backend.dev.yaml up -d postgres victoria-traces victoria-metrics grafana
```

Apply migrations:

```bash
just backend-migrate-up
```

Start the backend controller:

```bash
just backend-dev
```

Start the web app:

```bash
just web-dev
```

Start the docs site:

```bash
just docs-dev
```

Run a probe with probe credentials:

```bash
just backend-probe server/probe.env
```

## Configuration

The backend reads environment variables and an optional `.env` file from the repository root or `server/`.

Common controller settings:

| Variable                             | Purpose                                         | Default                             |
| ------------------------------------ | ----------------------------------------------- | ----------------------------------- |
| `APP_ENV`                            | Runtime environment name                        | `local`                             |
| `SERVICE_NAME`                       | Service name used in logs and telemetry         | `controller`                        |
| `APP_VERSION`                        | Service version used in logs and telemetry      | `0.1.0`                             |
| `API_VERSION`                        | API path version mounted at `/api/{version}`    | `v1`                                |
| `LOG_LEVEL`                          | Zap log level                                   | `info`                              |
| `LOG_PSEUDONYM_KEY`                  | Key for privacy-preserving log pseudonyms       | local development value             |
| `BACKEND_BASE_URL`                   | Absolute backend origin for OpenAPI server URLs | empty                               |
| `HTTP_ADDR`                          | Controller listen address                       | `:8080`                             |
| `REQUEST_TIMEOUT`                    | Request timeout middleware duration             | `10s`                               |
| `SHUTDOWN_TIMEOUT`                   | Graceful shutdown timeout                       | `10s`                               |
| `DATABASE_HOST`                      | PostgreSQL host                                 | `localhost`                         |
| `DATABASE_PORT`                      | PostgreSQL port                                 | `5432`                              |
| `DATABASE_USER`                      | PostgreSQL username                             | `netstamp`                          |
| `DATABASE_PASSWORD`                  | PostgreSQL password                             | `netstamp`                          |
| `DATABASE_NAME`                      | PostgreSQL database                             | `netstamp`                          |
| `DATABASE_SSLMODE`                   | PostgreSQL SSL mode                             | `disable`                           |
| `TIMESCALEDB_IMAGE`                  | Compose image for PostgreSQL/TimescaleDB        | `timescale/timescaledb:latest-pg16` |
| `AUTH_JWT_SECRET`                    | JWT signing secret                              | local development value             |
| `AUTH_ACCESS_TOKEN_TTL`              | JWT access token lifetime                       | `12h`                               |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | Optional OTLP traces endpoint                   | empty                               |

Probe settings:

| Variable                        | Purpose                             |
| ------------------------------- | ----------------------------------- |
| `NETSTAMP_PROBE_CONTROLLER_URL` | Controller origin used by the probe |
| `NETSTAMP_PROBE_ID`             | Probe UUID                          |
| `NETSTAMP_PROBE_SECRET`         | Plaintext probe secret              |
| `NETSTAMP_PROBE_HTTP_TIMEOUT`   | Probe HTTP client timeout           |
| `NETSTAMP_PROBE_MAX_WORKERS`    | Maximum concurrent check workers    |

Never commit production `.env` files, JWT secrets, database passwords, probe secrets, or telemetry endpoints with credentials.

## Common Commands

```bash
just                  # List available recipes
pnpm install          # Install workspace dependencies
just backend-dev      # Start the Go controller with Air
just backend-probe    # Run the probe runtime
just web-dev          # Start the Vite service app
just docs-dev         # Start the Astro docs app
just storybook-dev    # Start Storybook for @netstamp/ui
just build            # Build backend, web, and docs
just lint             # Run backend and web linting
just test             # Run backend tests
just backend-sqlc     # Regenerate sqlc Go code
just backend-openapi  # Regenerate docs/public/openapi.json
pnpm format           # Format JS, TS, CSS, JSON, Markdown, and Astro files
```

## Testing

Run backend unit tests:

```bash
just backend-test
```

Run integration tests against a PostgreSQL/TimescaleDB test database:

```bash
docker compose -f deployments/docker/compose.backend.dev.yaml up -d postgres
NETSTAMP_TEST_DATABASE_URL=postgres://netstamp:netstamp@localhost:5432/netstamp?sslmode=disable just backend-test-integration
```

Run web linting:

```bash
just web-lint
```

Run the full repository validation path:

```bash
just lint
just test
just build
```

## OpenAPI

The API contract is generated from backend Huma route definitions. Regenerate it after changing HTTP routes, request bodies, response bodies, security schemes, or operation metadata:

```bash
pnpm generate:openapi
```

The generated file is written to:

```text
docs/public/openapi.json
```

The docs app serves the API explorer at `/openapi/`.

## Local Observability

The backend can export metrics and traces locally through the development compose stack:

```bash
docker compose -f deployments/docker/compose.backend.dev.yaml up -d victoria-traces victoria-metrics grafana
```

Default local endpoints:

- Controller metrics: `http://localhost:8080/metrics`
- VictoriaMetrics: `http://127.0.0.1:8428`
- VictoriaTraces: `http://localhost:10428`
- Grafana: `http://127.0.0.1:3000`

The backend writes structured Zap logs and records application events for auth, projects, labels, checks, probes, and probe runtime workflows.

## Deployment

The production Docker Compose stack builds:

- the controller image from `server/Dockerfile`
- a migration job using the same backend image
- TimescaleDB for persistence
- an Nginx image for the web and docs surfaces

Start the production-style stack:

```bash
docker compose -f deployments/docker/compose.yaml up -d --build
```

Required production environment values include:

- `DATABASE_PASSWORD`
- `AUTH_JWT_SECRET`
- `LOG_PSEUDONYM_KEY`
- `GF_SECURITY_ADMIN_PASSWORD` when running the observability stack

Run migrations before serving the controller. The compose stack includes a `migrate` service that applies Goose migrations before the controller starts.

`TIMESCALEDB_IMAGE` may be overridden for local or deployment testing, but the selected image must include every extension enabled by migrations. The default lightweight image is enough for the current schema because it only requires core TimescaleDB features such as hypertables and `time_bucket`.

## Development Guidelines

- Prefer the root `Justfile` commands for repeatable local workflows.
- Keep backend changes inside the existing transport, application, domain, and infrastructure layers.
- Add database changes as Goose migrations under `server/db/migrations/`.
- Add SQL queries under `server/db/query/`, then run `just backend-sqlc`.
- Regenerate OpenAPI after route changes.
- Keep generated sqlc files, OpenAPI files, and Storybook output produced by their owning commands.
- Update the closest `AGENTS.md` when commands, architecture, configuration, or structure changes.

## Security Notes

- User routes use JWT bearer authentication.
- Probe runtime routes use probe-specific secrets, not user JWTs.
- Passwords are stored as Argon2id hashes.
- Probe secrets are shown only when created or rotated.
- Soft-deleted projects, labels, checks, and probes are excluded from normal access paths.
- Project authorization is enforced in application services using domain role policy.
- Technical error details are logged and traced, but public API responses keep error details conservative.

## License

Netstamp is licensed under the terms in [LICENSE](./LICENSE).
