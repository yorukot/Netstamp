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
const docsSectionLabels = html => Array.from(html.matchAll(/<h2 class="docsNavSectionTitle"[^>]*>([^<]+)<\/h2>/g), match => match[1]).filter(label => !["External links", "外部連結"].includes(label));
const sectionMarkup = (html, label) => {
	const headingIndex = html.indexOf(`>${label}</h2>`);
	if (headingIndex < 0) throw new Error(`Documentation navigation does not include section ${JSON.stringify(label)}`);

	const sectionStart = html.lastIndexOf("<section", headingIndex);
	const sectionEnd = html.indexOf("</section>", headingIndex);
	if (sectionStart < 0 || sectionEnd < 0) throw new Error(`Could not isolate documentation section ${JSON.stringify(label)}`);

	return html.slice(sectionStart, sectionEnd);
};

const expectedSectionKeys = ["start", "install", "use", "operate", "api", "development", "community"];
const expectedEnglishSections = ["Getting Started", "Installation", "Using Netstamp", "Operations", "API", "Development", "Community"];
const expectedTraditionalChineseSections = ["開始使用", "安裝", "使用 Netstamp", "維運", "API", "開發", "社群"];
const englishUi = JSON.parse(await readFile(path.join(process.cwd(), "docs/src/i18n/locales/en/ui.json"), "utf8"));
const traditionalChineseUi = JSON.parse(await readFile(path.join(process.cwd(), "docs/src/i18n/locales/zh-TW/ui.json"), "utf8"));

assertEqual(Object.keys(englishUi.docs.sections), expectedSectionKeys, "English documentation section keys");
assertEqual(Object.values(englishUi.docs.sections), expectedEnglishSections, "English documentation section labels");
assertEqual(Object.keys(traditionalChineseUi.docs.sections), expectedSectionKeys, "Traditional Chinese documentation section keys");
assertEqual(Object.values(traditionalChineseUi.docs.sections), expectedTraditionalChineseSections, "Traditional Chinese documentation section labels");

const englishHome = await readPage("");
const traditionalChineseHome = await readPage("zh-TW");
const englishNotFound = await readFile(path.join(dist, "404.html"), "utf8");
const traditionalChineseNotFound = await readPage("zh-TW/404.html");
const englishGuide = await readPage("docs/guides/getting-started");
const traditionalChineseGuide = await readPage("zh-TW/docs/guides/getting-started");
const traditionalChineseOpenApi = await readPage("zh-TW/openapi");
const englishDocsHome = await readPage("docs");
const traditionalChineseDocsHome = await readPage("zh-TW/docs");
const englishProduction = await readPage("docs/install/production");
const traditionalChineseProduction = await readPage("zh-TW/docs/install/production");
const englishDeploymentMarkdown = await readFile(path.join(dist, "docs/reference/deployment.md"), "utf8");
const traditionalChineseDeploymentMarkdown = await readFile(path.join(dist, "zh-TW/docs/reference/deployment.md"), "utf8");

const existingDocRoutes = [
	"docs",
	"docs/guides/api-explorer",
	"docs/guides/api-tokens",
	"docs/guides/getting-started",
	"docs/guides/probe-operations",
	"docs/guides/translating",
	"docs/reference/architecture",
	"docs/reference/configuration",
	"docs/reference/deployment",
	"docs/reference/ui-system"
];

await Promise.all(existingDocRoutes.flatMap(route => [readPage(route), readPage(`zh-TW/${route}`)]));

const englishNavHrefs = docsNavHrefs(englishDocsHome);
const traditionalChineseNavHrefs = docsNavHrefs(traditionalChineseDocsHome).map(href => href.replace(/^\/zh-TW/, ""));
assertEqual(traditionalChineseNavHrefs, englishNavHrefs, "English and Traditional Chinese documentation navigation order");

assertEqual(englishNavHrefs.slice(0, 3), ["/docs/", "/docs/guides/getting-started/", "/docs/start/concepts/"], "Getting Started page order");
assertExcludes(englishDocsHome, "--nav-depth", "English documentation navigation hierarchy");
assertExcludes(traditionalChineseDocsHome, "--nav-depth", "Traditional Chinese documentation navigation hierarchy");
assertEqual(activeDocsNavHrefs(englishDocsHome), ["/docs/"], "English documentation home active navigation item");
assertEqual(activeDocsNavHrefs(traditionalChineseDocsHome), ["/zh-TW/docs/"], "Traditional Chinese documentation home active navigation item");
assertEqual(activeDocsNavHrefs(englishProduction), ["/docs/install/production/"], "English production page active navigation item");
assertEqual(activeDocsNavHrefs(traditionalChineseProduction), ["/zh-TW/docs/install/production/"], "Traditional Chinese production page active navigation item");

const englishRenderedSections = docsSectionLabels(englishDocsHome);
const traditionalChineseRenderedSections = docsSectionLabels(traditionalChineseDocsHome);
assertEqual(englishRenderedSections, expectedEnglishSections, "Rendered English documentation section order");
assertEqual(traditionalChineseRenderedSections, expectedTraditionalChineseSections, "Rendered Traditional Chinese documentation section order");
assertIncludes(sectionMarkup(englishDocsHome, "Development"), 'href="/docs/guides/translating/"', "English Development navigation section");
assertIncludes(sectionMarkup(traditionalChineseDocsHome, "開發"), 'href="/zh-TW/docs/guides/translating/"', "Traditional Chinese Development navigation section");

assertIncludes(englishHome, '<html lang="en"', "/");
assertIncludes(traditionalChineseHome, '<html lang="zh-Hant-TW"', "/zh-TW/");
assertIncludes(traditionalChineseHome, "追蹤每一條路徑", "/zh-TW/");
assertIncludes(englishNotFound, "Page not found - Netstamp", "/404.html");
assertIncludes(englishNotFound, 'content="noindex,nofollow,noarchive"', "/404.html");
assertIncludes(englishNotFound, 'href="/zh-TW/404.html"', "/404.html");
assertIncludes(traditionalChineseNotFound, '<html lang="zh-Hant-TW"', "/zh-TW/404.html/");
assertIncludes(traditionalChineseNotFound, "找不到頁面", "/zh-TW/404.html/");
assertIncludes(traditionalChineseNotFound, 'href="/404.html"', "/zh-TW/404.html/");
assertIncludes(englishGuide, "Quick start - Netstamp Docs", "/docs/guides/getting-started/");
assertIncludes(traditionalChineseGuide, "快速開始 - Netstamp 說明文件", "/zh-TW/docs/guides/getting-started/");
assertIncludes(traditionalChineseGuide, 'href="/docs/guides/getting-started/"', "/zh-TW/docs/guides/getting-started/");
assertIncludes(traditionalChineseGuide, 'hreflang="en"', "/zh-TW/docs/guides/getting-started/");
assertIncludes(traditionalChineseGuide, 'hreflang="zh-TW"', "/zh-TW/docs/guides/getting-started/");
assertIncludes(traditionalChineseGuide, 'hreflang="x-default"', "/zh-TW/docs/guides/getting-started/");
assertIncludes(traditionalChineseGuide, 'href="https://netstamp.dev/zh-TW/docs/guides/getting-started/"', "/zh-TW/docs/guides/getting-started/");
assertIncludes(traditionalChineseGuide, 'href="/zh-TW/docs/reference/architecture/"', "/zh-TW/docs/guides/getting-started/");
assertExcludes(traditionalChineseGuide, 'href="/docs/reference/architecture/"', "/zh-TW/docs/guides/getting-started/");
assertIncludes(traditionalChineseGuide, 'data-copy-label="複製"', "/zh-TW/docs/guides/getting-started/");
assertIncludes(traditionalChineseGuide, 'data-copied-label="已複製"', "/zh-TW/docs/guides/getting-started/");
assertIncludes(traditionalChineseOpenApi, "正在載入 API 參考資料", "/zh-TW/openapi/");
assertIncludes(traditionalChineseOpenApi, "資料模型", "/zh-TW/openapi/");
assertIncludes(englishProduction, 'href="https://github.com/yorukot/Netstamp/blob/main/docs/src/content/docs/en/install/production.mdx"', "/docs/install/production/");
assertIncludes(traditionalChineseProduction, 'href="https://crowdin.com/project/netstamp"', "/zh-TW/docs/install/production/");
assertIncludes(englishDeploymentMarkdown, "# Docker Compose", "/docs/reference/deployment.md");
assertIncludes(englishDeploymentMarkdown, "| Service", "/docs/reference/deployment.md");
assertIncludes(englishDeploymentMarkdown, "```bash", "/docs/reference/deployment.md");
assertIncludes(englishDeploymentMarkdown, "`.env`", "/docs/reference/deployment.md");
assertIncludes(englishDeploymentMarkdown, "[Production deployment](/docs/install/production/)", "/docs/reference/deployment.md");
assertIncludes(traditionalChineseDeploymentMarkdown, "# Docker Compose", "/zh-TW/docs/reference/deployment.md");
assertIncludes(traditionalChineseDeploymentMarkdown, "| 服務", "/zh-TW/docs/reference/deployment.md");
assertIncludes(traditionalChineseDeploymentMarkdown, "```bash", "/zh-TW/docs/reference/deployment.md");
assertIncludes(traditionalChineseDeploymentMarkdown, "`.env`", "/zh-TW/docs/reference/deployment.md");
assertIncludes(traditionalChineseDeploymentMarkdown, "[正式環境部署](/zh-TW/docs/install/production/)", "/zh-TW/docs/reference/deployment.md");

console.log("Localized docs output check passed for English and Traditional Chinese IA, routes, active navigation, Markdown payloads, metadata, links, and OpenAPI UI.");
