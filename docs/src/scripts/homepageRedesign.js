const cleanupKey = "__netstampHomepageCleanup";
const eventsBoundKey = "__netstampHomepageEventsBound";

function clampNumber(value, min, max) {
	return Math.max(min, Math.min(max, value));
}

function cleanupHomepage() {
	if (typeof window[cleanupKey] === "function") {
		window[cleanupKey]();
		window[cleanupKey] = undefined;
	}
}

function initHomepage() {
	cleanupHomepage();

	const root = document.querySelector("[data-homepage]");
	if (!(root instanceof HTMLElement)) return;

	const cleanupTasks = [];
	initRouteField(root, cleanupTasks);
	initProductShowcase(root, cleanupTasks);
	initHomepageScrollEffects(root, cleanupTasks);

	window[cleanupKey] = () => {
		while (cleanupTasks.length) {
			const cleanup = cleanupTasks.pop();
			if (typeof cleanup === "function") cleanup();
		}
	};
}

const initProductShowcase = (root, cleanupTasks) => {
	const showcase = root.querySelector("[data-product-showcase]");
	if (!(showcase instanceof HTMLElement)) return;

	const controls = showcase.querySelector("[data-showcase-controls]");
	const caption = showcase.querySelector("[data-showcase-caption]");
	const triggers = Array.from(showcase.querySelectorAll("[data-showcase-trigger]")).filter(element => element instanceof HTMLButtonElement);
	const images = Array.from(showcase.querySelectorAll("[data-showcase-image]")).filter(element => element instanceof HTMLImageElement);
	if (!(controls instanceof HTMLElement) || !(caption instanceof HTMLElement) || !triggers.length || triggers.length !== images.length) return;

	const controller = new AbortController();
	const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
	let activeIndex = 0;
	let interval = null;
	let pointerPaused = false;
	let focusPaused = false;
	let visible = true;

	const setActiveIndex = nextIndex => {
		activeIndex = (nextIndex + triggers.length) % triggers.length;

		triggers.forEach((trigger, index) => {
			const active = index === activeIndex;
			trigger.classList.toggle("is-active", active);
			trigger.setAttribute("aria-pressed", String(active));
		});

		images.forEach((image, index) => image.classList.toggle("is-active", index === activeIndex));
		caption.textContent = triggers[activeIndex].dataset.showcaseDescription || triggers[activeIndex].textContent?.trim() || "Netstamp product highlight";
	};

	const stopSlideshow = () => {
		if (interval !== null) {
			window.clearInterval(interval);
			interval = null;
		}
	};

	const startSlideshow = () => {
		stopSlideshow();
		if (reduceMotion || pointerPaused || focusPaused || !visible || document.hidden) return;

		interval = window.setInterval(() => setActiveIndex(activeIndex + 1), 4800);
	};

	triggers.forEach((trigger, index) => {
		trigger.addEventListener(
			"pointerenter",
			() => {
				pointerPaused = true;
				stopSlideshow();
				setActiveIndex(index);
			},
			{ signal: controller.signal }
		);
		trigger.addEventListener("click", () => setActiveIndex(index), { signal: controller.signal });
		trigger.addEventListener(
			"focus",
			() => {
				focusPaused = true;
				stopSlideshow();
				setActiveIndex(index);
			},
			{ signal: controller.signal }
		);
		trigger.addEventListener(
			"keydown",
			event => {
				let nextIndex = null;
				if (event.key === "ArrowRight" || event.key === "ArrowDown") nextIndex = index + 1;
				if (event.key === "ArrowLeft" || event.key === "ArrowUp") nextIndex = index - 1;
				if (event.key === "Home") nextIndex = 0;
				if (event.key === "End") nextIndex = triggers.length - 1;
				if (nextIndex === null) return;

				event.preventDefault();
				triggers[(nextIndex + triggers.length) % triggers.length].focus();
			},
			{ signal: controller.signal }
		);
	});

	controls.addEventListener(
		"pointerleave",
		() => {
			pointerPaused = false;
			startSlideshow();
		},
		{ signal: controller.signal }
	);
	controls.addEventListener(
		"focusout",
		event => {
			if (event.relatedTarget instanceof Node && controls.contains(event.relatedTarget)) return;
			focusPaused = false;
			startSlideshow();
		},
		{ signal: controller.signal }
	);
	document.addEventListener("visibilitychange", startSlideshow, { signal: controller.signal });

	const observer = new IntersectionObserver(
		entries => {
			visible = entries.some(entry => entry.isIntersecting);
			startSlideshow();
		},
		{ threshold: 0.15 }
	);
	observer.observe(showcase);
	setActiveIndex(0);
	startSlideshow();

	cleanupTasks.push(() => {
		controller.abort();
		observer.disconnect();
		stopSlideshow();
	});
};

function initRouteField(root, cleanupTasks) {
	const canvas = root.querySelector("#ns-canvas");
	if (!(canvas instanceof HTMLCanvasElement)) return;

	const ctx = canvas.getContext("2d");
	if (!ctx) return;

	const controller = new AbortController();
	const CFG = {
		bgColor: "#0a0a0f",
		regularColor: [255, 255, 255],
		probeColor: [249, 115, 22],
		baseProbeCount: 20,
		baseRegularCount: 180,
		probeCount: 20,
		regularCount: 180,
		baseConnectionDist: 120,
		connectionDist: 120,
		baseMouseDist: 150,
		mouseDist: 150,
		mouseAttract: 0.012,
		probeRadius: 3,
		regularRadius: 1.5,
		probePulseSpeed: 0.04,
		routeEventInterval: 2800,
		routeEventDuration: 900,
		reducedMotion: window.matchMedia("(prefers-reduced-motion: reduce)").matches
	};

	let width = 0;
	let height = 0;
	let mouse = { x: -9999, y: -9999 };
	let particles = [];
	let routeEvents = [];
	let lastRouteEvent = 0;
	let raf = null;

	function setResponsiveParticleConfig() {
		const areaRatio = (width * height) / (1440 * 900);
		const density = clampNumber(Math.pow(areaRatio, 0.92), 0.23, 1.18);
		const compactScreenFactor = width < 640 ? 0.82 : width < 900 ? 0.92 : 1;
		const distanceScale = clampNumber(Math.sqrt(areaRatio), 0.64, 1.1);

		CFG.probeCount = Math.round(clampNumber(CFG.baseProbeCount * density * compactScreenFactor, 5, 24));
		CFG.regularCount = Math.round(clampNumber(CFG.baseRegularCount * density * compactScreenFactor, 42, 212));
		CFG.connectionDist = CFG.baseConnectionDist * distanceScale;
		CFG.mouseDist = CFG.baseMouseDist * distanceScale;
	}

	function resize() {
		const previousProbeCount = CFG.probeCount;
		const previousRegularCount = CFG.regularCount;
		width = canvas.width = window.innerWidth;
		height = canvas.height = window.innerHeight;
		setResponsiveParticleConfig();

		if (particles.length && (previousProbeCount !== CFG.probeCount || previousRegularCount !== CFG.regularCount)) {
			initParticles();
		}
	}

	function makeParticle(isProbe) {
		const speed = isProbe ? 0.25 + Math.random() * 0.35 : 0.08 + Math.random() * 0.18;
		const angle = Math.random() * Math.PI * 2;

		return {
			x: Math.random() * width,
			y: Math.random() * height,
			vx: Math.cos(angle) * speed,
			vy: Math.sin(angle) * speed,
			isProbe,
			r: isProbe ? CFG.probeRadius : CFG.regularRadius,
			phase: Math.random() * Math.PI * 2
		};
	}

	function initParticles() {
		particles = [];
		for (let index = 0; index < CFG.probeCount; index++) particles.push(makeParticle(true));
		for (let index = 0; index < CFG.regularCount; index++) particles.push(makeParticle(false));
	}

	function buildPath(src, dst) {
		const path = [src];
		const visited = new Set([src]);
		let current = src;
		const maxHops = 12;

		while (current !== dst && path.length <= maxHops) {
			let best = null;
			let bestScore = Infinity;
			for (let index = 0; index < particles.length; index++) {
				if (visited.has(index)) continue;
				const dx = particles[index].x - particles[current].x;
				const dy = particles[index].y - particles[current].y;
				const distToCurrent = Math.sqrt(dx * dx + dy * dy);
				if (distToCurrent > CFG.connectionDist * 2.5) continue;
				const ex = particles[dst].x - particles[index].x;
				const ey = particles[dst].y - particles[index].y;
				const distToDst = Math.sqrt(ex * ex + ey * ey);
				if (distToDst < bestScore) {
					bestScore = distToDst;
					best = index;
				}
			}
			if (best === null) break;
			visited.add(best);
			path.push(best);
			current = best;
		}
		if (current !== dst) path.push(dst);
		return path;
	}

	function spawnRouteEvent() {
		const probeIndices = [];
		for (let index = 0; index < particles.length; index++) {
			if (particles[index].isProbe) probeIndices.push(index);
		}
		if (probeIndices.length < 2) return;

		const ai = Math.floor(Math.random() * probeIndices.length);
		let bi;
		do {
			bi = Math.floor(Math.random() * probeIndices.length);
		} while (bi === ai);

		routeEvents.push({
			path: buildPath(probeIndices[ai], probeIndices[bi]),
			t: 0
		});
	}

	function update(ts) {
		for (const particle of particles) {
			const mdx = mouse.x - particle.x;
			const mdy = mouse.y - particle.y;
			const mdist = Math.sqrt(mdx * mdx + mdy * mdy);
			if (mdist < CFG.mouseDist && mdist > 0.5) {
				const mf = (1 - mdist / CFG.mouseDist) * CFG.mouseAttract;
				particle.vx += (mdx / mdist) * mf;
				particle.vy += (mdy / mdist) * mf;
			}

			const speed = Math.sqrt(particle.vx * particle.vx + particle.vy * particle.vy);
			const maxSpeed = particle.isProbe ? 0.7 : 0.35;
			if (speed > maxSpeed) {
				particle.vx = (particle.vx / speed) * maxSpeed;
				particle.vy = (particle.vy / speed) * maxSpeed;
			}

			particle.x += particle.vx;
			particle.y += particle.vy;

			if (particle.x < -10) particle.x = width + 10;
			else if (particle.x > width + 10) particle.x = -10;
			if (particle.y < -10) particle.y = height + 10;
			else if (particle.y > height + 10) particle.y = -10;

			particle.phase += CFG.probePulseSpeed;
		}

		const dt = 0.016;
		for (let index = routeEvents.length - 1; index >= 0; index--) {
			routeEvents[index].t += dt * (1000 / CFG.routeEventDuration);
			if (routeEvents[index].t >= 1) routeEvents.splice(index, 1);
		}

		if (ts - lastRouteEvent > CFG.routeEventInterval) {
			lastRouteEvent = ts;
			spawnRouteEvent();
		}
	}

	function draw() {
		ctx.clearRect(0, 0, width, height);
		ctx.fillStyle = CFG.bgColor;
		ctx.fillRect(0, 0, width, height);

		const regularColor = CFG.regularColor;
		const probeColor = CFG.probeColor;

		for (let i = 0; i < particles.length; i++) {
			const a = particles[i];
			for (let j = i + 1; j < particles.length; j++) {
				const b = particles[j];
				const dx = a.x - b.x;
				const dy = a.y - b.y;
				const dist = Math.sqrt(dx * dx + dy * dy);
				if (dist > CFG.connectionDist) continue;
				const alpha = (1 - dist / CFG.connectionDist) * 0.45;
				const isOrange = a.isProbe || b.isProbe;
				ctx.strokeStyle = isOrange ? `rgba(${probeColor[0]},${probeColor[1]},${probeColor[2]},${alpha * 0.85})` : `rgba(${regularColor[0]},${regularColor[1]},${regularColor[2]},${alpha * 0.35})`;
				ctx.lineWidth = isOrange ? 0.75 : 0.5;
				ctx.beginPath();
				ctx.moveTo(a.x, a.y);
				ctx.lineTo(b.x, b.y);
				ctx.stroke();
			}

			const mdx = mouse.x - a.x;
			const mdy = mouse.y - a.y;
			const mdist = Math.sqrt(mdx * mdx + mdy * mdy);
			if (mdist < CFG.mouseDist) {
				const malpha = (1 - mdist / CFG.mouseDist) * 0.55;
				ctx.strokeStyle = `rgba(255,255,255,${malpha})`;
				ctx.lineWidth = 0.6;
				ctx.beginPath();
				ctx.moveTo(a.x, a.y);
				ctx.lineTo(mouse.x, mouse.y);
				ctx.stroke();
			}
		}

		for (const event of routeEvents) {
			const pathLen = event.path.length - 1;
			if (pathLen < 1) continue;
			const globalT = event.t * pathLen;
			const segIdx = Math.min(Math.floor(globalT), pathLen - 1);
			const segT = globalT - segIdx;
			const nodeA = particles[event.path[segIdx]];
			const nodeB = particles[event.path[segIdx + 1]];
			if (!nodeA || !nodeB) continue;
			const px = nodeA.x + (nodeB.x - nodeA.x) * segT;
			const py = nodeA.y + (nodeB.y - nodeA.y) * segT;
			const grd = ctx.createRadialGradient(px, py, 0, px, py, 8);
			grd.addColorStop(0, "rgba(255,200,120,0.95)");
			grd.addColorStop(0.3, `rgba(${probeColor[0]},${probeColor[1]},${probeColor[2]},0.7)`);
			grd.addColorStop(1, `rgba(${probeColor[0]},${probeColor[1]},${probeColor[2]},0)`);
			ctx.beginPath();
			ctx.arc(px, py, 8, 0, Math.PI * 2);
			ctx.fillStyle = grd;
			ctx.fill();
			ctx.beginPath();
			ctx.arc(px, py, 2, 0, Math.PI * 2);
			ctx.fillStyle = "rgba(255,240,200,1)";
			ctx.fill();
		}

		for (const particle of particles) {
			if (particle.isProbe) {
				const pulseR = particle.r + 2.5 + Math.sin(particle.phase) * 1.5;
				const glow = ctx.createRadialGradient(particle.x, particle.y, 0, particle.x, particle.y, pulseR * 3.5);
				glow.addColorStop(0, `rgba(${probeColor[0]},${probeColor[1]},${probeColor[2]},0.22)`);
				glow.addColorStop(1, `rgba(${probeColor[0]},${probeColor[1]},${probeColor[2]},0)`);
				ctx.beginPath();
				ctx.arc(particle.x, particle.y, pulseR * 3.5, 0, Math.PI * 2);
				ctx.fillStyle = glow;
				ctx.fill();
				ctx.beginPath();
				ctx.arc(particle.x, particle.y, particle.r + Math.sin(particle.phase) * 0.6, 0, Math.PI * 2);
				ctx.fillStyle = `rgb(${probeColor[0]},${probeColor[1]},${probeColor[2]})`;
				ctx.fill();
			} else {
				ctx.beginPath();
				ctx.arc(particle.x, particle.y, particle.r, 0, Math.PI * 2);
				ctx.fillStyle = "rgba(255,255,255,0.72)";
				ctx.fill();
			}
		}

		if (mouse.x > 0) {
			const cr = ctx.createRadialGradient(mouse.x, mouse.y, 0, mouse.x, mouse.y, 14);
			cr.addColorStop(0, "rgba(255,255,255,0.25)");
			cr.addColorStop(1, "rgba(255,255,255,0)");
			ctx.beginPath();
			ctx.arc(mouse.x, mouse.y, 14, 0, Math.PI * 2);
			ctx.fillStyle = cr;
			ctx.fill();
			ctx.beginPath();
			ctx.arc(mouse.x, mouse.y, 2.5, 0, Math.PI * 2);
			ctx.fillStyle = "rgba(255,255,255,0.9)";
			ctx.fill();
		}
	}

	function drawStatic() {
		ctx.fillStyle = CFG.bgColor;
		ctx.fillRect(0, 0, width, height);
		const regularColor = CFG.regularColor;
		const probeColor = CFG.probeColor;
		for (let i = 0; i < particles.length; i++) {
			const a = particles[i];
			for (let j = i + 1; j < particles.length; j++) {
				const b = particles[j];
				const dx = a.x - b.x;
				const dy = a.y - b.y;
				const dist = Math.sqrt(dx * dx + dy * dy);
				if (dist > CFG.connectionDist) continue;
				const alpha = (1 - dist / CFG.connectionDist) * 0.3;
				const isOrange = a.isProbe || b.isProbe;
				ctx.strokeStyle = isOrange ? `rgba(${probeColor[0]},${probeColor[1]},${probeColor[2]},${alpha})` : `rgba(255,255,255,${alpha * 0.4})`;
				ctx.lineWidth = 0.5;
				ctx.beginPath();
				ctx.moveTo(a.x, a.y);
				ctx.lineTo(b.x, b.y);
				ctx.stroke();
			}
		}
		for (const particle of particles) {
			ctx.beginPath();
			ctx.arc(particle.x, particle.y, particle.r, 0, Math.PI * 2);
			ctx.fillStyle = particle.isProbe ? `rgb(${probeColor[0]},${probeColor[1]},${probeColor[2]})` : "rgba(255,255,255,0.65)";
			ctx.fill();
		}
	}

	function loop(ts) {
		update(ts);
		draw();
		raf = window.requestAnimationFrame(loop);
	}

	const resetMouse = () => {
		mouse = { x: -9999, y: -9999 };
	};

	window.addEventListener("resize", resize, { signal: controller.signal });
	window.addEventListener(
		"mousemove",
		event => {
			mouse.x = event.clientX;
			mouse.y = event.clientY;
		},
		{ signal: controller.signal }
	);
	window.addEventListener("mouseleave", resetMouse, { signal: controller.signal });
	window.addEventListener(
		"touchmove",
		event => {
			if (event.touches.length) {
				mouse.x = event.touches[0].clientX;
				mouse.y = event.touches[0].clientY;
			}
		},
		{ passive: true, signal: controller.signal }
	);
	window.addEventListener("touchend", resetMouse, { signal: controller.signal });

	resize();
	initParticles();

	if (CFG.reducedMotion) {
		drawStatic();
	} else {
		raf = window.requestAnimationFrame(loop);
	}

	cleanupTasks.push(() => {
		controller.abort();
		if (raf !== null) window.cancelAnimationFrame(raf);
	});
}

function initHomepageScrollEffects(root, cleanupTasks) {
	const controller = new AbortController();
	const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
	const progressSections = Array.from(root.querySelectorAll("[data-scroll-progress]")).filter(element => element instanceof HTMLElement);
	const introHighlights = Array.from(root.querySelectorAll("[data-intro-highlight]")).filter(element => element instanceof HTMLElement);
	const flowCards = Array.from(root.querySelectorAll("[data-flow-card]")).filter(element => element instanceof HTMLElement);
	let raf = null;

	function elementProgress(element, startRatio = 0.82, endRatio = 0.22) {
		const rect = element.getBoundingClientRect();
		const viewportHeight = window.innerHeight || document.documentElement.clientHeight || 1;
		const start = viewportHeight * startRatio;
		const end = viewportHeight * endRatio;
		const travel = start - end + rect.height * 0.3;
		return clampNumber((start - rect.top) / travel, 0, 1);
	}

	function update() {
		for (const section of progressSections) {
			const progress = reduceMotion ? 1 : elementProgress(section);
			section.style.setProperty("--section-progress", progress.toFixed(4));
			section.style.setProperty("--section-progress-percent", `${(progress * 100).toFixed(2)}%`);
			if (section.dataset.scrollProgress === "cta") {
				section.style.setProperty("--route-offset", ((1 - progress) * 1100).toFixed(2));
			}
		}

		const introSection = root.querySelector('[data-scroll-progress="intro"]');
		const introProgress = introSection instanceof HTMLElement ? (reduceMotion ? 1 : elementProgress(introSection, 0.78, 0.2)) : 1;
		for (let index = 0; index < introHighlights.length; index++) {
			const localProgress = reduceMotion ? 1 : clampNumber((introProgress - index * 0.075) / 0.54, 0, 1);
			introHighlights[index].style.setProperty("--highlight-clip", `${((1 - localProgress) * 100).toFixed(2)}%`);
		}

		for (const card of flowCards) {
			const cardProgress = reduceMotion ? 1 : elementProgress(card, 0.78, 0.38);
			card.style.setProperty("--card-progress", cardProgress.toFixed(4));
			card.style.setProperty("--card-height", `${(16 * cardProgress).toFixed(2)}rem`);
			card.style.setProperty("--card-shift", `${((1 - cardProgress) * -0.8).toFixed(2)}rem`);
			card.classList.toggle("is-open", cardProgress > 0.42);
		}
	}

	function scheduleUpdate() {
		if (raf !== null) return;
		raf = window.requestAnimationFrame(() => {
			raf = null;
			update();
		});
	}

	window.addEventListener("scroll", scheduleUpdate, { passive: true, signal: controller.signal });
	window.addEventListener("resize", scheduleUpdate, { signal: controller.signal });
	update();

	cleanupTasks.push(() => {
		controller.abort();
		if (raf !== null) window.cancelAnimationFrame(raf);
	});
}

if (!window[eventsBoundKey]) {
	document.addEventListener("astro:page-load", initHomepage);
	document.addEventListener("astro:before-swap", cleanupHomepage);
	window[eventsBoundKey] = true;
}

if (document.readyState === "loading") {
	document.addEventListener("DOMContentLoaded", initHomepage, { once: true });
} else {
	initHomepage();
}
