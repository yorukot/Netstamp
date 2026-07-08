import type { ComponentPropsWithoutRef, ReactNode } from "react";
import styles from "./EmptyState.module.css";

export interface EmptyStateProps extends Omit<ComponentPropsWithoutRef<"div">, "title"> {
	title: ReactNode;
	description?: ReactNode;
	action?: ReactNode;
}

export function EmptyState({ title, description, action, className, role = "status", ...props }: EmptyStateProps) {
	return (
		<div className={[styles.emptyState, className].filter(Boolean).join(" ")} role={role} {...props}>
			<strong>{title}</strong>
			{description ? <p>{description}</p> : null}
			{action ? <div className={styles.action}>{action}</div> : null}
		</div>
	);
}
