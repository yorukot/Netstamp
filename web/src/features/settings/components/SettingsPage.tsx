import { useSession } from "@/features/auth/session/SessionContext";
import { ActionRow } from "@/shared/components/ActionRow";
import { BodyCopy } from "@/shared/components/BodyCopy";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { Button, Panel, SignalAvatar, TextField } from "@netstamp/ui";
import type { FormEvent } from "react";
import styles from "./SettingsPage.module.css";

function handleSettingsSubmit(event: FormEvent<HTMLFormElement>) {
	event.preventDefault();
}

export function SettingsPage() {
	const { session } = useSession();

	if (!session) {
		return null;
	}

	const { user } = session;

	return (
		<PageStack>
			<ScreenHeader eyebrow="User settings" title="Account" copy="Set your username, rotate the login email, and change the password used for controller access." />

			<div className={styles.settingsGrid}>
				<Panel tone="glass" eyebrow="Identity" title="Set username">
					<form id="username-settings" className={styles.settingsForm} onSubmit={handleSettingsSubmit}>
						<TextField label="Display name" name="name" defaultValue={user.name} />
						<TextField label="Username" name="username" defaultValue={user.username} helper="Used in audit events and probe ownership trails." />
						<ActionRow>
							<Button type="submit">Save username</Button>
						</ActionRow>
					</form>
				</Panel>

				<Panel tone="deep" eyebrow="Profile image" title="Gravatar signal preview">
					<div className={styles.profilePreview}>
						<SignalAvatar size="lg" src={user.gravatarUrl} referrerPolicy="no-referrer" aria-hidden="true" />
						<div>
							<h3>{user.name}</h3>
							<p>{user.email}</p>
						</div>
					</div>
					<BodyCopy>The avatar is pulled using your email from Gravatar.</BodyCopy>
				</Panel>
			</div>

			<div className={styles.settingsGrid}>
				<Panel tone="glass" eyebrow="Email" title="Change email">
					<form className={styles.settingsForm} onSubmit={handleSettingsSubmit}>
						<TextField label="Current email" name="current-email" type="email" defaultValue={user.email} />
						<TextField label="New email" name="new-email" type="email" placeholder="operator@example.com" />
						<TextField label="Confirm password" name="email-password" type="password" autoComplete="current-password" />
						<ActionRow>
							<Button type="submit">Update email</Button>
						</ActionRow>
					</form>
				</Panel>

				<Panel tone="glass" eyebrow="Security" title="Change password">
					<form className={styles.settingsForm} onSubmit={handleSettingsSubmit}>
						<TextField label="Current password" name="current-password" type="password" autoComplete="current-password" />
						<TextField label="New password" name="new-password" type="password" autoComplete="new-password" />
						<TextField label="Confirm new password" name="confirm-password" type="password" autoComplete="new-password" helper="Use at least 12 characters for production accounts." />
						<ActionRow>
							<Button type="submit">Change password</Button>
						</ActionRow>
					</form>
				</Panel>
			</div>
		</PageStack>
	);
}
