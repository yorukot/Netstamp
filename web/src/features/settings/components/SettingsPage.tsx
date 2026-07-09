import { useSession } from "@/features/auth/session/SessionContext";
import { pathForRoute } from "@/routes/routePaths";
import {
	useAcceptProjectInviteMutation,
	useChangeCurrentUserEmailMutation,
	useChangeCurrentUserPasswordMutation,
	useDeactivateCurrentUserMutation,
	useRejectProjectInviteMutation,
	useUpdateCurrentUserMutation
} from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiProjectInvite } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { UnsavedChangesBar } from "@/shared/components/UnsavedChangesBar";
import { useConfirm } from "@/shared/components/confirmContext";
import { appFeatures, demoMode } from "@/shared/config/features";
import { pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { ActionRow, Badge, BodyCopy, Button, DataTable, Panel, SignalAvatar, TextField, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import styles from "./SettingsPage.module.css";

interface InviteRow {
	id: string;
	project: string;
	role: string;
	invitedBy: string;
	createdAt: string;
	source: ApiProjectInvite;
}

interface IdentityFormState {
	userId: string;
	displayName: string;
}

interface EmailFormState {
	userId: string;
	newEmail: string;
	password: string;
}

interface PasswordFormState {
	userId: string;
	currentPassword: string;
	newPassword: string;
	confirmPassword: string;
}

const emptyIdentityForm: IdentityFormState = {
	userId: "",
	displayName: ""
};

const emptyEmailForm: EmailFormState = {
	userId: "",
	newEmail: "",
	password: ""
};

const emptyPasswordForm: PasswordFormState = {
	userId: "",
	currentPassword: "",
	newPassword: "",
	confirmPassword: ""
};

function formatDateTime(value: string) {
	return new Date(value).toLocaleString();
}

function roleLabel(role: string) {
	return role.charAt(0).toUpperCase() + role.slice(1);
}

export function SettingsPage() {
	const { session } = useSession();
	const { setSelectedProjectRef } = useCurrentProject();
	const navigate = useNavigate();
	const confirm = useConfirm();
	const updateUserMutation = useUpdateCurrentUserMutation();
	const changeEmailMutation = useChangeCurrentUserEmailMutation();
	const changePasswordMutation = useChangeCurrentUserPasswordMutation();
	const deactivateUserMutation = useDeactivateCurrentUserMutation();
	const acceptInviteMutation = useAcceptProjectInviteMutation();
	const rejectInviteMutation = useRejectProjectInviteMutation();
	const invitesQuery = useQuery(projectQueries.currentUserInvites());
	const [identityForm, setIdentityForm] = useState<IdentityFormState>(emptyIdentityForm);
	const [emailForm, setEmailForm] = useState<EmailFormState>(emptyEmailForm);
	const [passwordForm, setPasswordForm] = useState<PasswordFormState>(emptyPasswordForm);

	if (!session) {
		return null;
	}

	const { user } = session;
	const activeIdentityForm = identityForm.userId === user.id ? identityForm : { userId: user.id, displayName: user.name };
	const activeEmailForm = emailForm.userId === user.id ? emailForm : { ...emptyEmailForm, userId: user.id };
	const activePasswordForm = passwordForm.userId === user.id ? passwordForm : { ...emptyPasswordForm, userId: user.id };
	const hasIdentityChanges = activeIdentityForm.displayName !== user.name;
	const hasEmailChanges = Boolean(activeEmailForm.newEmail || activeEmailForm.password);
	const hasPasswordChanges = Boolean(activePasswordForm.currentPassword || activePasswordForm.newPassword || activePasswordForm.confirmPassword);

	function handleIdentitySubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		updateUserMutation.mutate(
			{ displayName: activeIdentityForm.displayName.trim() },
			{
				onSuccess: data => {
					setIdentityForm({ userId: data.user.id, displayName: data.user.displayName });
				}
			}
		);
	}

	function handleEmailSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		if (!appFeatures.userCredentialChanges) {
			return;
		}

		changeEmailMutation.mutate(
			{
				newEmail: activeEmailForm.newEmail.trim(),
				password: activeEmailForm.password.trim()
			},
			{
				onSuccess: data => {
					setEmailForm({ ...emptyEmailForm, userId: data.user.id });
				}
			}
		);
	}

	function handlePasswordSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		if (!appFeatures.userCredentialChanges) {
			return;
		}

		const newPassword = activePasswordForm.newPassword.trim();

		if (newPassword !== activePasswordForm.confirmPassword.trim()) {
			pushToast({ title: "Password mismatch", message: "New passwords do not match.", tone: "critical" });
			return;
		}

		changePasswordMutation.mutate(
			{
				currentPassword: activePasswordForm.currentPassword.trim(),
				newPassword
			},
			{
				onSuccess: () => {
					setPasswordForm({ ...emptyPasswordForm, userId: user.id });
				}
			}
		);
	}

	function acceptInvite(invite: ApiProjectInvite) {
		acceptInviteMutation.mutate(invite.id, {
			onSuccess: data => {
				const projectRef = data.invite.project.slug || data.invite.project.id;
				setSelectedProjectRef(projectRef);
				pushToast({
					title: "Invite accepted",
					message: `Switched to ${data.invite.project.name}.`,
					tone: "success"
				});
				navigate(pathForRoute("dashboard", { projectRef }));
			}
		});
	}

	function rejectInvite(invite: ApiProjectInvite) {
		rejectInviteMutation.mutate(invite.id, {
			onSuccess: data => {
				pushToast({
					title: "Invite rejected",
					message: `${data.invite.project.name} was removed from pending invitations.`,
					tone: "success"
				});
			}
		});
	}

	async function deactivateAccount() {
		if (demoMode) {
			return;
		}

		const accepted = await confirm({
			title: "Deactivate account",
			message: "Your account will be disabled and you will not be able to sign in until a system administrator re-enables it.",
			confirmLabel: "Deactivate",
			tone: "danger"
		});
		if (!accepted) {
			return;
		}

		deactivateUserMutation.mutate(undefined, {
			onSuccess: () => {
				pushToast({ title: "Account deactivated", message: "A system administrator can re-enable it later.", tone: "success" });
				navigate(pathForRoute("login"));
			},
			onError: error => {
				pushToast({ title: "Deactivation failed", message: requestErrorMessage(error, "Could not deactivate your account."), tone: "critical" });
			}
		});
	}

	const inviteRows: InviteRow[] = (invitesQuery.data?.invites ?? []).map(invite => ({
		id: invite.id,
		project: invite.project.name,
		role: invite.role,
		invitedBy: invite.invitedByUser.displayName,
		createdAt: formatDateTime(invite.createdAt),
		source: invite
	}));
	const inviteColumns: DataColumn<InviteRow>[] = [
		{
			key: "project",
			label: "Project",
			render: row => <strong>{row.project}</strong>
		},
		{ key: "role", label: "Role", render: row => <Badge tone="accent">{roleLabel(row.role)}</Badge> },
		{ key: "invitedBy", label: "Invited by" },
		{ key: "createdAt", label: "Sent" },
		{
			key: "actions",
			label: "Actions",
			render: row => {
				if (demoMode) {
					return <Badge tone="muted">View only</Badge>;
				}

				const accepting = acceptInviteMutation.isPending && acceptInviteMutation.variables === row.id;
				const rejecting = rejectInviteMutation.isPending && rejectInviteMutation.variables === row.id;

				return (
					<div className={styles.inviteActions}>
						<Button size="sm" disabled={acceptInviteMutation.isPending || rejectInviteMutation.isPending} onClick={() => acceptInvite(row.source)}>
							{accepting ? "Accepting" : "Accept"}
						</Button>
						<Button variant="ghost" size="sm" disabled={acceptInviteMutation.isPending || rejectInviteMutation.isPending} onClick={() => rejectInvite(row.source)}>
							{rejecting ? "Rejecting" : "Reject"}
						</Button>
					</div>
				);
			}
		}
	];

	return (
		<PageStack>
			<ScreenHeader title="Account" />

			<Panel tone="glass" title={`${inviteRows.length} pending project invites`}>
				<DataTable
					columns={inviteColumns}
					rows={inviteRows}
					density="compact"
					minWidth="46rem"
					emptyLabel={invitesQuery.isLoading ? "Loading project invitations" : "No pending project invitations"}
					getRowKey={row => row.id}
				/>
			</Panel>

			<div className={styles.settingsGrid}>
				<Panel tone="glass" title="Profile">
					<form id="identity-settings" className={styles.settingsForm} onSubmit={handleIdentitySubmit}>
						<TextField
							label="Display name"
							name="name"
							value={activeIdentityForm.displayName}
							disabled={demoMode}
							onChange={event => setIdentityForm({ userId: user.id, displayName: event.currentTarget.value })}
						/>
						{demoMode ? (
							<BodyCopy>Profile changes are disabled for demo access.</BodyCopy>
						) : (
							<UnsavedChangesBar show={hasIdentityChanges} saveType="submit" saving={updateUserMutation.isPending} onReset={() => setIdentityForm({ userId: user.id, displayName: user.name })} />
						)}
					</form>
				</Panel>

				<Panel tone="deep" title="Gravatar signal preview">
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

			{appFeatures.userCredentialChanges ? (
				<div className={styles.settingsGrid}>
					<Panel tone="glass" title="Change email">
						<form className={styles.settingsForm} onSubmit={handleEmailSubmit}>
							<TextField label="Current email" name="current-email" type="email" value={user.email} disabled />
							<TextField
								label="New email"
								name="new-email"
								type="email"
								placeholder="operator@example.com"
								value={activeEmailForm.newEmail}
								onChange={event => setEmailForm(current => ({ ...current, userId: user.id, newEmail: event.currentTarget.value }))}
							/>
							<TextField
								label="Confirm password"
								name="email-password"
								type="password"
								autoComplete="current-password"
								value={activeEmailForm.password}
								onChange={event => setEmailForm(current => ({ ...current, userId: user.id, password: event.currentTarget.value }))}
							/>
							<UnsavedChangesBar
								show={hasEmailChanges}
								saveType="submit"
								saving={changeEmailMutation.isPending}
								savingLabel="Updating"
								disabled={!appFeatures.userCredentialChanges}
								onReset={() => setEmailForm({ ...emptyEmailForm, userId: user.id })}
							/>
						</form>
					</Panel>

					<Panel tone="glass" title="Change password">
						<form className={styles.settingsForm} onSubmit={handlePasswordSubmit}>
							<TextField
								label="Current password"
								name="current-password"
								type="password"
								autoComplete="current-password"
								value={activePasswordForm.currentPassword}
								onChange={event => setPasswordForm(current => ({ ...current, userId: user.id, currentPassword: event.currentTarget.value }))}
							/>
							<TextField
								label="New password"
								name="new-password"
								type="password"
								autoComplete="new-password"
								value={activePasswordForm.newPassword}
								onChange={event => setPasswordForm(current => ({ ...current, userId: user.id, newPassword: event.currentTarget.value }))}
							/>
							<TextField
								label="Confirm new password"
								name="confirm-password"
								type="password"
								autoComplete="new-password"
								helper="Use at least 12 characters for production accounts."
								value={activePasswordForm.confirmPassword}
								onChange={event => setPasswordForm(current => ({ ...current, userId: user.id, confirmPassword: event.currentTarget.value }))}
							/>
							<UnsavedChangesBar
								show={hasPasswordChanges}
								saveType="submit"
								saving={changePasswordMutation.isPending}
								savingLabel="Changing"
								disabled={!appFeatures.userCredentialChanges}
								onReset={() => setPasswordForm({ ...emptyPasswordForm, userId: user.id })}
							/>
						</form>
					</Panel>
				</div>
			) : null}

			<Panel tone="deep" title="Deactivate account">
				<BodyCopy>Disabled accounts cannot sign in or access protected routes. A system administrator can re-enable the account.</BodyCopy>
				<ActionRow>
					<Button type="button" variant="danger" disabled={demoMode || deactivateUserMutation.isPending} onClick={() => void deactivateAccount()}>
						{deactivateUserMutation.isPending ? "Deactivating" : "Deactivate account"}
					</Button>
				</ActionRow>
			</Panel>
		</PageStack>
	);
}
