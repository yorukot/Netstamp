# Contributing to Netstamp

Thanks for helping improve Netstamp. Contributions to the controller, probes, web app, documentation, deployment tooling, tests, and translations are welcome.

## Before you start

- Search [existing issues](https://github.com/yorukot/netstamp/issues) before opening a new one.
- Use the bug or feature issue template when it fits your change.
- Report security issues privately by following [SECURITY.md](./SECURITY.md).
- Never include real credentials, probe secrets, session cookies, private hostnames, or production data in issues, tests, screenshots, or commits.

For a substantial feature or architecture change, open an issue first so the approach can be discussed before a large implementation is written.

## Development setup

Install these prerequisites:

- Git
- Node.js 22.12 or newer
- pnpm 11.9
- Go 1.26 or newer
- [just](https://github.com/casey/just)
- Docker Engine with Docker Compose v2 for integration and deployment checks

Clone and prepare the workspace:

```bash
git clone https://github.com/yorukot/netstamp.git
cd netstamp
pnpm install
```

Common development commands are:

```bash
just dev          # backend with hot reload
just web-dev      # React/Vite application
just docs-dev     # documentation site
pnpm dev:storybook
```

The root [AGENTS.md](./AGENTS.md) describes the repository layout. Read the closest area guide before changing code:

- [server/AGENTS.md](./server/AGENTS.md) for Go, database, migrations, logging, and backend APIs;
- [api/AGENTS.md](./api/AGENTS.md) for the TypeSpec contract and generated OpenAPI artifacts;
- [web/AGENTS.md](./web/AGENTS.md) for the web app and browser behavior;
- [packages/ui/AGENTS.md](./packages/ui/AGENTS.md) for shared UI components; and
- [docs/AGENTS.md](./docs/AGENTS.md) for documentation.

Read [design.md](./design.md) for visible product or documentation UI changes.

## Branches and commits

The permanent `main` branch is the only branch without a prefix. Other branch names must use one of these forms:

```text
feat/short-description
fix/short-description
ui/short-description
refactor/short-description
docs/short-description
test/short-description
chore/short-description
release/short-description
```

Descriptions use lowercase ASCII letters, digits, and single hyphens. Validate the current branch with:

```bash
pnpm check:branch-name
```

Commit subjects use an area or component prefix followed by a concise imperative summary:

```text
server/auth: validate session cookie
web/routes: split route-level chunks
docs: clarify the deployment workflow
```

Keep unrelated changes in separate commits and do not rewrite generated artifacts by hand.

## Making changes

- Keep changes focused and follow `.editorconfig` plus the local formatter and linter configuration.
- Add or update tests for changed behavior.
- Update documentation when commands, configuration, APIs, or user-visible behavior change.
- Add database changes as new migrations; do not edit an already shared migration.
- Change the TypeSpec source for API contract updates, then run `pnpm generate:openapi` to refresh every generated consumer.
- Use design tokens and the existing shared components for UI changes. Include before-and-after screenshots in the pull request for visible changes, but do not commit review-only screenshots.

## Validation

Run checks that cover the area you changed. Before requesting review, the full baseline is:

```bash
pnpm check:frontend-style
just lint
just test
just build
```

Backend database or integration changes should also run:

```bash
just backend-test-integration
```

If a command cannot run in your environment, explain why in the pull request and list the checks that did run.

## Pull requests

A pull request should:

- explain the problem and the chosen solution;
- identify the affected repository areas;
- link related issues;
- list validation commands and their results;
- call out migrations, configuration changes, compatibility concerns, and follow-up work; and
- include screenshots or a short recording for visible web or documentation UI changes.

Keep the pull request reviewable. Maintainers may ask for a large change to be split when its parts can be reviewed and released independently.
