export const supportedLocales = ["en"] as const;

export type SupportedLocale = (typeof supportedLocales)[number];

export const defaultLocale: SupportedLocale = "en";
export const fallbackLocale: SupportedLocale = "en";
export const localeStorageKey = "netstamp:locale";
export const hasMultipleLocales = supportedLocales.length > 1;

export const localeMetadata = {
	en: {
		label: "English",
		htmlLang: "en",
		crowdinLocale: "en",
		openGraphLocale: "en_US"
	}
} as const satisfies Record<SupportedLocale, { label: string; htmlLang: string; crowdinLocale: string; openGraphLocale: string }>;

const supportedLocaleSet = new Set<string>(supportedLocales);

export const isSupportedLocale = (value: unknown): value is SupportedLocale => typeof value === "string" && supportedLocaleSet.has(value);

export const normalizeLocale = (value: string | null | undefined): SupportedLocale | undefined => {
	if (!value) return undefined;

	if (isSupportedLocale(value)) return value;

	const normalized = value.trim().replaceAll("_", "-").toLowerCase();
	if (normalized === "en" || normalized.startsWith("en-")) return "en";

	return undefined;
};

export const resolveLocale = (candidates: Iterable<string | null | undefined>): SupportedLocale => {
	for (const candidate of candidates) {
		const locale = normalizeLocale(candidate);
		if (locale) return locale;
	}

	return fallbackLocale;
};

export const htmlLangForLocale = (locale: SupportedLocale) => localeMetadata[locale].htmlLang;
