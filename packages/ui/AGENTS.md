# Shared UI Guidelines

## Scope

This guide applies to `packages/ui`, the shared primitive layer for the React web app, Astro docs islands, Storybook, and cross-surface visual patterns.

Read the root `AGENTS.md`, `web/AGENTS.md`, and `design.md` before changing component APIs, token roles, shared CSS, or Storybook coverage.

## Component Ownership

- Reusable controls, form primitives, tables, panels, surfaces, drawers, tabs, segmented controls, action rows, badges, metrics, empty states, and cross- surface layout primitives belong here.
- Feature-only workflows stay in `web/src/features/<feature>/components` until at least two surfaces need the same pattern.
- Export every public primitive from `src/index.ts`.
- Add or update Storybook stories for public primitives, including light/dark examples and key states such as disabled, invalid, loading, selected, empty, keyboard focus, and dense usage.

## Styling

- `src/styles/tokens.css` is the only shared token source. Component CSS should consume `--ns-*` tokens and should not define local raw color systems.
- Do not redefine the root `--ns-*` theme set inside a component. Add a semantic token to `tokens.css` instead when a reusable role is missing.
- Use `rem` units on a `0.25rem` grid for spacing, sizing, control dimensions, and breakpoints. Use `px` only for borders, dividers, outlines, hairlines, visually-hidden helpers, and unavoidable asset or canvas coordinates.
- Core frames are square by default. Shadows, glows, gradients, and decorative glass effects are not part of the shared primitive system.

## Accessibility

- Icon-only controls must require an accessible label at the type level when practical.
- Custom keyboard focus styling must use `:focus-visible` or `:has(:focus-visible)`. Do not use broad `:focus` or `:focus-within` rings that show keyboard focus styling on normal mouse interaction.
- Do not use `outline: none` or `outline: 0` without an equivalent visible keyboard focus style.
- Dialog, popover, select, combobox, menu, tab, and table interactions should use Radix or follow the corresponding WAI-ARIA keyboard pattern.
- Respect `prefers-reduced-motion` for any animation or repeated motion.

## Commands

- `pnpm --filter @netstamp/ui typecheck`: type-check the package.
- `pnpm --filter @netstamp/ui build`: build package output.
- `pnpm --filter @netstamp/ui build:storybook`: build Storybook.
