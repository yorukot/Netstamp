import { useSession } from "@/features/auth/session/SessionContext";
import { useGrantSystemAdminMutation, useRevokeSystemAdminMutation, useUpdateAdminSettingsMutation } from "@/shared/api/mutations";
import { adminQueries } from "@/shared/api/queries";
import type { ApiAdminSettings, ApiSystemAdminUser } from "@/shared/api/types";
import { useConfirm } from "@/shared/components/confirmContext";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { ActionRow, Badge, BodyCopy, Button, Checkbox, DataTable, LoadingState, Panel, SelectField, TextField, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import type { FormEvent } from "react";
import { useMemo, useState } from "react";
import styles from "./AdminPage.module.css";

interface AdminFormState {
	registrationEnabled: boolean;
	emailVerificationRequired: boolean;
	backendBaseUrl: string;
	publicWebBaseUrl: string;
	smtpHost: string;
	smtpPort: string;
	smtpUsername: string;
	smtpPassword: string;
	smtpClearPassword: boolean;
	smtpFrom: string;
	smtpTLSMode: "starttls" | "implicit" | "none";
	smtpTimeoutSeconds: string;
}

const defaultForm: AdminFormState = {
	registrationEnabled: true,
	emailVerificationRequired: false,
	backendBaseUrl: "",
	publicWebBaseUrl: "",
	smtpHost: "",
	smtpPort: "587",
	smtpUsername: "",
	smtpPassword: "",
	smtpClearPassword: false,
	smtpFrom: "",
	smtpTLSMode: "starttls",
	smtpTimeoutSeconds: "10"
};

function formFromSettings(settings: ApiAdminSettings): AdminFormState {
	return {
		registrationEnabled: settings.registrationEnabled,
		emailVerificationRequired: settings.emailVerificationRequired,
		backendBaseUrl: settings.backendBaseUrl,
		publicWebBaseUrl: settings.publicWebBaseUrl,
		smtpHost: settings.smtp.host,
		smtpPort: String(settings.smtp.port),
		smtpUsername: settings.smtp.username,
		smtpPassword: "",
		smtpClearPassword: false,
		smtpFrom: settings.smtp.from,
		smtpTLSMode: settings.smtp.tlsMode,
		smtpTimeoutSeconds: String(settings.smtp.timeoutSeconds)
	};
}

function numberValue(value: string, fallback: number) {
	const parsed = Number(value);
	return Number.isFinite(parsed) ? parsed : fallback;
}

function formatTimestamp(value: string) {
	const date = new Date(value);
	if (Number.isNaN(date.valueOf())) {
		return value;
	}
	return new Intl.DateTimeFormat(undefined, {
		dateStyle: "medium",
		timeStyle: "short"
	}).format(date);
}

export function AdminPage() {
	const { session } = useSession();
	const confirm = useConfirm();
	const settingsQuery = useQuery({ ...adminQueries.settings(), enabled: Boolean(session?.user.isSystemAdmin) });
	const adminsQuery = useQuery({ ...adminQueries.systemAdmins(), enabled: Boolean(session?.user.isSystemAdmin) });
	const updateSettingsMutation = useUpdateAdminSettingsMutation();
	const grantSystemAdminMutation = useGrantSystemAdminMutation();
	const revokeSystemAdminMutation = useRevokeSystemAdminMutation();
	const loadedSettings = settingsQuery.data?.settings;
	const serverForm = useMemo(() => {
		if (!loadedSettings) {
			return defaultForm;
		}
		return formFromSettings(loadedSettings);
	}, [loadedSettings]);
	const [editedForm, setEditedForm] = useState<AdminFormState | null>(null);
	const [grantEmail, setGrantEmail] = useState("");
	const form = editedForm ?? serverForm;
	const adminRows = adminsQuery.data?.admins ?? [];

	const smtpPasswordLabel = useMemo(() => {
		if (!settingsQuery.data?.settings.smtp.passwordSet) {
			return "No SMTP password is stored.";
		}
		return "A password is stored. Leave blank to keep it unchanged.";
	}, [settingsQuery.data?.settings.smtp.passwordSet]);
	const systemAdminColumns = useMemo<DataColumn<ApiSystemAdminUser>[]>(
		() => [
			{
				key: "user",
				label: "User",
				render: admin => (
					<span className={styles.adminCell}>
						<strong className={styles.adminName}>{admin.displayName}</strong>
						<span className={styles.adminMeta}>{admin.email}</span>
					</span>
				),
				sortable: true,
				sortValue: admin => admin.email
			},
			{
				key: "email",
				label: "Email",
				render: admin => <Badge tone={admin.emailVerified ? "success" : "warning"}>{admin.emailVerified ? "Verified" : "Unverified"}</Badge>,
				sortable: true,
				sortValue: admin => (admin.emailVerified ? 1 : 0)
			},
			{
				key: "grantedAt",
				label: "Granted",
				render: admin => <span className={styles.adminMeta}>{formatTimestamp(admin.grantedAt)}</span>,
				sortable: true,
				sortValue: admin => admin.grantedAt
			}
		],
		[]
	);

	if (!session) {
		return null;
	}

	if (!session.user.isSystemAdmin) {
		return (
			<PageStack>
				<ScreenHeader title="Admin" />
				<Panel tone="deep" title="System administrator access required">
					<BodyCopy>This page is limited to users with global administrator access.</BodyCopy>
				</Panel>
			</PageStack>
		);
	}

	function update<K extends keyof AdminFormState>(key: K, value: AdminFormState[K]) {
		setEditedForm(current => ({ ...(current ?? serverForm), [key]: value }));
	}

	function handleSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		updateSettingsMutation.mutate(
			{
				registrationEnabled: form.registrationEnabled,
				emailVerificationRequired: form.emailVerificationRequired,
				backendBaseUrl: form.backendBaseUrl,
				publicWebBaseUrl: form.publicWebBaseUrl,
				smtp: {
					host: form.smtpHost,
					port: numberValue(form.smtpPort, 587),
					username: form.smtpUsername,
					...(form.smtpPassword ? { password: form.smtpPassword } : {}),
					clearPassword: form.smtpClearPassword,
					from: form.smtpFrom,
					tlsMode: form.smtpTLSMode,
					timeoutSeconds: numberValue(form.smtpTimeoutSeconds, 10)
				}
			},
			{
				onSuccess: () => {
					setEditedForm(null);
					pushToast({ title: "Admin settings saved", message: "System settings were updated.", tone: "success" });
				},
				onError: error => {
					pushToast({ title: "Admin settings failed", message: requestErrorMessage(error, "Could not save admin settings."), tone: "critical" });
				}
			}
		);
	}

	function handleGrantSystemAdmin(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		grantSystemAdminMutation.mutate(
			{ email: grantEmail },
			{
				onSuccess: data => {
					setGrantEmail("");
					pushToast({ title: "System admin granted", message: `${data.admin.email} can now manage instance settings.`, tone: "success" });
				},
				onError: error => {
					pushToast({ title: "Grant failed", message: requestErrorMessage(error, "Could not grant system admin access."), tone: "critical" });
				}
			}
		);
	}

	async function handleRevokeSystemAdmin(admin: ApiSystemAdminUser) {
		const accepted = await confirm({
			title: "Revoke system admin",
			message: `${admin.email} will lose access to instance-level admin settings. Project memberships are not changed.`,
			confirmLabel: "Revoke",
			tone: "danger"
		});
		if (!accepted) {
			return;
		}

		revokeSystemAdminMutation.mutate(admin.id, {
			onSuccess: () => {
				pushToast({ title: "System admin revoked", message: `${admin.email} no longer has instance admin access.`, tone: "success" });
			},
			onError: error => {
				pushToast({ title: "Revoke failed", message: requestErrorMessage(error, "Could not revoke system admin access."), tone: "critical" });
			}
		});
	}

	return (
		<PageStack>
			<ScreenHeader title="Admin" actions={settingsQuery.data?.settings.smtp.configured ? <Badge tone="success">SMTP configured</Badge> : <Badge tone="warning">SMTP disabled</Badge>} />

			<Panel tone="glass" title="System administrators" summary="System admin access manages instance settings and other system admins. It does not grant project membership or project permissions.">
				<form className={styles.grantForm} onSubmit={handleGrantSystemAdmin}>
					<TextField label="User email" type="email" value={grantEmail} placeholder="admin@example.com" onChange={event => setGrantEmail(event.currentTarget.value)} />
					<Button type="submit" disabled={!grantEmail.trim() || grantSystemAdminMutation.isPending}>
						{grantSystemAdminMutation.isPending ? "Granting" : "Grant system admin"}
					</Button>
				</form>

				{adminsQuery.isLoading ? (
					<LoadingState label="Loading system admins" />
				) : adminsQuery.isError ? (
					<BodyCopy>{requestErrorMessage(adminsQuery.error, "Could not load system admins.")}</BodyCopy>
				) : (
					<DataTable<ApiSystemAdminUser>
						ariaLabel="System administrators"
						columns={systemAdminColumns}
						rows={adminRows}
						density="compact"
						minWidth="44rem"
						emptyLabel="No system administrators"
						getRowKey={admin => admin.id}
						rowActions={admin => {
							const isSelf = admin.id === session.user.id;
							const isLastAdmin = adminRows.length <= 1;
							return (
								<Button type="button" size="sm" variant="danger" disabled={isSelf || isLastAdmin || revokeSystemAdminMutation.isPending} onClick={() => handleRevokeSystemAdmin(admin)}>
									Revoke
								</Button>
							);
						}}
					/>
				)}
			</Panel>

			{settingsQuery.isLoading ? (
				<LoadingState label="Loading admin settings" />
			) : settingsQuery.isError ? (
				<Panel tone="deep" title="Admin settings unavailable">
					<BodyCopy>{requestErrorMessage(settingsQuery.error, "Could not load admin settings.")}</BodyCopy>
				</Panel>
			) : (
				<form className={styles.form} onSubmit={handleSubmit}>
					<div className={styles.grid}>
						<Panel tone="glass" title="Instance access">
							<label className={styles.checkboxRow}>
								<Checkbox checked={form.registrationEnabled} onChange={event => update("registrationEnabled", event.currentTarget.checked)} />
								<span>
									<strong>Allow new registrations</strong>
									<small>Disable this after bootstrap if accounts should be invite-only or operator-managed.</small>
								</span>
							</label>
							<label className={styles.checkboxRow}>
								<Checkbox checked={form.emailVerificationRequired} onChange={event => update("emailVerificationRequired", event.currentTarget.checked)} />
								<span>
									<strong>Require email verification</strong>
									<small>New non-admin registrations must confirm email before login. The first bootstrap admin is not blocked.</small>
								</span>
							</label>
						</Panel>

						<Panel tone="glass" title="Public origins">
							<div className={styles.fieldStack}>
								<TextField label="Backend base URL" value={form.backendBaseUrl} placeholder="https://app.netstamp.dev" onChange={event => update("backendBaseUrl", event.currentTarget.value)} />
								<TextField
									label="Public web base URL"
									value={form.publicWebBaseUrl}
									placeholder="https://app.netstamp.dev"
									helper="Password reset and email verification links use this origin when set."
									onChange={event => update("publicWebBaseUrl", event.currentTarget.value)}
								/>
							</div>
						</Panel>
					</div>

					<Panel tone="glass" title="SMTP delivery">
						<div className={styles.smtpGrid}>
							<TextField label="Host" value={form.smtpHost} placeholder="smtp.example.com" onChange={event => update("smtpHost", event.currentTarget.value)} />
							<TextField label="Port" type="number" min={1} max={65535} value={form.smtpPort} onChange={event => update("smtpPort", event.currentTarget.value)} />
							<TextField label="Username" value={form.smtpUsername} autoComplete="off" onChange={event => update("smtpUsername", event.currentTarget.value)} />
							<TextField
								label="Password"
								type="password"
								value={form.smtpPassword}
								autoComplete="new-password"
								helper={smtpPasswordLabel}
								disabled={form.smtpClearPassword}
								onChange={event => update("smtpPassword", event.currentTarget.value)}
							/>
							<TextField label="From" type="email" value={form.smtpFrom} placeholder="alerts@example.com" onChange={event => update("smtpFrom", event.currentTarget.value)} />
							<SelectField
								label="TLS mode"
								value={form.smtpTLSMode}
								options={[
									{ value: "starttls", label: "STARTTLS" },
									{ value: "implicit", label: "Implicit TLS" },
									{ value: "none", label: "None" }
								]}
								onChange={event => update("smtpTLSMode", event.currentTarget.value as AdminFormState["smtpTLSMode"])}
							/>
							<TextField label="Timeout seconds" type="number" min={1} value={form.smtpTimeoutSeconds} onChange={event => update("smtpTimeoutSeconds", event.currentTarget.value)} />
						</div>
						<label className={styles.checkboxRow}>
							<Checkbox
								checked={form.smtpClearPassword}
								onChange={event => {
									update("smtpClearPassword", event.currentTarget.checked);
									if (event.currentTarget.checked) {
										update("smtpPassword", "");
									}
								}}
							/>
							<span>
								<strong>Clear stored SMTP password</strong>
								<small>This also overrides an environment fallback with no password.</small>
							</span>
						</label>
					</Panel>

					<ActionRow>
						<Button type="submit" disabled={updateSettingsMutation.isPending}>
							{updateSettingsMutation.isPending ? "Saving" : "Save admin settings"}
						</Button>
					</ActionRow>
				</form>
			)}
		</PageStack>
	);
}
