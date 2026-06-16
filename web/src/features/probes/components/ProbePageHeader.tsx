import { pathForRoute } from "@/routes/routePaths";
import { classNames } from "@/shared/utils/classNames";
import { Button, SegmentedControl } from "@netstamp/ui";
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
				<SegmentedControl
					className={styles.viewActions}
					size="sm"
					ariaLabel="Probe view"
					value={view}
					options={[
						{ value: "grid", label: "Grid View" },
						{ value: "map", label: "Map View" }
					]}
					onValueChange={nextView => onViewChange(nextView as ProbeView)}
				/>
			</div>
			<div className={styles.primaryActions}>
				<Button asChild>
					<Link to={pathForRoute("newProbe", { projectRef })}>New probe</Link>
				</Button>
			</div>
		</header>
	);
}
