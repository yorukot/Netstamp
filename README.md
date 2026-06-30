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

Netstamp is an open-source, self-hosted network observability app for people who need to understand what the internet looks like from their own machines, regions, labs, edge nodes, private infrastructure, and real user-facing networks.

Most monitoring platforms tell you whether a service is up from somebody else's cloud. Netstamp lets you place probes where your users, servers, and networks actually are, then observe reachability, latency, packet loss, routes, uptime, certificates, probe health, and incidents from those real viewpoints.

## What you can use it for

- Monitor services from the networks, regions, ISPs, labs, and edge nodes you actually care about.
- Compare latency, packet loss, reachability, and route behavior across real-world viewpoints.
- Detect unstable paths, broken routes, degraded probes, failing checks, expired certificates, and abnormal network conditions.
- Understand whether an issue is global, regional, provider-specific, probe-specific, or target-specific.
- Build dashboards that summarize network health, probe status, check results, incidents, and historical trends.
- Send alerts to the notification channels your team already uses.
- Organize monitoring across projects, teams, probes, labels, dashboards, and scoped permissions.
- Keep historical measurements in PostgreSQL and TimescaleDB for debugging, reporting, and long-term visibility.
- Use APIs, API keys, OpenAPI, and integrations to connect Netstamp with your own tools and workflows.

## Features

- Self-hosted controller with lightweight probes that run from your own machines, regions, labs, edge nodes, and private networks.
- Real-world network checks including ping, TCP, traceroute, uptime, API payload, and TLS/SSL certificate monitoring.
- Visibility into reachability, latency, packet loss, route behavior, probe health, incidents, and historical trends.
- Project-based collaboration with users, invitations, roles, scoped permissions, and API keys.
- Flexible probe and check organization with labels, locations, assignment groups, and capability-aware targeting.
- Dashboards for network health, probe status, check results, incidents, charts, and public status views.
- Alert rules and incident tracking for degraded services, unstable routes, failed checks, and abnormal metrics.
- Notification integrations for Webhook, Discord, Telegram, Slack, LINE Bot, and Email.
- Result analysis by probe, target, check, and latest measurement.
- OpenAPI, health checks, metrics, root administration tools, and production-ready deployment documentation.

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
