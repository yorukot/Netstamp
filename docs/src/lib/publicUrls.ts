function trimTrailingSlash(value: string) {
	return value.replace(/\/+$/, "");
}

export const appBaseUrl = trimTrailingSlash(import.meta.env.PUBLIC_NETSTAMP_APP_BASE_URL || "https://app.netstamp.dev");

export function appUrl(path = "/") {
	const normalizedPath = path.startsWith("/") ? path : `/${path}`;
	return `${appBaseUrl}${normalizedPath}`;
}
