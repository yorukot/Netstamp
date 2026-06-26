let cleanupDocLayout = () => {};

function initDocLayout() {
	cleanupDocLayout();

	const cleanupTasks: Array<() => void> = [];
	const mobileNavToggle = document.querySelector("[data-docs-mobile-nav-toggle]");
	const mobileNavPanel = document.querySelector("[data-docs-nav-panel]");
	const pageActions = document.querySelector("[data-docs-page-actions]");
	const pageActionsToggle = document.querySelector("[data-docs-page-actions-toggle]");
	const pageActionsMenu = document.querySelector("[data-docs-page-actions-menu]");
	const copyPageButton = document.querySelector("[data-docs-copy-page]");
	const viewMarkdownButton = document.querySelector("[data-docs-view-markdown]");

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

	function pagePlainText() {
		const title = document.querySelector(".docHeader h1")?.textContent?.trim() ?? document.title;
		const summary = document.querySelector(".docHeader p")?.textContent?.trim() ?? "";
		const body = (document.querySelector(".docProse") as HTMLElement | null)?.innerText.trim() ?? "";

		return [`# ${title}`, summary, body].filter(Boolean).join("\n\n");
	}

	function setPageActionsOpen(open: boolean) {
		pageActionsToggle?.setAttribute("aria-expanded", String(open));
		if (pageActionsMenu instanceof HTMLElement) pageActionsMenu.hidden = !open;
	}

	function flashActionLabel(element: Element | null, label: string) {
		if (!(element instanceof HTMLElement)) return;

		const original = element.textContent ?? "";
		element.textContent = label;
		window.setTimeout(() => {
			element.textContent = original;
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

	if (pageActions && pageActionsToggle && pageActionsMenu) {
		const handlePageActionsToggle = () => {
			const willOpen = pageActionsToggle.getAttribute("aria-expanded") !== "true";
			setPageActionsOpen(willOpen);
			if (willOpen) {
				(pageActionsMenu.querySelector("[role='menuitem']") as HTMLElement | null)?.focus();
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
			await copyText(pagePlainText());
			flashActionLabel(copyPageButton, "Copied");
			setPageActionsOpen(false);
		};
		const handleViewMarkdown = () => {
			const blob = new Blob([pagePlainText()], { type: "text/plain;charset=utf-8" });
			const url = URL.createObjectURL(blob);
			window.open(url, "_blank", "noopener,noreferrer");
			window.setTimeout(() => URL.revokeObjectURL(url), 60000);
			setPageActionsOpen(false);
		};

		setPageActionsOpen(false);
		pageActionsToggle.addEventListener("click", handlePageActionsToggle);
		document.addEventListener("click", handleDocumentClick);
		document.addEventListener("keydown", handlePageActionsKeydown);
		copyPageButton?.addEventListener("click", handleCopyPage);
		viewMarkdownButton?.addEventListener("click", handleViewMarkdown);
		cleanupTasks.push(() => {
			pageActionsToggle.removeEventListener("click", handlePageActionsToggle);
			document.removeEventListener("click", handleDocumentClick);
			document.removeEventListener("keydown", handlePageActionsKeydown);
			copyPageButton?.removeEventListener("click", handleCopyPage);
			viewMarkdownButton?.removeEventListener("click", handleViewMarkdown);
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
document.addEventListener("astro:page-load", initDocLayout);
