# Repository Guidelines

## Project Overview

Netstamp is a pnpm workspace with a Go backend and React/Vite frontend. Use this root guide for repository-wide orientation only. Before making task-specific changes, read the closest area guide:

- Backend, database, migrations, logging, or API work: read `server/AGENTS.md`.
- API contract or OpenAPI generation work: read `api/AGENTS.md`.
- Web app, shared UI package, client styling, or browser behavior work: read `web/AGENTS.md`.
- Visual design, layout, product UI, or styling direction work: read `design.md`.

Only proceed from root guidance when the task is clearly limited to workspace metadata, documentation, deployment files, or cross-project maintenance.

## Project Structure

- `server/`: Go backend module with controller/probe commands in `cmd/`, private code in `internal/`, SQL and migrations in `db/`.
- `api/`: TypeSpec API contract that emits `docs/public/openapi.json`.
- `web/`: React 19 + Vite app source in `web/src/`.
- `packages/ui/`: shared React UI components and design tokens exported as `@netstamp/ui`.
- `docs/`: Astro public site and Markdown documentation. The docs build also publishes static Storybook output for shared UI components.
- `deployments/`: deployment and Docker-related configuration.
- `Justfile`: canonical local task runner for backend, web, docs, lint, build, and test commands.

## Common Commands

- `pnpm install`: install workspace dependencies. The repo enforces pnpm via `preinstall`.
- `just dev` or `pnpm dev:server`: start the backend with Air hot reload.
- `just web-dev` or `pnpm dev:web`: start the Vite web app.
- `just docs-dev` or `pnpm dev:docs`: start the documentation site.
- `pnpm dev:storybook`: start Storybook for `@netstamp/ui` components.
- `pnpm generate:openapi`: regenerate `docs/public/openapi.json` from TypeSpec and `web/src/shared/api/openapi.d.ts` from that contract.
- `just build`: build backend, web, and docs.
- `just test`: run available tests, currently backend tests.
- `just lint`: run web ESLint and backend linting.
- `pnpm format` or `just format`: format repository files with Prettier.

## Repository Conventions

Follow `.editorconfig` and local tool formatters. JavaScript, TypeScript, JSX, CSS, JSON, and Astro files use tabs with width 2 and Prettier. Go files use `gofmt`; backend linting is configured in `golangci.yaml`. Keep generated, migration, and API changes scoped to the relevant subproject.

When code, commands, architecture, configuration, or project structure changes make an `AGENTS.md` inaccurate, update the affected guide in the same change.

## Commit And PR Guidance

Use Linux kernel/Git-style commit subjects with an area, subsystem, or component prefix:

```text
area: concise patch summary
sub/sys: concise patch summary
```

The prefix should name the repository area changed, such as a directory, package, file, subsystem, or component. The summary after the colon should briefly describe what the patch does, because it becomes the first line shown in the git changelog. Keep it short, imperative, and specific. Use lowercase for the first word after the colon unless it is a proper noun, and do not end the subject with a period.

Examples:

```text
docs: clarify Storybook build ownership
web/routes: split route-level chunks
ui/field: fix select menu positioning
server/auth: validate session cookie
githooks.txt: improve the intro section
```

PRs should describe the changed area, list validation commands run, link related issues, and include screenshots for visible web UI changes.
