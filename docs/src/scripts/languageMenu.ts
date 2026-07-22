let cleanupLanguageMenus = () => {};

const initLanguageMenus = () => {
	cleanupLanguageMenus();

	const cleanupTasks: Array<() => void> = [];
	const languageMenus = document.querySelectorAll<HTMLElement>("[data-language-menu]");

	for (const languageMenu of languageMenus) {
		const trigger = languageMenu.querySelector<HTMLButtonElement>("[data-language-menu-trigger]");
		const menu = languageMenu.querySelector<HTMLElement>("[data-language-menu-content]");
		const items = Array.from(languageMenu.querySelectorAll<HTMLElement>("[data-language-menu-item]"));

		if (!trigger || !menu || items.length === 0) continue;

		let closeTimer: number | undefined;
		const prefersReducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)");
		const isOpen = () => trigger.getAttribute("aria-expanded") === "true";
		const cancelPendingClose = () => {
			if (closeTimer === undefined) return;

			window.clearTimeout(closeTimer);
			closeTimer = undefined;
		};
		const hideMenu = () => {
			menu.hidden = true;
			menu.dataset.state = "closed";
			closeTimer = undefined;
		};
		const openMenu = (focusIndex = 0) => {
			cancelPendingClose();
			trigger.setAttribute("aria-expanded", "true");
			menu.hidden = false;
			menu.dataset.state = "open";
			items[focusIndex]?.focus();
		};
		const closeMenu = (focusTrigger = false) => {
			cancelPendingClose();
			trigger.setAttribute("aria-expanded", "false");
			menu.dataset.state = "closed";

			if (focusTrigger) trigger.focus();

			if (prefersReducedMotion.matches) {
				hideMenu();
				return;
			}

			closeTimer = window.setTimeout(hideMenu, 160);
		};
		const handleTriggerClick = () => {
			if (isOpen()) {
				closeMenu();
				return;
			}

			openMenu();
		};
		const handleTriggerKeydown = (event: KeyboardEvent) => {
			if (event.key !== "ArrowDown" && event.key !== "ArrowUp") return;

			event.preventDefault();
			openMenu(event.key === "ArrowUp" ? items.length - 1 : 0);
		};
		const handleMenuKeydown = (event: KeyboardEvent) => {
			if (event.key === "Escape") {
				event.preventDefault();
				closeMenu(true);
				return;
			}

			if (event.key === "Tab") {
				closeMenu();
				return;
			}

			const activeIndex = items.indexOf(document.activeElement as HTMLElement);
			let nextIndex: number | undefined;

			if (event.key === "ArrowDown") nextIndex = (activeIndex + 1) % items.length;
			if (event.key === "ArrowUp") nextIndex = (activeIndex - 1 + items.length) % items.length;
			if (event.key === "Home") nextIndex = 0;
			if (event.key === "End") nextIndex = items.length - 1;

			if (nextIndex === undefined) return;

			event.preventDefault();
			items[nextIndex]?.focus();
		};
		const handleDocumentPointerDown = (event: PointerEvent) => {
			if (isOpen() && event.target instanceof Node && !languageMenu.contains(event.target)) closeMenu();
		};

		trigger.addEventListener("click", handleTriggerClick);
		trigger.addEventListener("keydown", handleTriggerKeydown);
		menu.addEventListener("keydown", handleMenuKeydown);
		document.addEventListener("pointerdown", handleDocumentPointerDown);
		cleanupTasks.push(() => {
			cancelPendingClose();
			trigger.removeEventListener("click", handleTriggerClick);
			trigger.removeEventListener("keydown", handleTriggerKeydown);
			menu.removeEventListener("keydown", handleMenuKeydown);
			document.removeEventListener("pointerdown", handleDocumentPointerDown);
		});
	}

	cleanupLanguageMenus = () => {
		for (const cleanup of cleanupTasks) cleanup();
		cleanupLanguageMenus = () => {};
	};
};

initLanguageMenus();
document.addEventListener("astro:page-load", initLanguageMenus);
