import type { AppResources } from "@/i18n/resources";

const APP_TITLE = "Netstamp";
type PageTitleKey = keyof AppResources["navigation"]["pageTitles"];
type PageTitleTranslationKey = `pageTitles.${PageTitleKey}`;

interface TitleRouteMatch {
	handle: unknown;
}

export interface PageTitleHandle {
	titleKey: PageTitleTranslationKey;
}

export function pageTitleHandle(titleKey: PageTitleTranslationKey): PageTitleHandle {
	return { titleKey };
}

export function formatPageTitle(title: string | null | undefined) {
	return title ? `${title} - ${APP_TITLE}` : APP_TITLE;
}

function isPageTitleHandle(handle: unknown): handle is PageTitleHandle {
	return typeof handle === "object" && handle !== null && "titleKey" in handle && typeof handle.titleKey === "string" && handle.titleKey.length > 0;
}

export function pageTitleFromMatches(matches: readonly TitleRouteMatch[], translate: (key: PageTitleTranslationKey) => string) {
	for (let index = matches.length - 1; index >= 0; index -= 1) {
		const handle = matches[index]?.handle;

		if (isPageTitleHandle(handle)) {
			return formatPageTitle(translate(handle.titleKey));
		}
	}

	return formatPageTitle(null);
}
