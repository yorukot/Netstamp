import { useSession } from "@/features/auth/session/SessionContext";
import { GlobalFooter } from "@/shared/components/GlobalFooter";
import { PageShell } from "@netstamp/ui";
import { Outlet } from "react-router-dom";
import styles from "./AppShell.module.css";
import { Sidebar } from "./components/Sidebar";

export function AppShell() {
	const { session, logout } = useSession();

	if (!session) {
		return null;
	}

	return (
		<PageShell as="div" className={styles.shell}>
			<Sidebar user={session.user} onLogout={logout} />

			<main className={styles.content}>
				<Outlet />
				<GlobalFooter />
			</main>
		</PageShell>
	);
}
