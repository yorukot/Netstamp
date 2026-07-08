# Repository Guidelines

## Project Structure & File Organization

This guide applies to `server/`, the Go backend for the Netstamp workspace. The root workspace also contains `web/`, `docs/`, and `packages/ui/`; use this file only for backend API, database, logging, tracing, and server runtime work.

- `cmd/controller/main.go`: controller process entry point. It creates the shutdown context, calls `app.New`, runs the app, and syncs the logger.
- `cmd/agent/main.go`: standalone probe agent process entry point. It loads `NETSTAMP_PROBE_*` configuration, starts the runtime client, scheduler, worker pool, ping executor, and result submitter.
- `cmd/migrate/main.go`: Goose migration CLI for `status`, `up`, and `down`.
- `internal/controller/app/`: composition root and lifecycle. `wiring.go` wires repositories, application services, event recorders, background workers, and HTTP dependencies. `lifecycle.go` starts and gracefully stops the HTTP listener.
- `internal/controller/transport/http/`: chi HTTP routing, local HTTP helpers, auth, project, label, probe management, result query, probe runtime, system health routes, and middleware.
- `internal/controller/transport/http/frontend.go`: optional static React app serving enabled by `WEB_DIR`; the self-host Docker image defaults this to `/app/web`.
- `internal/agent/cli/` and `internal/agent/service/`: Cobra-based probe agent CLI and Linux/systemd service management. `service install` writes probe env and systemd unit files; `update` downloads and replaces the installed Linux agent binary; the runtime still starts through `run` or no arguments.
- `internal/controller/transport/http/handler/install/`: unauthenticated Linux probe install assets. It serves the thin binary installer script, uninstaller wrapper, and the Linux amd64/arm64 agent binaries built into the backend image.
- `internal/controller/application/auth/`, `internal/controller/application/project/`, `internal/controller/application/label/`, `internal/controller/application/check/`, `internal/controller/application/probe/`, `internal/controller/application/result/{latest,ping,tcp,traceroute,shared}/`, `internal/controller/application/proberuntime/`, `internal/controller/application/pingquery/`, and `internal/controller/application/tcpquery/`: controller use cases, ports, DTOs, errors, shared query policy, and feature orchestration. `internal/controller/application/result/` itself is a thin compatibility facade for HTTP wiring.
- `internal/domain/identity/`, `internal/domain/project/`, `internal/domain/label/`, `internal/domain/check/`, `internal/domain/ping/`, `internal/domain/tcp/`, `internal/domain/traceroute/`, and `internal/domain/probe/`: stable domain structs and domain-level sentinel errors.
- `internal/agent/app/`: probe agent composition root. It wires config, slog logging, runtime client, scheduling, workers, executors, and runtime lifecycle.
- `internal/agent/config/`, `internal/agent/runtime/`, `internal/agent/scheduling/`, and `internal/agent/worker/`: probe agent environment loading, lifecycle orchestration, assignment scheduling, worker pool, result queue, and result submission. These packages use existing domain models for assignments, checks, ping/TCP/traceroute config and results, and IP family values.
- `internal/agent/infrastructure/`: probe agent outbound integrations, including the controller runtime HTTP client, raw ICMP ping executor, and local network status discovery.
- `internal/controller/infrastructure/`: PostgreSQL repositories and pool helpers, JWT issuing, Argon2id password hashing, and probe secret generation/verification.
- `internal/controller/logger/`: controller zap logging helpers and application event recording. `internal/platform/normalize/` and `internal/platform/observability/` are shared helpers that do not depend on controller features.
- `internal/architecture/`: architecture-boundary tests that enforce backend import direction. Update these when the intended architecture changes, not to bypass an individual feature.
- `ARCHITECTURE.md`: concise backend architecture note for the modular monolith, dependency rules, use-case service pattern, and async consistency guidance.
- `db/migrations/`: Goose SQL migrations. `db/query/`: sqlc query files. Generated sqlc Go files live in `internal/controller/infrastructure/postgres/sqlc/`.
- `tmp/` and `bin/`: local build artifacts; do not edit them as source.

The backend can serve a built frontend directory when `WEB_DIR` is set. In the self-host Docker image built by `deployments/docker/Dockerfile`, this defaults to `/app/web`; `server/Dockerfile` remains backend-only, and local backend development leaves `WEB_DIR` unset so Vite owns the web app.

## System Architecture Overview

The backend is a single Go service with one listener: HTTP on `HTTP_ADDR`. `internal/controller/app.New` loads validated configuration, creates a zap logger, initializes OpenTelemetry, opens a pgx pool, builds application services, and creates the HTTP server. `internal/controller/app.Run` starts the server and coordinates graceful shutdown with `errgroup`.

HTTP uses chi middleware and standard handlers under `/api/{version}`. `internal/controller/app/wiring.go` passes `cfg.APIVersion` (`API_VERSION`) into the router. The published OpenAPI contract is authored in `api/` with TypeSpec, emitted to `docs/public/openapi.json`, and copied into `internal/controller/transport/http/openapi/openapi.json` for runtime serving. Scalar API docs are always exposed at `/api/{version}/docs`, and the embedded schema is always exposed at `/api/{version}/openapi.json`. API system routes are `/` and `/healthz` under the versioned API prefix; when `WEB_DIR` is set, top-level `/healthz` is also routed to API health, and non-API browser routes fall back to the React `index.html`. Install routes are `/install/agent.sh`, `/install/uninstall-agent.sh`, `/install/netstamp-agent-linux-amd64`, and `/install/netstamp-agent-linux-arm64`; they are unauthenticated and install only the agent binary, while probe ID and plaintext secret are supplied later to `netstamp-agent service install`. Auth routes are `/auth/register`, `/auth/login`, `/auth/logout`, `/auth/password-resets`, and `/auth/me`; `/auth/me`, `/users/me/*`, project routes, project label routes, project check routes, project probe creation, and project-scoped result routes are protected by the session auth middleware in `internal/controller/transport/http/middleware`. Probe runtime routes live under `/runtime/probes/{probe_id}/*` and use `Authorization: Probe <secret>` with the probe's one-time plaintext secret, not user JWT auth.

The current auth request flow is:

`HTTP request -> chi route -> internal/controller/transport/http/auth handler -> internal/controller/application/auth.Service -> internal/controller/infrastructure/postgres/user repository -> sqlc.Queries -> PostgreSQL`

Probe runtime requests follow the same layer boundaries:

`HTTP request -> chi route -> internal/controller/transport/http/proberuntime handler -> internal/controller/application/proberuntime.Service -> internal/controller/infrastructure/postgres/{probe,ping,tcp,traceroute} repositories -> sqlc.Queries -> PostgreSQL`

No GraphQL, external message queues, payment, object-storage integrations, or functional DNS/HTTP probe executors are currently defined. The controller supports ping, TCP connect, and traceroute check definitions, assignment payloads, runtime result ingestion, result queries, typed result persistence, alert incident evaluation, a DB-backed assignment refresh worker, and a notification outbox worker for webhook, Slack incoming webhook, Discord, Telegram, and SMTP email deliveries. The standalone probe agent executes ping, TCP connect, and traceroute checks through the shared scheduler, worker pool, result queue, and runtime result submitter.

## Layer Responsibilities

- Transport (`internal/controller/transport/http`): route registration, request binding, response DTOs, protocol status mapping, and middleware. Do not put database calls or business rules here.
- Application (`internal/controller/application/*`): business orchestration, service methods, ports, app errors, feature event semantics, and use-case spans. Depend on interfaces, not concrete pgx, HTTP framework, or JWT types.
- Domain (`internal/domain`): stable domain structs, domain-level sentinel errors such as `identity.ErrUserNotFound`, and shared domain policy such as project role/action permission checks.
- Infrastructure (`internal/controller/infrastructure/postgres`, `internal/controller/infrastructure/security`): pgx/sqlc persistence, database error translation, JWT HS256 tokens, and Argon2id password hashing. Infrastructure packages should import domain packages for models and sentinel errors, but should not import application packages or depend on application DTOs.
- Config (`internal/controller/config`): Viper-based environment loading, defaults, and validation. Add new env keys here and mirror them in `.env.example` when operators need to set them.
- Cross-cutting (`internal/controller/logger`, `internal/platform/observability`): request-scoped loggers, auth event recording, trace fields, tracer provider setup, and span helpers.
- Agent runtime (`internal/agent/runtime`, `internal/agent/scheduling`, `internal/agent/worker`): keep probe process orchestration, scheduling state, worker queues, and result submission independent of concrete HTTP or ping libraries. Agent infrastructure packages implement runtime/worker ports and may import external libraries such as `x/net` ICMP helpers.
- Agent runtime config: heartbeat, assignment polling, retry timing, worker count, result buffering, and observability endpoints are local probe-agent settings only. The controller does not send runtime tuning config in `hello` or assignment responses. Configure probes with `NETSTAMP_PROBE_*` environment variables such as `NETSTAMP_PROBE_HEARTBEAT_INTERVAL`, `NETSTAMP_PROBE_ASSIGNMENT_POLL_INTERVAL`, `NETSTAMP_PROBE_INITIAL_BACKOFF`, `NETSTAMP_PROBE_MAX_BACKOFF`, `NETSTAMP_PROBE_MAX_ATTEMPTS`, `NETSTAMP_PROBE_METRICS_ADDR`, and `NETSTAMP_PROBE_PPROF_ADDR`.

## Application Package Taxonomy

Classify an application package before adding files or copying another package's shape. The expected structure depends on the package type:

- Command features handle state changes, authorization, and use-case-specific business errors. They must keep service orchestration, DTOs, ports, errors, validation, tracing, and flow helpers explicit with `service.go`, `dto.go`, `ports.go`, `errors.go`, `validate.go`, `trace.go`, and `flow.go`.
- Query features handle reads, aggregation, and time-series query policy. They should keep DTOs, ports, validation, errors where needed, and tracing explicit, but should not add an empty `flow.go` unless they also own command-style failure semantics.
- Orchestrators and workers coordinate background or cross-feature work. They should use `service.go` or `worker.go`, focused ports, and tracing/error conventions appropriate to the workflow rather than pretending to be HTTP-facing features.
- Shared application support packages provide application-layer helpers or policy. They should be named by purpose and do not need `service.go`, `dto.go`, `trace.go`, or `flow.go`.

`flow.go` is mandatory for command features. It should centralize span lifecycle, normalized identifiers, success/failure outcome handling, business-versus-technical failure classification, sentinel error mapping, and application event recording when the package has an event recorder. Do not add placeholder flow, trace, error, or DTO files to query or shared-support packages just to make directories look identical.

Current package classification:

- Command features: `auth`, `project`, `label`, `check`, `probe`, `proberuntime`, `assignment`, `user`, `alert`, and the mutating public status page/element use cases in `publicstatus`.
- Query features: `result/latest`, `result/ping`, `result/tcp`, `result/traceroute`, and the public read use cases in `publicstatus`.
- Orchestrators and workers: `alerteval`, `notification`, and assignment refresh worker behavior.
- Shared application support: `validation`, `tx`, `workerloop`, `pingquery`, `tcpquery`, `result/shared`, and the thin `result` compatibility facade.

## Cross-Feature Repository Reuse

When one application feature needs data or access checks that are already owned by another feature, define a narrow capability port in the consuming application package and wire an existing repository or dedicated adapter into it from `internal/controller/app/wiring.go`. Do not duplicate SQL, UUID/slug parsing, membership checks, or repository methods just because the consuming feature has its own repository.

Keep feature services independent: an application service should not call another feature's application service just to reuse repository behavior, because that imports the other feature's event semantics, tracing spans, and use-case policy. The consuming service should own its own events, spans, sentinel-error mapping, and business policy while depending on a small interface such as project access, user lookup, or membership lookup.

Repository packages should stay aligned with the data/capability they own. If two features need the same persistence capability, prefer reusing the existing repository through an application port or extracting a small infrastructure helper over copying sqlc calls into a second repository package. For example, a probe use case that needs to resolve a project for the current user should depend on a project-access port implemented by the project repository, while the probe repository remains focused on probe persistence.

Assignment matching, synchronization, and cleanup are application orchestration responsibilities. Feature repositories should not refresh or delete cross-feature assignment rows inside their own transactions; expose focused persistence capabilities and let the assignment service or application port coordinate assignment state after successful feature mutations. Assignment refresh uses `assignment_refresh_jobs` as a DB-backed retry table while keeping synchronous refresh for low-latency API behavior.

Project-scoped permission decisions should use the project domain policy rather than package-local role predicates. Application services remain responsible for use-case-specific invariants, event reasons, and error mapping.

## Authentication & Project Permissions

The backend currently uses authenticated users plus project-scoped membership roles. There is no global admin role, organization/project hierarchy, or route-level scope system. HTTP auth middleware proves identity only; application services own authorization.

Protected chi routes use `internal/controller/transport/http/middleware.RequireAuth`. It reads the `netstamp_session` HTTP-only cookie, verifies the JWT value through the auth `TokenVerifier`, and stores `identity.AccessTokenClaims` in the request context. Login and registration set this cookie, and `/auth/logout` clears it. Transport handlers read `claims.Subject` as `CurrentUserID` and pass it into application service inputs. Keep role checks out of HTTP handlers except for translating application errors into HTTP responses.

Project membership is stored in `project_members` with role enum values defined by `internal/domain/project`: `owner`, `admin`, `editor`, and `viewer`. Project repository access methods such as `ListProjectsForUser`, `GetProjectForUser`, and `GetMemberRole` join existing project membership with non-deleted projects, so soft-deleted projects are not accessible. Removing a member hard-deletes the membership row. Creating a project creates the creator's `owner` membership in the same repository operation. New non-owner access starts as a pending row in `project_invites`; accepting an invite creates the membership in the same repository transaction that resolves the invite.

All project-scoped permission rules belong in `internal/domain/project/permission.go`:

- `Can(role, action)` is the canonical action policy.
- `CanAssignRole(actorRole, targetRole)` controls member role assignment.
- `IsValidRole(role)` validates known role values.

Current action policy:

- `read:project`: any valid project role.
- `write:project`: `owner` and `admin`.
- `delete:project`: `owner` only.
- `write:project_members`: `owner` and `admin`.
- `write:project_labels`, `write:project_checks`, `write:project_probes`, and `create:probe`: `owner`, `admin`, and `editor`.

Current member and invite-management policy:

- Owners may assign `admin`, `editor`, or `viewer`, but not `owner`.
- Admins may assign `editor` or `viewer`, but not `owner` or `admin`.
- Admins cannot change an existing owner.
- Role changes must not remove the last active owner from a project.
- Owners cannot self-leave a project.
- Owners/admins create pending invites instead of directly adding users as members. The invitee must accept the invite before membership is created.

Feature services should enforce permissions after loading the project for the current user and before mutating project-scoped data:

- Project service: list/get require existing membership; update requires `write:project`; delete requires `delete:project`; invite create/list and member update require `write:project_members` plus assignability and last-owner checks where applicable. Member removal allows owner/admin-managed removal by policy and non-owner self-leave.
- Label service: list requires active project membership; create/update/delete require `write:project_labels`.
- Probe registry service: list/get require active project membership; create/update/delete and secret rotation require `write:project_probes`; label IDs are resolved inside the same project after the project permission check.
- Check service: list/get require active project membership; create/update/delete require `write:project_checks`.

Cross-feature authorization should use narrow application ports such as `ProjectAccess.GetProjectForUser` and `ProjectAccess.GetMemberRole`, usually implemented by the project repository and wired in `internal/controller/app/wiring.go`. Do not duplicate membership SQL or call another feature's application service just to check access.

Error mapping is intentionally conservative. Missing/invalid user auth cookies return `401`. Inaccessible or missing projects, users, members, invites, labels, checks, and probes generally map to `404` so project existence is not leaked through membership checks. Valid users without the required project role map to application `ErrForbidden` and HTTP `403`. Invalid role/input maps to `422`, and duplicate membership/invite or last-owner protection maps to `409`.

When adding a new project-scoped action, add it to the project domain policy first, update `internal/domain/project/permission_test.go`, then wire the relevant application service to call `domainproject.Can`. Add focused service tests for allowed roles, denied roles, role lookup failures, and any feature-specific invariants.

## Libraries & Dependencies

Direct backend dependencies are declared in `server/go.mod`.

- HTTP: `github.com/go-chi/chi/v5`, `otelhttp`, and local helpers under `internal/controller/transport/http/httpx`.
- Probe agent ping execution: raw ICMP using `golang.org/x/net/icmp`, `ipv4`, and `ipv6`; agent runtimes need raw socket privileges such as `CAP_NET_RAW` or root.
- Database: `github.com/jackc/pgx/v5`, `github.com/pressly/goose/v3`, and `github.com/google/uuid`.
- Config: `github.com/spf13/viper`.
- Auth/security: `github.com/golang-jwt/jwt/v4` and `golang.org/x/crypto/argon2`. Probe runtime authentication uses the hashed probe secret in `probe_credentials`; never pass or log plaintext probe secrets outside the HTTP auth boundary and verifier call.
- Logging: `go.uber.org/zap`.
- Tracing: OpenTelemetry SDK, trace API, and OTLP HTTP trace exporter.
- Tool tracking: `tools/` is a nested Go module that pins the sqlc CLI for code generation. `air` and `golangci-lint` are used by commands/config but are not pinned in `server/go.mod`.

## Logging Guidelines

Zap is configured in `internal/controller/logger/zap.go`. Every root logger includes `service`, `env`, and `version`; local env uses zap development config, other envs use production config. Valid `LOG_LEVEL` values are enforced in `internal/controller/config/config_validate.go`: `debug`, `info`, `warn`, `error`, `dpanic`, `panic`, and `fatal`.

Use request-scoped loggers from `logger.FromContext(ctx, fallback)` when handling requests. HTTP logging in `internal/controller/transport/http/middleware/logging.go` adds `request_id`, method, path, client address, user agent, status, bytes, duration, and trace fields.

Application-level events must follow the auth/project/label/check/probe/proberuntime/assignment/alert/publicstatus pattern:

- Define typed event names, actions, outcomes, reasons, and recorder ports in the application package `ports.go`.
- Keep zap out of application packages. Services call the package recorder interface; `internal/controller/logger` owns zap fields, privacy handling, and log levels.
- Pass concrete recorders from `internal/controller/app/wiring.go`. Do not silently install package-local no-op recorders; missing wiring should be visible in tests or startup composition.
- Use small package-internal flow helpers to keep `context.Context`, request-scoped loggers, OpenTelemetry spans, event metadata, and success/failure recording together.
- Log successful application events only for audit/security flows, not every expected endpoint. Auth register/login, project create/delete/member and invite access changes, check/label definition changes, alert rule/notification changes, and public status page/element changes are audit-worthy; normal successful reads/lists and probe create success are covered by the HTTP request logger.
- Expected business/security failures use package-internal helpers such as `businessFailure`, log at `warn`, and should not attach raw error details. Unexpected technical failures use helpers such as `technicalFailure`, log at `error`, record the span error with `span.RecordError`, and include the error field.
- Orchestrators and workers must not hide technical errors behind caller events. Alert evaluation errors go through `logger.AlertEvalEventRecorder`, while assignment refresh and notification outbox workers log swallowed `RunOnce` errors with `worker.name`, `worker.operation`, and `error` fields before continuing.
- Keep sentinel error imports centralized in each application package `errors.go` when an application layer needs to expose errors from another domain. Other files in that package and its HTTP handlers should use the application package error names.
- Never log raw passwords, password hashes, access tokens, JWT secrets, cookies, database passwords, probe plaintext secrets, probe secret hashes, raw request bodies, or raw personal data.

Auth security events must go through `logger.AuthEventRecorder`. It pseudonymizes email into `user.email_hash` using `LOG_PSEUDONYM_KEY`.

Project application events must go through `logger.ProjectEventRecorder`. Use these only for focused audit/security flows and failures where application semantics add value beyond HTTP request logs. Do not log project member or invite email addresses.

Label application events must go through `logger.LabelEventRecorder`. Label create, update, and delete successes are audit-worthy; successful list and resolve operations are covered by the HTTP request logger. Label failure events should preserve project and label identifiers when available, but must never include label key or value text.

Check application events must go through `logger.CheckEventRecorder`. Check create, update, and delete successes are audit-worthy; successful list and get operations are covered by the HTTP request logger. Check events should preserve project and check identifiers when available, but must never include check name, target, selector, or label text.

Alert application events must go through `logger.AlertEventRecorder`. Alert rule and notification create, update, delete, and notification test events are audit-worthy; successful list and get operations are covered by the HTTP request logger. Alert events should preserve project, rule, and notification identifiers when available, but must never include notification config, webhook URLs, condition JSON, check targets, selector text, or notification payloads.

Public status application events must go through `logger.PublicStatusEventRecorder`. Public status page and element create, update, and delete events are audit-worthy; successful public reads are covered by the HTTP request logger. Public status events should preserve project, page, and element identifiers when available, but must never include page titles, element titles, descriptions, rendered content, chart data, or incident text.

Alert evaluation application events must go through `logger.AlertEvalEventRecorder`. Alert evaluation records failure events only; successful evaluations are represented by spans and downstream incident/outbox state. Alert evaluation events should preserve project, probe, check, rule, and incident identifiers when available, but must never include notification payloads, condition JSON, check targets, selector text, metric summaries, or result samples.

Probe application events must go through `logger.ProbeEventRecorder`. Probe create currently records failure events only; probe update, delete, and secret rotation successes are audit-worthy. Successful probe list/get and create are covered by the HTTP request logger. Probe events must never include the plaintext secret or its hash.

Probe runtime application events must go through `logger.ProbeRuntimeEventRecorder`. Probe runtime currently records failure events only; successful hello, heartbeat, assignment polling, and result submission are covered by the HTTP request logger. Probe runtime events must never include plaintext secrets, secret hashes, agent version, IP addresses, raw result bodies, result error messages, check targets, or selector text.

## Tracing & Observability

OpenTelemetry tracing is configured in `internal/platform/observability/tracing/tracing.go`. The provider always samples locally and exports only when `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` is set. TraceContext and Baggage propagators are installed globally.

HTTP tracing is wired through `otelhttp.NewMiddleware` in `internal/controller/transport/http/router.go`, with span names from `internal/platform/observability/httptrace/span.go`. Application command flows create child spans through package `trace.go` and `flow.go` helpers. Query packages should create spans for read workflows and record technical failures with feature-specific failure reasons, while high-volume public reads should avoid application event logs unless they are audit-worthy. PostgreSQL repository methods use span helpers in `internal/controller/infrastructure/postgres/trace.go`.

Prometheus metrics for the controller are configured in `internal/platform/observability/metrics/metrics.go` and exposed at `/metrics` on the backend HTTP listener. The probe agent has separate opt-in metrics and pprof listeners configured by `NETSTAMP_PROBE_METRICS_ADDR` and `NETSTAMP_PROBE_PPROF_ADDR`; these are disabled by default. VictoriaMetrics scrapes configured endpoints from Docker compose and Grafana provisions a `VictoriaMetrics` Prometheus datasource.

Keep `context.Context` as the first parameter for request, service, repository, and token operations so trace context and request loggers propagate across layers. New database calls should either use existing DB span helpers or add equivalent attributes without recording raw SQL parameters.

## Build, Test, and Development Commands

Commands below come from the root `Justfile`, root `package.json`, `server/.air.toml`, `server/Dockerfile`, and compose files under `deployments/docker/`.

- `pnpm install`: install workspace dependencies; root `package.json` enforces pnpm.
- `just backend-dev` or `pnpm dev:server`: run Air hot reload using `server/.air.toml`.
- `just backend-build` or `pnpm build:server`: build `server/bin/controller` from `./cmd/controller` and `server/bin/agent` from `./cmd/agent`.
- `just api-openapi` or `pnpm generate:openapi`: write the TypeSpec OpenAPI schema to `docs/public/openapi.json`, copy it to the backend embedded OpenAPI artifact, and regenerate `web/src/shared/api/openapi.d.ts`.
- `just backend-test` or `pnpm test:server`: run `go test ./...` inside `server/`.
- `NETSTAMP_TEST_DATABASE_URL=postgres://netstamp:netstamp@localhost:5432/netstamp?sslmode=disable just backend-test-integration`: run opt-in API E2E tests against a local PostgreSQL/TimescaleDB instance.
- `just backend-fmt`: run configured golangci formatters with `../golangci.yaml`.
- `just backend-lint`: run `golangci-lint` with `../golangci.yaml`.
- `just backend-lint-fix`: apply safe `golangci-lint` fixes.
- `just backend-sqlc`: regenerate sqlc code from `sqlc.yaml`.
- `just backend-migrate-status`, `just backend-migrate-up`, `just backend-migrate-down`: run `cmd/migrate`.
- `docker compose -f deployments/docker/compose.yaml up -d`: run the self-host stack from Docker Hub with PostgreSQL, one-shot migrations, and the Netstamp app serving both frontend and API.
- `docker compose -f deployments/docker/compose.observability.yaml up --build`: build and run the observability Docker stack with PostgreSQL, VictoriaTraces, VictoriaMetrics, VictoriaLogs, Vector, Grafana, nginx, controller, migrations, and the backend image's agent install artifacts.

Use `server/.env.example` as the controller env template and `server/probe.env.example` as the probe env template. `server/.gitignore` intentionally ignores `.env`, `.env.*`, `*.env`, `bin/`, `tmp/`, and `coverage.out`.

## Coding Style & Naming Conventions

Go files use tabs and `gofmt` per root `.editorconfig`. `golangci.yaml` enables gofumpt, goimports, and gci formatting with local imports grouped under `github.com/yorukot/netstamp`. Keep package names short and lowercase, matching existing packages such as `auth`, `postgres`, and `httpserver`.

Follow existing feature file names: `service.go`, `validate.go`, `ports.go`, `errors.go`, `trace.go`, `flow.go`, `handler.go`, `register.go`, `login.go`, and `*_test.go`. Keep application service input validation, normalization, constants, and field-level validation errors in `validate.go` when a feature has enough validation logic to make `service.go` hard to scan. Export only types needed across packages. Use sentinel errors named `Err...` and compare with `errors.Is`.

For HTTP feature packages, keep `handler.go` focused on the `Handler` type, constructor, chi route registration, and thin binding/writing wrappers. Put application-facing operation methods and their request/response DTOs in separate operation files, matching `internal/controller/transport/http/auth` patterns such as `register.go`, `login.go`, and `me.go`. Put response DTOs and mappers in `response.go` only when they are shared by multiple operation files. Do not concentrate all endpoint handler logic and DTOs in one large `handler.go`.

## Testing Guidelines

Tests use Go's standard `testing` package and live beside the code as `*_test.go`, for example `internal/controller/application/auth/service_test.go` and `internal/controller/logger/auth_events_test.go`. Existing unit and handler tests use package-local fakes and zap observer cores rather than external test frameworks.

Run backend tests with `just backend-test` or `cd server && go test ./...`. API E2E tests live in `internal/e2e`, are gated by the `integration` build tag, and require `NETSTAMP_TEST_DATABASE_URL` to point at a local PostgreSQL/TimescaleDB database that can create/drop temporary databases. After providing that database, run `NETSTAMP_TEST_DATABASE_URL=postgres://netstamp:netstamp@localhost:5432/netstamp?sslmode=disable just backend-test-integration`. The integration recipe uses verbose test output so harness and API-flow checkpoints are visible. Coverage thresholds are not currently defined. Add unit tests beside changed packages, and use E2E tests only for complete API workflows that need real HTTP, services, repositories, migrations, and database behavior.

## Error Handling & Validation

HTTP input validation uses a mix of transport parsing and application-owned validation. The HTTP layer handles routing, auth middleware, JSON decoding, basic Go type binding, and transport-only parsing such as query integer conversion and IP address strings. Request semantic validation such as required strings, length limits, enum values, numeric bounds, selector validity, UUID canonicalization, empty patch checks, and result consistency is owned by the relevant `internal/controller/application/*` package. Handlers translate application errors to repo-local HTTP errors, for example duplicate email to `409` and invalid credentials to `401`.

Transport request DTOs should keep only Go/JSON concerns. TypeSpec in `api/` is the source of truth for frontend and documentation consumers. Mirror simple domain constraints such as min/max length, numeric ranges, enum values, UUID formats, examples, and straightforward patterns in TypeSpec where practical. Do not try to duplicate complex business validation, selector grammar, cross-field consistency, authorization, or database-backed existence checks in transport tags; those remain in application/domain logic.

Application and domain packages define sentinel errors. Repository-originated sentinel errors belong in `internal/domain/*`; repositories translate pgx-specific errors into domain errors, such as unique violation `uq_users_email` to `identity.ErrEmailAlreadyExists`. Application services should import domain errors directly when handling repository/domain failures, and should not alias domain repository errors in application `errors.go`. Keep service-only errors, such as forbidden actions, credentials-invalid responses, and use-case validation sentinels, in the owning application package. HTTP panic recovery is handled in `internal/controller/transport/http/middleware/recovery.go`. `httpx.WriteProblem` and middleware `WriteProblem` produce RFC 7807-style responses.

Do not add narrow storage input types when existing domain models already express the data cleanly. Infrastructure should use domain models directly for repository inputs and outputs; if persistence-specific translation is needed, keep it inside the repository implementation rather than introducing storage-only domain DTOs or importing application-layer input types.

Semantic validation rules belong in domain packages and should use `github.com/yorukot/spvalidator` through domain `VN*` helpers. Do not add application-level validation helper packages for required strings, numeric ranges, UUIDs, or similar reusable rules. Application `validate.go` files should orchestrate feature input normalization, call domain validators directly, and wrap domain validation failures with field metadata. Avoid thin wrappers around single domain validator calls; add small helpers only when they remove real branching or multi-step normalization complexity.

For feature-specific application validation, keep input normalization and field-level validation errors in `internal/controller/application/<feature>/validate.go` once the logic is more than a couple of lines. Application services should return sentinel-compatible errors such as `ErrInvalidInput`, `ErrCheckNotFound`, `ErrProjectNotFound`, `ErrLabelNotFound`, or `ErrForbidden`, usually through package flow helpers that also record event outcome and failure reason. The application layer should not construct HTTP transport errors or status codes.

When an application error needs to tell the client exactly which request field is invalid, wrap the sentinel error with `internal/controller/application/validation.New` or `NewFields`. The wrapped error must still satisfy `errors.Is(err, ErrInvalidInput)` through `Unwrap`, while carrying `FieldError` metadata: `Field`, `Message`, and `Value`. Transport handlers should extract this metadata with `appvalidation.FieldErrors(err)` and convert it to `httpx.ErrorDetail` values. Feature handlers follow this pattern in their `context.go`, mapping fields to locations such as `body.name`, `body.selector`, `body.labelIds`, `path.ref`, or `body` for whole-body validation failures like an empty patch.

Keep public error detail conservative. Not-found conditions for inaccessible projects, users, labels, and checks should remain collapsed to generic `404 not found`; forbidden role checks map to `403`; validation maps to `422` with field details only when `FieldError` metadata exists; unknown technical errors map to `500` with a route-level fallback message. Technical error details belong in logs and spans, not response bodies.

## Security & Configuration Tips

Secrets and runtime settings come from environment variables or `.env`; defaults and validation live in `internal/controller/config/config.go`. Assignment refresh worker behavior is controlled by `ASSIGNMENT_REFRESH_WORKER_ENABLED`, `ASSIGNMENT_REFRESH_WORKER_INTERVAL`, `ASSIGNMENT_REFRESH_WORKER_BATCH_SIZE`, and `ASSIGNMENT_REFRESH_WORKER_STALE_TIMEOUT`. `WEB_DIR` is optional and should point at built frontend static files only in packaged/self-host deployments. Never commit real `.env` files, JWT secrets, database passwords, trace endpoints with credentials, or production pseudonym keys.

The observability compose setup requires `LOG_PSEUDONYM_KEY`, `DATABASE_PASSWORD`, `POSTGRES_EXPORTER_PASSWORD`, `AUTH_JWT_SECRET`, and `GF_SECURITY_ADMIN_PASSWORD`. Passwords are hashed with Argon2id using `AUTH_ARGON2ID_*` settings. JWT session cookies use HS256 with `AUTH_JWT_SECRET` and `AUTH_ACCESS_TOKEN_TTL`. Password reset links use `PUBLIC_WEB_BASE_URL` when set, otherwise the request origin, and reset request throttling is controlled by `AUTH_PASSWORD_RESET_RATE_LIMIT_WINDOW`, `AUTH_PASSWORD_RESET_IP_LIMIT`, and `AUTH_PASSWORD_RESET_EMAIL_LIMIT`.

## Database & Persistence

The database is PostgreSQL with TimescaleDB in Docker. The compose files default `TIMESCALEDB_IMAGE` to the lightweight `timescale/timescaledb:latest-pg16` image and mount data at `/var/lib/postgresql/data`. Do not switch back to the heavier `timescale/timescaledb-ha` image unless the schema needs extensions or services that are not present in the lightweight image.

The initial Goose migration enables `pgcrypto`, `citext`, and `timescaledb`, then creates users, projects, project members, probes, selector-based ping checks, key/value labels, effective probe-check rows, ping results, and the ping results hypertable. Follow-up migrations add traceroute check configuration plus traceroute run and hop result tables, probe location naming, project invites, TCP connect check configuration/results, raw result retention, and Ping-specific Timescale rollups. Current Ping long-term queries use `ping_result_rollups_1m` for safely re-aggregated result count, success count, sent/received count, average RTT sum/count, min RTT, and max RTT data. Older unused continuous aggregate views for Ping/TCP/traceroute result rollups, Ping sample density, traceroute topology, and the traceroute topology observation hypertables are intentionally dropped by the Ping aggregate redesign migration; the legacy Ping RTT sample observation hypertable and trigger are also dropped because Ping histogram/sample-density aggregates are not part of the current backend API. Raw result hypertables and traceroute hop hypertables use 1-day chunk intervals with 3-day retention. Traceroute also keeps permanent 1-minute sampled run snapshots in `traceroute_sampled_runs_1m`; a Timescale user-defined job refreshes recent buckets by selecting the latest run per probe/check/bucket and storing ordered hop JSON. Result series downsampling uses core TimescaleDB `time_bucket`; `timescaledb_toolkit` and DB-side `lttb` are intentionally not required. Probes and checks keep public UUIDs but also have internal bigint IDs, while `ping_results`, `tcp_results`, and `traceroute_results` store bigint probe/check dimensions with `(probe_id, check_id, started_at)` as their primary key and no generated result ID. Traceroute hop rows store per-hop aggregate RTT fields and `rtt_samples_ms` arrays with `(probe_id, check_id, started_at, hop_index)` as their primary key. Because TimescaleDB does not support foreign keys from one hypertable to another, `traceroute_result_hops` does not keep a database foreign key to `traceroute_results`; repository writes create the run and hop rows in the same transaction, and query paths select runs first before joining hops.

Add schema changes as timestamped Goose migrations under `db/migrations/`, following the pattern in `db/migrations/README.md`, such as `202604300001_create_example_table.sql`. Add typed SQL queries under `db/query/*.sql`, then run `just backend-sqlc`. Do not edit `internal/controller/infrastructure/postgres/sqlc/*.go` manually. Keep repositories responsible for mapping sqlc rows and pgx errors into domain/application types.

## External Integrations

Current backend integrations are PostgreSQL/TimescaleDB, optional OTLP trace export to VictoriaTraces, Prometheus-compatible metrics scraped by VictoriaMetrics, VictoriaLogs for Grafana log panels, outbound HTTPS alert notifications, and optional SMTP email delivery when `SMTP_*` settings are configured. Docker observability also includes Vector container log collection, PostgreSQL metrics via `postgres_exporter`, and a probe-agent dashboard for agent metrics when a probe exposes `NETSTAMP_PROBE_METRICS_ADDR`. The observability stack mounts `deployments/docker/postgres-exporter/queries.yaml` to backfill dashboard-specific activity-state metrics. No third-party API SDKs, message queues, payment providers, or object storage clients are currently implemented.

## Commit & Pull Request Guidelines

Recent git history uses short conventional-style subjects such as `feat: implement login endpoint`, `fix: remove emoty spaces for docs`, and `refactor: refactor logging system to be better implement`. Prefer `feat:`, `fix:`, `refactor:`, `test:`, `docs:`, or `chore:` with an imperative summary.

PRs should include a clear change summary, related issue or ticket when applicable, validation commands run, and notes for migrations, environment changes, public API changes, deployment changes, or breaking behavior.

## Agent-Specific Instructions

Before changing backend code, inspect the nearest existing package patterns and use the current layers. Keep changes minimal and scoped. Update tests and documentation when behavior, routes, config, migrations, logging, or tracing change.

If a backend code, command, architecture, configuration, dependency, migration, logging, tracing, or testing change makes this guide inaccurate, update `server/AGENTS.md` in the same change.

Do not introduce new dependencies unless the repository evidence shows the existing stack cannot handle the task. Avoid public API, database schema, or deployment changes without explicitly documenting impact and required commands. Preserve `context.Context` propagation, request-scoped logging, OpenTelemetry spans, application validation, and sentinel-error mapping. Do not overwrite generated sqlc code by hand or commit local artifacts from `tmp/`, `bin/`, or `.env`.
