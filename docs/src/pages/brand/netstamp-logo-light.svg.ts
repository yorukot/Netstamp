import logo from "@netstamp/brand/assets/netstamp-logo-light.svg?raw";

export const prerender = true;

export function GET() {
	return new Response(logo, {
		headers: {
			"Cache-Control": "public, max-age=31536000, immutable",
			"Content-Type": "image/svg+xml; charset=utf-8"
		}
	});
}
