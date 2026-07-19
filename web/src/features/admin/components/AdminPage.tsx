import { useRequireSudo } from "@/features/auth/hooks/useRequireSudo";
import { useSession } from "@/features/auth/session/SessionContext";
import { formatDateTime } from "@/i18n/format";
import {
	useClearManagedUserPasswordMutation,
	useExportAdminDataMutation,
	useImportAdminDataMutation,
	useSetManagedUserPasswordMutation,
	useUpdateAdminSettingsMutation,
	useUpdateManagedUserMutation
} from "@/shared/api/mutations";
import { adminQueries } from "@/shared/api/queries";
import type { ApiAdminDataExport, ApiAdminSettings, ApiManagedUser } from "@/shared/api/types";
import { useConfirm, usePromptDialog } from "@/shared/components/confirmContext";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { UnsavedChangesBar } from "@/shared/components/UnsavedChangesBar";
import { pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Badge, BodyCopy, Button, Checkbox, DataTable, Panel, SelectField, Spinner, TextField, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import type { ChangeEvent, FormEvent } from "react";
import { useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
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
	return formatDateTime(date, {
		dateStyle: "medium",
		timeStyle: "short"
	});
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

function managedUserSearchText(user: ApiManagedUser, labels: string[]) {
	const [status, access, verification] = labels;

	return [user.displayName, user.email, status, access, verification, formatTimestamp(user.updatedAt), user.disabledAt ? formatTimestamp(user.disabledAt) : ""].join(" ").toLowerCase();
}

function filterManagedUsers(users: ApiManagedUser[], search: string, labels: (user: ApiManagedUser) => string[]) {
	const needle = search.trim().toLowerCase();

	if (!needle) {
		return users;
	}

	return users.filter(user => managedUserSearchText(user, labels(user)).includes(needle));
}

function sameValue(left: unknown, right: unknown) {
	return JSON.stringify(left) === JSON.stringify(right);
}

export function AdminPage() {
	const { t } = useTranslation("admin");
	const { session } = useSession();
	const confirm = useConfirm();
	const prompt = usePromptDialog();
	const requireSudo = useRequireSudo();
	const importInputRef = useRef<HTMLInputElement | null>(null);
	const settingsQuery = useQuery({ ...adminQueries.settings(), enabled: Boolean(session?.user.isSystemAdmin) });
	const usersQuery = useQuery({ ...adminQueries.users(), enabled: Boolean(session?.user.isSystemAdmin) });
	const updateSettingsMutation = useUpdateAdminSettingsMutation();
	const updateManagedUserMutation = useUpdateManagedUserMutation();
	const setManagedUserPasswordMutation = useSetManagedUserPasswordMutation();
	const clearManagedUserPasswordMutation = useClearManagedUserPasswordMutation();
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
	const hasAdminSettingsChanges = Boolean(editedForm && !sameValue(editedForm, serverForm));
	const userRows = useMemo(() => usersQuery.data?.users ?? [], [usersQuery.data?.users]);
	const filteredUserRows = useMemo(
		() =>
			filterManagedUsers(userRows, userSearch, user => [
				user.disabledAt ? t("states.disabled") : t("states.active"),
				user.isSystemAdmin ? t("states.systemAdmin") : t("states.user"),
				user.emailVerified ? t("states.verified") : t("states.unverified")
			]),
		[userRows, userSearch, t]
	);
	const userCountLabel = userSearch.trim() ? t("filteredUsersCount", { filtered: filteredUserRows.length, total: userRows.length }) : t("usersCount", { count: userRows.length });
	const activeAdminCount = userRows.filter(user => user.isSystemAdmin && !user.disabledAt).length;

	const smtpPasswordLabel = useMemo(() => {
		if (!settingsQuery.data?.settings.smtp.passwordSet) {
			return t("smtpPasswordEmpty");
		}
		return t("smtpPasswordStored");
	}, [settingsQuery.data?.settings.smtp.passwordSet, t]);

	const userColumns = useMemo<DataColumn<ApiManagedUser>[]>(
		() => [
			{
				key: "user",
				label: t("columns.user"),
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
				label: t("columns.status"),
				render: user => (
					<span className={styles.adminCell}>
						<Badge tone={user.disabledAt ? "critical" : "success"}>{user.disabledAt ? t("states.disabled") : t("states.active")}</Badge>
						{user.disabledAt ? <span className={styles.adminMeta}>{formatTimestamp(user.disabledAt)}</span> : null}
					</span>
				),
				sortable: true,
				sortValue: user => (user.disabledAt ? 0 : 1)
			},
			{
				key: "access",
				label: t("columns.access"),
				render: user => <Badge tone={user.isSystemAdmin ? "accent" : "neutral"}>{user.isSystemAdmin ? t("states.systemAdmin") : t("states.user")}</Badge>,
				sortable: true,
				sortValue: user => (user.isSystemAdmin ? 1 : 0)
			},
			{
				key: "email",
				label: t("columns.email"),
				render: user => <Badge tone={user.emailVerified ? "success" : "warning"}>{user.emailVerified ? t("states.verified") : t("states.unverified")}</Badge>,
				sortable: true,
				sortValue: user => (user.emailVerified ? 1 : 0)
			},
			{
				key: "updatedAt",
				label: t("columns.updated"),
				render: user => <span className={styles.adminMeta}>{formatTimestamp(user.updatedAt)}</span>,
				sortable: true,
				sortValue: user => user.updatedAt
			}
		],
		[t]
	);

	if (!session) {
		return null;
	}

	if (!session.user.isSystemAdmin) {
		return (
			<PageStack>
				<ScreenHeader title={t("title")} />
				<Panel tone="deep" title={t("accessRequired")}>
					<BodyCopy>{t("accessRequiredDescription")}</BodyCopy>
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
		void requireSudo(() =>
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
						pushToast({ title: t("settings.saved"), message: t("settings.savedDescription"), tone: "success" });
					},
					onError: error => {
						pushToast({ title: t("settings.saveFailed"), message: requestErrorMessage(error, t("settings.saveError")), tone: "critical" });
					}
				}
			)
		);
	}

	async function toggleDisabled(user: ApiManagedUser) {
		const nextDisabled = !user.disabledAt;
		if (nextDisabled) {
			const accepted = await confirm({
				title: t("account.disable"),
				message: t("account.disableDescription", { email: user.email }),
				confirmLabel: t("account.disableAction"),
				tone: "danger"
			});
			if (!accepted) {
				return;
			}
		}

		await requireSudo(() =>
			updateManagedUserMutation.mutate(
				{ userId: user.id, body: { disabled: nextDisabled } },
				{
					onSuccess: data => {
						pushToast({
							title: data.user.disabledAt ? t("account.disabled") : t("account.enabled"),
							message: t("account.updated", { email: data.user.email }),
							tone: "success"
						});
					},
					onError: error => {
						pushToast({ title: t("account.updateFailed"), message: requestErrorMessage(error, t("account.updateError")), tone: "critical" });
					}
				}
			)
		);
	}

	async function toggleSystemAdmin(user: ApiManagedUser) {
		const nextAdmin = !user.isSystemAdmin;
		if (!nextAdmin) {
			const accepted = await confirm({
				title: t("account.revokeAdmin"),
				message: t("account.revokeAdminDescription", { email: user.email }),
				confirmLabel: t("account.revoke"),
				tone: "danger"
			});
			if (!accepted) {
				return;
			}
		}

		await requireSudo(() =>
			updateManagedUserMutation.mutate(
				{ userId: user.id, body: { systemAdmin: nextAdmin } },
				{
					onSuccess: data => {
						pushToast({
							title: data.user.isSystemAdmin ? t("account.adminGranted") : t("account.adminRevoked"),
							message: t("account.updated", { email: data.user.email }),
							tone: "success"
						});
					},
					onError: error => {
						pushToast({ title: t("account.permissionFailed"), message: requestErrorMessage(error, t("account.permissionError")), tone: "critical" });
					}
				}
			)
		);
	}

	async function setPassword(user: ApiManagedUser) {
		const password = await prompt({
			title: t("account.setPassword"),
			message: t("account.setPasswordDescription", { email: user.email }),
			inputLabel: t("account.newPassword"),
			inputType: "password",
			confirmLabel: t("account.setPasswordAction"),
			validate: value => (value.length < 8 ? t("account.passwordTooShort") : null)
		});
		if (!password) {
			return;
		}

		await requireSudo(() =>
			setManagedUserPasswordMutation.mutate(
				{ userId: user.id, body: { password } },
				{
					onSuccess: data => {
						pushToast({ title: t("account.passwordUpdated"), message: t("account.passwordUpdatedDescription", { email: data.user.email }), tone: "success" });
					},
					onError: error => {
						pushToast({ title: t("account.passwordUpdateFailed"), message: requestErrorMessage(error, t("account.passwordUpdateError")), tone: "critical" });
					}
				}
			)
		);
	}

	async function clearPassword(user: ApiManagedUser) {
		const accepted = await confirm({
			title: t("account.removePassword"),
			message: t("account.removePasswordDescription", { email: user.email }),
			confirmLabel: t("account.removePasswordAction"),
			tone: "danger"
		});
		if (!accepted) return;

		await requireSudo(() =>
			clearManagedUserPasswordMutation.mutate(user.id, {
				onSuccess: data => {
					pushToast({ title: t("account.passwordRemoved"), message: t("account.passwordRemovedDescription", { email: data.user.email }), tone: "success" });
				},
				onError: error => {
					pushToast({ title: t("account.passwordRemovalFailed"), message: requestErrorMessage(error, t("account.passwordRemovalError")), tone: "critical" });
				}
			})
		);
	}

	function exportData() {
		void requireSudo(() =>
			exportDataMutation.mutate(undefined, {
				onSuccess: data => {
					downloadAdminDataExport(data);
					pushToast({ title: t("data.exported"), message: t("data.exportedDescription"), tone: "success" });
				},
				onError: error => {
					pushToast({ title: t("data.exportFailed"), message: requestErrorMessage(error, t("data.exportError")), tone: "critical" });
				}
			})
		);
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
			pushToast({ title: t("data.importFailed"), message: t("data.invalidJson"), tone: "critical" });
			return;
		}
		if (!isAdminDataExport(parsed)) {
			pushToast({ title: t("data.importFailed"), message: t("data.invalidExport"), tone: "critical" });
			return;
		}

		const accepted = await confirm({
			title: t("data.confirmImport"),
			message: t("data.confirmImportDescription"),
			confirmLabel: t("data.confirmImportAction"),
			tone: "danger"
		});
		if (!accepted) {
			return;
		}

		await requireSudo(() =>
			importDataMutation.mutate(parsed, {
				onSuccess: data => {
					pushToast({ title: t("data.imported"), message: t("data.importedRows", { count: data.result.importedRows }), tone: "success" });
				},
				onError: error => {
					pushToast({ title: t("data.importFailed"), message: requestErrorMessage(error, t("data.importError")), tone: "critical" });
				}
			})
		);
	}

	function userRowActions(user: ApiManagedUser) {
		const isSelf = user.id === currentUserID;
		const lastActiveAdmin = user.isSystemAdmin && !user.disabledAt && activeAdminCount <= 1;
		const updatingUser = updateManagedUserMutation.isPending && updateManagedUserMutation.variables?.userId === user.id;
		const settingPassword = setManagedUserPasswordMutation.isPending && setManagedUserPasswordMutation.variables?.userId === user.id;
		const clearingPassword = clearManagedUserPasswordMutation.isPending && clearManagedUserPasswordMutation.variables === user.id;

		return (
			<div className={styles.userActions}>
				<Button type="button" size="sm" variant={user.disabledAt ? "outline" : "danger"} disabled={isSelf || lastActiveAdmin || updatingUser} onClick={() => void toggleDisabled(user)}>
					{user.disabledAt ? t("account.enable") : t("account.disableAction")}
				</Button>
				<Button type="button" size="sm" variant="ghost" disabled={(isSelf && user.isSystemAdmin) || lastActiveAdmin || updatingUser} onClick={() => void toggleSystemAdmin(user)}>
					{user.isSystemAdmin ? t("account.revokeAdmin") : t("account.grantAdmin")}
				</Button>
				<Button type="button" size="sm" variant="outline" disabled={settingPassword} onClick={() => void setPassword(user)}>
					{settingPassword ? t("account.setting") : t("account.setPasswordAction")}
				</Button>
				{user.hasPassword ? (
					<Button type="button" size="sm" variant="danger" disabled={clearingPassword} onClick={() => void clearPassword(user)}>
						{clearingPassword ? t("account.removing") : t("account.removePasswordAction")}
					</Button>
				) : null}
			</div>
		);
	}

	return (
		<PageStack>
			<ScreenHeader
				title={t("title")}
				actions={settingsQuery.data?.settings.smtp.configured ? <Badge tone="success">{t("smtpConfigured")}</Badge> : <Badge tone="warning">{t("smtpDisabled")}</Badge>}
			/>

			{settingsQuery.isLoading ? (
				<Spinner label={t("loading")} layout="panel" size="lg" />
			) : settingsQuery.isError ? (
				<Panel tone="deep" title={t("unavailable")}>
					<BodyCopy>{requestErrorMessage(settingsQuery.error, t("loadError"))}</BodyCopy>
				</Panel>
			) : (
				<form className={styles.form} onSubmit={handleSubmit}>
					<div className={styles.grid}>
						<Panel tone="glass" title={t("settings.instanceAccess")} bodyClassName={styles.instanceAccessOptions}>
							<label className={styles.checkboxRow}>
								<Checkbox checked={form.registrationEnabled} onChange={event => update("registrationEnabled", event.currentTarget.checked)} />
								<span>
									<strong>{t("settings.registration")}</strong>
									<small>{t("settings.registrationDescription")}</small>
								</span>
							</label>
							<label className={styles.checkboxRow}>
								<Checkbox checked={form.emailVerificationRequired} onChange={event => update("emailVerificationRequired", event.currentTarget.checked)} />
								<span>
									<strong>{t("settings.verification")}</strong>
									<small>{t("settings.verificationDescription")}</small>
								</span>
							</label>
						</Panel>

						<Panel tone="glass" title={t("settings.publicOrigins")}>
							<div className={styles.fieldStack}>
								<TextField
									label={t("settings.backendUrl")}
									value={form.backendBaseUrl}
									placeholder="https://app.netstamp.dev"
									onChange={event => update("backendBaseUrl", event.currentTarget.value)}
								/>
								<TextField
									label={t("settings.webUrl")}
									value={form.publicWebBaseUrl}
									placeholder="https://app.netstamp.dev"
									helper={t("settings.webUrlHelper")}
									onChange={event => update("publicWebBaseUrl", event.currentTarget.value)}
								/>
							</div>
						</Panel>
					</div>

					<Panel tone="glass" title={t("settings.smtpDelivery")}>
						<div className={styles.smtpGrid}>
							<TextField label={t("settings.host")} value={form.smtpHost} placeholder="smtp.example.com" onChange={event => update("smtpHost", event.currentTarget.value)} />
							<TextField label={t("settings.port")} type="number" min={1} max={65535} value={form.smtpPort} onChange={event => update("smtpPort", event.currentTarget.value)} />
							<TextField label={t("settings.username")} value={form.smtpUsername} autoComplete="off" onChange={event => update("smtpUsername", event.currentTarget.value)} />
							<TextField
								label={t("settings.password")}
								type="password"
								value={form.smtpPassword}
								autoComplete="new-password"
								helper={smtpPasswordLabel}
								disabled={form.smtpClearPassword}
								onChange={event => update("smtpPassword", event.currentTarget.value)}
							/>
							<TextField label={t("settings.from")} type="email" value={form.smtpFrom} placeholder="alerts@example.com" onChange={event => update("smtpFrom", event.currentTarget.value)} />
							<SelectField
								label={t("settings.tlsMode")}
								value={form.smtpTLSMode}
								options={[
									{ value: "starttls", label: "STARTTLS" },
									{ value: "implicit", label: t("settings.implicitTls") },
									{ value: "none", label: t("settings.none") }
								]}
								onChange={event => update("smtpTLSMode", event.currentTarget.value as AdminFormState["smtpTLSMode"])}
							/>
							<TextField label={t("settings.timeout")} type="number" min={1} value={form.smtpTimeoutSeconds} onChange={event => update("smtpTimeoutSeconds", event.currentTarget.value)} />
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
								<strong>{t("settings.clearPassword")}</strong>
								<small>{t("settings.clearPasswordDescription")}</small>
							</span>
						</label>
					</Panel>

					<UnsavedChangesBar show={hasAdminSettingsChanges} saveType="submit" saving={updateSettingsMutation.isPending} onReset={() => setEditedForm(null)} />
				</form>
			)}

			<Panel
				tone="glass"
				title={t("data.title")}
				actions={
					<div className={styles.toolActions}>
						<Button type="button" variant="outline" disabled={exportDataMutation.isPending} onClick={exportData}>
							{exportDataMutation.isPending ? t("data.exporting") : t("data.export")}
						</Button>
						<Button type="button" variant="danger" disabled={importDataMutation.isPending} onClick={() => void requireSudo(() => importInputRef.current?.click())}>
							{importDataMutation.isPending ? t("data.importing") : t("data.import")}
						</Button>
						<input ref={importInputRef} className={styles.fileInput} type="file" accept="application/json,.json" onChange={event => void importData(event)} />
					</div>
				}
			>
				<BodyCopy>{t("data.description")}</BodyCopy>
			</Panel>

			<Panel
				tone="glass"
				title={t("users.title")}
				actions={usersQuery.isFetching ? <Badge tone="neutral">{t("users.syncing")}</Badge> : <Badge tone="neutral">{userCountLabel}</Badge>}
				bodySurface="transparent"
				padded={false}
			>
				{usersQuery.isLoading ? (
					<Spinner label={t("users.loading")} layout="panel" size="lg" />
				) : usersQuery.isError ? (
					<div className={styles.tableMessage}>
						<BodyCopy>{requestErrorMessage(usersQuery.error, t("users.loadError"))}</BodyCopy>
					</div>
				) : (
					<>
						<div className={styles.userToolbar}>
							<TextField label={t("users.search")} type="search" placeholder={t("users.searchPlaceholder")} value={userSearch} onChange={event => setUserSearch(event.currentTarget.value)} />
						</div>
						<DataTable<ApiManagedUser>
							ariaLabel={t("users.aria")}
							columns={userColumns}
							rows={filteredUserRows}
							density="compact"
							minWidth="72rem"
							emptyLabel={userSearch.trim() ? t("users.noMatch") : t("users.empty")}
							getRowKey={user => user.id}
							rowActions={userRowActions}
							rowActionsClassName={styles.userActionsCell}
							rowActionsHeaderClassName={styles.userActionsHeader}
						/>
					</>
				)}
			</Panel>
		</PageStack>
	);
}
