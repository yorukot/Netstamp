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
import { useMemo, useState, type FormEvent } from "react";
import styles from "./APITokensPanel.module.css";

const scopeOptions: Array<{ scope: ApiTokenScope; label: string }> = [
	{ scope: "projects:read", label: "Read projects" },
	{ scope: "projects:write", label: "Write projects" },
	{ scope: "probes:read", label: "Read probes" },
	{ scope: "probes:write", label: "Write probes" },
	{ scope: "checks:read", label: "Read checks" },
	{ scope: "checks:write", label: "Write checks" },
	{ scope: "labels:read", label: "Read labels" },
	{ scope: "labels:write", label: "Write labels" },
	{ scope: "assignments:read", label: "Read assignments" },
	{ scope: "results:read", label: "Read results" },
	{ scope: "alerts:read", label: "Read alerts" },
	{ scope: "alerts:write", label: "Write alerts" },
	{ scope: "status_pages:read", label: "Read status pages" },
	{ scope: "status_pages:write", label: "Write status pages" }
];

const defaultScopes = scopeOptions.filter(option => option.scope.endsWith(":read")).map(option => option.scope);
const expiryOptions = [7, 30, 90, 365].map(days => ({ value: String(days), label: `${days} days` }));

interface TokenForm {
	name: string;
	scopes: ApiTokenScope[];
	expiryDays: string;
}

const emptyForm: TokenForm = { name: "", scopes: defaultScopes, expiryDays: "90" };

function formatDateTime(value?: string) {
	return value ? new Date(value).toLocaleString() : "Never";
}

export function APITokensPanel({ requireSudo }: { requireSudo?: (action: () => void) => void }) {
	const query = useQuery(authQueries.apiTokens());
	const createMutation = useCreateAPITokenMutation();
	const revokeMutation = useRevokeAPITokenMutation();
	const confirm = useConfirm();
	const [open, setOpen] = useState(false);
	const [form, setForm] = useState<TokenForm>(emptyForm);
	const [createdValue, setCreatedValue] = useState<string | null>(null);

	const columns = useMemo<DataColumn<ApiToken>[]>(
		() => [
			{ key: "name", label: "Name", render: token => <strong>{token.name}</strong> },
			{ key: "token", label: "Token", render: token => <code>nst_pat_…{token.tokenHint}</code> },
			{ key: "scopes", label: "Scopes", render: token => <span className={styles.scopeSummary}>{token.scopes.length} scopes</span> },
			{ key: "lastUsed", label: "Last used", render: token => <span className={styles.time}>{formatDateTime(token.lastUsedAt)}</span> },
			{ key: "expires", label: "Expires", render: token => <span className={styles.time}>{formatDateTime(token.expiresAt)}</span> },
			{
				key: "status",
				label: "Status",
				render: token => {
					const active = new Date(token.expiresAt).valueOf() > Date.now();
					return <Badge tone={active ? "success" : "warning"}>{active ? "Active" : "Expired"}</Badge>;
				}
			}
		],
		[]
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
				onError: error => pushToast({ title: "Token creation failed", message: requestErrorMessage(error, "Could not create the API token."), tone: "critical" })
			}
		);
	}

	async function revoke(token: ApiToken) {
		const accepted = await confirm({ title: "Revoke API token?", message: `${token.name} will stop working immediately.`, confirmLabel: "Revoke", tone: "danger" });
		if (!accepted) return;
		revokeMutation.mutate(token.id, {
			onSuccess: () => pushToast({ title: "API token revoked", message: `${token.name} no longer has API access.`, tone: "success" }),
			onError: error => pushToast({ title: "Revocation failed", message: requestErrorMessage(error, "Could not revoke the API token."), tone: "critical" })
		});
	}

	const curlExample = createdValue
		? `curl -H "Authorization: Bearer ${createdValue}" -H "Accept: application/json" ${new URL(`${apiBaseUrl.replace(/\/$/, "")}/projects`, window.location.origin).toString()}`
		: "";

	return (
		<>
			<Panel
				tone="glass"
				title="API tokens"
				summary="Create scoped credentials for curl, CI, CLI tools, and backend integrations."
				actions={
					<Button type="button" size="sm" disabled={demoMode} onClick={startCreate}>
						<PlusIcon aria-hidden="true" focusable="false" />
						Create token
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
					ariaLabel="Personal API tokens"
					getRowKey={token => token.id}
					emptyLabel={query.isLoading ? <Spinner label="Loading API tokens" layout="compact" size="lg" /> : query.isError ? "API tokens could not be loaded" : "No API tokens"}
					rowActions={token => (
						<Button type="button" variant="danger" size="sm" disabled={demoMode || revokeMutation.isPending} onClick={() => void revoke(token)}>
							<TrashIcon aria-hidden="true" focusable="false" />
							Revoke
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
										<DialogTitle>Copy your API token</DialogTitle>
									</header>
									<DialogDescription id="api-token-dialog-description">This secret is shown once. Store it now; Netstamp cannot reveal it again.</DialogDescription>
									<CodeBlock title="Personal access token">{createdValue}</CodeBlock>
									<CodeBlock title="curl example">{curlExample}</CodeBlock>
									<div className={styles.dialogActions}>
										<Button type="button" onClick={closeDialog}>
											Done
										</Button>
									</div>
								</section>
							) : (
								<form className={styles.dialog} onSubmit={submit} onMouseDown={event => event.stopPropagation()}>
									<header className={styles.dialogHeader}>
										<KeyIcon aria-hidden="true" focusable="false" />
										<DialogTitle>Create API token</DialogTitle>
									</header>
									<DialogDescription id="api-token-dialog-description">Choose only the scopes this integration needs. Project roles still apply.</DialogDescription>
									<TextField label="Token name" value={form.name} maxLength={100} required onChange={event => setForm(current => ({ ...current, name: event.currentTarget.value }))} />
									<SelectField label="Expires after" value={form.expiryDays} options={expiryOptions} onChange={event => setForm(current => ({ ...current, expiryDays: event.currentTarget.value }))} />
									<fieldset className={styles.scopes}>
										<legend>Scopes</legend>
										{scopeOptions.map(option => (
											<label key={option.scope}>
												<Checkbox checked={form.scopes.includes(option.scope)} onChange={event => toggleScope(option.scope, event.currentTarget.checked)} />
												<span>{option.label}</span>
												<code>{option.scope}</code>
											</label>
										))}
									</fieldset>
									<BodyCopy>The token will expire automatically and can be revoked at any time.</BodyCopy>
									<div className={styles.dialogActions}>
										<Button type="button" variant="ghost" onClick={closeDialog}>
											Cancel
										</Button>
										<Button type="submit" disabled={!form.name.trim() || form.scopes.length === 0 || createMutation.isPending}>
											{createMutation.isPending ? "Creating" : "Create token"}
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
