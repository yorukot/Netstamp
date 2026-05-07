import { pathForRoute } from "@/routes/routePaths";
import type { Navigate } from "@/routes/routeTypes";
import { GlobalFooter } from "@/shared/components/GlobalFooter";
import { classNames } from "@/shared/utils/classNames";
import netstampLogo from "@netstamp/brand/assets/netstamp-logo-light.svg";
import { Badge, Button } from "@netstamp/ui";
import { ArrowUpRightIcon } from "@phosphor-icons/react/dist/csr/ArrowUpRight";
import { CheckCircleIcon } from "@phosphor-icons/react/dist/csr/CheckCircle";
import { GithubLogoIcon } from "@phosphor-icons/react/dist/csr/GithubLogo";
import { GlobeHemisphereWestIcon } from "@phosphor-icons/react/dist/csr/GlobeHemisphereWest";
import { NetworkIcon } from "@phosphor-icons/react/dist/csr/Network";
import { PulseIcon } from "@phosphor-icons/react/dist/csr/Pulse";
import { RocketLaunchIcon } from "@phosphor-icons/react/dist/csr/RocketLaunch";
import { gsap } from "gsap";
import { ScrollTrigger } from "gsap/ScrollTrigger";
import type { CSSProperties } from "react";
import { useEffect, useRef } from "react";
import { Helmet } from "react-helmet-async";
import { Link } from "react-router-dom";
import { BufferGeometry, Clock, Line, LineBasicMaterial, Mesh, MeshBasicMaterial, OctahedronGeometry, PerspectiveCamera, Scene, SphereGeometry, Vector3, WebGLRenderer } from "three";
import taiwanSubmarineCablesMap from "../../../../assets/taiwan_submarine_cables.svg";
import styles from "./LandingPage.module.css";

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
	navigate: Navigate;
}

export function LandingPage({ navigate }: LandingPageProps) {
	const landingRef = useRef<HTMLDivElement>(null);

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
			<Helmet>
				<title>Netstamp - See the network before it fails you</title>
				<meta name="description" content="Open-source network observability from probes you control. Measure latency, packet loss, DNS, and routes." />
			</Helmet>

			<nav className={styles.nav}>
				<Link className={styles.brand} to={pathForRoute("landing")}>
					<img className={styles.brandLogo} src={netstampLogo} alt="Netstamp" />
				</Link>

				<div className={styles.navActions}>
					<a className={styles.navLink} href={githubUrl} target="_blank" rel="noreferrer">
						<GithubLogoIcon size={16} weight="bold" aria-hidden="true" />
						<span>GitHub</span>
					</a>
					<Button size="md" onClick={() => navigate("register")}>
						<RocketLaunchIcon size={16} weight="bold" aria-hidden="true" />
						<span className={styles.navCtaText}>Deploy</span>
					</Button>
				</div>
			</nav>

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
							<Button size="xl" onClick={() => navigate("register")}>
								<RocketLaunchIcon size={20} weight="bold" aria-hidden="true" />
								Deploy Your Probe
							</Button>
							<Button size="xl" variant="secondary" asChild>
								<a href={githubUrl} target="_blank" rel="noreferrer">
									<GithubLogoIcon size={20} weight="bold" aria-hidden="true" />
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
								<GlobeHemisphereWestIcon size={24} weight="duotone" />
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
								<PulseIcon size={24} weight="duotone" />
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
								<NetworkIcon size={24} weight="duotone" />
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
									<CheckCircleIcon size={16} weight="fill" aria-hidden="true" />
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
								<Button size="xl" onClick={() => navigate("register")}>
									<RocketLaunchIcon size={20} weight="bold" aria-hidden="true" />
									Deploy Your Probe
								</Button>
								<Button size="xl" variant="outline" asChild>
									<a href={githubUrl} target="_blank" rel="noreferrer">
										<ArrowUpRightIcon size={20} weight="bold" aria-hidden="true" />
										View the source
									</a>
								</Button>
							</div>
						</div>

						<div className={styles.trustRight}>
							<div className={styles.trustMapBackdrop}>
								<img className={styles.trustMapImage} src={taiwanSubmarineCablesMap} alt="Map of Taiwan submarine cable routes" loading="lazy" decoding="async" />
							</div>
						</div>
					</div>
				</section>
			</main>

			<GlobalFooter />
		</div>
	);
}

// ── Three.js Network Scene ───────────────────────────────────────────────────

function NetworkScene() {
	const mountRef = useRef<HTMLDivElement>(null);

	useEffect(() => {
		const container = mountRef.current;
		if (!container) return;

		const reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

		const w = container.clientWidth || 480;
		const h = container.clientHeight || 480;

		const scene = new Scene();
		const camera = new PerspectiveCamera(52, w / h, 0.1, 100);
		camera.position.z = 6.5;

		const renderer = new WebGLRenderer({ alpha: true, antialias: true });
		renderer.setSize(w, h);
		renderer.setPixelRatio(Math.min(window.devicePixelRatio, 1.5));
		renderer.setClearColor(0x000000, 0);
		container.appendChild(renderer.domElement);

		// Distributed network topology — 10 nodes
		const positions: [number, number, number][] = [
			[-1.9, 1.1, 0.3], // 0 ams
			[-0.7, 1.9, -0.3], // 1 fra
			[0.6, 1.6, 0.5], // 2 lon
			[2.0, 0.7, -0.2], // 3 nyc
			[2.5, -0.5, 0.4], // 4 sfo
			[1.3, -1.7, 0.0], // 5 sin
			[-0.1, -2.0, 0.6], // 6 tok
			[-1.7, -0.9, -0.3], // 7 sgp
			[-2.4, 0.1, 0.5], // 8 cdg
			[0.2, 0.1, -0.9] // 9 IX hub
		];

		const nodeGeo = new OctahedronGeometry(0.072, 0);
		const nodes = positions.map(([x, y, z]) => {
			const mesh = new Mesh(nodeGeo, new MeshBasicMaterial({ color: 0xff7a1a }));
			mesh.position.set(x, y, z);
			scene.add(mesh);
			return mesh;
		});

		// Connections — [from, to, opacity]
		const connDefs: [number, number, number][] = [
			[0, 1, 0.5],
			[1, 2, 0.45],
			[2, 3, 0.5],
			[3, 4, 0.45],
			[4, 5, 0.35],
			[5, 6, 0.4],
			[6, 7, 0.35],
			[7, 8, 0.45],
			[8, 0, 0.4],
			[9, 1, 0.6],
			[9, 3, 0.55],
			[9, 5, 0.5],
			[9, 7, 0.5],
			[2, 9, 0.55],
			[4, 9, 0.5]
		];

		connDefs.forEach(([a, b, opacity]) => {
			const geo = new BufferGeometry().setFromPoints([nodes[a].position.clone(), nodes[b].position.clone()]);
			scene.add(new Line(geo, new LineBasicMaterial({ color: 0xff7a1a, transparent: true, opacity })));
		});

		// Animated packets — small spheres traveling along connections
		const pktGeo = new SphereGeometry(0.05, 8, 6);
		const packets = connDefs.slice(0, 8).map(([a, b], i) => {
			const mesh = new Mesh(pktGeo, new MeshBasicMaterial({ color: 0xffaa55 }));
			scene.add(mesh);
			return { mesh, from: a, to: b, t: i / 8 };
		});

		let raf: number;
		const clock = new Clock();

		function animate() {
			raf = requestAnimationFrame(animate);
			const elapsed = clock.getElapsedTime();

			if (!reduced) {
				scene.rotation.y = elapsed * 0.07;
				scene.rotation.x = Math.sin(elapsed * 0.035) * 0.13;
			}

			// Move packets
			packets.forEach(p => {
				p.t = (p.t + 0.0038) % 1;
				p.mesh.position.lerpVectors(nodes[p.from].position, nodes[p.to].position, p.t);
			});

			// Pulse nodes
			nodes.forEach((node, i) => {
				node.scale.setScalar(1 + Math.sin(elapsed * 1.6 + i * 0.65) * 0.11);
			});

			renderer.render(scene, camera);
		}
		animate();

		const ro = new ResizeObserver(entries => {
			const { width, height } = entries[0].contentRect;
			if (width > 0 && height > 0) {
				camera.aspect = width / height;
				camera.updateProjectionMatrix();
				renderer.setSize(width, height);
			}
		});
		ro.observe(container);

		return () => {
			cancelAnimationFrame(raf);
			ro.disconnect();
			renderer.dispose();
			if (container.contains(renderer.domElement)) container.removeChild(renderer.domElement);
		};
	}, []);

	return <div ref={mountRef} className={styles.networkScene} aria-hidden="true" />;
}

// ── Three.js Probe Scene (feature card 1) ────────────────────────────────────

function ProbeScene() {
	const mountRef = useRef<HTMLDivElement>(null);

	useEffect(() => {
		const container = mountRef.current;
		if (!container) return;

		const reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

		const w = container.clientWidth || 480;
		const h = container.clientHeight || 360;

		const scene = new Scene();
		const camera = new PerspectiveCamera(48, w / h, 0.1, 100);
		camera.position.set(0, 0, 7);

		const renderer = new WebGLRenderer({ alpha: true, antialias: true });
		renderer.setSize(w, h);
		renderer.setPixelRatio(Math.min(window.devicePixelRatio, 1.5));
		renderer.setClearColor(0x000000, 0);
		container.appendChild(renderer.domElement);

		// Central hub — larger octahedron
		const hubGeo = new OctahedronGeometry(0.22, 0);
		const hub = new Mesh(hubGeo, new MeshBasicMaterial({ color: 0xff7a1a }));
		scene.add(hub);

		// Two orbit rings of probe nodes
		const innerR = 1.6;
		const outerR = 2.8;
		const innerCount = 5;
		const outerCount = 7;

		const nodeGeo = new OctahedronGeometry(0.08, 0);
		const nodeMat = new MeshBasicMaterial({ color: 0xff9944 });

		type ProbeNode = { mesh: Mesh; radius: number; angle: number; speed: number; tiltX: number; tiltZ: number };
		const probeNodes: ProbeNode[] = [];

		for (let i = 0; i < innerCount; i++) {
			const mesh = new Mesh(nodeGeo, nodeMat);
			scene.add(mesh);
			probeNodes.push({
				mesh,
				radius: innerR,
				angle: (i / innerCount) * Math.PI * 2,
				speed: 0.38,
				tiltX: 0.3,
				tiltZ: 0.15
			});
		}
		for (let i = 0; i < outerCount; i++) {
			const mesh = new Mesh(nodeGeo, nodeMat);
			scene.add(mesh);
			probeNodes.push({
				mesh,
				radius: outerR,
				angle: (i / outerCount) * Math.PI * 2,
				speed: 0.22,
				tiltX: -0.2,
				tiltZ: 0.25
			});
		}

		// Thin orbit-ring lines
		const ringMat = new LineBasicMaterial({ color: 0xff7a1a, transparent: true, opacity: 0.12 });
		function makeRing(radius: number, tiltX: number) {
			const segments = 64;
			const pts = Array.from({ length: segments + 1 }, (_, i) => {
				const a = (i / segments) * Math.PI * 2;
				return new Vector3(Math.cos(a) * radius, Math.sin(a) * radius * Math.sin(tiltX), Math.sin(a) * radius * Math.cos(tiltX));
			});
			const geo = new BufferGeometry().setFromPoints(pts);
			return new Line(geo, ringMat);
		}
		scene.add(makeRing(innerR, 0.3));
		scene.add(makeRing(outerR, -0.2));

		// Connection lines from each probe to hub — updated each frame
		const connLines = probeNodes.map(() => {
			const geo = new BufferGeometry().setFromPoints([hub.position.clone(), hub.position.clone()]);
			const line = new Line(geo, new LineBasicMaterial({ color: 0xff7a1a, transparent: true, opacity: 0.25 }));
			scene.add(line);
			return line;
		});

		// Animated packets travelling hub ↔ probe
		const pktGeo = new SphereGeometry(0.045, 8, 6);
		const pktMat = new MeshBasicMaterial({ color: 0xffcc88 });
		const packets = probeNodes.slice(0, 6).map((node, i) => {
			const mesh = new Mesh(pktGeo, pktMat);
			scene.add(mesh);
			return { mesh, nodeIdx: i, t: i / 6 };
		});

		const clock = new Clock();
		let raf: number;

		function animate() {
			raf = requestAnimationFrame(animate);
			const elapsed = clock.getElapsedTime();

			// Gently rotate the whole group
			if (!reduced) {
				scene.rotation.y = elapsed * 0.12;
				scene.rotation.x = Math.sin(elapsed * 0.04) * 0.1;
			}

			// Pulse hub
			hub.scale.setScalar(1 + Math.sin(elapsed * 2.2) * 0.08);

			// Move probe nodes along their orbits
			probeNodes.forEach((p, i) => {
				const a = p.angle + elapsed * p.speed;
				p.mesh.position.set(Math.cos(a) * p.radius, Math.sin(a) * p.radius * Math.sin(p.tiltX), Math.sin(a) * p.radius * Math.cos(p.tiltX));
				p.mesh.scale.setScalar(1 + Math.sin(elapsed * 1.4 + i * 0.7) * 0.12);

				// Update connection line
				const positions = connLines[i].geometry.attributes.position;
				const arr = positions.array as Float32Array;
				arr[0] = hub.position.x;
				arr[1] = hub.position.y;
				arr[2] = hub.position.z;
				arr[3] = p.mesh.position.x;
				arr[4] = p.mesh.position.y;
				arr[5] = p.mesh.position.z;
				positions.needsUpdate = true;
			});

			// Move packets
			packets.forEach(pkt => {
				pkt.t = (pkt.t + 0.006) % 1;
				const node = probeNodes[pkt.nodeIdx];
				pkt.mesh.position.lerpVectors(hub.position, node.mesh.position, pkt.t);
			});

			renderer.render(scene, camera);
		}
		animate();

		const ro = new ResizeObserver(entries => {
			const { width, height } = entries[0].contentRect;
			if (width > 0 && height > 0) {
				camera.aspect = width / height;
				camera.updateProjectionMatrix();
				renderer.setSize(width, height);
			}
		});
		ro.observe(container);

		return () => {
			cancelAnimationFrame(raf);
			ro.disconnect();
			renderer.dispose();
			if (container.contains(renderer.domElement)) container.removeChild(renderer.domElement);
		};
	}, []);

	return <div ref={mountRef} className={styles.probeScene} aria-hidden="true" />;
}

// ── CSS Globe (hero, unchanged) ──────────────────────────────────────────────

function GlobalNetworkAnimation() {
	return (
		<div className={styles.globalStage} aria-hidden="true">
			<div className={styles.globeRig}>
				<div className={styles.globeCore} />
				<div className={styles.orbitA} />
				<div className={styles.orbitB} />
				<div className={styles.orbitC} />
				{[
					["ams", 8, 38],
					["fra", 43, 8],
					["sin", 92, 64],
					["nyc", 7, 66],
					["sfo", 90, 27]
				].map(([name, x, y]) => (
					<span key={name} className={styles.networkNode} style={{ "--x": `${x}%`, "--y": `${y}%` } as CSSProperties}>
						{name}
					</span>
				))}
				<span className={styles.packetOne} />
				<span className={styles.packetTwo} />
				<span className={styles.packetThree} />
			</div>
			<div className={styles.depthPlane} />
		</div>
	);
}
