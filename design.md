# Netstamp Design Guidelines

> Category: network observability and developer infrastructure. Current frontend: dark engineering console, orange-accented probe and route diagnostics, clipped industrial frames, monospace operational navigation, and dense technical documentation.

This document is the source of truth for Netstamp frontend design across the React web app, the Astro docs site, the public landing page, the OpenAPI explorer, Storybook, and shared `@netstamp/ui` primitives.

Keep new UI aligned with these files first:

- `packages/ui/src/styles/tokens.css`: shared fonts, colors, shell utilities, cut-frame utility, shadows, scrollbars, and base reset.
- `packages/ui/src/components`: shared primitives used by both web and docs.
- `web/src/index.css`: app root import of shared styles.
- `web/src/layouts/AppShell.module.css`: authenticated product shell.
- `web/src/shared/components`: product page layout helpers.
- `docs/src/styles/docs.css`: documentation shell, doc prose, cards, search, and navigation.
- `docs/src/components/landing/LandingPage.module.css`: public landing page and animated network storytelling.
- `docs/src/components/openapi/OpenAPIExplorer.module.css`: API reference and request console.

## 1. Product Feel

Netstamp should feel like a network operations console, not a generic SaaS dashboard. The design language is precise, dark, gridded, technical, and built around distributed probes, routes, checks, and controller output.

Core traits:

- Near-black operational canvas.
- Orange as the primary brand and interaction accent.
- Thin engineering grid lines, diagnostic overlays, and route/packet motifs.
- Cut-corner frames for primary surfaces and controls.
- Display typography for hero claims and screen titles.
- Monospace typography for navigation, labels, metadata, buttons, tables, telemetry, code, and API details.
- Sparse but dense layouts: few decorative objects, high information clarity.
- Infrastructure copy that names probes, checks, routes, DNS, latency, packet loss, topology, path hash, heartbeat, and controller actions.

The shared direction applies everywhere, but each surface has a different density:

- Public landing page: cinematic network storytelling with large display type, technical visuals, and clear deploy/GitHub actions.
- Web app: compact operational dashboard for repeated work.
- Docs: readable technical reference that still uses the same dark console structure.
- OpenAPI explorer: split reference and request console, optimized for scanning methods, paths, parameters, snippets, and responses.

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
- `--ns-text-on-accent`: `#0b0704`, text on orange controls.

### Accent And State

- `--ns-accent`: `#ff7a1a`, primary CTA and active brand color.
- `--ns-accent-hover`: `#ff9a3d`, hover and highlighted orange.
- `--ns-accent-active`: `#ff5f00`, pressed/focused orange.
- `--ns-accent-muted`: `rgba(255, 122, 26, 0.16)`, selected/hover fill.
- `--ns-accent-subtle`: `rgba(255, 122, 26, 0.22)`, stronger selected fill.
- `--ns-accent-border`: `rgba(255, 122, 26, 0.66)`, active frame color.
- `--ns-accent-glow`: `rgba(255, 122, 26, 0.3)`, restrained orange glow.
- `--ns-glass-orange`: `rgba(255, 122, 26, 0.22)`, orange glass overlay.
- `--ns-critical`: `#ff453a`, destructive/error/failed.
- `--ns-warning`: `#ff9f0a`, warning/waiting/degraded/pending.
- `--ns-success`: `#30d158`, healthy/online/success.
- `--ns-metal`: `#c4ccd9`, neutral technical accent.
- `--ns-slate-line`: `rgba(196, 204, 217, 0.22)`, neutral chart and divider line.

Orange is the brand and interaction color. Green, yellow, and red are reserved for state meaning. Do not add blue, purple, rainbow, pastel, or glossy gradients for visual variety. The only current non-state blue appears inside OpenAPI syntax highlighting for booleans/null values; keep that usage tiny and code-specific.

### Borders, Cuts, Shadows

- `--ns-border`: `rgba(255, 255, 255, 0.28)`, default frame.
- `--ns-border-strong`: `rgba(255, 255, 255, 0.48)`, hover/active neutral frame.
- `--ns-border-faint`: `rgba(255, 255, 255, 0.16)`, quiet dividers and nested frames.
- `--ns-radius-*`: all `0`.
- `--ns-cut-xs`: `0.375rem`, badges and compact labels.
- `--ns-cut-sm`: `0.5rem`, buttons, nav items, fields, small frames.
- `--ns-cut-md`: `0.75rem`, tables and medium cards.
- `--ns-cut-lg`: `1rem`, panels, landing feature cards, footer blocks.
- `--ns-shadow-sm`: dark technical elevation for small surfaces.
- `--ns-shadow-md`: stronger elevation for modals, hero panels, and large surfaces.
- `--ns-shadow-glow`: restrained orange focus/brand glow.
- `--ns-transition`: `180ms cubic-bezier(0.2, 0.8, 0.2, 1)`.

Use cut corners instead of rounded rectangles. Small circular details are allowed only for dots, status lights, orbit visuals, and similar diagnostic marks.

## 3. Page Families

### Shared App And Site Canvas

Most full-page surfaces use layered dark grids:

- Large orange grid around `6rem-8rem`, low opacity.
- Fine white grid around `2rem`, very low opacity.
- Optional radial orange glow near areas of attention.
- Optional diagonal micro-pattern on panels, sidebars, maps, user cards, and landing sections.

Use `.ns-grid-shell` or `.ns-grid-shell--constellation` when a full app surface can use the shared utility. Otherwise match the same layered background locally.

### Public Landing Page

Current implementation lives in `docs/src/components/landing`.

Use the landing page for high-impact storytelling:

- Full-height dark grid canvas.
- Sticky docs top nav above the page.
- Hero with large left-side claim and a full technical network animation.
- Primary CTA is orange fill; secondary CTA is dark/neutral.
- Story sections use animated network/topology scenes, probe scenes, route boards, check cards, and numbered clipped feature cards.
- Feature cards can be large and visual, but they must remain technical and specific.
- Final CTA/trust area should stay grounded in open source, deployability, probes, and measurable network behavior.

Landing copy should be short, concrete, and infrastructure-oriented:

- "See the network. Before it fails you."
- "Open-source network observability from probes you control."
- "Measure latency, packet loss, DNS, and routes."
- "Your traffic has a story. Netstamp shows the path."

Do not turn the landing page into a centered generic marketing template. Avoid stock photography, lifestyle imagery, decorative blobs, or abstract SaaS gradients.

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

- Sticky top nav with brand, docs/storybook/API links, GitHub, and app/deploy action.
- Three-column desktop shell: left docs navigation, central content, right table of contents.
- Left nav can fold to icon-only on wide layouts.
- Below about `68rem`, collapse to one column with sidebar, content, and TOC stacked.
- Doc hero is a framed orange-accent panel with compact display title and body summary.
- Prose uses generous line-height and clear heading spacing.
- Code blocks use terminal-style top bars, dark backgrounds, and orange borders.
- Cards, callouts, pager links, search results, and index grids use low-contrast dark surfaces with orange hover/active states.

Docs are allowed to be less clipped than the app in simple prose areas, but should still avoid rounded card/pill language.

### OpenAPI Explorer

Current implementation lives in `docs/src/components/openapi`.

The API explorer is denser than standard docs:

- Desktop shell has three columns: endpoint sidebar, reference content, sticky request console.
- Below about `86rem`, console drops below content.
- Below about `64rem`, sidebar/content/console collapse to one column.
- Method labels use state colors: GET green, POST orange, PUT/PATCH yellow, DELETE red.
- Endpoint rows use monospace method and path with active orange frame.
- Operation articles pair explanatory content with snippets and structured field/response rows.
- Console uses sticky dark glass, form controls, body textarea, cURL preview, and response panel.
- Snippets and response previews use monospace, deep black backgrounds, and horizontal/vertical scrolling when needed.

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
- Landing section padding can be larger, usually `clamp()` based, but must leave visual continuity with the grid.

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

- Landing: broad, visual, high contrast.
- Web app: compact, operational, panel-based.
- Docs prose: readable, less dense.
- OpenAPI: very dense, with sticky navigation and console tools.

## 6. Component Language

Use `@netstamp/ui` primitives before adding local controls. Shared components should stay in `packages/ui`; feature-only UI should stay colocated with the feature or docs area.

### Surface And Panel

`Surface` is the base clipped container. `Panel` adds structured header, eyebrow, title, actions, and separator.

Use tones intentionally:

- `glass`: default raised section with orange frame and subtle striped depth.
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

Buttons live in `@netstamp/ui` and remain rectangular with cut corners.

- Text is mono, uppercase, heavy weight, with `0.06em` letter spacing.
- Primary uses orange fill and dark text.
- Secondary uses dark surface and light text.
- Outline uses dark surface, muted text, and orange hover.
- Ghost is for low-priority actions.
- Danger is only for destructive actions.
- Hover may translate upward by `-0.0625rem`.
- Disabled buttons use opacity and no pointer events.

Button sizes:

- `sm`: `0.75rem`, min-height `2rem`.
- `md`: `0.8125rem`, min-height `2.5rem`.
- `lg`: `0.875rem`, min-height `3rem`.
- `xl`: `0.9375rem`, min-height `3.5rem`.

Use icons from the existing icon system in docs or Phosphor web components where already used. Do not create decorative emoji-style icons.

### Badges

Badges can represent operational labels or state.

- Use mono uppercase text.
- Optional dot is square and uses current color.
- Approved tones: `neutral`, `accent`, `success`, `warning`, `critical`, `muted`.
- Badge color must communicate category or state, not decoration.

### Fields And Selects

Fields use a clipped `controlFrame`.

- Labels are mono uppercase.
- Control background is near black.
- Focus uses `--ns-accent-active`, visible outline, and no hidden focus ring.
- Invalid state uses red border and low red glow.
- Select arrow is CSS-generated and turns with open state.
- Compact controls use mono and smaller sizing for dashboards/API panels.

### Data Tables

Data tables should feel like controller output.

- Use monospace for headers and cells.
- Sticky headers can use dark raised surfaces.
- Hover and selected rows use faint orange backgrounds.
- Thin row borders with low contrast.
- Allow horizontal scrolling for wide content.

### Metric Cards

Metric cards summarize state quickly.

- Large display numeric value.
- Mono orange label.
- Optional badge for category/state.
- Diagnostic bracket or corner detail is acceptable when restrained.

### Terminal And Code Blocks

Terminal blocks are command surfaces.

- Deep black background.
- Monospace text.
- Orange or faint white frame.
- Optional top bar with three status dots for prose code blocks.
- Snippets and response previews should scroll rather than wrap into unreadable layouts, except API response previews may use `pre-wrap` when readability is better.

### Navigation

Navigation is mono and uppercase.

- App sidebar active item uses orange leading accent and orange clipped frame.
- Docs top nav uses icon plus text, with text hidden on small screens when necessary.
- Docs sidebar groups are collapsible and show active state in orange.
- TOC active link uses orange and a subtle orange text glow.

### Network Visuals

Network visuals should be abstract, map-like, and diagnostic.

- Nodes are square, clipped, or small dot/square based.
- Routes, packets, rails, hops, matrices, and topology lines are preferred motifs.
- Active packets and important paths use orange.
- Green only means online/success.
- Scanline and route animations are acceptable when slow and subtle.
- Three.js scenes and animated CSS visuals must respect reduced motion.

## 7. Cut-Corner Frames

Cut corners are a core identity element. Use the shared `.ns-cut-frame` utility whenever possible.

### Standard Utility

```css
.frame {
	--ns-frame-cut: var(--ns-cut-sm);
	--ns-frame-color: var(--ns-border);
	--ns-frame-border-width: 1px;
}
```

Then apply `className="ns-cut-frame"` or the equivalent class composition.

The shared utility uses this polygon:

```css
clip-path: polygon(var(--ns-frame-cut) 0, 100% 0, 100% calc(100% - var(--ns-frame-cut)), calc(100% - var(--ns-frame-cut)) 100%, 0 100%, 0 var(--ns-frame-cut));
```

### Why Patches Exist

A regular `border: 1px` does not draw the diagonal cut edges of a clipped rectangle. The shared `.ns-cut-frame` patches those diagonal strokes with `::before` and `::after`.

```css
.ns-cut-frame::before,
.ns-cut-frame::after {
	width: calc(var(--ns-frame-cut) * sqrt(2));
	height: var(--ns-frame-border-width);
	background: var(--ns-frame-color);
}
```

When building a custom clipped frame:

- Prefer `.ns-cut-frame`.
- If a custom implementation is necessary, patch diagonal strokes or use a mask outline.
- Use `--ns-frame-color` and `--ns-frame-cut`; do not hardcode repeated frame values.
- Use `position: relative`.
- Use `isolation: isolate` when pseudo-elements, shadows, and children overlap.
- Be careful with `overflow: hidden`; do not clip visible focus outlines.

Never leave clipped borders visually open.

## 8. Motion

Motion should be restrained, mechanical, and diagnostic.

Use:

- `--ns-transition` for most UI state changes.
- `translateY(-0.0625rem)` for small hover lift.
- `180ms-280ms` for drawer/search/modal entry and exit.
- Slow scanlines, orbit movement, node blinking, packet routing, and topology reveals on landing visuals.
- GSAP/Three.js only where the page needs browser effects, currently the landing page.

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

- Primary series: `--ns-accent`, `#ff7a1a`, or `#ff8f3d`.
- Secondary/baseline series: `--ns-metal`, `--ns-slate-line`, or low-opacity white.
- Success state: `--ns-success`.
- Warning state: `--ns-warning`.
- Critical state: `--ns-critical`.
- Tooltip background: near black, around `rgba(10, 13, 18, 0.92)`.
- Tooltip border: faint white or orange when active.
- Axis labels: muted mono, around `10px`.
- Grid lines: very low contrast.
- Area fills: fade to transparent.

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
- Use `--ns-*` tokens for colors, fonts, cuts, shadows, and transitions.
- Use `class:list` in Astro and local `classNames` helpers in React when composing `ns-cut-frame`.
- Keep page sections full-width/unframed unless they are real tools, panels, repeated cards, or docs content blocks.
- Avoid nested decorative cards.
- Avoid new radius values for core UI.

### Assets And Icons

- Use `@netstamp/brand` assets for brand marks and favicon.
- Docs currently use Phosphor web components and `docs/src/components/icons/Icon.astro`.
- Use actual product/network visuals or generated technical scenes for landing visuals.
- Do not add stock-like atmospheric imagery when a diagnostic visual would communicate better.

## 13. Anti-Patterns

Do not add:

- Large rounded cards, pill-shaped core controls, or soft consumer-app surfaces.
- Pastel SaaS gradients, glossy blobs, heavy glassmorphism, or decorative orbs.
- Blue/purple/rainbow brand accents.
- Decorative emoji icons.
- Soft lifestyle photography or stock photos.
- Centered generic landing sections that ignore the engineering grid.
- Heavy shadows that overpower the frame system.
- Global CSS systems outside existing `--ns-*` tokens and CSS modules.
- Copy that sounds like a generic productivity app.

## 14. Review Checklist

Before shipping frontend UI, check:

- Does it use existing `--ns-*` tokens?
- Does it use the correct font family for display, body, and operational text?
- Are clipped frames using `.ns-cut-frame` or patched diagonal borders?
- Is orange the main interactive accent?
- Are green, yellow, and red reserved for state?
- Does the layout follow the `1rem` grid rhythm unless a larger landing/docs rhythm is intentional?
- Does the UI match the density of its surface: landing, app, docs, or OpenAPI?
- Does it collapse cleanly on the relevant breakpoints?
- Are focus states visible and not clipped?
- Are animations disabled or reduced for `prefers-reduced-motion`?
- Does copy name concrete network behavior instead of generic product benefits?
- If the change affects shared conventions, should this `design.md` or an area `AGENTS.md` be updated?
