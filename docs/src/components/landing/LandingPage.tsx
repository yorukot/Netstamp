import { Badge, Button, GlobalFooter } from "@netstamp/ui";
import { gsap } from "gsap";
import { ScrollTrigger } from "gsap/ScrollTrigger";
import { useEffect, useRef } from "react";
import taiwanSubmarineCablesMap from "../../assets/taiwan_submarine_cables.svg?url";
import { GlobalNetworkAnimation } from "./GlobalNetworkAnimation";
import styles from "./LandingPage.module.css";
import { NetworkScene } from "./NetworkScene";
import { ProbeScene } from "./ProbeScene";

gsap.registerPlugin(ScrollTrigger);

const githubUrl = "https://github.com/yorukot/netstamp";

const checkCards = [
	{ name: "Ping", metric: "p95 42ms", detail: "ICMP / TCP probes" },
	{ name: "DNS", metric: "NOERROR", detail: "resolver + authority" },
	{ name: "Traceroute", metric: "14 hops", detail: "route hash diff" }
];

const routeHops = ["AMS", "FRA", "IXP", "NYC", "SFO"];

const routeSignals = ["See latency.", "See packet loss.", "See DNS failures.", "See path changes.", "See where traffic takes the long way around."];

interface LandingPageProps {
	appHref?: string;
}

function classNames(...classes: Array<string | false | null | undefined>) {
	return classes.filter(Boolean).join(" ");
}

export function LandingPage({ appHref = "/register" }: LandingPageProps) {
	const landingRef = useRef<HTMLDivElement>(null);

	useEffect(() => {
		const image = new Image();
		image.decoding = "async";
		image.src = taiwanSubmarineCablesMap;
		if (image.decode) void image.decode().catch(() => undefined);
	}, []);

	useEffect(() => {
		const reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
		if (reduced) return;

		const ctx = gsap.context(() => {
			// Story section
			gsap.from("[data-gs='story']", {
				opacity: 0,
				y: 48,
				duration: 1.0,
				ease: "power3.out",
				scrollTrigger: {
					trigger: "[data-gs='story']",
					start: "top 80%"
				}
			});

			// Feature label
			gsap.from("[data-gs='feature-label']", {
				opacity: 0,
				y: -10,
				duration: 0.65,
				ease: "power2.out",
				scrollTrigger: {
					trigger: "[data-gs='feature-label']",
					start: "top 82%"
				}
			});

			// Feature cards stagger
			const cards = gsap.utils.toArray<Element>("[data-gs='feature-card']");
			if (cards.length) {
				gsap.from(cards, {
					opacity: 0,
					y: 64,
					duration: 0.8,
					ease: "power3.out",
					stagger: 0.14,
					scrollTrigger: {
						trigger: cards[0],
						start: "top 80%"
					}
				});
			}

			// Trust section
			gsap.from("[data-gs='trust']", {
				opacity: 0,
				y: 48,
				duration: 1.0,
				ease: "power3.out",
				scrollTrigger: {
					trigger: "[data-gs='trust']",
					start: "top 80%"
				}
			});
		}, landingRef);

		return () => ctx.revert();
	}, []);

	return (
		<div ref={landingRef} className={styles.landing}>
			<main>
				{/* Hero — unchanged */}
				<section className={styles.hero}>
					<GlobalNetworkAnimation />

					<div className={styles.heroCopy}>
						<h1>
							See the network.
							<span>Before it fails you.</span>
						</h1>
						<p>Open-source network observability from probes you control.</p>
						<p>Measure latency, packet loss, DNS, and routes.</p>

						<div className={styles.heroActions}>
							<Button size="xl" asChild>
								<a href={appHref}>
									<ph-rocket-launch size={20} weight="bold" aria-hidden="true" />
									Deploy Your Probe
								</a>
							</Button>
							<Button size="xl" variant="secondary" asChild>
								<a href={githubUrl} target="_blank" rel="noreferrer">
									<ph-github-logo size={20} weight="bold" aria-hidden="true" />
									View on GitHub
								</a>
							</Button>
						</div>
					</div>
				</section>

				{/* Story Section — redesigned with Three.js */}
				<section data-gs="story" className={styles.storySection}>
					<div className={styles.storyCopy}>
						<Badge tone="neutral">Path intelligence</Badge>
						<h2>
							Your traffic has a story.
							<br />
							Netstamp shows the path.
						</h2>
						<p>Traffic does not move through magic.</p>
						<p>It crosses cables, providers, exchanges, policies, failures, and cost decisions.</p>
						<p>Netstamp helps communities, operators, and builders understand the real paths their traffic takes.</p>
					</div>
					<div className={styles.storyViz}>
						<NetworkScene />
						<div className={styles.storyVizLabel} aria-hidden="true">
							<span className={styles.storyVizDot} />
							<span>live network topology</span>
						</div>
					</div>
				</section>

				{/* Feature Stack — redesigned */}
				<section className={styles.featureStack}>
					<div className={styles.featureHeader}>
						<p data-gs="feature-label" className={styles.featureLabel}>
							What Netstamp measures
						</p>
						<div className={styles.featureHeaderRule} aria-hidden="true" />
					</div>

					<article data-gs="feature-card" className={classNames("ns-cut-frame", styles.featureCard, styles.probeFeatureCard)}>
						<div className={styles.featureCardMain}>
							<div className={classNames("ns-cut-frame", styles.cardIcon)} aria-hidden="true">
								<ph-globe-hemisphere-west size={24} weight="duotone" />
							</div>
							<h2>Probes everywhere.</h2>
							<p>Install Netstamp probes on VPS nodes, servers, internal hosts, edge locations, classrooms, labs, or community networks.</p>
							<p>Each probe measures the Internet from its own point of view.</p>
						</div>
						<div className={styles.probeSceneCol} aria-hidden="true">
							<ProbeScene />
						</div>
						<span className={styles.featureBadge} aria-hidden="true">
							01
						</span>
					</article>

					<article data-gs="feature-card" className={classNames("ns-cut-frame", styles.featureCard)}>
						<div className={styles.featureCardTopline}>
							<div className={classNames("ns-cut-frame", styles.cardIcon)} aria-hidden="true">
								<ph-pulse size={24} weight="duotone" />
							</div>
							<span>scheduler / result stream</span>
						</div>
						<h2>Checks that matter.</h2>
						<p>Simple tools. Structured results. Historical visibility.</p>
						<div className={styles.checkGrid}>
							{checkCards.map(check => (
								<div className={classNames("ns-cut-frame", styles.checkCard)} key={check.name}>
									<div className={styles.checkCardHeader}>
										<strong>{check.name}</strong>
										<span>{check.metric}</span>
									</div>
									<div className={styles.checkPacketRail} aria-hidden="true">
										<i />
										<i />
										<i />
										<i />
									</div>
									<small>{check.detail}</small>
								</div>
							))}
						</div>
						<div className={styles.checkTelemetry} aria-hidden="true">
							<span>interval 30s</span>
							<span>jitter +/-4s</span>
							<span>retention 30d</span>
						</div>
						<span className={styles.featureBadge} aria-hidden="true">
							02
						</span>
					</article>

					<article data-gs="feature-card" className={classNames("ns-cut-frame", styles.featureCard)}>
						<div className={styles.featureCardTopline}>
							<div className={classNames("ns-cut-frame", styles.cardIcon)} aria-hidden="true">
								<ph-network size={24} weight="duotone" />
							</div>
							<span>path hash / hop timeline</span>
						</div>
						<h2>Routes you can compare.</h2>
						<div className={styles.routeBoard} aria-hidden="true">
							<div className={styles.routeRail}>
								{routeHops.map(hop => (
									<span className={styles.routeNode} key={hop}>
										<span className={styles.routeNodeLabel}>{hop}</span>
									</span>
								))}
							</div>
							<div className={styles.routeMetrics}>
								<span>delta +18ms</span>
								<span>hash changed</span>
								<span>hop 9 reroute</span>
							</div>
						</div>
						<ul className={styles.signalList}>
							{routeSignals.map(signal => (
								<li key={signal}>
									<ph-check-circle size={16} weight="fill" aria-hidden="true" />
									<span>{signal}</span>
								</li>
							))}
						</ul>
						<span className={styles.featureBadge} aria-hidden="true">
							03
						</span>
					</article>
				</section>

				{/* Trust / CTA Section — redesigned */}
				<section data-gs="trust" className={styles.trustSection}>
					<div className={styles.trustInner}>
						<div className={styles.trustLeft}>
							<Badge tone="accent">Open source</Badge>
							<h2>
								Open source.
								<br />
								Because trust needs visibility.
							</h2>
							<p>Netstamp is built in the open — for operators, researchers, students, communities, and anyone who wants to understand how the Internet actually behaves.</p>
							<p>Gives communities a way to measure, prove, and discuss what is happening.</p>
							<div className={styles.ctaActions}>
								<Button size="xl" asChild>
									<a href={appHref}>
										<ph-rocket-launch size={20} weight="bold" aria-hidden="true" />
										Deploy Your Probe
									</a>
								</Button>
								<Button size="xl" variant="outline" asChild>
									<a href={githubUrl} target="_blank" rel="noreferrer">
										<ph-arrow-up-right size={20} weight="bold" aria-hidden="true" />
										View the source
									</a>
								</Button>
							</div>
						</div>

						<div className={styles.trustRight}>
							<div className={styles.trustMapBackdrop}>
								<img className={styles.trustMapImage} src={taiwanSubmarineCablesMap} alt="Map of Taiwan submarine cable routes" loading="eager" fetchPriority="low" decoding="async" />
							</div>
						</div>
					</div>
				</section>
			</main>

			<GlobalFooter />
		</div>
	);
}
