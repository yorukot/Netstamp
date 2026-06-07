import type { MouseEvent, ReactNode } from "react";
import { useCallback, useEffect, useId, useRef, useState } from "react";
import { CloseButton } from "./CloseButton";
import styles from "./EditorDrawer.module.css";

const drawerCloseDurationMs = 180;

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
	const titleId = useId();
	const onCloseRef = useRef(onClose);
	const closeTimeoutRef = useRef<number | null>(null);
	const [closing, setClosing] = useState(false);

	const requestClose = useCallback(() => {
		if (closing || closeTimeoutRef.current) {
			return;
		}

		setClosing(true);
		closeTimeoutRef.current = window.setTimeout(() => {
			closeTimeoutRef.current = null;
			setClosing(false);
			onCloseRef.current();
		}, drawerCloseDurationMs);
	}, [closing]);

	useEffect(() => {
		onCloseRef.current = onClose;
	}, [onClose]);

	useEffect(() => {
		if (!open) {
			return;
		}

		const previousOverflow = document.body.style.overflow;
		document.body.style.overflow = "hidden";

		return () => {
			document.body.style.overflow = previousOverflow;
			if (closeTimeoutRef.current) {
				window.clearTimeout(closeTimeoutRef.current);
				closeTimeoutRef.current = null;
			}
		};
	}, [open]);

	useEffect(() => {
		if (!open) {
			return;
		}

		function handleKeyDown(event: KeyboardEvent) {
			if (event.key === "Escape") {
				requestClose();
			}
		}

		window.addEventListener("keydown", handleKeyDown);
		return () => window.removeEventListener("keydown", handleKeyDown);
	}, [open, requestClose]);

	if (!open) {
		return null;
	}

	function handleBackdropClick(event: MouseEvent<HTMLDivElement>) {
		if (event.target === event.currentTarget) {
			requestClose();
		}
	}

	return (
		<div className={[styles.backdrop, closing && styles.backdropClosing].filter(Boolean).join(" ")} onClick={handleBackdropClick}>
			<aside
				className={[styles.drawer, closing && styles.drawerClosing, className].filter(Boolean).join(" ")}
				role="dialog"
				aria-modal="true"
				aria-label={ariaLabel}
				aria-labelledby={ariaLabel ? undefined : titleId}
			>
				<div className={styles.header}>
					<div>
						<button type="button" className={styles.backLink} onClick={requestClose}>
							{backLabel}
						</button>
						<h2 id={titleId}>{title}</h2>
					</div>
					<div className={styles.actions}>
						{actions}
						<CloseButton ariaLabel="Close drawer" onClick={requestClose} />
					</div>
				</div>
				<div className={["ns-scrollbar", styles.content, contentClassName].filter(Boolean).join(" ")}>{children}</div>
			</aside>
		</div>
	);
}
