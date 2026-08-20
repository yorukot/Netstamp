import { useRequireSudo } from "@/features/auth/hooks/useRequireSudo";
import { useLocaleFormat } from "@/i18n/format";
import type { AdminUpdateSettings } from "@/shared/api/adminSettings";
import { useUpdateAdminUpdateSettingsMutation } from "@/shared/api/mutations";
import { adminQueries } from "@/shared/api/queries";
import { pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { ArrowSquareOutIcon } from "@phosphor-icons/react/dist/csr/ArrowSquareOut";
import { Button, Checkbox, KeyValueRow, Panel } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import type { FormEvent } from "react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { applySettingsIntent, hasSettingsIntent, updateSettingsIntent } from "./adminSettingsForm";
import { SettingsNotice, SettingsPanelError, SettingsPanelLoading } from "./AdminSettingsPanelState";
import styles from "./AdminSettingsPanels.module.css";

const displayVersion = (value: string | null | undefined, fallback: string) => {
	if (!value) {
		return fallback;
	}
	return value.startsWith("v") ? value : `v${value}`;
};

export const AdminSystemInformationPanel = () => {
	const { t } = useTranslation("admin");
	const format = useLocaleFormat();
	const requireSudo = useRequireSudo();
	const settingsQuery = useQuery(adminQueries.updateSettings());
	const statusQuery = useQuery(adminQueries.updateStatus());
	const mutation = useUpdateAdminUpdateSettingsMutation();
	const [draft, setDraft] = useState<Partial<AdminUpdateSettings> | null>(null);
	const [saveError, setSaveError] = useState<string | null>(null);
	const serverSettings = settingsQuery.data?.settings;
	const settings = serverSettings ? applySettingsIntent(serverSettings, draft) : undefined;
	const status = statusQuery.data;
	const dirty = hasSettingsIntent(draft);
	const busy = mutation.isPending;

	const updateCheckSetting = (value: boolean) => {
		if (!serverSettings) {
			return;
		}
		setDraft(current => updateSettingsIntent(current, serverSettings, "checkForUpdates", value));
		setSaveError(null);
	};

	const reset = () => {
		setDraft(null);
		setSaveError(null);
	};

	const submit = (event: FormEvent<HTMLFormElement>) => {
		event.preventDefault();
		if (!draft || !dirty) {
			return;
		}

		void requireSudo(() =>
			mutation.mutate(draft, {
				onSuccess: () => {
					reset();
					pushToast({ title: t("settings.saved"), message: t("settings.savedDescription"), tone: "success" });
				},
				onError: error => {
					const message = requestErrorMessage(error, t("settings.saveError"));
					setSaveError(message);
					pushToast({ title: t("settings.saveFailed"), message, tone: "critical" });
				}
			})
		);
	};

	return (
		<Panel tone="glass" title={t("settings.systemInformation")} summary={t("settings.systemInformationSummary")}>
			{settingsQuery.isLoading ? (
				<SettingsPanelLoading />
			) : settingsQuery.isError || !settings ? (
				<SettingsPanelError error={settingsQuery.error} onRetry={() => void settingsQuery.refetch()} />
			) : (
				<form className={styles.form} onSubmit={submit}>
					<div className={styles.statusRows}>
						<KeyValueRow label={t("settings.currentVersion")} value={displayVersion(status?.currentVersion, t("settings.unavailableValue"))} />
						<KeyValueRow
							label={t("settings.latestVersion")}
							value={statusQuery.isLoading ? t("settings.checking") : displayVersion(status?.latestVersion, t("settings.notChecked"))}
							tone={status?.updateAvailable ? "warning" : "neutral"}
						/>
						<KeyValueRow
							label={t("settings.publishedAt")}
							value={status?.publishedAt ? format.dateTime(status.publishedAt, { dateStyle: "medium", timeStyle: "short" }) : t("settings.unavailableValue")}
						/>
						<KeyValueRow
							label={t("settings.lastCheckedAt")}
							value={status?.lastCheckedAt ? format.dateTime(status.lastCheckedAt, { dateStyle: "medium", timeStyle: "short" }) : t("settings.notChecked")}
						/>
					</div>

					{statusQuery.isError ? <SettingsNotice tone="critical">{requestErrorMessage(statusQuery.error, t("settings.updateStatusError"))}</SettingsNotice> : null}
					{status?.checkError ? <SettingsNotice tone="critical">{t("settings.updateCheckError", { error: status.checkError })}</SettingsNotice> : null}

					{status?.updateAvailable && status.releaseUrl ? (
						<a className={styles.releaseLink} href={status.releaseUrl} target="_blank" rel="noreferrer">
							{t("settings.viewReleaseNotes")}
							<ArrowSquareOutIcon size="1rem" aria-hidden="true" focusable="false" />
						</a>
					) : null}

					<div className={styles.checkboxStack}>
						<label className={styles.checkboxRow}>
							<Checkbox disabled={busy} checked={settings.checkForUpdates} onChange={event => updateCheckSetting(event.currentTarget.checked)} />
							<span>
								<strong>{t("settings.checkForUpdates")}</strong>
								<small>{t("settings.checkForUpdatesDescription")}</small>
							</span>
						</label>
					</div>

					{saveError ? <SettingsNotice tone="critical">{saveError}</SettingsNotice> : null}

					<div className={styles.formActions}>
						<Button type="button" size="sm" variant="outline" disabled={!dirty || busy} onClick={reset}>
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
