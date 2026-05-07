import type { ReactNode } from "react";
import { Badge, type BadgeTone } from "../Badge/Badge";
import { Surface } from "../Surface/Surface";
import styles from "./MetricCard.module.css";

export interface MetricCardProps {
	label: ReactNode;
	value: ReactNode;
	detail?: ReactNode;
	tone?: BadgeTone;
	className?: string;
}

export function MetricCard({ label, value, detail, tone = "accent", className }: MetricCardProps) {
	const classes = [styles.card, className].filter(Boolean).join(" ");

	return (
		<Surface as="article" tone="glass" cut="lg" padding="md" className={classes}>
			<span className={styles.label}>{label}</span>
			<strong>{value}</strong>
			{detail ? <Badge tone={tone}>{detail}</Badge> : null}
			<span className={styles.corner} aria-hidden="true" />
		</Surface>
	);
}
