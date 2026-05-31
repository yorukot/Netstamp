import type { SessionUser } from "@/features/auth/services/authService";
import styles from "@/layouts/AppShell.module.css";
import { pathForRoute } from "@/routes/routePaths";
import { sidebarItems } from "@/routes/sidebarItems";
import { classNames } from "@/shared/utils/classNames";
import netstampLogo from "@netstamp/brand/assets/netstamp-logo-light.svg";
import { Link, NavLink } from "react-router-dom";
import { ProjectSwitcher } from "./ProjectSwitcher";
import { UserMenu } from "./UserMenu";

interface SidebarProps {
	user: SessionUser;
	onLogout: () => void;
}

export function Sidebar({ user, onLogout }: SidebarProps) {
	return (
		<aside className={classNames("ns-scrollbar", styles.sidebar)}>
			<Link className={styles.brand} to={pathForRoute("landing")}>
				<img className={styles.brandLogo} src={netstampLogo} alt="Netstamp" />
			</Link>

			<ProjectSwitcher />

			<nav className={styles.nav} aria-label="Primary app navigation">
				{sidebarItems.map(item => (
					<NavLink key={item.route} to={pathForRoute(item.route)} className={({ isActive }) => classNames("ns-cut-frame", isActive && styles.active)}>
						{item.label}
					</NavLink>
				))}
			</nav>

			<UserMenu user={user} onLogout={onLogout} />
		</aside>
	);
}
