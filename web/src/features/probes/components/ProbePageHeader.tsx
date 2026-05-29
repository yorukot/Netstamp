import { pathForRoute } from "@/routes/routePaths";
import { classNames } from "@/shared/utils/classNames";
import { Button } from "@netstamp/ui";
import { Link } from "react-router-dom";
import styles from "./ProbePageHeader.module.css";
import type { ProbeView } from "./types";

interface ProbePageHeaderProps {
	view: ProbeView;
	onViewChange: (view: ProbeView) => void;
	overlay?: boolean;
}

export function ProbePageHeader({ view, onViewChange, overlay = false }: ProbePageHeaderProps) {
	return (
		<header className={classNames(styles.header, overlay && styles.overlay)}>
			<div>
				<h1>Probe Fleet</h1>
			</div>
			<div className={styles.actions}>
				<Button type="button" size="sm" variant={view === "grid" ? "secondary" : "ghost"} onClick={() => onViewChange("grid")}>
					Grid View
				</Button>
				<Button type="button" size="sm" variant={view === "map" ? "secondary" : "ghost"} onClick={() => onViewChange("map")}>
					Map View
				</Button>
				<Button className={styles.createButton} size="sm" asChild>
					<Link to={pathForRoute("newProbe")}>Create Probe</Link>
				</Button>
			</div>
		</header>
	);
}
