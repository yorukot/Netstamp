import { Drawer, IconButton } from "@netstamp/ui";
import { createElement, useEffect, useRef, useState, type MouseEvent } from "react";
import styles from "./MobileDocsNavigation.module.css";

type PhosphorIconName = "ph-arrow-up-right" | "ph-check" | "ph-github-logo" | "ph-magnifying-glass" | "ph-moon-stars" | "ph-sidebar-simple" | "ph-sun";

interface PhosphorIconProps {
	name: PhosphorIconName;
	className?: string;
}

const PhosphorIcon = ({ name, className }: PhosphorIconProps) =>
	createElement(name, {
		className,
		size: "16",
		weight: "bold",
		"aria-hidden": "true",
		focusable: "false"
	});

interface MobileDocsNavItem {
	title: string;
	href: string;
	active: boolean;
}

interface MobileDocsNavGroup {
	id: string;
	title: string;
	items: MobileDocsNavItem[];
}

interface MobileDocsLanguageOption {
	href: string;
	label: string;
	locale: string;
	htmlLang: string;
	isCurrent: boolean;
}

interface MobileDocsLink {
	href: string;
	label: string;
	external?: boolean;
	active?: boolean;
	icon?: "external" | "github";
}

interface MobileDocsNavigationProps {
	groups: MobileDocsNavGroup[];
	languageOptions: MobileDocsLanguageOption[];
	links: MobileDocsLink[];
	labels: {
		close: string;
		language: string;
		navigation: string;
		open: string;
		search: string;
		switchToDark: string;
		switchToLight: string;
	};
}

export const MobileDocsNavigation = ({ groups, languageOptions, links, labels }: MobileDocsNavigationProps) => {
	const [open, setOpen] = useState(false);
	const activeItemRef = useRef<HTMLAnchorElement>(null);
	const drawerTriggerRef = useRef<HTMLSpanElement>(null);

	useEffect(() => {
		if (!open) return;

		const frame = window.requestAnimationFrame(() => {
			activeItemRef.current?.scrollIntoView({ block: "center" });
		});

		return () => window.cancelAnimationFrame(frame);
	}, [open]);

	const handleDrawerClick = (event: MouseEvent<HTMLElement>) => {
		if (event.target instanceof Element && event.target.closest("a")) {
			setOpen(false);
		}
	};

	return (
		<div className={styles.mobileActions}>
			<IconButton className={styles.topNavButton} variant="outline" aria-label={labels.search} aria-haspopup="dialog" aria-expanded="false" aria-controls="docs-search-dialog" data-search-open>
				<PhosphorIcon name="ph-magnifying-glass" />
			</IconButton>
			<span className={styles.triggerSlot} ref={drawerTriggerRef}>
				<IconButton className={styles.topNavButton} variant="outline" aria-label={labels.open} aria-expanded={open} aria-haspopup="dialog" onClick={() => setOpen(true)}>
					<PhosphorIcon name="ph-sidebar-simple" />
				</IconButton>
			</span>
			<Drawer
				open={open}
				onOpenChange={setOpen}
				title={<span className={styles.drawerTitle}>NETSTAMP / {labels.navigation}</span>}
				closeLabel={labels.close}
				side="right"
				size="full"
				className={styles.drawer}
				contentClassName={styles.drawerContent}
				onCloseAutoFocus={event => {
					event.preventDefault();
					drawerTriggerRef.current?.querySelector("button")?.focus();
				}}
				actions={
					<IconButton
						className={styles.themeToggle}
						variant="outline"
						aria-label={labels.switchToLight}
						data-theme-toggle
						data-theme-light-label={labels.switchToLight}
						data-theme-dark-label={labels.switchToDark}
					>
						<PhosphorIcon name="ph-sun" className={styles.themeIconSun} />
						<PhosphorIcon name="ph-moon-stars" className={styles.themeIconMoon} />
					</IconButton>
				}
			>
				<div className={styles.drawerBody} onClick={handleDrawerClick}>
					<nav className={styles.globalLinks} aria-label={labels.navigation}>
						{links.map(link => (
							<a
								key={link.href}
								className={[styles.globalLink, link.active && styles.active].filter(Boolean).join(" ")}
								href={link.href}
								target={link.external ? "_blank" : undefined}
								rel={link.external ? "noreferrer" : undefined}
								aria-current={link.active ? "page" : undefined}
							>
								<span>{link.label}</span>
								{link.icon === "github" ? <PhosphorIcon name="ph-github-logo" /> : null}
								{link.icon === "external" ? <PhosphorIcon name="ph-arrow-up-right" /> : null}
							</a>
						))}
					</nav>

					{languageOptions.length > 1 ? (
						<section className={styles.languageSection} aria-labelledby="mobile-docs-language-label">
							<h2 id="mobile-docs-language-label" className={styles.sectionTitle}>
								{labels.language}
							</h2>
							<div className={styles.languageOptions}>
								{languageOptions.map(option => (
									<a
										key={option.locale}
										className={[styles.languageOption, option.isCurrent && styles.active].filter(Boolean).join(" ")}
										href={option.href}
										lang={option.htmlLang}
										hrefLang={option.locale}
										aria-current={option.isCurrent ? "page" : undefined}
									>
										<span>{option.label}</span>
										{option.isCurrent ? <PhosphorIcon name="ph-check" /> : null}
									</a>
								))}
							</div>
						</section>
					) : null}

					<nav className={styles.docsTree} aria-label={labels.navigation}>
						{groups.map(group => (
							<section className={styles.docsGroup} aria-labelledby={`mobile-docs-group-${group.id}`} key={group.id}>
								<h2 className={styles.sectionTitle} id={`mobile-docs-group-${group.id}`}>
									{group.title}
								</h2>
								{group.items.map(item => (
									<a
										key={item.href}
										ref={item.active ? activeItemRef : undefined}
										href={item.href}
										className={[styles.docsLink, item.active && styles.active].filter(Boolean).join(" ")}
										aria-current={item.active ? "page" : undefined}
									>
										{item.title}
									</a>
								))}
							</section>
						))}
					</nav>
				</div>
			</Drawer>
		</div>
	);
};
