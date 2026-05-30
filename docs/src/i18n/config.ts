// Locale configuration for the docs site.
//
// The setup is table-driven so adding a language is a config change, not a
// code change: append the locale code to `locales`, then fill in the matching
// entries below. The default locale lives at the repository root of
// `src/content/docs/`; every other locale mirrors the same relative paths
// inside a `src/content/docs/<locale>/` folder.

export const locales = ["en", "zh-TW"] as const;

export type Locale = (typeof locales)[number];

export const defaultLocale: Locale = "en";

// Locales that render under a `/<locale>/` URL prefix (everything except the
// default, which stays at the site root for backwards-compatible URLs).
export const prefixedLocales: Locale[] = locales.filter(locale => locale !== defaultLocale);

// Human-readable name shown in the language switcher, written in its own script.
export const localeNames: Record<Locale, string> = {
	en: "English",
	"zh-TW": "繁體中文"
};

// Value for the `<html lang>` attribute (BCP 47).
export const localeHtmlLang: Record<Locale, string> = {
	en: "en",
	"zh-TW": "zh-Hant-TW"
};

// Value for the `og:locale` Open Graph tag and `hreflang` alternates.
export const localeOpenGraph: Record<Locale, string> = {
	en: "en_US",
	"zh-TW": "zh_TW"
};

// Text direction for the document. All current locales are left-to-right; the
// field exists so an RTL locale only needs a config entry.
export const localeDir: Record<Locale, "ltr" | "rtl"> = {
	en: "ltr",
	"zh-TW": "ltr"
};

// Short label used as the language-switcher trigger (kept compact for the nav).
export const localeShortLabel: Record<Locale, string> = {
	en: "EN",
	"zh-TW": "繁中"
};

export interface UiStrings {
	docs: string;
	signIn: string;
	previous: string;
	next: string;
	noPreviousPage: string;
	noNextPage: string;
	onThisPage: string;
	noHeadings: string;
	editOnGithub: string;
	foldSidebar: string;
	expandSidebar: string;
	foldLabel: string;
	expandLabel: string;
	searchTitle: string;
	searchPlaceholder: string;
	searchClose: string;
	searchEmpty: string;
	language: string;
	fallbackNotice: string;
}

// All chrome strings rendered by the docs shell, per locale. Content (the MDX
// body) is translated separately as files; this table covers the UI around it.
export const uiStrings: Record<Locale, UiStrings> = {
	en: {
		docs: "Docs",
		signIn: "Sign in",
		previous: "Previous",
		next: "Next",
		noPreviousPage: "No previous page",
		noNextPage: "No next page",
		onThisPage: "On This Page",
		noHeadings: "No headings",
		editOnGithub: "Edit on GitHub",
		foldSidebar: "Fold documentation sidebar",
		expandSidebar: "Expand documentation sidebar",
		foldLabel: "Fold",
		expandLabel: "Expand",
		searchTitle: "Search docs",
		searchPlaceholder: "Search documentation",
		searchClose: "Close",
		searchEmpty: "No matches",
		language: "Language",
		fallbackNotice: "This page has not been translated yet and is shown in English."
	},
	"zh-TW": {
		docs: "文件",
		signIn: "登入",
		previous: "上一頁",
		next: "下一頁",
		noPreviousPage: "沒有上一頁",
		noNextPage: "沒有下一頁",
		onThisPage: "本頁目錄",
		noHeadings: "沒有標題",
		editOnGithub: "在 GitHub 上編輯",
		foldSidebar: "收合文件側欄",
		expandSidebar: "展開文件側欄",
		foldLabel: "收合",
		expandLabel: "展開",
		searchTitle: "搜尋文件",
		searchPlaceholder: "搜尋文件內容",
		searchClose: "關閉",
		searchEmpty: "沒有符合的結果",
		language: "語言",
		fallbackNotice: "這個頁面尚未翻譯，先以英文呈現。"
	}
};

// Localized display labels for the navigation section headings. Sections are
// grouped and ordered by their canonical English key (see `data/docs.ts`); this
// table only controls how each key is shown. Missing entries fall back to the key.
export const sectionLabels: Record<Locale, Record<string, string>> = {
	en: {},
	"zh-TW": {
		Start: "開始",
		Guides: "指南",
		Reference: "參考"
	}
};

const localeSet = new Set<string>(locales);

// Astro's content loader lowercases entry ids, so a `zh-TW/` folder surfaces as
// the `zh-tw` segment. Match locale path segments case-insensitively and map
// them back to the canonical locale code.
const localeBySegment = new Map<string, Locale>(locales.map(locale => [locale.toLowerCase(), locale]));

export function isLocale(value: string | undefined): value is Locale {
	return value !== undefined && localeSet.has(value);
}

export function localeFromSegment(segment: string | undefined): Locale | undefined {
	return segment === undefined ? undefined : localeBySegment.get(segment.toLowerCase());
}

// Base path for the docs section of a given locale. The default locale keeps
// the historical `/docs` path; other locales are prefixed.
export function docsBasePath(locale: Locale): string {
	return locale === defaultLocale ? "/docs" : `/${locale}/docs`;
}

export function getUiStrings(locale: Locale): UiStrings {
	return uiStrings[locale] ?? uiStrings[defaultLocale];
}

export function getSectionLabel(locale: Locale, key: string): string {
	return sectionLabels[locale]?.[key] ?? sectionLabels[defaultLocale]?.[key] ?? key;
}
