import { useSession } from "@/features/auth/session/SessionContext";
import { demoMode } from "@/shared/config/features";
import { classNames } from "@/shared/utils/classNames";
import { GlobalFooter } from "@netstamp/ui";
import { useEffect, useState } from "react";
import { Outlet } from "react-router-dom";
import styles from "./AppShell.module.css";
import { Sidebar } from "./components/Sidebar";

const sidebarCollapsedStorageKey = "netstamp:sidebar-collapsed";

export function AppShell() {
	const { session, logout } = useSession();
	const [sidebarCollapsed, setSidebarCollapsed] = useState(() => {
		if (typeof window === "undefined") {
			return false;
		}

		try {
			return window.localStorage.getItem(sidebarCollapsedStorageKey) === "true";
		} catch {
			return false;
		}
	});

	useEffect(() => {
		try {
			window.localStorage.setItem(sidebarCollapsedStorageKey, String(sidebarCollapsed));
		} catch {
			// Keep the toggle usable even when persistence is unavailable.
		}
	}, [sidebarCollapsed]);

	if (!session) {
		return null;
	}

	return (
		<div className={classNames(styles.shell, sidebarCollapsed && styles.shellCollapsed)}>
			<a className={styles.skipLink} href="#app-content">
				Skip to content
			</a>
			<Sidebar user={session.user} collapsed={sidebarCollapsed} onToggleCollapsed={() => setSidebarCollapsed(collapsed => !collapsed)} onLogout={logout} />

			<main id="app-content" className={styles.content}>
				<div className={styles.contentMain}>
					<div className={styles.contentBody}>
						{demoMode ? (
							<div className={styles.demoBanner} role="status">
								<strong>Demo mode</strong>
								<span>Read-only access. Mutating requests are blocked; you can only view project data.</span>
							</div>
						) : null}
						<Outlet />
					</div>
				</div>
				<div className={styles.footerSlot}>
					<GlobalFooter />
				</div>
			</main>
		</div>
	);
}
