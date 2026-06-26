import type { ComponentPropsWithoutRef, ElementType } from "react";
import styles from "./SpecLabel.module.css";

export type SpecLabelTone = "neutral" | "primary" | "secondary" | "success" | "warning" | "critical" | "muted";
export type SpecLabelSize = "sm" | "md";

interface SpecLabelOwnProps {
	as?: ElementType;
	tone?: SpecLabelTone;
	size?: SpecLabelSize;
}

export type SpecLabelProps<T extends ElementType = "span"> = SpecLabelOwnProps & Omit<ComponentPropsWithoutRef<T>, keyof SpecLabelOwnProps>;

export function SpecLabel<T extends ElementType = "span">({ as, tone = "neutral", size = "sm", className, ...props }: SpecLabelProps<T>) {
	const Comp = as || "span";
	const classes = [styles.label, styles[tone], styles[size], className].filter(Boolean).join(" ");

	return <Comp className={classes} {...props} />;
}
