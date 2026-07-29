import { useRequireSudo } from "@/features/auth/hooks/useRequireSudo";
import type { AdminSMTPSettings, AdminSMTPSettingsPatch, AdminSMTPTLSMode } from "@/shared/api/adminSettings";
import { useTestAdminSMTPMutation, useUpdateAdminSMTPSettingsMutation } from "@/shared/api/mutations";
import { adminQueries } from "@/shared/api/queries";
import { pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Badge, Button, Checkbox, Panel, SelectField, TextField } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import type { FormEvent } from "react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { applySettingsIntent, hasSettingsIntent, isSettingsVersionConflict, secretValueFromForm, updateSettingsIntent } from "./adminSettingsForm";
import styles from "./AdminSettingsPanels.module.css";
import { SettingsNotice, SettingsPanelError, SettingsPanelLoading } from "./AdminSettingsPanelState";

interface SMTPFormState {
	host: string;
	port: string;
	username: string;
	password: string;
	clearPassword: boolean;
	from: string;
	tlsMode: AdminSMTPTLSMode;
	timeoutSeconds: string;
}

const formFromSettings = (settings: AdminSMTPSettings): SMTPFormState => ({
	host: settings.host,
	port: String(settings.port),
	username: settings.username,
	password: "",
	clearPassword: false,
	from: settings.from,
	tlsMode: settings.tlsMode,
	timeoutSeconds: String(settings.timeoutSeconds)
});

const patchFromIntent = (intent: Partial<SMTPFormState>): AdminSMTPSettingsPatch => {
	const patch: AdminSMTPSettingsPatch = {};

	if (intent.host !== undefined) patch.host = intent.host;
	if (intent.port !== undefined) patch.port = Number(intent.port);
	if (intent.username !== undefined) patch.username = intent.username;
	if (intent.from !== undefined) patch.from = intent.from;
	if (intent.tlsMode !== undefined) patch.tlsMode = intent.tlsMode;
	if (intent.timeoutSeconds !== undefined) patch.timeoutSeconds = Number(intent.timeoutSeconds);

	if (intent.password !== undefined || intent.clearPassword !== undefined) {
		const password = secretValueFromForm(intent.password ?? "", intent.clearPassword ?? false);
		if (password !== undefined) patch.password = password;
	}

	return patch;
};

export const AdminSMTPSettingsPanel = () => {
	const { t } = useTranslation("admin");
	const requireSudo = useRequireSudo();
	const query = useQuery(adminQueries.smtpSettings());
	const updateMutation = useUpdateAdminSMTPSettingsMutation();
	const testMutation = useTestAdminSMTPMutation();
	const [draft, setDraft] = useState<Partial<SMTPFormState> | null>(null);
	const [conflict, setConflict] = useState(false);
	const [saveError, setSaveError] = useState<string | null>(null);
	const [testError, setTestError] = useState<string | null>(null);
	const serverSettings = query.data?.settings;
	const serverForm = serverSettings ? formFromSettings(serverSettings) : null;
	const form = serverForm ? applySettingsIntent(serverForm, draft) : null;
	const dirty = hasSettingsIntent(draft);
	const busy = updateMutation.isPending || testMutation.isPending;

	const update = <TKey extends keyof SMTPFormState>(key: TKey, value: SMTPFormState[TKey]) => {
		if (!serverForm) {
			return;
		}
		setDraft(current => updateSettingsIntent(current, serverForm, key, value));
		setSaveError(null);
		setTestError(null);
	};

	const reset = () => {
		setDraft(null);
		setConflict(false);
		setSaveError(null);
		setTestError(null);
	};

	const handleConflict = () => {
		setConflict(true);
		setSaveError(null);
		setTestError(null);
		void query.refetch();
	};

	const submit = (event: FormEvent<HTMLFormElement>) => {
		event.preventDefault();
		if (!form || !query.data || !draft || !dirty) {
			return;
		}

		void requireSudo(() =>
			updateMutation.mutate(
				{ body: patchFromIntent(draft), etag: query.data.etag },
				{
					onSuccess: () => {
						reset();
						pushToast({ title: t("settings.saved"), message: t("settings.savedDescription"), tone: "success" });
					},
					onError: error => {
						if (isSettingsVersionConflict(error)) {
							handleConflict();
							return;
						}
						const message = requestErrorMessage(error, t("settings.saveError"));
						setSaveError(message);
						pushToast({ title: t("settings.saveFailed"), message, tone: "critical" });
					}
				}
			)
		);
	};

	const testSMTP = () => {
		if (!query.data || dirty || !serverSettings?.configured) {
			return;
		}

		void requireSudo(() =>
			testMutation.mutate(query.data.etag, {
				onSuccess: () => {
					setTestError(null);
					pushToast({ title: t("settings.testEmailSent"), message: t("settings.testEmailSentDescription"), tone: "success" });
				},
				onError: error => {
					if (isSettingsVersionConflict(error)) {
						handleConflict();
						return;
					}
					const message = requestErrorMessage(error, t("settings.testEmailError"));
					setTestError(message);
					pushToast({ title: t("settings.testEmailFailed"), message, tone: "critical" });
				}
			})
		);
	};

	const actions = serverSettings ? (
		<div className={styles.panelActions}>
			<Badge tone={serverSettings.configured ? "success" : "warning"}>{serverSettings.configured ? t("settings.configured") : t("settings.notConfigured")}</Badge>
			<Button type="button" size="sm" variant="outline" disabled={!serverSettings.configured || dirty || busy} onClick={testSMTP}>
				{testMutation.isPending ? t("settings.testingEmail") : t("settings.testEmail")}
			</Button>
		</div>
	) : undefined;

	return (
		<Panel tone="glass" title={t("settings.smtpDelivery")} summary={t("settings.smtpSummary")} actions={actions}>
			{query.isLoading ? (
				<SettingsPanelLoading />
			) : query.isError || !form || !serverSettings ? (
				<SettingsPanelError error={query.error} onRetry={() => void query.refetch()} />
			) : (
				<form className={styles.form} onSubmit={submit}>
					<div className={styles.fieldGrid}>
						<TextField disabled={busy} label={t("settings.host")} value={form.host} placeholder="smtp.example.com" onChange={event => update("host", event.currentTarget.value)} />
						<TextField disabled={busy} label={t("settings.port")} type="number" min={1} max={65535} required value={form.port} onChange={event => update("port", event.currentTarget.value)} />
						<TextField disabled={busy} label={t("settings.username")} value={form.username} autoComplete="off" onChange={event => update("username", event.currentTarget.value)} />
						<TextField
							label={t("settings.password")}
							type="password"
							value={form.password}
							autoComplete="new-password"
							helper={serverSettings.passwordSet ? t("smtpPasswordStored") : t("smtpPasswordEmpty")}
							disabled={busy || form.clearPassword}
							onChange={event => update("password", event.currentTarget.value)}
						/>
						<TextField disabled={busy} label={t("settings.from")} type="email" value={form.from} placeholder="alerts@example.com" onChange={event => update("from", event.currentTarget.value)} />
						<SelectField
							label={t("settings.tlsMode")}
							disabled={busy}
							value={form.tlsMode}
							options={[
								{ value: "starttls", label: "STARTTLS" },
								{ value: "implicit", label: t("settings.implicitTls") },
								{ value: "none", label: t("settings.none") }
							]}
							onChange={event => update("tlsMode", event.currentTarget.value as AdminSMTPTLSMode)}
						/>
						<TextField
							disabled={busy}
							label={t("settings.timeout")}
							type="number"
							min={1}
							required
							value={form.timeoutSeconds}
							onChange={event => update("timeoutSeconds", event.currentTarget.value)}
						/>
					</div>

					<label className={styles.checkboxRow}>
						<Checkbox
							checked={form.clearPassword}
							disabled={busy || (!serverSettings.passwordSet && !form.password)}
							onChange={event => {
								update("clearPassword", event.currentTarget.checked);
								if (event.currentTarget.checked) {
									update("password", "");
								}
							}}
						/>
						<span>
							<strong>{t("settings.clearPassword")}</strong>
							<small>{t("settings.clearPasswordDescription")}</small>
						</span>
					</label>

					{dirty ? <p className={styles.hint}>{t("settings.smtpUnsavedTestDescription")}</p> : null}
					{conflict ? (
						<SettingsNotice tone="warning" dismissLabel={t("settings.dismiss")} onDismiss={() => setConflict(false)}>
							<strong>{t("settings.conflict")}</strong>
							<span>{t("settings.conflictDescription")}</span>
						</SettingsNotice>
					) : null}
					{saveError ? <SettingsNotice tone="critical">{saveError}</SettingsNotice> : null}
					{testError ? <SettingsNotice tone="critical">{testError}</SettingsNotice> : null}

					<div className={styles.formActions}>
						<Button type="button" size="sm" variant="plain" disabled={!dirty || busy} onClick={reset}>
							{t("settings.reset")}
						</Button>
						<Button type="submit" size="sm" disabled={!dirty || busy}>
							{updateMutation.isPending ? t("settings.saving") : t("settings.save")}
						</Button>
					</div>
				</form>
			)}
		</Panel>
	);
};
