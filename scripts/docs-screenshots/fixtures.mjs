const projectId = "20000000-0000-4000-8000-000000000001";
const pageId = "70000000-0000-4000-8000-000000000001";
const userId = "10000000-0000-4000-8000-000000000001";
const now = "2026-08-13T03:20:00.000Z";
const createdAt = "2026-07-08T09:00:00.000Z";

export const ids = {
	project: projectId,
	page: pageId,
	user: userId,
	probes: {
		taipei: "30000000-0000-4000-8000-000000000001",
		tokyo: "30000000-0000-4000-8000-000000000002",
		frankfurt: "30000000-0000-4000-8000-000000000003",
		singapore: "30000000-0000-4000-8000-000000000004"
	},
	checks: {
		ping: "40000000-0000-4000-8000-000000000001",
		http: "40000000-0000-4000-8000-000000000002",
		tcp: "40000000-0000-4000-8000-000000000003",
		traceroute: "40000000-0000-4000-8000-000000000004"
	},
	rules: {
		http: "60000000-0000-4000-8000-000000000001",
		loss: "60000000-0000-4000-8000-000000000002"
	},
	incidents: {
		open: "61000000-0000-4000-8000-000000000001",
		acknowledged: "61000000-0000-4000-8000-000000000002",
		resolved: "61000000-0000-4000-8000-000000000003"
	}
};

const labelDefinitions = [
	["50000000-0000-4000-8000-000000000001", "region", "tw"],
	["50000000-0000-4000-8000-000000000002", "region", "jp"],
	["50000000-0000-4000-8000-000000000003", "network", "office"],
	["50000000-0000-4000-8000-000000000004", "provider", "cloud"],
	["50000000-0000-4000-8000-000000000005", "environment", "production"]
];

const labels = labelDefinitions.map(([id, key, value]) => ({ id, projectId, key, value, createdAt, updatedAt: now }));
const label = (key, value) => labels.find(item => item.key === key && item.value === value);

const probe = ({ id, name, locationName, latitude, longitude, enabled = true, publicV4, publicV6, probeLabels }) => ({
	id,
	projectId,
	name,
	enabled,
	locationName,
	latitude,
	longitude,
	labels: probeLabels,
	status: {
		probeId: id,
		state: enabled ? "online" : "offline",
		...(enabled ? { lastSeenAt: now, onlineSince: "2026-08-07T08:00:00.000Z", uptimeSeconds: 501600 } : {}),
		agentVersion: "netstamp-probe/1.4.0",
		...(publicV4 ? { publicV4 } : {}),
		...(publicV6 ? { publicV6 } : {}),
		addrs: ["192.0.2.10"],
		updatedAt: now
	},
	createdAt,
	updatedAt: now
});

const probes = [
	probe({
		id: ids.probes.taipei,
		name: "Taipei Office",
		locationName: "Taipei, Taiwan",
		latitude: 25.033,
		longitude: 121.5654,
		publicV4: "192.0.2.10",
		publicV6: "2001:db8::10",
		probeLabels: [label("region", "tw"), label("network", "office"), label("environment", "production")]
	}),
	probe({
		id: ids.probes.tokyo,
		name: "Tokyo Cloud",
		locationName: "Tokyo, Japan",
		latitude: 35.6762,
		longitude: 139.6503,
		publicV4: "198.51.100.20",
		publicV6: "2001:db8::20",
		probeLabels: [label("region", "jp"), label("provider", "cloud"), label("environment", "production")]
	}),
	probe({
		id: ids.probes.frankfurt,
		name: "Frankfurt Edge",
		locationName: "Frankfurt, Germany",
		latitude: 50.1109,
		longitude: 8.6821,
		publicV4: "203.0.113.30",
		probeLabels: [label("provider", "cloud"), label("environment", "production")]
	}),
	probe({
		id: ids.probes.singapore,
		name: "Singapore Branch",
		locationName: "Singapore",
		latitude: 1.3521,
		longitude: 103.8198,
		enabled: false,
		probeLabels: [label("network", "office")]
	})
];

const commonCheck = (id, name, type, target, description) => ({
	id,
	projectId,
	name,
	type,
	target,
	description,
	selector: { label: { key: "environment", op: "eq", value: "production" } },
	intervalSeconds: 60,
	labels: [label("environment", "production")],
	createdAt,
	updatedAt: now
});

const checks = [
	{
		...commonCheck(ids.checks.ping, "Cloudflare DNS", "ping", "1.1.1.1", "Track packet loss and round-trip latency."),
		pingConfig: { packetCount: 4, packetSizeBytes: 56, timeoutMs: 3000, ipFamily: "inet" }
	},
	{
		...commonCheck(ids.checks.http, "HTTPS Health", "http", "https://example.com/health", "Verify the public health endpoint."),
		httpConfig: {
			method: "GET",
			headers: [{ name: "Accept", value: "application/json" }],
			timeoutMs: 10000,
			ipFamily: "inet",
			followRedirects: true,
			skipTlsVerify: false,
			expectedStatuses: [{ kind: "class", class: "2xx" }],
			bodyContains: "ok",
			sensitiveFieldsRedacted: false
		}
	},
	{
		...commonCheck(ids.checks.tcp, "HTTPS Port", "tcp", "example.com", "Measure TCP connection time on port 443."),
		tcpConfig: { port: 443, timeoutMs: 3000, ipFamily: "inet" }
	},
	{
		...commonCheck(ids.checks.traceroute, "Public Route", "traceroute", "1.1.1.1", "Observe route changes to a public resolver."),
		tracerouteConfig: { protocol: "icmp", maxHops: 30, timeoutMs: 3000, queriesPerHop: 3, packetSizeBytes: 56, port: 33434, ipFamily: "inet" }
	}
];

const assignments = checks.flatMap((check, checkIndex) =>
	probes.slice(0, 3).map((currentProbe, probeIndex) => ({
		id: `80000000-0000-4000-8${checkIndex}00-00000000000${probeIndex + 1}`,
		projectId,
		probeId: currentProbe.id,
		checkId: check.id,
		checkVersion: "01J5DOCSEXAMPLE000000000001",
		selectorVersion: "01J5DOCSEXAMPLE000000000002",
		probe: currentProbe,
		check
	}))
);

const project = {
	id: projectId,
	name: "Docs Demo",
	slug: "docs-demo",
	createdByUserId: userId,
	createdAt,
	updatedAt: now
};

const user = {
	id: userId,
	email: "alex@example.com",
	displayName: "Alex Morgan",
	emailVerified: true,
	isSystemAdmin: true,
	hasPassword: true
};

const member = (id, memberUserId, displayName, email, role) => ({
	id,
	projectId,
	userId: memberUserId,
	role,
	user: { id: memberUserId, email, displayName },
	createdAt,
	updatedAt: now
});

const members = [
	member("11000000-0000-4000-8000-000000000001", userId, "Alex Morgan", "alex@example.com", "owner"),
	member("11000000-0000-4000-8000-000000000002", "10000000-0000-4000-8000-000000000002", "Sam Rivera", "sam@example.com", "admin"),
	member("11000000-0000-4000-8000-000000000003", "10000000-0000-4000-8000-000000000003", "Taylor Chen", "taylor@example.com", "viewer")
];

const invites = [
	{
		id: "12000000-0000-4000-8000-000000000001",
		projectId,
		invitedEmail: "viewer@example.com",
		invitedByUserId: userId,
		role: "viewer",
		status: "pending",
		project: { id: projectId, name: project.name, slug: project.slug },
		invitedUser: { id: "10000000-0000-4000-8000-000000000004", email: "viewer@example.com", displayName: "Jordan Lee" },
		invitedByUser: { id: userId, email: user.email, displayName: user.displayName },
		createdAt: "2026-08-12T02:00:00.000Z",
		updatedAt: "2026-08-12T02:00:00.000Z"
	}
];

const notifications = [
	{
		id: "62000000-0000-4000-8000-000000000001",
		name: "Platform On-call",
		type: "slack",
		enabled: true,
		config: { url: "https://hooks.slack.com/services/..." },
		createdAt,
		updatedAt: now
	},
	{
		id: "62000000-0000-4000-8000-000000000002",
		name: "SRE Email",
		type: "email",
		enabled: true,
		config: { to: ["sre@example.com"] },
		createdAt,
		updatedAt: now
	},
	{
		id: "62000000-0000-4000-8000-000000000003",
		name: "Incident Webhook",
		type: "webhook",
		enabled: true,
		config: { url: "https://hooks.example.com/..." },
		createdAt,
		updatedAt: now
	}
];

const rules = [
	{
		id: ids.rules.http,
		name: "HTTP Availability",
		description: "Open an incident when the health endpoint fails.",
		enabled: true,
		severity: "critical",
		scope: { checkType: "http", checkId: ids.checks.http },
		condition: { type: "metric_threshold", metric: "http.failure_percent", operator: "gte", threshold: 5, windowSeconds: 300, minSamples: 3 },
		triggerAfterSeconds: 60,
		cooldownSeconds: 900,
		notificationIds: [notifications[0].id, notifications[1].id],
		createdAt,
		updatedAt: now
	},
	{
		id: ids.rules.loss,
		name: "Packet Loss",
		description: "Warn when packet loss exceeds five percent.",
		enabled: true,
		severity: "warning",
		scope: { checkType: "ping", checkId: ids.checks.ping },
		condition: { type: "metric_threshold", metric: "ping.loss_percent", operator: "gte", threshold: 5, windowSeconds: 300, minSamples: 3 },
		triggerAfterSeconds: 120,
		cooldownSeconds: 600,
		notificationIds: [notifications[0].id],
		createdAt,
		updatedAt: now
	}
];

const incident = ({ id, ruleId, probeId, checkId, status, severity, lastValue, openedAt, resolvedAt }) => {
	const currentProbe = probes.find(item => item.id === probeId);
	const currentCheck = checks.find(item => item.id === checkId);
	return {
		id,
		ruleId,
		probeId,
		checkId,
		probe: { id: currentProbe.id, name: currentProbe.name },
		check: { id: currentCheck.id, name: currentCheck.name, type: currentCheck.type, target: currentCheck.target },
		checkType: currentCheck.type,
		status,
		...(resolvedAt ? { resolutionReason: "condition_cleared", resolvedAt } : {}),
		severity,
		lastEvaluationState: status === "resolved" ? "clear" : "firing",
		openedAt,
		lastEvaluatedAt: now,
		lastTriggeredAt: "2026-08-13T03:17:00.000Z",
		lastValue,
		lastSummary: {
			state: status === "resolved" ? "clear" : "firing",
			metric: currentCheck.type === "ping" ? "ping.loss_percent" : "http.failure_percent",
			value: lastValue,
			samples: 12
		},
		lastNotificationSentAt: "2026-08-13T03:18:00.000Z",
		nextNotificationEligibleAt: "2026-08-13T03:33:00.000Z",
		suppressedNotificationCount: status === "open" ? 1 : 0,
		createdAt: openedAt,
		updatedAt: now
	};
};

const incidents = [
	incident({
		id: ids.incidents.open,
		ruleId: ids.rules.http,
		probeId: ids.probes.tokyo,
		checkId: ids.checks.http,
		status: "open",
		severity: "critical",
		lastValue: 100,
		openedAt: "2026-08-13T02:48:00.000Z"
	}),
	incident({
		id: ids.incidents.acknowledged,
		ruleId: ids.rules.loss,
		probeId: ids.probes.taipei,
		checkId: ids.checks.ping,
		status: "acknowledged",
		severity: "warning",
		lastValue: 8.4,
		openedAt: "2026-08-13T01:20:00.000Z"
	}),
	incident({
		id: ids.incidents.resolved,
		ruleId: ids.rules.http,
		probeId: ids.probes.frankfurt,
		checkId: ids.checks.http,
		status: "resolved",
		severity: "warning",
		lastValue: 0,
		openedAt: "2026-08-12T22:10:00.000Z",
		resolvedAt: "2026-08-12T22:32:00.000Z"
	})
];

const statusPage = {
	id: pageId,
	projectId,
	slug: "docs-demo",
	title: "Docs Demo Status",
	description: "Live availability for the services used in this documentation example.",
	enabled: true,
	footerText: "Powered by Netstamp",
	theme: "light",
	showTargets: true,
	showProbeNames: true,
	showProbeLocations: true,
	showIncidentHistory: true,
	showGeneratedAt: true,
	defaultChartMode: "compact",
	defaultChartRange: "24h",
	createdAt,
	updatedAt: now
};

const statusElements = [
	{
		id: "71000000-0000-4000-8000-000000000001",
		publicPageId: pageId,
		kind: "folder",
		assignmentIds: [],
		title: "Core Services",
		description: "Customer-facing endpoints",
		sortOrder: 0,
		displayMode: "status",
		chartMode: "inherit",
		createdAt,
		updatedAt: now
	},
	{
		id: "71000000-0000-4000-8000-000000000002",
		publicPageId: pageId,
		parentElementId: "71000000-0000-4000-8000-000000000001",
		kind: "assignment_group",
		checkId: ids.checks.http,
		assignmentSelectionMode: "all_check",
		assignmentIds: assignments.filter(item => item.checkId === ids.checks.http).map(item => item.id),
		title: "HTTPS API",
		description: "Public health endpoint",
		sortOrder: 0,
		displayMode: "history",
		chartMode: "inherit",
		checkName: "HTTPS Health",
		checkType: "http",
		checkTarget: "https://example.com/health",
		checkIntervalSeconds: 60,
		createdAt,
		updatedAt: now
	},
	{
		id: "71000000-0000-4000-8000-000000000003",
		publicPageId: pageId,
		parentElementId: "71000000-0000-4000-8000-000000000001",
		kind: "assignment_group",
		checkId: ids.checks.ping,
		assignmentSelectionMode: "all_check",
		assignmentIds: assignments.filter(item => item.checkId === ids.checks.ping).map(item => item.id),
		title: "DNS Reachability",
		description: "Network reachability from all regions",
		sortOrder: 1,
		displayMode: "latency",
		chartMode: "inherit",
		checkName: "Cloudflare DNS",
		checkType: "ping",
		checkTarget: "1.1.1.1",
		checkIntervalSeconds: 60,
		createdAt,
		updatedAt: now
	}
];

const chartStart = Date.parse("2026-08-12T04:00:00.000Z");
const chartPoints = (base, variance = 6) =>
	Array.from({ length: 25 }, (_, index) => [chartStart + index * 60 * 60 * 1000, Number((base + Math.sin(index / 2.4) * variance + (index % 5 === 0 ? variance * 0.8 : 0)).toFixed(2))]);
const series = (name, unit, points) => ({ name, labels: { probeId: ids.probes.tokyo, checkId: ids.checks.ping }, unit, points });
const meta = { from: chartStart, to: chartStart + 24 * 60 * 60 * 1000, maxDataPoints: 600, source: "aggregate", resolution: "1h", totalPoints: 25 };

const pingSeries = {
	series: {
		latency_avg: series("latency_avg", "ms", chartPoints(31, 7)),
		latency_min: series("latency_min", "ms", chartPoints(22, 4)),
		latency_max: series("latency_max", "ms", chartPoints(48, 13)),
		loss_percent: series(
			"loss_percent",
			"%",
			chartPoints(1.2, 0.9).map(([timestamp, value], index) => [timestamp, index === 17 ? 8.4 : Math.max(0, value)])
		)
	},
	meta
};

const tcpSeries = {
	series: {
		connect_avg: series("connect_avg", "ms", chartPoints(46, 9)),
		connect_min: series("connect_min", "ms", chartPoints(31, 5)),
		connect_max: series("connect_max", "ms", chartPoints(71, 15)),
		failure_percent: series(
			"failure_percent",
			"%",
			chartPoints(0.4, 0.35).map(([timestamp, value]) => [timestamp, Math.max(0, value)])
		)
	},
	meta
};

const httpSeries = {
	series: {
		dns_avg: series("dns_avg", "ms", chartPoints(12, 3)),
		connect_avg: series("connect_avg", "ms", chartPoints(28, 6)),
		tls_avg: series("tls_avg", "ms", chartPoints(41, 8)),
		ttfb_avg: series("ttfb_avg", "ms", chartPoints(86, 14)),
		total_avg: series("total_avg", "ms", chartPoints(178, 24)),
		failure_percent: series(
			"failure_percent",
			"%",
			chartPoints(0.5, 0.4).map(([timestamp, value], index) => [timestamp, index === 18 ? 4.2 : Math.max(0, value)])
		)
	},
	meta
};

const latestHttpResults = probes.slice(0, 3).map((currentProbe, index) => ({
	probeId: currentProbe.id,
	checkId: ids.checks.http,
	result: {
		startedAt: "2026-08-13T03:19:00.000Z",
		finishedAt: "2026-08-13T03:19:00.186Z",
		durationMs: 186 + index * 14,
		status: "successful",
		dnsDurationMs: 12 + index,
		connectDurationMs: 29 + index * 2,
		tlsDurationMs: 41 + index * 3,
		ttfbDurationMs: 91 + index * 4,
		resolvedIp: "192.0.2.80",
		ipFamily: "inet",
		statusCode: 200,
		finalUrl: "https://example.com/health",
		redirectCount: 0,
		responseBytes: 512,
		responseTruncated: false,
		bodyMatched: true,
		tlsVersion: "TLS 1.3",
		tlsCipherSuite: "TLS_AES_128_GCM_SHA256",
		certificateNotBefore: "2026-06-01T00:00:00.000Z",
		certificateNotAfter: "2026-11-28T23:59:59.000Z"
	}
}));

const tracerouteHops = [
	[1, "192.0.2.1", "office-gateway.example", 1.8],
	[2, "192.0.2.2", "regional-edge.example", 6.4],
	[3, "198.51.100.8", "transit-1.example", 18.9],
	[4, "198.51.100.18", "transit-2.example", 28.2],
	[5, "203.0.113.53", "one.one.one.one", 32.7]
].map(([hopIndex, address, hostname, rttAvgMs]) => ({
	hopIndex,
	address,
	hostname,
	sentCount: 3,
	receivedCount: 3,
	lossPercent: 0,
	rttMinMs: rttAvgMs - 0.8,
	rttAvgMs,
	rttMedianMs: rttAvgMs,
	rttMaxMs: rttAvgMs + 1.2,
	rttStddevMs: 0.4,
	rttSamplesMs: [rttAvgMs - 0.8, rttAvgMs, rttAvgMs + 1.2]
}));

const tracerouteRuns = Array.from({ length: 6 }, (_, index) => ({
	startedAt: new Date(Date.parse("2026-08-13T03:15:00.000Z") - index * 15 * 60 * 1000).toISOString(),
	finishedAt: new Date(Date.parse("2026-08-13T03:15:04.000Z") - index * 15 * 60 * 1000).toISOString(),
	durationMs: 4000,
	status: "successful",
	resolvedIp: "203.0.113.53",
	ipFamily: "inet",
	destinationReached: true,
	hopCount: tracerouteHops.length,
	hops: tracerouteHops
}));

const topologyNodes = [
	{ id: `probe:${ids.probes.taipei}`, kind: "probe", label: "Taipei Office", probeId: ids.probes.taipei, seenCount: 24 },
	...tracerouteHops.map(hop => ({
		id: `ip:${hop.address}`,
		kind: hop.hopIndex === tracerouteHops.length ? "destination" : "hop",
		label: hop.hostname,
		address: hop.address,
		hostname: hop.hostname,
		hopIndex: hop.hopIndex,
		seenCount: 24,
		avgRttMs: hop.rttAvgMs,
		lossPercent: hop.lossPercent
	}))
];
const topologyEdges = topologyNodes.slice(1).map((node, index) => ({
	id: `${topologyNodes[index].id}->${node.id}`,
	source: topologyNodes[index].id,
	target: node.id,
	seenCount: 24,
	avgRttMs: node.avgRttMs,
	lossPercent: node.lossPercent
}));

const sessions = [
	{
		id: "13000000-0000-4000-8000-000000000001",
		userAgent: "Mozilla/5.0 (Macintosh; Apple Silicon) AppleWebKit/537.36 Chrome/140.0 Safari/537.36",
		createdAt: "2026-08-13T01:10:00.000Z",
		lastUsedAt: now,
		idleExpiresAt: "2026-08-20T03:20:00.000Z",
		absoluteExpiresAt: "2026-09-12T01:10:00.000Z",
		authenticatedAt: "2026-08-13T01:10:00.000Z",
		authenticationMethod: "password",
		isCurrent: true
	},
	{
		id: "13000000-0000-4000-8000-000000000002",
		userAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/139.0 Safari/537.36",
		createdAt: "2026-08-11T05:00:00.000Z",
		lastUsedAt: "2026-08-12T14:30:00.000Z",
		idleExpiresAt: "2026-08-19T14:30:00.000Z",
		absoluteExpiresAt: "2026-09-10T05:00:00.000Z",
		authenticatedAt: "2026-08-11T05:00:00.000Z",
		authenticationMethod: "github",
		isCurrent: false
	}
];

const apiTokens = [
	{
		id: "14000000-0000-4000-8000-000000000001",
		name: "Docs automation",
		tokenHint: "docsdemo",
		scopes: ["projects:read", "probes:read", "checks:read", "results:read"],
		createdAt,
		lastUsedAt: "2026-08-12T18:00:00.000Z",
		expiresAt: "2026-11-11T09:00:00.000Z"
	}
];

const managedUsers = [
	{ ...user, createdAt, updatedAt: now, grantedAt: createdAt },
	{ id: "10000000-0000-4000-8000-000000000002", email: "sam@example.com", displayName: "Sam Rivera", emailVerified: true, isSystemAdmin: false, hasPassword: true, createdAt, updatedAt: now },
	{ id: "10000000-0000-4000-8000-000000000003", email: "taylor@example.com", displayName: "Taylor Chen", emailVerified: true, isSystemAdmin: false, hasPassword: false, createdAt, updatedAt: now },
	{ id: "10000000-0000-4000-8000-000000000004", email: "viewer@example.com", displayName: "Jordan Lee", emailVerified: false, isSystemAdmin: false, hasPassword: true, createdAt, updatedAt: now }
];

const publicElements = [
	{
		id: statusElements[0].id,
		kind: "folder",
		title: "Core Services",
		description: "Customer-facing endpoints",
		status: "degraded",
		displayMode: "status",
		children: [
			{
				id: statusElements[1].id,
				kind: "assignment_group",
				title: "HTTPS API",
				description: "Public health endpoint",
				type: "http",
				target: "https://example.com/health",
				status: "degraded",
				latestStartedAt: "2026-08-13T03:19:00.000Z",
				latestStatus: "error",
				displayMode: "history",
				chartMode: "compact",
				chartRange: "24h",
				assignmentCount: 3,
				successfulAssignments: 2,
				failingAssignments: 1,
				staleAssignments: 0,
				metrics: { successRate: 99.7, averageLatencyMs: 184.2 },
				chart: { range: "24h", series: [httpSeries.series.total_avg] }
			},
			{
				id: statusElements[2].id,
				kind: "assignment_group",
				title: "DNS Reachability",
				description: "Network reachability from all regions",
				type: "ping",
				target: "1.1.1.1",
				status: "operational",
				latestStartedAt: "2026-08-13T03:19:00.000Z",
				latestStatus: "successful",
				displayMode: "latency",
				chartMode: "compact",
				chartRange: "24h",
				assignmentCount: 3,
				successfulAssignments: 3,
				failingAssignments: 0,
				staleAssignments: 0,
				metrics: { successRate: 99.9, averageLatencyMs: 31.4, packetLossPercent: 0.8 },
				chart: { range: "24h", series: [pingSeries.series.latency_avg] }
			}
		]
	}
];

const responseForGet = pathname => {
	if (pathname === "/api/v1/system/config") return { demoMode: false, capabilities: { accountCreationEnabled: true, projectCreationEnabled: true, credentialChangesEnabled: true } };
	if (pathname === "/api/v1/healthz") return { status: "ok" };
	if (pathname === "/api/v1/") return { name: "Netstamp API", version: "1.4.0", status: "ok" };
	if (pathname === "/api/v1/auth/methods") return { password: true, providers: [{ id: "github", name: "GitHub" }] };
	if (pathname === "/api/v1/auth/me") return { authenticated: true, user };
	if (pathname === "/api/v1/auth/sessions") return { sessions };
	if (pathname === "/api/v1/auth/api-tokens") return { tokens: apiTokens };
	if (pathname === "/api/v1/auth/sudo") return { active: true, expiresAt: "2026-08-13T03:35:00.000Z", methods: ["password", "github"] };
	if (pathname === "/api/v1/users/me/authentication-methods")
		return {
			hasPassword: true,
			identities: [{ id: "15000000-0000-4000-8000-000000000001", provider: "github", providerUserId: "docs-demo", email: "alex@example.com", displayName: "Alex Morgan", createdAt }]
		};
	if (pathname === "/api/v1/projects") return { projects: [project] };
	if (pathname === "/api/v1/me/project-invites") return { invites: [] };
	if (/^\/api\/v1\/projects\/[^/]+$/.test(pathname)) return { project };
	if (pathname.endsWith("/probes")) return { probes };
	if (pathname.endsWith("/probes/30000000-0000-4000-8000-000000000099")) {
		return {
			probe: {
				...probes[0],
				id: "30000000-0000-4000-8000-000000000099",
				name: "Taipei Office 2",
				status: { ...probes[0].status, probeId: "30000000-0000-4000-8000-000000000099" }
			}
		};
	}
	if (/\/probes\/[^/]+$/.test(pathname)) return { probe: probes.find(item => pathname.endsWith(item.id)) ?? probes[0] };
	if (pathname.endsWith("/checks")) return { checks, canManageChecks: true };
	if (/\/checks\/[^/]+$/.test(pathname)) return { check: checks.find(item => pathname.endsWith(item.id)) ?? checks[0] };
	if (pathname.endsWith("/assignments")) return { assignments };
	if (pathname.endsWith("/labels")) return { labels };
	if (pathname.endsWith("/members")) return { members };
	if (pathname.endsWith("/invites")) return { invites };
	if (pathname.endsWith("/alerts/rules")) return { rules };
	if (pathname.endsWith("/alerts/incidents")) return { incidents };
	if (/\/alerts\/incidents\/[^/]+$/.test(pathname)) return { incident: incidents.find(item => pathname.endsWith(item.id)) ?? incidents[0] };
	if (pathname.endsWith("/alerts/notifications")) return { notifications };
	if (pathname.endsWith("/status-pages")) return { pages: [statusPage] };
	if (pathname.endsWith(`/status-pages/${pageId}`)) return { page: statusPage, elements: statusElements };
	if (pathname.endsWith("/results/latest"))
		return { results: assignments.map(item => ({ type: item.check.type, probeId: item.probeId, checkId: item.checkId, latestStartedAt: "2026-08-13T03:19:00.000Z", latestStatus: "successful" })) };
	if (pathname.endsWith("/results/http/latest")) return { results: latestHttpResults };
	if (pathname.endsWith("/results/ping/insight")) return { summary: { averageRttMs: 31.4, maxRttMs: 67.2, lossPercent: 0.8, successRate: 99.2, samples: 1427 }, meta };
	if (pathname.endsWith("/results/ping/series")) return pingSeries;
	if (pathname.endsWith("/results/tcp/insight")) return { summary: { averageConnectMs: 46.8, maxConnectMs: 92.4, failurePercent: 0.3, successRate: 99.7, samples: 1430 }, meta };
	if (pathname.endsWith("/results/tcp/series")) return tcpSeries;
	if (pathname.endsWith("/results/http/insight"))
		return { summary: { averageTotalMs: 184.2, maxTotalMs: 288.4, averageTtfbMs: 91.5, maxTtfbMs: 142.1, failurePercent: 0.5, successRate: 99.5, certificateDaysRemaining: 107, samples: 1429 }, meta };
	if (pathname.endsWith("/results/http/series")) return httpSeries;
	if (pathname.endsWith("/results/traceroute/insight"))
		return {
			points: chartPoints(32, 4).map(([timestampMs, finalRttAvgMs], index) => ({
				timestampMs,
				bucketFromMs: timestampMs,
				bucketToMs: timestampMs + 60 * 60 * 1000,
				resultCount: 4,
				finalRttAvgMs,
				hasLoss: index === 18,
				hasRouteChange: index === 12,
				destinationReached: true
			})),
			query: { from: meta.from, to: meta.to, maxDataPoints: 600, resolution: "bucket", totalRuns: 96 }
		};
	if (pathname.endsWith("/results/traceroute/runs")) return { runs: tracerouteRuns, query: { from: meta.from, to: meta.to, limit: 200 } };
	if (pathname.endsWith("/results/traceroute/topology")) return { nodes: topologyNodes, edges: topologyEdges, query: { from: meta.from, to: meta.to, limit: 100 } };
	if (pathname === "/api/v1/public/status-pages/docs-demo/editor-context") return { projectRef: "docs-demo", pageId };
	if (pathname === "/api/v1/public/status-pages/docs-demo/summary")
		return {
			page: {
				id: pageId,
				slug: statusPage.slug,
				title: statusPage.title,
				description: statusPage.description,
				status: "degraded",
				footerText: statusPage.footerText,
				theme: "light",
				showIncidentHistory: true,
				showGeneratedAt: true,
				defaultChartMode: "compact",
				defaultChartRange: "24h",
				updatedAt: now
			},
			generatedAt: now
		};
	if (pathname === "/api/v1/public/status-pages/docs-demo/elements") return { elements: publicElements, generatedAt: now };
	if (pathname === "/api/v1/public/status-pages/docs-demo/incidents")
		return {
			incidents: {
				active: [
					{
						id: ids.incidents.open,
						checkTitle: "HTTPS API",
						status: "open",
						severity: "critical",
						openedAt: "2026-08-13T02:48:00.000Z",
						lastTriggeredAt: "2026-08-13T03:17:00.000Z",
						summary: { state: "firing", metric: "http.failure_percent", value: 100 }
					}
				],
				recentResolved: [
					{
						id: ids.incidents.resolved,
						checkTitle: "HTTPS API",
						status: "resolved",
						severity: "warning",
						openedAt: "2026-08-12T22:10:00.000Z",
						resolvedAt: "2026-08-12T22:32:00.000Z",
						lastTriggeredAt: "2026-08-12T22:30:00.000Z"
					}
				]
			},
			generatedAt: now
		};
	if (/\/public\/status-pages\/docs-demo\/elements\/[^/]+\/chart$/.test(pathname)) return { chart: { range: "24h", series: [httpSeries.series.total_avg] }, generatedAt: now };
	if (/\/public\/status-pages\/docs-demo\/elements\/[^/]+\/daily-status$/.test(pathname))
		return {
			range: "30d",
			days: Array.from({ length: 30 }, (_, index) => ({
				date: new Date(Date.parse("2026-08-13T00:00:00.000Z") - index * 86400000).toISOString().slice(0, 10),
				status: index === 3 ? "degraded" : "operational",
				successRate: index === 3 ? 97.8 : 99.9
			})),
			generatedAt: now
		};
	if (pathname === "/api/v1/admin/settings/access")
		return { settings: { accountCreationEnabled: true, emailVerificationRequired: true, projectCreationEnabled: true, credentialChangesEnabled: true } };
	if (pathname === "/api/v1/admin/settings/smtp")
		return {
			settings: { host: "smtp.example.com", port: 587, username: "netstamp", passwordSet: true, from: "Netstamp <alerts@example.com>", tlsMode: "starttls", timeoutSeconds: 10, configured: true }
		};
	if (pathname === "/api/v1/admin/settings/authentication-providers/oidc")
		return {
			settings: {
				enabled: true,
				issuerUrl: "https://identity.example.com",
				clientId: "docs-client-id",
				clientSecretSet: true,
				displayName: "Company SSO",
				jitEnabled: true,
				callbackUrl: "https://app.netstamp.dev/api/v1/auth/external/oidc/callback"
			}
		};
	if (pathname === "/api/v1/admin/settings/authentication-providers/google")
		return {
			settings: {
				enabled: true,
				clientId: "docs-client-id",
				clientSecretSet: true,
				displayName: "Google",
				jitEnabled: true,
				callbackUrl: "https://app.netstamp.dev/api/v1/auth/external/google/callback",
				allowedDomains: ["example.com"]
			}
		};
	if (pathname === "/api/v1/admin/settings/authentication-providers/github")
		return {
			settings: {
				enabled: true,
				clientId: "docs-client-id",
				clientSecretSet: true,
				displayName: "GitHub",
				jitEnabled: true,
				callbackUrl: "https://app.netstamp.dev/api/v1/auth/external/github/callback",
				allowSignup: false
			}
		};
	if (pathname === "/api/v1/admin/system-admins") return { admins: [{ id: userId, email: user.email, displayName: user.displayName, emailVerified: true, grantedAt: createdAt }] };
	if (pathname === "/api/v1/admin/users") return { users: managedUsers };
	return undefined;
};

const responseForMutation = (pathname, method) => {
	if (pathname === "/api/v1/auth/csrf" && method === "GET") return { csrfToken: "docs-example-not-valid-csrf" };
	if (pathname.endsWith("/selector-previews") && method === "POST")
		return {
			selector: { label: { key: "environment", op: "eq", value: "production" } },
			matchedCount: 3,
			probes: probes.slice(0, 3).map(({ id, projectId: currentProjectId, name, enabled, labels: currentLabels }) => ({ id, projectId: currentProjectId, name, enabled, labels: currentLabels }))
		};
	if (pathname.endsWith("/probes") && method === "POST")
		return { probe: { ...probes[0], id: "30000000-0000-4000-8000-000000000099", name: "Taipei Office 2", status: undefined }, secret: "docs-example-not-valid-probe-secret" };
	if (pathname === "/api/v1/auth/api-tokens" && method === "POST")
		return {
			token: { id: "14000000-0000-4000-8000-000000000099", name: "Docs API", tokenHint: "docsdemo", scopes: ["projects:read", "results:read"], createdAt: now, expiresAt: "2026-11-11T09:00:00.000Z" },
			value: "docs-example-not-valid-api-token"
		};
	return undefined;
};

export const fixtureResponse = (url, method) => {
	const { pathname } = new URL(url);
	const normalizedMethod = method.toUpperCase();
	const data =
		normalizedMethod === "GET" || normalizedMethod === "HEAD" ? (responseForGet(pathname) ?? responseForMutation(pathname, normalizedMethod)) : responseForMutation(pathname, normalizedMethod);
	return data === undefined ? null : { status: normalizedMethod === "POST" ? 201 : 200, body: data };
};

export const fixtureSummary = { project, user, probes, checks, assignments, labels, members, invites, notifications, rules, incidents, statusPage, statusElements };
