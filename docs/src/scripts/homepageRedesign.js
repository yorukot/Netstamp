import * as THREE from "three";
import { RoomEnvironment } from "three/examples/jsm/environments/RoomEnvironment.js";

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
	initCable(root, cleanupTasks);
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

function initCable(root, cleanupTasks) {
	const canvas = root.querySelector("#ns-cable-canvas");
	const title = root.querySelector("h1");
	if (!(canvas instanceof HTMLCanvasElement) || !(title instanceof HTMLElement)) return;

	const controller = new AbortController();
	const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
	const scene = new THREE.Scene();
	const renderer = new THREE.WebGLRenderer({
		canvas,
		alpha: true,
		antialias: true,
		preserveDrawingBuffer: true,
		powerPreference: "high-performance"
	});

	renderer.setClearColor(0x000000, 0);
	renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2));
	renderer.outputColorSpace = THREE.SRGBColorSpace;
	renderer.toneMapping = THREE.ACESFilmicToneMapping;
	renderer.toneMappingExposure = 1.0;

	const camera = new THREE.OrthographicCamera(-1, 1, 1, -1, 0.1, 2200);
	camera.position.set(0, 0, 920);
	camera.lookAt(0, 0, 0);

	const pmrem = new THREE.PMREMGenerator(renderer);
	const environment = pmrem.fromScene(new RoomEnvironment(), 0.045).texture;
	scene.environment = environment;

	const cableGroup = new THREE.Group();
	scene.add(cableGroup);
	let wholeCable = null;
	let leftSegment = null;
	let rightSegment = null;
	let leftCap = null;
	let rightCap = null;
	let breakMotion = null;
	let raf = null;
	let destroyed = false;

	const keyLight = new THREE.DirectionalLight("#f8fafc", 1.55);
	keyLight.position.set(-220, 280, 520);
	scene.add(keyLight);

	const rimLight = new THREE.DirectionalLight("#38bdf8", 2.2);
	rimLight.position.set(360, 260, 420);
	scene.add(rimLight);

	const coolFill = new THREE.DirectionalLight("#6ba6bb", 0.42);
	coolFill.position.set(-420, -180, 300);
	scene.add(coolFill);

	const ambient = new THREE.HemisphereLight("#c4ccd9", "#010509", 0.56);
	scene.add(ambient);

	function easeOutCubic(value) {
		return 1 - Math.pow(1 - value, 3);
	}

	function smoothstep(value) {
		const t = clampNumber(value, 0, 1);
		return t * t * (3 - 2 * t);
	}

	function driftWave(elapsed, frequency, phase = 0) {
		return (Math.sin(elapsed * frequency + phase) - Math.sin(phase)) * 0.72;
	}

	function makeRubberBumpTexture() {
		const size = 128;
		const noiseCanvas = document.createElement("canvas");
		noiseCanvas.width = size;
		noiseCanvas.height = size;
		const ctx = noiseCanvas.getContext("2d");
		if (!ctx) return null;

		const image = ctx.createImageData(size, size);
		for (let index = 0; index < image.data.length; index += 4) {
			const base = 124 + Math.random() * 10;
			image.data[index] = base;
			image.data[index + 1] = base;
			image.data[index + 2] = base;
			image.data[index + 3] = 255;
		}
		ctx.putImageData(image, 0, 0);

		const texture = new THREE.CanvasTexture(noiseCanvas);
		texture.wrapS = THREE.RepeatWrapping;
		texture.wrapT = THREE.RepeatWrapping;
		texture.repeat.set(18, 3);
		texture.colorSpace = THREE.NoColorSpace;
		return texture;
	}

	const rubberBump = makeRubberBumpTexture();
	const jacketMaterial = new THREE.MeshPhysicalMaterial({
		color: "#061923",
		roughness: 0.58,
		metalness: 0,
		clearcoat: 0.15,
		clearcoatRoughness: 0.65,
		sheen: 0.2,
		sheenRoughness: 0.9,
		envMapIntensity: 0.6,
		bumpMap: rubberBump || undefined,
		bumpScale: 0.018
	});
	const wholeCableMaterial = jacketMaterial.clone();
	wholeCableMaterial.transparent = true;
	wholeCableMaterial.opacity = 1;
	const jacketEndMaterial = new THREE.MeshStandardMaterial({
		color: "#041016",
		roughness: 0.68,
		metalness: 0,
		envMapIntensity: 0.38,
		side: THREE.DoubleSide
	});
	const sheathMaterial = new THREE.MeshStandardMaterial({
		color: "#d7dbe2",
		roughness: 0.52,
		metalness: 0,
		envMapIntensity: 0.42,
		side: THREE.DoubleSide
	});
	const innerSheathMaterial = new THREE.MeshStandardMaterial({
		color: "#9fa7b5",
		roughness: 0.58,
		metalness: 0,
		envMapIntensity: 0.36,
		side: THREE.DoubleSide
	});
	const fiberMaterials = ["#fb923c", "#38bdf8", "#f8fafc", "#ea6a1a", "#c4ccd9", "#2563eb"].map(
		color =>
			new THREE.MeshStandardMaterial({
				color,
				roughness: 0.34,
				metalness: color === "#c4ccd9" ? 0.14 : 0.04,
				envMapIntensity: 0.5,
				side: THREE.DoubleSide
			})
	);
	const ownedMaterials = [jacketMaterial, wholeCableMaterial, jacketEndMaterial, sheathMaterial, innerSheathMaterial, ...fiberMaterials];

	function disposeObject(object) {
		if (object.geometry) object.geometry.dispose();
		if (object.material) {
			if (Array.isArray(object.material)) {
				for (const material of object.material) {
					if (!ownedMaterials.includes(material)) material.dispose();
				}
			} else if (!ownedMaterials.includes(object.material)) {
				object.material.dispose();
			}
		}
	}

	function clearCable() {
		cableGroup.traverse(object => {
			if (object instanceof THREE.Mesh) disposeObject(object);
		});
		cableGroup.clear();
		wholeCable = null;
		leftCap = null;
		rightCap = null;
	}

	function createCableTube(points, radius, material = jacketMaterial, tubularSegments = 240) {
		const curve = new THREE.CatmullRomCurve3(points, false, "centripetal", 0.22);
		const geometry = new THREE.TubeGeometry(curve, tubularSegments, radius, 32, false);
		geometry.computeVertexNormals();
		const mesh = new THREE.Mesh(geometry, material);
		mesh.frustumCulled = false;
		return { mesh, curve };
	}

	function addLocalDisc(group, radius, material, x = 0, y = 0, z = 0.08, segments = 64) {
		const geometry = new THREE.CircleGeometry(radius, segments);
		geometry.computeVertexNormals();
		const mesh = new THREE.Mesh(geometry, material);
		mesh.position.set(x, y, z);
		group.add(mesh);
		return mesh;
	}

	function addLocalRing(group, innerRadius, outerRadius, material, z = 0.08) {
		const geometry = new THREE.RingGeometry(innerRadius, outerRadius, 96);
		geometry.computeVertexNormals();
		const mesh = new THREE.Mesh(geometry, material);
		mesh.position.z = z;
		group.add(mesh);
		return mesh;
	}

	function addLocalFiberCore(group, fiber, material) {
		const coreGeometry = new THREE.CylinderGeometry(fiber.radius, fiber.radius * 0.92, fiber.depth, 28, 1, false);
		coreGeometry.rotateX(Math.PI / 2);
		coreGeometry.computeVertexNormals();
		const core = new THREE.Mesh(coreGeometry, material);
		core.position.set(fiber.x, fiber.y, fiber.z + fiber.depth / 2);
		core.rotation.x = fiber.tiltY;
		core.rotation.y = fiber.tiltX;
		group.add(core);

		addLocalDisc(group, fiber.radius * 0.98, material, fiber.x + fiber.tipX, fiber.y + fiber.tipY, fiber.z + fiber.depth + 0.012, 28);
	}

	function makeCableCap({ curve, atEnd, radius }) {
		const t = atEnd === "start" ? 0 : 1;
		const position = curve.getPoint(t);
		let tangent = curve.getTangent(t).normalize();
		if (atEnd === "start") tangent = tangent.clone().negate();

		const cap = new THREE.Group();
		cap.position.copy(position);
		cap.quaternion.setFromUnitVectors(new THREE.Vector3(0, 0, 1), tangent);

		addLocalDisc(cap, radius * 1.02, jacketEndMaterial, 0, 0, 0, 96);
		addLocalRing(cap, radius * 0.72, radius * 0.98, jacketEndMaterial, 0.09);
		addLocalDisc(cap, radius * 0.7, sheathMaterial, 0, 0, 0.12, 96);
		addLocalRing(cap, radius * 0.44, radius * 0.61, innerSheathMaterial, 0.16);
		addLocalDisc(cap, radius * 0.34, jacketEndMaterial, 0, 0, 0.18, 64);

		const fibers = [
			{ x: 0.02, y: -0.01, radius: 0.098, depth: 0.22, tiltX: 0.02, tiltY: -0.03, tipX: 0.004, tipY: -0.003 },
			{ x: 0.26, y: 0.03, radius: 0.074, depth: 0.15, tiltX: -0.015, tiltY: 0.02, tipX: -0.002, tipY: 0.004 },
			{ x: -0.27, y: -0.04, radius: 0.069, depth: 0.28, tiltX: 0.035, tiltY: 0.015, tipX: 0.006, tipY: 0.002 },
			{ x: -0.03, y: 0.27, radius: 0.071, depth: 0.18, tiltX: -0.025, tiltY: -0.01, tipX: -0.004, tipY: -0.001 },
			{ x: 0.04, y: -0.28, radius: 0.067, depth: 0.24, tiltX: 0.018, tiltY: 0.03, tipX: 0.002, tipY: 0.006 },
			{ x: 0.2, y: 0.19, radius: 0.061, depth: 0.11, tiltX: -0.018, tiltY: 0.018, tipX: 0.002, tipY: -0.003 },
			{ x: -0.19, y: -0.21, radius: 0.064, depth: 0.19, tiltX: 0.012, tiltY: -0.026, tipX: -0.003, tipY: 0.003 }
		];

		for (let index = 0; index < fibers.length; index++) {
			const fiber = fibers[index];
			addLocalFiberCore(
				cap,
				{
					x: fiber.x * radius,
					y: fiber.y * radius,
					radius: fiber.radius * radius,
					depth: fiber.depth * radius,
					z: 0.2 + index * 0.006,
					tiltX: fiber.tiltX,
					tiltY: fiber.tiltY,
					tipX: fiber.tipX * radius,
					tipY: fiber.tipY * radius
				},
				fiberMaterials[index % fiberMaterials.length]
			);
		}

		return cap;
	}

	function applyBreakMotion(time) {
		if (!breakMotion || !leftSegment || !rightSegment) return;

		const breakDelay = 0;
		const breakDuration = 1900;
		const rawProgress = reduceMotion ? 1 : clampNumber((time - breakMotion.startedAt - breakDelay) / breakDuration, 0, 1);
		const progress = easeOutCubic(rawProgress);
		const turnProgress = easeOutCubic(clampNumber((rawProgress - 0.08) / 0.92, 0, 1));
		const idleProgress = smoothstep((rawProgress - 0.48) / 0.52);
		const breakElapsed = Math.max(0, time - breakMotion.startedAt - breakDelay) * 0.001;
		const idleStart = breakMotion.startedAt + breakDelay + breakDuration * 0.48;
		const idleElapsed = Math.max(0, time - idleStart) * 0.001;
		const showCaps = progress > 0.035;
		const showSegments = progress > 0.01;
		const wholeOpacity = clampNumber(1 - progress / 0.18, 0, 1);

		if (wholeCable) {
			wholeCableMaterial.opacity = wholeOpacity;
			wholeCable.visible = wholeOpacity > 0.02;
		}
		leftSegment.visible = showSegments;
		rightSegment.visible = showSegments;

		leftSegment.position.lerpVectors(breakMotion.leftStart, breakMotion.leftEnd, progress);
		rightSegment.position.lerpVectors(breakMotion.rightStart, breakMotion.rightEnd, progress);
		leftSegment.position.x += driftWave(idleElapsed, 0.7, 0.4) * breakMotion.floatX * idleProgress;
		leftSegment.position.y += driftWave(idleElapsed, 0.92, 1.7) * breakMotion.floatY * idleProgress;
		leftSegment.position.z += driftWave(idleElapsed, 0.58, 2.2) * breakMotion.floatZ * idleProgress;
		rightSegment.position.x += driftWave(idleElapsed, 0.68, 2.6) * breakMotion.floatX * idleProgress;
		rightSegment.position.y += driftWave(idleElapsed, 0.86, 0.1) * breakMotion.floatY * idleProgress;
		rightSegment.position.z += driftWave(idleElapsed, 0.62, 3.8) * breakMotion.floatZ * idleProgress;
		leftSegment.rotation.z =
			breakMotion.rotation * turnProgress + driftWave(breakElapsed, 0.55) * breakMotion.rotationDrift * turnProgress + driftWave(idleElapsed, 0.74, 0.9) * breakMotion.idleRotation * idleProgress;
		rightSegment.rotation.z =
			breakMotion.rotation * turnProgress + driftWave(breakElapsed, 0.5, 1.4) * breakMotion.rotationDrift * turnProgress + driftWave(idleElapsed, 0.68, 2.1) * breakMotion.idleRotation * idleProgress;

		if (leftCap) leftCap.visible = showCaps;
		if (rightCap) rightCap.visible = showCaps;
	}

	function makeCableScene() {
		clearCable();
		leftSegment = new THREE.Group();
		rightSegment = new THREE.Group();

		const width = window.innerWidth;
		const height = window.innerHeight;
		const isDesktop = width >= 900;
		const titleSize = Number.parseFloat(window.getComputedStyle(title).fontSize) || 104;
		const cableRadius = clampNumber(titleSize * 0.78, 60, 126);
		const centerY = clampNumber(0, -height / 2 + cableRadius * 1.08, height / 2 - cableRadius * 1.08);
		const offscreen = width / 2 + cableRadius * 5;
		const innerLead = clampNumber(width * 0.1, 76, 176);
		const softLead = clampNumber(width * 0.24, 150, 390);
		const shoulder = clampNumber(width * 0.43, 250, 650);
		const outerShoulder = clampNumber(width * 0.62, 380, 920);
		const arcScale = isDesktop ? 1.18 : 0.82;
		const centerZ = 108;
		const faceTiltZ = clampNumber(cableRadius * 0.72, 44, 92);
		const nearZ = 72;
		const midZ = -36;
		const farZ = -126;

		const leftPoints = [
			new THREE.Vector3(-offscreen, centerY - cableRadius * 0.3 * arcScale, farZ - 16),
			new THREE.Vector3(-outerShoulder, centerY - cableRadius * 0.24 * arcScale, farZ),
			new THREE.Vector3(-shoulder, centerY + cableRadius * 0.08 * arcScale, midZ),
			new THREE.Vector3(-softLead, centerY + cableRadius * 0.2 * arcScale, nearZ),
			new THREE.Vector3(-innerLead, centerY, centerZ - faceTiltZ),
			new THREE.Vector3(0, centerY, centerZ)
		];
		const rightPoints = [
			new THREE.Vector3(0, centerY, centerZ),
			new THREE.Vector3(innerLead, centerY, centerZ - faceTiltZ),
			new THREE.Vector3(softLead, centerY + cableRadius * 0.2 * arcScale, nearZ),
			new THREE.Vector3(shoulder, centerY + cableRadius * 0.08 * arcScale, midZ),
			new THREE.Vector3(outerShoulder, centerY - cableRadius * 0.24 * arcScale, farZ),
			new THREE.Vector3(offscreen, centerY - cableRadius * 0.3 * arcScale, farZ - 16)
		];
		const wholePoints = [
			new THREE.Vector3(-offscreen, centerY - cableRadius * 0.3 * arcScale, farZ - 16),
			new THREE.Vector3(-outerShoulder, centerY - cableRadius * 0.24 * arcScale, farZ),
			new THREE.Vector3(-shoulder, centerY + cableRadius * 0.08 * arcScale, midZ),
			new THREE.Vector3(-softLead, centerY + cableRadius * 0.2 * arcScale, nearZ),
			new THREE.Vector3(-innerLead, centerY, centerZ - faceTiltZ),
			new THREE.Vector3(0, centerY, centerZ),
			new THREE.Vector3(innerLead, centerY, centerZ - faceTiltZ),
			new THREE.Vector3(softLead, centerY + cableRadius * 0.2 * arcScale, nearZ),
			new THREE.Vector3(shoulder, centerY + cableRadius * 0.08 * arcScale, midZ),
			new THREE.Vector3(outerShoulder, centerY - cableRadius * 0.24 * arcScale, farZ),
			new THREE.Vector3(offscreen, centerY - cableRadius * 0.3 * arcScale, farZ - 16)
		];

		const wholeCableResult = createCableTube(wholePoints, cableRadius, wholeCableMaterial, 260);
		const leftCable = createCableTube(leftPoints, cableRadius);
		const rightCable = createCableTube(rightPoints, cableRadius);
		wholeCable = wholeCableResult.mesh;
		wholeCableMaterial.opacity = reduceMotion ? 0 : 1;
		wholeCable.visible = !reduceMotion;
		leftCap = makeCableCap({ curve: leftCable.curve, atEnd: "end", radius: cableRadius });
		rightCap = makeCableCap({ curve: rightCable.curve, atEnd: "start", radius: cableRadius });
		leftCap.visible = reduceMotion;
		rightCap.visible = reduceMotion;

		leftSegment.add(leftCable.mesh);
		leftSegment.add(leftCap);
		rightSegment.add(rightCable.mesh);
		rightSegment.add(rightCap);
		leftSegment.visible = reduceMotion;
		rightSegment.visible = reduceMotion;

		cableGroup.add(wholeCable);
		cableGroup.add(leftSegment);
		cableGroup.add(rightSegment);

		const horizontalFlightRatio = isDesktop ? clampNumber(0.23 + width * 0.00004, 0.26, 0.34) : clampNumber(0.13 + width * 0.0001, 0.17, 0.22);
		const driftX = width * horizontalFlightRatio;
		const driftY = isDesktop ? clampNumber(height * 0.19, 140, 260) : clampNumber(height * 0.075, 44, 78);
		breakMotion = {
			startedAt: performance.now(),
			leftStart: new THREE.Vector3(0, 0, 0),
			rightStart: new THREE.Vector3(0, 0, 0),
			leftEnd: new THREE.Vector3(-driftX, -driftY, 0),
			rightEnd: new THREE.Vector3(driftX, driftY, 0),
			rotation: isDesktop ? 0.27 : 0.09,
			rotationDrift: isDesktop ? 0.012 : 0.005,
			idleRotation: isDesktop ? 0.018 : 0.008,
			floatX: isDesktop ? clampNumber(width * 0.006, 8, 18) : clampNumber(width * 0.01, 3, 6),
			floatY: isDesktop ? clampNumber(height * 0.008, 6, 14) : clampNumber(height * 0.006, 3, 6),
			floatZ: isDesktop ? clampNumber(cableRadius * 0.12, 7, 16) : clampNumber(cableRadius * 0.08, 4, 9)
		};
		applyBreakMotion(reduceMotion ? performance.now() + 3000 : 0);
	}

	function resize() {
		const width = window.innerWidth;
		const height = window.innerHeight;

		renderer.setSize(width, height, false);
		camera.left = -width / 2;
		camera.right = width / 2;
		camera.top = height / 2;
		camera.bottom = -height / 2;
		camera.updateProjectionMatrix();
		makeCableScene();
		renderer.render(scene, camera);
	}

	function animate(time) {
		if (destroyed) return;
		applyBreakMotion(time);
		renderer.render(scene, camera);

		if (!reduceMotion) {
			raf = window.requestAnimationFrame(animate);
		}
	}

	window.addEventListener("resize", resize, { signal: controller.signal });
	resize();
	animate(0);

	cleanupTasks.push(() => {
		destroyed = true;
		controller.abort();
		if (raf !== null) window.cancelAnimationFrame(raf);
		clearCable();
		for (const material of ownedMaterials) material.dispose();
		if (rubberBump) rubberBump.dispose();
		environment.dispose();
		pmrem.dispose();
		renderer.dispose();
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
