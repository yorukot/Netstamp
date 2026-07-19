import { defaultLocale, isSupportedLocale, localePath, type SupportedLocale } from "@netstamp/i18n";
import en from "./locales/en/ui.json";
import zhTW from "./locales/zh-TW/ui.json";

export const localeFromAstro = (locale: string | undefined): SupportedLocale => (isSupportedLocale(locale) ? locale : defaultLocale);
export const uiForLocale = (locale: string | undefined): typeof en => (localeFromAstro(locale) === "zh-TW" ? zhTW : en);
export const pathForLocale = (path: string, locale: string | undefined) => localePath(path, localeFromAstro(locale));

export const interpolate = (template: string, values: Record<string, string>) => Object.entries(values).reduce((value, [key, replacement]) => value.replaceAll(`{{${key}}}`, replacement), template);
