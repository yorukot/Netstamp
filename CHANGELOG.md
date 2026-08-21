# ChangeLog

All notable changes to this project will be documented in this file. Dates are displayed in UTC (YYYY-MM-DD).

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
