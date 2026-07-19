import { classNames } from "@/shared/utils/classNames";
import { Button } from "@netstamp/ui";
import { useState, type AnimationEvent, type ButtonHTMLAttributes, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
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

type VisibilityState = "closed" | "open" | "closing";

export function UnsavedChangesBar({
	show = true,
	message,
	resetLabel,
	saveLabel,
	savingLabel,
	saving = false,
	disabled = false,
	form,
	saveType = "button",
	className,
	onReset,
	onSave
}: UnsavedChangesBarProps) {
	const { t } = useTranslation("common");
	const [visibility, setVisibility] = useState<VisibilityState>(() => (show ? "open" : "closed"));
	const visibleMessage = message ?? t("unsaved.message");
	const visibleResetLabel = resetLabel ?? t("actions.reset");
	const visibleSaveLabel = saveLabel ?? t("actions.saveChanges");
	const visibleSavingLabel = savingLabel ?? t("actions.saving");
	let visibleState = visibility;

	if (show && visibility !== "open") {
		visibleState = "open";
		setVisibility("open");
	} else if (!show && visibility === "open") {
		visibleState = "closing";
		setVisibility("closing");
	}

	function finishExit(event: AnimationEvent<HTMLDivElement>) {
		if (event.currentTarget !== event.target || visibility !== "closing") {
			return;
		}

		setVisibility("closed");
	}

	if (visibleState === "closed") {
		return null;
	}

	const closing = visibleState === "closing";

	return (
		<div className={classNames(styles.bar, className)} data-state={visibleState} role="status" aria-live="polite" aria-hidden={closing || undefined} onAnimationEnd={finishExit}>
			<strong className={styles.message}>{visibleMessage}</strong>
			<div className={styles.actions}>
				<Button type="button" variant="plain" className={styles.resetButton} disabled={saving || closing} onClick={onReset}>
					{visibleResetLabel}
				</Button>
				<Button type={saveType} form={form} disabled={disabled || saving || closing} onClick={onSave}>
					{saving ? visibleSavingLabel : visibleSaveLabel}
				</Button>
			</div>
		</div>
	);
}
