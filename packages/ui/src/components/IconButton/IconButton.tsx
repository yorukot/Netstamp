import type { ReactNode } from "react";
import { Button, type ButtonProps } from "../Button/Button";
import styles from "./IconButton.module.css";

export interface IconButtonProps extends Omit<ButtonProps, "children" | "size"> {
	children: ReactNode;
	"aria-label": string;
	danger?: boolean;
	size?: "sm" | "md";
}

export function IconButton({ children, className, danger, size = "sm", title, type = "button", variant = "ghost", ...props }: IconButtonProps) {
	const label = props["aria-label"];

	return (
		<Button className={[styles.button, styles[size], danger && styles.danger, className].filter(Boolean).join(" ")} type={type} variant={variant} size="sm" title={title ?? label} {...props}>
			{children}
		</Button>
	);
}
