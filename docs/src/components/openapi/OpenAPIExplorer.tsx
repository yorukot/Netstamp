import { Badge, Button, FieldLabel, Input, Panel } from "@netstamp/ui";
import { useDeferredValue, useEffect, useState } from "react";
import styles from "./OpenAPIExplorer.module.css";

type HTTPMethod = "get" | "post" | "put" | "patch" | "delete";

interface OpenAPISpec {
	openapi?: string;
	info?: { title?: string; version?: string; description?: string };
	servers?: Array<{ url: string; description?: string }>;
	components?: { schemas?: Record<string, JSONSchema> };
	paths?: Record<string, Partial<Record<HTTPMethod, OpenAPIOperation>>>;
}

interface OpenAPIOperation {
	operationId?: string;
	summary?: string;
	description?: string;
	tags?: string[];
	parameters?: OpenAPIParameter[];
	requestBody?: {
		required?: boolean;
		content?: Record<string, { schema?: JSONSchema; example?: unknown; examples?: Record<string, { value?: unknown }> }>;
	};
	responses?: Record<string, OpenAPIResponse>;
}

interface OpenAPIParameter {
	name: string;
	in: string;
	required?: boolean;
	description?: string;
	schema?: JSONSchema;
}

interface OpenAPIResponse {
	description?: string;
	content?: Record<string, { schema?: JSONSchema }>;
}

interface JSONSchema {
	$ref?: string;
	type?: string;
	format?: string;
	description?: string;
	properties?: Record<string, JSONSchema>;
	required?: string[];
	items?: JSONSchema;
	enum?: unknown[];
	example?: unknown;
	examples?: unknown[];
	default?: unknown;
	readOnly?: boolean;
}

interface OperationItem extends OpenAPIOperation {
	key: string;
	method: HTTPMethod;
	path: string;
	tag: string;
}

interface OperationGroup {
	tag: string;
	operations: OperationItem[];
}

interface SchemaField {
	name: string;
	type: string;
	description?: string;
	required: boolean;
}

interface OpenAPIExplorerProps {
	specUrl: string;
}

const httpMethods = ["get", "post", "put", "patch", "delete"] as const satisfies readonly HTTPMethod[];

const methodLabels: Record<HTTPMethod, string> = {
	get: "GET",
	post: "POST",
	put: "PUT",
	patch: "PATCH",
	delete: "DELETE"
};

const methodTones: Record<HTTPMethod, "success" | "accent" | "warning" | "critical"> = {
	get: "success",
	post: "accent",
	put: "warning",
	patch: "warning",
	delete: "critical"
};

function classNames(...classes: Array<string | false | undefined>) {
	return classes.filter(Boolean).join(" ");
}

function slugify(value: string) {
	return (
		value
			.toLowerCase()
			.replace(/[^a-z0-9]+/g, "-")
			.replace(/^-+|-+$/g, "") || "section"
	);
}

function tagAnchor(tag: string) {
	return `tag-${slugify(tag)}`;
}

function operationAnchor(operation: OperationItem) {
	return `operation-${slugify(operation.operationId ?? `${operation.method}-${operation.path}`)}`;
}

function collectOperations(spec: OpenAPISpec): OperationItem[] {
	const operations: OperationItem[] = [];

	Object.entries(spec.paths ?? {}).forEach(([path, pathItem]) => {
		httpMethods.forEach(method => {
			const operation = pathItem[method];
			if (!operation) return;

			operations.push({
				...operation,
				key: `${method}:${path}`,
				method,
				path,
				tag: operation.tags?.[0] ?? "API"
			});
		});
	});

	return operations;
}

function groupOperations(operations: OperationItem[]): OperationGroup[] {
	const groups = new Map<string, OperationItem[]>();

	for (const operation of operations) {
		groups.set(operation.tag, [...(groups.get(operation.tag) ?? []), operation]);
	}

	return Array.from(groups, ([tag, groupedOperations]) => ({ tag, operations: groupedOperations }));
}

function resolveSchema(schema: JSONSchema | undefined, spec: OpenAPISpec | undefined, seen: Set<string>): JSONSchema | undefined {
	if (!schema?.$ref) return schema;

	const ref = schema.$ref.replace("#/components/schemas/", "");
	if (seen.has(ref)) return schema;

	seen.add(ref);
	return resolveSchema(spec?.components?.schemas?.[ref], spec, seen);
}

function valueFromSchema(schema: JSONSchema | undefined, spec: OpenAPISpec | undefined, seen = new Set<string>()): unknown {
	schema = resolveSchema(schema, spec, seen);
	if (!schema) return undefined;
	if (schema.example !== undefined) return schema.example;
	if (schema.examples?.[0] !== undefined) return schema.examples[0];
	if (schema.default !== undefined) return schema.default;
	if (schema.enum?.[0] !== undefined) return schema.enum[0];

	if (schema.type === "object" && schema.properties) {
		const value: Record<string, unknown> = {};
		Object.entries(schema.properties).forEach(([key, property]) => {
			if (property.readOnly) return;
			value[key] = valueFromSchema(property, spec, new Set(seen)) ?? "";
		});
		return value;
	}

	if (schema.type === "array") return schema.items ? [valueFromSchema(schema.items, spec, new Set(seen))] : [];
	if (schema.type === "number" || schema.type === "integer") return 0;
	if (schema.type === "boolean") return false;
	return "";
}

function requestContent(operation: OperationItem | undefined) {
	return operation?.requestBody?.content?.["application/json"];
}

function requestBodyExample(operation: OperationItem | undefined, spec: OpenAPISpec | undefined) {
	const json = requestContent(operation);
	if (!json) return "";
	if (json.example !== undefined) return JSON.stringify(json.example, null, 2);

	const example = Object.values(json.examples ?? {})[0]?.value;
	if (example !== undefined) return JSON.stringify(example, null, 2);

	const schemaValue = valueFromSchema(json.schema, spec);
	return schemaValue === undefined ? "" : JSON.stringify(schemaValue, null, 2);
}

function serverUrl(spec: OpenAPISpec) {
	return spec.servers?.[0]?.url ?? "/api/v1";
}

function schemaType(schema: JSONSchema | undefined, spec: OpenAPISpec | undefined): string {
	if (!schema) return "unknown";
	if (schema.$ref) return schema.$ref.replace("#/components/schemas/", "");

	const resolved = resolveSchema(schema, spec, new Set());
	if (resolved?.enum?.length) return "enum";
	if (resolved?.type === "array") return `${schemaType(resolved.items, spec)}[]`;
	return [resolved?.type, resolved?.format].filter(Boolean).join("/") || "value";
}

function requestFields(operation: OperationItem | undefined, spec: OpenAPISpec | undefined): SchemaField[] {
	const schema = resolveSchema(requestContent(operation)?.schema, spec, new Set());
	if (!schema?.properties) return [];

	const requiredFields = new Set(schema.required ?? []);
	return Object.entries(schema.properties)
		.filter(([, property]) => !property.readOnly)
		.map(([name, property]) => ({
			name,
			type: schemaType(property, spec),
			description: property.description,
			required: requiredFields.has(name)
		}));
}

function responseEntries(operation: OperationItem | undefined) {
	return Object.entries(operation?.responses ?? {});
}

function statusTone(status: string) {
	if (status.startsWith("2")) return styles.statusSuccess;
	if (status.startsWith("4")) return styles.statusWarning;
	if (status.startsWith("5")) return styles.statusCritical;
	return styles.statusMuted;
}

function operationSearchText(operation: OperationItem) {
	return [operation.tag, operation.path, operation.summary, operation.description, operation.operationId, methodLabels[operation.method]].filter(Boolean).join(" ").toLowerCase();
}

function shellEscape(value: string) {
	return value.replaceAll("'", "'\"'\"'");
}

function requestUrl(baseUrl: string, path: string) {
	const trimmedBase = baseUrl.replace(/\/$/, "");
	const requestPath = path.startsWith("/") ? path : `/${path}`;
	return `${trimmedBase}${requestPath}`;
}

function curlCommand(operation: OperationItem | undefined, baseUrl: string, path: string, body: string, token: string) {
	if (!operation) return "Select an operation to generate a request.";

	const lines = [`curl -X ${methodLabels[operation.method]} '${shellEscape(requestUrl(baseUrl, path))}'`, "  -H 'Accept: application/json'"];

	if (token.trim()) {
		lines.push(`  -H 'Authorization: Bearer ${shellEscape(token.trim())}'`);
	}

	if (body.trim() && operation.method !== "get") {
		lines.push("  -H 'Content-Type: application/json'", `  --data '${shellEscape(body.trim())}'`);
	}

	return lines.join(" \\" + "\n");
}

export default function OpenAPIExplorer({ specUrl }: OpenAPIExplorerProps) {
	const [spec, setSpec] = useState<OpenAPISpec>();
	const [operations, setOperations] = useState<OperationItem[]>([]);
	const [selectedKey, setSelectedKey] = useState("");
	const [activeKey, setActiveKey] = useState("");
	const [baseUrl, setBaseUrl] = useState("/api/v1");
	const [requestPath, setRequestPath] = useState("");
	const [token, setToken] = useState("");
	const [body, setBody] = useState("");
	const [response, setResponse] = useState("");
	const [sending, setSending] = useState(false);
	const [searchQuery, setSearchQuery] = useState("");
	const [openSidebarGroups, setOpenSidebarGroups] = useState<Set<string>>(() => new Set());
	const [openContentGroups, setOpenContentGroups] = useState<Set<string>>(() => new Set());
	const [loadError, setLoadError] = useState("");
	const deferredSearchQuery = useDeferredValue(searchQuery);

	useEffect(() => {
		let cancelled = false;

		fetch(specUrl)
			.then(result => result.json())
			.then((nextSpec: OpenAPISpec) => {
				if (cancelled) return;

				const nextOperations = collectOperations(nextSpec);
				const firstOperation = nextOperations[0]?.key ?? "";
				setSpec(nextSpec);
				setOperations(nextOperations);
				setBaseUrl(serverUrl(nextSpec));
				setSelectedKey(firstOperation);
				setActiveKey(firstOperation);
			})
			.catch(error => {
				if (!cancelled) setLoadError(`Failed to load OpenAPI spec: ${error instanceof Error ? error.message : String(error)}`);
			});

		return () => {
			cancelled = true;
		};
	}, [specUrl]);

	const selected = operations.find(operation => operation.key === selectedKey);
	const groups = groupOperations(operations);
	const normalizedQuery = deferredSearchQuery.trim().toLowerCase();
	const filteredOperations = normalizedQuery ? operations.filter(operation => operationSearchText(operation).includes(normalizedQuery)) : operations;
	const filteredGroups = groupOperations(filteredOperations);
	const selectedFields = requestFields(selected, spec);
	const selectedResponses = responseEntries(selected);
	const selectedCurl = curlCommand(selected, baseUrl, requestPath, body, token);

	useEffect(() => {
		setRequestPath(selected?.path ?? "");
		setBody(requestBodyExample(selected, spec));
		setResponse("");
	}, [selectedKey, spec]);

	useEffect(() => {
		if (!operations.length) return;

		let updateQueued = false;

		function updateActiveOperation() {
			updateQueued = false;
			let nextActiveKey = operations[0]?.key ?? "";

			for (const operation of operations) {
				const section = document.getElementById(operationAnchor(operation));
				if (!section || section.getClientRects().length === 0) continue;

				if (section.getBoundingClientRect().top <= 160) {
					nextActiveKey = operation.key;
				} else {
					break;
				}
			}

			setActiveKey(nextActiveKey);
		}

		function scheduleUpdate() {
			if (updateQueued) return;

			updateQueued = true;
			requestAnimationFrame(updateActiveOperation);
		}

		updateActiveOperation();
		window.addEventListener("scroll", scheduleUpdate, { passive: true });
		window.addEventListener("resize", scheduleUpdate);

		return () => {
			window.removeEventListener("scroll", scheduleUpdate);
			window.removeEventListener("resize", scheduleUpdate);
		};
	}, [operations]);

	function groupSetWith(previous: Set<string>, tag: string, open: boolean) {
		const next = new Set(previous);

		if (open) {
			next.add(tag);
		} else {
			next.delete(tag);
		}

		return next;
	}

	function setSidebarGroupOpen(tag: string, open: boolean) {
		setOpenSidebarGroups(previous => groupSetWith(previous, tag, open));
	}

	function setContentGroupOpen(tag: string, open: boolean) {
		setOpenContentGroups(previous => groupSetWith(previous, tag, open));
	}

	function selectOperation(operation: OperationItem) {
		setSelectedKey(operation.key);
		setActiveKey(operation.key);
		setContentGroupOpen(operation.tag, true);
	}

	function jumpToOperation(operation: OperationItem) {
		selectOperation(operation);

		requestAnimationFrame(() => {
			document.getElementById(operationAnchor(operation))?.scrollIntoView({ block: "start" });
			window.history.replaceState(null, "", `#${operationAnchor(operation)}`);
		});
	}

	async function sendRequest() {
		if (!selected) return;

		setSending(true);
		setResponse("");

		try {
			const headers: Record<string, string> = { Accept: "application/json" };

			if (token.trim()) {
				headers.Authorization = `Bearer ${token.trim()}`;
			}

			const init: RequestInit = { method: methodLabels[selected.method], headers };
			if (body.trim() && selected.method !== "get") {
				headers["Content-Type"] = "application/json";
				init.body = body;
			}

			const result = await fetch(requestUrl(baseUrl, requestPath), init);
			const text = await result.text();
			let formatted = text || "<empty response>";
			try {
				formatted = text ? JSON.stringify(JSON.parse(text), null, 2) : formatted;
			} catch {
				formatted = text;
			}
			setResponse(`${result.status} ${result.statusText}\n${formatted}`);
		} catch (error) {
			setResponse(error instanceof Error ? error.message : String(error));
		} finally {
			setSending(false);
		}
	}

	return (
		<div className={styles.referenceShell}>
			<aside className={styles.sidebar} aria-label="API reference navigation">
				<div className={styles.sidebarHeader}>
					<span>API Reference</span>
					<Badge tone="muted" dot={false}>
						OAS {spec?.openapi ?? "..."}
					</Badge>
				</div>

				<label className={styles.sidebarSearch}>
					<FieldLabel>Search</FieldLabel>
					<Input variant="compact" value={searchQuery} onChange={event => setSearchQuery(event.currentTarget.value)} placeholder="Find endpoint" />
				</label>

				<nav className={styles.sidebarNav}>
					<a href="#api-overview" className={styles.overviewLink}>
						Overview
					</a>
					{filteredGroups.map(group => (
						<details
							className={styles.navGroup}
							key={group.tag}
							open={normalizedQuery ? true : openSidebarGroups.has(group.tag)}
							onToggle={event => {
								if (!normalizedQuery) setSidebarGroupOpen(group.tag, event.currentTarget.open);
							}}
						>
							<summary className={styles.navGroupSummary}>
								<span className={styles.folderToggle} aria-hidden="true" />
								<span>{group.tag}</span>
								<small>{group.operations.length}</small>
							</summary>
							{group.operations.map(operation => (
								<a
									key={operation.key}
									href={`#${operationAnchor(operation)}`}
									className={classNames(styles.navOperation, activeKey === operation.key && styles.activeNavOperation)}
									onClick={event => {
										event.preventDefault();
										jumpToOperation(operation);
									}}
								>
									<span className={styles.method} data-method={operation.method}>
										{methodLabels[operation.method]}
									</span>
									<span>{operation.summary ?? operation.path}</span>
								</a>
							))}
						</details>
					))}
				</nav>

				<a className={styles.downloadLink} href={specUrl} download="netstamp-openapi.json">
					Download OpenAPI <span>JSON</span>
				</a>
			</aside>

			<main className={styles.content}>
				<section id="api-overview" className={styles.hero}>
					<div className={styles.heroCopy}>
						<div className={styles.metaRow}>
							<Badge tone="accent" dot={false}>
								{spec?.info?.version ?? "v1"}
							</Badge>
							<Badge tone="muted" dot={false}>
								OAS {spec?.openapi ?? "3.x"}
							</Badge>
						</div>
						<h1>{spec?.info?.title ?? "Netstamp API"}</h1>
						<p>{spec?.info?.description ?? "Generated reference for the Netstamp controller API."}</p>
						<a href={specUrl} download="netstamp-openapi.json" className={styles.inlineDownload}>
							Download OpenAPI document <span>json</span>
						</a>
					</div>

					<Panel title="Server" eyebrow="request target" tone="deep" className={styles.serverPanel}>
						<Input value={baseUrl} onChange={event => setBaseUrl(event.currentTarget.value)} aria-label="API server URL" />
						<p>Used by the sticky request console. Point it at a local or deployed backend before testing.</p>
					</Panel>
				</section>

				{loadError ? (
					<Panel title="Spec load failed" tone="deep" className={styles.errorPanel}>
						<p>{loadError}</p>
					</Panel>
				) : null}

				{groups.map(group => (
					<details
						className={styles.tagSection}
						id={tagAnchor(group.tag)}
						key={group.tag}
						open={normalizedQuery ? true : openContentGroups.has(group.tag)}
						onToggle={event => {
							if (!normalizedQuery) setContentGroupOpen(group.tag, event.currentTarget.open);
						}}
					>
						<summary className={styles.tagSummary}>
							<span className={styles.folderToggle} aria-hidden="true" />
							<span>
								<h2>{group.tag}</h2>
								<p>
									{group.operations.length} operation{group.operations.length === 1 ? "" : "s"} in this API area.
								</p>
							</span>
							<Badge tone="muted" dot={false}>
								Folder
							</Badge>
						</summary>

						<div className={styles.sectionBody}>
							<Panel title="Operations" tone="deep" padded={false} className={styles.operationsPanel}>
								{group.operations.map(operation => (
									<a
										href={`#${operationAnchor(operation)}`}
										key={operation.key}
										onClick={event => {
											event.preventDefault();
											jumpToOperation(operation);
										}}
									>
										<span className={styles.method} data-method={operation.method}>
											{methodLabels[operation.method]}
										</span>
										<code>{operation.path}</code>
									</a>
								))}
							</Panel>

							<div className={styles.operationStack}>
								{group.operations.map(operation => {
									const fields = requestFields(operation, spec);
									const responses = responseEntries(operation);

									return (
										<article className={classNames(styles.operationArticle, activeKey === operation.key && styles.activeOperationArticle)} id={operationAnchor(operation)} key={operation.key}>
											<div className={styles.operationCopy}>
												<div className={styles.operationMeta}>
													<Badge tone={methodTones[operation.method]} dot={false}>
														{methodLabels[operation.method]}
													</Badge>
													{operation.operationId ? <span>{operation.operationId}</span> : null}
												</div>
												<h3>{operation.summary ?? operation.path}</h3>
												<p>{operation.description ?? `Endpoint path: ${operation.path}`}</p>

												{operation.parameters?.length ? (
													<div className={styles.detailBlock}>
														<h4>Parameters</h4>
														{operation.parameters.map(parameter => (
															<div className={styles.fieldRow} key={`${parameter.in}:${parameter.name}`}>
																<code>{parameter.name}</code>
																<span>{schemaType(parameter.schema, spec)}</span>
																<small>{parameter.required ? "required" : parameter.in}</small>
																<p>{parameter.description ?? "No description provided."}</p>
															</div>
														))}
													</div>
												) : null}

												{fields.length ? (
													<div className={styles.detailBlock}>
														<h4>Request body</h4>
														{fields.map(field => (
															<div className={styles.fieldRow} key={field.name}>
																<code>{field.name}</code>
																<span>{field.type}</span>
																<small>{field.required ? "required" : "optional"}</small>
																<p>{field.description ?? "No description provided."}</p>
															</div>
														))}
													</div>
												) : null}

												<div className={styles.detailBlock}>
													<h4>Responses</h4>
													{responses.map(([status, responseValue]) => (
														<div className={styles.responseRow} key={status}>
															<code className={classNames(styles.statusCode, statusTone(status))}>{status}</code>
															<span>{responseValue.description ?? "Response"}</span>
														</div>
													))}
												</div>
											</div>

											<Panel tone="deep" padded={false} className={styles.snippetPanel}>
												<div className={styles.snippetHeader}>
													<span className={styles.method} data-method={operation.method}>
														{methodLabels[operation.method]}
													</span>
													<code>{operation.path}</code>
												</div>
												<pre>
													<code>{curlCommand(operation, baseUrl, operation.path, requestBodyExample(operation, spec), token)}</code>
												</pre>
												<button type="button" onClick={() => selectOperation(operation)}>
													Load in console
												</button>
											</Panel>
										</article>
									);
								})}
							</div>
						</div>
					</details>
				))}
			</main>

			<aside className={styles.console} aria-label="Request console">
				<Panel title="Request Console" eyebrow={selected ? `${methodLabels[selected.method]} ${selected.path}` : "select operation"} tone="deep" className={styles.consolePanel}>
					<div className={styles.formGrid}>
						<label>
							<FieldLabel>Server</FieldLabel>
							<Input value={baseUrl} onChange={event => setBaseUrl(event.currentTarget.value)} />
						</label>
						<label>
							<FieldLabel>Path</FieldLabel>
							<Input value={requestPath} onChange={event => setRequestPath(event.currentTarget.value)} />
						</label>
						<label className={styles.fullWidth}>
							<FieldLabel>Bearer token</FieldLabel>
							<Input value={token} onChange={event => setToken(event.currentTarget.value)} placeholder="Optional access token" />
						</label>
					</div>

					{requestContent(selected) ? (
						<label className={styles.bodyField}>
							<FieldLabel>JSON body</FieldLabel>
							<textarea value={body} onChange={event => setBody(event.currentTarget.value)} spellCheck={false} />
						</label>
					) : (
						<div className={styles.emptyBody}>No JSON request body for this operation.</div>
					)}

					{selectedFields.length ? (
						<div className={styles.consoleFields}>
							{selectedFields.map(field => (
								<span key={field.name}>
									<code>{field.name}</code> {field.required ? "required" : "optional"}
								</span>
							))}
						</div>
					) : null}

					<pre className={styles.curlPreview}>
						<code>{selectedCurl}</code>
					</pre>

					<div className={styles.actions}>
						<Button type="button" size="sm" onClick={sendRequest} disabled={!selected || sending}>
							{sending ? "Sending" : "Test Request"}
						</Button>
					</div>

					<div className={styles.consoleResponses}>
						{selectedResponses.map(([status]) => (
							<span className={classNames(styles.statusCode, statusTone(status))} key={status}>
								{status}
							</span>
						))}
					</div>

					<pre className={styles.response} aria-live="polite">
						<code>{response || "Response will appear here."}</code>
					</pre>
				</Panel>
			</aside>
		</div>
	);
}
