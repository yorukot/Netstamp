import type { SessionUser } from "@/features/auth/services/authService";
import styles from "@/layouts/AppShell.module.css";
import { pathForRoute } from "@/routes/routePaths";
import { projectQueries } from "@/shared/api/queries";
import { classNames } from "@/shared/utils/classNames";
import { Button, PopoverContent, PopoverPortal, PopoverRoot, PopoverTrigger, SignalAvatar } from "@netstamp/ui";
import { EnvelopeSimple, GearSix, SignOut } from "@phosphor-icons/react";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { Link } from "react-router-dom";

interface UserMenuProps {
	user: SessionUser;
	onLogout: () => void;
}

interface UserMenuPanelProps extends UserMenuProps {
	className?: string;
	onClose?: () => void;
}

export function UserMenu({ user, onLogout }: UserMenuProps) {
	const [profileOpen, setProfileOpen] = useState(false);
	const [avatarOpen, setAvatarOpen] = useState(false);
	const invitesQuery = useQuery(projectQueries.currentUserInvites());
	const pendingInviteCount = invitesQuery.data?.invites.length ?? 0;
	const inviteCountLabel = pendingInviteCount > 99 ? "99+" : String(pendingInviteCount);

	function closeMenus() {
		setProfileOpen(false);
		setAvatarOpen(false);
	}

	function logout() {
		closeMenus();
		onLogout();
	}

	const content = <UserMenuContent user={user} pendingInviteCount={pendingInviteCount} inviteCountLabel={inviteCountLabel} onClose={closeMenus} onLogout={logout} />;

	return (
		<>
			<PopoverRoot open={profileOpen} onOpenChange={setProfileOpen}>
				<PopoverTrigger asChild>
					<button type="button" className={classNames("ns-cut-frame", styles.userCard)} aria-label={`Open user menu for ${user.name}`} title={user.name}>
						<div className={styles.userProfile}>
							<SignalAvatar size="sm" src={user.gravatarUrl} referrerPolicy="no-referrer" aria-hidden="true" />
							<div className={styles.userMeta}>
								<strong>{user.name}</strong>
								<span>{user.role}</span>
							</div>
						</div>
					</button>
				</PopoverTrigger>

				<PopoverPortal>
					<PopoverContent className={classNames("ns-cut-frame", styles.userPopover)} align="start" side="right" sideOffset={10} collisionPadding={8}>
						{content}
					</PopoverContent>
				</PopoverPortal>
			</PopoverRoot>

			<PopoverRoot open={avatarOpen} onOpenChange={setAvatarOpen}>
				<div className={styles.userMenu}>
					<PopoverTrigger asChild>
						<button type="button" className={styles.userAvatarButton} aria-label={`Open user menu for ${user.name}`} title={user.name}>
							<SignalAvatar className={styles.userAvatar} size="sm" src={user.gravatarUrl} referrerPolicy="no-referrer" aria-hidden="true" />
							{pendingInviteCount > 0 ? <span className={styles.inviteBadge}>{inviteCountLabel}</span> : null}
						</button>
					</PopoverTrigger>
				</div>

				<PopoverPortal>
					<PopoverContent className={classNames("ns-cut-frame", styles.userPopover)} align="end" side="right" sideOffset={10} collisionPadding={8}>
						{content}
					</PopoverContent>
				</PopoverPortal>
			</PopoverRoot>
		</>
	);
}

export function UserMenuPanel({ user, onLogout, className, onClose }: UserMenuPanelProps) {
	const invitesQuery = useQuery(projectQueries.currentUserInvites());
	const pendingInviteCount = invitesQuery.data?.invites.length ?? 0;
	const inviteCountLabel = pendingInviteCount > 99 ? "99+" : String(pendingInviteCount);

	function closePanel() {
		onClose?.();
	}

	function logout() {
		closePanel();
		onLogout();
	}

	return (
		<section className={classNames("ns-cut-frame", styles.mobileDrawerUserPanel, className)} aria-label="User menu">
			<UserMenuContent user={user} pendingInviteCount={pendingInviteCount} inviteCountLabel={inviteCountLabel} onClose={closePanel} onLogout={logout} />
		</section>
	);
}

function UserMenuContent({
	user,
	pendingInviteCount,
	inviteCountLabel,
	onClose,
	onLogout
}: {
	user: SessionUser;
	pendingInviteCount: number;
	inviteCountLabel: string;
	onClose: () => void;
	onLogout: () => void;
}) {
	return (
		<>
			<div className={styles.userPopoverProfile}>
				<SignalAvatar size="sm" src={user.gravatarUrl} referrerPolicy="no-referrer" aria-hidden="true" />
				<div className={styles.userMeta}>
					<strong>{user.name}</strong>
					<span>{user.role}</span>
				</div>
			</div>
			<div className={styles.userPopoverActions}>
				<Button className={styles.userPopoverAction} variant="ghost" size="sm" aria-label={`${pendingInviteCount} pending project invites`} title="Project invitations" onClick={onClose} asChild>
					<Link to={pathForRoute("settings")}>
						<EnvelopeSimple size={18} weight="bold" aria-hidden="true" />
						<span>Invites {pendingInviteCount > 0 ? `(${inviteCountLabel})` : ""}</span>
					</Link>
				</Button>
				<Button className={styles.userPopoverAction} variant="ghost" size="sm" onClick={onClose} asChild>
					<Link to={pathForRoute("settings")}>
						<GearSix size={18} weight="bold" aria-hidden="true" />
						<span>Settings</span>
					</Link>
				</Button>
				<Button className={styles.userPopoverAction} variant="ghost" size="sm" onClick={onLogout} asChild>
					<Link to={pathForRoute("landing")}>
						<SignOut size={18} weight="bold" aria-hidden="true" />
						<span>logout</span>
					</Link>
				</Button>
			</div>
		</>
	);
}
