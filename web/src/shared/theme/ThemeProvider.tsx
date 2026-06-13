import { useEffect, useLayoutEffect, useMemo, useState, type ReactNode } from "react";
import { defaultTheme, isAppTheme, ThemeContext, themeStorageKey, type AppTheme, type ThemeContextValue } from "./themeContext";

function readStoredTheme(): AppTheme {
	if (typeof window === "undefined") {
		return defaultTheme;
	}

	try {
		const storedTheme = window.localStorage.getItem(themeStorageKey);
		return isAppTheme(storedTheme) ? storedTheme : defaultTheme;
	} catch {
		return defaultTheme;
	}
}

export function ThemeProvider({ children }: { children: ReactNode }) {
	const [theme, setTheme] = useState<AppTheme>(readStoredTheme);

	useLayoutEffect(() => {
		document.documentElement.dataset.theme = theme;
	}, [theme]);

	useEffect(() => {
		try {
			window.localStorage.setItem(themeStorageKey, theme);
		} catch {
			// Keep theme switching usable even when persistence is unavailable.
		}
	}, [theme]);

	const value = useMemo<ThemeContextValue>(
		() => ({
			theme,
			setTheme,
			toggleTheme: () => setTheme(currentTheme => (currentTheme === "dark" ? "light" : "dark"))
		}),
		[theme]
	);

	return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}
