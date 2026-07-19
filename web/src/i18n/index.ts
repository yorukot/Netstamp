import { defaultLocale, htmlLangForLocale, isSupportedLocale, localeStorageKey, resolveLocale, type SupportedLocale } from "@netstamp/i18n";
import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import { defaultNamespace, namespaces, resources } from "./resources";

let initializationPromise: Promise<void> | undefined;

const storedLocale = () => {
	try {
		return window.localStorage.getItem(localeStorageKey);
	} catch {
		return null;
	}
};

export const detectInitialLocale = (): SupportedLocale => {
	if (typeof window === "undefined") return defaultLocale;

	return resolveLocale([storedLocale(), ...(window.navigator.languages ?? []), window.navigator.language]);
};

const synchronizeDocumentLocale = (locale: SupportedLocale) => {
	document.documentElement.lang = htmlLangForLocale(locale);

	try {
		window.localStorage.setItem(localeStorageKey, locale);
	} catch {
		// The UI can still switch languages when storage is unavailable.
	}
};

export const initializeI18n = () => {
	if (initializationPromise) return initializationPromise;

	initializationPromise = i18n
		.use(initReactI18next)
		.init({
			resources,
			lng: detectInitialLocale(),
			fallbackLng: defaultLocale,
			supportedLngs: ["en", "zh-TW"],
			ns: namespaces,
			defaultNS: defaultNamespace,
			returnNull: false,
			returnEmptyString: false,
			interpolation: {
				escapeValue: false
			},
			initAsync: false
		})
		.then(() => {
			const locale = isSupportedLocale(i18n.resolvedLanguage) ? i18n.resolvedLanguage : defaultLocale;
			synchronizeDocumentLocale(locale);
			i18n.on("languageChanged", language => {
				synchronizeDocumentLocale(isSupportedLocale(language) ? language : defaultLocale);
			});
		});

	return initializationPromise;
};

export const currentLocale = (): SupportedLocale => (isSupportedLocale(i18n.resolvedLanguage) ? i18n.resolvedLanguage : defaultLocale);

export const changeLocale = async (locale: SupportedLocale) => {
	await initializeI18n();
	await i18n.changeLanguage(locale);
};

export { i18n };
