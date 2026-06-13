import { useContext } from "react";
import { ThemeContext } from "./themeContext";

export function useTheme() {
	const value = useContext(ThemeContext);

	if (!value) {
		throw new Error("useTheme must be used within ThemeProvider");
	}

	return value;
}
