import "@typespec/http";

using Http;

namespace Netstamp;

alias PingSeriesKey = "latency_avg" | "latency_min" | "latency_max" | "loss_percent";
alias TcpSeriesKey = "connect_avg" | "connect_min" | "connect_max" | "failure_percent";
alias LatestResultStatus = "successful" | "timeout" | "error" | "partial";
alias SeriesResolution = "raw" | "bucket" | "1m";
alias SeriesSource = "raw" | "aggregate";
alias TracerouteTopologyNodeKind = "probe" | "hop" | "destination" | "unknown";

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
	startedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
	finishedAt: utcDateTime.fromISO("2026-05-13T10:00:00.042Z"),
	durationMs: 42,
	status: "successful",
	connectDurationMs: 42,
	resolvedIp: "93.184.216.34",
	ipFamily: "inet",
})
model TcpResult {
	startedAt: utcDateTime;
	finishedAt: utcDateTime;

	@minValue(0)
	durationMs: int32;

	status: TcpStatus;

	@minValue(0)
	connectDurationMs?: float64;

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
	tcp?: TcpResult[];
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
	@minValue(0)
	accepted: int32;

	serverTime: utcDateTime;
}

alias SeriesPoint = [int64, float64];

@example(#{
	name: "latency_avg",
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

@example(#{ from: 1778662800000, to: 1778749200000, maxDataPoints: 600, source: "raw", resolution: "bucket", totalPoints: 487 })
model PingSeriesQueryMetadata {
	@minValue(1)
	from: int64;

	@minValue(1)
	to: int64;

	@minValue(1)
	@maxValue(2000)
	maxDataPoints: int32;

	source?: SeriesSource;
	resolution: SeriesResolution;

	@minValue(0)
	totalPoints: int64;
}

@example(#{
	series: #{
		latency_avg: #{
			name: "latency_avg",
			labels: #{ probeId: "33333333-3333-3333-3333-333333333333", checkId: "44444444-4444-4444-4444-444444444444", checkType: "ping" },
			unit: "ms",
			points: #[#[1778666400000, 12.3]],
		}
	},
	meta: #{ from: 1778662800000, to: 1778749200000, maxDataPoints: 600, source: "raw", resolution: "bucket", totalPoints: 487 },
})
model PingSeriesResponse {
	series: Record<Series>;
	meta: PingSeriesQueryMetadata;
}

@example(#{
	series: #{
		connect_avg: #{
			name: "connect_avg",
			labels: #{ probeId: "33333333-3333-3333-3333-333333333333", checkId: "44444444-4444-4444-4444-444444444444", checkType: "tcp" },
			unit: "ms",
			points: #[#[1778666400000, 12.3]],
		}
	},
	meta: #{ from: 1778662800000, to: 1778749200000, maxDataPoints: 600, source: "raw", resolution: "bucket", totalPoints: 487 },
})
model TcpSeriesResponse {
	series: Record<Series>;
	meta: PingSeriesQueryMetadata;
}

@example(#{
	averageRttMs: 12.3,
	maxRttMs: 42.5,
	lossPercent: 1.4,
	successRate: 98.6,
	samples: 1920,
})
model PingInsightSummary {
	@minValue(0)
	averageRttMs?: float64;

	@minValue(0)
	maxRttMs?: float64;

	@minValue(0)
	@maxValue(100)
	lossPercent?: float64;

	@minValue(0)
	@maxValue(100)
	successRate?: float64;

	@minValue(0)
	samples: int64;
}

@example(#{
	summary: #{ averageRttMs: 12.3, maxRttMs: 42.5, lossPercent: 1.4, successRate: 98.6, samples: 1920 },
	meta: #{ from: 1778662800000, to: 1778749200000, maxDataPoints: 600, source: "aggregate", resolution: "1m", totalPoints: 487 },
})
model PingInsightResponse {
	summary: PingInsightSummary;
	meta: PingSeriesQueryMetadata;
}

@example(#{
	averageConnectMs: 12.3,
	maxConnectMs: 42.5,
	failurePercent: 1.4,
	successRate: 98.6,
	samples: 487,
})
model TcpInsightSummary {
	@minValue(0)
	averageConnectMs?: float64;

	@minValue(0)
	maxConnectMs?: float64;

	@minValue(0)
	@maxValue(100)
	failurePercent?: float64;

	@minValue(0)
	@maxValue(100)
	successRate?: float64;

	@minValue(0)
	samples: int64;
}

@example(#{
	summary: #{ averageConnectMs: 12.3, maxConnectMs: 42.5, failurePercent: 1.4, successRate: 98.6, samples: 487 },
	meta: #{ from: 1778662800000, to: 1778749200000, maxDataPoints: 600, source: "aggregate", resolution: "1m", totalPoints: 487 },
})
model TcpInsightResponse {
	summary: TcpInsightSummary;
	meta: PingSeriesQueryMetadata;
}

@example(#{ from: 1778662800000, to: 1778749200000, limit: 100, nextCursor: 1778666400000 })
model TracerouteRunsQueryMetadata {
	@minValue(1)
	from: int64;

	@minValue(1)
	to: int64;

	@minValue(1)
	@maxValue(500)
	limit: int32;

	@minValue(1)
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

@example(#{ from: 1778662800000, to: 1778749200000, maxDataPoints: 600, resolution: "bucket", totalRuns: 487 })
model TracerouteInsightQueryMetadata {
	@minValue(1)
	from: int64;

	@minValue(1)
	to: int64;

	@minValue(1)
	@maxValue(2000)
	maxDataPoints: int32;

	resolution: SeriesResolution;

	@minValue(0)
	totalRuns: int64;
}

@example(#{
	timestampMs: 1778666400000,
	bucketFromMs: 1778666400000,
	bucketToMs: 1778667000000,
	resultCount: 5,
	finalRttAvgMs: 12.4,
	finalLossPercent: 0,
	hasLoss: false,
	hasRouteChange: true,
	destinationReached: true,
})
model TracerouteInsightPoint {
	@minValue(1)
	timestampMs: int64;

	@minValue(1)
	bucketFromMs: int64;

	@minValue(1)
	bucketToMs: int64;

	runStartedAt?: utcDateTime;

	@minValue(1)
	resultCount: int64;

	@minValue(0)
	finalRttAvgMs?: float64;

	@minValue(0)
	@maxValue(100)
	finalLossPercent?: float64;

	hasLoss: boolean;
	hasRouteChange: boolean;
	destinationReached: boolean;
}

@example(#{
	points: #[
		#{ timestampMs: 1778666400000, bucketFromMs: 1778666400000, bucketToMs: 1778667000000, resultCount: 5, finalRttAvgMs: 12.4, hasLoss: false, hasRouteChange: true, destinationReached: true }
	],
	query: #{ from: 1778662800000, to: 1778749200000, maxDataPoints: 600, resolution: "bucket", totalRuns: 487 },
})
model TracerouteInsightResponse {
	points: TracerouteInsightPoint[];
	query: TracerouteInsightQueryMetadata;
}

@example(#{ id: "ip:192.0.2.1", kind: "hop", label: "gateway.local", address: "192.0.2.1", hostname: "gateway.local", hopIndex: 1, seenCount: 12, avgRttMs: 1.8, lossPercent: 0 })
model TracerouteTopologyNode {
	id: string;
	kind: TracerouteTopologyNodeKind;
	label: string;
	address?: ipAddress;
	hostname?: string;
	probeId?: uuid;
	checkId?: uuid;
	target?: string;

	@minValue(1)
	@maxValue(64)
	hopIndex?: int32;

	@minValue(0)
	seenCount: int32;

	@minValue(0)
	avgRttMs?: float64;

	@minValue(0)
	@maxValue(100)
	lossPercent?: float64;
}

@example(#{
	id: "probe:33333333-3333-3333-3333-333333333333->ip:192.0.2.1",
	source: "probe:33333333-3333-3333-3333-333333333333",
	target: "ip:192.0.2.1",
	seenCount: 12,
	avgRttMs: 1.8,
	lossPercent: 0,
})
model TracerouteTopologyEdge {
	id: string;
	source: string;
	target: string;

	@minValue(0)
	seenCount: int32;

	@minValue(0)
	avgRttMs?: float64;

	@minValue(0)
	@maxValue(100)
	lossPercent?: float64;
}

@example(#{ from: 1778662800000, to: 1778749200000, limit: 100 })
model TracerouteTopologyQueryMetadata {
	@minValue(1)
	from: int64;

	@minValue(1)
	to: int64;

	@minValue(1)
	@maxValue(500)
	limit: int32;
}

@example(#{
	nodes: #[
		#{ id: "probe:33333333-3333-3333-3333-333333333333", kind: "probe", label: "fra-bm-02", probeId: "33333333-3333-3333-3333-333333333333", seenCount: 12 },
		#{ id: "ip:192.0.2.1", kind: "hop", label: "gateway.local", address: "192.0.2.1", hostname: "gateway.local", hopIndex: 1, seenCount: 12, avgRttMs: 1.8, lossPercent: 0 }
	],
	edges: #[
		#{ id: "probe:33333333-3333-3333-3333-333333333333->ip:192.0.2.1", source: "probe:33333333-3333-3333-3333-333333333333", target: "ip:192.0.2.1", seenCount: 12, avgRttMs: 1.8, lossPercent: 0 }
	],
	query: #{ from: 1778662800000, to: 1778749200000, limit: 100 },
})
model TracerouteTopologyResponse {
	nodes: TracerouteTopologyNode[];
	edges: TracerouteTopologyEdge[];
	query: TracerouteTopologyQueryMetadata;
}

@example(#{
	type: "ping",
	probeId: "33333333-3333-3333-3333-333333333333",
	checkId: "44444444-4444-4444-4444-444444444444",
	latestStartedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
	latestStatus: "successful",
})
model LatestResult {
	type: CheckType;
	probeId: uuid;
	checkId: uuid;
	latestStartedAt: utcDateTime;
	latestStatus: LatestResultStatus;
}

@example(#{
	results: #[
		#{
			type: "ping",
			probeId: "33333333-3333-3333-3333-333333333333",
			checkId: "44444444-4444-4444-4444-444444444444",
			latestStartedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
			latestStatus: "successful",
		}
	],
})
model LatestResultsResponse {
	results: LatestResult[];
}

model LatestResultsQuery {
	@query
	probeId?: uuid;

	@query
	checkId?: uuid;

	@query
	type?: CheckType;
}

@example(#{ agentVersion: "netstamp-probe/0.1.0", publicV4: "203.0.113.10", publicV6: "2001:db8::10", as: "AS15169 Google LLC", addrs: #["10.0.0.10"] })
model RuntimeStatusRequest {
	agentVersion?: string;
	publicV4?: ipv4Address;
	publicV6?: ipv6Address;
	as?: string;
	addrs?: ipAddress[];
}

@example(#{ families: #["inet", "inet6"] })
model RuntimeIPFamilyCapabilitiesRequest {
	families?: IPFamily[];
}

@example(#{
	serverTime: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
	minimumSupportedAgentVersion: "0.1.0",
})
model RuntimeHelloResponse {
	serverTime: utcDateTime;
	minimumSupportedAgentVersion: string;
}

@example(#{ serverTime: utcDateTime.fromISO("2026-05-13T10:00:00Z") })
model RuntimeHeartbeatResponse {
	serverTime: utcDateTime;
}

@example(#{ serverTime: utcDateTime.fromISO("2026-05-13T10:00:00Z") })
model RuntimeIPFamilyCapabilitiesResponse {
	serverTime: utcDateTime;
}

@example(#{
	serverTime: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
	assignments: #[],
})
model RuntimeAssignmentsResponse {
	serverTime: utcDateTime;
	assignments: Assignment[];
}
