import { useSession } from "@/features/auth/session/SessionContext";
import { pathForRoute } from "@/routes/routePaths";
import {
	useAcceptProjectInviteMutation,
	useChangeCurrentUserEmailMutation,
	useChangeCurrentUserPasswordMutation,
	useRejectProjectInviteMutation,
	useUpdateCurrentUserMutation
} from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiProjectInvite } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { ActionRow } from "@/shared/components/ActionRow";
import { BodyCopy } from "@/shared/components/BodyCopy";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { appFeatures } from "@/shared/config/features";
import { pushToast } from "@/shared/toast/toastStore";
import { Badge, Button, DataTable, Panel, SignalAvatar, TextField, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import type { FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import styles from "./SettingsPage.module.css";

function formValue(form: HTMLFormElement, name: string) {
	const value = new FormData(form).get(name);
	return typeof value === "string" ? value.trim() : "";
}

interface InviteRow {
	id: string;
	project: string;
	role: string;
	invitedBy: string;
	createdAt: string;
	source: ApiProjectInvite;
}

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
	const updateUserMutation = useUpdateCurrentUserMutation();
	const changeEmailMutation = useChangeCurrentUserEmailMutation();
	const changePasswordMutation = useChangeCurrentUserPasswordMutation();
	const acceptInviteMutation = useAcceptProjectInviteMutation();
	const rejectInviteMutation = useRejectProjectInviteMutation();
	const invitesQuery = useQuery(projectQueries.currentUserInvites());

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
		if (!appFeatures.userCredentialChanges) {
			return;
		}

		changeEmailMutation.mutate({
			newEmail: formValue(event.currentTarget, "new-email"),
			password: formValue(event.currentTarget, "email-password")
		});
	}

	function handlePasswordSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		if (!appFeatures.userCredentialChanges) {
			return;
		}

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
						<TextField label="Display name" name="name" defaultValue={user.name} />
						<ActionRow>
							<Button type="submit" disabled={updateUserMutation.isPending}>
								{updateUserMutation.isPending ? "Saving" : "Save identity"}
							</Button>
						</ActionRow>
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

					<Panel tone="glass" title="Change password">
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
			) : null}
		</PageStack>
	);
}
