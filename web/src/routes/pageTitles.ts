const APP_TITLE = "Netstamp";

interface TitleRouteMatch {
	handle: unknown;
}

export interface PageTitleHandle {
	title: string;
}

export function pageTitleHandle(title: string): PageTitleHandle {
	return { title };
}

export function formatPageTitle(title: string | null | undefined) {
	return title ? `${title} - ${APP_TITLE}` : APP_TITLE;
}

function isPageTitleHandle(handle: unknown): handle is PageTitleHandle {
	return typeof handle === "object" && handle !== null && "title" in handle && typeof handle.title === "string" && handle.title.length > 0;
}

export function pageTitleFromMatches(matches: readonly TitleRouteMatch[]) {
	for (let index = matches.length - 1; index >= 0; index -= 1) {
		const handle = matches[index]?.handle;

		if (isPageTitleHandle(handle)) {
			return formatPageTitle(handle.title);
		}
	}

	return formatPageTitle(null);
}
