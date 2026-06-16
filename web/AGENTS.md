# Frontend Guidelines

## Structure

- Route-level product areas live under `web/src/features/<feature>/components`.
- Keep feature-only UI and CSS modules colocated with that feature.
- Put reused app-level components in `web/src/shared/components`.
- Put reusable chart, map, timeline, and other domain-neutral visualization primitives in `web/src/shared/visualizations`.
- Put shared utilities, mock data, and data helpers in `web/src/shared/utils`.
- Use `@netstamp/ui` for reusable primitives before adding app-local controls.
- Use `@` aliases for imports to avoid relative paths and improve readability.
- Keep `web/src/shared/components` scoped to authenticated app layout, workflow, and provider helpers. Move cross-surface controls, form primitives, tables, panels, drawers, tabs, segmented controls, and action buttons to `@netstamp/ui`.

## Styling

- Prefer one CSS module per component or route section.
- Avoid shared catch-all page stylesheets; extract repeated patterns into shared components instead.
- Follow the existing dark, technical console visual language unless the task explicitly targets a new surface.

## Commands

- `pnpm --filter @netstamp/web typecheck`: run TypeScript checks.
- `pnpm --filter @netstamp/web lint`: run frontend ESLint.
- `pnpm --filter @netstamp/web build`: build the web app.
- `pnpm --filter @netstamp/web generate:api-types`: regenerate `src/shared/api/openapi.d.ts` from `docs/public/openapi.json`.
