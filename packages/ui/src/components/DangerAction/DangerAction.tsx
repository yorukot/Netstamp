import type { ComponentPropsWithoutRef, ReactNode } from "react";
import styles from "./DangerAction.module.css";

export interface DangerActionProps extends Omit<ComponentPropsWithoutRef<"article">, "title"> {
	action: ReactNode;
	description: ReactNode;
	descriptionId?: string;
	title: ReactNode;
}

export function DangerAction({ action, className, description, descriptionId, title, ...props }: DangerActionProps) {
	return (
		<article className={[styles.action, className].filter(Boolean).join(" ")} {...props}>
			<div className={styles.copy}>
				<h3>{title}</h3>
				<p id={descriptionId}>{description}</p>
			</div>
			<div className={styles.control}>{action}</div>
		</article>
	);
}
