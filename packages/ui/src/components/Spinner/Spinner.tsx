import type { ComponentPropsWithoutRef } from "react";
import styles from "./Spinner.module.css";

export type SpinnerSize = "sm" | "md" | "lg";
export type SpinnerLayout = "inline" | "compact" | "panel" | "page";

export interface SpinnerProps extends ComponentPropsWithoutRef<"span"> {
	label?: string;
	layout?: SpinnerLayout;
	size?: SpinnerSize;
}

export function Spinner({ label = "Loading", layout = "inline", size = "md", className, ...props }: SpinnerProps) {
	const classes = [styles.spinner, styles[size], styles[`layout${layout}`], className].filter(Boolean).join(" ");

	return (
		<span className={classes} role="status" aria-label={label} {...props}>
			<span className={styles.mark} aria-hidden="true" />
			<span className={styles.label}>{label}</span>
		</span>
	);
}
