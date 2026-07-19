import { i18n } from "@/i18n";
import type { ApiCheck, CreateCheckInput } from "@/shared/api/types";

const checkT = i18n.getFixedT(null, "checks") as (key: string, options?: Record<string, unknown>) => string;

export type PingConfigPayload = NonNullable<CreateCheckInput["pingConfig"]>;
export type TCPConfigPayload = NonNullable<CreateCheckInput["tcpConfig"]>;
export type TracerouteConfigPayload = NonNullable<CreateCheckInput["tracerouteConfig"]>;
export type HTTPConfigPayload = NonNullable<CreateCheckInput["httpConfig"]>;
export type IPFamilyFormValue = "" | NonNullable<PingConfigPayload["ipFamily"]>;
export type TracerouteProtocolFormValue = NonNullable<TracerouteConfigPayload["protocol"]>;

export interface PingConfigFormState {
	packetCount: string;
	packetSizeBytes: string;
	timeoutMs: string;
	ipFamily: IPFamilyFormValue;
}

export interface TracerouteConfigFormState {
	protocol: TracerouteProtocolFormValue;
	maxHops: string;
	timeoutMs: string;
	queriesPerHop: string;
	packetSizeBytes: string;
	port: string;
	ipFamily: IPFamilyFormValue;
}

export interface TCPConfigFormState {
	port: string;
	timeoutMs: string;
	ipFamily: IPFamilyFormValue;
}

export type HTTPStatusClass = "1xx" | "2xx" | "3xx" | "4xx" | "5xx";
export interface HTTPHeaderFormState {
	id: string;
	name: string;
	value: string;
}
export interface HTTPConfigFormState {
	method: NonNullable<HTTPConfigPayload["method"]>;
	headers: HTTPHeaderFormState[];
	body: string;
	timeoutMs: string;
	ipFamily: IPFamilyFormValue;
	followRedirects: boolean;
	skipTlsVerify: boolean;
	statusClasses: HTTPStatusClass[];
	statusCodes: string;
	bodyContains: string;
}

export interface NumericConfigValidation {
	value: number;
	error: string;
}

export type PingConfigValidation = Record<keyof Pick<PingConfigFormState, "packetCount" | "packetSizeBytes" | "timeoutMs">, NumericConfigValidation>;
export type TCPConfigValidation = Record<keyof Pick<TCPConfigFormState, "port" | "timeoutMs">, NumericConfigValidation>;
export type TracerouteConfigValidation = Record<keyof Pick<TracerouteConfigFormState, "maxHops" | "timeoutMs" | "queriesPerHop" | "packetSizeBytes" | "port">, NumericConfigValidation>;
export interface HTTPConfigValidation {
	timeout: NumericConfigValidation;
	codes: number[];
	statusError: string;
	bodyError: string;
	error: string;
}

export interface CheckConfigValidation {
	ping: PingConfigValidation;
	tcp: TCPConfigValidation;
	traceroute: TracerouteConfigValidation;
}

export const defaultPingConfigFormState: PingConfigFormState = {
	packetCount: "4",
	packetSizeBytes: "56",
	timeoutMs: "3000",
	ipFamily: ""
};

export const defaultTracerouteConfigFormState: TracerouteConfigFormState = {
	protocol: "icmp",
	maxHops: "30",
	timeoutMs: "3000",
	queriesPerHop: "3",
	packetSizeBytes: "56",
	port: "33434",
	ipFamily: ""
};

export const defaultTCPConfigFormState: TCPConfigFormState = {
	port: "443",
	timeoutMs: "3000",
	ipFamily: ""
};

export const defaultHTTPConfigFormState: HTTPConfigFormState = {
	method: "GET",
	headers: [],
	body: "",
	timeoutMs: "10000",
	ipFamily: "",
	followRedirects: true,
	skipTlsVerify: false,
	statusClasses: ["2xx", "3xx"],
	statusCodes: "",
	bodyContains: ""
};

function validateIntegerField(label: string, value: string, options: { min: number; max?: number }): NumericConfigValidation {
	const trimmed = value.trim();

	if (!trimmed) {
		return { value: Number.NaN, error: checkT("config.required", { label }) };
	}

	if (!/^\d+$/.test(trimmed)) {
		return { value: Number.NaN, error: checkT("config.wholeNumber", { label }) };
	}

	const parsed = Number.parseInt(trimmed, 10);

	if (!Number.isFinite(parsed)) {
		return { value: Number.NaN, error: checkT("config.number", { label }) };
	}

	if (parsed < options.min) {
		return { value: parsed, error: checkT("config.minimum", { label, min: options.min }) };
	}

	if (typeof options.max === "number" && parsed > options.max) {
		return { value: parsed, error: checkT("config.maximum", { label, max: options.max }) };
	}

	return { value: parsed, error: "" };
}

export function validatePingConfig(state: PingConfigFormState): PingConfigValidation {
	return {
		packetCount: validateIntegerField(checkT("config.packetCount"), state.packetCount, { min: 1, max: 10000 }),
		packetSizeBytes: validateIntegerField(checkT("config.packetSizeBytes"), state.packetSizeBytes, { min: 1, max: 65507 }),
		timeoutMs: validateIntegerField(checkT("config.timeoutMs"), state.timeoutMs, { min: 1, max: 60000 })
	};
}

export function validateTCPConfig(state: TCPConfigFormState): TCPConfigValidation {
	return {
		port: validateIntegerField(checkT("config.port"), state.port, { min: 1, max: 65535 }),
		timeoutMs: validateIntegerField(checkT("config.timeoutMs"), state.timeoutMs, { min: 1, max: 60000 })
	};
}

export function validateTracerouteConfig(state: TracerouteConfigFormState): TracerouteConfigValidation {
	return {
		maxHops: validateIntegerField(checkT("config.maxHops"), state.maxHops, { min: 1, max: 64 }),
		timeoutMs: validateIntegerField(checkT("config.timeoutMs"), state.timeoutMs, { min: 1, max: 60000 }),
		queriesPerHop: validateIntegerField(checkT("config.queriesPerHop"), state.queriesPerHop, { min: 1, max: 10 }),
		packetSizeBytes: validateIntegerField(checkT("config.packetSizeBytes"), state.packetSizeBytes, { min: 1, max: 65507 }),
		port: validateIntegerField(checkT("config.port"), state.port, { min: 1, max: 65535 })
	};
}

export function firstConfigValidationError(validation: Record<string, NumericConfigValidation>) {
	return Object.values(validation).find(field => field.error)?.error ?? "";
}

export function pingConfigFormStateFromApi(check: Pick<ApiCheck, "pingConfig"> | null | undefined): PingConfigFormState {
	const config = check?.pingConfig;

	return {
		packetCount: String(config?.packetCount ?? defaultPingConfigFormState.packetCount),
		packetSizeBytes: String(config?.packetSizeBytes ?? defaultPingConfigFormState.packetSizeBytes),
		timeoutMs: String(config?.timeoutMs ?? defaultPingConfigFormState.timeoutMs),
		ipFamily: config?.ipFamily ?? defaultPingConfigFormState.ipFamily
	};
}

export function tcpConfigFormStateFromApi(check: Pick<ApiCheck, "tcpConfig"> | null | undefined): TCPConfigFormState {
	const config = check?.tcpConfig;

	return {
		port: String(config?.port ?? defaultTCPConfigFormState.port),
		timeoutMs: String(config?.timeoutMs ?? defaultTCPConfigFormState.timeoutMs),
		ipFamily: config?.ipFamily ?? defaultTCPConfigFormState.ipFamily
	};
}

export function tracerouteConfigFormStateFromApi(check: Pick<ApiCheck, "tracerouteConfig"> | null | undefined): TracerouteConfigFormState {
	const config = check?.tracerouteConfig;

	return {
		protocol: config?.protocol ?? defaultTracerouteConfigFormState.protocol,
		maxHops: String(config?.maxHops ?? defaultTracerouteConfigFormState.maxHops),
		timeoutMs: String(config?.timeoutMs ?? defaultTracerouteConfigFormState.timeoutMs),
		queriesPerHop: String(config?.queriesPerHop ?? defaultTracerouteConfigFormState.queriesPerHop),
		packetSizeBytes: String(config?.packetSizeBytes ?? defaultTracerouteConfigFormState.packetSizeBytes),
		port: String(config?.port ?? defaultTracerouteConfigFormState.port),
		ipFamily: config?.ipFamily ?? defaultTracerouteConfigFormState.ipFamily
	};
}

export function httpConfigFormStateFromApi(check: Pick<ApiCheck, "httpConfig"> | null | undefined): HTTPConfigFormState {
	const config = check?.httpConfig;
	const statuses = config?.expectedStatuses ?? [];
	return {
		method: config?.method ?? "GET",
		headers: (config?.headers ?? []).map((header, index) => ({ id: `header-${index}-${header.name}`, ...header })),
		body: config?.body ?? "",
		timeoutMs: String(config?.timeoutMs ?? 10000),
		ipFamily: config?.ipFamily ?? "",
		followRedirects: config?.followRedirects ?? true,
		skipTlsVerify: config?.skipTlsVerify ?? false,
		statusClasses: statuses.filter((status): status is Extract<(typeof statuses)[number], { kind: "class" }> => status.kind === "class").map(status => status.class),
		statusCodes: statuses
			.filter((status): status is Extract<(typeof statuses)[number], { kind: "code" }> => status.kind === "code")
			.map(status => status.code)
			.join(", "),
		bodyContains: config?.bodyContains ?? ""
	};
}

export function validateHTTPConfig(state: HTTPConfigFormState): HTTPConfigValidation {
	const timeout = validateIntegerField(checkT("config.timeoutMs"), state.timeoutMs, { min: 1, max: 60000 });
	const codes: number[] = [];
	let statusError = "";
	for (const value of state.statusCodes
		.split(",")
		.map(value => value.trim())
		.filter(Boolean)) {
		if (!/^\d{3}$/.test(value) || Number(value) < 100 || Number(value) > 599) {
			statusError = checkT("config.statusRange");
			break;
		}
		codes.push(Number(value));
	}
	if (!statusError && state.statusClasses.length === 0 && codes.length === 0) statusError = checkT("config.expectedStatusRequired");
	const bodyError = (state.method === "GET" || state.method === "HEAD") && state.body ? checkT("config.bodyNotAllowed", { method: state.method }) : "";
	return { timeout, codes: Array.from(new Set(codes)).sort((a, b) => a - b), statusError, bodyError, error: timeout.error || statusError || bodyError };
}

export function buildHTTPConfigPayload(state: HTTPConfigFormState): HTTPConfigPayload {
	const validation = validateHTTPConfig(state);
	if (validation.error) throw new Error(validation.error);
	const payload: HTTPConfigPayload = {
		method: state.method,
		headers: state.headers.filter(header => header.name.trim()).map(({ name, value }) => ({ name: name.trim(), value })),
		timeoutMs: validation.timeout.value,
		followRedirects: state.followRedirects,
		skipTlsVerify: state.skipTlsVerify,
		expectedStatuses: [...state.statusClasses.map(statusClass => ({ kind: "class" as const, class: statusClass })), ...validation.codes.map(code => ({ kind: "code" as const, code }))]
	};
	payload.ipFamily = state.ipFamily || null;
	if (state.method !== "GET" && state.method !== "HEAD") payload.body = state.body;
	payload.bodyContains = state.bodyContains;
	return payload;
}

export function buildPingConfigPayload(state: PingConfigFormState): PingConfigPayload {
	const validation = validatePingConfig(state);
	const validationError = firstConfigValidationError(validation);

	if (validationError) {
		throw new Error(validationError);
	}

	const config: PingConfigPayload = {
		packetCount: validation.packetCount.value,
		packetSizeBytes: validation.packetSizeBytes.value,
		timeoutMs: validation.timeoutMs.value
	};

	if (state.ipFamily) {
		config.ipFamily = state.ipFamily;
	}

	return config;
}

export function buildTCPConfigPayload(state: TCPConfigFormState): TCPConfigPayload {
	const validation = validateTCPConfig(state);
	const validationError = firstConfigValidationError(validation);

	if (validationError) {
		throw new Error(validationError);
	}

	const config: TCPConfigPayload = {
		port: validation.port.value,
		timeoutMs: validation.timeoutMs.value
	};

	if (state.ipFamily) {
		config.ipFamily = state.ipFamily;
	}

	return config;
}

export function buildTracerouteConfigPayload(state: TracerouteConfigFormState): TracerouteConfigPayload {
	const validation = validateTracerouteConfig(state);
	const validationError = firstConfigValidationError(state.protocol === "udp" ? validation : { ...validation, port: { value: 1, error: "" } });

	if (validationError) {
		throw new Error(validationError);
	}

	const config: TracerouteConfigPayload = {
		protocol: state.protocol,
		maxHops: validation.maxHops.value,
		timeoutMs: validation.timeoutMs.value,
		queriesPerHop: validation.queriesPerHop.value,
		packetSizeBytes: validation.packetSizeBytes.value
	};

	if (state.protocol === "udp") {
		config.port = validation.port.value;
	}

	if (state.ipFamily) {
		config.ipFamily = state.ipFamily;
	}

	return config;
}

export function checkConfigSummaryFields(check: ApiCheck): Array<[label: string, value: string]> {
	if (check.type === "traceroute") {
		const config = tracerouteConfigFormStateFromApi(check);

		return [
			[checkT("summary.protocol"), config.protocol],
			[checkT("summary.maxHops"), config.maxHops],
			[checkT("summary.queriesPerHop"), config.queriesPerHop],
			[checkT("summary.timeout"), `${config.timeoutMs}ms`]
		];
	}

	if (check.type === "tcp") {
		const config = tcpConfigFormStateFromApi(check);

		return [
			[checkT("summary.port"), config.port],
			[checkT("summary.timeout"), `${config.timeoutMs}ms`],
			[checkT("summary.ipFamily"), config.ipFamily || checkT("config.auto")]
		];
	}
	if (check.type === "http") {
		const config = httpConfigFormStateFromApi(check);
		return [
			[checkT("summary.method"), config.method],
			[checkT("summary.expected"), [...config.statusClasses, config.statusCodes].filter(Boolean).join(", ")],
			[checkT("summary.timeout"), `${config.timeoutMs}ms`]
		];
	}

	const config = pingConfigFormStateFromApi(check);

	return [
		[checkT("summary.packets"), config.packetCount],
		[checkT("summary.packetSize"), checkT("summary.bytes", { count: Number(config.packetSizeBytes) })],
		[checkT("summary.timeout"), `${config.timeoutMs}ms`]
	];
}
