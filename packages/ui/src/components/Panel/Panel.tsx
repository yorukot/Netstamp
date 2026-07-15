import type { ComponentPropsWithoutRef, ElementType, ReactNode } from "react";
import styles from "./Panel.module.css";

export type PanelTone = "glass" | "matte" | "deep";
export type PanelBodySurface = "default" | "transparent";

export interface PanelProps extends Omit<ComponentPropsWithoutRef<"section">, "title"> {
	as?: ElementType;
	tone?: PanelTone;
	eyebrow?: ReactNode;
	title?: ReactNode;
	summary?: ReactNode;
	actions?: ReactNode;
	actionsClassName?: string;
	footer?: ReactNode;
	padded?: boolean;
	bodySurface?: PanelBodySurface;
	bodyClassName?: string;
}

export function Panel({
	as: Comp = "section",
	tone = "glass",
	eyebrow,
	title,
	summary,
	actions,
	actionsClassName,
	footer,
	padded = true,
	bodySurface = "default",
	className,
	bodyClassName,
	children,
	...props
}: PanelProps) {
	const classes = [styles.panel, className].filter(Boolean).join(" ");
	const bodyClasses = [styles.body, bodyClassName].filter(Boolean).join(" ");
	const actionClasses = [styles.actions, actionsClassName].filter(Boolean).join(" ");

	return (
		<Comp className={classes} data-tone={tone} {...props}>
			{eyebrow || title || summary || actions ? (
				<div className={styles.header}>
					<div className={styles.copy}>
						{eyebrow ? <span className={styles.eyebrow}>{eyebrow}</span> : null}
						{title ? (
							<div className={styles.titleRow}>
								<span className={styles.marker} aria-hidden="true" />
								<h2>{title}</h2>
							</div>
						) : null}
						{summary ? <p>{summary}</p> : null}
					</div>
					{actions ? <div className={actionClasses}>{actions}</div> : null}
				</div>
			) : null}
			{children ? (
				<div className={bodyClasses} data-padded={padded} data-surface={bodySurface}>
					{children}
				</div>
			) : null}
			{footer ? <div className={styles.footer}>{footer}</div> : null}
		</Comp>
	);
}
