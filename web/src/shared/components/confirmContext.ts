import { createContext, type ReactNode, useContext } from "react";

export interface ConfirmOptions {
	title: ReactNode;
	message?: ReactNode;
	confirmLabel?: ReactNode;
	cancelLabel?: ReactNode;
	confirmationText?: string;
	confirmationLabel?: ReactNode;
	tone?: "danger" | "default";
}

export type ConfirmFn = (options: ConfirmOptions) => Promise<boolean>;

export interface PromptOptions extends ConfirmOptions {
	defaultValue?: string;
	inputLabel?: ReactNode;
	inputType?: string;
	placeholder?: string;
	validate?: (value: string) => ReactNode;
}

export type PromptFn = (options: PromptOptions) => Promise<string | null>;

export interface AlertDialogOptions {
	title: ReactNode;
	message?: ReactNode;
	confirmLabel?: ReactNode;
	tone?: "danger" | "default";
}

export type AlertDialogFn = (options: AlertDialogOptions) => Promise<void>;

export const ConfirmContext = createContext<ConfirmFn | null>(null);
export const PromptContext = createContext<PromptFn | null>(null);
export const AlertDialogContext = createContext<AlertDialogFn | null>(null);

export function useConfirm() {
	const confirm = useContext(ConfirmContext);

	if (!confirm) {
		throw new Error("useConfirm must be used within ConfirmProvider");
	}

	return confirm;
}

export function usePromptDialog() {
	const prompt = useContext(PromptContext);

	if (!prompt) {
		throw new Error("usePromptDialog must be used within ConfirmProvider");
	}

	return prompt;
}

export function useAlertDialog() {
	const alertDialog = useContext(AlertDialogContext);

	if (!alertDialog) {
		throw new Error("useAlertDialog must be used within ConfirmProvider");
	}

	return alertDialog;
}
