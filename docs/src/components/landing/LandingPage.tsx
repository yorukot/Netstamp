import { Badge, Button, CodePreview, GlobalFooter, KeyValueRow, MetricTile, SpecCard, SpecLabel, Surface } from "@netstamp/ui";
import dashboardDark from "../../assets/homepage-dashboard-dark.png?url";
import { appUrl } from "../../lib/publicUrls";
import styles from "./LandingPage.module.css";

const githubUrl = "https://github.com/yorukot/netstamp";

const productStats = [
	{ label: "Probe fleet", value: "18", detail: "online", tone: "success" },
	{ label: "Latency p95", value: "42ms", detail: "healthy", tone: "success" },
	{ label: "Packet loss", value: "0.08%", detail: "watch", tone: "warning" },
	{ label: "Route diff", value: "2", detail: "changed", tone: "accent" }
] as const;

const productSections = [
	{
		kicker: "Fleet",
		title: "Probes you control",
		copy: "Install probes on VPS nodes, lab hosts, edge machines, or internal networks. Track heartbeat freshness, labels, and capability from one controller.",
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
		copy: "Compare probe and check pairs over time. Detect route hash changes, latency shifts, packet loss, and topology movement before symptoms reach users.",
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
] as const;

const routeHops = ["TPE", "IXP", "NRT", "SJC", "SFO"] as const;

interface LandingPageProps {
	appHref?: string;
}

function classNames(...classes: Array<string | false | null | undefined>) {
	return classes.filter(Boolean).join(" ");
}

export function LandingPage({ appHref = appUrl("/register") }: LandingPageProps) {
	return (
		<div className={styles.landing}>
			<main id="content">
				<section className={styles.hero}>
					<div className={styles.heroCopy}>
						<SpecLabel tone="primary">Open-source network observability</SpecLabel>
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
						<Surface tone="glass" frameSize="lg" padding="md" className={styles.heroSpec}>
							<KeyValueRow label="Controller" value="self-hosted" meta="Go API" tone="primary" />
							<KeyValueRow label="Probe work" value="ping, tcp, dns, traceroute" meta="scheduled" />
							<KeyValueRow label="Evidence" value="latency, loss, route hash, heartbeat" meta="stored" />
						</Surface>
					</div>

					<figure className={styles.productShot} aria-label="Netstamp dashboard showing probe fleet metrics, route topology, and alert state">
						<div className={styles.shotHeader}>
							<SpecLabel>Product snapshot</SpecLabel>
							<Badge tone="success">live dashboard</Badge>
						</div>
						<img src={dashboardDark} alt="" width="1440" height="960" loading="eager" decoding="async" aria-hidden="true" />
					</figure>
				</section>

				<section className={styles.telemetryStrip} aria-label="Product telemetry snapshot">
					{productStats.map(stat => (
						<MetricTile key={stat.label} label={stat.label} value={stat.value} detail={stat.detail} tone={stat.tone} />
					))}
				</section>

				<section className={styles.productBand}>
					<div className={styles.sectionHeader}>
						<SpecLabel tone="secondary">Product surface</SpecLabel>
						<h2>Designed for repeated network operations.</h2>
						<p>Netstamp keeps the interface close to the work: fleet state, check definitions, result insight, alert routing, status pages, and API automation.</p>
					</div>

					<div className={styles.productGrid}>
						{productSections.map(section => (
							<SpecCard key={section.kicker} eyebrow={section.kicker} title={section.title} description={section.copy} active={section.kicker === "Fleet"}>
								<ul className={styles.pointList}>
									{section.points.map(point => (
										<li key={point}>{point}</li>
									))}
								</ul>
							</SpecCard>
						))}
					</div>
				</section>

				<section className={styles.routeBand}>
					<div className={styles.routeCopy}>
						<SpecLabel tone="primary">Route intelligence</SpecLabel>
						<h2>See when the path changes.</h2>
						<p>Traceroute runs become route timelines and topology views. Operators can compare the path hash, hop changes, and latency movement across probe locations.</p>
						<div className={styles.routeFacts}>
							<KeyValueRow label="Path hash" value="b94c.22f9.changed" meta="diff" tone="warning" />
							<KeyValueRow label="Hop shift" value="provider changed at hop 3" meta="NRT" />
							<KeyValueRow label="Latency" value="+18ms p95" meta="watch" tone="primary" />
						</div>
					</div>
					<div className={styles.routePanel}>
						<div className={styles.routeRail} aria-label="Example route from Taipei to San Francisco">
							{routeHops.map((hop, index) => (
								<span className={classNames(styles.routeNode, index === 2 && styles.routeNodeActive)} key={hop}>
									{hop}
								</span>
							))}
						</div>
						<CodePreview title="route diff" meta="trace">
							{`hop 03 changed: 203.69.35.1 -> 203.69.35.9
p95 latency +18ms
path hash b94c.22f9.changed`}
						</CodePreview>
					</div>
				</section>

				<section className={styles.ctaBand}>
					<div>
						<SpecLabel tone="secondary">Self-hosted controller</SpecLabel>
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
