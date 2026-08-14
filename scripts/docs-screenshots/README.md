# Documentation screenshots

This directory regenerates the product screenshots used by the homepage and both documentation locales.

The capture opens the current Netstamp frontend, intercepts every `/api/v1/` request before the first navigation, and supplies anonymous deterministic fixtures. It never needs an account or password. Unhandled API writes are blocked and fail the run.

## Requirements

- Google Chrome
- `cwebp`
- Workspace dependencies installed with `pnpm install`

## Capture

```bash
pnpm capture:docs-screenshots
```

The default frontend is `https://app.netstamp.dev`. To capture another build without changing the fixtures:

```bash
DOCS_SCREENSHOT_BASE_URL=http://localhost:5173 pnpm capture:docs-screenshots
```

Generated documentation images are written to `docs/src/assets/screenshots/`. The eight homepage images are written to `docs/src/assets/homepage-highlights/` at the `1200x780` size required by the device frame.

Use `--audit` to inspect one route without replacing the image set:

```bash
DOCS_SCREENSHOT_AUDIT_ROUTE=/projects/docs-demo/checks pnpm capture:docs-screenshots -- --audit
```

Temporary browser artifacts are written under `output/playwright/docs-screenshots/`.
