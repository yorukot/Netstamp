let cleanupDocLayout = () => {};

function initDocLayout() {
	cleanupDocLayout();

	const cleanupTasks: Array<() => void> = [];
	const mobileNavToggle = document.querySelector("[data-docs-mobile-nav-toggle]");
	const mobileNavPanel = document.querySelector("[data-docs-nav-panel]");

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
