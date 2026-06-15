# Netstamp Docs

This package owns the public Netstamp site and documentation. It uses Astro for static pages, MDX files under `src/content/docs/` for documentation content, and React islands only where interaction is required.

## Commands

Run from the repository root:

- `pnpm --filter @netstamp/docs dev`: start the Astro docs dev server.
- `pnpm --filter @netstamp/docs build`: build static Storybook into `docs/public/storybook`, then build the Astro site into `docs/dist`.
- `pnpm generate:openapi`: regenerate the TypeSpec OpenAPI contract used by the docs explorer and backend runtime docs.
- `pnpm --filter @netstamp/docs preview`: preview the built docs output.
- `pnpm --filter @netstamp/ui storybook`: run Storybook locally for shared UI components.

## Public App Links

The public site links to the dashboard app through `PUBLIC_NETSTAMP_APP_BASE_URL`, which defaults to `https://dashboard.netstamp.dev`.

## Structure

- `src/pages/index.astro`: public landing page shell.
- `src/pages/docs/[...slug].astro`: renders docs content through `DocLayout.astro`.
- `src/content/docs/**/*.mdx`: MDX documentation content with `title`, `description`, `icon`, and optional `head` frontmatter.
- `src/components/landing/`: React landing page island and visual scenes.
- `src/components/openapi/`: React OpenAPI explorer used by the public Markdown OpenAPI page.
- `public/openapi.json`: generated backend OpenAPI contract.
- `public/storybook/`: ignored static Storybook build that Astro copies into `docs/dist/storybook`.
