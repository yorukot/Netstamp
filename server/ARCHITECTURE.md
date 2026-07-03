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
