export const supportedLocales = ["en", "zh-TW"] as const;

export type SupportedLocale = (typeof supportedLocales)[number];

export const defaultLocale: SupportedLocale = "en";
export const fallbackLocale: SupportedLocale = "en";
export const localeStorageKey = "netstamp:locale";

export const localeMetadata = {
	en: {
		label: "English",
		htmlLang: "en",
		crowdinLocale: "en"
	},
	"zh-TW": {
		label: "繁體中文",
		htmlLang: "zh-Hant-TW",
		crowdinLocale: "zh-TW"
	}
} as const satisfies Record<SupportedLocale, { label: string; htmlLang: string; crowdinLocale: string }>;

const supportedLocaleSet = new Set<string>(supportedLocales);

export const isSupportedLocale = (value: unknown): value is SupportedLocale => typeof value === "string" && supportedLocaleSet.has(value);

export const normalizeLocale = (value: string | null | undefined): SupportedLocale | undefined => {
	if (!value) return undefined;

	if (isSupportedLocale(value)) return value;

	const normalized = value.trim().replaceAll("_", "-").toLowerCase();
	if (normalized === "en" || normalized.startsWith("en-")) return "en";
	if (normalized === "zh-tw" || normalized === "zh-hk" || normalized === "zh-mo" || normalized === "zh-hant" || normalized.startsWith("zh-hant-")) return "zh-TW";

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
