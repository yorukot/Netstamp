import type { ButtonHTMLAttributes } from "react";
import styles from "./DisclosureToggle.module.css";

export type DisclosureToggleSize = "sm" | "md";

export interface DisclosureToggleProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, "children"> {
	open: boolean;
	label: string;
	size?: DisclosureToggleSize;
}

export function DisclosureToggle({ open, label, size = "sm", className, type = "button", ...props }: DisclosureToggleProps) {
	return (
		<button className={[styles.toggle, styles[size], open && styles.open, className].filter(Boolean).join(" ")} type={type} aria-expanded={open} aria-label={label} title={label} {...props}>
			<span aria-hidden="true" />
		</button>
	);
}
