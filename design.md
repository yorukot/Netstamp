# Netstamp Design System

> Category: network observability and developer infrastructure. Direction: spec-documentation product UI for the authenticated web app, public homepage, docs, Storybook, and API reference. The system is light-default, high-contrast, restrained, orange-led, and built from explicit tokens, dense information surfaces, and practical network telemetry.

This file is the source of truth for Netstamp visual design. Treat it as an implementation spec, not mood guidance. Every shared primitive, app route, docs page, Storybook story, and API explorer surface should be traceable back to the foundations below.

Primary implementation files:

- `packages/ui/src/styles/tokens.css`: global fonts, color tokens, state tokens, focus, frame utilities, reset, scrollbars, and compatibility utilities.
- `packages/ui/src/components`: reusable primitives for both the React app and Astro docs.
- `packages/ui/src/stories`: Storybook documentation for every shared primitive and system pattern.
- `web/src/index.css`: app root import of shared styles.
- `web/src/layouts`: authenticated product shell.
- `web/src/shared/components`: app-only layout helpers and domain-neutral product patterns.
- `docs/src/layouts/DocLayout.astro`: documentation shell, prose, doc hero, callouts, cards, pager, and TOC.
- `docs/src/components/docs`: top navigation, search, and docs-local controls.
- `docs/src/components/landing`: public homepage.
- `docs/src/components/openapi`: API reference and request console.

## 1. Design Thesis

Netstamp should look like an engineering design system and operations console that happens to ship as a polished product. It must feel serious, modern, and expensive without relying on decorative gradients, heavy radius, glossy cards, or empty marketing ornament.

The target visual language is "spec-docs":

- Every page has a clear information hierarchy.
- Tokens and component roles are obvious.
- Frames are precise and mostly square.
- Panels look engineered, not decorative.
- Examples use real product telemetry, API paths, checks, routes, probes, labels, alerts, and status pages.
- Copy sounds like a reference manual or controller UI.
- Color is intentional: orange for primary interaction, blue for secondary reference/data, semantic colors only for state.
- Density is high enough for repeated operational work, but spacing and contrast keep the UI comfortable.

Core traits:

- Light default, with matching dark mode.
- Warm orange primary color, tuned for Netstamp rather than inherited unchanged from another brand.
- Neutral ink, graphite, white, and restrained cool gray foundations. Public chrome uses the Sui-style black-green tone `#131518`.
- Square or nearly-square frames. Core controls use zero radius.
- Minimal shadows. If depth is needed, use borders, surface contrast, and spatial grouping first.
- Gradients are not part of the product UI. Use flat token colors.
- Richness comes from layout, data, typography, state, and component completeness, not decoration.

## 2. Token Contract

Use `--ns-*` tokens from `packages/ui/src/styles/tokens.css`. Do not create new local color systems in docs, Storybook, or web routes.

### Typography Tokens

- `--ns-font-sans`: `"TASAOrbiter", "TASA Orbiter", sans-serif`.
- `--ns-font-display`: `"TASAExplorer", "TASA Explorer", sans-serif`.
- `--ns-font-mono`: `"JetBrainsMono", "JetBrains Mono", monospace`.

Usage:

- Display: page titles, landing hero, major docs headings, large metric values.
- Sans: body copy, form labels, app navigation, table text, buttons, compact panel headings.
- Mono: IDs, tokens, timestamps, paths, command text, code, method labels, response previews, technical metadata.

Do not use display-size type inside compact cards, sidebars, tables, nav, or controls.

### Light Theme Color Targets

These values follow the corrected Sui-reference direction and brand orange assets. `tokens.css` must match these values.

- `--ns-bg`: `#f4f5f7`, page canvas and public body background.
- `--ns-bg-section`: `#ffffff`, broad docs and landing bands.
- `--ns-bg-subtle`: `#edeff3`, nested quiet surface.
- `--ns-surface`: `#ffffff`, primary panels and cards.
- `--ns-surface-raised`: `#ffffff`, toolbars, controls, raised rows.
- `--ns-surface-deep`: `#e8ebef`, code, topology, maps, consoles, response previews.
- `--ns-glass-dark`: `rgba(19, 21, 24, 0.96)`, dark floating chrome.
- `--ns-glass-light`: `rgba(19, 21, 24, 0.04)`, faint overlay.
- `--ns-overlay-scrim`: `rgba(19, 21, 24, 0.18)`, modal/search scrim.
- `--ns-chrome-bg`: `#131518`, public floating nav and dark chrome.
- `--ns-chrome-text`: `#ffffff`, text on public floating nav.
- `--ns-chrome-text-muted`: `rgba(255, 255, 255, 0.72)`, secondary chrome text.
- `--ns-chrome-border`: `rgba(255, 255, 255, 0.14)`, dark chrome frame.
- `--ns-chrome-hover`: `rgba(255, 255, 255, 0.1)`, dark chrome hover fill.
- `--ns-text`: `#172033`, primary text.
- `--ns-text-muted`: `#405168`, body and secondary text.
- `--ns-text-subtle`: `#657389`, supporting text.
- `--ns-text-low`: `#8793a3`, metadata and inactive controls.
- `--ns-text-on-accent`: `#ffffff`, text on filled primary controls.
- `--ns-primary`: `#ea6a1a`, Netstamp orange primary.
- `--ns-primary-hover`: `#d85a0f`, primary hover.
- `--ns-primary-active`: `#b94a0b`, primary pressed and focus.
- `--ns-primary-muted`: `rgba(234, 106, 26, 0.1)`, faint selected/hover fill.
- `--ns-primary-subtle`: `rgba(234, 106, 26, 0.15)`, stronger selected fill.
- `--ns-primary-border`: `rgba(234, 106, 26, 0.38)`, active frame.
- `--ns-secondary`: `#2563eb`, blue secondary action/data.
- `--ns-secondary-hover`: `#1d4ed8`, blue hover.
- `--ns-secondary-active`: `#1e40af`, blue pressed.
- `--ns-secondary-muted`: `rgba(37, 99, 235, 0.08)`, faint blue fill.
- `--ns-secondary-subtle`: `rgba(37, 99, 235, 0.12)`, stronger blue fill.
- `--ns-secondary-border`: `rgba(37, 99, 235, 0.28)`, blue frame.
- `--ns-critical`: `#c9362c`, destructive/error/failed.
- `--ns-warning`: `#b7791f`, warning/degraded/pending.
- `--ns-success`: `#168a45`, healthy/online/success.
- `--ns-metal`: `#64748b`, neutral technical accent.
- `--ns-slate-line`: `rgba(100, 116, 139, 0.24)`, chart and divider line.
- `--ns-border`: `#d7dbe2`, default frame.
- `--ns-border-strong`: `#9fa7b5`, strong frame.
- `--ns-border-faint`: `#e7e9ee`, subtle divider.

### Dark Theme Color Targets

Dark mode is the same product system on a black canvas. Do not use blue-black as the default background; blue is reserved for secondary actions, links, and data.

- `--ns-bg`: `#000000`.
- `--ns-bg-section`: `#000000`.
- `--ns-bg-subtle`: `#131518`.
- `--ns-surface`: `#070707`.
- `--ns-surface-raised`: `#131518`.
- `--ns-surface-deep`: `#000000`.
- `--ns-glass-dark`: `rgba(19, 21, 24, 0.96)`.
- `--ns-glass-light`: `rgba(255, 255, 255, 0.045)`.
- `--ns-overlay-scrim`: `rgba(0, 0, 0, 0.56)`.
- `--ns-chrome-bg`: `#131518`.
- `--ns-chrome-text`: `#ffffff`.
- `--ns-chrome-text-muted`: `rgba(255, 255, 255, 0.72)`.
- `--ns-chrome-border`: `rgba(255, 255, 255, 0.14)`.
- `--ns-chrome-hover`: `rgba(255, 255, 255, 0.1)`.
- `--ns-text`: `#f8fafc`.
- `--ns-text-muted`: `#d8dce8`.
- `--ns-text-subtle`: `#98a8b4`.
- `--ns-text-low`: `#77736b`.
- `--ns-text-on-accent`: `#090b10`.
- `--ns-primary`: `#fb923c`.
- `--ns-primary-hover`: `#fdba74`.
- `--ns-primary-active`: `#f97316`.
- `--ns-primary-muted`: `rgba(251, 146, 60, 0.13)`.
- `--ns-primary-subtle`: `rgba(251, 146, 60, 0.19)`.
- `--ns-primary-border`: `rgba(251, 146, 60, 0.44)`.
- `--ns-secondary`: `#38bdf8`.
- `--ns-secondary-hover`: `#7dd3fc`.
- `--ns-secondary-active`: `#2563eb`.
- `--ns-secondary-muted`: `rgba(56, 189, 248, 0.11)`.
- `--ns-secondary-subtle`: `rgba(56, 189, 248, 0.16)`.
- `--ns-secondary-border`: `rgba(56, 189, 248, 0.34)`.
- `--ns-critical`: `#ff6b63`.
- `--ns-warning`: `#facc15`.
- `--ns-success`: `#34c77b`.
- `--ns-metal`: `#c4ccd9`.
- `--ns-slate-line`: `rgba(196, 204, 217, 0.22)`.
- `--ns-border`: `#2d3035`.
- `--ns-border-strong`: `#4a4f58`.
- `--ns-border-faint`: `#1b1d21`.

### Reference Palette Evidence

Reference palette extraction is based on the local image and SVG assets currently in this repository.

- `docs/src/assets/homepage-dashboard-light.png` dominant exact sample colors: `#f3f7fb`, `#ffffff`, `#e0e7e9`, `#f8fbff`, `#fdf0e8`, `#d8e1ec`, `#e4ebf3`, `#ea6a1a`, `#172033`, `#657389`, `#0b0f16`.
- `docs/src/assets/homepage-dashboard-dark.png` dominant exact sample colors: `#0d1624`, `#151f2e`, `#0b0b0b`, `#2a2a2a`, `#1b283a`, `#2b3a4e`, `#223045`, `#fb923c`, `#f8fafc`, `#1b1b1b`.
- Brand SVG literals: `#090b10`, `#0b0f16`, `#11151d`, `#ea6a1a`, `#fb923c`, `#fff7ed`, `#f8fafc`.
- Network map SVG literals: `#2563eb`, `#38bdf8`, `#77736b`, `#c4ccd9`.

Implementation decision: public body uses `#f4f5f7` in light mode and `#000000` in dark mode. Public floating navigation and dark chrome use `#131518` with white text. The blue-black colors present in the old screenshot (`#0d1624`, `#151f2e`, `#1b283a`) are not default background tokens.

### Compatibility Aliases

- `--ns-accent` and related `--ns-accent-*` tokens are aliases for primary orange. New code should prefer `--ns-primary-*` when the role is primary.
- `--ns-accent-glow` remains for compatibility only. Do not use it to create visible glow effects.
- `.ns-grid-shell` and `.ns-grid-shell--constellation` are compatibility classes. They must render as plain token backgrounds.
- `.ns-cut-frame` is a compatibility class. It must render as a standard square frame, with no clipped corners.

### Radius, Shadow, And Motion

- `--ns-radius-*`: `0` for app and docs core UI.
- `--ns-cut-*`: `0`, legacy only.
- `--ns-shadow-sm`, `--ns-shadow-md`, `--ns-shadow-glow`: `none`.
- `--ns-transition`: `180ms cubic-bezier(0.2, 0.8, 0.2, 1)`.
- Focus uses `--ns-focus-outline`, usually `2px solid var(--ns-primary-active)`.

If a surface needs hierarchy, use border strength, surface tone, section spacing, and structured headers before shadows or radius.

## 3. Surface Families

### Product App

The authenticated app is the densest surface.

- Desktop shell: sticky left sidebar plus fluid content.
- Sidebar: brand, project switcher, navigation, user menu.
- Main content: compact page stack, screen header, panels, data tables, metric rows, maps, timelines, charts, drawers.
- Route titles are large but not theatrical.
- Repeated data should use tables, key-value grids, timeline rows, terminal blocks, or compact panels.
- Wide technical data can scroll horizontally.

### Public Homepage

The homepage is product-led, not generic marketing.

- First viewport must immediately signal Netstamp and the product UI.
- Hero uses real dashboard/product screenshots or constructed product panels.
- CTAs are concrete: deploy, read docs, view GitHub, inspect API.
- Sections cover probes, checks, insight, alerts, status pages, API automation, and open source.
- Landing can have more breathing room than the app, but should still use framed product panels and spec-like content.

Approved copy direction:

- "See the network before it fails you."
- "Open-source network observability from probes you control."
- "Measure latency, packet loss, DNS, and routes."
- "Controller, probes, checks, alerts, and status pages in one visible loop."

Avoid generic SaaS copy such as "unlock potential", "beautifully simple", "supercharge", or "AI-powered for everyone".

### Documentation

Docs should feel like a design-system/spec site for Netstamp.

- Top nav: brand, docs, API, Storybook, GitHub, deploy/app actions, color mode.
- Desktop docs shell: left navigation, central prose, right TOC.
- Docs hero: compact framed panel with title, summary, metadata, and optional links.
- Prose: readable line length, strong heading rhythm, precise callouts.
- Code blocks: deep token surface, simple bar, copy-friendly spacing.
- Index pages: spec cards, not marketing cards.
- Docs primitives should come from `@netstamp/ui` when reusable.

### OpenAPI Explorer

The API explorer is a dense controller and reference manual.

- Three-column desktop layout: endpoint sidebar, operation reference, sticky request console.
- Method labels are stateful and legible.
- Paths and examples use mono.
- Parameter, body, and response details use compact structured rows.
- Request console uses tokenized fields, cURL preview, and response panel.
- It should feel like part of the same docs system, not a separate app.

### Storybook

Storybook is the living design-system spec.

- Every exported primitive must have stories.
- Every major primitive needs light and dark examples through the toolbar `data-theme` path.
- Stories should include usage, states, sizes, tones, invalid/disabled/loading variants, and dense composition examples.
- Overview stories should document tokens, surface families, typography, and patterns.
- Avoid one-off demo art. Use Netstamp telemetry and API language.

## 4. Component Specifications

### Surface

`Surface` is the base framed container.

Required behavior:

- Square frame.
- Tokenized tone: `glass`, `matte`, `deep`, `flat`, `accent`, `danger`.
- No decorative shadow.
- Focusable surfaces must expose visible focus outlines.

Tone intent:

- `glass`: legacy name; render as a flat dashboard surface with neutral border.
- `matte`: lower-contrast nested surface.
- `deep`: code, topology, maps, terminals, response previews.
- `flat`: quiet low-emphasis surface.
- `accent`: important CTA or selected product block.
- `danger`: destructive or critical state block.

### Panel

`Panel` adds structure to `Surface`.

- Header supports eyebrow, title, summary, actions.
- Body supports compact content without nested decorative cards.
- Separator uses `--ns-border-faint`.
- Actions should align right on desktop and wrap cleanly on mobile.

### Button

Buttons are direct controls, not decorative pills.

- Text uses sans, normal case, medium-heavy weight.
- Primary: orange fill.
- Secondary: blue muted fill or blue outline.
- Outline: neutral frame.
- Ghost: low-priority action.
- Danger: destructive only.
- Disabled: opacity plus no pointer events.
- Loading: stable dimensions, no layout shift.
- Icon buttons use actual icon primitives and accessible labels.

Sizes:

- `sm`: min-height `2rem`.
- `md`: min-height `2.5rem`.
- `lg`: min-height `3rem`.
- `xl`: min-height `3.5rem`.

### Badge

Badges communicate category or state.

- Tones: `neutral`, `accent`, `success`, `warning`, `critical`, `muted`.
- Optional dot is square.
- State badges include readable labels; color alone is not enough.

### Fields And Selects

Fields use a square control frame.

- Labels are sans, clear, and close to the input.
- Helper and error text are visible and specific.
- Focus uses token outline and must not be clipped.
- Invalid state uses red border plus helper text.
- Selects need visible trigger, popover, active item, disabled item, and keyboard state.

### Data Table

Tables are enterprise dashboard surfaces.

- Headers are sticky when useful.
- Rows use subtle dividers.
- Hover and selected states are tokenized.
- Technical values may use mono within cells.
- Wide tables scroll instead of crushing content.

### Metric Card

Metric cards summarize current state.

- Label, value, delta, state badge, and optional description.
- Values may use display or mono depending on meaning.
- Cards align heights in grids.

### Terminal And Code

Terminal blocks represent commands, snippets, logs, and responses.

- Deep surface.
- Mono text.
- Optional simple top bar.
- Horizontal scrolling for long lines.
- Orange for primary command prompt or active marker, blue for secondary reference.

### Navigation

Navigation is quiet and precise.

- Public homepage/docs navigation follows the Sui-style structure: a fixed floating top bar over the page, not a sticky header and not a full-width strip attached to the viewport edge. Logo stays on the left, primary nav text links sit centered, and the right CTA is `Login`.
- Public nav text links are plain text, not buttons: no border, no hover background, `#ffffff` text, `400` font weight, and hover/active changes text color only.
- Public floating nav background is `--ns-chrome-bg` (`#131518`) with `--ns-chrome-text` (`#ffffff`).
- Authenticated web app navigation stays on the left, but the sidebar itself is a floating rail inside the app canvas rather than a full-height wall attached to the viewport edge.
- Active items use orange leading marker or frame.
- Secondary links use blue only when they are real links or secondary actions.
- Collapsed icon-only nav must keep accessible labels.
- Docs TOC active state uses orange text or frame, not glow.

### Overlays

Dialogs, drawers, popovers, select menus, search, and menus share the same system.

- Token scrim.
- Square surface.
- Clear header/action footer.
- Keyboard and focus management through Radix or equivalent primitives.
- No decorative backdrop blur dependency for readability.

## 5. Layout System

Spacing:

- Page stack gap: `1.25rem`.
- Grid gap: `1rem`.
- Compact panel padding: `0.875rem`.
- Standard panel padding: `1.25rem`.
- Rich docs/landing section padding: `clamp(2rem, 5vw, 5rem)`.

Breakpoints:

- Web app sidebar collapses around `58rem`.
- Web app single-column shell below `38rem`.
- Docs sidebars collapse below about `68rem`.
- API explorer console drops below content below about `86rem`.
- API explorer becomes one column below about `64rem`.

Rules:

- Use CSS Grid for page skeletons and table-like product layouts.
- Use Flexbox for action rows, metadata, toolbars, and compact controls.
- Avoid cards inside cards unless the inner item is a repeated row/cell.
- Use stable dimensions for controls, boards, counters, tiles, and toolbar items.
- Do not scale font size with viewport width for controls or compact UI.

## 6. Visual Data

Network visuals are diagnostic, not decorative.

- Use orange for the primary active path or selected series.
- Use blue for secondary paths or comparison series.
- Use green, amber, and red only for semantic health/state.
- Use neutral slate for baselines, grids, and inactive lines.
- Tooltips use high-contrast deep surfaces.
- Grid lines and axes are low contrast.
- Area fills are flat low-opacity colors, never gradients.
- Motion is slow, useful, and disabled for reduced motion.

Approved motifs:

- Probe grids.
- Route paths.
- Latency rails.
- Hop timelines.
- DNS/result tables.
- Packet or heartbeat movement when it communicates state.
- Controller/probe topology maps.

## 7. Accessibility

Accessibility is part of the visual system.

- Use semantic landmarks: `nav`, `main`, `section`, `article`, `aside`, `table`, `header`, `footer`.
- Use visible `:focus-visible` outlines.
- Do not rely on color alone for state.
- Keep touch targets near `2rem-2.75rem` or larger.
- Use `100svh` for full-height shells.
- Let dense tables and code blocks scroll horizontally.
- Keep sticky regions from trapping mobile scroll.
- Provide labels or titles for icon-only controls.
- Respect `prefers-reduced-motion`.

## 8. Implementation Rules

- Reusable primitives belong in `@netstamp/ui`.
- Feature-only UI belongs under `web/src/features/<feature>/components`.
- App-level layout helpers stay in `web/src/shared/components`.
- Docs-specific Astro/React UI stays under `docs/src/components`.
- Prefer one CSS module per component or route section.
- Avoid broad new global stylesheets.
- Use `--ns-*` tokens for colors, fonts, borders, state, focus, and transitions.
- Use `@netstamp/brand` assets for brand marks and favicon.
- Use actual product screenshots or product-like panels on the homepage.
- Use icons from the existing icon systems. Do not add decorative emoji icons.

## 9. Anti-Patterns

Do not add:

- Decorative blobs, orbs, bokeh, radial glows, scanlines, dot fields, or grid backgrounds.
- Pastel SaaS gradients or glossy glassmorphism.
- Heavy shadows.
- Rounded consumer-app cards or pill-shaped core controls.
- Purple/rainbow accents.
- Stock photography or atmospheric images.
- Centered generic marketing sections that hide product detail.
- Custom clipped polygons for ordinary UI.
- Duplicate local primitives when `@netstamp/ui` can own the pattern.
- Copy that sounds like generic productivity software.

## 10. Verification Checklist

Before shipping a frontend change, verify:

- `tokens.css` matches the token values and roles in this file.
- The root `data-theme` path works in app, docs, and Storybook.
- Light and dark mode use the same layout and component structure.
- Orange is the primary interaction color.
- Blue is secondary/reference/data.
- Semantic colors are reserved for semantic state.
- Core controls and panels are square or nearly square.
- Shadows, glows, gradients, and decorative patterns are absent.
- Text fits at mobile and desktop widths.
- Wide tables and code blocks scroll instead of overflowing.
- Focus states are visible and not clipped.
- Storybook covers exported primitives and major states.
- Docs and app reuse `@netstamp/ui` primitives instead of recreating them locally.
- Copy names concrete network behavior.
