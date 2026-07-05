# Backend Architecture

Netstamp backend is a modular monolith. Keep one deployable controller service and one probe agent binary, while preserving clear package boundaries inside the Go module.

## Target Shape

- `internal/controller/transport/http`: HTTP routing, request binding, response mapping, auth middleware, and protocol status translation.
- `internal/controller/application/*`: use-case orchestration, feature ports, application errors, validation flow, event semantics, and tracing spans.
- `internal/domain/*`: stable domain models, validation, normalization, and policies.
- `internal/controller/infrastructure/*`: adapters for PostgreSQL, security, and notification delivery.
- `internal/controller/app`: composition root. It wires concrete adapters into application services and owns lifecycle startup/shutdown.
- `internal/agent/*`: standalone probe runtime. It may share domain models, but it must not depend on controller application internals.

This is a pragmatic hexagonal architecture: application packages define the ports they need, infrastructure packages implement them, and the app composition root connects the two.

## Application Package Taxonomy

Application packages are not all the same shape. Classify each package before adding or reorganizing files:

- Command features handle state changes, authorization, and use-case-specific business errors. They must keep service orchestration, DTOs, ports, errors, validation, tracing, and command flow helpers explicit with `service.go`, `dto.go`, `ports.go`, `errors.go`, `validate.go`, `trace.go`, and `flow.go`.
- Query features handle reads, aggregation, and time-series query policy. They should keep DTOs, ports, validation, errors where needed, and tracing explicit, but should not add an empty `flow.go` unless they also own command-style failure semantics.
- Orchestrators and workers coordinate background or cross-feature work. They should use `service.go` or `worker.go`, focused ports, and tracing/error conventions appropriate to the workflow rather than pretending to be HTTP-facing features.
- Shared application support packages provide application-layer helpers or policy. They should be named by purpose and do not need `service.go`, `dto.go`, `trace.go`, or `flow.go`.

`flow.go` is mandatory for command features. It should centralize span lifecycle, normalized identifiers, success/failure outcome handling, business-versus-technical failure classification, sentinel error mapping, and application event recording when the package has an event recorder.

Current controller application classification:

- Command features: `auth`, `project`, `label`, `check`, `probe`, `proberuntime`, `assignment`, `user`, `alert`, and the mutating page/element use cases in `publicstatus`.
- Query features: `result/latest`, `result/ping`, `result/tcp`, `result/traceroute`, and public read use cases in `publicstatus`.
- Orchestrators and workers: `alerteval`, `notification`, and assignment refresh worker behavior.
- Shared application support: `validation`, `tx`, `pingquery`, `tcpquery`, `result/shared`, and the thin `result` compatibility facade.

## Dependency Rules

- Domain packages must not import controller, agent, platform, or command packages.
- Application packages must not import HTTP transport, infrastructure, app, config, logger, agent, or command packages.
- PostgreSQL repository packages must not import application DTOs or transport packages. Security and notification adapters may reference the application port or result types they implement until those contracts are moved to a narrower shared package.
- Transport packages may import application packages, but must not contain business rules or database calls.
- Cross-feature reuse should happen through narrow application ports, not by calling another feature's service for its side effects.

## Use-Case Service Pattern

Application services should model complete use cases, not CRUD wrappers. A mutating use case usually follows this flow:

1. Normalize and validate input.
2. Load project/user/probe/check state through ports.
3. Enforce domain policy.
4. Persist state through a focused repository port.
5. Trigger required assignment, alert, or notification orchestration.
6. Record application events and span errors.

Keep request/response DTO details in transport packages. Keep SQL and pgx/sqlc details in infrastructure packages.

## Async Consistency

Use workers and reconcilers for side effects that can be retried safely, such as notification delivery and assignment refresh. Assignment refresh writes `assignment_refresh_jobs` before the synchronous refresh/delete path so the worker can repair derived probe-check assignment state after partial failures. Synchronous refresh may remain for low-latency API behavior, but durable background repair should exist for cross-feature derived state that can become stale after partial failures.
