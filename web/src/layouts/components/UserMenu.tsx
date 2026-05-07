import type { MockUser } from "@/features/auth/services/authService";
import { pathForRoute } from "@/routes/routePaths";
import { classNames } from "@/shared/utils/classNames";
import { Button, SignalAvatar } from "@netstamp/ui";
import { Link } from "react-router-dom";
import styles from "../AppShell.module.css";

interface UserMenuProps {
	user: MockUser;
	onLogout: () => void;
}

export function UserMenu({ user, onLogout }: UserMenuProps) {
	return (
		<div className={classNames("ns-cut-frame", styles.userCard)}>
			<div className={styles.userProfile}>
				<SignalAvatar size="sm" src={user.gravatarUrl} referrerPolicy="no-referrer" aria-hidden="true" />
				<div className={styles.userMeta}>
					<strong>{user.name}</strong>
					<span>{user.role}</span>
				</div>
			</div>
			<div className={styles.userActions}>
				<Button variant="ghost" size="sm" onClick={onLogout} asChild>
					<Link to={pathForRoute("landing")}>logout</Link>
				</Button>
				<Button variant="ghost" size="sm" asChild>
					<Link to={pathForRoute("settings")}>Settings</Link>
				</Button>
			</div>
		</div>
	);
}
