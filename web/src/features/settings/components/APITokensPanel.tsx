import { useLocaleFormat } from "@/i18n/format";
import { apiBaseUrl } from "@/shared/api/client";
import { useCreateAPITokenMutation, useRevokeAPITokenMutation } from "@/shared/api/mutations";
import { authQueries } from "@/shared/api/queries";
import type { ApiToken, ApiTokenScope } from "@/shared/api/types";
import { useConfirm } from "@/shared/components/confirmContext";
import { demoMode } from "@/shared/config/features";
import { pushToast } from "@/shared/toast/toastStore";
import { requestErrorMessage } from "@/shared/utils/requestErrorMessage";
import {
	Badge,
	BodyCopy,
	Button,
	Checkbox,
	CodeBlock,
	DataTable,
	DialogContent,
	DialogDescription,
	DialogOverlay,
	DialogPortal,
	DialogRoot,
	DialogTitle,
	Panel,
	SelectField,
	Spinner,
	TextField,
	type DataColumn
} from "@netstamp/ui";
import { KeyIcon } from "@phosphor-icons/react/dist/csr/Key";
import { PlusIcon } from "@phosphor-icons/react/dist/csr/Plus";
import { TrashIcon } from "@phosphor-icons/react/dist/csr/Trash";
import { useQuery } from "@tanstack/react-query";
import { useCallback, useMemo, useState, type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import styles from "./APITokensPanel.module.css";

const scopeOptions = [
	{ scope: "projects:read", labelKey: "projectsRead" },
	{ scope: "projects:write", labelKey: "projectsWrite" },
	{ scope: "probes:read", labelKey: "probesRead" },
	{ scope: "probes:write", labelKey: "probesWrite" },
	{ scope: "checks:read", labelKey: "checksRead" },
	{ scope: "checks:write", labelKey: "checksWrite" },
	{ scope: "labels:read", labelKey: "labelsRead" },
	{ scope: "labels:write", labelKey: "labelsWrite" },
	{ scope: "assignments:read", labelKey: "assignmentsRead" },
	{ scope: "results:read", labelKey: "resultsRead" },
	{ scope: "alerts:read", labelKey: "alertsRead" },
	{ scope: "alerts:write", labelKey: "alertsWrite" },
	{ scope: "status_pages:read", labelKey: "statusPagesRead" },
	{ scope: "status_pages:write", labelKey: "statusPagesWrite" }
] as const satisfies ReadonlyArray<{ scope: ApiTokenScope; labelKey: string }>;

const defaultScopes = scopeOptions.filter(option => option.scope.endsWith(":read")).map(option => option.scope);

interface TokenForm {
	name: string;
	scopes: ApiTokenScope[];
	expiryDays: string;
}

const emptyForm: TokenForm = { name: "", scopes: defaultScopes, expiryDays: "90" };

export function APITokensPanel({ requireSudo }: { requireSudo?: (action: () => void) => void }) {
	const { t } = useTranslation(["settings", "common"]);
	const format = useLocaleFormat();
	const query = useQuery(authQueries.apiTokens());
	const createMutation = useCreateAPITokenMutation();
	const revokeMutation = useRevokeAPITokenMutation();
	const confirm = useConfirm();
	const [open, setOpen] = useState(false);
	const [form, setForm] = useState<TokenForm>(emptyForm);
	const [createdValue, setCreatedValue] = useState<string | null>(null);
	const expiryOptions = [7, 30, 90, 365].map(days => ({ value: String(days), label: t("apiTokens.expiryDays", { count: days }) }));
	const formatDateTime = useCallback((value?: string) => (value ? format.dateTime(value) : t("apiTokens.never")), [format, t]);

	const columns = useMemo<DataColumn<ApiToken>[]>(
		() => [
			{ key: "name", label: t("apiTokens.name"), render: token => <strong>{token.name}</strong> },
			{ key: "token", label: t("apiTokens.token"), render: token => <code>nst_pat_…{token.tokenHint}</code> },
			{ key: "scopes", label: t("apiTokens.scopes"), render: token => <span className={styles.scopeSummary}>{t("apiTokens.scopeCount", { count: token.scopes.length })}</span> },
			{ key: "lastUsed", label: t("apiTokens.lastUsed"), render: token => <span className={styles.time}>{formatDateTime(token.lastUsedAt)}</span> },
			{ key: "expires", label: t("apiTokens.expires"), render: token => <span className={styles.time}>{formatDateTime(token.expiresAt)}</span> },
			{
				key: "status",
				label: t("apiTokens.status"),
				render: token => {
					const active = new Date(token.expiresAt).valueOf() > Date.now();
					return <Badge tone={active ? "success" : "warning"}>{active ? t("apiTokens.active") : t("apiTokens.expired")}</Badge>;
				}
			}
		],
		[formatDateTime, t]
	);

	function startCreate() {
		const openDialog = () => {
			setForm(emptyForm);
			setCreatedValue(null);
			createMutation.reset();
			setOpen(true);
		};
		if (requireSudo) {
			requireSudo(openDialog);
		} else {
			openDialog();
		}
	}

	function closeDialog() {
		setOpen(false);
		setCreatedValue(null);
		setForm(emptyForm);
	}

	function toggleScope(scope: ApiTokenScope, checked: boolean) {
		setForm(current => ({ ...current, scopes: checked ? [...current.scopes, scope] : current.scopes.filter(candidate => candidate !== scope) }));
	}

	function submit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		const expiresAt = new Date(Date.now() + Number(form.expiryDays) * 24 * 60 * 60 * 1000).toISOString();
		createMutation.mutate(
			{ name: form.name.trim(), scopes: form.scopes, expiresAt },
			{
				onSuccess: response => {
					setCreatedValue(response.value);
				},
				onError: error => pushToast({ title: t("apiTokens.creationFailed"), message: requestErrorMessage(error, t("apiTokens.creationError")), tone: "critical" })
			}
		);
	}

	async function revoke(token: ApiToken) {
		const accepted = await confirm({ title: t("apiTokens.revokeQuestion"), message: t("apiTokens.revokeMessage", { name: token.name }), confirmLabel: t("apiTokens.revoke"), tone: "danger" });
		if (!accepted) return;
		revokeMutation.mutate(token.id, {
			onSuccess: () => pushToast({ title: t("apiTokens.revoked"), message: t("apiTokens.revokedMessage", { name: token.name }), tone: "success" }),
			onError: error => pushToast({ title: t("apiTokens.revokeFailed"), message: requestErrorMessage(error, t("apiTokens.revokeError")), tone: "critical" })
		});
	}

	const curlExample = createdValue
		? `curl -H "Authorization: Bearer ${createdValue}" -H "Accept: application/json" ${new URL(`${apiBaseUrl.replace(/\/$/, "")}/projects`, window.location.origin).toString()}`
		: "";

	return (
		<>
			<Panel
				tone="glass"
				title={t("apiTokens.title")}
				summary={t("apiTokens.summary")}
				actions={
					<Button type="button" size="sm" disabled={demoMode} onClick={startCreate}>
						<PlusIcon aria-hidden="true" focusable="false" />
						{t("apiTokens.create")}
					</Button>
				}
				padded={false}
				bodySurface="transparent"
			>
				<DataTable
					columns={columns}
					rows={query.data?.tokens ?? []}
					density="compact"
					minWidth="68rem"
					ariaLabel={t("apiTokens.aria")}
					getRowKey={token => token.id}
					emptyLabel={query.isLoading ? <Spinner label={t("apiTokens.loading")} layout="compact" size="lg" /> : query.isError ? t("apiTokens.loadError") : t("apiTokens.empty")}
					rowActions={token => (
						<Button type="button" variant="danger" size="sm" disabled={demoMode || revokeMutation.isPending} onClick={() => void revoke(token)}>
							<TrashIcon aria-hidden="true" focusable="false" />
							{t("apiTokens.revoke")}
						</Button>
					)}
				/>
			</Panel>

			<DialogRoot open={open} onOpenChange={next => !next && closeDialog()}>
				<DialogPortal>
					<DialogOverlay onMouseDown={closeDialog}>
						<DialogContent asChild aria-describedby="api-token-dialog-description">
							{createdValue ? (
								<section className={styles.dialog} onMouseDown={event => event.stopPropagation()}>
									<header className={styles.dialogHeader}>
										<KeyIcon aria-hidden="true" focusable="false" />
										<DialogTitle>{t("apiTokens.copyTitle")}</DialogTitle>
									</header>
									<DialogDescription id="api-token-dialog-description">{t("apiTokens.copyDescription")}</DialogDescription>
									<CodeBlock title={t("apiTokens.personalToken")}>{createdValue}</CodeBlock>
									<CodeBlock title={t("apiTokens.curlExample")}>{curlExample}</CodeBlock>
									<div className={styles.dialogActions}>
										<Button type="button" onClick={closeDialog}>
											{t("common:actions.done")}
										</Button>
									</div>
								</section>
							) : (
								<form className={styles.dialog} onSubmit={submit} onMouseDown={event => event.stopPropagation()}>
									<header className={styles.dialogHeader}>
										<KeyIcon aria-hidden="true" focusable="false" />
										<DialogTitle>{t("apiTokens.createTitle")}</DialogTitle>
									</header>
									<DialogDescription id="api-token-dialog-description">{t("apiTokens.createDescription")}</DialogDescription>
									<TextField label={t("apiTokens.tokenName")} value={form.name} maxLength={100} required onChange={event => setForm(current => ({ ...current, name: event.currentTarget.value }))} />
									<SelectField
										label={t("apiTokens.expiresAfter")}
										value={form.expiryDays}
										options={expiryOptions}
										onChange={event => setForm(current => ({ ...current, expiryDays: event.currentTarget.value }))}
									/>
									<fieldset className={styles.scopes}>
										<legend>{t("apiTokens.scopes")}</legend>
										{scopeOptions.map(option => (
											<label key={option.scope}>
												<Checkbox checked={form.scopes.includes(option.scope)} onChange={event => toggleScope(option.scope, event.currentTarget.checked)} />
												<span>{t(`apiTokens.scopeLabels.${option.labelKey}`)}</span>
												<code>{option.scope}</code>
											</label>
										))}
									</fieldset>
									<BodyCopy>{t("apiTokens.expiryDescription")}</BodyCopy>
									<div className={styles.dialogActions}>
										<Button type="button" variant="ghost" onClick={closeDialog}>
											{t("common:actions.cancel")}
										</Button>
										<Button type="submit" disabled={!form.name.trim() || form.scopes.length === 0 || createMutation.isPending}>
											{createMutation.isPending ? t("apiTokens.creating") : t("apiTokens.create")}
										</Button>
									</div>
								</form>
							)}
						</DialogContent>
					</DialogOverlay>
				</DialogPortal>
			</DialogRoot>
		</>
	);
}
