import { Drawer } from "@netstamp/ui";
import type { ReactNode } from "react";

interface EditorDrawerProps {
	open: boolean;
	title: ReactNode;
	ariaLabel?: string;
	actions?: ReactNode;
	backLabel?: string;
	children: ReactNode;
	className?: string;
	contentClassName?: string;
	onClose: () => void;
}

export function EditorDrawer({ open, title, ariaLabel, actions, backLabel = "back", children, className, contentClassName, onClose }: EditorDrawerProps) {
	return (
		<Drawer
			open={open}
			title={title}
			aria-label={ariaLabel}
			actions={actions}
			backLabel={backLabel}
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
