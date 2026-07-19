import { useRequireSudo } from "@/features/auth/hooks/useRequireSudo";
import { useSession } from "@/features/auth/session/SessionContext";
import { useLocaleFormat } from "@/i18n/format";
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
import type { ApiAuthSession, ApiProjectInvite, ProjectMemberRole } from "@/shared/api/types";
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
import { useTranslation } from "react-i18next";
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

function effectiveSessionExpiry(authSession: ApiAuthSession) {
	const idleExpiry = new Date(authSession.idleExpiresAt).valueOf();
	const absoluteExpiry = new Date(authSession.absoluteExpiresAt).valueOf();
	return new Date(Math.min(idleExpiry, absoluteExpiry)).toISOString();
}

function externalAuthenticationErrorKey(code: string) {
	switch (code) {
		case "identity_conflict":
			return "identityConflict" as const;
		case "sudo_expired":
			return "sudoExpired" as const;
		default:
			return "default" as const;
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
	const { t } = useTranslation(["settings", "common", "project"]);
	const format = useLocaleFormat();
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
				title: t("account.authenticationFailed"),
				message: t(`account.authErrors.${externalAuthenticationErrorKey(authCallbackError)}`),
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
	}, [authCallbackError, resumeAction, setSearchParams, t]);

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
	const passwordActionLabel = hasPassword ? t("account.changePassword") : t("account.setPassword");

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
					pushToast({ title: t("account.emailChanged"), message: t("account.emailChangedMessage"), tone: "success" });
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
			pushToast({ title: t("account.passwordMismatch"), message: t("account.passwordMismatchMessage"), tone: "critical" });
			return;
		}

		changePasswordMutation.mutate(
			{
				newPassword
			},
			{
				onSuccess: () => {
					setIsCredentialDialogDismissed(true);
					pushToast({ title: t("account.passwordChanged"), message: t("account.passwordChangedMessage"), tone: "success" });
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
					title: t("account.invites.accepted"),
					message: t("account.invites.acceptedMessage", { project: data.invite.project.name }),
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
					title: t("account.invites.rejected"),
					message: t("account.invites.rejectedMessage", { project: data.invite.project.name }),
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
			title: t("account.danger.question"),
			message: t("account.danger.message"),
			confirmLabel: t("account.danger.deactivate"),
			confirmationText: user.email,
			confirmationLabel: t("account.danger.confirmationLabel"),
			tone: "danger"
		});
		if (!accepted) {
			return;
		}

		deactivateUserMutation.mutate(undefined, {
			onSuccess: () => {
				pushToast({ title: t("account.danger.deactivated"), message: t("account.danger.deactivatedMessage"), tone: "success" });
				navigate(pathForRoute("login"));
			},
			onError: error => {
				pushToast({ title: t("account.danger.failed"), message: requestErrorMessage(error, t("account.danger.error")), tone: "critical" });
			}
		});
	}

	async function revokeSession(authSession: ApiAuthSession) {
		if (demoMode) {
			return;
		}

		const accepted = await confirm({
			title: authSession.isCurrent ? t("account.sessions.signOutQuestion") : t("account.sessions.revokeQuestion"),
			message: authSession.isCurrent ? t("account.sessions.currentMessage") : t("account.sessions.revokeMessage"),
			confirmLabel: authSession.isCurrent ? t("account.sessions.signOut") : t("account.sessions.revoke"),
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
					pushToast({ title: t("account.sessions.signedOut"), message: t("account.sessions.signedOutMessage"), tone: "success" });
					navigate(pathForRoute("login"));
					return;
				}
				pushToast({ title: t("account.sessions.revoked"), message: t("account.sessions.revokedMessage"), tone: "success" });
			},
			onError: error => {
				pushToast({ title: t("account.sessions.revokeFailed"), message: requestErrorMessage(error, t("account.sessions.revokeError")), tone: "critical" });
			}
		});
	}

	async function revokeAllSessions() {
		if (demoMode) {
			return;
		}

		const accepted = await confirm({
			title: t("account.sessions.logOutAllQuestion"),
			message: t("account.sessions.logOutAllMessage"),
			confirmLabel: t("account.sessions.logOutAll"),
			tone: "danger"
		});
		if (!accepted) {
			return;
		}

		revokeAllSessionsMutation.mutate(undefined, {
			onSuccess: () => {
				pushToast({ title: t("account.sessions.loggedOutEverywhere"), message: t("account.sessions.loggedOutEverywhereMessage"), tone: "success" });
			},
			onError: error => {
				pushToast({ title: t("account.sessions.logOutFailed"), message: requestErrorMessage(error, t("account.sessions.logOutError")), tone: "critical" });
			}
		});
	}

	const inviteRows: InviteRow[] = (invitesQuery.data?.invites ?? []).map(invite => ({
		id: invite.id,
		project: invite.project.name,
		role: invite.role,
		invitedBy: invite.invitedByUser.displayName,
		createdAt: format.dateTime(invite.createdAt),
		source: invite
	}));
	const inviteColumns: DataColumn<InviteRow>[] = [
		{
			key: "project",
			label: t("account.invites.project"),
			render: row => <strong>{row.project}</strong>
		},
		{ key: "role", label: t("account.invites.role"), render: row => <Badge tone="accent">{t(`project:roles.${row.role as ProjectMemberRole}`)}</Badge> },
		{ key: "invitedBy", label: t("account.invites.invitedBy") },
		{ key: "createdAt", label: t("account.invites.sent") },
		{
			key: "actions",
			label: t("account.invites.actions"),
			render: row => {
				if (demoMode) {
					return <Badge tone="muted">{t("account.invites.viewOnly")}</Badge>;
				}

				const accepting = acceptInviteMutation.isPending && acceptInviteMutation.variables === row.id;
				const rejecting = rejectInviteMutation.isPending && rejectInviteMutation.variables === row.id;

				return (
					<div className={styles.inviteActions}>
						<Button size="sm" disabled={acceptInviteMutation.isPending || rejectInviteMutation.isPending} onClick={() => acceptInvite(row.source)}>
							{accepting ? t("account.invites.accepting") : t("account.invites.accept")}
						</Button>
						<Button variant="ghost" size="sm" disabled={acceptInviteMutation.isPending || rejectInviteMutation.isPending} onClick={() => rejectInvite(row.source)}>
							{rejecting ? t("account.invites.rejecting") : t("account.invites.reject")}
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
			label: t("account.sessions.client"),
			render: authSession => (
				<div className={styles.sessionClient}>
					<div className={styles.sessionState}>
						<strong>{authSession.isCurrent ? t("account.sessions.currentSession") : t("account.sessions.authenticatedSession")}</strong>
						{authSession.isCurrent ? <Badge tone="success">{t("account.sessions.current")}</Badge> : null}
					</div>
					<span className={styles.sessionAgent}>{authSession.userAgent || t("account.sessions.userAgentUnavailable")}</span>
				</div>
			)
		},
		{ key: "lastUsedAt", label: t("account.sessions.lastActive"), render: authSession => <span className={styles.sessionTime}>{format.dateTime(authSession.lastUsedAt)}</span> },
		{ key: "createdAt", label: t("account.sessions.signedIn"), render: authSession => <span className={styles.sessionTime}>{format.dateTime(authSession.createdAt)}</span> },
		{ key: "absoluteExpiresAt", label: t("account.sessions.expires"), render: authSession => <span className={styles.sessionTime}>{format.dateTime(effectiveSessionExpiry(authSession))}</span> }
	];

	return (
		<PageStack>
			<ScreenHeader title={t("account.title")} />

			<Panel tone="glass" title={t("account.invites.title")} padded={false} bodySurface="transparent">
				<DataTable
					columns={inviteColumns}
					rows={inviteRows}
					density="compact"
					minWidth="46rem"
					emptyLabel={invitesQuery.isLoading ? <Spinner label={t("account.invites.loading")} layout="compact" size="lg" /> : t("account.invites.empty")}
					getRowKey={row => row.id}
				/>
			</Panel>

			<Panel tone="glass" title={t("account.profile.title")}>
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
							label={t("account.profile.displayName")}
							name="name"
							value={activeIdentityForm.displayName}
							disabled={demoMode}
							onChange={event => setIdentityForm({ userId: user.id, displayName: event.currentTarget.value })}
						/>
						{demoMode ? (
							<BodyCopy>{t("account.profile.demoDisabled")}</BodyCopy>
						) : (
							<UnsavedChangesBar show={hasIdentityChanges} saveType="submit" saving={updateUserMutation.isPending} onReset={() => setIdentityForm({ userId: user.id, displayName: user.name })} />
						)}
					</form>

					<BodyCopy>{t("account.profile.avatarDescription")}</BodyCopy>

					{appFeatures.userCredentialChanges ? (
						<div className={styles.credentialActions}>
							<Button type="button" variant="outline" disabled={demoMode} onClick={() => void requireSudo(() => openCredentialDialog("email"))}>
								<EnvelopeSimpleIcon size="1rem" weight="bold" aria-hidden="true" focusable="false" />
								{t("account.changeEmail")}
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
				title={t("account.sessions.title")}
				summary={t("account.sessions.summary")}
				actions={
					<Button type="button" variant="danger" size="sm" disabled={demoMode || revokeAllSessionsMutation.isPending} onClick={() => void revokeAllSessions()}>
						<SignOutIcon aria-hidden="true" focusable="false" />
						{revokeAllSessionsMutation.isPending ? t("account.sessions.loggingOut") : t("account.sessions.logOutAll")}
					</Button>
				}
				padded={false}
				bodySurface="transparent"
			>
				<DataTable
					columns={sessionColumns}
					rows={activeSessions}
					density="compact"
					minWidth="64rem"
					ariaLabel={t("account.sessions.aria")}
					getRowKey={authSession => authSession.id}
					emptyLabel={
						sessionsQuery.isLoading ? (
							<Spinner label={t("account.sessions.loading")} layout="compact" size="lg" />
						) : sessionsQuery.isError ? (
							t("account.sessions.loadError")
						) : (
							t("account.sessions.empty")
						)
					}
					rowActions={authSession => {
						const revoking = revokeSessionMutation.isPending && revokeSessionMutation.variables === authSession.id;
						return (
							<Button
								type="button"
								variant="danger"
								size="sm"
								disabled={demoMode || revokeSessionMutation.isPending}
								aria-label={authSession.isCurrent ? t("account.sessions.revokeCurrentAria") : t("account.sessions.revokeAria", { date: format.dateTime(authSession.createdAt) })}
								onClick={() => void revokeSession(authSession)}
							>
								<SignOutIcon aria-hidden="true" focusable="false" />
								{revoking ? t("account.sessions.revoking") : authSession.isCurrent ? t("account.sessions.signOut") : t("account.sessions.revoke")}
							</Button>
						);
					}}
				/>
			</Panel>

			<Panel tone="glass" title={t("account.loginMethods.title")} summary={t("account.loginMethods.summary")} padded={false} bodySurface="transparent">
				<div className={styles.loginMethodList}>
					<article className={styles.loginMethodRow}>
						<div className={styles.loginMethodMain}>
							<div className={styles.loginMethodIcon}>
								<KeyIcon size="1.5rem" weight="bold" aria-hidden="true" focusable="false" />
							</div>
							<div className={styles.loginMethodCopy}>
								<div className={styles.loginMethodHeading}>
									<h3>{t("account.loginMethods.password")}</h3>
									<Badge tone={hasPassword ? "success" : "muted"}>{hasPassword ? t("account.loginMethods.configured") : t("account.loginMethods.notConfigured")}</Badge>
								</div>
								<p>{hasPassword ? t("account.loginMethods.passwordDescription") : t("account.loginMethods.passwordFallback")}</p>
							</div>
						</div>
						<div className={styles.loginMethodControls}>
							<Button type="button" size="sm" variant="outline" disabled={demoMode} onClick={() => void requireSudo(() => openCredentialDialog("password"), { returnTo: passwordReauthReturnTo })}>
								{hasPassword ? t("account.changePassword") : t("account.setPassword")}
							</Button>
							{hasPassword && authenticationMethodsQuery.data?.identities.length ? (
								<Button type="button" size="sm" variant="danger" disabled={demoMode || removePasswordMutation.isPending} onClick={() => void requireSudo(() => removePasswordMutation.mutate())}>
									{removePasswordMutation.isPending ? t("account.loginMethods.removing") : t("account.loginMethods.removePassword")}
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
										<Badge tone="success">{t("account.loginMethods.connected")}</Badge>
									</div>
									<p className={styles.loginMethodIdentifier}>
										{identity.displayName || t("account.loginMethods.externalIdentity")}
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
									aria-label={t("account.loginMethods.disconnectAria", { provider: configuredProviders.find(provider => provider.id === identity.provider)?.displayName || identity.provider })}
									onClick={() => void requireSudo(() => removeIdentityMutation.mutate(identity.id))}
								>
									{removeIdentityMutation.isPending && removeIdentityMutation.variables === identity.id ? t("account.loginMethods.disconnecting") : t("account.loginMethods.disconnect")}
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
											<Badge tone="muted">{t("account.loginMethods.notConnected")}</Badge>
										</div>
										<p>{t("account.loginMethods.connectDescription")}</p>
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
										{t("account.loginMethods.connect", { provider: provider.displayName })}
									</Button>
								</div>
							</article>
						))}
				</div>
			</Panel>

			<APITokensPanel requireSudo={action => void requireSudo(action)} />

			<Panel tone="deep" title={t("account.danger.title")} padded={false} bodySurface="transparent">
				<DangerAction
					title={t("account.danger.deactivate")}
					description={t("account.danger.description")}
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
							{deactivateUserMutation.isPending ? t("account.danger.deactivating") : t("account.danger.deactivate")}
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
										<span>{t("account.credentialDialog.eyebrow")}</span>
										<DialogTitle asChild>
											<strong>{credentialDialog === "email" ? t("account.changeEmail") : passwordActionLabel}</strong>
										</DialogTitle>
										<DialogDescription asChild>
											<p id={credentialDialogDescriptionId}>
												{credentialDialog === "email"
													? t("account.credentialDialog.emailDescription")
													: hasPassword
														? t("account.credentialDialog.changePasswordDescription")
														: t("account.credentialDialog.setPasswordDescription")}
											</p>
										</DialogDescription>
									</div>

									{credentialDialog === "email" ? (
										<>
											<TextField label={t("account.credentialDialog.currentEmail")} name="current-email" type="email" value={user.email} disabled />
											<TextField
												label={t("account.credentialDialog.newEmail")}
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
												label={t("account.credentialDialog.newPassword")}
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
												label={t("account.credentialDialog.confirmPassword")}
												name="confirm-password"
												type="password"
												autoComplete="new-password"
												helper={t("account.credentialDialog.passwordHelper")}
												value={activePasswordForm.confirmPassword}
												disabled={changePasswordMutation.isPending}
												onChange={event => updatePasswordForm("confirmPassword", event.currentTarget.value)}
												required
											/>
										</>
									)}

									<div className={styles.dialogActions}>
										<Button type="button" variant="ghost" disabled={isCredentialMutationPending} onClick={closeCredentialDialog}>
											{t("common:actions.cancel")}
										</Button>
										<Button type="submit" disabled={credentialDialog === "email" ? !canChangeEmail : !canChangePassword}>
											{credentialDialog === "email"
												? changeEmailMutation.isPending
													? t("account.credentialDialog.updating")
													: t("account.changeEmail")
												: changePasswordMutation.isPending
													? t("common:actions.saving")
													: passwordActionLabel}
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
