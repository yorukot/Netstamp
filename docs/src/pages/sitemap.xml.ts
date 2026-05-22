import { docsPages } from "../data/docs";
import { absoluteUrl, resolveSiteUrl } from "../lib/seo";

export const prerender = true;

interface SitemapRoute {
	path: string;
	changefreq: "daily" | "weekly" | "monthly";
	priority: string;
}

const staticRoutes: SitemapRoute[] = [
	{ path: "/", changefreq: "weekly", priority: "1.0" },
	{ path: "/docs/", changefreq: "weekly", priority: "0.8" },
	{ path: "/openapi/", changefreq: "weekly", priority: "0.7" }
];

function escapeXml(value: string) {
	return value.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");
}

function routeForDoc(path: string): SitemapRoute {
	return {
		path,
		changefreq: "monthly",
		priority: path === "/docs/" ? "0.8" : "0.6"
	};
}

export function GET({ site }: { site?: URL }) {
	const siteUrl = resolveSiteUrl(site);
	const routes = new Map<string, SitemapRoute>();

	for (const route of staticRoutes) {
		routes.set(route.path, route);
	}

	for (const page of docsPages) {
		if (!routes.has(page.href)) {
			routes.set(page.href, routeForDoc(page.href));
		}
	}

	const urls = Array.from(routes.values())
		.sort((a, b) => a.path.localeCompare(b.path))
		.map(
			route => `	<url>
		<loc>${escapeXml(absoluteUrl(route.path, siteUrl))}</loc>
		<changefreq>${route.changefreq}</changefreq>
		<priority>${route.priority}</priority>
	</url>`
		)
		.join("\n");

	return new Response(`<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n${urls}\n</urlset>\n`, {
		headers: {
			"Content-Type": "application/xml; charset=utf-8"
		}
	});
}
