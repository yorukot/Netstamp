import type { ComponentPropsWithoutRef, ReactNode } from "react";
import { Badge, type BadgeTone } from "../Badge/Badge";
import { Surface } from "../Surface/Surface";
import styles from "./MetricTile.module.css";

export interface MetricTileProps extends ComponentPropsWithoutRef<"article"> {
	label: ReactNode;
	value: ReactNode;
	detail?: ReactNode;
	description?: ReactNode;
	tone?: BadgeTone;
	trend?: ReactNode;
}

export function MetricTile({ label, value, detail, description, tone = "neutral", trend, className, ...props }: MetricTileProps) {
	return (
		<Surface as="article" tone="glass" frameSize="lg" padding="sm" className={[styles.tile, className].filter(Boolean).join(" ")} {...props}>
			<div className={styles.header}>
				<span className={styles.label}>{label}</span>
				{trend ? <span className={styles.trend}>{trend}</span> : null}
			</div>
			<strong className={styles.value}>{value}</strong>
			{description ? <span className={styles.description}>{description}</span> : null}
			{detail ? <Badge tone={tone}>{detail}</Badge> : null}
		</Surface>
	);
}
