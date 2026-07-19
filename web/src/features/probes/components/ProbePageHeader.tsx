import { pathForRoute } from "@/routes/routePaths";
import { classNames } from "@/shared/utils/classNames";
import { Button, SegmentedControl } from "@netstamp/ui";
import { useTranslation } from "react-i18next";
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
	const { t } = useTranslation("probes");
	return (
		<header className={classNames(styles.header, overlay && styles.overlay)}>
			<div className={styles.titleArea}>
				<h1>{t("title")}</h1>
				<SegmentedControl
					className={styles.viewActions}
					size="sm"
					ariaLabel={t("views.aria")}
					value={view}
					options={[
						{ value: "grid", label: t("views.grid") },
						{ value: "map", label: t("views.map") }
					]}
					onValueChange={nextView => onViewChange(nextView as ProbeView)}
				/>
			</div>
			<div className={styles.primaryActions}>
				<Button asChild>
					<Link to={pathForRoute("newProbe", { projectRef })}>{t("new")}</Link>
				</Button>
			</div>
		</header>
	);
}
