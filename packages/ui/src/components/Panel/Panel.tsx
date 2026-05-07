import * as Separator from "@radix-ui/react-separator";
import type { ComponentPropsWithoutRef, ElementType, ReactNode } from "react";
import { Surface, type SurfaceTone } from "../Surface/Surface";
import styles from "./Panel.module.css";

export type PanelTone = Extract<SurfaceTone, "glass" | "matte" | "deep">;

export interface PanelProps extends Omit<ComponentPropsWithoutRef<"section">, "title"> {
	as?: ElementType;
	tone?: PanelTone;
	eyebrow?: ReactNode;
	title?: ReactNode;
	actions?: ReactNode;
	padded?: boolean;
}

export function Panel({ as: Comp = "section", tone = "glass", eyebrow, title, actions, padded = true, className, children, ...props }: PanelProps) {
	const classes = [styles.panel, className].filter(Boolean).join(" ");

	return (
		<Surface as={Comp} tone={tone} cut="lg" padding={padded ? "md" : "none"} className={classes} {...props}>
			{eyebrow || title || actions ? (
				<div className={styles.header}>
					<div className={styles.copy}>
						{eyebrow ? <div className={styles.eyebrow}>{eyebrow}</div> : null}
						{title ? <h3>{title}</h3> : null}
					</div>
					{actions ? <div className={styles.actions}>{actions}</div> : null}
				</div>
			) : null}
			{eyebrow || title || actions ? <Separator.Root className={styles.separator} /> : null}
			{children}
		</Surface>
	);
}
