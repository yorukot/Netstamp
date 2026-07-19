# Contributing to Netstamp

Thank you for contributing to Netstamp. This repository contains a Go backend, a TypeSpec API contract, a React/Vite web app, a shared React UI package, and an Astro documentation site. Keep each change focused, follow the closest area guide, and include the validation needed for the behavior you changed.

## Before You Start

1. Read the root [`AGENTS.md`](./AGENTS.md).
2. Read the closest area guide before editing:
   - Backend, database, migrations, logging, or server APIs: [`server/AGENTS.md`](./server/AGENTS.md)
   - TypeSpec and generated OpenAPI artifacts: [`api/AGENTS.md`](./api/AGENTS.md)
   - Web app and browser behavior: [`web/AGENTS.md`](./web/AGENTS.md)
   - Shared UI components: [`packages/ui/AGENTS.md`](./packages/ui/AGENTS.md)
   - Documentation site: [`docs/AGENTS.md`](./docs/AGENTS.md)
3. Read [`design.md`](./design.md) for visible product or documentation UI work.
4. Check the issue tracker and open pull requests before starting overlapping work.

## Development Setup

Prerequisites:

- Node.js 22.12 or newer
- pnpm 11.9
- Go 1.26 for backend work
- `just` for the repository task commands
- Docker when running PostgreSQL, TimescaleDB, or the full deployment stack

Install workspace dependencies:

```bash
pnpm install
```

The install step configures Git to use the repository hooks in `.githooks`.

Common development commands:

```bash
just dev
just web-dev
just docs-dev
pnpm dev:storybook
```

## Branch Names

Create work from an up-to-date `main` branch:

```bash
git switch main
git pull --ff-only
git switch -c feat/example-change
```

Except for the permanent `main` branch, every branch name must use this format:

```text
<type>/<short-kebab-case-description>
```

Allowed types:

| Prefix      | Use it for                                                       |
| ----------- | ---------------------------------------------------------------- |
| `feat/`     | New product, API, or platform behavior                           |
| `fix/`      | Bug fixes and regressions                                        |
| `ui/`       | Visual, layout, interaction, or design-system changes            |
| `refactor/` | Internal restructuring without intended behavior changes         |
| `docs/`     | Documentation-only changes                                       |
| `test/`     | Test-only additions or corrections                               |
| `chore/`    | Maintenance, tooling, dependencies, or repository administration |
| `release/`  | Release preparation and versioned release work                   |

The description must contain lowercase ASCII letters, digits, and single hyphens. Do not use spaces, underscores, uppercase letters, consecutive hyphens, or another slash.

Valid examples:

```text
feat/probe-registration
fix/session-cookie-expiry
ui/insight-scope-selectors
refactor/auth-session-store
docs/deployment-guide
test/traceroute-handler
chore/update-dependencies
release/v1-4-0
```

Invalid examples:

```text
feature/probe-registration
fix_session_cookie
UI/insight-selectors
docs/product/overview
backup0616
```

Validate the current branch locally with:

```bash
pnpm check:branch-name
```

The pre-push hook validates every branch ref being pushed, and pull requests run the same check in GitHub Actions.

## Commits

Use Linux kernel/Git-style subjects with an area, subsystem, or component prefix:

```text
area: concise patch summary
sub/sys: concise patch summary
```

Examples:

```text
docs: clarify Storybook build ownership
web/routes: split route-level chunks
ui/field: fix select menu positioning
server/auth: validate session cookie
```

Keep the first word after the colon lowercase unless it is a proper noun. Use an imperative, specific summary and do not end the subject with a period. Keep commits focused so reviewers can understand and validate one logical change at a time.

## Frontend Code

For new or modified JavaScript, TypeScript, JSX, TSX, and Astro code:

- Use arrow functions by default for components, callbacks, and local helpers.
- Use a `function` declaration only when hoisting, generator syntax, dynamic `this`/`arguments`, or a framework contract requires it.
- Within `web/src` and `docs/src`, use the application-local `@/` alias for imports that cross directories.
- Keep `./` imports for modules in the same directory and use package names such as `@netstamp/ui` for workspace dependencies.
- Do not introduce parent-relative `../` import chains. Existing parent-relative imports may be migrated when their files are otherwise being changed; avoid unrelated repository-wide rewrites in a focused pull request.

Example:

```tsx
import type { ApiProject } from "@/shared/api/types";
import { ProjectCard } from "./ProjectCard";

const ProjectSummary = ({ project }: { project: ApiProject }) => <ProjectCard project={project} />;
```

## Generated API Artifacts

TypeSpec is the API contract source of truth. After changing API models or operations, run:

```bash
pnpm generate:openapi
```

Commit the generated OpenAPI documents, the backend embedded copy, and the web API types with the source contract change. Do not edit generated artifacts by hand.

## Localization

English UI resources and English MDX are the source of truth. Traditional Chinese resources are synchronized through Crowdin. Read [`docs/localization.md`](./docs/localization.md) before changing user-visible text, locale routing, documentation content, or translation workflow.

Translation-only contributors can work directly in the [Netstamp Crowdin project](https://crowdin.com/project/netstamp) without repository write access or a Crowdin API token. The public [Translating Netstamp guide](./docs/src/content/docs/en/guides/translating.mdx) explains how to select strings, preserve code and placeholders, use Taiwanese terminology, handle QA warnings, and request context. Do not open a pull request that changes only Crowdin-managed output unless the same correction has already been made in Crowdin.

Run the localization checks for every user-visible text change:

```bash
pnpm check:i18n
pnpm test:i18n
pnpm test:web
```

Never commit Crowdin tokens or populated `.env` files. Do not manually change Crowdin-managed `zh-TW` output without applying the same correction in Crowdin, or the next translation download may overwrite it.

## Validation

Run checks that match the affected area. The canonical repository-level commands are:

```bash
pnpm check:branch-name
pnpm check:frontend-style
just lint
just test
just build
```

Useful focused commands include:

```bash
pnpm --filter @netstamp/web typecheck
pnpm --filter @netstamp/web lint
pnpm --filter @netstamp/web build
pnpm --filter @netstamp/ui typecheck
pnpm --filter @netstamp/ui build
pnpm --filter @netstamp/ui build:storybook
pnpm --filter @netstamp/api check
go test ./...
```

Run backend commands from `server/` when invoking `go` directly. For visible UI work, verify the affected workflow at desktop and mobile widths and include screenshots in the pull request.

## Pull Requests

Before opening a pull request:

1. Rebase or merge the latest `main` as appropriate for the active work.
2. Confirm the branch name and PR title follow their conventions.
3. Keep the PR focused on one coherent outcome.
4. Describe the changed area and concrete behavior.
5. List the validation commands and manual checks performed.
6. Link related issues with `Closes`, `Fixes`, or `Refs` when applicable.
7. Include desktop and mobile screenshots for visible UI changes. Upload review-only screenshots directly to the pull request description or comments; do not commit temporary preview images under `.github` or application source directories.
8. Update generated files, documentation, and area `AGENTS.md` files when the source-of-truth behavior changes.

Pull requests target `main`. The repository requires review, resolved review threads, and passing status checks before squash merge. Delete the working branch after the pull request is merged.
