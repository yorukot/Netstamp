import { useRequireSudo } from "@/features/auth/hooks/useRequireSudo";
import { useSession } from "@/features/auth/session/SessionContext";
import { pathForRoute } from "@/routes/routePaths";
import { absoluteExternalAuthStartUrl, clearCSRFToken } from "@/shared/api/client";
import {
	useAcceptProjectInviteMutation,
	useChangeCurrentUserEmailMutation,
	useChangeCurrentUserPasswordMutation,
	useDeactivateCurrentUserMutation,
	useRejectProjectInviteMutation,
	useRemoveCurrentUserIdentityMutation,
	useRemoveCurrentUserPasswordMutation,
	useRevokeAllAuthSessionsMutation,
	useRevokeAuthSessionMutation,
	useUpdateCurrentUserMutation
} from "@/shared/api/mutations";
import { authQueries, projectQueries } from "@/shared/api/queries";
import { apiQueryKeys } from "@/shared/api/queryKeys";
import type { ApiAuthSession, ApiProjectInvite } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { UnsavedChangesBar } from "@/shared/components/UnsavedChangesBar";
import { useConfirm } from "@/shared/components/confirmContext";
import { appFeatures, demoMode } from "@/shared/config/features";
import { pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import {
	Badge,
	BodyCopy,
	Button,
	DangerAction,
	DataTable,
	DialogContent,
	DialogDescription,
	DialogOverlay,
	DialogPortal,
	DialogRoot,
	DialogTitle,
	Panel,
	SignalAvatar,
	Spinner,
	TextField,
	type DataColumn
} from "@netstamp/ui";
import { EnvelopeSimpleIcon } from "@phosphor-icons/react/dist/csr/EnvelopeSimple";
import { GithubLogoIcon } from "@phosphor-icons/react/dist/csr/GithubLogo";
import { GoogleLogoIcon } from "@phosphor-icons/react/dist/csr/GoogleLogo";
import { IdentificationCardIcon } from "@phosphor-icons/react/dist/csr/IdentificationCard";
import { KeyIcon } from "@phosphor-icons/react/dist/csr/Key";
import { SignOutIcon } from "@phosphor-icons/react/dist/csr/SignOut";
import { UserMinusIcon } from "@phosphor-icons/react/dist/csr/UserMinus";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useId, useRef, useState, type AnimationEvent, type FormEvent } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { APITokensPanel } from "./APITokensPanel";
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
}

interface PasswordFormState {
	userId: string;
	newPassword: string;
	confirmPassword: string;
}

type CredentialDialog = "email" | "password";

const emptyIdentityForm: IdentityFormState = {
	userId: "",
	displayName: ""
};

const emptyEmailForm: EmailFormState = {
	userId: "",
	newEmail: ""
};

const emptyPasswordForm: PasswordFormState = {
	userId: "",
	newPassword: "",
	confirmPassword: ""
};

const passwordReauthReturnTo = "/settings?reauth=set-password";

function formatDateTime(value: string) {
	return new Date(value).toLocaleString();
}

function effectiveSessionExpiry(authSession: ApiAuthSession) {
	const idleExpiry = new Date(authSession.idleExpiresAt).valueOf();
	const absoluteExpiry = new Date(authSession.absoluteExpiresAt).valueOf();
	return new Date(Math.min(idleExpiry, absoluteExpiry)).toISOString();
}

function roleLabel(role: string) {
	return role.charAt(0).toUpperCase() + role.slice(1);
}

function externalAuthenticationErrorMessage(code: string) {
	switch (code) {
		case "identity_conflict":
			return "The external identity does not belong to this Netstamp account.";
		case "sudo_expired":
			return "The verification attempt expired. Please try again.";
		default:
			return "External authentication could not be completed. Please try again.";
	}
}

function AuthenticationProviderIcon({ provider }: { provider: string }) {
	if (provider === "github") {
		return <GithubLogoIcon size="1.5rem" weight="fill" aria-hidden="true" focusable="false" />;
	}

	if (provider === "google") {
		return <GoogleLogoIcon size="1.5rem" weight="bold" aria-hidden="true" focusable="false" />;
	}

	return <IdentificationCardIcon size="1.5rem" weight="bold" aria-hidden="true" focusable="false" />;
}

export function SettingsPage() {
	const { session } = useSession();
	const queryClient = useQueryClient();
	const { setSelectedProjectRef } = useCurrentProject();
	const navigate = useNavigate();
	const [searchParams, setSearchParams] = useSearchParams();
	const confirm = useConfirm();
	const requireSudo = useRequireSudo("/settings");
	const updateUserMutation = useUpdateCurrentUserMutation();
	const changeEmailMutation = useChangeCurrentUserEmailMutation();
	const changePasswordMutation = useChangeCurrentUserPasswordMutation();
	const deactivateUserMutation = useDeactivateCurrentUserMutation();
	const acceptInviteMutation = useAcceptProjectInviteMutation();
	const rejectInviteMutation = useRejectProjectInviteMutation();
	const revokeAllSessionsMutation = useRevokeAllAuthSessionsMutation();
	const revokeSessionMutation = useRevokeAuthSessionMutation();
	const removePasswordMutation = useRemoveCurrentUserPasswordMutation();
	const removeIdentityMutation = useRemoveCurrentUserIdentityMutation();
	const invitesQuery = useQuery(projectQueries.currentUserInvites());
	const sessionsQuery = useQuery({ ...authQueries.sessions(), enabled: Boolean(session) });
	const authenticationMethodsQuery = useQuery({ ...authQueries.authenticationMethods(), enabled: Boolean(session) });
	const authMethodsQuery = useQuery(authQueries.methods());
	const credentialDialogDescriptionId = useId();
	const resumeAction = searchParams.get("reauth");
	const authCallbackError = searchParams.get("auth_error");
	const handledAuthCallbackRef = useRef(false);
	const [identityForm, setIdentityForm] = useState<IdentityFormState>(emptyIdentityForm);
	const [emailForm, setEmailForm] = useState<EmailFormState>(emptyEmailForm);
	const [passwordForm, setPasswordForm] = useState<PasswordFormState>(emptyPasswordForm);
	const [credentialDialog, setCredentialDialog] = useState<CredentialDialog | null>(() => (resumeAction === "set-password" && !authCallbackError ? "password" : null));
	const [isCredentialDialogDismissed, setIsCredentialDialogDismissed] = useState(false);

	useEffect(() => {
		if (handledAuthCallbackRef.current || (resumeAction !== "set-password" && !authCallbackError)) {
			return;
		}
		handledAuthCallbackRef.current = true;

		if (authCallbackError) {
			pushToast({
				title: "Authentication failed",
				message: externalAuthenticationErrorMessage(authCallbackError),
				tone: "critical"
			});
		}

		setSearchParams(
			current => {
				const next = new URLSearchParams(current);
				next.delete("reauth");
				next.delete("auth_error");
				next.delete("auth_provider");
				return next;
			},
			{ replace: true }
		);
	}, [authCallbackError, resumeAction, setSearchParams]);

	if (!session) {
		return null;
	}

	const { user } = session;
	const activeIdentityForm = identityForm.userId === user.id ? identityForm : { userId: user.id, displayName: user.name };
	const activeEmailForm = emailForm.userId === user.id ? emailForm : { ...emptyEmailForm, userId: user.id };
	const activePasswordForm = passwordForm.userId === user.id ? passwordForm : { ...emptyPasswordForm, userId: user.id };
	const hasIdentityChanges = activeIdentityForm.displayName !== user.name;
	const isCredentialDialogOpen = credentialDialog !== null && !isCredentialDialogDismissed;
	const isCredentialMutationPending = changeEmailMutation.isPending || changePasswordMutation.isPending;
	const canChangeEmail = Boolean(activeEmailForm.newEmail.trim() && !changeEmailMutation.isPending);
	const canChangePassword = Boolean(activePasswordForm.newPassword.trim() && activePasswordForm.confirmPassword.trim() && !changePasswordMutation.isPending);
	const hasPassword = authenticationMethodsQuery.data?.hasPassword ?? user.hasPassword;
	const passwordActionLabel = hasPassword ? "Change Password" : "Set Password";

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
				newEmail: activeEmailForm.newEmail.trim()
			},
			{
				onSuccess: () => {
					setIsCredentialDialogDismissed(true);
					pushToast({ title: "Email changed", message: "Your sign-in email and Gravatar source have been updated.", tone: "success" });
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
				newPassword
			},
			{
				onSuccess: () => {
					setIsCredentialDialogDismissed(true);
					pushToast({ title: "Password changed", message: "Your new password is ready for future sign-ins.", tone: "success" });
				}
			}
		);
	}

	function openCredentialDialog(dialog: CredentialDialog) {
		if (dialog === "email") {
			setEmailForm({ ...emptyEmailForm, userId: user.id });
			changeEmailMutation.reset();
		} else {
			setPasswordForm({ ...emptyPasswordForm, userId: user.id });
			changePasswordMutation.reset();
		}

		setCredentialDialog(dialog);
		setIsCredentialDialogDismissed(false);
	}

	function closeCredentialDialog() {
		if (isCredentialMutationPending) {
			return;
		}

		setIsCredentialDialogDismissed(true);
	}

	function finishCredentialDialogClose(event: AnimationEvent<HTMLFormElement>) {
		if (event.target !== event.currentTarget || event.currentTarget.dataset.state !== "closed") {
			return;
		}

		if (credentialDialog === "email") {
			setEmailForm({ ...emptyEmailForm, userId: user.id });
		} else if (credentialDialog === "password") {
			setPasswordForm({ ...emptyPasswordForm, userId: user.id });
		}

		setCredentialDialog(null);
		setIsCredentialDialogDismissed(false);
	}

	function updateEmailForm(field: "newEmail", value: string) {
		setEmailForm(current => ({ ...current, userId: user.id, [field]: value }));
	}

	function updatePasswordForm(field: "newPassword" | "confirmPassword", value: string) {
		setPasswordForm(current => ({ ...current, userId: user.id, [field]: value }));
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
			title: "Deactivate this account?",
			message: "Your account will be disabled and you will not be able to sign in until a system administrator re-enables it.",
			confirmLabel: "Deactivate",
			confirmationText: user.email,
			confirmationLabel: "Account email",
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

	async function revokeSession(authSession: ApiAuthSession) {
		if (demoMode) {
			return;
		}

		const accepted = await confirm({
			title: authSession.isCurrent ? "Sign out this session" : "Revoke session",
			message: authSession.isCurrent ? "This is your current session. You will be signed out on this device." : "This device will lose access immediately. Other active sessions will stay signed in.",
			confirmLabel: authSession.isCurrent ? "Sign out" : "Revoke",
			tone: "danger"
		});
		if (!accepted) {
			return;
		}

		revokeSessionMutation.mutate(authSession.id, {
			onSuccess: () => {
				if (authSession.isCurrent) {
					clearCSRFToken();
					queryClient.removeQueries({ queryKey: apiQueryKeys.auth.all });
					queryClient.removeQueries({ queryKey: apiQueryKeys.projects.all });
					pushToast({ title: "Signed out", message: "The current session was revoked.", tone: "success" });
					navigate(pathForRoute("login"));
					return;
				}
				pushToast({ title: "Session revoked", message: "The selected session can no longer access Netstamp.", tone: "success" });
			},
			onError: error => {
				pushToast({ title: "Revoke failed", message: requestErrorMessage(error, "Could not revoke the selected session."), tone: "critical" });
			}
		});
	}

	async function revokeAllSessions() {
		if (demoMode) {
			return;
		}

		const accepted = await confirm({
			title: "Log out all sessions?",
			message: "Every device signed in to this account, including this one, will lose access immediately.",
			confirmLabel: "Log out all",
			tone: "danger"
		});
		if (!accepted) {
			return;
		}

		revokeAllSessionsMutation.mutate(undefined, {
			onSuccess: () => {
				pushToast({ title: "Logged out everywhere", message: "All sessions for this account have been revoked.", tone: "success" });
			},
			onError: error => {
				pushToast({ title: "Log out failed", message: requestErrorMessage(error, "Could not revoke all sessions."), tone: "critical" });
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
	const activeSessions = sessionsQuery.data?.sessions ?? [];
	const configuredProviders = authMethodsQuery.data?.providers ?? [];
	const connectedProviders = new Set((authenticationMethodsQuery.data?.identities ?? []).map(identity => identity.provider));
	const sessionColumns: DataColumn<ApiAuthSession>[] = [
		{
			key: "client",
			label: "Client",
			render: authSession => (
				<div className={styles.sessionClient}>
					<div className={styles.sessionState}>
						<strong>{authSession.isCurrent ? "Current session" : "Authenticated session"}</strong>
						{authSession.isCurrent ? <Badge tone="success">Current</Badge> : null}
					</div>
					<span className={styles.sessionAgent}>{authSession.userAgent || "User-Agent unavailable for this session"}</span>
				</div>
			)
		},
		{ key: "lastUsedAt", label: "Last active", render: authSession => <span className={styles.sessionTime}>{formatDateTime(authSession.lastUsedAt)}</span> },
		{ key: "createdAt", label: "Signed in", render: authSession => <span className={styles.sessionTime}>{formatDateTime(authSession.createdAt)}</span> },
		{ key: "absoluteExpiresAt", label: "Expires", render: authSession => <span className={styles.sessionTime}>{formatDateTime(effectiveSessionExpiry(authSession))}</span> }
	];

	return (
		<PageStack>
			<ScreenHeader title="Account" />

			<Panel tone="glass" title="Pending invites" padded={false} bodySurface="transparent">
				<DataTable
					columns={inviteColumns}
					rows={inviteRows}
					density="compact"
					minWidth="46rem"
					emptyLabel={invitesQuery.isLoading ? <Spinner label="Loading project invitations" layout="compact" size="lg" /> : "No pending project invitations"}
					getRowKey={row => row.id}
				/>
			</Panel>

			<Panel tone="glass" title="Profile">
				<div className={styles.profileLayout}>
					<div className={styles.profileIdentity}>
						<SignalAvatar size="lg" src={user.gravatarUrl} referrerPolicy="no-referrer" aria-hidden="true" />
						<div className={styles.profileIdentityCopy}>
							<strong>{user.name}</strong>
							<span>{user.email}</span>
						</div>
					</div>

					<form id="identity-settings" className={styles.settingsForm} onSubmit={handleIdentitySubmit}>
						<TextField
							label="Display Name"
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

					<BodyCopy>Your avatar comes from Gravatar and is generated from your account email.</BodyCopy>

					{appFeatures.userCredentialChanges ? (
						<div className={styles.credentialActions}>
							<Button type="button" variant="outline" disabled={demoMode} onClick={() => void requireSudo(() => openCredentialDialog("email"))}>
								<EnvelopeSimpleIcon size="1rem" weight="bold" aria-hidden="true" focusable="false" />
								Change Email
							</Button>
							<Button type="button" variant="outline" disabled={demoMode} onClick={() => void requireSudo(() => openCredentialDialog("password"), { returnTo: passwordReauthReturnTo })}>
								<KeyIcon size="1rem" weight="bold" aria-hidden="true" focusable="false" />
								{passwordActionLabel}
							</Button>
						</div>
					) : null}
				</div>
			</Panel>

			<Panel
				tone="glass"
				title="Active sessions"
				summary="Review the clients authenticated with your account and revoke access you no longer recognize."
				actions={
					<Button type="button" variant="danger" size="sm" disabled={demoMode || revokeAllSessionsMutation.isPending} onClick={() => void revokeAllSessions()}>
						<SignOutIcon aria-hidden="true" focusable="false" />
						{revokeAllSessionsMutation.isPending ? "Logging out" : "Log out all"}
					</Button>
				}
				actionsClassName={styles.sessionActions}
				padded={false}
				bodySurface="transparent"
			>
				<DataTable
					columns={sessionColumns}
					rows={activeSessions}
					density="compact"
					minWidth="64rem"
					ariaLabel="Active auth sessions"
					getRowKey={authSession => authSession.id}
					emptyLabel={
						sessionsQuery.isLoading ? <Spinner label="Loading active sessions" layout="compact" size="lg" /> : sessionsQuery.isError ? "Active sessions could not be loaded" : "No active sessions"
					}
					rowActions={authSession => {
						const revoking = revokeSessionMutation.isPending && revokeSessionMutation.variables === authSession.id;
						return (
							<Button
								type="button"
								variant="danger"
								size="sm"
								disabled={demoMode || revokeSessionMutation.isPending}
								aria-label={authSession.isCurrent ? "Revoke current session" : `Revoke session signed in ${formatDateTime(authSession.createdAt)}`}
								onClick={() => void revokeSession(authSession)}
							>
								<SignOutIcon aria-hidden="true" focusable="false" />
								{revoking ? "Revoking" : authSession.isCurrent ? "Sign out" : "Revoke"}
							</Button>
						);
					}}
				/>
			</Panel>

			<Panel tone="glass" title="Login methods" summary="Manage the credentials and external identities that can access this account." padded={false} bodySurface="transparent">
				<div className={styles.loginMethodList}>
					<article className={styles.loginMethodRow}>
						<div className={styles.loginMethodMain}>
							<div className={styles.loginMethodIcon}>
								<KeyIcon size="1.5rem" weight="bold" aria-hidden="true" focusable="false" />
							</div>
							<div className={styles.loginMethodCopy}>
								<div className={styles.loginMethodHeading}>
									<h3>Password</h3>
									<Badge tone={hasPassword ? "success" : "muted"}>{hasPassword ? "Configured" : "Not configured"}</Badge>
								</div>
								<p>{hasPassword ? "Use your account email and password to sign in." : "Add a password as a fallback sign-in method."}</p>
							</div>
						</div>
						<div className={styles.loginMethodControls}>
							<Button type="button" size="sm" variant="outline" disabled={demoMode} onClick={() => void requireSudo(() => openCredentialDialog("password"), { returnTo: passwordReauthReturnTo })}>
								{hasPassword ? "Change password" : "Set password"}
							</Button>
							{hasPassword && authenticationMethodsQuery.data?.identities.length ? (
								<Button type="button" size="sm" variant="danger" disabled={demoMode || removePasswordMutation.isPending} onClick={() => void requireSudo(() => removePasswordMutation.mutate())}>
									{removePasswordMutation.isPending ? "Removing" : "Remove password"}
								</Button>
							) : null}
						</div>
					</article>
					{authenticationMethodsQuery.data?.identities.map(identity => (
						<article key={identity.id} className={styles.loginMethodRow}>
							<div className={styles.loginMethodMain}>
								<div className={styles.loginMethodIcon}>
									<AuthenticationProviderIcon provider={identity.provider} />
								</div>
								<div className={styles.loginMethodCopy}>
									<div className={styles.loginMethodHeading}>
										<h3>{configuredProviders.find(provider => provider.id === identity.provider)?.displayName || identity.provider}</h3>
										<Badge tone="success">Connected</Badge>
									</div>
									<p className={styles.loginMethodIdentifier}>
										{identity.displayName || "External identity"}
										<span aria-hidden="true"> · </span>
										{identity.username ? `@${identity.username}` : identity.email || identity.issuer}
									</p>
								</div>
							</div>
							<div className={styles.loginMethodControls}>
								<Button
									type="button"
									size="sm"
									variant="danger"
									disabled={demoMode || removeIdentityMutation.isPending}
									aria-label={`Disconnect ${configuredProviders.find(provider => provider.id === identity.provider)?.displayName || identity.provider}`}
									onClick={() => void requireSudo(() => removeIdentityMutation.mutate(identity.id))}
								>
									{removeIdentityMutation.isPending && removeIdentityMutation.variables === identity.id ? "Disconnecting" : "Disconnect"}
								</Button>
							</div>
						</article>
					))}
					{configuredProviders
						.filter(provider => !connectedProviders.has(provider.id))
						.map(provider => (
							<article key={provider.id} className={styles.loginMethodRow}>
								<div className={styles.loginMethodMain}>
									<div className={styles.loginMethodIcon}>
										<AuthenticationProviderIcon provider={provider.id} />
									</div>
									<div className={styles.loginMethodCopy}>
										<div className={styles.loginMethodHeading}>
											<h3>{provider.displayName}</h3>
											<Badge tone="muted">Not connected</Badge>
										</div>
										<p>Connect this provider as another way to sign in.</p>
									</div>
								</div>
								<div className={styles.loginMethodControls}>
									<Button
										type="button"
										size="sm"
										variant="outline"
										disabled={demoMode}
										onClick={() =>
											void requireSudo(() => {
												const url = new URL(absoluteExternalAuthStartUrl(provider.id));
												url.searchParams.set("intent", "link");
												url.searchParams.set("returnTo", "/settings");
												window.location.assign(url.toString());
											})
										}
									>
										Connect {provider.displayName}
									</Button>
								</div>
							</article>
						))}
				</div>
			</Panel>

			<APITokensPanel requireSudo={action => void requireSudo(action)} />

			<Panel tone="deep" title="Dangerous account actions" padded={false} bodySurface="transparent">
				<DangerAction
					title="Deactivate account"
					description="Disable sign-in and protected route access until a system administrator re-enables the account."
					descriptionId="deactivate-account-description"
					action={
						<Button
							type="button"
							variant="danger"
							disabled={demoMode || deactivateUserMutation.isPending}
							aria-describedby="deactivate-account-description"
							onClick={() => void requireSudo(() => void deactivateAccount())}
						>
							<UserMinusIcon size="1rem" weight="bold" aria-hidden="true" focusable="false" />
							{deactivateUserMutation.isPending ? "Deactivating" : "Deactivate account"}
						</Button>
					}
				/>
			</Panel>

			{credentialDialog ? (
				<DialogRoot
					open={isCredentialDialogOpen}
					onOpenChange={open => {
						if (!open) {
							closeCredentialDialog();
						}
					}}
				>
					<DialogPortal>
						<DialogOverlay onMouseDown={closeCredentialDialog}>
							<DialogContent asChild aria-describedby={credentialDialogDescriptionId}>
								<form
									className={styles.credentialDialog}
									onSubmit={credentialDialog === "email" ? handleEmailSubmit : handlePasswordSubmit}
									onMouseDown={event => event.stopPropagation()}
									onAnimationEnd={finishCredentialDialogClose}
								>
									<div className={styles.dialogHeader}>
										<span>Account credentials</span>
										<DialogTitle asChild>
											<strong>{credentialDialog === "email" ? "Change Email" : passwordActionLabel}</strong>
										</DialogTitle>
										<DialogDescription asChild>
											<p id={credentialDialogDescriptionId}>
												{credentialDialog === "email"
													? "Update the email used to sign in and generate your Gravatar."
													: hasPassword
														? "Replace the password used to sign in to your account."
														: "Add a local password so this account can confirm sensitive changes."}
											</p>
										</DialogDescription>
									</div>

									{credentialDialog === "email" ? (
										<>
											<TextField label="Current Email" name="current-email" type="email" value={user.email} disabled />
											<TextField
												label="New Email"
												name="new-email"
												type="email"
												placeholder="operator@example.com"
												value={activeEmailForm.newEmail}
												disabled={changeEmailMutation.isPending}
												onChange={event => updateEmailForm("newEmail", event.currentTarget.value)}
												autoFocus
												required
											/>
										</>
									) : (
										<>
											<TextField
												label="New Password"
												name="new-password"
												type="password"
												autoComplete="new-password"
												value={activePasswordForm.newPassword}
												autoFocus
												disabled={changePasswordMutation.isPending}
												onChange={event => updatePasswordForm("newPassword", event.currentTarget.value)}
												required
											/>
											<TextField
												label="Confirm New Password"
												name="confirm-password"
												type="password"
												autoComplete="new-password"
												helper="Use at least 12 characters for production accounts."
												value={activePasswordForm.confirmPassword}
												disabled={changePasswordMutation.isPending}
												onChange={event => updatePasswordForm("confirmPassword", event.currentTarget.value)}
												required
											/>
										</>
									)}

									<div className={styles.dialogActions}>
										<Button type="button" variant="ghost" disabled={isCredentialMutationPending} onClick={closeCredentialDialog}>
											Cancel
										</Button>
										<Button type="submit" disabled={credentialDialog === "email" ? !canChangeEmail : !canChangePassword}>
											{credentialDialog === "email" ? (changeEmailMutation.isPending ? "Updating" : "Change Email") : changePasswordMutation.isPending ? "Saving" : passwordActionLabel}
										</Button>
									</div>
								</form>
							</DialogContent>
						</DialogOverlay>
					</DialogPortal>
				</DialogRoot>
			) : null}
		</PageStack>
	);
}
