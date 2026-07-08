import { Drawer } from "@netstamp/ui";
import type { ReactNode } from "react";

interface EditorDrawerProps {
	open: boolean;
	title: ReactNode;
	ariaLabel?: string;
	actions?: ReactNode;
	children: ReactNode;
	className?: string;
	contentClassName?: string;
	onClose: () => void;
}

export function EditorDrawer({ open, title, ariaLabel, actions, children, className, contentClassName, onClose }: EditorDrawerProps) {
	return (
		<Drawer
			open={open}
			title={title}
			aria-label={ariaLabel}
			actions={actions}
			className={className}
			contentClassName={contentClassName}
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
