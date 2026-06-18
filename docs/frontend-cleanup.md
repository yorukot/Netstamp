# Frontend Cleanup Checklist

This checklist tracks the cleanup pass for `web/` and `docs/`. Complete items in small commits.

## Error Handling

- [x] Remove duplicate mutation error toasts by making global React Query error reporting opt-in or locally suppressible.
- [x] Keep page-level success/error messages only where the page adds useful context.

## Domain Boundaries

- [x] Move cross-feature view/domain models shared by probes, checks, dashboard, and insight out of feature component folders.
- [x] Remove feature API adapters that import types from another feature's component directory.
- [ ] Keep feature-only UI state colocated with its route feature.

## API Layer

- [x] Split the monolithic shared query registry by domain.
- [x] Split the monolithic shared mutation registry by domain.
- [x] Keep query keys and low-level OpenAPI client helpers shared.
- [x] Preserve existing cache invalidation behavior while splitting.

## Routing

- [ ] Split route guards and legacy redirects out of `AppRouter`.
- [ ] Keep the top-level router file focused on providers and route table wiring.
- [ ] Rename ambiguous route keys so project settings and account settings are distinct.

## Shared UI Boundary

- [ ] Promote reusable primitives currently in `web/src/shared/components` to `@netstamp/ui` where they are not app-workflow specific.
- [ ] Leave authenticated app workflow/provider helpers in `web/src/shared/components`.
- [ ] Add package exports and Storybook coverage for promoted primitives where appropriate.

## Styling

- [ ] Split `AppShell.module.css` so Sidebar, ProjectSwitcher, and UserMenu styles live with their components.
- [ ] Keep one CSS module per component or route section.

## Large Pages

- [ ] Split Alerts page table/form/drawer logic into smaller feature modules.
- [ ] Split Checks page editor and selector logic into smaller feature modules or hooks.
- [ ] Split Insight URL state and scope derivation into smaller feature modules or hooks.

## Docs

- [ ] Extract docs layout inline browser behavior into reusable scripts.
- [ ] Extract docs search inline browser behavior into reusable scripts.
- [ ] Keep Astro components focused on markup and data injection.

## Housekeeping

- [ ] Remove duplicated frontend assets shared between web and docs.
- [ ] Remove empty frontend directories.
- [ ] Keep generated and ignored build outputs out of commits.

## Validation

- [ ] Run `pnpm --filter @netstamp/web typecheck`.
- [ ] Run `pnpm --filter @netstamp/web lint`.
- [ ] Run `pnpm --filter @netstamp/docs build`.
