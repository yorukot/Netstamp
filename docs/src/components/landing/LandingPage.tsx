import { Badge, Button, GlobalFooter } from "@netstamp/ui";
import dashboardDesktop from "../../assets/homepage-dashboard-desktop.svg?url";
import dashboardMobile from "../../assets/homepage-dashboard-mobile.svg?url";
import { appUrl } from "../../lib/publicUrls";
import styles from "./LandingPage.module.css";

const githubUrl = "https://github.com/yorukot/netstamp";

const productStats = [
	{ label: "Probe fleet", value: "18 online" },
	{ label: "Latency p95", value: "42ms" },
	{ label: "Packet loss", value: "0.08%" },
	{ label: "Route diff", value: "2 changed" }
];

const productSections = [
	{
		kicker: "Fleet",
		title: "Probes you control",
		copy: "Install probes on VPS nodes, lab hosts, edge machines, or internal networks. Netstamp shows where each agent is running, when it last checked in, and what it can measure.",
		points: ["Heartbeat status", "Location and labels", "IPv4 / IPv6 capability"]
	},
	{
		kicker: "Checks",
		title: "Latency, loss, DNS, and routes",
		copy: "Define repeatable checks once and assign them by labels. The controller schedules probe work and stores results for operational review.",
		points: ["Ping and TCP checks", "Traceroute runs", "Probe selectors"]
	},
	{
		kicker: "Insight",
		title: "Route intelligence for real paths",
		copy: "Compare probe and check pairs over time. Detect route hash changes, latency shifts, packet loss, and topology movement before users report symptoms.",
		points: ["Scope by probe or check", "Time range URLs", "Route topology views"]
	},
	{
		kicker: "Alerts",
		title: "Incidents routed to operators",
		copy: "Create threshold rules for measurable network behavior and send notifications to the tools your operators already watch.",
		points: ["Metric thresholds", "Cooldown windows", "Webhook and chat targets"]
	},
	{
		kicker: "API",
		title: "Automation without mystery",
		copy: "Use the generated OpenAPI contract to inspect controller routes, build scripts, and keep integrations aligned with backend behavior.",
		points: ["Generated contract", "Request console", "Typed web client"]
	},
	{
		kicker: "Open Source",
		title: "Deploy it where trust matters",
		copy: "Netstamp is built in the open so teams, researchers, and communities can inspect how measurements are collected and represented.",
		points: ["Self-hosted controller", "Portable probes", "Public docs"]
	}
];

interface LandingPageProps {
	appHref?: string;
}

function classNames(...classes: Array<string | false | null | undefined>) {
	return classes.filter(Boolean).join(" ");
}

export function LandingPage({ appHref = appUrl("/register") }: LandingPageProps) {
	return (
		<div className={styles.landing}>
			<main>
				<section className={styles.hero}>
					<div className={styles.heroCopy}>
						<Badge tone="accent" dot={false}>
							Open-source network observability
						</Badge>
						<h1>Netstamp network observability</h1>
						<p>Probes you control measure latency, packet loss, DNS, and routes from the networks that matter to you.</p>
						<div className={styles.heroActions}>
							<Button size="xl" asChild>
								<a href={appHref}>
									<ph-rocket-launch size={20} weight="bold" aria-hidden="true" />
									Deploy a probe
								</a>
							</Button>
							<Button size="xl" variant="secondary" asChild>
								<a href={githubUrl} target="_blank" rel="noreferrer">
									<ph-github-logo size={20} weight="bold" aria-hidden="true" />
									View GitHub
								</a>
							</Button>
						</div>
					</div>

					<figure className={styles.productShot}>
						<picture>
							<source media="(max-width: 48rem)" srcSet={dashboardMobile} width="720" height="960" />
							<img src={dashboardDesktop} alt="Netstamp dashboard showing probe fleet metrics, route topology, and alert state" width="1440" height="960" loading="eager" decoding="async" />
						</picture>
					</figure>
				</section>

				<section className={styles.telemetryStrip} aria-label="Product telemetry snapshot">
					{productStats.map(stat => (
						<div className={styles.telemetryItem} key={stat.label}>
							<span>{stat.label}</span>
							<strong>{stat.value}</strong>
						</div>
					))}
				</section>

				<section className={styles.productBand}>
					<div className={styles.sectionHeader}>
						<Badge tone="neutral" dot={false}>
							Product surface
						</Badge>
						<h2>Designed for repeated network operations.</h2>
						<p>Netstamp keeps the interface close to the work: fleet state, check definitions, result insight, alert routing, and API automation.</p>
					</div>

					<div className={styles.productGrid}>
						{productSections.map(section => (
							<article className={styles.productCard} key={section.kicker}>
								<div className={styles.cardHeader}>
									<span>{section.kicker}</span>
									<i aria-hidden="true" />
								</div>
								<h3>{section.title}</h3>
								<p>{section.copy}</p>
								<ul>
									{section.points.map(point => (
										<li key={point}>{point}</li>
									))}
								</ul>
							</article>
						))}
					</div>
				</section>

				<section className={styles.routeBand}>
					<div className={styles.routeCopy}>
						<Badge tone="accent" dot={false}>
							Route intelligence
						</Badge>
						<h2>See when the path changes.</h2>
						<p>Traceroute runs become route timelines and topology views. Operators can compare the path hash, hop changes, and latency movement across probe locations.</p>
					</div>
					<div className={styles.routePanel} aria-hidden="true">
						<div className={styles.routeRail}>
							{["TPE", "IXP", "NRT", "SJC", "SFO"].map((hop, index) => (
								<span className={classNames(styles.routeNode, index === 2 && styles.routeNodeActive)} key={hop}>
									{hop}
								</span>
							))}
						</div>
						<div className={styles.routeRows}>
							<span>path hash changed</span>
							<span>hop 3 provider shift</span>
							<span>p95 +18ms</span>
						</div>
					</div>
				</section>

				<section className={styles.ctaBand}>
					<div>
						<Badge tone="neutral" dot={false}>
							Self-hosted controller
						</Badge>
						<h2>Run measurements from networks you own.</h2>
						<p>Deploy Netstamp, register probes, and start collecting measurable evidence about the paths your traffic actually takes.</p>
					</div>
					<div className={styles.ctaActions}>
						<Button size="xl" asChild>
							<a href={appHref}>
								<ph-rocket-launch size={20} weight="bold" aria-hidden="true" />
								Deploy a probe
							</a>
						</Button>
						<Button size="xl" variant="outline" asChild>
							<a href="/docs/">
								<ph-book-open size={20} weight="bold" aria-hidden="true" />
								Read docs
							</a>
						</Button>
					</div>
				</section>
			</main>
			<GlobalFooter />
		</div>
	);
}
