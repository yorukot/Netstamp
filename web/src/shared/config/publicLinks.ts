function trimTrailingSlash(value: string) {
	return value.replace(/\/+$/, "");
}

export const docsBaseUrl = trimTrailingSlash(import.meta.env.VITE_NETSTAMP_DOCS_BASE_URL || (import.meta.env.DEV ? "http://localhost:4321" : "https://netstamp.dev"));

export function docsUrl(path = "/docs/") {
	const normalizedPath = path.startsWith("/") ? path : `/${path}`;
	return `${docsBaseUrl}${normalizedPath}`;
}
