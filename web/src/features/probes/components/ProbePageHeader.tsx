import { pathForRoute } from "@/routes/routePaths";
import { classNames } from "@/shared/utils/classNames";
import { Button } from "@netstamp/ui";
import { Link } from "react-router-dom";
import styles from "./ProbePageHeader.module.css";
import type { ProbeView } from "./types";

interface ProbePageHeaderProps {
	view: ProbeView;
	projectRef?: string | null;
	onViewChange: (view: ProbeView) => void;
	overlay?: boolean;
}

export function ProbePageHeader({ view, projectRef, onViewChange, overlay = false }: ProbePageHeaderProps) {
	return (
		<header className={classNames(styles.header, overlay && styles.overlay)}>
			<div className={styles.titleArea}>
				<h1>Probe</h1>
				<div className={styles.viewActions} role="group" aria-label="Probe view">
					<Button type="button" size="sm" variant={view === "grid" ? "secondary" : "ghost"} onClick={() => onViewChange("grid")}>
						Grid View
					</Button>
					<Button type="button" size="sm" variant={view === "map" ? "secondary" : "ghost"} onClick={() => onViewChange("map")}>
						Map View
					</Button>
				</div>
			</div>
			<div className={styles.primaryActions}>
				<Button asChild>
					<Link to={pathForRoute("newProbe", { projectRef })}>New probe</Link>
				</Button>
			</div>
		</header>
	);
}
