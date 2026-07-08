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
- Project membership screens stay under `features/project` unless membership becomes an independently routed, cross-project domain with its own API and navigation boundary.
- The Insight assignment picker remains feature-local until another route needs the same probe/check multi-select behavior; promote it to `@netstamp/ui` only when the generic contract is clear.

## Styling

- Prefer one CSS module per component or route section.
- Avoid shared catch-all page stylesheets; extract repeated patterns into shared components instead.
- Follow `design.md` as the source of truth. The current target is a dark-default, spec-docs/product design system with restrained orange primary accents, precise square frames, minimal shadows, no decorative gradients, and matching light mode where the surface supports theme switching.
- Do not introduce raw color literals in implementation CSS, TSX, or Astro. Use `var(--ns-*)` tokens from `packages/ui/src/styles/tokens.css`; raw colors belong only in token definitions, vendor bridge overrides, browser/asset metadata, canvas/WebGL scene data that cannot consume CSS variables directly, and tests.
- Use `rem` units on a `0.25rem` spacing and sizing grid for layout, gaps, padding, margins, control dimensions, and breakpoints. Use `px` only for borders, dividers, outlines, hairlines, visually-hidden helpers, and unavoidable asset or canvas coordinates.
- Custom keyboard focus styling must use `:focus-visible` or `:has(:focus-visible)`. Pointer and mouse focus must not show unexpected keyboard focus rings or border changes, and `outline: none` or `outline: 0` needs an equivalent visible keyboard focus state.
- Dialogs, menus, popovers, search overlays, and drawers should use `@netstamp/ui` Radix-based primitives or implement equivalent focus management, keyboard navigation, escape handling, and labelled semantics.

## Commands

- `pnpm --filter @netstamp/web typecheck`: run TypeScript checks.
- `pnpm --filter @netstamp/web lint`: run frontend ESLint.
- `pnpm --filter @netstamp/web build`: build the web app.
- `pnpm check:frontend-style`: run token, focus-visible, and px-unit guardrails for web, docs, and shared UI implementation files.
- `pnpm --filter @netstamp/web generate:api-types`: regenerate `src/shared/api/openapi.d.ts` from `docs/public/openapi.json`.
