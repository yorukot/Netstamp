import type { ComponentPropsWithoutRef, ElementType, ReactNode } from "react";
import styles from "./SelectableRow.module.css";

export type SelectableRowTone = "neutral" | "primary" | "secondary" | "danger";
export type SelectableRowDensity = "normal" | "compact";

interface SelectableRowOwnProps {
	as?: ElementType;
	active?: boolean;
	disabled?: boolean;
	tone?: SelectableRowTone;
	density?: SelectableRowDensity;
	leading?: ReactNode;
	title: ReactNode;
	description?: ReactNode;
	meta?: ReactNode;
	trailing?: ReactNode;
}

export type SelectableRowProps<T extends ElementType = "button"> = SelectableRowOwnProps & Omit<ComponentPropsWithoutRef<T>, keyof SelectableRowOwnProps | "title">;

export function SelectableRow<T extends ElementType = "button">({
	as,
	active = false,
	disabled = false,
	tone = "neutral",
	density = "normal",
	leading,
	title,
	description,
	meta,
	trailing,
	className,
	...props
}: SelectableRowProps<T>) {
	const Comp = as || "button";
	const classes = [styles.row, styles[tone], styles[density], active && styles.active, disabled && styles.disabled, className].filter(Boolean).join(" ");
	const rowProps = {
		"aria-current": active ? "page" : undefined,
		"aria-disabled": disabled || undefined,
		"data-active": active || undefined,
		"data-disabled": disabled || undefined,
		disabled: Comp === "button" ? disabled : undefined,
		...props
	};

	return (
		<Comp className={classes} {...rowProps}>
			{leading ? <span className={styles.leading}>{leading}</span> : null}
			<span className={styles.copy}>
				<span className={["ns-title", styles.title].join(" ")}>{title}</span>
				{description ? <span className={styles.description}>{description}</span> : null}
			</span>
			{meta ? <span className={styles.meta}>{meta}</span> : null}
			{trailing ? <span className={styles.trailing}>{trailing}</span> : null}
		</Comp>
	);
}
