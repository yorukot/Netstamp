type Theme = "light" | "dark";

const themeStorageKey = "netstamp:theme";
const eventBoundKey = "__netstampThemeEventsBound";
const defaultTheme: Theme = "dark";
type ThemeWindow = Window & {
	[eventBoundKey]?: boolean;
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

function readLockedTheme(): Theme | null {
	const lockedTheme = document.documentElement.dataset.themeLocked;
	return isTheme(lockedTheme) ? lockedTheme : null;
}

function writeStoredTheme(theme: Theme) {
	try {
		window.localStorage.setItem(themeStorageKey, theme);
	} catch {
		// Persistence can be unavailable; keep the current page state usable.
	}
}

function resolveTheme(): Theme {
	return readLockedTheme() ?? readStoredTheme() ?? defaultTheme;
}

function readAppliedTheme(): Theme {
	const theme = document.documentElement.dataset.theme;
	return isTheme(theme) ? theme : resolveTheme();
}

function syncThemeToggle(themeToggle: HTMLElement, theme: Theme) {
	const isDark = theme === "dark";
	const nextTheme = isDark ? "light" : "dark";
	const label = nextTheme === "light" ? themeToggle.dataset.themeLightLabel : themeToggle.dataset.themeDarkLabel;

	if (label) {
		themeToggle.setAttribute("aria-label", label);
		themeToggle.setAttribute("title", label);
	}
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
	const lockedTheme = readLockedTheme();

	document.querySelectorAll<HTMLElement>("[data-theme-toggle]").forEach(themeToggle => {
		syncThemeToggle(themeToggle, readAppliedTheme());
		if (lockedTheme) {
			themeToggle.hidden = true;
			return;
		}
		themeToggle.hidden = false;
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

initDocsTheme();

const themeWindow = window as ThemeWindow;
if (!themeWindow[eventBoundKey]) {
	document.addEventListener("astro:after-swap", () => {
		applyTheme();
	});
	document.addEventListener("astro:page-load", initDocsTheme);
	themeWindow[eventBoundKey] = true;
}
