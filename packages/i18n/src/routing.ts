import { defaultLocale, normalizeLocale, supportedLocales, type SupportedLocale } from "./locales.js";

interface PathParts {
	pathname: string;
	suffix: string;
}

const splitPath = (value: string): PathParts => {
	const suffixIndex = value.search(/[?#]/);
	const rawPathname = suffixIndex >= 0 ? value.slice(0, suffixIndex) : value;
	const suffix = suffixIndex >= 0 ? value.slice(suffixIndex) : "";
	const pathname = rawPathname.startsWith("/") ? rawPathname : `/${rawPathname}`;

	return { pathname: pathname || "/", suffix };
};

export const localeFromPath = (value: string): SupportedLocale => {
	const { pathname } = splitPath(value);
	const firstSegment = pathname.split("/").filter(Boolean)[0];

	return normalizeLocale(firstSegment) ?? defaultLocale;
};

export const stripLocalePrefix = (value: string) => {
	const { pathname, suffix } = splitPath(value);
	const segments = pathname.split("/").filter(Boolean);
	const firstSegment = segments[0];

	if (firstSegment && supportedLocales.some(locale => locale === firstSegment)) {
		segments.shift();
	}

	const trailingSlash = pathname.endsWith("/") || segments.length === 0;
	const unprefixed = `/${segments.join("/")}${trailingSlash && segments.length > 0 ? "/" : ""}`;

	return `${unprefixed}${suffix}`;
};

export const localePath = (value: string, locale: SupportedLocale) => {
	const unprefixed = stripLocalePrefix(value);
	const { pathname, suffix } = splitPath(unprefixed);

	if (locale === defaultLocale) return `${pathname}${suffix}`;

	return pathname === "/" ? `/${locale}/${suffix}` : `/${locale}${pathname}${suffix}`;
};

export const alternateLocalePaths = (value: string) => Object.fromEntries(supportedLocales.map(locale => [locale, localePath(value, locale)])) as Record<SupportedLocale, string>;
