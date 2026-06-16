import { X } from "@phosphor-icons/react";
import type { ComponentPropsWithoutRef } from "react";
import styles from "./CloseButton.module.css";

interface CloseButtonProps extends Omit<ComponentPropsWithoutRef<"button">, "children"> {
	ariaLabel?: string;
}

export function CloseButton({ ariaLabel = "Close panel", className, title, type = "button", ...props }: CloseButtonProps) {
	const classes = [styles.button, className].filter(Boolean).join(" ");

	return (
		<button type={type} className={classes} aria-label={ariaLabel} title={title ?? ariaLabel} {...props}>
			<X size={16} weight="bold" aria-hidden="true" focusable="false" />
		</button>
	);
}
