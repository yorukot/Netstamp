import { createContext, type ReactNode, useContext } from "react";

export interface ConfirmOptions {
	title: ReactNode;
	message?: ReactNode;
	confirmLabel?: ReactNode;
	cancelLabel?: ReactNode;
	tone?: "danger" | "default";
}

export type ConfirmFn = (options: ConfirmOptions) => Promise<boolean>;

export const ConfirmContext = createContext<ConfirmFn | null>(null);

export function useConfirm() {
	const confirm = useContext(ConfirmContext);

	if (!confirm) {
		throw new Error("useConfirm must be used within ConfirmProvider");
	}

	return confirm;
}
