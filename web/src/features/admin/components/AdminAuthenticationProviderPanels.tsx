import { useRequireSudo } from "@/features/auth/hooks/useRequireSudo";
import type {
	AdminGitHubProviderSettings,
	AdminGitHubProviderSettingsPatch,
	AdminGoogleProviderSettings,
	AdminGoogleProviderSettingsPatch,
	AdminOIDCProviderSettings,
	AdminOIDCProviderSettingsPatch,
	AdminSettingsProvider,
	VersionedAdminSettings
} from "@/shared/api/adminSettings";
import {
	useUpdateAdminGitHubSettingsMutation,
	useUpdateAdminGoogleSettingsMutation,
	useUpdateAdminOIDCSettingsMutation,
	useValidateAdminGitHubSettingsMutation,
	useValidateAdminGoogleSettingsMutation,
	useValidateAdminOIDCSettingsMutation
} from "@/shared/api/mutations";
import { adminQueries } from "@/shared/api/queries";
import { pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { Badge, Button, Checkbox, Panel, TextAreaField, TextField } from "@netstamp/ui";
import { useQuery, type UseMutationResult, type UseQueryResult } from "@tanstack/react-query";
import type { FormEvent } from "react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { applySettingsIntent, googleAllowedDomainsFromText, hasSettingsIntent, isSettingsVersionConflict, secretValueFromForm, updateSettingsIntent } from "./adminSettingsForm";
import styles from "./AdminSettingsPanels.module.css";
import { SettingsNotice, SettingsPanelError, SettingsPanelLoading } from "./AdminSettingsPanelState";

type ProviderSettings = AdminOIDCProviderSettings | AdminGoogleProviderSettings | AdminGitHubProviderSettings;
type ProviderPatch = AdminOIDCProviderSettingsPatch | AdminGoogleProviderSettingsPatch | AdminGitHubProviderSettingsPatch;

interface ProviderFormState {
	enabled: boolean;
	issuerUrl: string;
	clientId: string;
	clientSecret: string;
	clearClientSecret: boolean;
	displayName: string;
	jitEnabled: boolean;
	allowedDomains: string;
	allowSignup: boolean;
}

interface VersionedProviderMutation<TPatch> {
	body: TPatch;
	etag: string;
}

interface ProviderSettingsPanelProps<TSettings extends ProviderSettings, TPatch extends ProviderPatch> {
	provider: AdminSettingsProvider;
	query: UseQueryResult<VersionedAdminSettings<TSettings>, Error>;
	updateMutation: UseMutationResult<VersionedAdminSettings<TSettings>, Error, VersionedProviderMutation<TPatch>>;
	validateMutation: UseMutationResult<void, Error, VersionedProviderMutation<TPatch>>;
	formFromSettings: (settings: TSettings) => ProviderFormState;
	patchFromIntent: (intent: Partial<ProviderFormState>) => TPatch;
}

const providerNames: Record<AdminSettingsProvider, string> = {
	oidc: "OIDC",
	google: "Google",
	github: "GitHub"
};

const providerBaseForm = (settings: ProviderSettings): ProviderFormState => ({
	enabled: settings.enabled,
	issuerUrl: "",
	clientId: settings.clientId,
	clientSecret: "",
	clearClientSecret: false,
	displayName: settings.displayName,
	jitEnabled: settings.jitEnabled,
	allowedDomains: "",
	allowSignup: true
});

const oidcFormFromSettings = (settings: AdminOIDCProviderSettings): ProviderFormState => ({
	...providerBaseForm(settings),
	issuerUrl: settings.issuerUrl
});

const googleFormFromSettings = (settings: AdminGoogleProviderSettings): ProviderFormState => ({
	...providerBaseForm(settings),
	allowedDomains: settings.allowedDomains.join(", ")
});

const githubFormFromSettings = (settings: AdminGitHubProviderSettings): ProviderFormState => ({
	...providerBaseForm(settings),
	allowSignup: settings.allowSignup
});

const secretPatch = (intent: Partial<ProviderFormState>) => {
	if (intent.clientSecret === undefined && intent.clearClientSecret === undefined) {
		return {};
	}
	const clientSecret = secretValueFromForm(intent.clientSecret ?? "", intent.clearClientSecret ?? false);
	return clientSecret === undefined ? {} : { clientSecret };
};

const basePatchFromIntent = (intent: Partial<ProviderFormState>) => ({
	...(intent.enabled === undefined ? {} : { enabled: intent.enabled }),
	...(intent.clientId === undefined ? {} : { clientId: intent.clientId }),
	...secretPatch(intent),
	...(intent.displayName === undefined ? {} : { displayName: intent.displayName }),
	...(intent.jitEnabled === undefined ? {} : { jitEnabled: intent.jitEnabled })
});

const oidcPatchFromIntent = (intent: Partial<ProviderFormState>): AdminOIDCProviderSettingsPatch => ({
	...basePatchFromIntent(intent),
	...(intent.issuerUrl === undefined ? {} : { issuerUrl: intent.issuerUrl })
});

const googlePatchFromIntent = (intent: Partial<ProviderFormState>): AdminGoogleProviderSettingsPatch => {
	return {
		...basePatchFromIntent(intent),
		...(intent.allowedDomains === undefined ? {} : { allowedDomains: googleAllowedDomainsFromText(intent.allowedDomains) })
	};
};

const githubPatchFromIntent = (intent: Partial<ProviderFormState>): AdminGitHubProviderSettingsPatch => ({
	...basePatchFromIntent(intent),
	...(intent.allowSignup === undefined ? {} : { allowSignup: intent.allowSignup })
});

const ProviderSettingsPanel = <TSettings extends ProviderSettings, TPatch extends ProviderPatch>({
	provider,
	query,
	updateMutation,
	validateMutation,
	formFromSettings,
	patchFromIntent
}: ProviderSettingsPanelProps<TSettings, TPatch>) => {
	const { t } = useTranslation("admin");
	const requireSudo = useRequireSudo();
	const [draft, setDraft] = useState<Partial<ProviderFormState> | null>(null);
	const [conflict, setConflict] = useState(false);
	const [saveError, setSaveError] = useState<string | null>(null);
	const [validation, setValidation] = useState<{ tone: "critical" | "success"; message: string } | null>(null);
	const serverSettings = query.data?.settings;
	const serverForm = serverSettings ? formFromSettings(serverSettings) : null;
	const form = serverForm ? applySettingsIntent(serverForm, draft) : null;
	const dirty = hasSettingsIntent(draft);
	const configured = Boolean(serverSettings?.clientId && serverSettings.clientSecretSet);
	const busy = updateMutation.isPending || validateMutation.isPending;

	const update = <TKey extends keyof ProviderFormState>(key: TKey, value: ProviderFormState[TKey]) => {
		if (!serverForm) {
			return;
		}
		setDraft(current => updateSettingsIntent(current, serverForm, key, value));
		setSaveError(null);
		setValidation(null);
	};

	const reset = () => {
		setDraft(null);
		setConflict(false);
		setSaveError(null);
		setValidation(null);
	};

	const handleConflict = () => {
		setConflict(true);
		setSaveError(null);
		setValidation(null);
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

	const validate = () => {
		if (!form?.enabled || !query.data) {
			return;
		}

		void requireSudo(() =>
			validateMutation.mutate(
				{ body: patchFromIntent(draft ?? {}), etag: query.data.etag },
				{
					onSuccess: () => {
						setValidation({ tone: "success", message: t("settings.configurationValidDescription") });
						pushToast({ title: t("settings.configurationValid"), message: t("settings.configurationValidDescription"), tone: "success" });
					},
					onError: error => {
						if (isSettingsVersionConflict(error)) {
							handleConflict();
							return;
						}
						const message = requestErrorMessage(error, t("settings.configurationInvalidError"));
						setValidation({ tone: "critical", message });
						pushToast({ title: t("settings.configurationInvalid"), message, tone: "critical" });
					}
				}
			)
		);
	};

	const actions = serverSettings ? (
		<div className={styles.panelActions}>
			<Badge tone={configured ? "success" : "warning"}>{configured ? t("settings.configured") : t("settings.notConfigured")}</Badge>
			<Button type="button" size="sm" variant="outline" disabled={!form?.enabled || busy} onClick={validate}>
				{validateMutation.isPending ? t("settings.validatingConfiguration") : t("settings.validateConfiguration")}
			</Button>
		</div>
	) : undefined;

	return (
		<Panel tone="deep" title={providerNames[provider]} actions={actions}>
			{query.isLoading ? (
				<SettingsPanelLoading />
			) : query.isError || !form || !serverSettings ? (
				<SettingsPanelError error={query.error} onRetry={() => void query.refetch()} />
			) : (
				<form className={styles.form} onSubmit={submit}>
					<label className={styles.checkboxRow}>
						<Checkbox disabled={busy} checked={form.enabled} onChange={event => update("enabled", event.currentTarget.checked)} />
						<span>
							<strong>{t("settings.enabled")}</strong>
						</span>
					</label>

					<div className={styles.fieldStack}>
						{provider === "oidc" ? (
							<TextField disabled={busy} label={t("settings.issuerUrl")} type="url" value={form.issuerUrl} onChange={event => update("issuerUrl", event.currentTarget.value)} />
						) : null}
						<TextField disabled={busy} label={t("settings.clientId")} value={form.clientId} onChange={event => update("clientId", event.currentTarget.value)} />
						<TextField
							label={t("settings.clientSecret")}
							type="password"
							value={form.clientSecret}
							autoComplete="new-password"
							helper={serverSettings.clientSecretSet ? t("settings.clientSecretStored") : t("settings.clientSecretEmpty")}
							disabled={busy || form.clearClientSecret}
							onChange={event => update("clientSecret", event.currentTarget.value)}
						/>
						<TextField disabled={busy} label={t("settings.displayName")} value={form.displayName} onChange={event => update("displayName", event.currentTarget.value)} />
						{provider === "google" ? (
							<TextAreaField
								label={t("settings.allowedDomains")}
								helper={t("settings.allowedDomainsDescription")}
								disabled={busy}
								rows={3}
								value={form.allowedDomains}
								onChange={event => update("allowedDomains", event.currentTarget.value)}
							/>
						) : null}
					</div>

					<div className={styles.checkboxStack}>
						<label className={styles.checkboxRow}>
							<Checkbox disabled={busy} checked={form.jitEnabled} onChange={event => update("jitEnabled", event.currentTarget.checked)} />
							<span>
								<strong>{t("settings.jitProvisioning")}</strong>
								<small>{t("settings.jitProvisioningDescription")}</small>
							</span>
						</label>
						<label className={styles.checkboxRow}>
							<Checkbox
								checked={form.clearClientSecret}
								disabled={busy || (!serverSettings.clientSecretSet && !form.clientSecret)}
								onChange={event => {
									update("clearClientSecret", event.currentTarget.checked);
									if (event.currentTarget.checked) {
										update("clientSecret", "");
									}
								}}
							/>
							<span>
								<strong>{t("settings.clearClientSecret")}</strong>
								<small>{t("settings.clearClientSecretDescription")}</small>
							</span>
						</label>
						{provider === "github" ? (
							<label className={styles.checkboxRow}>
								<Checkbox disabled={busy} checked={form.allowSignup} onChange={event => update("allowSignup", event.currentTarget.checked)} />
								<span>
									<strong>{t("settings.allowSignup")}</strong>
									<small>{t("settings.allowSignupDescription")}</small>
								</span>
							</label>
						) : null}
					</div>

					<div className={styles.callback}>
						<span className={styles.callbackLabel}>{t("settings.callbackUrl")}</span>
						<code className={styles.callbackValue}>{serverSettings.callbackUrl || t("settings.notConfigured")}</code>
						<span className={styles.callbackDescription}>{t("settings.callbackUrlDescription")}</span>
					</div>

					{conflict ? (
						<SettingsNotice tone="warning" dismissLabel={t("settings.dismiss")} onDismiss={() => setConflict(false)}>
							<strong>{t("settings.conflict")}</strong>
							<span>{t("settings.conflictDescription")}</span>
						</SettingsNotice>
					) : null}
					{saveError ? <SettingsNotice tone="critical">{saveError}</SettingsNotice> : null}
					{validation ? <SettingsNotice tone={validation.tone}>{validation.message}</SettingsNotice> : null}

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

const AdminOIDCSettingsPanel = () => {
	const query = useQuery(adminQueries.oidcSettings());
	const updateMutation = useUpdateAdminOIDCSettingsMutation();
	const validateMutation = useValidateAdminOIDCSettingsMutation();

	return (
		<ProviderSettingsPanel
			provider="oidc"
			query={query}
			updateMutation={updateMutation}
			validateMutation={validateMutation}
			formFromSettings={oidcFormFromSettings}
			patchFromIntent={oidcPatchFromIntent}
		/>
	);
};

const AdminGoogleSettingsPanel = () => {
	const query = useQuery(adminQueries.googleSettings());
	const updateMutation = useUpdateAdminGoogleSettingsMutation();
	const validateMutation = useValidateAdminGoogleSettingsMutation();

	return (
		<ProviderSettingsPanel
			provider="google"
			query={query}
			updateMutation={updateMutation}
			validateMutation={validateMutation}
			formFromSettings={googleFormFromSettings}
			patchFromIntent={googlePatchFromIntent}
		/>
	);
};

const AdminGitHubSettingsPanel = () => {
	const query = useQuery(adminQueries.githubSettings());
	const updateMutation = useUpdateAdminGitHubSettingsMutation();
	const validateMutation = useValidateAdminGitHubSettingsMutation();

	return (
		<ProviderSettingsPanel
			provider="github"
			query={query}
			updateMutation={updateMutation}
			validateMutation={validateMutation}
			formFromSettings={githubFormFromSettings}
			patchFromIntent={githubPatchFromIntent}
		/>
	);
};

export const AdminAuthenticationProviderPanels = () => {
	const { t } = useTranslation("admin");

	return (
		<Panel tone="glass" title={t("settings.authenticationProviders")} summary={t("settings.authenticationProvidersSummary")} bodySurface="transparent">
			<div className={styles.providerGrid}>
				<AdminOIDCSettingsPanel />
				<AdminGoogleSettingsPanel />
				<AdminGitHubSettingsPanel />
			</div>
		</Panel>
	);
};
