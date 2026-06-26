import type { ComponentPropsWithoutRef, ReactNode } from "react";
import styles from "./KeyValueRow.module.css";

export type KeyValueRowTone = "neutral" | "primary" | "secondary" | "success" | "warning" | "critical";

export interface KeyValueRowProps extends ComponentPropsWithoutRef<"div"> {
	label: ReactNode;
	value: ReactNode;
	meta?: ReactNode;
	tone?: KeyValueRowTone;
}

export function KeyValueRow({ label, value, meta, tone = "neutral", className, ...props }: KeyValueRowProps) {
	return (
		<div className={[styles.row, styles[tone], className].filter(Boolean).join(" ")} {...props}>
			<span className={styles.label}>{label}</span>
			<strong className={styles.value}>{value}</strong>
			{meta ? <span className={styles.meta}>{meta}</span> : null}
		</div>
	);
}
