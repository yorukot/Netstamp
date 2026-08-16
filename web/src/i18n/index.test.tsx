// @vitest-environment jsdom

import { pageTitleFromMatches, pageTitleHandle } from "@/routes/pageTitles";
import { ApiError } from "@/shared/api/client";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { cleanup, render, screen } from "@testing-library/react";
import { useTranslation } from "react-i18next";
import { afterEach, beforeAll, describe, expect, it } from "vitest";
import { changeLocale, detectInitialLocale, i18n, initializeI18n } from "./index";
import { LanguageSwitcher } from "./LanguageSwitcher";

const NavigationLabel = () => {
	const { t } = useTranslation("navigation");
	return <span>{t("overview")}</span>;
};

const storageValues = new Map<string, string>();
const localStorageMock: Storage = {
	get length() {
		return storageValues.size;
	},
	clear: () => storageValues.clear(),
	getItem: key => storageValues.get(key) ?? null,
	key: index => [...storageValues.keys()][index] ?? null,
	removeItem: key => storageValues.delete(key),
	setItem: (key, value) => storageValues.set(key, String(value))
};

const setBrowserLanguages = (languages: string[]) => {
	Object.defineProperty(window.navigator, "languages", { configurable: true, value: languages });
	Object.defineProperty(window.navigator, "language", { configurable: true, value: languages[0] ?? "en" });
};

beforeAll(async () => {
	Object.defineProperty(window, "localStorage", { configurable: true, value: localStorageMock });
	window.localStorage.clear();
	setBrowserLanguages(["en-US"]);
	await initializeI18n();
});

afterEach(async () => {
	cleanup();
	await changeLocale("en");
	window.localStorage.clear();
	setBrowserLanguages(["en-US"]);
});

describe("React i18n", () => {
	it("renders English resources", async () => {
		await changeLocale("en");
		render(<NavigationLabel />);
		expect(screen.getByText("Overview")).toBeTruthy();
	});

	it("hides the language switcher while only one locale is enabled", () => {
		render(<LanguageSwitcher />);
		expect(screen.queryByRole("button", { name: "Language" })).toBeNull();
	});

	it("ignores a stored locale that is no longer supported", () => {
		setBrowserLanguages(["en-US"]);
		window.localStorage.setItem("netstamp:locale", "zh-TW");
		expect(detectInitialLocale()).toBe("en");
	});

	it("falls back to English for unsupported browser locales", () => {
		window.localStorage.clear();
		setBrowserLanguages(["zh-HK"]);
		expect(detectInitialLocale()).toBe("en");

		setBrowserLanguages(["fr-FR"]);
		expect(detectInitialLocale()).toBe("en");
	});

	it("interpolates values from the English resources", () => {
		expect(i18n.t("openUserMenu", { ns: "navigation", name: "Elvis" })).toBe("Open user menu for Elvis");
	});

	it("uses English route page titles", () => {
		const matches = [{ handle: pageTitleHandle("pageTitles.login") }];
		const translate = (key: Parameters<typeof pageTitleHandle>[0]) => i18n.t(key, { ns: "navigation" });

		expect(pageTitleFromMatches(matches, translate)).toBe("Log in - Netstamp");
	});

	it("uses English API, network, and runtime error messages", () => {
		expect(requestErrorMessage(new ApiError("Bad Gateway", 502))).toBe("Bad Gateway");
		expect(requestErrorMessage(new TypeError("Failed to fetch"))).toBe("Unable to reach the Netstamp controller.");
		expect(requestErrorMessage(new Error("Internal English detail"), "Unable to save settings.")).toBe("Internal English detail");
	});
});
