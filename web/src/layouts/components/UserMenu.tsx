import type { SessionUser } from "@/features/auth/services/authService";
import { pathForRoute } from "@/routes/routePaths";
import { useTheme } from "@/shared/theme/useTheme";
import { classNames } from "@/shared/utils/classNames";
import { PopoverContent, PopoverPortal, PopoverRoot, PopoverTrigger, SignalAvatar } from "@netstamp/ui";
import { GearSix, MoonStars, SignOut, Sun } from "@phosphor-icons/react";
import { useState } from "react";
import { Link } from "react-router-dom";
import styles from "./UserMenu.module.css";

interface UserMenuProps {
	collapsed?: boolean;
	user: SessionUser;
	onLogout: () => void;
}

interface UserMenuPanelProps extends UserMenuProps {
	className?: string;
	onClose?: () => void;
}

export function UserMenu({ user, collapsed = false, onLogout }: UserMenuProps) {
	const [profileOpen, setProfileOpen] = useState(false);

	function closeMenus() {
		setProfileOpen(false);
	}

	function logout() {
		closeMenus();
		onLogout();
	}

	const content = <UserMenuContent onClose={closeMenus} onLogout={logout} />;

	return (
		<PopoverRoot open={profileOpen} onOpenChange={setProfileOpen}>
			<PopoverTrigger asChild>
				<button type="button" className={classNames(styles.userCard, collapsed && styles.userCardCollapsed)} aria-label={`Open user menu for ${user.name}`} title={user.name}>
					<div className={styles.userProfile}>
						<SignalAvatar className={styles.userAvatar} size="sm" src={user.gravatarUrl} referrerPolicy="no-referrer" aria-hidden="true" />
						<div className={styles.userMeta}>
							<strong>{user.name}</strong>
							<span>{user.role}</span>
						</div>
					</div>
				</button>
			</PopoverTrigger>

			<PopoverPortal>
				<PopoverContent className={styles.userPopover} align={collapsed ? "center" : "start"} side="top" sideOffset={10} collisionPadding={8}>
					{content}
				</PopoverContent>
			</PopoverPortal>
		</PopoverRoot>
	);
}

export function UserMenuPanel({ onLogout, className, onClose }: UserMenuPanelProps) {
	function closePanel() {
		onClose?.();
	}

	function logout() {
		closePanel();
		onLogout();
	}

	return (
		<section className={classNames(styles.mobileDrawerUserPanel, className)} aria-label="User menu">
			<UserMenuContent onClose={closePanel} onLogout={logout} />
		</section>
	);
}

function UserMenuContent({ onClose, onLogout }: { onClose: () => void; onLogout: () => void }) {
	const { theme, toggleTheme } = useTheme();
	const ThemeIcon = theme === "dark" ? Sun : MoonStars;
	const themeToggleLabel = theme === "dark" ? "Switch to light mode" : "Switch to dark mode";

	function switchTheme() {
		toggleTheme();
		onClose();
	}

	return (
		<div className={styles.userPopoverActions}>
			<Link className={styles.userPopoverAction} to={pathForRoute("accountSettings")} onClick={onClose}>
				<GearSix size={18} weight="bold" aria-hidden="true" />
				<span>Settings</span>
			</Link>
			<button type="button" className={styles.userPopoverAction} aria-label={themeToggleLabel} title={themeToggleLabel} aria-pressed={theme === "light"} onClick={switchTheme}>
				<ThemeIcon size={18} weight="bold" aria-hidden="true" />
				<span>{theme === "dark" ? "Light mode" : "Dark mode"}</span>
			</button>
			<Link className={styles.userPopoverAction} to={pathForRoute("login")} onClick={onLogout}>
				<SignOut size={18} weight="bold" aria-hidden="true" />
				<span>Logout</span>
			</Link>
		</div>
	);
}
