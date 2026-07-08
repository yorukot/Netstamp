import type { ComponentPropsWithoutRef } from "react";
import styles from "./Spinner.module.css";

export type SpinnerSize = "sm" | "md" | "lg";

export interface SpinnerProps extends ComponentPropsWithoutRef<"span"> {
	label?: string;
	size?: SpinnerSize;
}

export function Spinner({ label = "Loading", size = "md", className, ...props }: SpinnerProps) {
	const classes = [styles.spinner, styles[size], className].filter(Boolean).join(" ");

	return (
		<span className={classes} role="status" aria-label={label} {...props}>
			<span className={styles.mark} aria-hidden="true" />
			<span className={styles.label}>{label}</span>
		</span>
	);
}
