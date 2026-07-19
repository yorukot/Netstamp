// @vitest-environment jsdom

import { pageTitleFromMatches, pageTitleHandle } from "@/routes/pageTitles";
import { ApiError } from "@/shared/api/client";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
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
	it("renders English and Traditional Chinese resources", async () => {
		await changeLocale("en");
		const view = render(<NavigationLabel />);
		expect(screen.getByText("Overview")).toBeTruthy();

		await changeLocale("zh-TW");
		view.rerender(<NavigationLabel />);
		expect(screen.getByText("總覽")).toBeTruthy();
	});

	it("switches immediately, persists the locale, and updates html lang", async () => {
		await changeLocale("en");
		render(<LanguageSwitcher />);
		fireEvent.click(screen.getByRole("button", { name: /繁體中文/ }));

		await screen.findByRole("button", { name: /English/ });
		expect(i18n.resolvedLanguage).toBe("zh-TW");
		expect(window.localStorage.getItem("netstamp:locale")).toBe("zh-TW");
		expect(document.documentElement.lang).toBe("zh-Hant-TW");
	});

	it("restores an explicit locale before checking browser languages", () => {
		setBrowserLanguages(["en-US"]);
		window.localStorage.setItem("netstamp:locale", "zh-TW");
		expect(detectInitialLocale()).toBe("zh-TW");
	});

	it("maps Traditional Chinese browser locales and falls back for unsupported locales", () => {
		window.localStorage.clear();
		setBrowserLanguages(["zh-HK"]);
		expect(detectInitialLocale()).toBe("zh-TW");

		setBrowserLanguages(["fr-FR"]);
		expect(detectInitialLocale()).toBe("en");
	});

	it("interpolates values and falls back to English for a missing translation", async () => {
		await changeLocale("zh-TW");
		expect(i18n.t("openUserMenu", { ns: "navigation", name: "Elvis" })).toBe("開啟 Elvis 的使用者選單");

		i18n.addResource("en", "common", "testOnlyFallback", "English fallback");
		const translateDynamicCommonKey = i18n.getFixedT(null, "common") as (key: string) => string;
		expect(translateDynamicCommonKey("testOnlyFallback")).toBe("English fallback");
		i18n.removeResourceBundle("en", "common");
		i18n.addResourceBundle("en", "common", (await import("./locales/en/common.json")).default, true, true);
	});

	it("localizes route page titles", async () => {
		const matches = [{ handle: pageTitleHandle("pageTitles.login") }];
		const translate = (key: Parameters<typeof pageTitleHandle>[0]) => i18n.t(key, { ns: "navigation" });

		await changeLocale("en");
		expect(pageTitleFromMatches(matches, translate)).toBe("Log in - Netstamp");

		await changeLocale("zh-TW");
		expect(pageTitleFromMatches(matches, translate)).toBe("登入 - Netstamp");
	});

	it("localizes generic API and network errors without exposing English transport text", async () => {
		await changeLocale("zh-TW");
		expect(requestErrorMessage(new ApiError("Bad Gateway", 502))).toBe("Netstamp 控制器回傳伺服器錯誤。");
		expect(requestErrorMessage(new TypeError("Failed to fetch"))).toBe("無法連線至 Netstamp 控制器。");
		expect(requestErrorMessage(new Error("Internal English detail"), "無法儲存設定。")).toBe("無法儲存設定。");
	});
});
