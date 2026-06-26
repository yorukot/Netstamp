import type { ComponentPropsWithoutRef, ElementType, ReactNode } from "react";
import { Surface, type SurfaceTone } from "../Surface/Surface";
import styles from "./SpecCard.module.css";

export type SpecCardTone = Extract<SurfaceTone, "glass" | "matte" | "deep" | "accent" | "danger">;

interface SpecCardOwnProps {
	as?: ElementType;
	tone?: SpecCardTone;
	active?: boolean;
	eyebrow?: ReactNode;
	title: ReactNode;
	description?: ReactNode;
	meta?: ReactNode;
	icon?: ReactNode;
	actions?: ReactNode;
}

export type SpecCardProps<T extends ElementType = "article"> = SpecCardOwnProps & Omit<ComponentPropsWithoutRef<T>, keyof SpecCardOwnProps | "title">;

export function SpecCard<T extends ElementType = "article">({ as, tone = "glass", active = false, eyebrow, title, description, meta, icon, actions, className, children, ...props }: SpecCardProps<T>) {
	const classes = [styles.card, active && styles.active, className].filter(Boolean).join(" ");

	return (
		<Surface as={as || "article"} tone={tone} frameSize="lg" padding="md" className={classes} data-active={active || undefined} {...props}>
			<div className={styles.header}>
				<div className={styles.copy}>
					{eyebrow ? <span className={styles.eyebrow}>{eyebrow}</span> : null}
					<strong className={styles.title}>{title}</strong>
				</div>
				{icon ? <span className={styles.icon}>{icon}</span> : null}
			</div>
			{description ? <p className={styles.description}>{description}</p> : null}
			{children ? <div className={styles.body}>{children}</div> : null}
			{meta || actions ? (
				<div className={styles.footer}>
					{meta ? <span className={styles.meta}>{meta}</span> : null}
					{actions ? <span className={styles.actions}>{actions}</span> : null}
				</div>
			) : null}
		</Surface>
	);
}
