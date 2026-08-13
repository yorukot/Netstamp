import { readFile } from "node:fs/promises";
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
const docsNavHrefs = html => Array.from(html.matchAll(/<a href="([^"]+)" class="docsNavItem(?: active)?"/g), match => match[1]);
const activeDocsNavHrefs = html => Array.from(html.matchAll(/<a href="([^"]+)" class="docsNavItem active" aria-current="page"/g), match => match[1]);
const docsSectionLabels = html =>
	Array.from(html.matchAll(/<h2 class="docsNavSectionTitle"[^>]*>([^<]+)<\/h2>/g), match => match[1].replaceAll("&amp;", "&")).filter(label => !["External links", "外部連結"].includes(label));
const expectedSectionKeys = ["start", "install", "use", "operate"];
const expectedEnglishSections = ["Getting Started", "Installation", "Using Netstamp", "Operations"];
const expectedTraditionalChineseSections = ["開始使用", "安裝", "使用 Netstamp", "維運"];
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
const englishUi = JSON.parse(await readFile(path.join(process.cwd(), "docs/src/i18n/locales/en/ui.json"), "utf8"));
const traditionalChineseUi = JSON.parse(await readFile(path.join(process.cwd(), "docs/src/i18n/locales/zh-TW/ui.json"), "utf8"));

assertEqual(Object.keys(englishUi.docs.sections), expectedSectionKeys, "English documentation section keys");
assertEqual(Object.values(englishUi.docs.sections), expectedEnglishSections, "English documentation section labels");
assertEqual(Object.keys(traditionalChineseUi.docs.sections), expectedSectionKeys, "Traditional Chinese documentation section keys");
assertEqual(Object.values(traditionalChineseUi.docs.sections), expectedTraditionalChineseSections, "Traditional Chinese documentation section labels");

const [
	englishHome,
	traditionalChineseHome,
	englishNotFound,
	traditionalChineseNotFound,
	traditionalChineseOpenApi,
	englishDocsRedirect,
	traditionalChineseDocsRedirect,
	englishQuickStart,
	traditionalChineseQuickStart,
	englishInstallation,
	englishProbes,
	sitemap
] = await Promise.all([
	readPage(""),
	readPage("zh-TW"),
	readFile(path.join(dist, "404.html"), "utf8"),
	readPage("zh-TW/404.html"),
	readPage("zh-TW/openapi"),
	readPage("docs"),
	readPage("zh-TW/docs"),
	readPage("docs/getting-started/quick-start"),
	readPage("zh-TW/docs/getting-started/quick-start"),
	readPage("docs/installation"),
	readPage("docs/guides/probes"),
	readFile(path.join(dist, "sitemap.xml"), "utf8")
]);

await Promise.all(expectedDocRoutes.flatMap(route => [readPage(route), readPage(`zh-TW/${route}`)]));

const englishNavHrefs = docsNavHrefs(englishQuickStart);
const traditionalChineseNavHrefs = docsNavHrefs(traditionalChineseQuickStart).map(href => href.replace(/^\/zh-TW/, ""));
assertEqual(englishNavHrefs, expectedNavHrefs, "English documentation navigation order");
assertEqual(traditionalChineseNavHrefs, expectedNavHrefs, "Traditional Chinese documentation navigation order");
assertExcludes(englishQuickStart, "--nav-depth", "English documentation navigation hierarchy");
assertExcludes(traditionalChineseQuickStart, "--nav-depth", "Traditional Chinese documentation navigation hierarchy");
assertEqual(activeDocsNavHrefs(englishQuickStart), ["/docs/getting-started/quick-start/"], "English Quick Start active navigation item");
assertEqual(activeDocsNavHrefs(traditionalChineseQuickStart), ["/zh-TW/docs/getting-started/quick-start/"], "Traditional Chinese Quick Start active navigation item");
assertEqual(activeDocsNavHrefs(englishInstallation), ["/docs/installation/"], "English installation overview active navigation item");
assertEqual(activeDocsNavHrefs(englishProbes), ["/docs/guides/probes/"], "English probes active navigation item");
assertIncludes(englishQuickStart, 'data-astro-transition-persist="docs-sidebar"', "English persistent documentation sidebar");
assertIncludes(englishQuickStart, 'data-docs-locale="en"', "English documentation sidebar locale");
assertIncludes(traditionalChineseQuickStart, 'data-astro-transition-persist="docs-sidebar"', "Traditional Chinese persistent documentation sidebar");
assertIncludes(traditionalChineseQuickStart, 'data-docs-locale="zh-TW"', "Traditional Chinese documentation sidebar locale");

assertEqual(docsSectionLabels(englishQuickStart), expectedEnglishSections, "Rendered English documentation section order");
assertEqual(docsSectionLabels(traditionalChineseQuickStart), expectedTraditionalChineseSections, "Rendered Traditional Chinese documentation section order");

assertIncludes(englishHome, '<html lang="en"', "/");
assertIncludes(traditionalChineseHome, '<html lang="zh-Hant-TW"', "/zh-TW/");
assertIncludes(englishNotFound, "Page not found - Netstamp", "/404.html");
assertIncludes(englishNotFound, 'content="noindex,nofollow,noarchive"', "/404.html");
assertIncludes(englishNotFound, 'href="/zh-TW/404.html"', "/404.html");
assertIncludes(traditionalChineseNotFound, '<html lang="zh-Hant-TW"', "/zh-TW/404.html/");
assertIncludes(traditionalChineseNotFound, "找不到頁面", "/zh-TW/404.html/");
assertIncludes(traditionalChineseNotFound, 'href="/404.html"', "/zh-TW/404.html/");
assertIncludes(englishDocsRedirect, 'content="0;url=/docs/getting-started/quick-start/"', "/docs/ redirect");
assertIncludes(englishDocsRedirect, 'href="/docs/getting-started/quick-start/"', "/docs/ redirect");
assertIncludes(traditionalChineseDocsRedirect, 'content="0;url=/zh-TW/docs/getting-started/quick-start/"', "/zh-TW/docs/ redirect");
assertIncludes(traditionalChineseDocsRedirect, 'href="/zh-TW/docs/getting-started/quick-start/"', "/zh-TW/docs/ redirect");
assertIncludes(englishQuickStart, "By the end, you will have", "/docs/getting-started/quick-start/");
assertIncludes(traditionalChineseQuickStart, "完成後，你將會有", "/zh-TW/docs/getting-started/quick-start/");
assertExcludes(traditionalChineseQuickStart, traditionalChineseUi.docs.fallbackNotice, "/zh-TW/docs/getting-started/quick-start/");
assertExcludes(englishQuickStart, "&quot;href&quot;:&quot;/docs/&quot;", "English documentation search index");
assertExcludes(traditionalChineseQuickStart, "&quot;href&quot;:&quot;/zh-TW/docs/&quot;", "Traditional Chinese documentation search index");
assertIncludes(traditionalChineseQuickStart, 'href="/docs/getting-started/quick-start/"', "/zh-TW/docs/getting-started/quick-start/");
assertIncludes(traditionalChineseQuickStart, 'hreflang="en"', "/zh-TW/docs/getting-started/quick-start/");
assertIncludes(traditionalChineseQuickStart, 'hreflang="zh-TW"', "/zh-TW/docs/getting-started/quick-start/");
assertIncludes(traditionalChineseQuickStart, 'hreflang="x-default"', "/zh-TW/docs/getting-started/quick-start/");
assertIncludes(englishQuickStart, "data-language-menu-trigger", "/docs/getting-started/quick-start/");
assertIncludes(englishQuickStart, 'href="/zh-TW/docs/getting-started/quick-start/"', "/docs/getting-started/quick-start/");
assertIncludes(sitemap, "<loc>https://netstamp.dev/docs/getting-started/quick-start/</loc>", "English Quick Start sitemap entry");
assertIncludes(sitemap, "<loc>https://netstamp.dev/zh-TW/docs/getting-started/quick-start/</loc>", "Traditional Chinese Quick Start sitemap entry");
assertExcludes(sitemap, "<loc>https://netstamp.dev/docs/</loc>", "English documentation redirect sitemap entry");
assertExcludes(sitemap, "<loc>https://netstamp.dev/zh-TW/docs/</loc>", "Traditional Chinese documentation redirect sitemap entry");
assertIncludes(traditionalChineseOpenApi, "正在載入 API 參考資料", "/zh-TW/openapi/");
assertIncludes(traditionalChineseOpenApi, "資料模型", "/zh-TW/openapi/");

console.log(
	`Localized docs architecture check passed for ${expectedDocRoutes.length} English routes, their matching Traditional Chinese routes, and ${expectedSectionKeys.length} navigation sections.`
);
