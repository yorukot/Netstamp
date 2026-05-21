namespace Netstamp;

alias CheckType = "ping" | "traceroute";
alias IPFamily = "inet" | "inet6";
alias PingStatus = "successful" | "timeout" | "error";
alias TracerouteProtocol = "icmp" | "udp";
alias TracerouteStatus = "successful" | "timeout" | "error" | "partial";

@example(#{ packetCount: 4, packetSizeBytes: 56, timeoutMs: 3000, ipFamily: "inet" })
model PingConfig {
	@minValue(1)
	packetCount: int32;

	@minValue(1)
	@maxValue(65507)
	packetSizeBytes: int32;

	@minValue(1)
	timeoutMs: int32;

	ipFamily?: IPFamily;
}

@example(#{ packetCount: 4, packetSizeBytes: 56, timeoutMs: 3000, ipFamily: "inet" })
model PingConfigPatch {
	@minValue(1)
	packetCount?: int32;

	@minValue(1)
	@maxValue(65507)
	packetSizeBytes?: int32;

	@minValue(1)
	timeoutMs?: int32;

	ipFamily?: IPFamily;
}

@example(#{ protocol: "icmp", maxHops: 30, timeoutMs: 3000, queriesPerHop: 3, packetSizeBytes: 56, port: 33434, ipFamily: "inet" })
model TracerouteConfig {
	protocol: TracerouteProtocol;

	@minValue(1)
	@maxValue(64)
	maxHops: int32;

	@minValue(1)
	@maxValue(60000)
	timeoutMs: int32;

	@minValue(1)
	@maxValue(10)
	queriesPerHop: int32;

	@minValue(1)
	@maxValue(65507)
	packetSizeBytes: int32;

	@minValue(1)
	@maxValue(65535)
	port: int32;

	ipFamily?: IPFamily;
}

@example(#{ protocol: "icmp", maxHops: 30, timeoutMs: 3000, queriesPerHop: 3, packetSizeBytes: 56, port: 33434, ipFamily: "inet" })
model TracerouteConfigPatch {
	protocol?: TracerouteProtocol;

	@minValue(1)
	@maxValue(64)
	maxHops?: int32;

	@minValue(1)
	@maxValue(60000)
	timeoutMs?: int32;

	@minValue(1)
	@maxValue(10)
	queriesPerHop?: int32;

	@minValue(1)
	@maxValue(65507)
	packetSizeBytes?: int32;

	@minValue(1)
	@maxValue(65535)
	port?: int32;

	ipFamily?: IPFamily;
}

@example(#{
	id: "44444444-4444-4444-4444-444444444444",
	projectId: "22222222-2222-2222-2222-222222222222",
	name: "api-latency",
	type: "ping",
	target: "api.netstamp.io",
	selector: #{ label: #{ key: "region", `op`: "eq", value: "tokyo" } },
	description: "Latency and loss to controller API.",
	intervalSeconds: 30,
	labels: #[
		#{
			id: "66666666-6666-6666-6666-666666666666",
			projectId: "22222222-2222-2222-2222-222222222222",
			key: "region",
			value: "tokyo",
			createdAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
			updatedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
		}
	],
	createdAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
	updatedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
	pingConfig: #{ packetCount: 4, packetSizeBytes: 56, timeoutMs: 3000, ipFamily: "inet" },
})
model Check {
	id: uuid;
	projectId: uuid;

	@minLength(1)
	@maxLength(128)
	name: string;

	type: CheckType;

	@minLength(1)
	@maxLength(128)
	target: string;

	selector?: Selector;

	@minLength(1)
	@maxLength(1024)
	description?: string;

	@minValue(1)
	intervalSeconds: int32;

	labels: Label[];
	createdAt: utcDateTime;
	updatedAt: utcDateTime;
	pingConfig?: PingConfig;
	tracerouteConfig?: TracerouteConfig;
}

@example(#{
	check: #{
		id: "44444444-4444-4444-4444-444444444444",
		projectId: "22222222-2222-2222-2222-222222222222",
		name: "api-latency",
		type: "ping",
		target: "api.netstamp.io",
		selector: #{ label: #{ key: "region", `op`: "eq", value: "tokyo" } },
		description: "Latency and loss to controller API.",
		intervalSeconds: 30,
		labels: #[],
		createdAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
		updatedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
		pingConfig: #{ packetCount: 4, packetSizeBytes: 56, timeoutMs: 3000, ipFamily: "inet" },
	},
})
model CheckResponse {
	check: Check;
}

@example(#{
	checks: #[
		#{
			id: "44444444-4444-4444-4444-444444444444",
			projectId: "22222222-2222-2222-2222-222222222222",
			name: "api-latency",
			type: "ping",
			target: "api.netstamp.io",
			selector: #{ label: #{ key: "region", `op`: "eq", value: "tokyo" } },
			intervalSeconds: 30,
			labels: #[],
			createdAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
			updatedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
			pingConfig: #{ packetCount: 4, packetSizeBytes: 56, timeoutMs: 3000, ipFamily: "inet" },
		}
	],
})
model CheckListResponse {
	checks: Check[];
}

@example(#{
	name: "api-latency",
	type: "ping",
	target: "api.netstamp.io",
	selector: #{ label: #{ key: "region", `op`: "eq", value: "tokyo" } },
	description: "Latency and loss to controller API.",
	intervalSeconds: 30,
	labelIds: #["66666666-6666-6666-6666-666666666666"],
	pingConfig: #{ packetCount: 4, packetSizeBytes: 56, timeoutMs: 3000, ipFamily: "inet" },
})
model CreateCheckRequest {
	@minLength(1)
	@maxLength(128)
	name: string;

	type: CheckType;

	@minLength(1)
	@maxLength(128)
	target: string;

	selector?: Selector;

	@minLength(1)
	@maxLength(1024)
	description?: string;

	@minValue(1)
	intervalSeconds: int32;

	labelIds?: uuid[];
	pingConfig?: PingConfigPatch;
	tracerouteConfig?: TracerouteConfigPatch;
}

@example(#{ name: "api-latency", target: "api.netstamp.io", intervalSeconds: 60, labelIds: #["66666666-6666-6666-6666-666666666666"], pingConfig: #{ timeoutMs: 5000 } })
@doc("Patch payload. At least one field must be provided.")
model UpdateCheckRequest {
	@minLength(1)
	@maxLength(128)
	name?: string;

	type?: CheckType;

	@minLength(1)
	@maxLength(128)
	target?: string;

	selector?: Selector;

	@minLength(1)
	@maxLength(1024)
	description?: string;

	@minValue(1)
	intervalSeconds?: int32;

	pingConfig?: PingConfigPatch;
	tracerouteConfig?: TracerouteConfigPatch;
	labelIds?: uuid[];
}
