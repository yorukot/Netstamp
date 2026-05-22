import type { ApiCheck, CreateCheckInput } from "@/shared/api/types";

export type PingConfigPayload = NonNullable<CreateCheckInput["pingConfig"]>;
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

function parseIntOrDefault(value: string, fallback: string) {
	const parsed = Number.parseInt(value, 10);
	const defaultValue = Number.parseInt(fallback, 10);

	return Number.isFinite(parsed) ? parsed : defaultValue;
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
	const config: PingConfigPayload = {
		packetCount: parseIntOrDefault(state.packetCount, defaultPingConfigFormState.packetCount),
		packetSizeBytes: parseIntOrDefault(state.packetSizeBytes, defaultPingConfigFormState.packetSizeBytes),
		timeoutMs: parseIntOrDefault(state.timeoutMs, defaultPingConfigFormState.timeoutMs)
	};

	if (state.ipFamily) {
		config.ipFamily = state.ipFamily;
	}

	return config;
}

export function buildTracerouteConfigPayload(state: TracerouteConfigFormState): TracerouteConfigPayload {
	const config: TracerouteConfigPayload = {
		protocol: state.protocol,
		maxHops: parseIntOrDefault(state.maxHops, defaultTracerouteConfigFormState.maxHops),
		timeoutMs: parseIntOrDefault(state.timeoutMs, defaultTracerouteConfigFormState.timeoutMs),
		queriesPerHop: parseIntOrDefault(state.queriesPerHop, defaultTracerouteConfigFormState.queriesPerHop),
		packetSizeBytes: parseIntOrDefault(state.packetSizeBytes, defaultTracerouteConfigFormState.packetSizeBytes),
		port: parseIntOrDefault(state.port, defaultTracerouteConfigFormState.port)
	};

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

	const config = pingConfigFormStateFromApi(check);

	return [
		["Packets", config.packetCount],
		["Packet size", `${config.packetSizeBytes} bytes`],
		["Timeout", `${config.timeoutMs}ms`]
	];
}
