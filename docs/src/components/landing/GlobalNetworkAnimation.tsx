import type { CSSProperties } from "react";
import styles from "./LandingPage.module.css";

export function GlobalNetworkAnimation() {
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
