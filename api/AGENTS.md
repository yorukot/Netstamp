# API Contract Guidelines

## Project Overview

The `api/` package owns the public Netstamp API contract in TypeSpec. It emits `docs/public/openapi.json`, which the docs site uses for API reference and the web app uses to generate TypeScript API types.

Use the backend application and domain packages as the source of truth for resources, request bodies, response bodies, and error states. Do not copy transport DTO tags as the contract source; TypeSpec owns the published contract and the backend serves the generated artifact at runtime.

## Structure

- `main.tsp`: service metadata, server base path, tag metadata, and imports.
- `models/`: shared scalars, auth, project/member, label, check, probe, runtime, and result models.
- `services/`: route groups split by application feature: system, auth, projects, labels, checks, probes, results, and probe runtime.
- `tspconfig.yaml`: OpenAPI 3.1 JSON emitter configuration. Output is `../docs/public/openapi.json`.

## Commands

- `pnpm --filter @netstamp/api generate:openapi`: compile TypeSpec and write `docs/public/openapi.json`.
- `pnpm --filter @netstamp/api check`: compile without emitting output.
- `pnpm --filter @netstamp/api format`: format TypeSpec files.
- `pnpm generate:openapi`: compile TypeSpec, copy the generated artifact into the backend embedded OpenAPI location, regenerate web API types, and format generated artifacts.

## Contract Conventions

Keep route paths aligned with the controller HTTP API under `/api/v1`, but model semantics from the application service inputs, outputs, and sentinel error mapping. Every operation should list all expected status-specific responses, including validation, auth, not-found, conflict, forbidden, unavailable, and internal failures where the service can return them.

Use shared response wrappers from `models/common.tsp` for consistent JSON, empty, session-cookie, and problem responses. Add concrete examples on request and response body models when adding or changing public payloads.
