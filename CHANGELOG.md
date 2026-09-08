# ChangeLog

All notable changes to this project will be documented in this file. Dates are displayed in UTC (YYYY-MM-DD).

# [**v0.1.0**](https://github.com/yorukot/netstamp/releases/tag/v0.1.0)

> 2026-09-08

#### Update

- Promote Netstamp to its first product release. The `v0.0.x` tags were distribution and listing tests; `v0.1.0` is the first version intended for self-hosted use.

#### Optimization

- Decouple the minimum supported probe agent version from the product version so a release no longer forces every probe to update. The probe runtime contract is unchanged since `v0.0.0`, so agents from any earlier release keep running against this controller.

#### Misc

- Pin the backend lint toolchain to `go1.26.2` so the bundled staticcheck matches the Go release the module targets.

# [**v0.0.2**](https://github.com/yorukot/netstamp/releases/tag/v0.0.2)

> 2026-09-05

#### Update

- Switch the map basemap to OpenFreeMap and refine the sidebar update indicator.

#### Fix

- Fix the MapLibre worker bundling under Vite.
- Bound the traceroute sampled-run refresh job's hop join to its refresh window and stop rewriting unchanged buckets; the unbounded join hashed the whole hop hypertable every minute and spilled terabytes of temp files per day.

#### Documentation

- Refresh the homepage copy, add a deployment summary, drop the three.js hero animation, and hand the quick start off to a first two-probe comparison.

#### Misc

- Prepare `v0.0.2` as a distribution test ahead of the first product release, `v0.1.0`.

# [**v0.0.1**](https://github.com/yorukot/netstamp/releases/tag/v0.0.1)

> 2026-08-21

#### Update

- Add automatic GitHub release checks and administrator controls for update-check settings and status.

#### Optimization

- Remove the deprecated administrator JSON import and export paths.
- Refresh workspace and Go dependencies, remove dead backend helpers, and move shared version handling under the platform package.

#### Documentation

- Add contribution guidance, align installation documentation with the release artifact names, and remove obsolete internal planning documents.

#### Fix

- Restore HTTP result service wiring in the TimescaleDB integration harness and fix its workflow formatting.

#### Misc

- Prepare `v0.0.1` as the final distribution test before the first product release, `v0.1.0`.

# [**v0.0.0**](https://github.com/yorukot/netstamp/releases/tag/v0.0.0)

> 2026-08-19

#### Update

- Introduce the self-hosted Netstamp controller, React web application, and lightweight Linux probes.
- Add Ping, TCP connect, HTTP/HTTPS, Traceroute, and TLS certificate monitoring from probes you control.
- Add projects, members, labels, probe selectors, check assignments, scoped permissions, and personal API tokens.
- Add dashboards and Insight views for reachability, latency, packet loss, HTTP timings, certificates, probe health, incidents, and historical trends.
- Add alert rules, incident tracking, public status pages, and Webhook, Discord, Telegram, Slack, and Email notifications.

#### Optimization

- Centralize the product, API, and minimum probe version contract across the controller and probe runtime.
- Harden self-hosted Docker Compose assets, release images, and probe installation metadata for reproducible deployments.
- Simplify analytics consent and remove deprecated runtime configuration from the web and documentation builds.

#### Documentation

- Publish the quick start, installation, configuration, operations, security, and task-oriented product guides.
- Publish the generated OpenAPI reference and shared component Storybook alongside the documentation site.

#### Misc

- Prepare `v0.0.0` as the final distribution and listing test before the first product release, `v0.1.0`.
