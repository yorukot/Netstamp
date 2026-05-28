export const PUBLIC_PAGE_SLUG_MAX_LENGTH = 64;
export const PUBLIC_PAGE_SLUG_PATTERN = /^[a-z0-9-]+$/;
export const PUBLIC_PAGE_SLUG_HELPER = "Use lowercase letters, numbers, and hyphens.";

export function sanitizePublicPageSlug(value: string) {
	return value
		.trim()
		.replace(/[^a-z0-9-]/g, "")
		.slice(0, PUBLIC_PAGE_SLUG_MAX_LENGTH);
}

export function isValidPublicPageSlug(value: string) {
	return value.length > 0 && value.length <= PUBLIC_PAGE_SLUG_MAX_LENGTH && PUBLIC_PAGE_SLUG_PATTERN.test(value);
}
