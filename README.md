<div align="center">
  <img src="./packages/brand/assets/netstamp-logo.svg" alt="Netstamp" width="360" />

  <h3>Self-hosted network observability from probes you control.</h3>

  <p>
    See latency, packet loss, routes, TCP reachability, and probe health from the networks that matter to you.
  </p>

  <p>
    <a href="./LICENSE"><img alt="License" src="https://img.shields.io/github/license/yorukot/netstamp?style=flat-square" /></a>
    <img alt="Docker" src="https://img.shields.io/badge/Docker-self--hosted-2496ED?style=flat-square&logo=docker&logoColor=white" />
    <img alt="Go" src="https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go&logoColor=white" />
    <img alt="React" src="https://img.shields.io/badge/React-19-61DAFB?style=flat-square&logo=react&logoColor=111111" />
  </p>
</div>

---

## What is Netstamp?

Netstamp is an open-source, self-hosted network monitoring app for people who need to know what the internet looks like from their own machines, regions, labs, edge nodes, and private infrastructure.

Most monitoring tells you if a service is up from somebody else's cloud. Netstamp lets you place probes where your users, servers, or networks actually are, then watch reachability, latency, packet loss, routes, and probe health from those real viewpoints.

## Why self-host it?

- Your network data stays on your infrastructure.
- Your probes can run from home labs, offices, VPS regions, edge nodes, or private networks.
- You can monitor internal services that public SaaS checks cannot reach.
- You get one place to compare network behavior across multiple locations.
- You can start with Docker Compose and grow into a more serious deployment later.

## What you can use it for

- Check if a service is reachable from multiple real locations.
- Compare latency and packet loss between regions, ISPs, or hosting providers.
- Detect route changes and unstable network paths.
- Track probe health so you know which viewpoints are still reporting.
- Send alerts to the channels your team already watches.
- Keep historical network measurements in PostgreSQL and TimescaleDB.

## Features

- Self-hosted controller with a React web app and Go API.
- Lightweight probe agents that poll assignments and submit results.
- Ping, TCP connect, and traceroute checks.
- Project workspaces with roles and scoped access.
- Label-based probe and check organization.
- Dashboards for latency, packet loss, route behavior, and probe status.
- Alert rules, incidents, and notification channels.
- Webhook, Discord, and Telegram notifications.
- Docker Compose deployment with built-in migrations.
- Generated OpenAPI contract for integrations.

## Quick Start

Run Netstamp with Docker Compose:

```bash
mkdir netstamp
cd netstamp
curl -O https://raw.githubusercontent.com/yorukot/netstamp/main/deployments/docker/compose.yaml
curl -O https://raw.githubusercontent.com/yorukot/netstamp/main/deployments/docker/example.env
cp example.env .env
docker compose up -d
```

Open Netstamp:

```text
http://localhost:3000
```

Before exposing Netstamp publicly, edit `.env`, replace every `change-me` value, set `APP_ENV=production`, pin `NETSTAMP_VERSION` to a release tag, and put Netstamp behind HTTPS with your reverse proxy of choice.

## First Setup

After the app is running:

1. Create your first account.
2. Create a project for the services or networks you want to watch.
3. Add a probe for each network viewpoint you care about.
4. Install or run the probe agent with its probe ID and secret.
5. Create checks for the hosts, ports, and routes you want to monitor.
6. Add alerts and notification channels once the first measurements are flowing.

## How it works

Netstamp has one controller and many probes.

```text
Your browser -> Netstamp web app -> Netstamp controller -> PostgreSQL / TimescaleDB
```

```text
Probe agent -> poll assignments -> run checks -> submit results -> alerts and dashboards
```

The controller stores projects, users, probes, checks, results, incidents, and notification settings. Probes run near the networks being measured and only need the controller URL, probe ID, and probe secret.

## Links

- Docker Compose: [`deployments/docker/compose.yaml`](./deployments/docker/compose.yaml)
- Example environment: [`deployments/docker/example.env`](./deployments/docker/example.env)
- Documentation source: [`docs/`](./docs/)
- API contract: [`api/`](./api/)
- Backend: [`server/`](./server/)
- Web app: [`web/`](./web/)

## Development

Netstamp is a pnpm workspace with a Go backend and React/Vite frontend. If you want to contribute, install dependencies with `pnpm install`, then use the root `Justfile` for local development, linting, testing, and builds.

```bash
just
```

## License

Netstamp is licensed under the [Apache License 2.0](./LICENSE).

### Contributors

**Thanks to all the contributors for making this project even greater!**

<a href="https://github.com/yorukot/netstamp/graphs/contributors">
  <img src="https://gthanks.yorukot.me/image?target=yorukot%2Fnetstamp" />
</a>

### Star History

**THANKS FOR All OF YOUR STARS!** Your stars are my motivation to keep updating!

<a href="https://star-history.com/#yorukot/netstamp&Timeline">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=yorukot/netstamp&type=Timeline&theme=dark" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=yorukot/netstamp&type=Timeline" />
   <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=yorukot/netstamp&type=Timeline" />
 </picture>
</a>

<div align="center">

## ༼ つ ◕_◕ ༽つ Please share.

</div>
