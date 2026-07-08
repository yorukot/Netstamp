import type { SessionUser } from "@/features/auth/services/authService";
import { pathForRoute } from "@/routes/routePaths";
import { sidebarItems } from "@/routes/sidebarItems";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { EditorDrawer } from "@/shared/components/EditorDrawer";
import { classNames } from "@/shared/utils/classNames";
import netstampLogo from "@netstamp/brand/assets/netstamp-logo-light.svg";
import netstampMark from "@netstamp/brand/assets/netstamp-mark-light.svg";
import { CaretLeft, CaretRight, List, X } from "@phosphor-icons/react";
import { useState } from "react";
import { Link, NavLink } from "react-router-dom";
import { ProjectSwitcher } from "./ProjectSwitcher";
import styles from "./Sidebar.module.css";
import { UserMenu, UserMenuPanel } from "./UserMenu";

interface SidebarProps {
	collapsed: boolean;
	user: SessionUser;
	onToggleCollapsed: () => void;
	onLogout: () => void;
}

export function Sidebar({ collapsed, user, onToggleCollapsed, onLogout }: SidebarProps) {
	const { projectRef } = useCurrentProject();
	const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
	const MobileMenuIcon = mobileMenuOpen ? X : List;
	const visibleSidebarItems = sidebarItems.filter(item => !item.systemAdminOnly || user.isSystemAdmin);

	function closeMobileMenu() {
		setMobileMenuOpen(false);
	}

	return (
		<aside className={classNames(styles.sidebar, collapsed && styles.collapsed)}>
			<div className={classNames("ns-scrollbar", styles.sidebarScroll)}>
				<div className={styles.brandRow}>
					<Link className={styles.brand} to={pathForRoute("dashboard", { projectRef })} onClick={closeMobileMenu}>
						<img className={classNames(styles.brandLogo, styles.brandLogoFull)} src={netstampLogo} alt="Netstamp" />
						<img className={classNames(styles.brandLogo, styles.brandLogoMark)} src={netstampMark} alt="" aria-hidden="true" />
					</Link>
					<button type="button" className={classNames(styles.brandToggle, styles.brandToggleExpanded)} aria-label="Collapse sidebar" onClick={onToggleCollapsed}>
						<CaretLeft size={17} weight="bold" aria-hidden="true" />
					</button>
					<button type="button" className={classNames(styles.brandToggle, styles.brandToggleCollapsed)} aria-label="Expand sidebar" onClick={onToggleCollapsed}>
						<img className={classNames(styles.brandLogo, styles.brandLogoMark)} src={netstampMark} alt="" aria-hidden="true" />
						<CaretRight className={styles.brandToggleArrow} size={18} weight="bold" aria-hidden="true" />
					</button>
					<button
						type="button"
						className={styles.mobileMenuButton}
						aria-label={mobileMenuOpen ? "Close navigation menu" : "Open navigation menu"}
						aria-expanded={mobileMenuOpen}
						onClick={() => setMobileMenuOpen(open => !open)}
					>
						<MobileMenuIcon size={20} weight="bold" aria-hidden="true" />
					</button>
				</div>

				<ProjectSwitcher collapsed={collapsed} />

				<nav className={styles.nav} aria-label="Primary app navigation">
					{visibleSidebarItems.map(item => {
						const ItemIcon = item.icon;

						return (
							<NavLink
								key={item.route}
								to={pathForRoute(item.route, { projectRef })}
								className={({ isActive }) => classNames(isActive && styles.active)}
								aria-label={item.label}
								title={item.label}
								onClick={closeMobileMenu}
							>
								<ItemIcon className={styles.navIcon} size={18} weight="bold" aria-hidden="true" />
								<span className={styles.navLabel}>{item.label}</span>
							</NavLink>
						);
					})}
				</nav>

				<UserMenu user={user} collapsed={collapsed} onLogout={onLogout} />
			</div>

			<EditorDrawer
				open={mobileMenuOpen}
				title="Menu"
				ariaLabel="Primary navigation menu"
				className={styles.mobileNavDrawer}
				contentClassName={styles.mobileNavDrawerContent}
				onClose={closeMobileMenu}
			>
				<div className={styles.mobileDrawerProject}>
					<ProjectSwitcher variant="drawer" />
				</div>
				<nav className={styles.mobileDrawerNav} aria-label="Primary app navigation">
					{visibleSidebarItems.map(item => {
						const ItemIcon = item.icon;

						return (
							<NavLink
								key={item.route}
								to={pathForRoute(item.route, { projectRef })}
								className={({ isActive }) => classNames(styles.mobileDrawerNavLink, isActive && styles.active)}
								onClick={closeMobileMenu}
							>
								<ItemIcon className={styles.navIcon} size={20} weight="bold" aria-hidden="true" />
								<span>{item.label}</span>
							</NavLink>
						);
					})}
				</nav>
				<div className={styles.mobileDrawerUser}>
					<UserMenuPanel user={user} onLogout={onLogout} onClose={closeMobileMenu} />
				</div>
			</EditorDrawer>
		</aside>
	);
}
