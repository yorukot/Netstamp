import { Button } from "@netstamp/ui";
import { type ReactNode, useCallback, useEffect, useId, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { ConfirmContext, type ConfirmFn, type ConfirmOptions } from "./confirmContext";
import styles from "./ConfirmProvider.module.css";

const closeDurationMs = 180;

interface ConfirmState {
	options: ConfirmOptions;
	closing: boolean;
}

export function ConfirmProvider({ children }: { children: ReactNode }) {
	const [state, setState] = useState<ConfirmState | null>(null);
	const resolverRef = useRef<((confirmed: boolean) => void) | null>(null);
	const closeTimeoutRef = useRef<number | null>(null);
	const titleId = useId();
	const messageId = useId();

	const clearCloseTimer = useCallback(() => {
		if (closeTimeoutRef.current) {
			window.clearTimeout(closeTimeoutRef.current);
			closeTimeoutRef.current = null;
		}
	}, []);

	const completeClose = useCallback(() => {
		clearCloseTimer();
		setState(null);
	}, [clearCloseTimer]);

	const close = useCallback(
		(confirmed: boolean) => {
			resolverRef.current?.(confirmed);
			resolverRef.current = null;
			setState(current => (current ? { ...current, closing: true } : current));
			clearCloseTimer();
			closeTimeoutRef.current = window.setTimeout(completeClose, closeDurationMs);
		},
		[clearCloseTimer, completeClose]
	);

	const confirm = useCallback<ConfirmFn>(options => {
		if (resolverRef.current) {
			return Promise.resolve(false);
		}

		setState({ options, closing: false });

		return new Promise(resolve => {
			resolverRef.current = resolve;
		});
	}, []);

	useEffect(() => {
		if (!state) {
			return undefined;
		}

		const previousOverflow = document.body.style.overflow;
		document.body.style.overflow = "hidden";

		function handleKeyDown(event: KeyboardEvent) {
			if (event.key === "Escape") {
				close(false);
			}
		}

		window.addEventListener("keydown", handleKeyDown);

		return () => {
			window.removeEventListener("keydown", handleKeyDown);
			document.body.style.overflow = previousOverflow;
		};
	}, [close, state]);

	useEffect(() => {
		return () => {
			clearCloseTimer();
			resolverRef.current?.(false);
			resolverRef.current = null;
		};
	}, [clearCloseTimer]);

	const dialog =
		state && typeof document !== "undefined"
			? createPortal(
					<div className={styles.overlay} data-closing={state.closing} role="presentation" onMouseDown={() => close(false)}>
						<div className={styles.lineTop} aria-hidden="true" />
						<div className={styles.lineBottom} aria-hidden="true" />
						<section
							className={["ns-cut-frame", styles.dialog].join(" ")}
							role="alertdialog"
							aria-modal="true"
							aria-labelledby={titleId}
							aria-describedby={state.options.message ? messageId : undefined}
							onMouseDown={event => event.stopPropagation()}
						>
							<div className={styles.header}>
								<span>{state.options.tone === "danger" ? "Destructive action" : "Confirm action"}</span>
								<strong id={titleId}>{state.options.title}</strong>
							</div>
							{state.options.message ? (
								<p id={messageId} className={styles.message}>
									{state.options.message}
								</p>
							) : null}
							<div className={styles.actions}>
								<Button type="button" variant="ghost" onClick={() => close(false)} autoFocus>
									{state.options.cancelLabel ?? "Cancel"}
								</Button>
								<Button type="button" variant={state.options.tone === "danger" ? "danger" : "primary"} onClick={() => close(true)}>
									{state.options.confirmLabel ?? "Confirm"}
								</Button>
							</div>
						</section>
					</div>,
					document.body
				)
			: null;

	return (
		<ConfirmContext.Provider value={confirm}>
			{children}
			{dialog}
		</ConfirmContext.Provider>
	);
}
