interface SearchEntry {
	title: string;
	description: string;
	keywords?: string | string[];
	content?: string;
	href: string;
}

let cleanupSearch = () => {};

function normalized(value: string) {
	return value.toLocaleLowerCase().replace(/\s+/g, " ").trim();
}

function keywordsText(keywords: SearchEntry["keywords"]) {
	return Array.isArray(keywords) ? keywords.join(" ") : (keywords ?? "");
}

function scoreSearchEntry(entry: SearchEntry, query: string) {
	const title = normalized(entry.title);
	const description = normalized(entry.description);
	const keywords = normalized(keywordsText(entry.keywords));
	const content = normalized(entry.content ?? "");
	const href = normalized(entry.href);
	const terms = normalized(query).split(" ").filter(Boolean);

	if (!terms.length) return 0;

	let score = 0;
	for (const term of terms) {
		if (title === term) score += 120;
		if (title.startsWith(term)) score += 80;
		if (title.includes(term)) score += 60;
		if (keywords.includes(term)) score += 36;
		if (href.includes(term)) score += 28;
		if (description.includes(term)) score += 20;
		if (content.includes(term)) score += 8;
	}

	return score;
}

function initDocsSearch() {
	cleanupSearch();

	const root = document.querySelector("[data-docs-search]");
	if (!(root instanceof HTMLElement)) return;

	const trigger = root.querySelector("[data-search-open]");
	const dialog = root.querySelector("[data-search-dialog]");
	const panel = root.querySelector(".docsSearchPanel");
	const input = root.querySelector("[data-search-input]");
	const results = root.querySelector("[data-search-results]");
	const closeButtons = root.querySelectorAll("[data-search-close]");
	const entries = JSON.parse(root.dataset.searchIndex ?? "[]") as SearchEntry[];

	let restoreFocusTo: Element | null = null;
	let closeTimer: number | undefined;
	let activeResultIndex = -1;

	function clearResults() {
		if (!results) return;

		activeResultIndex = -1;
		results.replaceChildren();
		(results as HTMLElement).hidden = true;
	}

	function resultLinks() {
		return Array.from(results?.querySelectorAll("[data-search-result]") ?? []).filter((element): element is HTMLAnchorElement => element instanceof HTMLAnchorElement);
	}

	function focusablePanelElements() {
		if (!(panel instanceof HTMLElement)) return [];

		return Array.from(panel.querySelectorAll<HTMLElement>('a[href], button:not([disabled]), input:not([disabled]), [tabindex]:not([tabindex="-1"])')).filter(
			element => !element.hidden && element.offsetParent !== null
		);
	}

	function focusResult(index: number) {
		const links = resultLinks();

		if (!links.length) return;

		activeResultIndex = (index + links.length) % links.length;
		links[activeResultIndex]?.focus();
	}

	function openSearch() {
		if (!dialog || !input) return;
		const dialogEl = dialog as HTMLElement;
		const triggerEl = trigger as HTMLButtonElement | null;

		if (!dialogEl.hidden && dialogEl.dataset.state === "open") return;

		if (closeTimer) {
			window.clearTimeout(closeTimer);
			closeTimer = undefined;
		}

		restoreFocusTo = document.activeElement;
		dialogEl.hidden = false;
		dialogEl.dataset.state = "open";
		triggerEl?.setAttribute("aria-expanded", "true");
		document.documentElement.classList.add("docsSearchOpen");
		(input as HTMLInputElement).value = "";
		clearResults();
		requestAnimationFrame(() => (input as HTMLElement).focus());
	}

	function closeSearch() {
		if (!dialog || !input) return;
		const dialogEl = dialog as HTMLElement;
		const triggerEl = trigger as HTMLButtonElement | null;

		if (dialogEl.hidden || dialogEl.dataset.state === "closing") return;

		dialogEl.dataset.state = "closing";
		triggerEl?.setAttribute("aria-expanded", "false");
		(input as HTMLInputElement).value = "";
		clearResults();

		closeTimer = window.setTimeout(() => {
			dialogEl.hidden = true;
			dialogEl.dataset.state = "closed";
			document.documentElement.classList.remove("docsSearchOpen");
			closeTimer = undefined;

			if (restoreFocusTo instanceof HTMLElement) {
				restoreFocusTo.focus();
			}
		}, 260);
	}

	function renderResults(matches: SearchEntry[]) {
		if (!results) return;

		results.replaceChildren();

		if (!matches.length) {
			const empty = document.createElement("div");
			empty.className = "docsSearchEmpty";
			empty.setAttribute("role", "status");
			empty.textContent = root.dataset.searchEmpty || "No matches";
			results.append(empty);
			(results as HTMLElement).hidden = false;
			return;
		}

		const fragment = document.createDocumentFragment();
		for (const entry of matches) {
			const link = document.createElement("a");
			const title = document.createElement("strong");
			const description = document.createElement("span");

			link.className = "docsSearchResult";
			link.href = entry.href;
			link.setAttribute("data-search-result", "");
			title.textContent = entry.title;
			description.textContent = entry.description;
			link.append(title, description);
			fragment.append(link);
		}

		results.append(fragment);
		(results as HTMLElement).hidden = false;
	}

	function handleInput(event: Event) {
		const query = (event.currentTarget as HTMLInputElement).value.trim();
		if (!query) {
			clearResults();
			return;
		}

		const matches = entries
			.map(entry => ({ entry, score: scoreSearchEntry(entry, query) }))
			.filter(result => result.score > 0)
			.sort((left, right) => right.score - left.score || left.entry.title.localeCompare(right.entry.title))
			.map(result => result.entry)
			.slice(0, 6);
		renderResults(matches);
	}

	function handleSearchKeydown(event: KeyboardEvent) {
		if (event.key === "ArrowDown") {
			event.preventDefault();
			focusResult(0);
			return;
		}

		if (event.key === "Escape") {
			event.preventDefault();
			closeSearch();
		}
	}

	function handleResultsKeydown(event: KeyboardEvent) {
		if (event.key === "ArrowDown") {
			event.preventDefault();
			focusResult(activeResultIndex + 1);
			return;
		}

		if (event.key === "ArrowUp") {
			event.preventDefault();
			if (activeResultIndex <= 0) {
				(input as HTMLElement | null)?.focus();
				activeResultIndex = -1;
				return;
			}
			focusResult(activeResultIndex - 1);
			return;
		}

		if (event.key === "Escape") {
			event.preventDefault();
			closeSearch();
		}
	}

	function handleDialogKeydown(event: KeyboardEvent) {
		if (event.key !== "Tab" || !dialog || (dialog as HTMLElement).hidden) {
			return;
		}

		const focusable = focusablePanelElements();
		const first = focusable[0];
		const last = focusable[focusable.length - 1];

		if (!first || !last) {
			event.preventDefault();
			return;
		}

		if (event.shiftKey && document.activeElement === first) {
			event.preventDefault();
			last.focus();
			return;
		}

		if (!event.shiftKey && document.activeElement === last) {
			event.preventDefault();
			first.focus();
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
			event.preventDefault();
			openSearch();
			return;
		}

		if (event.key === "Escape" && dialog && !(dialog as HTMLElement).hidden) {
			closeSearch();
		}
	}

	trigger?.addEventListener("click", openSearch);
	closeButtons.forEach(button => button.addEventListener("click", closeSearch));
	input?.addEventListener("input", handleInput);
	input?.addEventListener("keydown", handleSearchKeydown);
	results?.addEventListener("keydown", handleResultsKeydown);
	dialog?.addEventListener("keydown", handleDialogKeydown);
	document.addEventListener("keydown", handleKeydown);

	cleanupSearch = () => {
		if (closeTimer) window.clearTimeout(closeTimer);
		trigger?.removeEventListener("click", openSearch);
		closeButtons.forEach(button => button.removeEventListener("click", closeSearch));
		input?.removeEventListener("input", handleInput);
		input?.removeEventListener("keydown", handleSearchKeydown);
		results?.removeEventListener("keydown", handleResultsKeydown);
		dialog?.removeEventListener("keydown", handleDialogKeydown);
		document.removeEventListener("keydown", handleKeydown);
		document.documentElement.classList.remove("docsSearchOpen");
		cleanupSearch = () => {};
	};
}

initDocsSearch();
document.addEventListener("astro:page-load", initDocsSearch);
