import { XIcon } from "@phosphor-icons/react/dist/csr/X";
import type { ComponentPropsWithoutRef } from "react";
import { useTranslation } from "react-i18next";
import styles from "./CloseButton.module.css";

interface CloseButtonProps extends Omit<ComponentPropsWithoutRef<"button">, "children"> {
	ariaLabel?: string;
}

export function CloseButton({ ariaLabel, className, title, type = "button", ...props }: CloseButtonProps) {
	const { t } = useTranslation("common");
	const classes = [styles.button, className].filter(Boolean).join(" ");
	const visibleAriaLabel = ariaLabel ?? t("a11y.closePanel");

	return (
		<button type={type} className={classes} aria-label={visibleAriaLabel} title={title ?? visibleAriaLabel} {...props}>
			<XIcon size={16} weight="bold" aria-hidden="true" focusable="false" />
		</button>
	);
}
