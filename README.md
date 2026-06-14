<div align="center">
  <img src="./packages/brand/assets/netstamp-logo.svg" alt="Netstamp" width="360" />

  <h3>Open-source network observability from probes you control.</h3>

  <p>
    Measure latency, packet loss, routes, and probe health from real network viewpoints.
  </p>

  <p>
    <a href="./LICENSE"><img alt="License" src="https://img.shields.io/github/license/yorukot/netstamp?style=flat-square" /></a>
    <img alt="Go" src="https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go&logoColor=white" />
    <img alt="React" src="https://img.shields.io/badge/React-19-61DAFB?style=flat-square&logo=react&logoColor=111111" />
    <img alt="pnpm" src="https://img.shields.io/badge/pnpm-workspace-F69220?style=flat-square&logo=pnpm&logoColor=white" />
  </p>
</div>

---

## Overview

Netstamp is built for teams that need to know whether a service is reachable from the places that matter: home labs, edge nodes, regional servers, private infrastructure, or any machine that can run a probe.

The product is split into:

| Surface           | Purpose                                                                                                          |
| ----------------- | ---------------------------------------------------------------------------------------------------------------- |
| `server/`         | Go controller API, database access, migrations, probe runtime endpoints, alert evaluation, notification delivery |
| `web/`            | React service app for dashboards, probes, checks, insights, alerts, projects, and settings                       |
| `docs/`           | Astro public site, documentation, OpenAPI explorer, and Storybook publishing                                     |
| `packages/ui/`    | Shared React UI components and design tokens                                                                     |
| `packages/brand/` | Netstamp logo and brand assets                                                                                   |

## Content

- [Overview](#overview)
- [Features](#features)
- [How It Works](#how-it-works)
- [Installation](#installation)
- [Start Netstamp](#start-netstamp)
- [Probe Agent](#probe-agent)
- [Commands](#commands)
- [Configuration](#configuration)
- [OpenAPI](#openapi)
- [Docker](#docker)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## Features

- Multi-project workspaces with owner, admin, editor, and viewer roles.
- User registration, login, JWT sessions, and project-scoped authorization.
- Probe inventory with labels, location metadata, online status, and secret rotation.
- Label-based check assignment through selector expressions.
- Ping checks with packet count, packet size, timeout, and IP family settings.
- Probe runtime API for hello, heartbeat, assignment polling, and result submission.
- PostgreSQL plus TimescaleDB storage for relational state and time-series measurements.
- Insight views for latency, packet loss, route behavior, and probe health.
- Alert rules, incidents, notification channels, and test notifications.
- Webhook, Discord, and Telegram notification channels.
- Generated OpenAPI contract from TypeSpec.
- Structured logs, Prometheus-compatible metrics, and optional OpenTelemetry traces.

## How It Works

Netstamp has one controller and many probes.

```text
Browser
  -> React web app
  -> /api/v1/*
  -> Go controller
  -> PostgreSQL / TimescaleDB
```

```text
Probe agent
  -> authenticate with probe secret
  -> poll assignments
  -> run checks
  -> submit results
  -> controller evaluates alerts and sends notifications
```

The controller owns authentication, authorization, projects, labels, checks, probes, assignments, results, incidents, notification channels, metrics, and traces. Probe agents run near the network being measured and only need their controller URL, probe ID, and probe secret.

## Installation

### Requirements

- [Node.js](https://nodejs.org/) 22.12 or newer
- [pnpm](https://pnpm.io/) 11
- [Go](https://go.dev/doc/install) version matching `server/go.mod`
- [Just](https://github.com/casey/just)
- [Docker](https://docs.docker.com/get-docker/) and Docker Compose
- PostgreSQL with TimescaleDB for backend development
- [Air](https://github.com/air-verse/air) for backend hot reload
- [golangci-lint](https://golangci-lint.run/) for backend linting

Clone the repository and install workspace dependencies.

```bash
git clone https://github.com/yorukot/netstamp.git
cd netstamp
pnpm install
```

Create local backend and probe configuration files.

```bash
cp server/.env.example server/.env
cp server/probe.env.example server/probe.env
```

Prepare a PostgreSQL/TimescaleDB database, then update the `DATABASE_*` values in `server/.env`.

Apply migrations.

```bash
just backend-migrate-up
```

## Start Netstamp

Run the controller API.

```bash
just backend-dev
```

Run the service app.

```bash
just web-dev
```

Run the documentation site.

```bash
just docs-dev
```

By default, the backend listens on `http://localhost:8080`. Vite and Astro print their local URLs when they start.

## Probe Agent

Create or rotate a probe in the web app to get a probe ID and secret, then place those values in `server/probe.env`.

```bash
NETSTAMP_PROBE_CONTROLLER_URL=http://localhost:8080
NETSTAMP_PROBE_ID=<probe-id>
NETSTAMP_PROBE_SECRET=<probe-secret>
```

Run a local probe.

```bash
just backend-probe server/probe.env
```

Install a probe agent on a Linux systemd host.

```bash
curl -fsSL https://example.com/api/v1/install/agent.sh | sudo sh
sudo netstamp-agent service install \
  --url https://example.com \
  --probe-id <probe-id> \
  --probe-secret <probe-secret>
```

Uninstall a probe agent.

```bash
curl -fsSL https://example.com/api/v1/install/uninstall-agent.sh | sudo sh
```

## Commands

Use the root `Justfile` for repeatable local workflows.

| Command              | Purpose                                             |
| -------------------- | --------------------------------------------------- |
| `just`               | List available recipes                              |
| `pnpm install`       | Install workspace dependencies                      |
| `just backend-dev`   | Start the Go controller with Air                    |
| `just backend-probe` | Run the probe runtime                               |
| `just web-dev`       | Start the React service app                         |
| `just docs-dev`      | Start the Astro docs app                            |
| `just storybook-dev` | Start Storybook for `@netstamp/ui`                  |
| `just build`         | Build backend, web, docs, and API contract targets  |
| `just lint`          | Run backend and web linting                         |
| `just test`          | Run backend tests                                   |
| `just backend-sqlc`  | Regenerate sqlc Go code                             |
| `just api-openapi`   | Regenerate OpenAPI JSON and web API types           |
| `pnpm format`        | Format JS, TS, CSS, JSON, Markdown, and Astro files |

## Configuration

The backend reads environment variables and optional `.env` files from the repository root or `server/`.

Common controller settings:

| Variable                             | Purpose                                      |
| ------------------------------------ | -------------------------------------------- |
| `APP_ENV`                            | Runtime environment name                     |
| `API_VERSION`                        | API path version mounted at `/api/{version}` |
| `HTTP_ADDR`                          | Controller listen address                    |
| `DATABASE_HOST`                      | PostgreSQL host                              |
| `DATABASE_PORT`                      | PostgreSQL port                              |
| `DATABASE_USER`                      | PostgreSQL username                          |
| `DATABASE_PASSWORD`                  | PostgreSQL password                          |
| `DATABASE_NAME`                      | PostgreSQL database                          |
| `DATABASE_SSLMODE`                   | PostgreSQL SSL mode                          |
| `AUTH_JWT_SECRET`                    | JWT signing secret                           |
| `ALERT_EVALUATION_ENABLED`           | Enables alert evaluation                     |
| `NOTIFICATION_WORKER_ENABLED`        | Enables notification delivery                |
| `NOTIFICATION_HTTP_TIMEOUT`          | Timeout for outbound notification requests   |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | Optional OTLP trace endpoint                 |

Probe settings:

| Variable                           | Purpose                             |
| ---------------------------------- | ----------------------------------- |
| `NETSTAMP_PROBE_CONTROLLER_URL`    | Controller origin used by the probe |
| `NETSTAMP_PROBE_ID`                | Probe UUID                          |
| `NETSTAMP_PROBE_SECRET`            | Plaintext probe secret              |
| `NETSTAMP_PROBE_HTTP_TIMEOUT`      | Probe HTTP client timeout           |
| `NETSTAMP_PROBE_MAX_WORKERS`       | Maximum concurrent check workers    |
| `NETSTAMP_PROBE_RESULT_BATCH_SIZE` | Maximum results sent per flush      |
| `NETSTAMP_PROBE_ASSIGNMENT_TTL`    | Local assignment cache lifetime     |

Frontend and docs tracking variables use build-time public prefixes:

- `VITE_NETSTAMP_*` for `web/`
- `PUBLIC_NETSTAMP_*` for `docs/`

Never commit production `.env` files, JWT secrets, database passwords, probe secrets, webhook URLs, bot tokens, or telemetry endpoints with credentials.

## OpenAPI

The API contract is authored in TypeSpec and emitted to OpenAPI. Regenerate it after changing HTTP routes, request bodies, response bodies, security schemes, or operation metadata.

```bash
just api-openapi
```

Generated API assets:

| File                                                             | Purpose                                               |
| ---------------------------------------------------------------- | ----------------------------------------------------- |
| `docs/public/openapi.json`                                       | Public OpenAPI document used by the docs API explorer |
| `server/internal/controller/transport/http/openapi/openapi.json` | Embedded backend OpenAPI artifact                     |
| `web/src/shared/api/openapi.d.ts`                                | Generated web API types                               |

The docs app serves the API explorer at `/openapi/`.

## Docker

Start the production-style stack.

```bash
docker compose -f deployments/docker/compose.yaml up -d --build
```

Start the observability stack.

```bash
LOG_PSEUDONYM_KEY=local \
DATABASE_PASSWORD=netstamp \
POSTGRES_EXPORTER_PASSWORD=netstamp \
AUTH_JWT_SECRET=local-dev-secret \
GF_SECURITY_ADMIN_PASSWORD=admin \
docker compose -f deployments/docker/compose.observability.yaml up --build
```

Useful local endpoints:

| Surface            | URL                             |
| ------------------ | ------------------------------- |
| Controller metrics | `http://localhost:8080/metrics` |
| VictoriaMetrics    | `http://127.0.0.1:8428`         |
| VictoriaTraces     | `http://localhost:10428`        |
| Grafana            | `http://127.0.0.1:3000`         |

The production compose stack builds the controller image, runs migrations, serves Linux probe install assets, runs TimescaleDB, and serves web/docs surfaces through Nginx.

## Development

Run backend tests.

```bash
just backend-test
```

Run backend integration tests against an explicit PostgreSQL URL.

```bash
NETSTAMP_TEST_DATABASE_URL=postgres://netstamp:netstamp@localhost:5432/netstamp?sslmode=disable just backend-test-integration
```

Run the full validation path.

```bash
just lint
just test
just build
```

Development rules:

- Prefer root `Justfile` commands for repeatable workflows.
- Keep backend changes inside the existing transport, application, domain, and infrastructure layers.
- Add database changes as Goose migrations under `server/db/migrations/`.
- Add SQL queries under `server/db/query/`, then run `just backend-sqlc`.
- Regenerate OpenAPI after route or contract changes.
- Update the closest `AGENTS.md` when commands, architecture, configuration, or structure changes.

## Contributing

Issues, ideas, and pull requests are welcome. For larger changes, open an issue first so the implementation direction is clear.

Commit subjects should use the repository style:

```text
area: concise patch summary
```

Examples:

```text
web/alerts: improve channel setup flow
server/probe: validate assignment result windows
docs: clarify local setup
```

## License

Netstamp is licensed under the [Apache License 2.0](./LICENSE).
