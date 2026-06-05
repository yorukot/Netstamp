const GRAVATAR_BASE_URL = "https://www.gravatar.com/avatar";

function createDefaultGravatarUrl(size: number) {
	const params = new URLSearchParams({
		s: String(size),
		d: "identicon"
	});

	return `${GRAVATAR_BASE_URL}/?${params.toString()}`;
}

function hexFromBuffer(buffer: ArrayBuffer) {
	return Array.from(new Uint8Array(buffer), byte => byte.toString(16).padStart(2, "0")).join("");
}

export async function createGravatarUrl(email: string, size = 160) {
	const normalizedEmail = email.trim().toLowerCase();

	if (!normalizedEmail || !globalThis.crypto?.subtle) {
		return createDefaultGravatarUrl(size);
	}

	const digest = await globalThis.crypto.subtle.digest("SHA-256", new TextEncoder().encode(normalizedEmail));
	const hash = hexFromBuffer(digest);
	const params = new URLSearchParams({
		s: String(size),
		d: "identicon"
	});

	return `${GRAVATAR_BASE_URL}/${hash}?${params.toString()}`;
}
