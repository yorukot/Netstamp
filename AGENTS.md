# Repository Guidelines

## Project Overview

Netstamp is a pnpm workspace with a Go backend and React/Vite frontend. Use this root guide for repository-wide orientation only. Before making task-specific changes, read the closest area guide:

- Backend, database, migrations, logging, or API work: read `server/AGENTS.md`.
- Web app, shared UI package, client styling, or browser behavior work: read `web/AGENTS.md`.

Only proceed from root guidance when the task is clearly limited to workspace metadata, documentation, deployment files, or cross-project maintenance.

## Project Structure

- `server/`: Go API service, commands in `cmd/`, app code in `internal/`, SQL and migrations in `db/`.
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
- `pnpm generate:openapi`: regenerate `docs/public/openapi.json` from backend Huma routes.
- `just build`: build backend, web, and docs.
- `just test`: run available tests, currently backend tests.
- `just lint`: run web ESLint and backend linting.
- `pnpm format` or `just format`: format repository files with Prettier.

## Repository Conventions

Follow `.editorconfig` and local tool formatters. JavaScript, TypeScript, JSX, CSS, JSON, and Astro files use tabs with width 2 and Prettier. Go files use `gofmt`; backend linting is configured in `golangci.yaml`. Keep generated, migration, and API changes scoped to the relevant subproject.

When code, commands, architecture, configuration, or project structure changes make an `AGENTS.md` inaccurate, update the affected guide in the same change.

## Commit And PR Guidance

Recent history uses short conventional-style subjects such as `feat: implement login endpoint` and `refactor: refactor logging system to be better implement`. Prefer `feat:`, `fix:`, `refactor:`, `docs:`, or `chore:` with an imperative summary. PRs should describe the changed area, list validation commands run, link related issues, and include screenshots for visible web UI changes.
