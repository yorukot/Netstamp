const sidebarStorageKey = "netstamp:docs-sidebar-folded";
let cleanupDocLayout = () => {};

function readSidebarState() {
	try {
		return localStorage.getItem(sidebarStorageKey) === "true";
	} catch {
		return false;
	}
}

function writeSidebarState(folded: boolean) {
	try {
		localStorage.setItem(sidebarStorageKey, String(folded));
	} catch {
		// Storage can be unavailable in private contexts; the in-page state still works.
	}
}

function initDocLayout() {
	cleanupDocLayout();

	const sidebarShell = document.querySelector("[data-docs-shell]");
	if (!sidebarShell) return;

	const cleanupTasks: Array<() => void> = [];
	const sidebarToggle = document.querySelector("[data-docs-sidebar-toggle]");
	const sidebarToggleLabel = sidebarToggle?.querySelector("[data-sidebar-toggle-label]");

	function setSidebarFolded(folded: boolean) {
		sidebarShell.classList.toggle("sidebarFolded", folded);
		sidebarToggle?.setAttribute("aria-pressed", String(folded));
		sidebarToggle?.setAttribute("aria-label", folded ? "Expand documentation sidebar" : "Fold documentation sidebar");
		sidebarToggle?.setAttribute("title", folded ? "Expand sidebar" : "Fold sidebar");
		if (sidebarToggleLabel) sidebarToggleLabel.textContent = folded ? "Expand" : "Fold";
		writeSidebarState(folded);
	}

	if (sidebarToggle) {
		const handleSidebarToggle = () => {
			setSidebarFolded(!sidebarShell.classList.contains("sidebarFolded"));
		};

		setSidebarFolded(readSidebarState());
		sidebarToggle.addEventListener("click", handleSidebarToggle);
		cleanupTasks.push(() => sidebarToggle.removeEventListener("click", handleSidebarToggle));
	}

	for (const group of document.querySelectorAll("[data-docs-nav-group]")) {
		const toggle = group.querySelector("[data-docs-nav-group-toggle]");
		if (!toggle) continue;

		const itemsId = toggle.getAttribute("aria-controls");
		const items = itemsId ? document.getElementById(itemsId) : null;
		const handleGroupToggle = () => {
			const expanded = toggle.getAttribute("aria-expanded") === "true";
			const nextExpanded = !expanded;

			toggle.setAttribute("aria-expanded", String(nextExpanded));
			group.classList.toggle("collapsed", !nextExpanded);
			items?.setAttribute("aria-hidden", String(!nextExpanded));
		};

		toggle.addEventListener("click", handleGroupToggle);
		cleanupTasks.push(() => toggle.removeEventListener("click", handleGroupToggle));
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
