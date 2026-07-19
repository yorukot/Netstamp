import { changeLocale, currentLocale } from "@/i18n";
import { localeMetadata, supportedLocales } from "@netstamp/i18n";
import { TranslateIcon } from "@phosphor-icons/react/dist/csr/Translate";
import { useTranslation } from "react-i18next";

interface LanguageSwitcherProps {
	className?: string;
	onChange?: () => void;
}

export const LanguageSwitcher = ({ className, onChange }: LanguageSwitcherProps) => {
	const { t } = useTranslation("common");
	const locale = currentLocale();
	const targetLocale = supportedLocales.find(candidate => candidate !== locale) ?? "en";
	const targetLabel = localeMetadata[targetLocale].label;

	const switchLanguage = async () => {
		await changeLocale(targetLocale);
		onChange?.();
	};

	return (
		<button
			type="button"
			className={className}
			aria-label={t("language.switchTo", { language: targetLabel })}
			title={t("language.switchTo", { language: targetLabel })}
			onClick={() => void switchLanguage()}
		>
			<TranslateIcon size={18} weight="bold" aria-hidden="true" focusable="false" />
			<span>{targetLabel}</span>
		</button>
	);
};
