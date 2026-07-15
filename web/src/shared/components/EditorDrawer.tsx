import { Drawer, type DrawerSide } from "@netstamp/ui";
import type { ReactNode } from "react";

interface EditorDrawerProps {
	open: boolean;
	title: ReactNode;
	ariaLabel?: string;
	actions?: ReactNode;
	children: ReactNode;
	className?: string;
	contentClassName?: string;
	side?: DrawerSide;
	onClose: () => void;
}

export function EditorDrawer({ open, title, ariaLabel, actions, children, className, contentClassName, side, onClose }: EditorDrawerProps) {
	return (
		<Drawer
			open={open}
			title={title}
			aria-label={ariaLabel}
			actions={actions}
			className={className}
			contentClassName={contentClassName}
			side={side}
			onOpenChange={nextOpen => {
				if (!nextOpen) {
					onClose();
				}
			}}
		>
			{children}
		</Drawer>
	);
}
