// @vitest-environment jsdom

import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { ThemeProvider } from "./ThemeProvider";
import { themeStorageKey } from "./themeContext";
import { useTheme } from "./useTheme";

const ThemeProbe = () => {
	const { setTheme, theme, toggleTheme } = useTheme();

	return (
		<div data-root-theme-at-render={document.documentElement.dataset.theme}>
			<span>{theme}</span>
			<button type="button" onClick={toggleTheme}>
				toggle
			</button>
			<button type="button" onClick={() => setTheme("light")}>
				set light
			</button>
		</div>
	);
};

const renderThemeProvider = () =>
	render(
		<ThemeProvider>
			<ThemeProbe />
		</ThemeProvider>
	);

afterEach(() => {
	cleanup();
	window.localStorage.clear();
	delete document.documentElement.dataset.theme;
});

describe("ThemeProvider", () => {
	it("applies the default theme before descendants render", () => {
		renderThemeProvider();

		expect(screen.getByText("dark")).toBeTruthy();
		expect(screen.getByText("dark").parentElement?.dataset.rootThemeAtRender).toBe("dark");
		expect(document.documentElement.dataset.theme).toBe("dark");
	});

	it("restores a stored theme before descendants render", () => {
		window.localStorage.setItem(themeStorageKey, "light");
		renderThemeProvider();

		expect(screen.getByText("light")).toBeTruthy();
		expect(screen.getByText("light").parentElement?.dataset.rootThemeAtRender).toBe("light");
		expect(document.documentElement.dataset.theme).toBe("light");
	});

	it("keeps the document, context, and persisted theme synchronized", () => {
		renderThemeProvider();

		fireEvent.click(screen.getByRole("button", { name: "toggle" }));
		expect(screen.getByText("light")).toBeTruthy();
		expect(document.documentElement.dataset.theme).toBe("light");
		expect(window.localStorage.getItem(themeStorageKey)).toBe("light");

		fireEvent.click(screen.getByRole("button", { name: "toggle" }));
		expect(screen.getByText("dark")).toBeTruthy();
		expect(document.documentElement.dataset.theme).toBe("dark");

		fireEvent.click(screen.getByRole("button", { name: "set light" }));
		expect(screen.getByText("light")).toBeTruthy();
		expect(document.documentElement.dataset.theme).toBe("light");
	});
});
