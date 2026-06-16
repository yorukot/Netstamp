import { Slot } from "@radix-ui/react-slot";
import type { ComponentPropsWithoutRef } from "react";
import styles from "./Button.module.css";

export type ButtonVariant = "primary" | "secondary" | "outline" | "ghost" | "danger" | "plain";
export type ButtonSize = "sm" | "md" | "lg" | "xl";

export interface ButtonProps extends ComponentPropsWithoutRef<"button"> {
	variant?: ButtonVariant;
	size?: ButtonSize;
	/** Render as child element (polymorphic) */
	asChild?: boolean;
}

export function Button({ variant = "primary", size = "md", asChild = false, className, ...props }: ButtonProps) {
	const Comp = asChild ? Slot : "button";
	const classes = ["ns-frame", styles.btn, styles[variant], styles[size], className].filter(Boolean).join(" ");

	return <Comp className={classes} {...props} />;
}
