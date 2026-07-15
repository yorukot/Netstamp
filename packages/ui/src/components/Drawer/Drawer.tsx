import { XIcon } from "@phosphor-icons/react/dist/csr/X";
import * as DialogPrimitive from "@radix-ui/react-dialog";
import { useState, type ComponentPropsWithoutRef, type ReactNode } from "react";
import { Button } from "../Button/Button";
import { wasOverlayPointerDownHandled } from "../overlayInteractions";
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

export function Drawer({
	open,
	title,
	description,
	actions,
	closeLabel = "Close",
	children,
	side = "right",
	size = "md",
	className,
	contentClassName,
	onOpenChange,
	onAnimationEnd,
	onPointerDownOutside,
	...props
}: DrawerProps) {
	const [closeRequested, setCloseRequested] = useState(false);
	const renderedOpen = open && !closeRequested;

	function handleOpenChange(nextOpen: boolean) {
		if (!nextOpen) {
			setCloseRequested(true);
			return;
		}

		setCloseRequested(false);
		onOpenChange(true);
	}

	return (
		<DialogPrimitive.Root open={renderedOpen} onOpenChange={handleOpenChange}>
			<DialogPrimitive.Portal>
				<DialogPrimitive.Overlay className={styles.overlay} />
				<DialogPrimitive.Content
					className={[styles.drawer, className].filter(Boolean).join(" ")}
					data-side={side}
					data-size={size}
					onAnimationEnd={event => {
						onAnimationEnd?.(event);

						if (!closeRequested || event.target !== event.currentTarget || event.currentTarget.dataset.state !== "closed") {
							return;
						}

						setCloseRequested(false);
						onOpenChange(false);
					}}
					onPointerDownOutside={event => {
						onPointerDownOutside?.(event);

						if (wasOverlayPointerDownHandled(event)) {
							event.preventDefault();
						}
					}}
					{...props}
				>
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
