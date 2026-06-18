const themeStorageKey = "netstamp:theme";
let cleanupDocsTheme = () => {};

function readDocsTheme() {
	const theme = document.documentElement.dataset.theme;
	return theme === "light" || theme === "dark" ? theme : "light";
}

function writeDocsTheme(theme: "light" | "dark") {
	document.documentElement.dataset.theme = theme;

	try {
		window.localStorage.setItem(themeStorageKey, theme);
	} catch {
		// Persistence can be unavailable; keep the current page state usable.
	}
}

function initDocsTheme() {
	cleanupDocsTheme();

	const themeToggle = document.querySelector("[data-theme-toggle]");
	if (!themeToggle) return;

	function syncThemeToggle() {
		const theme = readDocsTheme();
		const nextTheme = theme === "dark" ? "light" : "dark";
		const label = nextTheme === "light" ? "Switch to light mode" : "Switch to dark mode";

		themeToggle.setAttribute("aria-label", label);
		themeToggle.setAttribute("title", label);
		themeToggle.setAttribute("aria-pressed", String(theme === "light"));
	}

	const handleThemeToggle = () => {
		writeDocsTheme(readDocsTheme() === "dark" ? "light" : "dark");
		syncThemeToggle();
	};

	syncThemeToggle();
	themeToggle.addEventListener("click", handleThemeToggle);
	cleanupDocsTheme = () => {
		themeToggle.removeEventListener("click", handleThemeToggle);
		cleanupDocsTheme = () => {};
	};
}

initDocsTheme();
document.addEventListener("astro:page-load", initDocsTheme);
