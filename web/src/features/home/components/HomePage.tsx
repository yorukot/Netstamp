import { pathForRoute } from "@/routes/routePaths";
import taiwanSubmarineCablesMap from "@/shared/assets/taiwan_submarine_cables.svg?url";
import { docsUrl } from "@/shared/config/publicLinks";
import netstampLogo from "@netstamp/brand/assets/netstamp-logo-dark.svg";
import { Button, GlobalFooter } from "@netstamp/ui";
import { ArrowUpRight, CheckCircle, GitBranch, GithubLogo, GlobeHemisphereWest, Pulse, RocketLaunch, Users } from "@phosphor-icons/react";
import { useLayoutEffect, type CSSProperties } from "react";
import { Helmet } from "react-helmet-async";
import { Link } from "react-router-dom";
import styles from "./HomePage.module.css";

const githubUrl = "https://github.com/yorukot/netstamp";
const salesHref = "mailto:sales@netstamp.dev?subject=Netstamp%20demo";

const proofMetrics = [
	{ value: "30s", label: "probe intervals for live checks" },
	{ value: "3d", label: "raw telemetry retention by default" },
	{ value: "4", label: "probe, check, route, and result views" },
	{ value: "OSS", label: "self-hostable controller and probes" }
];

const buyerCards = [
	{
		audience: "Product",
		title: "Connect network quality to customer impact.",
		copy: "Turn route changes, latency spikes, and packet loss into plain evidence for roadmap tradeoffs and launch readiness."
	},
	{
		audience: "Sales",
		title: "Show proof before the customer asks.",
		copy: "Bring clean probe history into enterprise conversations when prospects ask about regions, uptime, or connectivity risk."
	},
	{
		audience: "Operations",
		title: "Find the path that changed.",
		copy: "Compare probes, checks, and route hashes so the team can separate app incidents from network behavior."
	}
];

const workflowSteps = [
	{ step: "01", title: "Deploy probes", copy: "Place lightweight agents in the regions, offices, edge networks, and customer paths that matter." },
	{ step: "02", title: "Measure continuously", copy: "Run ping, TCP, and traceroute checks on a schedule with consistent labels and project ownership." },
	{ step: "03", title: "Explain the result", copy: "Use dashboard evidence to answer what changed, where, and how often." }
];

const routeHops = ["TPE", "HKG", "SIN", "FRA", "AMS"];
const orbitNodes = [
	{ label: "TPE", x: 10, y: 48 },
	{ label: "HKG", x: 38, y: 12 },
	{ label: "SIN", x: 88, y: 62 },
	{ label: "FRA", x: 20, y: 78 },
	{ label: "AMS", x: 82, y: 26 }
];
const checks = [
	{ name: "Ping", metric: "p95 42ms", detail: "latency and loss" },
	{ name: "TCP", metric: "open 98%", detail: "service reachability" },
	{ name: "Trace", metric: "14 hops", detail: "path hash changes" }
];

function classNames(...classes: Array<string | false | null | undefined>) {
	return classes.filter(Boolean).join(" ");
}

function useLightHomeTheme() {
	useLayoutEffect(() => {
		const previousTheme = document.documentElement.dataset.theme;
		document.documentElement.dataset.theme = "light";

		return () => {
			if (previousTheme) {
				document.documentElement.dataset.theme = previousTheme;
			} else {
				delete document.documentElement.dataset.theme;
			}
		};
	}, []);
}

export function HomePage() {
	useLightHomeTheme();

	return (
		<div className={styles.home}>
			<Helmet>
				<title>Netstamp - Network clarity for product and sales teams</title>
				<meta name="description" content="Open-source network observability from probes you control. Measure latency, packet loss, TCP reachability, and route changes." />
			</Helmet>

			<header className={styles.nav}>
				<Link className={styles.brand} to={pathForRoute("landing")} aria-label="Netstamp home">
					<img src={netstampLogo} alt="Netstamp" />
				</Link>
				<nav className={styles.navLinks} aria-label="Home navigation">
					<a href="#product">Product</a>
					<a href="#use-cases">Use cases</a>
					<a href={docsUrl("/docs/")}>Docs</a>
					<a href="/api/v1/docs">API</a>
				</nav>
				<div className={styles.navActions}>
					<Button variant="ghost" asChild>
						<Link to={pathForRoute("login")}>Log in</Link>
					</Button>
					<Button asChild>
						<Link to={pathForRoute("register")}>Start measuring</Link>
					</Button>
				</div>
			</header>

			<main>
				<section className={styles.hero}>
					<div className={styles.heroCopy}>
						<span className={styles.eyebrow}>Open-source network observability</span>
						<h1>Network clarity your team can sell, ship, and defend.</h1>
						<p>Netstamp turns probes you control into customer-ready evidence for latency, packet loss, TCP reachability, and route changes.</p>
						<div className={styles.heroActions}>
							<Button size="xl" asChild>
								<Link to={pathForRoute("register")}>
									<RocketLaunch size={20} weight="bold" aria-hidden="true" />
									Start measuring
								</Link>
							</Button>
							<Button size="xl" variant="secondary" asChild>
								<a href={salesHref}>
									<Users size={20} weight="bold" aria-hidden="true" />
									Contact sales
								</a>
							</Button>
						</div>
						<div className={styles.heroProof} aria-label="Netstamp positioning">
							<span>Self-hostable</span>
							<span>Probe-owned data</span>
							<span>Built for network evidence</span>
						</div>
					</div>

					<div className={styles.heroVisual} aria-label="Netstamp product preview">
						<div className={styles.orbitScene} aria-hidden="true">
							<div className={styles.orbitCore} />
							<div className={styles.orbitA} />
							<div className={styles.orbitB} />
							<div className={styles.orbitC} />
							{orbitNodes.map(node => (
								<span className={styles.orbitNode} style={{ "--x": `${node.x}%`, "--y": `${node.y}%` } as CSSProperties} key={node.label}>
									{node.label}
								</span>
							))}
							<span className={styles.packetOne} />
							<span className={styles.packetTwo} />
						</div>
						<div className={classNames("ns-cut-frame", styles.dashboardMock)}>
							<div className={styles.mockHeader}>
								<div>
									<span>Project / APAC edge</span>
									<strong>Network health</strong>
								</div>
								<span className={styles.liveBadge}>Live</span>
							</div>
							<div className={styles.mockMetrics}>
								{proofMetrics.slice(0, 3).map(metric => (
									<div key={metric.value} className={styles.mockMetric}>
										<strong>{metric.value}</strong>
										<span>{metric.label}</span>
									</div>
								))}
							</div>
							<div className={styles.mockMap}>
								<div className={styles.mockRoute} aria-hidden="true">
									{routeHops.map(hop => (
										<span key={hop}>{hop}</span>
									))}
								</div>
							</div>
						</div>
					</div>
				</section>

				<section className={styles.proofBand} aria-label="Product proof">
					{proofMetrics.map(metric => (
						<div className={styles.proofMetric} key={metric.value}>
							<strong>{metric.value}</strong>
							<span>{metric.label}</span>
						</div>
					))}
				</section>

				<section className={styles.buyerSection} id="use-cases">
					<div className={styles.sectionHeader}>
						<span className={styles.eyebrow}>Why teams buy it</span>
						<h2>One measurement layer for product, sales, and operations.</h2>
						<p>Make the business case early: show who uses it, what decision it improves, and why the data is credible.</p>
					</div>
					<div className={styles.buyerGrid}>
						{buyerCards.map(card => (
							<article className={classNames("ns-cut-frame", styles.buyerCard)} key={card.audience}>
								<span>{card.audience}</span>
								<h3>{card.title}</h3>
								<p>{card.copy}</p>
							</article>
						))}
					</div>
				</section>

				<section className={styles.productSection} id="product">
					<div className={styles.productCopy}>
						<span className={styles.eyebrow}>Product tour</span>
						<h2>From probe to proof in three steps.</h2>
						<p>Install agents, collect comparable checks, and share the result with the people who need the answer.</p>
						<div className={styles.workflowList}>
							{workflowSteps.map(item => (
								<div className={styles.workflowStep} key={item.step}>
									<strong>{item.step}</strong>
									<div>
										<h3>{item.title}</h3>
										<p>{item.copy}</p>
									</div>
								</div>
							))}
						</div>
					</div>
					<div className={classNames("ns-cut-frame", styles.productVisual)}>
						<div className={styles.probeScene} aria-hidden="true">
							<span className={styles.probeHub} />
							{Array.from({ length: 10 }, (_, index) => (
								<span className={styles.probeNode} style={{ "--i": index } as CSSProperties} key={index} />
							))}
						</div>
						<div className={styles.productVisualOverlay}>
							<span>Probe fleet</span>
							<strong>12 active agents</strong>
						</div>
					</div>
				</section>

				<section className={styles.signalSection}>
					<div className={styles.sectionHeader}>
						<span className={styles.eyebrow}>What it measures</span>
						<h2>Evidence that survives the incident review.</h2>
					</div>
					<div className={styles.signalGrid}>
						<article className={classNames("ns-cut-frame", styles.signalCard)}>
							<div className={styles.cardTopline}>
								<Pulse size={22} weight="bold" aria-hidden="true" />
								<span>Scheduler / result stream</span>
							</div>
							<h3>Checks that explain customer experience.</h3>
							<div className={styles.checkGrid}>
								{checks.map(check => (
									<div className={classNames("ns-cut-frame", styles.checkCard)} key={check.name}>
										<strong>{check.name}</strong>
										<span>{check.metric}</span>
										<small>{check.detail}</small>
									</div>
								))}
							</div>
						</article>
						<article className={classNames("ns-cut-frame", styles.signalCard)}>
							<div className={styles.cardTopline}>
								<GitBranch size={22} weight="bold" aria-hidden="true" />
								<span>Path hash / hop timeline</span>
							</div>
							<h3>Routes your team can compare.</h3>
							<div className={styles.routeBoard} aria-hidden="true">
								<div className={styles.routeRail}>
									{routeHops.map(hop => (
										<span className={styles.routeNode} key={hop}>
											{hop}
										</span>
									))}
								</div>
								<div className={styles.routeMetrics}>
									<span>delta +18ms</span>
									<span>hash changed</span>
									<span>hop 9 reroute</span>
								</div>
							</div>
						</article>
					</div>
				</section>

				<section className={styles.mapSection}>
					<div className={styles.mapCopy}>
						<span className={styles.eyebrow}>Regional confidence</span>
						<h2>Tell a clearer story when the network is outside your app.</h2>
						<p>Use Netstamp to explain regional connectivity, submarine cable exposure, route churn, and edge reachability without asking every team to read raw traceroute output.</p>
						<div className={styles.heroActions}>
							<Button size="lg" asChild>
								<a href={docsUrl("/docs/")}>Read the docs</a>
							</Button>
							<Button size="lg" variant="outline" asChild>
								<a href={githubUrl} target="_blank" rel="noreferrer">
									<GithubLogo size={18} weight="bold" aria-hidden="true" />
									View source
								</a>
							</Button>
						</div>
					</div>
					<div className={styles.mapFrame}>
						<img src={taiwanSubmarineCablesMap} alt="" loading="lazy" decoding="async" />
					</div>
				</section>

				<section className={styles.finalCta}>
					<GlobeHemisphereWest size={72} weight="duotone" aria-hidden="true" />
					<span className={styles.eyebrow}>Ready for proof</span>
					<h2>Make network behavior visible before it becomes a renewal problem.</h2>
					<p>Start with a self-hosted controller, add probes where customers care, and give every team the same network facts.</p>
					<div className={styles.heroActions}>
						<Button size="xl" asChild>
							<Link to={pathForRoute("register")}>Start measuring</Link>
						</Button>
						<Button size="xl" variant="secondary" asChild>
							<a href={salesHref}>
								<ArrowUpRight size={20} weight="bold" aria-hidden="true" />
								Contact sales
							</a>
						</Button>
					</div>
					<ul className={styles.finalChecks}>
						<li>
							<CheckCircle size={18} weight="fill" aria-hidden="true" />
							Probe-owned data
						</li>
						<li>
							<CheckCircle size={18} weight="fill" aria-hidden="true" />
							Open-source controller
						</li>
						<li>
							<CheckCircle size={18} weight="fill" aria-hidden="true" />
							Dashboard-ready evidence
						</li>
					</ul>
				</section>
			</main>

			<GlobalFooter className={styles.footer} />
		</div>
	);
}
