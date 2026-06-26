import type { ComponentPropsWithoutRef, ElementType, ReactNode } from "react";
import { Surface, type SurfaceTone } from "../Surface/Surface";
import styles from "./Panel.module.css";

export type PanelTone = Extract<SurfaceTone, "glass" | "matte" | "deep">;

export interface PanelProps extends Omit<ComponentPropsWithoutRef<"section">, "title"> {
	as?: ElementType;
	tone?: PanelTone;
	eyebrow?: ReactNode;
	title?: ReactNode;
	summary?: ReactNode;
	actions?: ReactNode;
	footer?: ReactNode;
	padded?: boolean;
}

export function Panel({ as: Comp = "section", tone = "glass", eyebrow, title, summary, actions, footer, padded = true, className, children, ...props }: PanelProps) {
	const classes = [styles.panel, className].filter(Boolean).join(" ");

	return (
		<Surface as={Comp} tone={tone} frameSize="lg" padding={padded ? "md" : "none"} className={classes} {...props}>
			{eyebrow || title || summary || actions ? (
				<div className={styles.header}>
					<div className={styles.copy}>
						{eyebrow ? <span className={styles.eyebrow}>{eyebrow}</span> : null}
						{title ? <h3>{title}</h3> : null}
						{summary ? <p>{summary}</p> : null}
					</div>
					{actions ? <div className={styles.actions}>{actions}</div> : null}
				</div>
			) : null}
			{children ? <div className={styles.body}>{children}</div> : null}
			{footer ? <div className={styles.footer}>{footer}</div> : null}
		</Surface>
	);
}
