import { useEffect, useRef } from "react";
import { BufferGeometry, Clock, Line, LineBasicMaterial, Mesh, MeshBasicMaterial, OctahedronGeometry, PerspectiveCamera, Scene, SphereGeometry, Vector3, WebGLRenderer } from "three";
import styles from "./LandingPage.module.css";
import { disposeSceneObjects } from "./threeSceneCleanup";

type ProbeNode = { mesh: Mesh; radius: number; angle: number; speed: number; tiltX: number; tiltZ: number };

export function ProbeScene() {
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

		const hubGeo = new OctahedronGeometry(0.22, 0);
		const hub = new Mesh(hubGeo, new MeshBasicMaterial({ color: 0xff7a1a }));
		scene.add(hub);

		const innerR = 1.6;
		const outerR = 2.8;
		const innerCount = 5;
		const outerCount = 7;
		const nodeGeo = new OctahedronGeometry(0.08, 0);
		const nodeMat = new MeshBasicMaterial({ color: 0xff9944 });
		const probeNodes: ProbeNode[] = [];

		for (let i = 0; i < innerCount; i++) {
			const mesh = new Mesh(nodeGeo, nodeMat);
			scene.add(mesh);
			probeNodes.push({ mesh, radius: innerR, angle: (i / innerCount) * Math.PI * 2, speed: 0.38, tiltX: 0.3, tiltZ: 0.15 });
		}

		for (let i = 0; i < outerCount; i++) {
			const mesh = new Mesh(nodeGeo, nodeMat);
			scene.add(mesh);
			probeNodes.push({ mesh, radius: outerR, angle: (i / outerCount) * Math.PI * 2, speed: 0.22, tiltX: -0.2, tiltZ: 0.25 });
		}

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

		const connLines = probeNodes.map(() => {
			const geo = new BufferGeometry().setFromPoints([hub.position.clone(), hub.position.clone()]);
			const line = new Line(geo, new LineBasicMaterial({ color: 0xff7a1a, transparent: true, opacity: 0.25 }));
			scene.add(line);
			return line;
		});

		const pktGeo = new SphereGeometry(0.045, 8, 6);
		const pktMat = new MeshBasicMaterial({ color: 0xffcc88 });
		const packets = probeNodes.slice(0, 6).map((node, i) => {
			const mesh = new Mesh(pktGeo, pktMat);
			scene.add(mesh);
			return { mesh, nodeIdx: i, t: i / 6 };
		});

		const clock = new Clock();
		let raf = 0;

		function renderFrame(elapsed: number) {
			scene.rotation.y = elapsed * 0.12;
			scene.rotation.x = Math.sin(elapsed * 0.04) * 0.1;
			hub.scale.setScalar(1 + Math.sin(elapsed * 2.2) * 0.08);

			probeNodes.forEach((probeNode, i) => {
				const a = probeNode.angle + elapsed * probeNode.speed;
				probeNode.mesh.position.set(Math.cos(a) * probeNode.radius, Math.sin(a) * probeNode.radius * Math.sin(probeNode.tiltX), Math.sin(a) * probeNode.radius * Math.cos(probeNode.tiltX));
				probeNode.mesh.scale.setScalar(1 + Math.sin(elapsed * 1.4 + i * 0.7) * 0.12);

				const positions = connLines[i].geometry.attributes.position;
				const arr = positions.array as Float32Array;
				arr[0] = hub.position.x;
				arr[1] = hub.position.y;
				arr[2] = hub.position.z;
				arr[3] = probeNode.mesh.position.x;
				arr[4] = probeNode.mesh.position.y;
				arr[5] = probeNode.mesh.position.z;
				positions.needsUpdate = true;
			});

			packets.forEach(packet => {
				packet.t = (packet.t + 0.006) % 1;
				const node = probeNodes[packet.nodeIdx];
				packet.mesh.position.lerpVectors(hub.position, node.mesh.position, packet.t);
			});

			renderer.render(scene, camera);
		}

		function animate() {
			raf = requestAnimationFrame(animate);
			renderFrame(clock.getElapsedTime());
		}

		if (reduced) {
			renderer.render(scene, camera);
		} else {
			animate();
		}

		const ro = new ResizeObserver(entries => {
			const { width, height } = entries[0].contentRect;
			if (width > 0 && height > 0) {
				camera.aspect = width / height;
				camera.updateProjectionMatrix();
				renderer.setSize(width, height);
				renderer.render(scene, camera);
			}
		});
		ro.observe(container);

		return () => {
			cancelAnimationFrame(raf);
			ro.disconnect();
			disposeSceneObjects(scene);
			renderer.dispose();
			if (container.contains(renderer.domElement)) container.removeChild(renderer.domElement);
		};
	}, []);

	return <div ref={mountRef} className={styles.probeScene} aria-hidden="true" />;
}
