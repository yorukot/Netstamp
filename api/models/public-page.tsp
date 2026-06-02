namespace Netstamp;

alias PublicProbeStatus = "online" | "offline";

@example(#{
	id: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
	folderId: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
	name: "edge-api",
	description: "Public latency to the edge API.",
	intervalSeconds: 30,
	sortOrder: 0,
	createdAt: utcDateTime.fromISO("2026-05-28T10:00:00Z"),
	updatedAt: utcDateTime.fromISO("2026-05-28T10:00:00Z"),
})
model PublicPageCheck {
	id: uuid;
	folderId: uuid;

	@minLength(1)
	@maxLength(128)
	name: string;

	@minLength(1)
	@maxLength(1024)
	description?: string;

	@minValue(1)
	intervalSeconds: int32;

	@minValue(0)
	sortOrder: int32;

	createdAt: utcDateTime;
	updatedAt: utcDateTime;
}

@example(#{
	id: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
	name: "North America",
	description: "Public probes in North America.",
	sortOrder: 0,
	checks: #[],
	createdAt: utcDateTime.fromISO("2026-05-28T10:00:00Z"),
	updatedAt: utcDateTime.fromISO("2026-05-28T10:00:00Z"),
})
model PublicPageFolder {
	id: uuid;
	parentId?: uuid;

	@minLength(1)
	@maxLength(128)
	name: string;

	@minLength(1)
	@maxLength(1024)
	description?: string;

	@minValue(0)
	sortOrder: int32;

	checks?: PublicPageCheck[];
	createdAt: utcDateTime;
	updatedAt: utcDateTime;
}

@example(#{
	folderId: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
	probeId: "33333333-3333-3333-3333-333333333333",
	probeName: "iad-edge-1",
	probeLocationName: "Ashburn, VA",
	probeStatus: "online",
	checkId: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
	checkName: "edge-api",
	checkIntervalSeconds: 30,
})
model PublicPingPair {
	folderId: uuid;
	probeId: uuid;

	@minLength(1)
	@maxLength(128)
	probeName: string;

	probeLocationName?: string;
	probeStatus: PublicProbeStatus;
	checkId: uuid;

	@minLength(1)
	@maxLength(128)
	checkName: string;

	@minLength(1)
	@maxLength(1024)
	checkDescription?: string;

	@minValue(1)
	checkIntervalSeconds: int32;
}

@example(#{
	id: "99999999-9999-9999-9999-999999999999",
	slug: "public-edge",
	title: "Public Edge Latency",
	description: "Public ping insight for selected edge checks.",
	enabled: true,
	folders: #[],
	pairs: #[],
	createdAt: utcDateTime.fromISO("2026-05-28T10:00:00Z"),
	updatedAt: utcDateTime.fromISO("2026-05-28T10:00:00Z"),
})
model PublicPage {
	id: uuid;

	@minLength(1)
	@maxLength(64)
	@pattern("^[a-z0-9-]+$")
	slug: string;

	@minLength(1)
	@maxLength(128)
	title: string;

	@minLength(1)
	@maxLength(1024)
	description?: string;

	enabled: boolean;
	folders?: PublicPageFolder[];
	pairs?: PublicPingPair[];
	createdAt: utcDateTime;
	updatedAt: utcDateTime;
}

model PublicPageResponse {
	publicPage: PublicPage;
}

model PublicPageListResponse {
	publicPages: PublicPage[];
}

model PublicPageFolderResponse {
	folder: PublicPageFolder;
}

model PublicPageFolderChecksResponse {
	checks: PublicPageCheck[];
}

@example(#{ slug: "public-edge", title: "Public Edge Latency", description: "Public ping insight.", enabled: true })
model CreatePublicPageRequest {
	@minLength(1)
	@maxLength(64)
	@pattern("^[a-z0-9-]+$")
	slug: string;

	@minLength(1)
	@maxLength(128)
	title: string;

	@minLength(1)
	@maxLength(1024)
	description?: string;

	enabled?: boolean = true;
}

@example(#{ title: "Public Edge Latency", description: null, enabled: false })
model UpdatePublicPageRequest {
	@minLength(1)
	@maxLength(64)
	@pattern("^[a-z0-9-]+$")
	slug?: string;

	@minLength(1)
	@maxLength(128)
	title?: string;

	@minLength(1)
	@maxLength(1024)
	description?: string | null;

	enabled?: boolean;
}

@example(#{ parentId: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", name: "Ashburn", sortOrder: 0 })
model CreatePublicPageFolderRequest {
	parentId?: uuid;

	@minLength(1)
	@maxLength(128)
	name: string;

	@minLength(1)
	@maxLength(1024)
	description?: string;

	@minValue(0)
	sortOrder?: int32 = 0;
}

@example(#{ parentId: null, name: "US East", description: null, sortOrder: 1 })
model UpdatePublicPageFolderRequest {
	parentId?: uuid | null;

	@minLength(1)
	@maxLength(128)
	name?: string;

	@minLength(1)
	@maxLength(1024)
	description?: string | null;

	@minValue(0)
	sortOrder?: int32;
}

@example(#{ checkIds: #["aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"] })
model SetPublicPageFolderChecksRequest {
	checkIds: uuid[];
}

model PublicPingInsightSummary {
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

model PublicPingInsightResponse {
	summary: PublicPingInsightSummary;
	meta: PingSeriesQueryMetadata;
}
