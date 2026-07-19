# Docs Guidelines

## Scope

This guide applies to `docs/`, including Astro pages, MDX content rendering, public homepage work, docs navigation/search, OpenAPI theming, and Storybook static publishing.

Read the root `AGENTS.md`, `web/AGENTS.md`, and `design.md` before changing docs UI, styling, theme behavior, or browser interaction.

## Structure

- Keep English source MDX under `src/content/docs/en/` and Crowdin-synchronized Traditional Chinese MDX under `src/content/docs/zh-TW/` with matching relative paths.
- Keep localizable Astro shell and landing-page text in `src/i18n/locales/en/ui.json`; `src/i18n/locales/zh-TW/ui.json` is the Crowdin translation output.
- Keep docs-only Astro components under `src/components/docs`.
- Keep public homepage components under `src/components/landing`.
- Prefer small docs-local components over growing `DocLayout.astro`; navigation, pager, page actions, callouts, cards, and TOC behavior should not all live in one layout file.
- Use `@/` for imports that cross directories under `src`; reserve `./` for same-directory modules and do not introduce parent-relative `../` import chains.
- Use arrow functions by default for new or modified JavaScript, TypeScript, and Astro helpers and callbacks. Use `function` only when its hoisting, generator, or dynamic `this`/`arguments` semantics are required.
- Use `@netstamp/ui` for reusable primitives before creating docs-local controls. If Astro-only rendering prevents direct React usage, create a tokenized docs-local primitive with the same visual and accessibility contract.

## Styling And Theme

- Use `@netstamp/ui/styles` once through `BaseLayout.astro`.
- Use `--ns-*` tokens for colors, fonts, borders, state, focus, and transitions. Do not create `--home-*`, `--docs-*`, or raw color systems for production UI. Browser metadata such as `theme-color` may use the concrete token value that browsers require at build time.
- Use `rem` units on a `0.25rem` grid for layout, gaps, padding, margins, control dimensions, and breakpoints. Use `px` only for borders, outlines, hairlines, visually-hidden helpers, and unavoidable asset or canvas coordinates.
- Keep docs on the same dark-default root `data-theme` and `netstamp:theme` behavior as the web app. The public homepage is the exception: it is a single fixed dark product design and should not expose a light/dark split.
- Avoid decorative gradients, glows, or shadows in product and docs UI. Use borders, surface contrast, and layout hierarchy first.

## Accessibility

- Provide skip links for fixed or repeated navigation.
- Use `:focus-visible` or `:has(:focus-visible)` for keyboard focus. Pointer focus must not create unexpected keyboard focus rings.
- Search overlays, menus, dialogs, and popovers must use `@netstamp/ui` primitives or equivalent focus management, keyboard navigation, escape handling, and labelled semantics.
- Add `prefers-reduced-motion` fallbacks for animations and infinite motion.
- Provide stable `width` and `height` attributes for important images where the dimensions are known.

## Commands

- `pnpm --filter @netstamp/docs dev`: start the docs dev server.
- `pnpm --filter @netstamp/docs build`: build static Storybook into `docs/public/storybook`, then build Astro into `docs/dist`.
- `pnpm --filter @netstamp/docs preview`: preview the built docs output.
- `pnpm --filter @netstamp/ui build:storybook`: build shared UI Storybook.
- `pnpm check:frontend-style`: run token, focus-visible, and px-unit guardrails for docs, web, and shared UI implementation files.
- `pnpm check:i18n`: validate Docs UI JSON, document parity, code blocks, placeholders, locale directories, and React resources.
- `pnpm test:docs:i18n`: validate localized routes, metadata, links, language markup, and the OpenAPI shell after building Docs.
