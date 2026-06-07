import { useSession } from "@/features/auth/session/SessionContext";
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
			<Sidebar user={session.user} collapsed={sidebarCollapsed} onToggleCollapsed={() => setSidebarCollapsed(collapsed => !collapsed)} onLogout={logout} />

			<main className={styles.content}>
				<Outlet />
				<GlobalFooter />
			</main>
		</div>
	);
}
