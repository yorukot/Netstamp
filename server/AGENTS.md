# Repository Guidelines

## Project Structure & File Organization

This guide applies to `server/`, the Go backend for the Netstamp workspace. The root workspace also contains `web/`, `docs/`, and `packages/ui/`; use this file only for backend API, database, logging, tracing, and server runtime work.

- `cmd/api/main.go`: API process entry point. It creates the shutdown context, calls `app.New`, runs the app, and syncs the logger.
- `cmd/migrate/main.go`: Goose migration CLI for `status`, `up`, and `down`.
- `internal/app/`: composition root and lifecycle. `bootstrap.go` wires config, logging, tracing, PostgreSQL, auth, and HTTP. `lifecycle.go` starts and gracefully stops the HTTP listener.
- `internal/transport/http/`: chi/Huma HTTP routing, auth, project, probe, system health routes, and middleware.
- `internal/application/auth/`, `internal/application/project/`, and `internal/application/probe/`: use cases, ports, DTOs, errors, and feature orchestration.
- `internal/domain/identity/`, `internal/domain/project/`, and `internal/domain/probe/`: stable domain structs and domain-level sentinel errors.
- `internal/infrastructure/`: PostgreSQL repositories and pool helpers, JWT issuing, Argon2id password hashing, and probe secret generation.
- `internal/logger/` and `internal/observability/`: zap logging helpers, auth event recording, OpenTelemetry setup, and HTTP span naming.
- `db/migrations/`: Goose SQL migrations. `db/query/`: sqlc query files. Generated sqlc Go files live in `internal/infrastructure/postgres/sqlc/`.
- `tmp/` and `bin/`: local build artifacts; do not edit them as source.

No backend public asset directory is currently defined.

## System Architecture Overview

The backend is a single Go service with one listener: HTTP on `HTTP_ADDR`. `internal/app.New` loads validated configuration, creates a zap logger, initializes OpenTelemetry, opens a pgx pool, builds application services, and creates the HTTP server. `internal/app.Run` starts the server and coordinates graceful shutdown with `errgroup`.

HTTP uses chi middleware plus Huma route registration under `/api/{version}`. `internal/app/bootstrap.go` passes `cfg.APIVersion` (`API_VERSION`) into the router, and `BACKEND_BASE_URL` can publish an absolute OpenAPI server URL for deployed environments. System routes are `/`, `/livez`, and `/readyz`. Auth routes are `/auth/register`, `/auth/login`, and `/auth/me`; `/auth/me`, project routes, and project probe creation are protected by the Huma auth middleware in `internal/transport/http/middleware`.

The current auth request flow is:

`HTTP request -> chi/Huma route -> internal/transport/http/auth handler -> internal/application/auth.Service -> internal/infrastructure/postgres/user repository -> sqlc.Queries -> PostgreSQL`

No GraphQL, message queues, background workers, scheduled jobs, email, payment, or object-storage integrations are currently defined.

## Layer Responsibilities

- Transport (`internal/transport/http`): route registration, request/response DTOs, Huma validation tags, protocol status mapping, and middleware. Do not put database calls or business rules here.
- Application (`internal/application/auth`): business orchestration, service methods, ports, app errors, auth event semantics, and use-case spans. Depend on interfaces, not concrete pgx, Huma, or JWT types.
- Domain (`internal/domain/identity`): stable domain structs and domain-level sentinel errors such as `identity.ErrUserNotFound`.
- Infrastructure (`internal/infrastructure/postgres`, `internal/infrastructure/security`): pgx/sqlc persistence, database error translation, JWT HS256 tokens, and Argon2id password hashing.
- Config (`internal/config`): Viper-based environment loading, defaults, and validation. Add new env keys here and mirror them in `.env.example` when operators need to set them.
- Cross-cutting (`internal/logger`, `internal/observability`): request-scoped loggers, auth event recording, trace fields, tracer provider setup, and span helpers.

## Libraries & Dependencies

Direct backend dependencies are declared in `server/go.mod`.

- HTTP: `github.com/go-chi/chi/v5`, `github.com/danielgtaylor/huma/v2`, and `otelhttp`.
- Database: `github.com/jackc/pgx/v5`, `github.com/pressly/goose/v3`, `github.com/sqlc-dev/sqlc`, and `github.com/google/uuid`.
- Config: `github.com/spf13/viper`.
- Auth/security: `github.com/golang-jwt/jwt/v4` and `golang.org/x/crypto/argon2`.
- Logging: `go.uber.org/zap`.
- Tracing: OpenTelemetry SDK, trace API, and OTLP HTTP trace exporter.
- Tool tracking: `tools/tools.go` pins goose and sqlc. `air` and `golangci-lint` are used by commands/config but are not pinned in `server/go.mod`.

## Logging Guidelines

Zap is configured in `internal/logger/zap.go`. Every root logger includes `service`, `env`, and `version`; local env uses zap development config, other envs use production config. Valid `LOG_LEVEL` values are enforced in `internal/config/config_validate.go`: `debug`, `info`, `warn`, `error`, `dpanic`, `panic`, and `fatal`.

Use request-scoped loggers from `logger.FromContext(ctx, fallback)` when handling requests. HTTP logging in `internal/transport/http/middleware/logging.go` adds `request_id`, method, path, client address, user agent, status, bytes, duration, and trace fields.

Auth security events must go through `logger.AuthEventRecorder`. It pseudonymizes email into `user.email_hash` using `LOG_PSEUDONYM_KEY`. Do not log raw passwords, password hashes, access tokens, JWT secrets, cookies, database passwords, or raw personal data. Expected auth failures log at `warn`; technical failures log at `error`.

## Tracing & Observability

OpenTelemetry is configured in `internal/observability/tracing/tracing.go`. The provider always samples locally and exports only when `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` is set. TraceContext and Baggage propagators are installed globally.

HTTP tracing is wired through `otelhttp.NewMiddleware` in `internal/transport/http/router.go`, with span names from `internal/observability/httptrace/span.go`. Auth service methods create child spans in `internal/application/auth/trace.go` and `flow.go`. PostgreSQL repository methods use span helpers in `internal/infrastructure/postgres/trace.go`.

Keep `context.Context` as the first parameter for request, service, repository, and token operations so trace context and request loggers propagate across layers. New database calls should either use existing DB span helpers or add equivalent attributes without recording raw SQL parameters. Metrics are not currently configured.

## Build, Test, and Development Commands

Commands below come from the root `Justfile`, root `package.json`, `server/.air.toml`, `server/Dockerfile`, and compose files under `deployments/docker/`.

- `pnpm install`: install workspace dependencies; root `package.json` enforces pnpm.
- `just backend-dev` or `pnpm dev:server`: run Air hot reload using `server/.air.toml`.
- `just backend-build` or `pnpm build:server`: build `server/bin/api` from `./cmd/api`.
- `just backend-openapi` or `pnpm generate:openapi`: write the Huma OpenAPI schema to `docs/public/openapi.json`.
- `just backend-test` or `pnpm test:server`: run `go test ./...` inside `server/`.
- `just backend-fmt`: run `go fmt ./...`.
- `just golangci-lint`: run `golangci-lint` with `../golangci.yaml`.
- `just golangci-fmt`: run configured golangci formatters.
- `just backend-sqlc`: regenerate sqlc code from `sqlc.yaml`.
- `just backend-migrate-status`, `just backend-migrate-up`, `just backend-migrate-down`: run `cmd/migrate`.
- `docker compose -f deployments/docker/compose.backend.dev.yaml up -d`: start local PostgreSQL/TimescaleDB, VictoriaTraces, and Grafana dependencies for a host-run backend.
- `docker compose -f deployments/docker/compose.observability.yaml up --build`: build and run the Docker stack with PostgreSQL, VictoriaTraces, VictoriaLogs, Vector, Grafana, nginx, backend, and migrations.

Use `server/.env.example` as the env template. `server/.gitignore` intentionally ignores `.env`, `.env.*`, `bin/`, `tmp/`, and `coverage.out`.

## Coding Style & Naming Conventions

Go files use tabs and `gofmt` per root `.editorconfig`. `golangci.yaml` enables gofumpt, goimports, and gci formatting with local imports grouped under `github.com/yorukot/netstamp`. Keep package names short and lowercase, matching existing packages such as `auth`, `postgres`, and `httpserver`.

Follow existing feature file names: `service.go`, `ports.go`, `errors.go`, `trace.go`, `handler.go`, `register.go`, `login.go`, and `*_test.go`. Export only types needed across packages. Use sentinel errors named `Err...` and compare with `errors.Is`.

For HTTP feature packages, keep `handler.go` focused on the `Handler` type, constructor, and Huma route registration. Put each operation handler and its request/response DTOs in a separate operation file, matching `internal/transport/http/auth` patterns such as `register.go`, `login.go`, and `me.go`. Put response DTOs and mappers in `response.go` only when they are shared by multiple operation files. Do not concentrate all endpoint handler logic and DTOs in one large `handler.go`.

## Testing Guidelines

Tests use Go's standard `testing` package and live beside the code as `*_test.go`, for example `internal/application/auth/service_test.go` and `internal/logger/auth_events_test.go`. Existing tests use package-local fakes and zap observer cores rather than external test frameworks.

Run backend tests with `just backend-test` or `cd server && go test ./...`. Integration tests, end-to-end tests, fixtures, a test database setup, and coverage thresholds are not currently defined. Add unit tests beside changed packages, and document any new integration-test setup before relying on it in CI.

## Error Handling & Validation

HTTP input validation is primarily expressed with Huma struct tags in transport DTOs, such as `format:"email"`, `required:"true"`, and password length constraints in `internal/transport/http/auth/register.go`. Handlers translate application errors to Huma HTTP errors, for example duplicate email to `409` and invalid credentials to `401`.

Application and domain packages define sentinel errors. Repositories translate pgx-specific errors into domain/application errors, such as unique violation `uq_users_email` to `auth.ErrEmailAlreadyExists`. HTTP panic recovery is handled in `internal/transport/http/middleware/recovery.go`. `WriteProblem` exists for RFC 7807-style responses, but Huma errors are the current route pattern.

## Security & Configuration Tips

Secrets and runtime settings come from environment variables or `.env`; defaults and validation live in `internal/config/config.go`. Never commit real `.env` files, JWT secrets, database passwords, trace endpoints with credentials, or production pseudonym keys.

The observability compose setup requires `LOG_PSEUDONYM_KEY`, `DATABASE_PASSWORD`, `AUTH_JWT_SECRET`, and `GF_SECURITY_ADMIN_PASSWORD`. Passwords are hashed with Argon2id using `AUTH_ARGON2ID_*` settings. JWT access tokens use HS256 with `AUTH_JWT_SECRET` and `AUTH_ACCESS_TOKEN_TTL`. Configuration fields for login rate limits exist, but no rate-limiting middleware is currently wired.

## Database & Persistence

The database is PostgreSQL with TimescaleDB in Docker (`timescale/timescaledb:latest-pg16`). The initial Goose migration enables `pgcrypto`, `citext`, and `timescaledb`, then creates users, projects, project members, probes, selector-based ping checks, key/value labels, effective probe-check rows, ping results, and the ping results hypertable. Traceroute schema is intentionally not present yet.

Add schema changes as timestamped Goose migrations under `db/migrations/`, following the pattern in `db/migrations/README.md`, such as `202604300001_create_example_table.sql`. Add typed SQL queries under `db/query/*.sql`, then run `just backend-sqlc`. Do not edit `internal/infrastructure/postgres/sqlc/*.go` manually. Keep repositories responsible for mapping sqlc rows and pgx errors into domain/application types.

## External Integrations

Current backend integrations are PostgreSQL/TimescaleDB and optional OTLP trace export to VictoriaTraces, as shown in `deployments/docker/compose.yaml` and `compose.observability.yaml`. Docker observability also includes VictoriaLogs with Vector container log collection. No third-party API SDKs, queues, email services, payment providers, or object storage clients are currently implemented.

## Commit & Pull Request Guidelines

Recent git history uses short conventional-style subjects such as `feat: implement login endpoint`, `fix: remove emoty spaces for docs`, and `refactor: refactor logging system to be better implement`. Prefer `feat:`, `fix:`, `refactor:`, `test:`, `docs:`, or `chore:` with an imperative summary.

PRs should include a clear change summary, related issue or ticket when applicable, validation commands run, and notes for migrations, environment changes, public API changes, deployment changes, or breaking behavior.

## Agent-Specific Instructions

Before changing backend code, inspect the nearest existing package patterns and use the current layers. Keep changes minimal and scoped. Update tests and documentation when behavior, routes, config, migrations, logging, or tracing change.

If a backend code, command, architecture, configuration, dependency, migration, logging, tracing, or testing change makes this guide inaccurate, update `server/AGENTS.md` in the same change.

Do not introduce new dependencies unless the repository evidence shows the existing stack cannot handle the task. Avoid public API, database schema, or deployment changes without explicitly documenting impact and required commands. Preserve `context.Context` propagation, request-scoped logging, OpenTelemetry spans, Huma validation, and sentinel-error mapping. Do not overwrite generated sqlc code by hand or commit local artifacts from `tmp/`, `bin/`, or `.env`.
