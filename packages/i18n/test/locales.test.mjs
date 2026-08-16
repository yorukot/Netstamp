import assert from "node:assert/strict";
import test from "node:test";
import { alternateLocalePaths, defaultLocale, htmlLangForLocale, isSupportedLocale, localeFromPath, localePath, normalizeLocale, resolveLocale, stripLocalePrefix } from "../dist/index.js";

test("validates and normalizes supported locales", () => {
	assert.equal(isSupportedLocale("en"), true);
	assert.equal(isSupportedLocale("zh-TW"), false);
	assert.equal(isSupportedLocale("zh"), false);
	assert.equal(normalizeLocale("zh-Hant"), undefined);
	assert.equal(normalizeLocale("zh_HK"), undefined);
	assert.equal(normalizeLocale("en-US"), "en");
	assert.equal(normalizeLocale("fr-FR"), undefined);
});

test("resolves candidates in order and falls back to English", () => {
	assert.equal(resolveLocale(["fr-FR", "zh-Hant-TW", "en"]), "en");
	assert.equal(resolveLocale(["fr-FR", undefined]), defaultLocale);
	assert.equal(htmlLangForLocale("en"), "en");
});

test("keeps English routes unprefixed and leaves unsupported prefixes untouched", () => {
	assert.equal(localePath("/docs/getting-started/?tab=agent#install", "en"), "/docs/getting-started/?tab=agent#install");
	assert.equal(stripLocalePrefix("/en/docs/getting-started/"), "/docs/getting-started/");
	assert.equal(stripLocalePrefix("/zh-TW/openapi/?view=client"), "/zh-TW/openapi/?view=client");
	assert.equal(localeFromPath("/zh-TW/docs/"), "en");
	assert.equal(localeFromPath("/docs/"), "en");
	assert.deepEqual(alternateLocalePaths("/docs/"), { en: "/docs/" });
});
