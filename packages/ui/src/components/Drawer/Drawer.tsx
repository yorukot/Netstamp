import { XIcon } from "@phosphor-icons/react/dist/csr/X";
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
	closeLabel?: string;
	side?: DrawerSide;
	size?: DrawerSize;
	contentClassName?: string;
	onOpenChange: (open: boolean) => void;
}

export function Drawer({ open, title, description, actions, closeLabel = "Close", children, side = "right", size = "md", className, contentClassName, onOpenChange, ...props }: DrawerProps) {
	return (
		<DialogPrimitive.Root open={open} onOpenChange={onOpenChange}>
			<DialogPrimitive.Portal>
				<DialogPrimitive.Overlay className={styles.overlay} />
				<DialogPrimitive.Content className={[styles.drawer, className].filter(Boolean).join(" ")} data-side={side} data-size={size} {...props}>
					<header className={styles.header}>
						<div className={styles.copy}>
							<DialogPrimitive.Title className={styles.title}>{title}</DialogPrimitive.Title>
							{description ? <DialogPrimitive.Description className={styles.description}>{description}</DialogPrimitive.Description> : null}
						</div>
						<div className={styles.actions}>
							{actions}
							<DialogPrimitive.Close asChild>
								<Button className={styles.closeButton} type="button" variant="outline" size="sm" aria-label={closeLabel} title={closeLabel}>
									<XIcon className={styles.closeIcon} size={16} weight="bold" aria-hidden="true" focusable="false" />
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
