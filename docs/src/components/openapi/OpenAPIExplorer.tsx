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
	security?: Array<Record<string, string[]>>;
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
	type?: string | string[];
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

interface HighlightToken {
	value: string;
	className?: string;
}

interface OpenAPIExplorerProps {
	specUrl: string;
}

type CodeLanguage = "shell" | "json" | "response";

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

const jsonTokenPattern = /("(?:\\.|[^"\\])*"\s*:)|("(?:\\.|[^"\\])*")|\b(?:true|false)\b|\bnull\b|-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?|[{}\[\]:,]/g;
const shellTokenPattern = /\bcurl\b|--?[A-Za-z0-9-]+|'(?:[^']|'"'"')*'|"(?:\\.|[^"\\])*"|\\\n|\b(?:GET|POST|PUT|PATCH|DELETE)\b/g;
const numericTokenPattern = /^-?\d/;
const shellMethodNames = new Set(["GET", "POST", "PUT", "PATCH", "DELETE"]);

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

function schemaHasType(schema: JSONSchema | undefined, type: string) {
	const schemaTypeValue = schema?.type;
	return Array.isArray(schemaTypeValue) ? schemaTypeValue.includes(type) : schemaTypeValue === type;
}

function valueFromSchema(schema: JSONSchema | undefined, spec: OpenAPISpec | undefined, seen = new Set<string>()): unknown {
	schema = resolveSchema(schema, spec, seen);
	if (!schema) return undefined;
	if (schema.example !== undefined) return schema.example;
	if (schema.examples?.[0] !== undefined) return schema.examples[0];
	if (schema.default !== undefined) return schema.default;
	if (schema.enum?.[0] !== undefined) return schema.enum[0];

	if (schemaHasType(schema, "object") && schema.properties) {
		const value: Record<string, unknown> = {};
		Object.entries(schema.properties).forEach(([key, property]) => {
			if (property.readOnly) return;
			value[key] = valueFromSchema(property, spec, new Set(seen)) ?? "";
		});
		return value;
	}

	if (schemaHasType(schema, "array")) return schema.items ? [valueFromSchema(schema.items, spec, new Set(seen))] : [];
	if (schemaHasType(schema, "number") || schemaHasType(schema, "integer")) return 0;
	if (schemaHasType(schema, "boolean")) return false;
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
	const schemaTypeValue = resolved?.type;
	const typeNames = (Array.isArray(schemaTypeValue) ? schemaTypeValue : schemaTypeValue ? [schemaTypeValue] : []).filter(Boolean);
	if (typeNames.includes("array")) {
		return `${schemaType(resolved?.items, spec)}[]${typeNames.includes("null") ? " | null" : ""}`;
	}

	const baseType = typeNames.filter(type => type !== "null").join(" | ") || "value";
	const formattedType = resolved?.format && baseType !== "value" ? `${baseType}/${resolved.format}` : baseType;
	return typeNames.includes("null") && baseType !== "null" ? `${formattedType} | null` : formattedType;
}

function schemaFields(schema: JSONSchema | undefined, spec: OpenAPISpec | undefined, includeReadOnly: boolean): SchemaField[] {
	const resolved = resolveSchema(schema, spec, new Set());
	const objectSchema = schemaHasType(resolved, "array") ? resolveSchema(resolved?.items, spec, new Set()) : resolved;
	if (!objectSchema?.properties) return [];

	const requiredFields = new Set(objectSchema.required ?? []);
	return Object.entries(objectSchema.properties)
		.filter(([, property]) => includeReadOnly || !property.readOnly)
		.map(([name, property]) => ({
			name,
			type: schemaType(property, spec),
			description: property.description,
			required: requiredFields.has(name)
		}));
}

function requestFields(operation: OperationItem | undefined, spec: OpenAPISpec | undefined): SchemaField[] {
	const schema = requestContent(operation)?.schema;
	return schemaFields(schema, spec, false);
}

function responseFields(schema: JSONSchema | undefined, spec: OpenAPISpec | undefined): SchemaField[] {
	return schemaFields(schema, spec, true);
}

function responseContentEntries(response: OpenAPIResponse | undefined) {
	return Object.entries(response?.content ?? {});
}

function responseSchemaLabel(response: OpenAPIResponse, spec: OpenAPISpec | undefined) {
	const content = responseContentEntries(response)[0]?.[1];
	return content?.schema ? schemaType(content.schema, spec) : "no body";
}

function ResponseSchemaDetails({ response, spec }: { response: OpenAPIResponse; spec: OpenAPISpec | undefined }) {
	const contentEntries = responseContentEntries(response);
	if (!contentEntries.length) {
		return <div className={styles.emptySchema}>No response body.</div>;
	}

	return (
		<div className={styles.responseSchemaStack}>
			{contentEntries.map(([contentType, mediaType]) => {
				const fields = responseFields(mediaType.schema, spec);

				return (
					<div className={styles.responseSchema} key={contentType}>
						<div className={styles.responseSchemaHeader}>
							<code>{contentType}</code>
							<span>{schemaType(mediaType.schema, spec)}</span>
						</div>
						{fields.length ? (
							<div className={styles.responseFields}>
								{fields.map(field => (
									<div className={styles.responseField} key={field.name}>
										<code>{field.name}</code>
										<span>{field.type}</span>
										<small>{field.required ? "required" : "optional"}</small>
										<p>{field.description ?? "No description provided."}</p>
									</div>
								))}
							</div>
						) : (
							<p className={styles.emptySchema}>No fields documented for this response body.</p>
						)}
					</div>
				);
			})}
		</div>
	);
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

function tokenizeWithPattern(source: string, pattern: RegExp, classForMatch: (value: string, match: RegExpExecArray) => string | undefined) {
	const tokens: HighlightToken[] = [];
	let cursor = 0;
	let match: RegExpExecArray | null;

	pattern.lastIndex = 0;
	while ((match = pattern.exec(source))) {
		if (match.index > cursor) {
			tokens.push({ value: source.slice(cursor, match.index) });
		}

		tokens.push({ value: match[0], className: classForMatch(match[0], match) });
		cursor = match.index + match[0].length;
		if (!match[0].length) pattern.lastIndex += 1;
	}

	if (cursor < source.length) {
		tokens.push({ value: source.slice(cursor) });
	}

	return tokens;
}

function tokenizeJson(source: string) {
	return tokenizeWithPattern(source, jsonTokenPattern, (value, match) => {
		if (match[1]) return styles.syntaxJsonKey;
		if (match[2]) return styles.syntaxString;
		if (value === "true" || value === "false") return styles.syntaxBoolean;
		if (value === "null") return styles.syntaxNull;
		if (numericTokenPattern.test(value)) return styles.syntaxNumber;
		return styles.syntaxPunctuation;
	});
}

function tokenizeShell(source: string) {
	return tokenizeWithPattern(source, shellTokenPattern, value => {
		if (value === "curl") return styles.syntaxCommand;
		if (value.startsWith("-")) return styles.syntaxOption;
		if (value.startsWith("'") || value.startsWith('"')) return styles.syntaxString;
		if (shellMethodNames.has(value)) return styles.syntaxHttpMethod;
		return styles.syntaxPunctuation;
	});
}

function tokenizeResponse(source: string) {
	const statusLine = source.match(/^(\d{3}[^\n]*)(\n)?/);
	if (statusLine) {
		return [{ value: statusLine[1], className: styles.syntaxMeta }, ...(statusLine[2] ? [{ value: statusLine[2] }] : []), ...tokenizeMaybeJson(source.slice(statusLine[0].length))];
	}

	return tokenizeMaybeJson(source);
}

function tokenizeMaybeJson(source: string) {
	const trimmed = source.trimStart();
	return trimmed.startsWith("{") || trimmed.startsWith("[") ? tokenizeJson(source) : [{ value: source }];
}

function highlightedTokens(source: string, language: CodeLanguage) {
	if (language === "shell") return tokenizeShell(source);
	if (language === "json") return tokenizeJson(source);
	return tokenizeResponse(source);
}

function SyntaxCode({ code, language }: { code: string; language: CodeLanguage }) {
	return (
		<code className={styles.syntaxCode}>
			{highlightedTokens(code, language).map((token, index) =>
				token.className ? (
					<span className={token.className} key={index}>
						{token.value}
					</span>
				) : (
					token.value
				)
			)}
		</code>
	);
}

function requestUrl(baseUrl: string, path: string) {
	const trimmedBase = baseUrl.replace(/\/$/, "");
	const requestPath = path.startsWith("/") ? path : `/${path}`;
	return `${trimmedBase}${requestPath}`;
}

function usesSessionCookie(operation: OperationItem | undefined) {
	return operation?.security?.some(requirement => Object.prototype.hasOwnProperty.call(requirement, "sessionCookieAuth")) ?? false;
}

function writesSessionCookie(operation: OperationItem | undefined) {
	return operation?.path === "/auth/login" || operation?.path === "/auth/register" || operation?.path === "/auth/logout";
}

function curlCommand(operation: OperationItem | undefined, baseUrl: string, path: string, body: string) {
	if (!operation) return "Select an operation to generate a request.";

	const lines = [`curl -X ${methodLabels[operation.method]} '${shellEscape(requestUrl(baseUrl, path))}'`, "  -H 'Accept: application/json'"];

	if (usesSessionCookie(operation) || operation.path === "/auth/logout") {
		lines.push("  --cookie netstamp.cookies");
	}
	if (usesSessionCookie(operation) || writesSessionCookie(operation)) {
		lines.push("  --cookie-jar netstamp.cookies");
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
	const selectedCurl = curlCommand(selected, baseUrl, requestPath, body);

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

			const init: RequestInit = { method: methodLabels[selected.method], headers, credentials: "include" };
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

					<Panel title="Server" tone="deep" className={styles.serverPanel}>
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
														<details className={styles.responseDetails} key={status} open={status === "200"}>
															<summary className={styles.responseSummary}>
																<span className={styles.responseToggle} aria-hidden="true" />
																<code className={classNames(styles.statusCode, statusTone(status))}>{status}</code>
																<span className={styles.responseDescription}>{responseValue.description ?? "Response"}</span>
																<small>{responseSchemaLabel(responseValue, spec)}</small>
															</summary>
															<ResponseSchemaDetails response={responseValue} spec={spec} />
														</details>
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
													<SyntaxCode code={curlCommand(operation, baseUrl, operation.path, requestBodyExample(operation, spec))} language="shell" />
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
				<Panel title="Request Console" tone="deep" className={styles.consolePanel}>
					<div className={styles.formGrid}>
						<label>
							<FieldLabel>Server</FieldLabel>
							<Input value={baseUrl} onChange={event => setBaseUrl(event.currentTarget.value)} />
						</label>
						<label>
							<FieldLabel>Path</FieldLabel>
							<Input value={requestPath} onChange={event => setRequestPath(event.currentTarget.value)} />
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
						<SyntaxCode code={selectedCurl} language="shell" />
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
						<SyntaxCode code={response || "Response will appear here."} language="response" />
					</pre>
				</Panel>
			</aside>
		</div>
	);
}
