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

const englishHome = await readPage("");
const traditionalChineseHome = await readPage("zh-TW");
const englishNotFound = await readFile(path.join(dist, "404.html"), "utf8");
const traditionalChineseNotFound = await readPage("zh-TW/404.html");
const englishGuide = await readPage("docs/guides/getting-started");
const traditionalChineseGuide = await readPage("zh-TW/docs/guides/getting-started");
const traditionalChineseOpenApi = await readPage("zh-TW/openapi");

assertIncludes(englishHome, '<html lang="en"', "/");
assertIncludes(traditionalChineseHome, '<html lang="zh-Hant-TW"', "/zh-TW/");
assertIncludes(traditionalChineseHome, "追蹤每一條路徑", "/zh-TW/");
assertIncludes(englishNotFound, "Page not found - Netstamp", "/404.html");
assertIncludes(englishNotFound, 'content="noindex,nofollow,noarchive"', "/404.html");
assertIncludes(englishNotFound, 'href="/zh-TW/404.html"', "/404.html");
assertIncludes(traditionalChineseNotFound, '<html lang="zh-Hant-TW"', "/zh-TW/404.html/");
assertIncludes(traditionalChineseNotFound, "找不到頁面", "/zh-TW/404.html/");
assertIncludes(traditionalChineseNotFound, 'href="/404.html"', "/zh-TW/404.html/");
assertIncludes(englishGuide, "Getting Started | Netstamp Docs", "/docs/guides/getting-started/");
assertIncludes(traditionalChineseGuide, "開始使用 | Netstamp 說明文件", "/zh-TW/docs/guides/getting-started/");
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

console.log("Localized docs output check passed for English and Traditional Chinese routes, metadata, links, and OpenAPI UI.");
