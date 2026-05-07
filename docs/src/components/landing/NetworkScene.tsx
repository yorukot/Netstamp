import { useEffect, useRef } from "react";
import { BufferGeometry, Clock, Line, LineBasicMaterial, Mesh, MeshBasicMaterial, OctahedronGeometry, PerspectiveCamera, Scene, SphereGeometry, WebGLRenderer } from "three";
import styles from "./LandingPage.module.css";
import { disposeSceneObjects } from "./threeSceneCleanup";

export function NetworkScene() {
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

		const positions: [number, number, number][] = [
			[-1.9, 1.1, 0.3],
			[-0.7, 1.9, -0.3],
			[0.6, 1.6, 0.5],
			[2.0, 0.7, -0.2],
			[2.5, -0.5, 0.4],
			[1.3, -1.7, 0.0],
			[-0.1, -2.0, 0.6],
			[-1.7, -0.9, -0.3],
			[-2.4, 0.1, 0.5],
			[0.2, 0.1, -0.9]
		];

		const nodeGeo = new OctahedronGeometry(0.072, 0);
		const nodes = positions.map(([x, y, z]) => {
			const mesh = new Mesh(nodeGeo, new MeshBasicMaterial({ color: 0xff7a1a }));
			mesh.position.set(x, y, z);
			scene.add(mesh);
			return mesh;
		});

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

		const pktGeo = new SphereGeometry(0.05, 8, 6);
		const packets = connDefs.slice(0, 8).map(([a, b], i) => {
			const mesh = new Mesh(pktGeo, new MeshBasicMaterial({ color: 0xffaa55 }));
			scene.add(mesh);
			return { mesh, from: a, to: b, t: i / 8 };
		});

		const clock = new Clock();
		let raf = 0;
		let visible = false;

		function renderFrame(elapsed: number) {
			scene.rotation.y = elapsed * 0.07;
			scene.rotation.x = Math.sin(elapsed * 0.035) * 0.13;

			packets.forEach(packet => {
				packet.t = (packet.t + 0.0038) % 1;
				packet.mesh.position.lerpVectors(nodes[packet.from].position, nodes[packet.to].position, packet.t);
			});

			nodes.forEach((node, i) => {
				node.scale.setScalar(1 + Math.sin(elapsed * 1.6 + i * 0.65) * 0.11);
			});

			renderer.render(scene, camera);
		}

		function animate() {
			if (!visible) {
				raf = 0;
				return;
			}

			renderFrame(clock.getElapsedTime());
			raf = requestAnimationFrame(animate);
		}

		function startAnimation() {
			if (raf || reduced) return;

			visible = true;
			animate();
		}

		function stopAnimation() {
			visible = false;
			if (raf) {
				cancelAnimationFrame(raf);
				raf = 0;
			}
		}

		let observer: IntersectionObserver | undefined;
		if (reduced) {
			renderer.render(scene, camera);
		} else if ("IntersectionObserver" in window) {
			observer = new IntersectionObserver(
				entries => {
					if (entries[0]?.isIntersecting) {
						startAnimation();
					} else {
						stopAnimation();
					}
				},
				{ rootMargin: "160px 0px" }
			);
			observer.observe(container);
		} else {
			startAnimation();
		}

		const ro = new ResizeObserver(entries => {
			const { width, height } = entries[0].contentRect;
			if (width > 0 && height > 0) {
				camera.aspect = width / height;
				camera.updateProjectionMatrix();
				renderer.setSize(width, height);
				if (visible || reduced) renderer.render(scene, camera);
			}
		});
		ro.observe(container);

		return () => {
			observer?.disconnect();
			stopAnimation();
			ro.disconnect();
			disposeSceneObjects(scene);
			renderer.dispose();
			if (container.contains(renderer.domElement)) container.removeChild(renderer.domElement);
		};
	}, []);

	return <div ref={mountRef} className={styles.networkScene} aria-hidden="true" />;
}
