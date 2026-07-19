# Localization

Netstamp supports English (`en`) and Traditional Chinese for Taiwan (`zh-TW`). English is the source language, English is the fallback language, and localized URLs keep existing English routes unprefixed.

> English files are the source of truth. `zh-TW` files are synchronized from Crowdin.

## Contributing A Translation

Translation contributors can work directly in the [Netstamp Crowdin project](https://crowdin.com/project/netstamp). You do not need repository write access, a local development environment, `CROWDIN_PROJECT_ID`, or `CROWDIN_PERSONAL_TOKEN` to translate strings in the Crowdin editor.

Read the public [Translating Netstamp guide](./src/content/docs/en/guides/translating.mdx) for the contributor workflow, Taiwanese terminology, protected technical content, placeholder rules, QA review, and the maintainer synchronization process. Use Crowdin comments when the English source is ambiguous instead of guessing from an isolated string.

## Locale Architecture

Shared locale metadata and routing helpers live in `packages/i18n`. That package defines the supported locale codes, default and fallback locale, HTML language values, Crowdin mapping, locale validation, language-switcher metadata, and path conversion helpers.

| Surface                      | English source                      | Traditional Chinese output             |
| ---------------------------- | ----------------------------------- | -------------------------------------- |
| React app                    | `web/src/i18n/locales/en/*.json`    | `web/src/i18n/locales/zh-TW/*.json`    |
| Astro shell and landing page | `docs/src/i18n/locales/en/ui.json`  | `docs/src/i18n/locales/zh-TW/ui.json`  |
| Documentation content        | `docs/src/content/docs/en/**/*.mdx` | `docs/src/content/docs/zh-TW/**/*.mdx` |

The React app uses `i18next` and `react-i18next`. Resources are bundled with the app and initialized before the first React render, so loading a page does not create a translation network waterfall or briefly display the wrong language. The selected language is stored under `netstamp:locale`; an explicit selection wins over browser preferences. `zh-TW`, `zh-Hant`, and `zh-HK` browser locales resolve to `zh-TW`; unsupported locales resolve to English.

The Astro site uses unprefixed English routes and a `zh-TW` prefix:

```text
/docs/guides/getting-started/        English
/zh-TW/docs/guides/getting-started/  Traditional Chinese
```

Every localized page emits the correct `<html lang>`, canonical URL, `hreflang="en"`, `hreflang="zh-TW"`, and `hreflang="x-default"`. The language switcher preserves the current page path. Documentation routes are generated from the English page inventory; if a Traditional Chinese MDX file is missing, the route renders the English source with a visible fallback notice and a link to the English URL.

The localized OpenAPI route translates Netstamp's page shell, metadata, loading state, and schema section label. Scalar's embedded API-reference controls and the generated operation text remain English because the installed Scalar configuration does not expose a locale option and the OpenAPI contract is the English source of truth. Do not send API identifiers or generated OpenAPI artifacts through Crowdin; localizing the embedded client requires upstream Scalar locale support or a separately maintained integration.

## Adding Or Changing App Text

1. Choose the namespace that owns the feature. Do not put unrelated strings into `common.json`.
2. Add a semantic, stable key to the English namespace. Use `navigation.docs`, not an English sentence as the key.
3. Use `useTranslation("namespace")` in React components. Use `<Trans>` for structured component interpolation instead of embedding HTML in a translation value.
4. Preserve interpolation variables in every translation. For example, `Welcome, {{name}}` must keep `{{name}}`.
5. Run `pnpm check:i18n`, the focused React tests, typecheck, and the web build.

New English keys do not have to block a pull request while Crowdin translation is pending: the app falls back to English. The checked-in `zh-TW` files must still pass structural validation when they are updated or downloaded.

Do not translate product names, API names, code, commands, file paths, URLs, placeholders inside code samples, or protocol identifiers such as `Netstamp`, `OpenAPI`, `HTTP`, `TLS`, and `CROWDIN_PROJECT_ID`.

## Adding Documentation

1. Add the English MDX file below `docs/src/content/docs/en/`.
2. Assign its sidebar section with `navSection` and its position within that section with `navOrder`. These structural values are owned by the English source and must match in every locale.
3. Include localizable frontmatter fields such as `title`, `description`, and page-title metadata; keep structural frontmatter, imports, component tags, anchors, code blocks, and file paths intact.
4. Upload sources to Crowdin and download the completed translation. Crowdin writes the matching path below `docs/src/content/docs/zh-TW/`.
5. Check relative links, localized internal links, headings, images, code examples, sidebar order, search results, the page switcher, and both direct URLs.

The supported sidebar sections, in display order, are `start`, `install`, `use`, `operate`, `api`, `development`, and `community`. Section labels are translated in `docs/src/i18n/locales/*/ui.json`; localized MDX cannot move itself to a different section or change its order. Keep Getting Started at `start` / `0`, and keep translation and contribution material under `development` rather than mixing it into product reference material.

Astro shell, search, navigation, pagination, tracking consent, page actions, landing-page copy, accessibility labels, and metadata come from `docs/src/i18n/locales/*/ui.json` and follow the same English-source workflow.

## Local Testing

Install dependencies and run both app surfaces:

```bash
pnpm install
pnpm dev:web
pnpm dev:docs
```

Use the language switcher without reloading. Verify that the app retains its current route and state, the Docs switcher opens the equivalent localized URL, and `document.documentElement.lang` changes to `en` or `zh-Hant-TW`.

Run automated checks:

```bash
pnpm check:i18n
pnpm test:i18n
pnpm test:web
pnpm build:web
pnpm build:docs
pnpm test:docs:i18n
```

`scripts/check-i18n.mjs` reports the surface, namespace or MDX path, key path, issue type, source value, and translation value. It detects invalid JSON, duplicate keys, missing or extra keys, empty translations, object/string type mismatches, interpolation mismatches, invalid locale directory names, missing namespaces or documents, and changed MDX code blocks.

## Crowdin Setup

The root `crowdin.yml` uploads only English sources and maps Crowdin's Traditional Chinese locale to the exact `zh-TW` directory name. It preserves the documentation hierarchy.

MDX sources disable Crowdin content segmentation. Crowdin otherwise splits long English and Chinese paragraphs differently, which prevents a reviewed translated MDX file from mapping back to every source segment. Paragraph-level segmentation keeps checked-in human translations aligned with their English document structure; changed source text is retained as unapproved so it can be reviewed again.

Create a personal local `.env` or export these variables in the shell or CI secret store:

```text
CROWDIN_PROJECT_ID=
CROWDIN_PERSONAL_TOKEN=
```

Never commit a Crowdin token, project secret, or populated `.env` file. `.env.example` documents the variable names and `.gitignore` excludes local credentials.

Upload English sources:

```bash
pnpm crowdin:upload
```

Download Traditional Chinese translations:

```bash
pnpm crowdin:download
```

Crowdin manages these translation outputs:

- `web/src/i18n/locales/zh-TW/*.json`
- `docs/src/i18n/locales/zh-TW/ui.json`
- `docs/src/content/docs/zh-TW/**/*.mdx`

Avoid editing those files directly because the next download can overwrite the change. Make translation corrections in Crowdin, then download again. If an urgent local correction is unavoidable, apply the same correction in Crowdin before the next synchronization.

When a maintainer reviews or improves a checked-in `zh-TW` file locally, import that exact reviewed file into Crowdin before running the next download. Confirm the Crowdin import finishes successfully, allow its QA checks and glossary terminology checks to run, and only then download translations again. This keeps the human-reviewed Taiwanese wording from being replaced by an older machine or translation-memory suggestion.

The Netstamp Crowdin project glossary defines preferred product terminology. Use it for recurring terms such as probe, check, incident, token, availability, and status-page states. A glossary suggestion is context, not permission to translate identifiers, commands, file paths, code, or brand names.

The repository does not automatically upload or download translations in pull requests. This avoids competing with a Crowdin VCS integration. CI validates checked-in resources, tests fallback behavior, and builds both locales.

## Adding Another Language

1. Add the locale and metadata to `packages/i18n/src/locales.ts` and update routing tests.
2. Add its Crowdin locale mapping in `crowdin.yml`.
3. Add translated resource directories for the React app, Astro UI, and MDX content.
4. Add the locale to React resources and Astro route generation.
5. Extend language detection only when related browser locale aliases are safe and intentional.
6. Update validation, route-output tests, SEO alternates, documentation, and Crowdin project languages.

Do not add a second independent locale list to a feature package. Consumers must use the shared metadata from `@netstamp/i18n`.

## Troubleshooting

### A key is displayed instead of text

Confirm that the component requests the correct namespace, the English key exists, and its key path matches the JSON structure. Run `pnpm check:i18n` and the web typecheck.

### Traditional Chinese displays English

English fallback is expected when a translation is missing. Confirm that the `zh-TW` file contains the key, Crowdin downloaded into `zh-TW` rather than `zh` or `zh_Hant`, and the resource is imported by `web/src/i18n/resources.ts` or `docs/src/i18n/ui.ts`.

### Interpolation is missing

Keep every English interpolation name exactly, including case. `{{name}}` and `{{userName}}` are different. The localization check reports mismatched variables.

### A Docs link changes language unexpectedly

Use the locale path helpers instead of manually concatenating prefixes. MDX links in Traditional Chinese should point to the equivalent `/zh-TW/...` route unless the link intentionally opens the English fallback.

### Crowdin writes the wrong directory

Verify the `languages_mapping.locale.zh-TW: zh-TW` entries in `crowdin.yml`, and confirm the Crowdin project target language is Traditional Chinese. Do not add `zh-TW` files as sources.

### Crowdin CLI authentication fails

Confirm both environment variables are available to the process, the token can access the configured project, and no populated credential file is being committed. With valid credentials, use `pnpm exec crowdin upload sources --dryrun --tree` to validate file matching without uploading sources.
