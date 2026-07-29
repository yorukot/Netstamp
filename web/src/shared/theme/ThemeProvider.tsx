import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState, type ReactNode } from "react";
import { defaultTheme, isAppTheme, ThemeContext, themeStorageKey, type AppTheme, type ThemeContextValue } from "./themeContext";

const readStoredTheme = (): AppTheme => {
	if (typeof window === "undefined") {
		return defaultTheme;
	}

	try {
		const storedTheme = window.localStorage.getItem(themeStorageKey);
		return isAppTheme(storedTheme) ? storedTheme : defaultTheme;
	} catch {
		return defaultTheme;
	}
};

const applyDocumentTheme = (theme: AppTheme) => {
	if (typeof document !== "undefined") {
		document.documentElement.dataset.theme = theme;
	}
};

export const ThemeProvider = ({ children }: { children: ReactNode }) => {
	const [theme, setThemeState] = useState<AppTheme>(() => {
		const initialTheme = readStoredTheme();
		applyDocumentTheme(initialTheme);
		return initialTheme;
	});
	const themeRef = useRef(theme);

	const setTheme = useCallback((nextTheme: AppTheme) => {
		themeRef.current = nextTheme;
		applyDocumentTheme(nextTheme);
		setThemeState(nextTheme);
	}, []);

	const toggleTheme = useCallback(() => {
		setTheme(themeRef.current === "dark" ? "light" : "dark");
	}, [setTheme]);

	useLayoutEffect(() => {
		themeRef.current = theme;
		applyDocumentTheme(theme);
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
			toggleTheme
		}),
		[setTheme, theme, toggleTheme]
	);

	return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
};
