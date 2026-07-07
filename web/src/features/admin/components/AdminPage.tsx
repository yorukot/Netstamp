import { useSession } from "@/features/auth/session/SessionContext";
import { useUpdateAdminSettingsMutation } from "@/shared/api/mutations";
import { adminQueries } from "@/shared/api/queries";
import type { ApiAdminSettings } from "@/shared/api/types";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import { ActionRow, Badge, BodyCopy, Button, Checkbox, LoadingState, Panel, SelectField, TextField } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import type { FormEvent } from "react";
import { useMemo, useState } from "react";
import styles from "./AdminPage.module.css";

interface AdminFormState {
	registrationEnabled: boolean;
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

export function AdminPage() {
	const { session } = useSession();
	const settingsQuery = useQuery({ ...adminQueries.settings(), enabled: Boolean(session?.user.isSystemAdmin) });
	const updateSettingsMutation = useUpdateAdminSettingsMutation();
	const loadedSettings = settingsQuery.data?.settings;
	const serverForm = useMemo(() => {
		if (!loadedSettings) {
			return defaultForm;
		}
		return formFromSettings(loadedSettings);
	}, [loadedSettings]);
	const [editedForm, setEditedForm] = useState<AdminFormState | null>(null);
	const form = editedForm ?? serverForm;

	const smtpPasswordLabel = useMemo(() => {
		if (!settingsQuery.data?.settings.smtp.passwordSet) {
			return "No SMTP password is stored.";
		}
		return "A password is stored. Leave blank to keep it unchanged.";
	}, [settingsQuery.data?.settings.smtp.passwordSet]);

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

	return (
		<PageStack>
			<ScreenHeader title="Admin" actions={settingsQuery.data?.settings.smtp.configured ? <Badge tone="success">SMTP configured</Badge> : <Badge tone="warning">SMTP disabled</Badge>} />

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
						</Panel>

						<Panel tone="glass" title="Public origins">
							<div className={styles.fieldStack}>
								<TextField label="Backend base URL" value={form.backendBaseUrl} placeholder="https://app.netstamp.dev" onChange={event => update("backendBaseUrl", event.currentTarget.value)} />
								<TextField
									label="Public web base URL"
									value={form.publicWebBaseUrl}
									placeholder="https://app.netstamp.dev"
									helper="Password reset links use this origin when set."
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
