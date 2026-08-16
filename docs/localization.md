# Localization

Netstamp currently publishes English (`en`) only. English is the source, default, and fallback language. Additional language releases and translation downloads are paused, but the locale architecture remains available for future use.

## Locale Architecture

Shared locale metadata and routing helpers live in `packages/i18n`. The package owns supported locale codes, default and fallback behavior, HTML language metadata, locale validation, language-switcher metadata, and localized path generation.

| Surface                      | Current source                      |
| ---------------------------- | ----------------------------------- |
| React app                    | `web/src/i18n/locales/en/*.json`    |
| Astro shell and landing page | `docs/src/i18n/locales/en/ui.json`  |
| Documentation content        | `docs/src/content/docs/en/**/*.mdx` |

The React app continues to use `i18next` and `react-i18next`. Resources are bundled before the first render, locale-aware date and number formatters remain active, and unsupported stored or browser locales resolve to English.

The Astro site keeps its locale-aware layouts, content lookup, route helpers, metadata generation, and language-menu component. With one supported locale, English routes remain unprefixed and language switchers render nothing. Unsupported locale prefixes are not generated or redirected.

## Adding Or Changing App Text

1. Choose the namespace that owns the feature; do not put unrelated strings in `common.json`.
2. Add a semantic, stable key to the English namespace.
3. Use `useTranslation("namespace")` in React components and `<Trans>` for structured interpolation.
4. Preserve interpolation variables such as `{{name}}`.
5. Run `pnpm check:i18n`, focused React tests, typecheck, and the Web build.

Do not localize product names, API identifiers, code, commands, file paths, URLs, or protocol names such as `Netstamp`, `OpenAPI`, `HTTP`, and `TLS`.

## Adding Documentation

1. Add English MDX below `docs/src/content/docs/en/`.
2. Set `navSection` and `navOrder`; the English inventory owns route and sidebar structure.
3. Keep imports, component tags, anchors, code blocks, and file paths stable so future translations can preserve document topology.
4. Run the localization check, Docs build, and Docs architecture test.

The supported sidebar sections, in order, are `start`, `install`, `use`, and `operate`. Astro shell, search, navigation, pagination, tracking consent, page actions, landing-page copy, accessibility labels, and metadata remain sourced from locale JSON.

## Validation

Run:

```bash
pnpm check:i18n
pnpm test:i18n
pnpm test:web
pnpm build:web
pnpm build:docs
pnpm test:docs:i18n
```

`scripts/check-i18n.mjs` validates active locale directories, English JSON syntax and duplicate keys, source document presence, and translation parity when additional locales are enabled. `scripts/check-docs-i18n.mjs` verifies the currently published English routes, metadata, navigation, search output, and absence of disabled locale output.

## Crowdin

The root `crowdin.yml` retains generic mappings from English Web, Docs UI, and MDX sources to future translation paths. Translation downloads are intentionally disabled while localized releases are paused.

Create a personal `.env` or export these variables only when maintaining Crowdin sources:

```text
CROWDIN_PROJECT_ID=
CROWDIN_PERSONAL_TOKEN=
```

Upload English sources with:

```bash
pnpm crowdin:upload
```

Never commit Crowdin credentials or a populated `.env` file. Do not add a download script until at least one additional locale has been reviewed and re-enabled.

## Re-enabling A Language

1. Add the locale and metadata to `packages/i18n/src/locales.ts`, including browser aliases only when intentional.
2. Add its Crowdin locale mapping and reviewed translation directories for Web JSON, Docs UI JSON, and MDX content.
3. Add locale resources to the React bundle and locale route generation to Astro.
4. Restore a scoped Crowdin download command for the approved locale.
5. Extend locale, routing, fallback, language-switcher, SEO, Sitemap, and rendered-route tests.
6. Verify direct localized URLs, current-path switching, canonical URLs, `hreflang`, Open Graph metadata, search, Markdown output, and fallback behavior.

Do not add independent locale lists inside feature packages. Consumers must use shared metadata from `@netstamp/i18n`.
