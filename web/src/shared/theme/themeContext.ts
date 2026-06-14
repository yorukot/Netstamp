import { createContext } from "react";

export type AppTheme = "dark" | "light";

export interface ThemeContextValue {
	theme: AppTheme;
	setTheme: (theme: AppTheme) => void;
	toggleTheme: () => void;
}

export const themeStorageKey = "netstamp:theme";
export const defaultTheme: AppTheme = "dark";
export const ThemeContext = createContext<ThemeContextValue | null>(null);

export function isAppTheme(value: string | null): value is AppTheme {
	return value === "dark" || value === "light";
}
