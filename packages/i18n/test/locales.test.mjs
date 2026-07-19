import assert from "node:assert/strict";
import test from "node:test";
import { alternateLocalePaths, defaultLocale, htmlLangForLocale, isSupportedLocale, localeFromPath, localePath, normalizeLocale, resolveLocale, stripLocalePrefix } from "../dist/index.js";

test("validates and normalizes supported locales", () => {
	assert.equal(isSupportedLocale("en"), true);
	assert.equal(isSupportedLocale("zh-TW"), true);
	assert.equal(isSupportedLocale("zh"), false);
	assert.equal(normalizeLocale("zh-Hant"), "zh-TW");
	assert.equal(normalizeLocale("zh_HK"), "zh-TW");
	assert.equal(normalizeLocale("en-US"), "en");
	assert.equal(normalizeLocale("fr-FR"), undefined);
});

test("resolves candidates in order and falls back to English", () => {
	assert.equal(resolveLocale(["fr-FR", "zh-Hant-TW", "en"]), "zh-TW");
	assert.equal(resolveLocale(["fr-FR", undefined]), defaultLocale);
	assert.equal(htmlLangForLocale("zh-TW"), "zh-Hant-TW");
});

test("converts locale URLs without changing the English public route", () => {
	assert.equal(localePath("/docs/getting-started/?tab=agent#install", "zh-TW"), "/zh-TW/docs/getting-started/?tab=agent#install");
	assert.equal(localePath("/zh-TW/docs/getting-started/", "en"), "/docs/getting-started/");
	assert.equal(localePath("/", "zh-TW"), "/zh-TW/");
	assert.equal(stripLocalePrefix("/zh-TW/openapi/?view=client"), "/openapi/?view=client");
	assert.equal(localeFromPath("/zh-TW/docs/"), "zh-TW");
	assert.equal(localeFromPath("/docs/"), "en");
	assert.deepEqual(alternateLocalePaths("/docs/"), { en: "/docs/", "zh-TW": "/zh-TW/docs/" });
});
