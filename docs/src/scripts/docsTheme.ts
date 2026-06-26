type Theme = "light" | "dark";

const themeStorageKey = "netstamp:theme";
const eventBoundKey = "__netstampThemeEventsBound";
const mediaBoundKey = "__netstampThemeMediaBound";
type ThemeWindow = Window & {
	[eventBoundKey]?: boolean;
	[mediaBoundKey]?: boolean;
};

function isTheme(theme: string | null | undefined): theme is Theme {
	return theme === "light" || theme === "dark";
}

function readStoredTheme(): Theme | null {
	try {
		const storedTheme = window.localStorage.getItem(themeStorageKey);
		return isTheme(storedTheme) ? storedTheme : null;
	} catch {
		return null;
	}
}

function writeStoredTheme(theme: Theme) {
	try {
		window.localStorage.setItem(themeStorageKey, theme);
	} catch {
		// Persistence can be unavailable; keep the current page state usable.
	}
}

function readSystemTheme(): Theme {
	return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

function resolveTheme(): Theme {
	return readStoredTheme() ?? readSystemTheme();
}

function readAppliedTheme(): Theme {
	const theme = document.documentElement.dataset.theme;
	return isTheme(theme) ? theme : resolveTheme();
}

function syncThemeToggle(themeToggle: HTMLElement, theme: Theme) {
	const isDark = theme === "dark";
	const nextTheme = isDark ? "light" : "dark";
	const label = nextTheme === "light" ? "Switch to light mode" : "Switch to dark mode";

	themeToggle.setAttribute("aria-label", label);
	themeToggle.setAttribute("title", label);
	themeToggle.setAttribute("aria-pressed", String(isDark));

	if (themeToggle.getAttribute("role") === "switch" || themeToggle.getAttribute("role") === "checkbox") {
		themeToggle.setAttribute("aria-checked", String(isDark));
	}

	if (themeToggle instanceof HTMLInputElement && (themeToggle.type === "checkbox" || themeToggle.type === "radio")) {
		themeToggle.checked = isDark;
	}
}

export function applyTheme(theme: Theme = resolveTheme()) {
	document.documentElement.classList.toggle("dark", theme === "dark");
	document.documentElement.dataset.theme = theme;
	document.querySelectorAll<HTMLElement>("[data-theme-toggle]").forEach(themeToggle => syncThemeToggle(themeToggle, theme));

	return theme;
}

export function setupThemeToggle() {
	document.querySelectorAll<HTMLElement>("[data-theme-toggle]").forEach(themeToggle => {
		syncThemeToggle(themeToggle, readAppliedTheme());
		if (themeToggle.dataset.themeBound === "true") return;

		themeToggle.dataset.themeBound = "true";
		themeToggle.addEventListener("click", () => {
			const nextTheme = readAppliedTheme() === "dark" ? "light" : "dark";

			writeStoredTheme(nextTheme);
			applyTheme(nextTheme);
		});
	});
}

function initDocsTheme() {
	applyTheme();
	setupThemeToggle();
}

function bindSystemThemeSync() {
	const themeWindow = window as ThemeWindow;
	if (themeWindow[mediaBoundKey]) return;

	const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
	const handleSystemThemeChange = () => {
		if (readStoredTheme() === null) {
			applyTheme();
		}
	};

	if ("addEventListener" in mediaQuery) {
		mediaQuery.addEventListener("change", handleSystemThemeChange);
	} else {
		mediaQuery.addListener(handleSystemThemeChange);
	}
	themeWindow[mediaBoundKey] = true;
}

initDocsTheme();
bindSystemThemeSync();

const themeWindow = window as ThemeWindow;
if (!themeWindow[eventBoundKey]) {
	document.addEventListener("astro:after-swap", () => {
		applyTheme();
	});
	document.addEventListener("astro:page-load", initDocsTheme);
	themeWindow[eventBoundKey] = true;
}
