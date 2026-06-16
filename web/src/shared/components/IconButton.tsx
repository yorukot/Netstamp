import { Button, type ButtonProps } from "@netstamp/ui";
import type { ReactNode } from "react";
import { classNames } from "../utils/classNames";
import styles from "./IconButton.module.css";

interface IconButtonProps extends Omit<ButtonProps, "children" | "size"> {
	children: ReactNode;
	"aria-label": string;
	danger?: boolean;
	size?: "sm" | "md";
}

export function IconButton({ children, className, danger, size = "sm", title, type = "button", variant = "ghost", ...props }: IconButtonProps) {
	const label = props["aria-label"];

	return (
		<Button className={classNames(styles.button, styles[size], danger && styles.danger, className)} type={type} variant={variant} size="sm" title={title ?? label} {...props}>
			{children}
		</Button>
	);
}
