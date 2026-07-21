import { changeLocale, currentLocale } from "@/i18n";
import { classNames } from "@/shared/utils/classNames";
import { localeMetadata, supportedLocales } from "@netstamp/i18n";
import { DropdownMenuContent, DropdownMenuItem, DropdownMenuPortal, DropdownMenuRoot, DropdownMenuTrigger } from "@netstamp/ui";
import { CheckIcon } from "@phosphor-icons/react/dist/csr/Check";
import { TranslateIcon } from "@phosphor-icons/react/dist/csr/Translate";
import { useTranslation } from "react-i18next";
import styles from "./LanguageSwitcher.module.css";

interface LanguageSwitcherProps {
	className?: string;
	menuClassName?: string;
	onChange?: () => void;
}

export const LanguageSwitcher = ({ className, menuClassName, onChange }: LanguageSwitcherProps) => {
	const { t } = useTranslation("common");
	const locale = currentLocale();

	const selectLanguage = async (selectedLocale: (typeof supportedLocales)[number]) => {
		if (selectedLocale !== locale) await changeLocale(selectedLocale);
		onChange?.();
	};

	return (
		<DropdownMenuRoot>
			<DropdownMenuTrigger asChild>
				<button type="button" className={classNames(styles.trigger, className)} aria-label={t("language.label")} title={t("language.label")}>
					<TranslateIcon size="1.25rem" weight="bold" aria-hidden="true" focusable="false" />
				</button>
			</DropdownMenuTrigger>
			<DropdownMenuPortal>
				<DropdownMenuContent className={classNames(styles.menu, menuClassName)} align="end" sideOffset={8} collisionPadding={8} aria-label={t("language.label")}>
					{supportedLocales.map(candidate => {
						const selected = candidate === locale;

						return (
							<DropdownMenuItem key={candidate} className={styles.item} aria-current={selected ? "true" : undefined} onSelect={() => void selectLanguage(candidate)}>
								<span>{localeMetadata[candidate].label}</span>
								{selected ? <CheckIcon size="1rem" weight="bold" aria-hidden="true" focusable="false" /> : null}
							</DropdownMenuItem>
						);
					})}
				</DropdownMenuContent>
			</DropdownMenuPortal>
		</DropdownMenuRoot>
	);
};
