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
- Neutral ink, graphite, white, cool gray, and pale blue foundations.
- Square or nearly-square frames. Core controls use zero radius.
- Minimal shadows. If depth is needed, use borders, surface contrast, and spatial grouping first.
- Gradients are not part of the product UI. Use flat token colors.
- Richness comes from layout, data, typography, state, and component completeness, not decoration.

## 2. Token Contract

Use `--ns-*` tokens from `packages/ui/src/styles/tokens.css`. Do not create new local color systems in docs, Storybook, or web routes.

### Typography Tokens

- `--ns-font-sans`: `"TASAOrbiter", "TASA Orbiter", sans-serif`.
- `--ns-font-display`: `"Arial Black", "TASAExplorer", "TASA Explorer", sans-serif`.
- `--ns-font-mono`: `"JetBrainsMono", "JetBrains Mono", monospace`.

Usage:

- Display: page titles, landing hero, major docs headings, large metric values.
- Sans: body copy, form labels, app navigation, table text, buttons, compact panel headings.
- Mono: IDs, tokens, timestamps, paths, command text, code, method labels, response previews, technical metadata.

Do not use display-size type inside compact cards, sidebars, tables, nav, or controls.

### Light Theme Color Targets

These are target token values for the spec-docs system. `tokens.css` must match these values after the implementation refactor.

- `--ns-bg`: `#f6f8fb`, page canvas.
- `--ns-bg-section`: `#edf2f7`, broad docs and landing bands.
- `--ns-bg-subtle`: `#f1f5f9`, nested quiet surface.
- `--ns-surface`: `#ffffff`, primary panels and cards.
- `--ns-surface-raised`: `#fbfcfe`, toolbars, sticky controls, raised rows.
- `--ns-surface-deep`: `#e7edf5`, code, topology, maps, consoles, response previews.
- `--ns-glass-dark`: `rgba(255, 255, 255, 0.94)`, sticky light chrome.
- `--ns-glass-light`: `rgba(15, 23, 42, 0.035)`, faint overlay.
- `--ns-overlay-scrim`: `rgba(15, 23, 42, 0.18)`, modal/search scrim.
- `--ns-text`: `#111827`, primary text.
- `--ns-text-muted`: `#3f4b5f`, body and secondary text.
- `--ns-text-subtle`: `#64748b`, supporting text.
- `--ns-text-low`: `#8a96a8`, metadata and inactive controls.
- `--ns-text-on-accent`: `#ffffff`, text on filled primary controls.
- `--ns-primary`: `#e65f1a`, Netstamp orange primary.
- `--ns-primary-hover`: `#c94f12`, primary hover.
- `--ns-primary-active`: `#9f3d0d`, primary pressed and focus.
- `--ns-primary-muted`: `rgba(230, 95, 26, 0.1)`, faint selected/hover fill.
- `--ns-primary-subtle`: `rgba(230, 95, 26, 0.16)`, stronger selected fill.
- `--ns-primary-border`: `rgba(230, 95, 26, 0.42)`, active frame.
- `--ns-secondary`: `#2867b2`, blue secondary action/data.
- `--ns-secondary-hover`: `#1f5597`, blue hover.
- `--ns-secondary-active`: `#173f73`, blue pressed.
- `--ns-secondary-muted`: `rgba(40, 103, 178, 0.08)`, faint blue fill.
- `--ns-secondary-subtle`: `rgba(40, 103, 178, 0.14)`, stronger blue fill.
- `--ns-secondary-border`: `rgba(40, 103, 178, 0.3)`, blue frame.
- `--ns-critical`: `#c9362c`, destructive/error/failed.
- `--ns-warning`: `#b7791f`, warning/degraded/pending.
- `--ns-success`: `#168a45`, healthy/online/success.
- `--ns-metal`: `#667085`, neutral technical accent.
- `--ns-slate-line`: `rgba(102, 112, 133, 0.24)`, chart and divider line.
- `--ns-border`: `#d7dee8`, default frame.
- `--ns-border-strong`: `#aab6c6`, strong frame.
- `--ns-border-faint`: `#e6ebf2`, subtle divider.

### Dark Theme Color Targets

Dark mode is the same product system on a darker canvas, not a separate cyber-console theme.

- `--ns-bg`: `#0d1520`.
- `--ns-bg-section`: `#101b29`.
- `--ns-bg-subtle`: `#142132`.
- `--ns-surface`: `#151f2d`.
- `--ns-surface-raised`: `#1b2838`.
- `--ns-surface-deep`: `#08111b`.
- `--ns-glass-dark`: `rgba(21, 31, 45, 0.94)`.
- `--ns-glass-light`: `rgba(255, 255, 255, 0.045)`.
- `--ns-overlay-scrim`: `rgba(8, 17, 27, 0.42)`.
- `--ns-text`: `#f8fafc`.
- `--ns-text-muted`: `#d7dde7`.
- `--ns-text-subtle`: `#9ca8b8`.
- `--ns-text-low`: `#738196`.
- `--ns-text-on-accent`: `#1f1207`.
- `--ns-primary`: `#ff9448`.
- `--ns-primary-hover`: `#ffb172`.
- `--ns-primary-active`: `#f97316`.
- `--ns-primary-muted`: `rgba(255, 148, 72, 0.13)`.
- `--ns-primary-subtle`: `rgba(255, 148, 72, 0.2)`.
- `--ns-primary-border`: `rgba(255, 148, 72, 0.46)`.
- `--ns-secondary`: `#6ea8ff`.
- `--ns-secondary-hover`: `#9cc3ff`.
- `--ns-secondary-active`: `#4c8df0`.
- `--ns-secondary-muted`: `rgba(110, 168, 255, 0.11)`.
- `--ns-secondary-subtle`: `rgba(110, 168, 255, 0.17)`.
- `--ns-secondary-border`: `rgba(110, 168, 255, 0.36)`.
- `--ns-critical`: `#ff6b63`.
- `--ns-warning`: `#facc15`.
- `--ns-success`: `#34c77b`.
- `--ns-metal`: `#98a4b5`.
- `--ns-slate-line`: `rgba(152, 164, 181, 0.24)`.
- `--ns-border`: `#2b3a4d`.
- `--ns-border-strong`: `#4a5f78`.
- `--ns-border-faint`: `#223044`.

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
