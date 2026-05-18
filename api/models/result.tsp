namespace Netstamp;

alias PingMetric = "rttAvgMs" | "lossPercent" | "successRate";

@example(#{
	startedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
	finishedAt: utcDateTime.fromISO("2026-05-13T10:00:01Z"),
	durationMs: 1000,
	status: "successful",
	sentCount: 4,
	receivedCount: 4,
	lossPercent: 0,
	rttMinMs: 10.1,
	rttAvgMs: 12.3,
	rttMedianMs: 12,
	rttMaxMs: 15.6,
	rttStddevMs: 1.7,
	rttSamplesMs: #[10.1, 12.0, 11.5, 15.6],
	resolvedIp: "1.1.1.1",
	ipFamily: "inet",
})
model PingResult {
	startedAt: utcDateTime;
	finishedAt: utcDateTime;

	@minValue(0)
	durationMs: int32;

	status: PingStatus;

	@minValue(0)
	sentCount: int32;

	@minValue(0)
	receivedCount: int32;

	@minValue(0)
	@maxValue(100)
	lossPercent: float64;

	@minValue(0)
	rttMinMs?: float64;

	@minValue(0)
	rttAvgMs?: float64;

	@minValue(0)
	rttMedianMs?: float64;

	@minValue(0)
	rttMaxMs?: float64;

	@minValue(0)
	rttStddevMs?: float64;

	rttSamplesMs?: float64[];
	resolvedIp?: ipAddress;
	ipFamily?: IPFamily;
	errorCode?: string;
	errorMessage?: string;
}

@example(#{
	hopIndex: 1,
	address: "192.0.2.1",
	hostname: "gateway.local",
	sentCount: 3,
	receivedCount: 3,
	lossPercent: 0,
	rttMinMs: 1.5,
	rttAvgMs: 1.7,
	rttMedianMs: 1.7,
	rttMaxMs: 1.9,
	rttStddevMs: 0.2,
	rttSamplesMs: #[1.5, 1.7, 1.9],
})
model TracerouteHop {
	@minValue(1)
	@maxValue(64)
	hopIndex: int32;

	address?: ipAddress;
	hostname?: string;

	@minValue(0)
	sentCount: int32;

	@minValue(0)
	receivedCount: int32;

	@minValue(0)
	@maxValue(100)
	lossPercent: float64;

	@minValue(0)
	rttMinMs?: float64;

	@minValue(0)
	rttAvgMs?: float64;

	@minValue(0)
	rttMedianMs?: float64;

	@minValue(0)
	rttMaxMs?: float64;

	@minValue(0)
	rttStddevMs?: float64;

	rttSamplesMs?: float64[];
	errorCode?: string;
	errorMessage?: string;
}

@example(#{
	startedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
	finishedAt: utcDateTime.fromISO("2026-05-13T10:00:04Z"),
	durationMs: 4000,
	status: "partial",
	resolvedIp: "93.184.216.34",
	ipFamily: "inet",
	destinationReached: false,
	hopCount: 12,
	hops: #[#{ hopIndex: 1, address: "192.0.2.1", sentCount: 3, receivedCount: 3, lossPercent: 0 }],
	errorCode: "destination_unreached",
	errorMessage: "destination was not reached before max hops",
})
model TracerouteResult {
	startedAt: utcDateTime;
	finishedAt: utcDateTime;

	@minValue(0)
	durationMs: int32;

	status: TracerouteStatus;
	resolvedIp?: ipAddress;
	ipFamily?: IPFamily;
	destinationReached: boolean;

	@minValue(0)
	hopCount: int32;

	hops?: TracerouteHop[];
	errorCode?: string;
	errorMessage?: string;
}

@example(#{
	checkId: "44444444-4444-4444-4444-444444444444",
	type: "ping",
	ping: #[
		#{
			startedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
			finishedAt: utcDateTime.fromISO("2026-05-13T10:00:01Z"),
			durationMs: 1000,
			status: "successful",
			sentCount: 4,
			receivedCount: 4,
			lossPercent: 0,
			rttAvgMs: 12.3,
			rttSamplesMs: #[10.1, 12.0, 11.5, 15.6],
			resolvedIp: "1.1.1.1",
			ipFamily: "inet",
		}
	],
})
model RuntimeResultGroup {
	checkId: uuid;
	type: CheckType;
	ping?: PingResult[];
	traceroute?: TracerouteResult[];
}

@example(#{
	results: #[
		#{
			checkId: "44444444-4444-4444-4444-444444444444",
			type: "ping",
			ping: #[
				#{
					startedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
					finishedAt: utcDateTime.fromISO("2026-05-13T10:00:01Z"),
					durationMs: 1000,
					status: "successful",
					sentCount: 4,
					receivedCount: 4,
					lossPercent: 0,
				}
			],
		}
	],
})
model SubmitResultsRequest {
	@minItems(1)
	@maxItems(100)
	results: RuntimeResultGroup[];
}

@example(#{ accepted: 1, serverTime: utcDateTime.fromISO("2026-05-13T10:00:02Z") })
model SubmitResultsResponse {
	accepted: int32;
	serverTime: utcDateTime;
}

alias SeriesPoint = [int64, float64];

@example(#{
	name: "rttAvgMs",
	labels: #{ probeId: "33333333-3333-3333-3333-333333333333", checkId: "44444444-4444-4444-4444-444444444444", checkType: "ping" },
	unit: "ms",
	points: #[#[1778666400000, 12.3]],
})
model Series {
	name: string;
	labels: Record<string>;
	unit: string;
	points: SeriesPoint[];
}

@example(#{ from: 1778662800000, to: 1778749200000, maxDataPoints: 600, resolution: "bucket", totalPoints: 487 })
model PingSeriesQueryMetadata {
	from: int64;
	to: int64;
	maxDataPoints: int32;
	resolution: string;
	totalPoints: int64;
}

@example(#{
	series: #[
		#{
			name: "rttAvgMs",
			labels: #{ probeId: "33333333-3333-3333-3333-333333333333", checkId: "44444444-4444-4444-4444-444444444444", checkType: "ping" },
			unit: "ms",
			points: #[#[1778666400000, 12.3]],
		}
	],
	query: #{ from: 1778662800000, to: 1778749200000, maxDataPoints: 600, resolution: "bucket", totalPoints: 487 },
})
model PingSeriesResponse {
	series: Series[];
	query: PingSeriesQueryMetadata;
}

@example(#{ from: 1778662800000, to: 1778749200000, limit: 100, nextCursor: 1778666400000 })
model TracerouteRunsQueryMetadata {
	from: int64;
	to: int64;
	limit: int32;
	nextCursor?: int64;
}

@example(#{
	runs: #[
		#{
			startedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
			finishedAt: utcDateTime.fromISO("2026-05-13T10:00:04Z"),
			durationMs: 4000,
			status: "partial",
			destinationReached: false,
			hopCount: 12,
			hops: #[#{ hopIndex: 1, sentCount: 3, receivedCount: 3, lossPercent: 0 }],
		}
	],
	query: #{ from: 1778662800000, to: 1778749200000, limit: 100 },
})
model TracerouteRunsResponse {
	runs: TracerouteResult[];
	query: TracerouteRunsQueryMetadata;
}

@example(#{ agentVersion: "netstamp-probe/0.1.0", publicV4: "203.0.113.10", publicV6: "2001:db8::10", as: "AS15169 Google LLC", addrs: #["10.0.0.10"] })
model RuntimeStatusRequest {
	agentVersion?: string;
	publicV4?: ipAddress;
	publicV6?: ipAddress;
	as?: string;
	addrs?: ipAddress[];
}

@example(#{
	serverTime: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
	minimumSupportedAgentVersion: "0.1.0",
	config: #{ heartbeatIntervalSeconds: 30, assignmentPollIntervalSeconds: 30, maxConcurrentWorkers: 16, initialBackoffSeconds: 1, maxBackoffSeconds: 30, maxAttempts: 5 },
})
model RuntimeHelloResponse {
	serverTime: utcDateTime;
	minimumSupportedAgentVersion: string;
	config: RuntimeConfig;
}

@example(#{ serverTime: utcDateTime.fromISO("2026-05-13T10:00:00Z") })
model RuntimeHeartbeatResponse {
	serverTime: utcDateTime;
}

@example(#{
	serverTime: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
	config: #{ heartbeatIntervalSeconds: 30, assignmentPollIntervalSeconds: 30, maxConcurrentWorkers: 16, initialBackoffSeconds: 1, maxBackoffSeconds: 30, maxAttempts: 5 },
	assignments: #[],
})
model RuntimeAssignmentsResponse {
	serverTime: utcDateTime;
	config: RuntimeConfig;
	assignments: Assignment[];
}
