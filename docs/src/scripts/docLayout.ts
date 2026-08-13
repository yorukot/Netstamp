let cleanupDocLayout = () => {};
let preservedDocsSidebarScrollTop: number | undefined;

const normalizedPathname = (pathname: string) => (pathname.endsWith("/") ? pathname : `${pathname}/`);

const syncDocsNavActiveState = () => {
	const sidebar = document.querySelector("[data-docs-sidebar]");
	if (!(sidebar instanceof HTMLElement)) return;

	const pathname = normalizedPathname(window.location.pathname);
	const navItems = Array.from(sidebar.querySelectorAll<HTMLAnchorElement>("[data-docs-nav-item]"));

	for (const navItem of navItems) {
		const isActive = normalizedPathname(new URL(navItem.href, window.location.origin).pathname) === pathname;
		navItem.classList.toggle("active", isActive);

		if (isActive) {
			navItem.setAttribute("aria-current", "page");
		} else {
			navItem.removeAttribute("aria-current");
		}
	}

	for (const section of sidebar.querySelectorAll("[data-docs-nav-section]")) {
		section.classList.toggle("active", Boolean(section.querySelector('[data-docs-nav-item][aria-current="page"]')));
	}
};

const prepareDocsSidebarSwap = (event: DocumentEventMap["astro:before-swap"]) => {
	preservedDocsSidebarScrollTop = undefined;

	const currentSidebar = document.querySelector<HTMLElement>("[data-docs-sidebar]");
	const nextSidebar = event.newDocument.querySelector<HTMLElement>("[data-docs-sidebar]");
	if (!currentSidebar || !nextSidebar) return;

	if (currentSidebar.dataset.docsLocale === nextSidebar.dataset.docsLocale) {
		preservedDocsSidebarScrollTop = currentSidebar.scrollTop;
		return;
	}

	nextSidebar.removeAttribute("data-astro-transition-persist");
};

const completeDocsSidebarSwap = () => {
	syncDocsNavActiveState();

	const sidebar = document.querySelector<HTMLElement>("[data-docs-sidebar]");
	if (sidebar) {
		sidebar.setAttribute("data-astro-transition-persist", "docs-sidebar");
		if (preservedDocsSidebarScrollTop !== undefined) sidebar.scrollTop = preservedDocsSidebarScrollTop;
	}
	preservedDocsSidebarScrollTop = undefined;
};

function initDocLayout() {
	cleanupDocLayout();
	syncDocsNavActiveState();

	const cleanupTasks: Array<() => void> = [];
	const mobileNavToggle = document.querySelector("[data-docs-mobile-nav-toggle]");
	const mobileNavPanel = document.querySelector("[data-docs-nav-panel]");
	const pageActions = document.querySelector("[data-docs-page-actions]");
	const copiedLabel = pageActions instanceof HTMLElement ? pageActions.dataset.copiedLabel || "Copied" : "Copied";
	const noContentLabel = pageActions instanceof HTMLElement ? pageActions.dataset.noContentLabel || "No page content was available." : "No page content was available.";
	const markdownUrl = pageActions instanceof HTMLElement ? pageActions.dataset.markdownUrl : undefined;
	const pageActionsToggle = document.querySelector("[data-docs-page-actions-toggle]");
	const pageActionsMenu = document.querySelector("[data-docs-page-actions-menu]");
	const copyPageButton = document.querySelector("[data-docs-copy-page]");

	function setMobileNavExpanded(expanded: boolean) {
		mobileNavToggle?.setAttribute("aria-expanded", String(expanded));
		mobileNavPanel?.setAttribute("data-state", expanded ? "open" : "closed");
	}

	if (mobileNavToggle && mobileNavPanel) {
		const handleMobileNavToggle = () => {
			setMobileNavExpanded(mobileNavToggle.getAttribute("aria-expanded") !== "true");
		};
		const handleMobileNavLinkClick = (event: Event) => {
			if (event.target instanceof Element && event.target.closest("a")) {
				setMobileNavExpanded(false);
			}
		};

		setMobileNavExpanded(false);
		mobileNavToggle.addEventListener("click", handleMobileNavToggle);
		mobileNavPanel.addEventListener("click", handleMobileNavLinkClick);
		cleanupTasks.push(() => {
			mobileNavToggle.removeEventListener("click", handleMobileNavToggle);
			mobileNavPanel.removeEventListener("click", handleMobileNavLinkClick);
		});
	}

	function setPageActionsOpen(open: boolean) {
		pageActionsToggle?.setAttribute("aria-expanded", String(open));
		if (pageActionsMenu instanceof HTMLElement) pageActionsMenu.hidden = !open;
	}

	function flashActionLabel(element: Element | null, label: string) {
		if (!(element instanceof HTMLElement)) return;

		const labelTarget = element.querySelector<HTMLElement>("[data-docs-action-label]") ?? element;
		const original = labelTarget.textContent ?? "";
		labelTarget.textContent = label;
		window.setTimeout(() => {
			labelTarget.textContent = original;
		}, 1400);
	}

	async function copyText(text: string) {
		if (navigator.clipboard?.writeText) {
			try {
				await navigator.clipboard.writeText(text);
				return;
			} catch {
				// Fall back to a temporary textarea when clipboard permissions are unavailable.
			}
		}

		const textarea = document.createElement("textarea");
		textarea.value = text;
		textarea.setAttribute("readonly", "");
		textarea.style.position = "fixed";
		textarea.style.inset = "0 auto auto 0";
		textarea.style.opacity = "0";
		document.body.append(textarea);
		textarea.select();
		document.execCommand("copy");
		textarea.remove();
	}

	function flashCodeCopyLabel(button: HTMLElement, label: string) {
		const labelTarget = button.querySelector<HTMLElement>("[data-ns-code-copy-label]") ?? button;
		const original = labelTarget.textContent ?? "";
		labelTarget.textContent = label;
		window.setTimeout(() => {
			labelTarget.textContent = original;
		}, 1400);
	}

	const codeCopyButtons = Array.from(document.querySelectorAll("[data-ns-code-copy]")).filter((button): button is HTMLButtonElement => button instanceof HTMLButtonElement);

	for (const button of codeCopyButtons) {
		const handleCodeCopy = async () => {
			if (button.disabled) {
				return;
			}

			const block = button.closest("[data-ns-code-block]");
			const code = block?.querySelector("code")?.textContent ?? "";

			if (!code) {
				return;
			}

			await copyText(code);
			flashCodeCopyLabel(button, button.dataset.copiedLabel || "Copied");
		};

		button.addEventListener("click", handleCodeCopy);
		cleanupTasks.push(() => button.removeEventListener("click", handleCodeCopy));
	}

	if (pageActions && pageActionsToggle && pageActionsMenu) {
		const handlePageActionsToggle = () => {
			const willOpen = pageActionsToggle.getAttribute("aria-expanded") !== "true";
			setPageActionsOpen(willOpen);
			if (willOpen) {
				(pageActionsMenu.querySelector("button, a[href]") as HTMLElement | null)?.focus();
			}
		};
		const handleDocumentClick = (event: MouseEvent) => {
			if (event.target instanceof Node && !pageActions.contains(event.target)) {
				setPageActionsOpen(false);
			}
		};
		const handlePageActionsKeydown = (event: KeyboardEvent) => {
			if (pageActionsToggle.getAttribute("aria-expanded") !== "true") return;

			if (event.key === "Escape") {
				event.preventDefault();
				setPageActionsOpen(false);
				(pageActionsToggle as HTMLElement).focus();
			}
		};
		const handleCopyPage = async () => {
			try {
				if (!markdownUrl) throw new Error("Missing documentation Markdown URL.");

				const response = await fetch(markdownUrl, { headers: { Accept: "text/markdown" } });
				if (!response.ok) throw new Error(`Documentation Markdown request failed with ${response.status}.`);

				const markdown = await response.text();
				if (!markdown.trim()) throw new Error("Documentation Markdown response was empty.");

				await copyText(markdown);
				flashActionLabel(pageActionsToggle, copiedLabel);
			} catch {
				flashActionLabel(pageActionsToggle, noContentLabel);
			}
			setPageActionsOpen(false);
		};

		setPageActionsOpen(false);
		pageActionsToggle.addEventListener("click", handlePageActionsToggle);
		document.addEventListener("click", handleDocumentClick);
		document.addEventListener("keydown", handlePageActionsKeydown);
		copyPageButton?.addEventListener("click", handleCopyPage);
		cleanupTasks.push(() => {
			pageActionsToggle.removeEventListener("click", handlePageActionsToggle);
			document.removeEventListener("click", handleDocumentClick);
			document.removeEventListener("keydown", handlePageActionsKeydown);
			copyPageButton?.removeEventListener("click", handleCopyPage);
		});
	}

	const toc = document.querySelector("[data-docs-toc]");
	const tocLinks = Array.from(toc?.querySelectorAll("[data-toc-link]") ?? []);
	const tocSections = tocLinks
		.map(link => {
			const id = link.getAttribute("href")?.slice(1);
			const heading = id ? document.getElementById(decodeURIComponent(id)) : null;

			return heading ? { heading, link } : null;
		})
		.filter((section): section is { heading: HTMLElement; link: Element } => Boolean(section));

	function setActiveTocLink(activeLink: Element) {
		for (const { link } of tocSections) {
			const isActive = link === activeLink;
			link.classList.toggle("active", isActive);
			if (isActive) {
				link.setAttribute("aria-current", "location");
			} else {
				link.removeAttribute("aria-current");
			}
		}
	}

	function updateActiveTocLink() {
		if (!tocSections.length) return;

		const activationOffset = 128;
		let activeSection = tocSections[0];

		for (const section of tocSections) {
			if (section.heading.getBoundingClientRect().top <= activationOffset) {
				activeSection = section;
			} else {
				break;
			}
		}

		setActiveTocLink(activeSection.link);
	}

	let tocUpdateQueued = false;
	function scheduleTocUpdate() {
		if (tocUpdateQueued) return;

		tocUpdateQueued = true;
		requestAnimationFrame(() => {
			tocUpdateQueued = false;
			updateActiveTocLink();
		});
	}

	if (tocSections.length) {
		updateActiveTocLink();

		const scrollOptions = { passive: true };
		window.addEventListener("scroll", scheduleTocUpdate, scrollOptions);
		window.addEventListener("resize", scheduleTocUpdate);
		window.addEventListener("hashchange", scheduleTocUpdate);
		cleanupTasks.push(() => {
			window.removeEventListener("scroll", scheduleTocUpdate, scrollOptions);
			window.removeEventListener("resize", scheduleTocUpdate);
			window.removeEventListener("hashchange", scheduleTocUpdate);
		});
	}

	cleanupDocLayout = () => {
		for (const cleanup of cleanupTasks) {
			cleanup();
		}

		cleanupDocLayout = () => {};
	};
}

initDocLayout();
document.addEventListener("astro:before-swap", prepareDocsSidebarSwap);
document.addEventListener("astro:after-swap", completeDocsSidebarSwap);
document.addEventListener("astro:page-load", initDocLayout);
