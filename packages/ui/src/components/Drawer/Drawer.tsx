import * as DialogPrimitive from "@radix-ui/react-dialog";
import type { ComponentPropsWithoutRef, ReactNode } from "react";
import { Button } from "../Button/Button";
import styles from "./Drawer.module.css";

export type DrawerSide = "left" | "right";
export type DrawerSize = "sm" | "md" | "lg" | "full";

export interface DrawerProps extends Omit<ComponentPropsWithoutRef<typeof DialogPrimitive.Content>, "title" | "children"> {
	open: boolean;
	title: ReactNode;
	children: ReactNode;
	description?: ReactNode;
	actions?: ReactNode;
	backLabel?: ReactNode;
	closeLabel?: string;
	side?: DrawerSide;
	size?: DrawerSize;
	contentClassName?: string;
	onOpenChange: (open: boolean) => void;
}

function CloseIcon() {
	return (
		<svg className={styles.closeIcon} viewBox="0 0 16 16" aria-hidden="true" focusable="false">
			<path d="M3.76 3.05 8 7.29l4.24-4.24.71.71L8.71 8l4.24 4.24-.71.71L8 8.71l-4.24 4.24-.71-.71L7.29 8 3.05 3.76z" fill="currentColor" />
		</svg>
	);
}

export function Drawer({
	open,
	title,
	description,
	actions,
	backLabel = "back",
	closeLabel = "Close",
	children,
	side = "right",
	size = "md",
	className,
	contentClassName,
	onOpenChange,
	...props
}: DrawerProps) {
	return (
		<DialogPrimitive.Root open={open} onOpenChange={onOpenChange}>
			<DialogPrimitive.Portal>
				<DialogPrimitive.Overlay className={styles.overlay} />
				<DialogPrimitive.Content className={[styles.drawer, className].filter(Boolean).join(" ")} data-side={side} data-size={size} {...props}>
					<header className={styles.header}>
						<div className={styles.copy}>
							<DialogPrimitive.Close className={styles.backButton}>{backLabel}</DialogPrimitive.Close>
							<DialogPrimitive.Title className={styles.title}>{title}</DialogPrimitive.Title>
							{description ? <DialogPrimitive.Description className={styles.description}>{description}</DialogPrimitive.Description> : null}
						</div>
						<div className={styles.actions}>
							{actions}
							<DialogPrimitive.Close asChild>
								<Button className={styles.closeButton} type="button" variant="outline" size="sm" aria-label={closeLabel} title={closeLabel}>
									<CloseIcon />
								</Button>
							</DialogPrimitive.Close>
						</div>
					</header>
					<div className={["ns-scrollbar", styles.content, contentClassName].filter(Boolean).join(" ")}>{children}</div>
				</DialogPrimitive.Content>
			</DialogPrimitive.Portal>
		</DialogPrimitive.Root>
	);
}
