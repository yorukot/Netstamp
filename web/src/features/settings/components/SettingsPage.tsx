import { useSession } from "@/features/auth/session/SessionContext";
import { useChangeCurrentUserEmailMutation, useChangeCurrentUserPasswordMutation, useUpdateCurrentUserMutation } from "@/shared/api/mutations";
import { ActionRow } from "@/shared/components/ActionRow";
import { BodyCopy } from "@/shared/components/BodyCopy";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { pushToast } from "@/shared/toast/toastStore";
import { Button, Panel, SignalAvatar, TextField } from "@netstamp/ui";
import type { FormEvent } from "react";
import styles from "./SettingsPage.module.css";

function formValue(form: HTMLFormElement, name: string) {
	const value = new FormData(form).get(name);
	return typeof value === "string" ? value.trim() : "";
}

export function SettingsPage() {
	const { session } = useSession();
	const updateUserMutation = useUpdateCurrentUserMutation();
	const changeEmailMutation = useChangeCurrentUserEmailMutation();
	const changePasswordMutation = useChangeCurrentUserPasswordMutation();

	if (!session) {
		return null;
	}

	const { user } = session;

	function handleIdentitySubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		updateUserMutation.mutate({ displayName: formValue(event.currentTarget, "name") });
	}

	function handleEmailSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		changeEmailMutation.mutate({
			newEmail: formValue(event.currentTarget, "new-email"),
			password: formValue(event.currentTarget, "email-password")
		});
	}

	function handlePasswordSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		const newPassword = formValue(event.currentTarget, "new-password");

		if (newPassword !== formValue(event.currentTarget, "confirm-password")) {
			pushToast({ title: "Password mismatch", message: "New passwords do not match.", tone: "critical" });
			return;
		}

		changePasswordMutation.mutate({
			currentPassword: formValue(event.currentTarget, "current-password"),
			newPassword
		});
	}

	return (
		<PageStack>
			<ScreenHeader eyebrow="User settings" title="Account" copy="Set your display name, rotate the login email, and change the password used for controller access." />

			<div className={styles.settingsGrid}>
				<Panel tone="glass" eyebrow="Identity" title="Profile">
					<form id="identity-settings" className={styles.settingsForm} onSubmit={handleIdentitySubmit}>
						<TextField label="Display name" name="name" defaultValue={user.name} />
						<ActionRow>
							<Button type="submit" disabled={updateUserMutation.isPending}>
								{updateUserMutation.isPending ? "Saving" : "Save identity"}
							</Button>
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
					<form className={styles.settingsForm} onSubmit={handleEmailSubmit}>
						<TextField label="Current email" name="current-email" type="email" defaultValue={user.email} />
						<TextField label="New email" name="new-email" type="email" placeholder="operator@example.com" />
						<TextField label="Confirm password" name="email-password" type="password" autoComplete="current-password" />
						<ActionRow>
							<Button type="submit" disabled={changeEmailMutation.isPending}>
								{changeEmailMutation.isPending ? "Updating" : "Update email"}
							</Button>
						</ActionRow>
					</form>
				</Panel>

				<Panel tone="glass" eyebrow="Security" title="Change password">
					<form className={styles.settingsForm} onSubmit={handlePasswordSubmit}>
						<TextField label="Current password" name="current-password" type="password" autoComplete="current-password" />
						<TextField label="New password" name="new-password" type="password" autoComplete="new-password" />
						<TextField label="Confirm new password" name="confirm-password" type="password" autoComplete="new-password" helper="Use at least 12 characters for production accounts." />
						<ActionRow>
							<Button type="submit" disabled={changePasswordMutation.isPending}>
								{changePasswordMutation.isPending ? "Changing" : "Change password"}
							</Button>
						</ActionRow>
					</form>
				</Panel>
			</div>
		</PageStack>
	);
}
