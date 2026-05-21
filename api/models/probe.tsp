namespace Netstamp;

alias ProbeState = "online" | "offline";

@example(#{
	probeId: "33333333-3333-3333-3333-333333333333",
	state: "online",
	lastSeenAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
	agentVersion: "netstamp-probe/0.1.0",
	publicV4: "203.0.113.10",
	publicV6: "2001:db8::10",
	as: "AS15169 Google LLC",
	addrs: #["10.0.0.10"],
	updatedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
})
model ProbeStatus {
	probeId: uuid;
	state: ProbeState;
	lastSeenAt?: utcDateTime;
	agentVersion?: string;
	publicV4?: ipv4Address;
	publicV6?: ipv6Address;
	as?: string;
	addrs?: ipAddress[];
	updatedAt: utcDateTime;
}

@example(#{
	id: "33333333-3333-3333-3333-333333333333",
	projectId: "22222222-2222-2222-2222-222222222222",
	name: "tokyo-vps-1",
	enabled: true,
	subdivisionCode: "JP-13",
	latitude: 35.6762,
	longitude: 139.6503,
	labels: #[],
	status: #{ probeId: "33333333-3333-3333-3333-333333333333", state: "online", lastSeenAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"), updatedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z") },
	createdAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
	updatedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
})
model Probe {
	id: uuid;
	projectId: uuid;

	@minLength(1)
	@maxLength(128)
	name: string;

	enabled: boolean;

	@pattern("^[A-Z]{2}-[A-Z0-9]{1,3}$")
	subdivisionCode?: string;

	@minValue(-90)
	@maxValue(90)
	latitude?: float64;

	@minValue(-180)
	@maxValue(180)
	longitude?: float64;

	labels: Label[];
	status?: ProbeStatus;
	createdAt: utcDateTime;
	updatedAt: utcDateTime;
}

@example(#{
	probe: #{
		id: "33333333-3333-3333-3333-333333333333",
		projectId: "22222222-2222-2222-2222-222222222222",
		name: "tokyo-vps-1",
		enabled: true,
		subdivisionCode: "JP-13",
		latitude: 35.6762,
		longitude: 139.6503,
		labels: #[],
		createdAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
		updatedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
	},
})
model ProbeResponse {
	probe: Probe;
}

@example(#{
	probes: #[
		#{
			id: "33333333-3333-3333-3333-333333333333",
			projectId: "22222222-2222-2222-2222-222222222222",
			name: "tokyo-vps-1",
			enabled: true,
			labels: #[],
			createdAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
			updatedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
		}
	],
})
model ProbeListResponse {
	probes: Probe[];
}

@example(#{
	probe: #{
		id: "33333333-3333-3333-3333-333333333333",
		projectId: "22222222-2222-2222-2222-222222222222",
		name: "tokyo-vps-1",
		enabled: true,
		labels: #[],
		createdAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
		updatedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
	},
	secret: "nsp_6b5fbd35d3a1417f9c0c8c2e",
})
model ProbeSecretResponse {
	probe: Probe;

	@secret
	@doc("Plaintext probe secret. It is returned only at creation and rotation time.")
	secret: string;
}

@example(#{ name: "tokyo-vps-1", enabled: true, subdivisionCode: "JP-13", latitude: 35.6762, longitude: 139.6503, labelIds: #["66666666-6666-6666-6666-666666666666"] })
model CreateProbeRequest {
	@minLength(1)
	@maxLength(128)
	name: string;

	enabled?: boolean = true;

	@pattern("^[A-Z]{2}-[A-Z0-9]{1,3}$")
	subdivisionCode?: string;

	@minValue(-90)
	@maxValue(90)
	latitude?: float64;

	@minValue(-180)
	@maxValue(180)
	longitude?: float64;

	labelIds?: uuid[];
}

@example(#{ name: "tokyo-vps-2", enabled: false, labelIds: #["66666666-6666-6666-6666-666666666666"] })
@doc("Patch payload. At least one field must be provided.")
model UpdateProbeRequest {
	@minLength(1)
	@maxLength(128)
	name?: string;

	enabled?: boolean;

	@pattern("^[A-Z]{2}-[A-Z0-9]{1,3}$")
	subdivisionCode?: string;

	@minValue(-90)
	@maxValue(90)
	latitude?: float64;

	@minValue(-180)
	@maxValue(180)
	longitude?: float64;

	labelIds?: uuid[];
}

@example(#{ heartbeatIntervalSeconds: 30, assignmentPollIntervalSeconds: 30, maxConcurrentWorkers: 16, initialBackoffSeconds: 1, maxBackoffSeconds: 30, maxAttempts: 5 })
model RuntimeConfig {
	@minValue(1)
	heartbeatIntervalSeconds: int32;

	@minValue(1)
	assignmentPollIntervalSeconds: int32;

	@minValue(1)
	maxConcurrentWorkers: int32;

	@minValue(1)
	initialBackoffSeconds: int32;

	@minValue(1)
	maxBackoffSeconds: int32;

	@minValue(1)
	maxAttempts: int32;
}

@example(#{
	id: "77777777-7777-7777-7777-777777777777",
	projectId: "22222222-2222-2222-2222-222222222222",
	checkVersion: "01HXYZ7K6P4E4M3F0CB1W9DYJ5",
	selectorVersion: "01HXYZ7K6P4E4M3F0CB1W9DYJ6",
	check: #{
		id: "44444444-4444-4444-4444-444444444444",
		projectId: "22222222-2222-2222-2222-222222222222",
		name: "api-latency",
		type: "ping",
		target: "api.netstamp.io",
		intervalSeconds: 30,
		labels: #[],
		createdAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
		updatedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
		pingConfig: #{ packetCount: 4, packetSizeBytes: 56, timeoutMs: 3000 },
	},
})
model Assignment {
	id: uuid;
	projectId: uuid;
	checkVersion: string;
	selectorVersion: string;
	check?: Check;
	probe?: Probe;
}
