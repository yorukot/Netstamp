import type { SessionUser } from "@/features/auth/services/authService";
import { pathForRoute } from "@/routes/routePaths";
import { projectQueries } from "@/shared/api/queries";
import { classNames } from "@/shared/utils/classNames";
import { Button, SignalAvatar } from "@netstamp/ui";
import { EnvelopeSimple } from "@phosphor-icons/react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import styles from "../AppShell.module.css";

interface UserMenuProps {
	user: SessionUser;
	onLogout: () => void;
}

export function UserMenu({ user, onLogout }: UserMenuProps) {
	const invitesQuery = useQuery(projectQueries.currentUserInvites());
	const pendingInviteCount = invitesQuery.data?.invites.length ?? 0;
	const inviteCountLabel = pendingInviteCount > 99 ? "99+" : String(pendingInviteCount);

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
				<span className={styles.inviteInboxSlot}>
					<Button className={styles.inviteInboxButton} variant="ghost" size="sm" aria-label={`${pendingInviteCount} pending project invites`} title="Project invitations" asChild>
						<Link to={pathForRoute("settings")}>
							<EnvelopeSimple size={18} weight="bold" aria-hidden="true" />
						</Link>
					</Button>
					{pendingInviteCount > 0 ? <span className={styles.inviteBadge}>{inviteCountLabel}</span> : null}
				</span>
				<Button className={styles.userTextAction} variant="ghost" size="sm" onClick={onLogout} asChild>
					<Link to={pathForRoute("landing")}>logout</Link>
				</Button>
				<Button className={styles.userTextAction} variant="ghost" size="sm" asChild>
					<Link to={pathForRoute("settings")}>Settings</Link>
				</Button>
			</div>
		</div>
	);
}
