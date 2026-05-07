import { pathForRoute } from "@/routes/routePaths";
import { GlobalFooter } from "@/shared/components/GlobalFooter";
import { classNames } from "@/shared/utils/classNames";
import { currentUser, sidebarItems } from "@/shared/utils/mockData";
import netstampLogo from "@netstamp/brand/assets/netstamp-logo-light.svg";
import { Button, PageShell, Select, SignalAvatar } from "@netstamp/ui";
import { Link, NavLink, Outlet } from "react-router-dom";
import styles from "./AppShell.module.css";

export function AppShell() {
	return (
		<PageShell as="div" className={styles.shell}>
			<aside className={styles.sidebar}>
				<Link className={styles.brand} to={pathForRoute("landing")}>
					<img className={styles.brandLogo} src={netstampLogo} alt="Netstamp" />
				</Link>

				<label className={styles.teamSelect}>
					<span>team</span>
					<Select variant="compact" frameClassName={styles.teamFrame} className={styles.teamControl} defaultValue="vector-ix">
						<option value="vector-ix">Vector IX / prod</option>
						<option value="helio">Helio Validators</option>
						<option value="lab">Lab Network</option>
					</Select>
				</label>

				<nav className={styles.nav} aria-label="Primary app navigation">
					{sidebarItems.map(item => (
						<NavLink key={item.route} to={pathForRoute(item.route)} className={({ isActive }) => classNames("ns-cut-frame", isActive && styles.active)}>
							{item.label}
						</NavLink>
					))}
				</nav>

				<div className={classNames("ns-cut-frame", styles.userCard)}>
					<div className={styles.userProfile}>
						<SignalAvatar size="sm" src={currentUser.gravatarUrl} referrerPolicy="no-referrer" aria-hidden="true" />
						<div className={styles.userMeta}>
							<strong>{currentUser.name}</strong>
							<span>{currentUser.role}</span>
						</div>
					</div>
					<div className={styles.userActions}>
						<Button variant="ghost" size="sm" asChild>
							<Link to={pathForRoute("landing")}>logout</Link>
						</Button>
						<Button variant="ghost" size="sm" asChild>
							<Link to={pathForRoute("settings")}>Settings</Link>
						</Button>
					</div>
				</div>
			</aside>

			<main className={styles.content}>
				<Outlet />
				<GlobalFooter />
			</main>
		</PageShell>
	);
}
