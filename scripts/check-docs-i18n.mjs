import { access, readFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";

const dist = path.join(process.cwd(), "docs/dist");

const readPage = relativePath => readFile(path.join(dist, relativePath, "index.html"), "utf8");
const assertIncludes = (html, expected, page) => {
	if (!html.includes(expected)) throw new Error(`${page} does not include ${JSON.stringify(expected)}`);
};
const assertExcludes = (html, unexpected, page) => {
	if (html.includes(unexpected)) throw new Error(`${page} unexpectedly includes ${JSON.stringify(unexpected)}`);
};
const assertEqual = (actual, expected, label) => {
	if (JSON.stringify(actual) !== JSON.stringify(expected)) {
		throw new Error(`${label} mismatch.\nExpected: ${JSON.stringify(expected)}\nReceived: ${JSON.stringify(actual)}`);
	}
};
const assertMissing = async (target, label) => {
	try {
		await access(target);
	} catch {
		return;
	}
	throw new Error(`${label} unexpectedly exists at ${target}`);
};
const docsNavHrefs = html => Array.from(html.matchAll(/<a href="([^"]+)" class="docsNavItem(?: active)?"/g), match => match[1]);
const activeDocsNavHrefs = html => Array.from(html.matchAll(/<a href="([^"]+)" class="docsNavItem active" aria-current="page"/g), match => match[1]);
const docsSectionLabels = html =>
	Array.from(html.matchAll(/<h2 class="docsNavSectionTitle"[^>]*>([^<]+)<\/h2>/g), match => match[1].replaceAll("&amp;", "&")).filter(label => label !== "External links");

const expectedSectionKeys = ["start", "install", "use", "operate"];
const expectedEnglishSections = ["Getting Started", "Installation", "Using Netstamp", "Operations"];
const expectedDocRoutes = [
	"docs/getting-started/quick-start",
	"docs/getting-started/core-concepts",
	"docs/installation",
	"docs/installation/docker-compose",
	"docs/installation/configuration",
	"docs/installation/reverse-proxy-and-https",
	"docs/installation/authentication-and-email",
	"docs/installation/backup-and-restore",
	"docs/guides",
	"docs/guides/projects-and-members",
	"docs/guides/probes",
	"docs/guides/labels-and-assignments",
	"docs/guides/checks",
	"docs/guides/results-and-insight",
	"docs/guides/results-and-insight/analyze-measurements",
	"docs/guides/results-and-insight/investigate-traceroute",
	"docs/guides/alerts-and-incidents",
	"docs/guides/notifications",
	"docs/guides/status-pages",
	"docs/guides/account-settings",
	"docs/guides/system-administration",
	"docs/operations",
	"docs/operations/probe-agent",
	"docs/operations/security"
];
const expectedNavHrefs = expectedDocRoutes.map(route => `/${route}/`);
expectedNavHrefs.push("/changelog/");
const englishUi = JSON.parse(await readFile(path.join(process.cwd(), "docs/src/i18n/locales/en/ui.json"), "utf8"));

assertEqual(Object.keys(englishUi.docs.sections), expectedSectionKeys, "English documentation section keys");
assertEqual(Object.values(englishUi.docs.sections), expectedEnglishSections, "English documentation section labels");

const [
	englishHome,
	englishNotFound,
	englishOpenApi,
	englishChangelog,
	englishDocsRedirect,
	englishQuickStart,
	englishInstallation,
	englishProbes,
	englishAnalyzeMeasurements,
	englishInvestigateTraceroute,
	sitemap
] = await Promise.all([
	readPage(""),
	readFile(path.join(dist, "404.html"), "utf8"),
	readPage("openapi"),
	readPage("changelog"),
	readPage("docs"),
	readPage("docs/getting-started/quick-start"),
	readPage("docs/installation"),
	readPage("docs/guides/probes"),
	readPage("docs/guides/results-and-insight/analyze-measurements"),
	readPage("docs/guides/results-and-insight/investigate-traceroute"),
	readFile(path.join(dist, "sitemap.xml"), "utf8")
]);

await Promise.all(expectedDocRoutes.map(route => readPage(route)));
await assertMissing(path.join(dist, "zh-TW"), "Traditional Chinese output directory");

assertEqual(docsNavHrefs(englishQuickStart), expectedNavHrefs, "English documentation navigation order");
assertExcludes(englishQuickStart, "--nav-depth", "English documentation navigation hierarchy");
assertEqual(activeDocsNavHrefs(englishQuickStart), ["/docs/getting-started/quick-start/"], "English Quick Start active navigation item");
assertEqual(activeDocsNavHrefs(englishInstallation), ["/docs/installation/"], "English installation overview active navigation item");
assertEqual(activeDocsNavHrefs(englishProbes), ["/docs/guides/probes/"], "English probes active navigation item");
assertEqual(activeDocsNavHrefs(englishAnalyzeMeasurements), ["/docs/guides/results-and-insight/analyze-measurements/"], "English measurement analysis active navigation item");
assertEqual(activeDocsNavHrefs(englishInvestigateTraceroute), ["/docs/guides/results-and-insight/investigate-traceroute/"], "English Traceroute investigation active navigation item");
assertEqual(activeDocsNavHrefs(englishChangelog), ["/changelog/"], "English Changelog active navigation item");
assertIncludes(englishQuickStart, 'data-astro-transition-persist="docs-sidebar"', "English persistent documentation sidebar");
assertIncludes(englishQuickStart, 'data-docs-locale="en"', "English documentation sidebar locale");
assertEqual(docsSectionLabels(englishQuickStart), [...expectedEnglishSections, "Project"], "Rendered English documentation section order");

assertIncludes(englishHome, '<html lang="en"', "/");
assertIncludes(englishNotFound, "Page not found - Netstamp", "/404.html");
assertIncludes(englishNotFound, 'content="noindex,nofollow,noarchive"', "/404.html");
assertIncludes(englishDocsRedirect, 'content="0;url=/docs/getting-started/quick-start/"', "/docs/ redirect");
assertIncludes(englishDocsRedirect, 'href="/docs/getting-started/quick-start/"', "/docs/ redirect");
assertIncludes(englishQuickStart, "By the end, you will have", "/docs/getting-started/quick-start/");
assertExcludes(englishQuickStart, "&quot;href&quot;:&quot;/docs/&quot;", "English documentation search index");
assertIncludes(englishQuickStart, 'hreflang="en"', "/docs/getting-started/quick-start/");
assertIncludes(englishQuickStart, 'hreflang="x-default"', "/docs/getting-started/quick-start/");
assertExcludes(englishHome, '<div class="siteLanguageMenu"', "/");
assertExcludes(englishQuickStart, '<div class="siteLanguageMenu"', "/docs/getting-started/quick-start/");
assertExcludes(englishQuickStart, 'hreflang="zh-TW"', "/docs/getting-started/quick-start/");
assertExcludes(englishQuickStart, "og:locale:alternate", "/docs/getting-started/quick-start/");
assertIncludes(englishOpenApi, "Loading API reference", "/openapi/");
assertIncludes(englishChangelog, "v0.0.0", "/changelog/");
assertIncludes(englishChangelog, "2026-08-19", "/changelog/");
assertIncludes(sitemap, "<loc>https://netstamp.dev/docs/getting-started/quick-start/</loc>", "English Quick Start sitemap entry");
assertIncludes(sitemap, "<loc>https://netstamp.dev/changelog/</loc>", "English Changelog sitemap entry");
assertExcludes(sitemap, "/zh-TW/", "English-only sitemap");

console.log(`English-only docs architecture check passed for ${expectedDocRoutes.length} documentation routes and ${expectedSectionKeys.length} navigation sections.`);
