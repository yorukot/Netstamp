import type { ApiCheck, CreateCheckInput } from "@/shared/api/types";

export type PingConfigPayload = NonNullable<CreateCheckInput["pingConfig"]>;
export type TCPConfigPayload = NonNullable<CreateCheckInput["tcpConfig"]>;
export type TracerouteConfigPayload = NonNullable<CreateCheckInput["tracerouteConfig"]>;
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

export interface NumericConfigValidation {
	value: number;
	error: string;
}

export type PingConfigValidation = Record<keyof Pick<PingConfigFormState, "packetCount" | "packetSizeBytes" | "timeoutMs">, NumericConfigValidation>;
export type TCPConfigValidation = Record<keyof Pick<TCPConfigFormState, "port" | "timeoutMs">, NumericConfigValidation>;
export type TracerouteConfigValidation = Record<keyof Pick<TracerouteConfigFormState, "maxHops" | "timeoutMs" | "queriesPerHop" | "packetSizeBytes" | "port">, NumericConfigValidation>;

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

function validateIntegerField(label: string, value: string, options: { min: number; max?: number }): NumericConfigValidation {
	const trimmed = value.trim();

	if (!trimmed) {
		return { value: Number.NaN, error: `${label} is required.` };
	}

	if (!/^\d+$/.test(trimmed)) {
		return { value: Number.NaN, error: `${label} must be a whole number.` };
	}

	const parsed = Number.parseInt(trimmed, 10);

	if (!Number.isFinite(parsed)) {
		return { value: Number.NaN, error: `${label} must be a number.` };
	}

	if (parsed < options.min) {
		return { value: parsed, error: `${label} must be at least ${options.min}.` };
	}

	if (typeof options.max === "number" && parsed > options.max) {
		return { value: parsed, error: `${label} must be at most ${options.max}.` };
	}

	return { value: parsed, error: "" };
}

export function validatePingConfig(state: PingConfigFormState): PingConfigValidation {
	return {
		packetCount: validateIntegerField("Packet count", state.packetCount, { min: 1, max: 10000 }),
		packetSizeBytes: validateIntegerField("Packet size bytes", state.packetSizeBytes, { min: 1, max: 65507 }),
		timeoutMs: validateIntegerField("Timeout ms", state.timeoutMs, { min: 1, max: 60000 })
	};
}

export function validateTCPConfig(state: TCPConfigFormState): TCPConfigValidation {
	return {
		port: validateIntegerField("Port", state.port, { min: 1, max: 65535 }),
		timeoutMs: validateIntegerField("Timeout ms", state.timeoutMs, { min: 1, max: 60000 })
	};
}

export function validateTracerouteConfig(state: TracerouteConfigFormState): TracerouteConfigValidation {
	return {
		maxHops: validateIntegerField("Max hops", state.maxHops, { min: 1, max: 64 }),
		timeoutMs: validateIntegerField("Timeout ms", state.timeoutMs, { min: 1, max: 60000 }),
		queriesPerHop: validateIntegerField("Queries per hop", state.queriesPerHop, { min: 1, max: 10 }),
		packetSizeBytes: validateIntegerField("Packet size bytes", state.packetSizeBytes, { min: 1, max: 65507 }),
		port: validateIntegerField("Port", state.port, { min: 1, max: 65535 })
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
			["Protocol", config.protocol],
			["Max hops", config.maxHops],
			["Queries/hop", config.queriesPerHop],
			["Timeout", `${config.timeoutMs}ms`]
		];
	}

	if (check.type === "tcp") {
		const config = tcpConfigFormStateFromApi(check);

		return [
			["Port", config.port],
			["Timeout", `${config.timeoutMs}ms`],
			["IP family", config.ipFamily || "Auto"]
		];
	}

	const config = pingConfigFormStateFromApi(check);

	return [
		["Packets", config.packetCount],
		["Packet size", `${config.packetSizeBytes} bytes`],
		["Timeout", `${config.timeoutMs}ms`]
	];
}
