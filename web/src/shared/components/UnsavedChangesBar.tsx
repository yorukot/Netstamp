import { classNames } from "@/shared/utils/classNames";
import { Button } from "@netstamp/ui";
import type { ButtonHTMLAttributes, ReactNode } from "react";
import styles from "./UnsavedChangesBar.module.css";

interface UnsavedChangesBarProps {
	show?: boolean;
	message?: ReactNode;
	resetLabel?: string;
	saveLabel?: string;
	savingLabel?: string;
	saving?: boolean;
	disabled?: boolean;
	form?: string;
	saveType?: ButtonHTMLAttributes<HTMLButtonElement>["type"];
	className?: string;
	onReset: () => void;
	onSave?: () => void;
}

export function UnsavedChangesBar({
	show = true,
	message = "Careful, you have unsaved changes.",
	resetLabel = "Reset",
	saveLabel = "Save Changes",
	savingLabel = "Saving",
	saving = false,
	disabled = false,
	form,
	saveType = "button",
	className,
	onReset,
	onSave
}: UnsavedChangesBarProps) {
	if (!show) {
		return null;
	}

	return (
		<div className={classNames(styles.bar, className)} role="status" aria-live="polite">
			<strong className={styles.message}>{message}</strong>
			<div className={styles.actions}>
				<Button type="button" variant="plain" className={styles.resetButton} disabled={saving} onClick={onReset}>
					{resetLabel}
				</Button>
				<Button type={saveType} form={form} disabled={disabled || saving} onClick={onSave}>
					{saving ? savingLabel : saveLabel}
				</Button>
			</div>
		</div>
	);
}
