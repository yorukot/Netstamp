# Netstamp Design Guidelines

> Category: network observability and developer infrastructure. Current direction: authenticated web app, public homepage, docs, Storybook, and API reference all align to the dashboard surface system: light-default B2B product UI, matching dark mode, restrained blue/cyan brand accents, square frames, flat panels, no gradients, and no grid or decorative background patterns.

This document is the source of truth for Netstamp frontend design across the React web app, the Astro docs site, the public landing page, the OpenAPI explorer, Storybook, and shared `@netstamp/ui` primitives.

Keep new UI aligned with these files first:

- `packages/ui/src/styles/tokens.css`: shared fonts, colors, shell utilities, frame utility, scrollbars, and base reset.
- `packages/ui/src/components`: shared primitives used by both web and docs.
- `web/src/index.css`: app root import of shared styles.
- `web/src/layouts/AppShell.module.css`: authenticated product shell.
- `web/src/shared/components`: product page layout helpers.
- `docs/src/styles/docs.css`: documentation shell, doc prose, cards, search, and navigation.
- `docs/src/components/landing/LandingPage.module.css`: public landing page with dashboard-aligned product screenshots and telemetry panels.
- `docs/src/components/openapi/OpenAPIExplorer.module.css`: API reference and request console.

## 1. Product Feel

Netstamp should feel like serious infrastructure software for repeated operational work. The authenticated app should be clear, dense, calm, and trustworthy rather than cinematic or decorative.

Core authenticated app traits:

- Light, neutral workspace with white and gray surfaces.
- Dark mode must use the same dashboard structure, density, and token roles rather than a separate console-only visual system.
- Blue/cyan as the reserved brand accent for primary actions, active navigation, focus, and selected state.
- Status colors carry semantic meaning only: healthy, warning, critical, pending, or neutral.
- Square, flat frames; no rounded corners and no cut-corner framing in the web app.
- Sans typography for navigation, headings, controls, tables, labels, and buttons.
- Monospace typography only for IDs, tokens, timestamps, code, CLI commands, and machine-readable telemetry.
- Dense layouts: few decorative objects, high information clarity.
- Plain token backgrounds only; do not use CSS gradients, SVG gradients, ECharts gradient color stops, glow filters, scanline overlays, or grid/background patterns.
- Infrastructure copy that names probes, checks, routes, DNS, latency, packet loss, topology, path hash, heartbeat, and controller actions.

Each surface has a different density:

- Web app: compact light-default B2B dashboard for repeated work, with matching dark mode.
- Public landing page: dashboard-aligned product homepage with real dashboard screenshots, compact telemetry, and clear deploy/GitHub actions.
- Docs: readable technical reference using the same dashboard surfaces, navigation density, and color mode behavior.
- OpenAPI explorer: dashboard-aligned split reference and request console, optimized for scanning methods, paths, parameters, snippets, and responses.

## 2. Tokens

Use existing `--ns-*` tokens. Do not introduce parallel token names unless a new primitive is genuinely required and belongs in `packages/ui/src/styles/tokens.css`.

### Fonts

- `--ns-font-sans`: `"TASAOrbiter", "TASA Orbiter", sans-serif`.
- `--ns-font-display`: `"Arial Black", "TASAExplorer", "TASA Explorer", sans-serif`.
- `--ns-font-mono`: `"JetBrainsMono", "JetBrains Mono", monospace`.

Use the display face for big product statements and screen titles. Use the sans face for explanatory body copy. Use the mono face for operational UI, navigation, labels, data, code, buttons, and metadata.

### Backgrounds And Surfaces

- `--ns-bg`: `#010203`, main page canvas.
- `--ns-bg-section`: `#030507`, major section background.
- `--ns-bg-subtle`: `#05080c`, subtle elevated area.
- `--ns-surface`: `#07090d`, base panel surface.
- `--ns-surface-raised`: `#0c1016`, stronger raised surface.
- `--ns-surface-deep`: `#020304`, deep terminal/map/code surface.
- `--ns-glass-dark`: `rgba(3, 5, 8, 0.94)`, sticky dark glass.
- `--ns-glass-light`: `rgba(255, 255, 255, 0.012)`, faint overlay.

### Text

- `--ns-text`: `#fff7ec`, primary text.
- `--ns-text-muted`: `#ddd4c8`, body and secondary text.
- `--ns-text-subtle`: `#b8b3aa`, supporting copy and inactive controls.
- `--ns-text-low`: `#77736b`, metadata and low-priority labels.
- `--ns-text-on-accent`: high-contrast text on accent controls.

### Accent And State

- `--ns-accent`: blue/cyan brand action color.
- `--ns-accent-hover`: brighter blue/cyan hover and highlight color.
- `--ns-accent-active`: pressed/focused blue color.
- `--ns-accent-muted`: low-opacity selected/hover fill.
- `--ns-accent-subtle`: stronger selected fill.
- `--ns-accent-border`: active frame color.
- `--ns-accent-glow`: restrained blue/cyan glow.
- `--ns-glass-accent`: low-opacity blue/cyan glass overlay.
- `--ns-critical`: `#ff453a`, destructive/error/failed.
- `--ns-warning`: `#ff9f0a`, warning/waiting/degraded/pending.
- `--ns-success`: `#30d158`, healthy/online/success.
- `--ns-metal`: `#c4ccd9`, neutral technical accent.
- `--ns-slate-line`: `rgba(196, 204, 217, 0.22)`, neutral chart and divider line.

Blue/cyan is the reserved brand interaction color. Green, amber, and red are reserved for state meaning. Slate may be used for chart/data series where it improves dashboard readability. Avoid rainbow, pastel, purple, orange-first, glossy gradient, or decorative color systems.

### Borders, Radius, Shadows

- `--ns-border`, `--ns-border-strong`, and `--ns-border-faint`: neutral frame and divider colors.
- `--ns-radius-*`: app values should remain `0`.
- `--ns-cut-*`: legacy tokens only; do not use them for new cut-corner UI.
- `--ns-shadow-sm`, `--ns-shadow-md`, and `--ns-shadow-glow`: app values should remain `none`.
- `--ns-transition`: `180ms cubic-bezier(0.2, 0.8, 0.2, 1)`.

Use square frames instead of clipped, cut-corner, rounded, or pill-shaped rectangles. Avoid decorative shadows, glows, and inset lighting effects.

## 3. Page Families

### Shared App And Site Canvas

Full-page surfaces use plain token backgrounds:

- Page canvas: `var(--ns-bg)`.
- App, docs, and landing panels: `var(--ns-surface)`.
- Raised rows, sidebars, code bars, and tool headers: `var(--ns-surface-raised)`.
- Deep code, map, terminal, or response blocks: `var(--ns-surface-deep)`.

Do not add gradient washes, radial glows, dot fields, scanlines, diagonal micro-patterns, or grid backgrounds. Existing `.ns-grid-shell` and `.ns-grid-shell--constellation` are compatibility names only and must render as plain token backgrounds.

### Public Homepage

Current implementation lives in `docs/src/components/landing`.

Use the homepage to show Netstamp as a real product:

- Light-default dashboard canvas with matching dark mode and blue/cyan product accent.
- Sticky docs top nav above the page.
- Hero with Netstamp/product category headline, concise offer copy, and a real product dashboard screenshot.
- Primary CTA is blue/cyan fill; secondary CTA is dark/neutral.
- Product sections cover Fleet, Checks, Insight, Alerts, API/automation, and Open Source.
- Product context should come from real UI screenshots, telemetry chips, route paths, status indicators, and compact panels, not decorative backgrounds.
- Final CTA/trust area should stay grounded in open source, deployability, probes, and measurable network behavior.

Landing copy should be short, concrete, and infrastructure-oriented:

- "See the network. Before it fails you."
- "Open-source network observability from probes you control."
- "Measure latency, packet loss, DNS, and routes."
- "Your traffic has a story. Netstamp shows the path."

Do not turn the homepage into a centered generic marketing template. Avoid stock photography, lifestyle imagery, decorative blobs, abstract SaaS gradients, grid backgrounds, orange-first styling, or cinematic scenes that hide the product.

### Web App

Current implementation lives under `web/src`.

The authenticated app is optimized for repeated operational work:

- Two-column shell at desktop: `17rem` sticky sidebar and fluid content.
- Sidebar uses brand mark, project selector, nav links, and user card.
- Content uses `3rem 1.5rem 0` desktop padding and compact page stacks.
- Sidebar collapses to `6rem` below about `58rem`.
- The shell becomes single-column with horizontal nav below about `38rem`.
- Product pages use `ScreenHeader`, `PageStack`, `ResponsiveGrid`, `ActionRow`, panels, metric cards, data tables, network maps, and key-value grids.

Keep product pages scan-friendly:

- One large screen title per route.
- Actions aligned to the header when possible.
- Panels are the main grouping unit.
- Dense data belongs in tables, key-value grids, timelines, or terminal/code blocks.
- Wide data may scroll horizontally rather than becoming unreadably compressed.

### Documentation Site

Current implementation lives under `docs/src/styles/docs.css` and `docs/src/layouts/DocLayout.astro`.

Docs must feel like Netstamp, but prioritize reading and navigation:

- Sticky top nav with brand, docs/storybook/API links, GitHub, color mode toggle, and app/deploy action.
- Three-column desktop shell: left docs navigation, central content, right table of contents.
- Left nav can fold to icon-only on wide layouts.
- Below about `68rem`, collapse to one column with sidebar, content, and TOC stacked.
- Doc hero is a framed blue-accent panel with compact display title and body summary.
- Prose uses generous line-height and clear heading spacing.
- Code blocks use token deep surfaces, simple top bars, and accent borders.
- Cards, callouts, pager links, search results, and index grids use dashboard token surfaces with accent hover/active states.

Docs must follow the dashboard surface system and color mode behavior. Do not reintroduce a separate dark-only docs theme or cut-corner treatments.

### OpenAPI Explorer

Current implementation lives in `docs/src/components/openapi`.

The API explorer is denser than standard docs:

- Desktop shell has three columns: endpoint sidebar, reference content, sticky request console.
- Below about `86rem`, console drops below content.
- Below about `64rem`, sidebar/content/console collapse to one column.
- Method labels use state colors: GET green, POST amber, PUT/PATCH yellow, DELETE red.
- Endpoint rows use monospace method and path with active accent frame.
- Operation articles pair explanatory content with snippets and structured field/response rows.
- Console uses sticky token surfaces, form controls, body textarea, cURL preview, and response panel.
- Snippets and response previews use monospace, deep token backgrounds, and horizontal/vertical scrolling when needed.

Do not make the API explorer airy or editorial. It should feel like a controller plus reference manual.

## 4. Typography

Use typography to separate marketing claims, app structure, and machine-readable data.

### Display

Use `--ns-font-display` for:

- Landing hero and section headlines.
- Web app screen titles.
- Docs hero titles and major prose headings.
- OpenAPI tag and operation headings.

Typical scale:

- Landing hero `h1`: `clamp(2.4rem, 5.2vw, 6rem)`, line-height `0.9`.
- Landing story headline: `clamp(3.25rem, 7vw, 8rem)` when used, line-height `0.9`.
- Web app screen title: `clamp(3rem, 6vw, 5.75rem)`, line-height `0.9`.
- Auth hero title: up to `clamp(3rem, 5vw, 7rem)`.
- Docs hero title: `clamp(1.4rem, 2.5vw, 2rem)`, line-height `1.05`.
- Docs prose `h2`: `clamp(1.6rem, 3vw, 2.6rem)`, line-height `1`.
- OpenAPI hero title: `clamp(2.5rem, 7vw, 4.5rem)`, line-height `0.9`.

### Body

Use `--ns-font-sans` for:

- Landing body copy.
- Docs prose paragraphs and list text.
- Product summaries and form input text.

Typical body text:

- Standard body: `1rem`, line-height `1.65-1.75`.
- Landing/body emphasis: `clamp(1rem, 1.4vw, 1.25rem)`.
- Auth and large intro copy: around `1.125rem`.

### Operational Mono

Use `--ns-font-mono` for:

- Navigation.
- Buttons.
- Labels and eyebrows.
- Badges.
- Tables and cells.
- Metric metadata.
- Code, snippets, endpoint paths, method labels, and response previews.
- TOC and pager metadata.

Typical mono label style:

- Size `0.68rem-0.875rem`.
- Weight `700-900`.
- Uppercase for labels, nav, buttons, and metadata.
- Letter spacing around `0.06em-0.14em` where the current UI already uses it.

Do not use display-size type inside compact panels, sidebars, tables, or controls. Long labels must truncate, wrap, or collapse before they overflow.

## 5. Layout System

### Spacing

Default layout rhythm:

- Page stack gap: `1.25rem`.
- Grid gap: `1rem`.
- Panel padding: `0.875rem`, `1.25rem`, or `1.35rem`.
- Docs shell gap: `2rem` desktop, `1rem` mobile.
- Landing section padding can be larger, usually `clamp()` based, but must keep continuity with the dashboard panel rhythm.

Prefer CSS Grid for page structure and Flexbox for button rows, metadata rows, and compact horizontal controls.

### Product Page Layouts

Use:

- `ScreenHeader` for page title, eyebrow, description, and actions.
- `PageStack` for vertical route rhythm.
- `ResponsiveGrid` with two or three columns for panels.
- Single-column layout below `58rem-78rem`, based on content density.

### Docs Layouts

Use:

- `DocLayout.astro` for all markdown documentation.
- `docHero` for title/description.
- `docProse` for markdown content.
- `cardGrid` and `docIndexGrid` for index navigation.
- `callout-*` classes for notes, tips, warnings, and cautions.
- `docPager` and `editLink` for end-of-page navigation.

### Density

Match density to context:

- Landing: dashboard-aligned, product-led, moderately broad, screenshot-forward.
- Web app: compact, operational, panel-based.
- Docs prose: readable, less dense.
- OpenAPI: very dense, with sticky navigation and console tools.

## 6. Component Language

Use `@netstamp/ui` primitives before adding local controls. Shared components should stay in `packages/ui`; feature-only UI should stay colocated with the feature or docs area.

### Surface And Panel

`Surface` is the base framed container. `Panel` adds structured header, title, actions, and separator.

Use tones intentionally:

- `glass`: legacy tone name; render it as a flat dashboard surface with neutral border.
- `matte`: lower-contrast nested section.
- `deep`: maps, terminals, code, diagnostics, and high-depth blocks.
- `flat`: quiet simple surface.
- `accent`: CTA or important highlighted block.
- `danger`: destructive or critical state block.

Panel rules:

- Keep one panel header structure: eyebrow, title, optional actions.
- Avoid cards inside cards unless the inner item is a repeated row/cell.
- Use nested cells for key-value data, event feeds, steps, timelines, route diffs, and compact summaries.

### Buttons

Buttons live in `@netstamp/ui` and use square rectangular frames.

- Text is sans, normal case, and medium-heavy weight.
- Primary uses blue/cyan fill and high-contrast text from `--ns-text-on-accent`.
- Secondary uses neutral raised surface and standard text.
- Outline uses neutral surface and neutral hover.
- Ghost is for low-priority actions.
- Danger is only for destructive actions.
- Avoid hover lift for app controls.
- Disabled buttons use opacity and no pointer events.

Button sizes:

- `sm`: `0.75rem`, min-height `2rem`.
- `md`: `0.8125rem`, min-height `2.5rem`.
- `lg`: `0.875rem`, min-height `3rem`.
- `xl`: `0.9375rem`, min-height `3.5rem`.

Use icons from the existing icon system in docs or Phosphor web components where already used. Do not create decorative emoji-style icons.

### Badges

Badges can represent operational labels or state.

- Use sans, normal-case text.
- Optional dot is square and uses current color.
- Approved tones: `neutral`, `accent`, `success`, `warning`, `critical`, `muted`.
- Badge color must communicate category or state, not decoration.

### Fields And Selects

Fields use a square `controlFrame`.

- Labels are sans and normal case.
- Control background is neutral surface.
- Focus uses `--ns-accent-active`, visible outline, and no hidden focus ring.
- Invalid state uses red border and clear helper text.
- Select arrow is CSS-generated and turns with open state.
- Compact controls use smaller sizing for dashboards/API panels.

### Data Tables

Data tables should feel like enterprise dashboard data surfaces.

- Use sans for headers and cells; reserve mono for technical values inside cells.
- Sticky headers can use neutral raised surfaces.
- Hover rows use neutral backgrounds; selected rows may use faint accent backgrounds.
- Thin row borders with low contrast.
- Allow horizontal scrolling for wide content.

### Metric Cards

Metric cards summarize state quickly.

- Large display numeric value.
- Sans muted label.
- Optional badge for category/state.

### Terminal And Code Blocks

Terminal blocks are command surfaces.

- Deep black background.
- Monospace text.
- Blue/cyan accent or faint white frame.
- Optional top bar or compact status markers for prose code blocks.
- Snippets and response previews should scroll rather than wrap into unreadable layouts, except API response previews may use `pre-wrap` when readability is better.

### Navigation

Navigation is sans and normal case.

- App sidebar active item uses blue/cyan leading accent and a neutral square frame.
- Docs top nav uses icon plus text, with text hidden on small screens when necessary.
- Docs sidebar groups are collapsible and show active state in the shared accent color.
- TOC active link uses the shared accent color without glow.

### Network Visuals

Network visuals should be abstract, map-like, and diagnostic.

- In the authenticated app, prefer simple square labels and markers.
- Routes, packets, rails, hops, matrices, and topology lines are preferred motifs.
- Active packets and important paths use blue/cyan.
- Green only means online/success.
- Route animations are acceptable when slow, subtle, and useful for the diagnostic state.
- Three.js scenes and animated CSS visuals must respect reduced motion.

## 7. Frame Utility

The legacy `.ns-cut-frame` class name is retained for compatibility, but it now renders as a standard square frame. Do not create new cut-corner polygon frames.

### Standard Utility

```css
.frame {
	--ns-frame-color: var(--ns-border);
	--ns-frame-border-width: 1px;
}
```

Then apply the square frame utility already used by the component, or compose the same border variables locally.

When building a custom frame:

- Existing `.ns-cut-frame` usage is acceptable for compatibility because it now renders square; do not introduce it for new components.
- Use `--ns-frame-color` and `--ns-frame-border-width`; do not hardcode repeated frame values.
- Avoid `clip-path` for ordinary cards, controls, tables, dialogs, and popovers.
- Be careful with `overflow: hidden`; do not clip visible focus outlines.

## 8. Motion

Motion should be restrained, mechanical, and diagnostic.

Use:

- `--ns-transition` for most UI state changes.
- `translateY(-0.0625rem)` for small hover lift.
- `180ms-280ms` for drawer/search/modal entry and exit.
- Slow node blinking, packet routing, and topology reveals on functional network visuals.
- GSAP/Three.js only for functional browser effects; do not use them for decorative backgrounds.

Always support reduced motion:

```css
@media (prefers-reduced-motion: reduce) {
	* {
		animation-duration: 0.01ms;
		animation-iteration-count: 1;
		scroll-behavior: auto;
	}
}
```

Or disable the specific decorative animation in component code, as the landing page does.

Avoid springy, playful, elastic, large parallax, or attention-seeking motion.

## 9. Data Visualization

Charts, maps, route boards, and telemetry widgets should look embedded in the console.

- Primary series: `--ns-accent`, `#2563eb`, or `#38bdf8`.
- Secondary/baseline series: `--ns-metal`, `--ns-slate-line`, or low-opacity white.
- Success state: `--ns-success`.
- Warning state: `--ns-warning`.
- Critical state: `--ns-critical`.
- Tooltip background: near black, around `rgba(10, 13, 18, 0.92)`.
- Tooltip border: faint white or accent when active.
- Axis labels: muted mono, around `10px`.
- Grid lines: very low contrast.
- Area fills: flat low-opacity color only; do not use gradient color stops.

Do not use multicolor chart palettes unless every color maps to a meaningful operational state.

## 10. Copywriting

Netstamp copy should sound like network infrastructure and measurement tooling.

Use phrases like:

- "See the network before it fails you."
- "Open-source network observability from probes you control."
- "Measure latency, packet loss, DNS, and routes."
- "Grid and map views for distributed measurement agents."
- "Path hash changed from previous run."
- "Waiting for first heartbeat."
- "Scheduler / result stream."
- "Live network topology."
- "Route hash diff."

Avoid phrases like:

- "Unlock your potential."
- "Supercharge your workflow."
- "Beautifully simple."
- "AI-powered for everyone."
- "Seamless experience."
- "Delightful collaboration."

The interface should sound like a controller, reference manual, or operator console, not a lifestyle brand.

## 11. Accessibility And Responsiveness

Accessibility is part of the visual system.

- Keep `:focus-visible` outlines visible, usually `2px solid var(--ns-accent)` or `--ns-accent-active`.
- Use semantic landmarks: `nav`, `main`, `section`, `article`, `aside`, `table`, `header`, `footer`.
- Use `aria-hidden="true"` for decorative geometry and icons.
- Do not rely on color alone for important state; include labels, badges, status text, method names, or icons.
- Keep touch targets around `2rem-2.75rem` high or larger for controls.
- Use `100svh` for full-height shells.
- Collapse dense grids to one column at mobile breakpoints.
- Let wide data tables and code blocks scroll horizontally.
- Keep sticky sidebars from becoming trapped scroll regions on small screens.
- Ensure icon-only collapsed nav items have accessible labels and useful `title` text where applicable.

## 12. Implementation Rules

### File Organization

- Use `@netstamp/ui` for reusable primitives before creating local controls.
- Put feature-only React UI under `web/src/features/<feature>/components`.
- Put app-level shared React components under `web/src/shared/components`.
- Put docs-specific Astro/React UI under `docs/src/components`.
- Prefer one CSS module per component or route section.
- Avoid broad new global stylesheets; shared global primitives belong in `packages/ui/src/styles/tokens.css`.

### CSS

- Use CSS modules for component styling.
- Use `--ns-*` tokens for colors, fonts, borders, and transitions.
- Use `class:list` in Astro and local `classNames` helpers in React when composing frame classes.
- Keep page sections full-width/unframed unless they are real tools, panels, repeated cards, or docs content blocks.
- Avoid nested decorative cards.
- Keep core UI square.

### Assets And Icons

- Use `@netstamp/brand` assets for brand marks and favicon.
- Docs currently use Phosphor web components and `docs/src/components/icons/Icon.astro`.
- Use actual product screenshots and product/network visuals for landing visuals; generated scenes are secondary.
- Do not add stock-like atmospheric imagery when a diagnostic visual would communicate better.

## 13. Anti-Patterns

Do not add:

- Oversized decorative cards, pill-shaped core controls, or soft consumer-app surfaces.
- Pastel SaaS gradients, glossy blobs, heavy glassmorphism, decorative orbs, grid backgrounds, radial glows, scanline overlays, SVG/CSS/ECharts gradients, or non-product hero decoration.
- Purple/rainbow/non-token brand accents.
- Decorative emoji icons.
- Soft lifestyle photography or stock photos.
- Centered generic landing sections that ignore the dashboard panel rhythm.
- Decorative or heavy shadows.
- Global CSS systems outside existing `--ns-*` tokens and CSS modules.
- Copy that sounds like a generic productivity app.

## 14. Review Checklist

Before shipping frontend UI, check:

- Does it use existing `--ns-*` tokens?
- Does it align with the dashboard surface system in both light and dark mode?
- Are backgrounds plain token colors, with no gradients, grids, glows, or decorative patterns?
- Does it use the correct font family for display, body, and operational text?
- Are framed controls square and free of custom clipped polygons?
- Is blue/cyan the main interactive accent?
- Are green, amber, and red reserved for state?
- Does the layout follow the established spacing rhythm unless a larger landing/docs rhythm is intentional?
- Does the UI match the density of its surface: landing, app, docs, or OpenAPI?
- Does it collapse cleanly on the relevant breakpoints?
- Are focus states visible and not clipped by overflow?
- Are animations disabled or reduced for `prefers-reduced-motion`?
- Does copy name concrete network behavior instead of generic product benefits?
- If the change affects shared conventions, should this `design.md` or an area `AGENTS.md` be updated?
