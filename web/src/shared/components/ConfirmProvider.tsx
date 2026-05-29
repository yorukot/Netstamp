import { Button, Input } from "@netstamp/ui";
import { type FormEvent, type ReactNode, useCallback, useEffect, useId, useRef, useState } from "react";
import { createPortal } from "react-dom";
import {
	AlertDialogContext,
	type AlertDialogFn,
	type AlertDialogOptions,
	ConfirmContext,
	type ConfirmFn,
	type ConfirmOptions,
	PromptContext,
	type PromptFn,
	type PromptOptions
} from "./confirmContext";
import styles from "./ConfirmProvider.module.css";

const closeDurationMs = 260;

interface BaseDialogState {
	closing: boolean;
}

interface ConfirmState extends BaseDialogState {
	kind: "confirm";
	options: ConfirmOptions;
}

interface PromptState extends BaseDialogState {
	kind: "prompt";
	inputError?: ReactNode;
	inputValue: string;
	options: PromptOptions;
}

interface AlertState extends BaseDialogState {
	kind: "alert";
	options: AlertDialogOptions;
}

type DialogState = AlertState | ConfirmState | PromptState;
type DialogResult = boolean | string | null | undefined;

export function ConfirmProvider({ children }: { children: ReactNode }) {
	const [state, setState] = useState<DialogState | null>(null);
	const resolverRef = useRef<((result: DialogResult) => void) | null>(null);
	const closeTimeoutRef = useRef<number | null>(null);
	const titleId = useId();
	const messageId = useId();
	const inputId = useId();
	const inputErrorId = useId();

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
		(result: DialogResult) => {
			resolverRef.current?.(result);
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

		setState({ kind: "confirm", options, closing: false });

		return new Promise(resolve => {
			resolverRef.current = result => resolve(result === true);
		});
	}, []);

	const prompt = useCallback<PromptFn>(options => {
		if (resolverRef.current) {
			return Promise.resolve(null);
		}

		setState({
			kind: "prompt",
			options,
			closing: false,
			inputValue: options.defaultValue ?? ""
		});

		return new Promise(resolve => {
			resolverRef.current = result => resolve(typeof result === "string" ? result : null);
		});
	}, []);

	const alertDialog = useCallback<AlertDialogFn>(options => {
		if (resolverRef.current) {
			return Promise.resolve();
		}

		setState({ kind: "alert", options, closing: false });

		return new Promise(resolve => {
			resolverRef.current = () => resolve();
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

	function cancelDialog() {
		if (!state) {
			return;
		}

		if (state.kind === "prompt") {
			close(null);
			return;
		}

		close(state.kind === "confirm" ? false : undefined);
	}

	function updatePromptValue(value: string) {
		setState(current => (current?.kind === "prompt" ? { ...current, inputValue: value, inputError: undefined } : current));
	}

	function submitDialog(event: FormEvent) {
		event.preventDefault();

		if (!state) {
			return;
		}

		if (state.kind === "prompt") {
			const inputError = state.options.validate?.(state.inputValue);

			if (inputError) {
				setState(current => (current?.kind === "prompt" ? { ...current, inputError } : current));
				return;
			}

			close(state.inputValue);
			return;
		}

		close(state.kind === "confirm");
	}

	const message = state?.options.message;
	const tone = state?.options.tone ?? "default";
	const eyebrow = state ? dialogEyebrow(state) : "";
	const confirmLabel = state?.options.confirmLabel ?? (state?.kind === "alert" ? "OK" : "Confirm");

	const dialog =
		state && typeof document !== "undefined"
			? createPortal(
					<div className={styles.overlay} data-closing={state.closing} role="presentation" onMouseDown={cancelDialog}>
						<div className={styles.axis} aria-hidden="true" />
						<div className={styles.curtainTop} aria-hidden="true" />
						<div className={styles.curtainBottom} aria-hidden="true" />
						<form
							className={["ns-cut-frame", styles.dialog].join(" ")}
							data-tone={tone}
							onSubmit={submitDialog}
							role="alertdialog"
							aria-modal="true"
							aria-labelledby={titleId}
							aria-describedby={message ? messageId : undefined}
							onMouseDown={event => event.stopPropagation()}
						>
							<div className={styles.header}>
								<span>{eyebrow}</span>
								<strong id={titleId}>{state.options.title}</strong>
							</div>
							{message ? (
								<p id={messageId} className={styles.message}>
									{message}
								</p>
							) : null}
							{state.kind === "prompt" ? (
								<div className={styles.field}>
									<label htmlFor={inputId}>{state.options.inputLabel ?? "Value"}</label>
									<Input
										id={inputId}
										type={state.options.inputType ?? "text"}
										value={state.inputValue}
										placeholder={state.options.placeholder}
										invalid={Boolean(state.inputError)}
										aria-describedby={state.inputError ? inputErrorId : undefined}
										autoFocus
										onChange={event => updatePromptValue(event.currentTarget.value)}
									/>
									{state.inputError ? (
										<p id={inputErrorId} className={styles.inputError}>
											{state.inputError}
										</p>
									) : null}
								</div>
							) : null}
							<div className={styles.actions}>
								{state.kind !== "alert" ? (
									<Button type="button" variant="ghost" onClick={cancelDialog} autoFocus={state.kind === "confirm"}>
										{state.options.cancelLabel ?? "Cancel"}
									</Button>
								) : null}
								<Button type="submit" variant={tone === "danger" ? "danger" : "primary"} autoFocus={state.kind === "alert"}>
									{confirmLabel}
								</Button>
							</div>
						</form>
					</div>,
					document.body
				)
			: null;

	return (
		<ConfirmContext.Provider value={confirm}>
			<PromptContext.Provider value={prompt}>
				<AlertDialogContext.Provider value={alertDialog}>
					{children}
					{dialog}
				</AlertDialogContext.Provider>
			</PromptContext.Provider>
		</ConfirmContext.Provider>
	);
}

function dialogEyebrow(state: DialogState) {
	if (state.kind === "alert") {
		return state.options.tone === "danger" ? "Action required" : "Notice";
	}

	if (state.kind === "prompt") {
		return state.options.tone === "danger" ? "Input required" : "Input required";
	}

	return state.options.tone === "danger" ? "Destructive action" : "Confirm action";
}
