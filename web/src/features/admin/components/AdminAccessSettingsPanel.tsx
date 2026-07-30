import { useRequireSudo } from "@/features/auth/hooks/useRequireSudo";
import type { AdminAccessSettings } from "@/shared/api/adminSettings";
import { useUpdateAdminAccessSettingsMutation } from "@/shared/api/mutations";
import { adminQueries } from "@/shared/api/queries";
import { loadRuntimeConfig } from "@/shared/config/runtime";
import { pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Button, Checkbox, Panel } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import type { FormEvent } from "react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { applySettingsIntent, hasSettingsIntent, isSettingsVersionConflict, updateSettingsIntent } from "./adminSettingsForm";
import styles from "./AdminSettingsPanels.module.css";
import { SettingsNotice, SettingsPanelError, SettingsPanelLoading } from "./AdminSettingsPanelState";

export const AdminAccessSettingsPanel = () => {
	const { t } = useTranslation("admin");
	const requireSudo = useRequireSudo();
	const query = useQuery(adminQueries.accessSettings());
	const mutation = useUpdateAdminAccessSettingsMutation();
	const [draft, setDraft] = useState<Partial<AdminAccessSettings> | null>(null);
	const [conflict, setConflict] = useState(false);
	const [saveError, setSaveError] = useState<string | null>(null);
	const serverSettings = query.data?.settings;
	const settings = serverSettings ? applySettingsIntent(serverSettings, draft) : undefined;
	const dirty = hasSettingsIntent(draft);
	const busy = mutation.isPending;

	const update = <TKey extends keyof AdminAccessSettings>(key: TKey, value: AdminAccessSettings[TKey]) => {
		if (!serverSettings) {
			return;
		}
		setDraft(current => updateSettingsIntent(current, serverSettings, key, value));
		setSaveError(null);
	};

	const reset = () => {
		setDraft(null);
		setConflict(false);
		setSaveError(null);
	};

	const handleConflict = () => {
		setConflict(true);
		setSaveError(null);
		void query.refetch();
	};

	const submit = (event: FormEvent<HTMLFormElement>) => {
		event.preventDefault();
		if (!settings || !query.data || !draft || !dirty) {
			return;
		}

		void requireSudo(() =>
			mutation.mutate(
				{ body: draft, etag: query.data.etag },
				{
					onSuccess: () => {
						reset();
						void loadRuntimeConfig().catch(() => undefined);
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

	return (
		<Panel tone="glass" title={t("settings.instanceAccess")} summary={t("settings.accessSummary")}>
			{query.isLoading ? (
				<SettingsPanelLoading />
			) : query.isError || !settings ? (
				<SettingsPanelError error={query.error} onRetry={() => void query.refetch()} />
			) : (
				<form className={styles.form} onSubmit={submit}>
					<div className={styles.checkboxStack}>
						<label className={styles.checkboxRow}>
							<Checkbox disabled={busy} checked={settings.accountCreationEnabled} onChange={event => update("accountCreationEnabled", event.currentTarget.checked)} />
							<span>
								<strong>{t("settings.accountCreation")}</strong>
								<small>{t("settings.accountCreationDescription")}</small>
							</span>
						</label>
						<label className={styles.checkboxRow}>
							<Checkbox disabled={busy} checked={settings.emailVerificationRequired} onChange={event => update("emailVerificationRequired", event.currentTarget.checked)} />
							<span>
								<strong>{t("settings.verification")}</strong>
								<small>{t("settings.verificationDescription")}</small>
							</span>
						</label>
						<label className={styles.checkboxRow}>
							<Checkbox disabled={busy} checked={settings.projectCreationEnabled} onChange={event => update("projectCreationEnabled", event.currentTarget.checked)} />
							<span>
								<strong>{t("settings.projectCreation")}</strong>
								<small>{t("settings.projectCreationDescription")}</small>
							</span>
						</label>
						<label className={styles.checkboxRow}>
							<Checkbox disabled={busy} checked={settings.credentialChangesEnabled} onChange={event => update("credentialChangesEnabled", event.currentTarget.checked)} />
							<span>
								<strong>{t("settings.credentialChanges")}</strong>
								<small>{t("settings.credentialChangesDescription")}</small>
							</span>
						</label>
					</div>

					{conflict ? (
						<SettingsNotice tone="warning" dismissLabel={t("settings.dismiss")} onDismiss={() => setConflict(false)}>
							<strong>{t("settings.conflict")}</strong>
							<span>{t("settings.conflictDescription")}</span>
						</SettingsNotice>
					) : null}
					{saveError ? <SettingsNotice tone="critical">{saveError}</SettingsNotice> : null}

					<div className={styles.formActions}>
						<Button type="button" size="sm" variant="plain" disabled={!dirty || busy} onClick={reset}>
							{t("settings.reset")}
						</Button>
						<Button type="submit" size="sm" disabled={!dirty || busy}>
							{mutation.isPending ? t("settings.saving") : t("settings.save")}
						</Button>
					</div>
				</form>
			)}
		</Panel>
	);
};
