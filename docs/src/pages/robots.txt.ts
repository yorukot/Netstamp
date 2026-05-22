import { absoluteUrl, resolveSiteUrl } from "../lib/seo";

export const prerender = true;

const disallowedPaths = ["/api/", "/login", "/register", "/onboarding", "/dashboard", "/probes", "/insight", "/checks", "/alerts", "/project", "/settings", "/storybook/"];

export function GET({ site }: { site?: URL }) {
	const siteUrl = resolveSiteUrl(site);
	const rules = ["User-agent: *", "Allow: /", ...disallowedPaths.map(path => `Disallow: ${path}`), "", `Sitemap: ${absoluteUrl("/sitemap.xml", siteUrl)}`, ""];

	return new Response(rules.join("\n"), {
		headers: {
			"Content-Type": "text/plain; charset=utf-8"
		}
	});
}
