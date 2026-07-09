import { XIcon } from "@phosphor-icons/react/dist/csr/X";
import type { ComponentPropsWithoutRef } from "react";
import styles from "./Toast.module.css";

export type ToastTone = "neutral" | "success" | "warning" | "critical";

export interface ToastViewportProps extends ComponentPropsWithoutRef<"div"> {
	ariaLabel?: string;
}

export interface ToastProps extends ComponentPropsWithoutRef<"div"> {
	tone?: ToastTone;
}

export type ToastTitleProps = ComponentPropsWithoutRef<"strong">;
export type ToastDescriptionProps = ComponentPropsWithoutRef<"p">;

export interface ToastCloseProps extends Omit<ComponentPropsWithoutRef<"button">, "aria-label" | "children"> {
	ariaLabel?: string;
}

export function ToastViewport({ ariaLabel = "Notifications", className, role = "region", ...props }: ToastViewportProps) {
	return <div className={[styles.viewport, className].filter(Boolean).join(" ")} role={role} aria-label={ariaLabel} {...props} />;
}

export function Toast({ tone = "neutral", className, role = "status", ...props }: ToastProps) {
	return <div className={[styles.toast, className].filter(Boolean).join(" ")} data-tone={tone} role={role} {...props} />;
}

export function ToastTitle({ className, ...props }: ToastTitleProps) {
	return <strong className={[styles.title, className].filter(Boolean).join(" ")} {...props} />;
}

export function ToastDescription({ className, ...props }: ToastDescriptionProps) {
	return <p className={[styles.description, className].filter(Boolean).join(" ")} {...props} />;
}

export function ToastClose({ ariaLabel = "Dismiss notification", className, type = "button", ...props }: ToastCloseProps) {
	return (
		<button className={[styles.close, className].filter(Boolean).join(" ")} type={type} aria-label={ariaLabel} title={ariaLabel} {...props}>
			<XIcon className={styles.closeIcon} size="1rem" weight="bold" aria-hidden="true" focusable="false" />
		</button>
	);
}
