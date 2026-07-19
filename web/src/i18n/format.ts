import { defaultLocale, htmlLangForLocale, isSupportedLocale } from "@netstamp/i18n";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { currentLocale } from "./index";

export const localeForFormatting = () => htmlLangForLocale(currentLocale());

export const formatDateTime = (value: string | number | Date, options?: Intl.DateTimeFormatOptions) => new Intl.DateTimeFormat(localeForFormatting(), options).format(new Date(value));

export const formatNumber = (value: number, options?: Intl.NumberFormatOptions) => new Intl.NumberFormat(localeForFormatting(), options).format(value);

export const useLocaleFormat = () => {
	const { i18n } = useTranslation();
	const language = i18n.resolvedLanguage;

	return useMemo(() => {
		const locale = htmlLangForLocale(isSupportedLocale(language) ? language : defaultLocale);

		return {
			dateTime: (value: string | number | Date, options?: Intl.DateTimeFormatOptions) => new Intl.DateTimeFormat(locale, options).format(new Date(value)),
			number: (value: number, options?: Intl.NumberFormatOptions) => new Intl.NumberFormat(locale, options).format(value),
			relativeTime: (value: number, unit: Intl.RelativeTimeFormatUnit, options?: Intl.RelativeTimeFormatOptions) => new Intl.RelativeTimeFormat(locale, options).format(value, unit)
		};
	}, [language]);
};
