# Netstamp Docs

This package owns the public Netstamp site and documentation. It uses Astro for static pages, MDX files under `src/content/docs/` for documentation content, and React islands only where interaction is required.

## Commands

Run from the repository root:

- `pnpm --filter @netstamp/docs dev`: start the Astro docs dev server.
- `pnpm --filter @netstamp/docs build`: build static Storybook into `docs/public/storybook`, then build the Astro site into `docs/dist`.
- `pnpm generate:openapi`: regenerate the TypeSpec OpenAPI contract used by the docs explorer and backend runtime docs.
- `pnpm --filter @netstamp/docs preview`: preview the built docs output.
- `pnpm --filter @netstamp/ui storybook`: run Storybook locally for shared UI components.

## Structure

- `src/pages/index.astro`: public landing page shell.
- `src/pages/docs/[...slug].astro`: renders default-locale docs content through `DocLayout.astro`.
- `src/pages/[locale]/docs/[...slug].astro`: renders translated docs (or English fallback) for prefixed locales.
- `src/content/docs/**/*.mdx`: MDX documentation content with `title`, `description`, `icon`, and optional `head` frontmatter.
- `src/i18n/config.ts`: locale registry and UI string tables (see below).
- `src/components/landing/`: React landing page island and visual scenes.
- `src/components/openapi/`: React OpenAPI explorer used by the public Markdown OpenAPI page.
- `public/openapi.json`: generated backend OpenAPI contract.
- `public/storybook/`: ignored static Storybook build that Astro copies into `docs/dist/storybook`.

## Internationalization (i18n)

Docs are multi-locale and table-driven from `src/i18n/config.ts`.

- The default locale (`en`) lives at the content root and keeps its `/docs/...` URLs.
- Every other locale mirrors the same relative paths inside `src/content/docs/<locale>/` and is served under `/<locale>/docs/...`.
- A page with no translation falls back to the English content automatically and shows a "not translated yet" notice; the language switcher and navigation always stay within the active locale.

To translate a page, copy the English file into the matching `src/content/docs/<locale>/` path and translate its frontmatter and body. To add a language, append its code to `locales` in `src/i18n/config.ts` and fill in the matching name, `<html lang>`, Open Graph locale, direction, and UI strings.
