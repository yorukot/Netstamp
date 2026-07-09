import { useSession } from "@/features/auth/session/SessionContext";
import { useExportAdminDataMutation, useImportAdminDataMutation, useSetManagedUserPasswordMutation, useUpdateAdminSettingsMutation, useUpdateManagedUserMutation } from "@/shared/api/mutations";
import { adminQueries } from "@/shared/api/queries";
import type { ApiAdminDataExport, ApiAdminSettings, ApiManagedUser } from "@/shared/api/types";
import { useConfirm, usePromptDialog } from "@/shared/components/confirmContext";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { ActionRow, Badge, BodyCopy, Button, Checkbox, DataTable, LoadingState, Panel, SelectField, TextField, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import type { ChangeEvent, FormEvent } from "react";
import { useMemo, useRef, useState } from "react";
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

function formatTimestamp(value: string | undefined) {
	if (!value) {
		return "";
	}
	const date = new Date(value);
	if (Number.isNaN(date.valueOf())) {
		return value;
	}
	return new Intl.DateTimeFormat(undefined, {
		dateStyle: "medium",
		timeStyle: "short"
	}).format(date);
}

function isAdminDataExport(value: unknown): value is ApiAdminDataExport {
	if (!value || typeof value !== "object") {
		return false;
	}
	const candidate = value as Partial<ApiAdminDataExport>;
	return typeof candidate.format === "string" && typeof candidate.exportedAt === "string" && Boolean(candidate.tables) && typeof candidate.tables === "object";
}

function downloadAdminDataExport(data: ApiAdminDataExport) {
	const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" });
	const href = URL.createObjectURL(blob);
	const link = document.createElement("a");
	link.href = href;
	link.download = `netstamp-data-export-${new Date().toISOString().slice(0, 10)}.json`;
	document.body.append(link);
	link.click();
	link.remove();
	URL.revokeObjectURL(href);
}

function managedUserSearchText(user: ApiManagedUser) {
	const status = user.disabledAt ? "disabled inactive deactivated" : "active enabled";
	const access = user.isSystemAdmin ? "system admin administrator" : "user member";
	const verification = user.emailVerified ? "verified" : "unverified";

	return [user.displayName, user.email, status, access, verification, formatTimestamp(user.updatedAt), user.disabledAt ? formatTimestamp(user.disabledAt) : ""].join(" ").toLowerCase();
}

function filterManagedUsers(users: ApiManagedUser[], search: string) {
	const needle = search.trim().toLowerCase();

	if (!needle) {
		return users;
	}

	return users.filter(user => managedUserSearchText(user).includes(needle));
}

export function AdminPage() {
	const { session } = useSession();
	const confirm = useConfirm();
	const prompt = usePromptDialog();
	const importInputRef = useRef<HTMLInputElement | null>(null);
	const settingsQuery = useQuery({ ...adminQueries.settings(), enabled: Boolean(session?.user.isSystemAdmin) });
	const usersQuery = useQuery({ ...adminQueries.users(), enabled: Boolean(session?.user.isSystemAdmin) });
	const updateSettingsMutation = useUpdateAdminSettingsMutation();
	const updateManagedUserMutation = useUpdateManagedUserMutation();
	const setManagedUserPasswordMutation = useSetManagedUserPasswordMutation();
	const exportDataMutation = useExportAdminDataMutation();
	const importDataMutation = useImportAdminDataMutation();
	const loadedSettings = settingsQuery.data?.settings;
	const [userSearch, setUserSearch] = useState("");
	const serverForm = useMemo(() => {
		if (!loadedSettings) {
			return defaultForm;
		}
		return formFromSettings(loadedSettings);
	}, [loadedSettings]);
	const [editedForm, setEditedForm] = useState<AdminFormState | null>(null);
	const form = editedForm ?? serverForm;
	const userRows = useMemo(() => usersQuery.data?.users ?? [], [usersQuery.data?.users]);
	const filteredUserRows = useMemo(() => filterManagedUsers(userRows, userSearch), [userRows, userSearch]);
	const userCountLabel = userSearch.trim() ? `${filteredUserRows.length}/${userRows.length} users` : `${userRows.length} users`;
	const activeAdminCount = userRows.filter(user => user.isSystemAdmin && !user.disabledAt).length;

	const smtpPasswordLabel = useMemo(() => {
		if (!settingsQuery.data?.settings.smtp.passwordSet) {
			return "No SMTP password is stored.";
		}
		return "A password is stored. Leave blank to keep it unchanged.";
	}, [settingsQuery.data?.settings.smtp.passwordSet]);

	const userColumns = useMemo<DataColumn<ApiManagedUser>[]>(
		() => [
			{
				key: "user",
				label: "User",
				render: user => (
					<span className={styles.adminCell}>
						<strong className={styles.adminName}>{user.displayName}</strong>
						<span className={styles.adminMeta}>{user.email}</span>
					</span>
				),
				sortable: true,
				sortValue: user => user.email
			},
			{
				key: "status",
				label: "Status",
				render: user => (
					<span className={styles.adminCell}>
						<Badge tone={user.disabledAt ? "critical" : "success"}>{user.disabledAt ? "Disabled" : "Active"}</Badge>
						{user.disabledAt ? <span className={styles.adminMeta}>{formatTimestamp(user.disabledAt)}</span> : null}
					</span>
				),
				sortable: true,
				sortValue: user => (user.disabledAt ? 0 : 1)
			},
			{
				key: "access",
				label: "Access",
				render: user => <Badge tone={user.isSystemAdmin ? "accent" : "neutral"}>{user.isSystemAdmin ? "System admin" : "User"}</Badge>,
				sortable: true,
				sortValue: user => (user.isSystemAdmin ? 1 : 0)
			},
			{
				key: "email",
				label: "Email",
				render: user => <Badge tone={user.emailVerified ? "success" : "warning"}>{user.emailVerified ? "Verified" : "Unverified"}</Badge>,
				sortable: true,
				sortValue: user => (user.emailVerified ? 1 : 0)
			},
			{
				key: "updatedAt",
				label: "Updated",
				render: user => <span className={styles.adminMeta}>{formatTimestamp(user.updatedAt)}</span>,
				sortable: true,
				sortValue: user => user.updatedAt
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
				<ScreenHeader title="System Settings" />
				<Panel tone="deep" title="System administrator access required">
					<BodyCopy>This page is limited to users with global administrator access.</BodyCopy>
				</Panel>
			</PageStack>
		);
	}
	const currentUserID = session.user.id;

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

	async function toggleDisabled(user: ApiManagedUser) {
		const nextDisabled = !user.disabledAt;
		if (nextDisabled) {
			const accepted = await confirm({
				title: "Disable account",
				message: `${user.email} will no longer be able to sign in or access protected routes. A system administrator can re-enable this account later.`,
				confirmLabel: "Disable",
				tone: "danger"
			});
			if (!accepted) {
				return;
			}
		}

		updateManagedUserMutation.mutate(
			{ userId: user.id, body: { disabled: nextDisabled } },
			{
				onSuccess: data => {
					pushToast({
						title: data.user.disabledAt ? "Account disabled" : "Account enabled",
						message: `${data.user.email} was updated.`,
						tone: "success"
					});
				},
				onError: error => {
					pushToast({ title: "User update failed", message: requestErrorMessage(error, "Could not update account status."), tone: "critical" });
				}
			}
		);
	}

	async function toggleSystemAdmin(user: ApiManagedUser) {
		const nextAdmin = !user.isSystemAdmin;
		if (!nextAdmin) {
			const accepted = await confirm({
				title: "Revoke system admin",
				message: `${user.email} will lose access to instance-level system settings. Project memberships are not changed.`,
				confirmLabel: "Revoke",
				tone: "danger"
			});
			if (!accepted) {
				return;
			}
		}

		updateManagedUserMutation.mutate(
			{ userId: user.id, body: { systemAdmin: nextAdmin } },
			{
				onSuccess: data => {
					pushToast({
						title: data.user.isSystemAdmin ? "System admin granted" : "System admin revoked",
						message: `${data.user.email} was updated.`,
						tone: "success"
					});
				},
				onError: error => {
					pushToast({ title: "Permission update failed", message: requestErrorMessage(error, "Could not update user permissions."), tone: "critical" });
				}
			}
		);
	}

	async function setPassword(user: ApiManagedUser) {
		const password = await prompt({
			title: "Set user password",
			message: `Set a new password for ${user.email}.`,
			inputLabel: "New password",
			inputType: "password",
			confirmLabel: "Set password",
			validate: value => (value.length < 8 ? "Password must be at least 8 characters." : null)
		});
		if (!password) {
			return;
		}

		setManagedUserPasswordMutation.mutate(
			{ userId: user.id, body: { password } },
			{
				onSuccess: data => {
					pushToast({ title: "Password updated", message: `${data.user.email} can use the new password.`, tone: "success" });
				},
				onError: error => {
					pushToast({ title: "Password update failed", message: requestErrorMessage(error, "Could not update user password."), tone: "critical" });
				}
			}
		);
	}

	function exportData() {
		exportDataMutation.mutate(undefined, {
			onSuccess: data => {
				downloadAdminDataExport(data);
				pushToast({ title: "Data exported", message: "A JSON backup was downloaded.", tone: "success" });
			},
			onError: error => {
				pushToast({ title: "Export failed", message: requestErrorMessage(error, "Could not export admin data."), tone: "critical" });
			}
		});
	}

	async function importData(event: ChangeEvent<HTMLInputElement>) {
		const file = event.currentTarget.files?.[0];
		event.currentTarget.value = "";
		if (!file) {
			return;
		}

		let parsed: unknown;
		try {
			parsed = JSON.parse(await file.text());
		} catch {
			pushToast({ title: "Import failed", message: "The selected file is not valid JSON.", tone: "critical" });
			return;
		}
		if (!isAdminDataExport(parsed)) {
			pushToast({ title: "Import failed", message: "The selected file is not a Netstamp data export.", tone: "critical" });
			return;
		}

		const accepted = await confirm({
			title: "Import data",
			message: "Existing managed data will be replaced by this backup.",
			confirmLabel: "Import",
			tone: "danger"
		});
		if (!accepted) {
			return;
		}

		importDataMutation.mutate(parsed, {
			onSuccess: data => {
				pushToast({ title: "Data imported", message: `${data.result.importedRows} rows were imported.`, tone: "success" });
			},
			onError: error => {
				pushToast({ title: "Import failed", message: requestErrorMessage(error, "Could not import admin data."), tone: "critical" });
			}
		});
	}

	function userRowActions(user: ApiManagedUser) {
		const isSelf = user.id === currentUserID;
		const lastActiveAdmin = user.isSystemAdmin && !user.disabledAt && activeAdminCount <= 1;
		const updatingUser = updateManagedUserMutation.isPending && updateManagedUserMutation.variables?.userId === user.id;
		const settingPassword = setManagedUserPasswordMutation.isPending && setManagedUserPasswordMutation.variables?.userId === user.id;

		return (
			<div className={styles.userActions}>
				<Button type="button" size="sm" variant={user.disabledAt ? "outline" : "danger"} disabled={isSelf || lastActiveAdmin || updatingUser} onClick={() => void toggleDisabled(user)}>
					{user.disabledAt ? "Enable" : "Disable"}
				</Button>
				<Button type="button" size="sm" variant="ghost" disabled={(isSelf && user.isSystemAdmin) || lastActiveAdmin || updatingUser} onClick={() => void toggleSystemAdmin(user)}>
					{user.isSystemAdmin ? "Revoke admin" : "Grant admin"}
				</Button>
				<Button type="button" size="sm" variant="outline" disabled={settingPassword} onClick={() => void setPassword(user)}>
					{settingPassword ? "Setting" : "Set password"}
				</Button>
			</div>
		);
	}

	return (
		<PageStack>
			<ScreenHeader title="System Settings" actions={settingsQuery.data?.settings.smtp.configured ? <Badge tone="success">SMTP configured</Badge> : <Badge tone="warning">SMTP disabled</Badge>} />

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

			<Panel
				tone="glass"
				title="Data tools"
				actions={
					<div className={styles.toolActions}>
						<Button type="button" variant="outline" disabled={exportDataMutation.isPending} onClick={exportData}>
							{exportDataMutation.isPending ? "Exporting" : "Export data"}
						</Button>
						<Button type="button" variant="danger" disabled={importDataMutation.isPending} onClick={() => importInputRef.current?.click()}>
							{importDataMutation.isPending ? "Importing" : "Import data"}
						</Button>
						<input ref={importInputRef} className={styles.fileInput} type="file" accept="application/json,.json" onChange={event => void importData(event)} />
					</div>
				}
			>
				<BodyCopy>Exports include account, project, probe, check, alert, public status, result, and system setting data. Imported backups replace existing managed data.</BodyCopy>
			</Panel>

			<Panel
				tone="glass"
				title="User management"
				actions={
					<div className={styles.userManagementActions}>
						<TextField label="Search" type="search" placeholder="name, email, status, access" value={userSearch} onChange={event => setUserSearch(event.currentTarget.value)} />
						{usersQuery.isFetching ? <Badge tone="neutral">Syncing</Badge> : <Badge tone="neutral">{userCountLabel}</Badge>}
					</div>
				}
				padded={false}
			>
				{usersQuery.isLoading ? (
					<LoadingState label="Loading users" />
				) : usersQuery.isError ? (
					<div className={styles.tableMessage}>
						<BodyCopy>{requestErrorMessage(usersQuery.error, "Could not load users.")}</BodyCopy>
					</div>
				) : (
					<DataTable<ApiManagedUser>
						ariaLabel="Managed users"
						columns={userColumns}
						rows={filteredUserRows}
						density="compact"
						minWidth="72rem"
						emptyLabel={userSearch.trim() ? "No users match this search" : "No users"}
						getRowKey={user => user.id}
						rowActions={userRowActions}
						rowActionsClassName={styles.userActionsCell}
						rowActionsHeaderClassName={styles.userActionsHeader}
					/>
				)}
			</Panel>
		</PageStack>
	);
}
