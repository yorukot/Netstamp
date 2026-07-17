import { AlertDialogContent, AlertDialogDescription, AlertDialogOverlay, AlertDialogPortal, AlertDialogRoot, AlertDialogTitle, Button, Input } from "@netstamp/ui";
import { type FormEvent, type ReactNode, useCallback, useEffect, useId, useRef, useState } from "react";
import {
	AlertDialogContext,
	type AlertDialogFn,
	type AlertDialogOptions,
	ChoiceDialogContext,
	type ChoiceDialogFn,
	type ChoiceDialogOptions,
	ConfirmContext,
	type ConfirmFn,
	type ConfirmOptions,
	PromptContext,
	type PromptFn,
	type PromptOptions
} from "./confirmContext";
import styles from "./ConfirmProvider.module.css";

const closeDurationMs = 180;

interface BaseDialogState {
	closing: boolean;
}

interface ConfirmState extends BaseDialogState {
	kind: "confirm";
	inputValue: string;
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

interface ChoiceState extends BaseDialogState {
	kind: "choice";
	options: ChoiceDialogOptions;
}

type DialogState = AlertState | ChoiceState | ConfirmState | PromptState;
type DialogResult = boolean | string | null | undefined;

export function ConfirmProvider({ children }: { children: ReactNode }) {
	const [state, setState] = useState<DialogState | null>(null);
	const resolverRef = useRef<((result: DialogResult) => void) | null>(null);
	const resultRef = useRef<DialogResult>(undefined);
	const closeTimeoutRef = useRef<number | null>(null);
	const confirmationInputRef = useRef<HTMLInputElement | null>(null);
	const titleId = useId();
	const messageId = useId();
	const inputId = useId();
	const inputErrorId = useId();
	const confirmationPromptId = useId();

	const clearCloseTimer = useCallback(() => {
		if (closeTimeoutRef.current) {
			window.clearTimeout(closeTimeoutRef.current);
			closeTimeoutRef.current = null;
		}
	}, []);

	const completeClose = useCallback(() => {
		clearCloseTimer();
		const resolve = resolverRef.current;
		const result = resultRef.current;
		resolverRef.current = null;
		resultRef.current = undefined;
		setState(null);
		resolve?.(result);
	}, [clearCloseTimer]);

	const close = useCallback(
		(result: DialogResult) => {
			resultRef.current = result;
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

		setState({ kind: "confirm", options, closing: false, inputValue: "" });

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

	const choiceDialog = useCallback<ChoiceDialogFn>(options => {
		if (resolverRef.current || options.choices.length === 0) {
			return Promise.resolve(null);
		}

		setState({ kind: "choice", options, closing: false });

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

		return () => {
			document.body.style.overflow = previousOverflow;
		};
	}, [state]);

	useEffect(() => {
		return () => {
			clearCloseTimer();
			resolverRef.current?.(false);
			resolverRef.current = null;
			resultRef.current = undefined;
		};
	}, [clearCloseTimer]);

	function cancelDialog() {
		if (!state) {
			return;
		}

		if (state.kind === "prompt" || state.kind === "choice") {
			close(null);
			return;
		}

		close(state.kind === "confirm" ? false : undefined);
	}

	function updatePromptValue(value: string) {
		setState(current => (current?.kind === "prompt" ? { ...current, inputValue: value, inputError: undefined } : current));
	}

	function updateConfirmationValue(value: string) {
		setState(current => (current?.kind === "confirm" ? { ...current, inputValue: value } : current));
	}

	function submitDialog(event: FormEvent) {
		event.preventDefault();

		if (!state) {
			return;
		}
		if (state.kind === "choice") {
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

		if (state.kind === "confirm" && state.options.confirmationText !== undefined && state.inputValue !== state.options.confirmationText) {
			return;
		}

		close(state.kind === "confirm");
	}

	function handleOpenChange(nextOpen: boolean) {
		if (!nextOpen && state && !state.closing) {
			cancelDialog();
		}
	}

	const message = state?.options.message;
	const tone = state?.options.tone ?? "default";
	const eyebrow = state ? dialogEyebrow(state) : "";
	const confirmLabel = state?.kind === "choice" ? undefined : (state?.options.confirmLabel ?? (state?.kind === "alert" ? "OK" : "Confirm"));
	const confirmationText = state?.kind === "confirm" ? state.options.confirmationText : undefined;
	const confirmationRequired = confirmationText !== undefined;
	const confirmationMatches = !confirmationRequired || (state?.kind === "confirm" && state.inputValue === confirmationText);
	const descriptionIds = [message ? messageId : "", confirmationRequired ? confirmationPromptId : ""].filter(Boolean).join(" ") || undefined;

	const dialog =
		state && typeof document !== "undefined" ? (
			<AlertDialogRoot open={Boolean(state)} onOpenChange={handleOpenChange}>
				<AlertDialogPortal>
					<AlertDialogOverlay className={styles.overlay} data-closing={state.closing} onMouseDown={cancelDialog}>
						<AlertDialogContent
							asChild
							aria-describedby={descriptionIds}
							onOpenAutoFocus={event => {
								if (confirmationRequired) {
									event.preventDefault();
									confirmationInputRef.current?.focus();
								}
							}}
						>
							<form className={styles.dialog} data-tone={tone} onSubmit={submitDialog} onMouseDown={event => event.stopPropagation()}>
								<div className={styles.header}>
									<span>{eyebrow}</span>
									<AlertDialogTitle asChild>
										<strong id={titleId} className="ns-title">
											{state.options.title}
										</strong>
									</AlertDialogTitle>
								</div>
								{message ? (
									<AlertDialogDescription asChild>
										<p id={messageId} className={styles.message}>
											{message}
										</p>
									</AlertDialogDescription>
								) : null}
								{state.kind === "confirm" && confirmationText !== undefined ? (
									<div className={styles.confirmation}>
										<p id={confirmationPromptId} className={styles.confirmationPrompt}>
											Type{" "}
											<button className={styles.confirmationText} type="button" title="Fill the confirmation input" onClick={() => updateConfirmationValue(confirmationText)}>
												{confirmationText}
											</button>{" "}
											to confirm.
										</p>
										<div className={styles.field}>
											<label htmlFor={inputId}>{state.options.confirmationLabel ?? "Name"}</label>
											<Input
												ref={confirmationInputRef}
												id={inputId}
												type="text"
												value={state.inputValue}
												aria-describedby={confirmationPromptId}
												autoComplete="off"
												autoFocus
												spellCheck={false}
												onChange={event => updateConfirmationValue(event.currentTarget.value)}
											/>
										</div>
									</div>
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
								{state.kind === "choice" ? (
									<div className={styles.choices}>
										{state.options.choices.map(choice => (
											<Button key={choice.value} type="button" variant="outline" onClick={() => close(choice.value)}>
												{choice.label}
											</Button>
										))}
									</div>
								) : null}
								<div className={styles.actions}>
									{state.kind !== "alert" ? (
										<Button type="button" variant="ghost" onClick={cancelDialog} autoFocus={state.kind === "confirm" && !confirmationRequired}>
											{state.options.cancelLabel ?? "Cancel"}
										</Button>
									) : null}
									{state.kind !== "choice" ? (
										<Button type="submit" variant={tone === "danger" ? "danger" : "primary"} disabled={!confirmationMatches} autoFocus={state.kind === "alert"}>
											{confirmLabel}
										</Button>
									) : null}
								</div>
							</form>
						</AlertDialogContent>
					</AlertDialogOverlay>
				</AlertDialogPortal>
			</AlertDialogRoot>
		) : null;

	return (
		<ConfirmContext.Provider value={confirm}>
			<PromptContext.Provider value={prompt}>
				<ChoiceDialogContext.Provider value={choiceDialog}>
					<AlertDialogContext.Provider value={alertDialog}>
						{children}
						{dialog}
					</AlertDialogContext.Provider>
				</ChoiceDialogContext.Provider>
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
	if (state.kind === "choice") {
		return "Authentication required";
	}

	return state.options.tone === "danger" ? "Destructive action" : "Confirm action";
}
