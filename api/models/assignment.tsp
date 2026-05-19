import "@typespec/http";

using Http;

namespace Netstamp;

@example(#{ selector: #{} })
model SelectorPreviewRequest {
	@doc("Selector AST to validate and preview against active project probes. Empty or omitted selectors match all probes.")
	selector?: Record<unknown>;
}

@example(#{
	id: "33333333-3333-3333-3333-333333333333",
	projectId: "22222222-2222-2222-2222-222222222222",
	name: "tokyo-vps-1",
	enabled: true,
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
})
model SelectorPreviewProbe {
	id: uuid;
	projectId: uuid;

	@minLength(1)
	@maxLength(128)
	name: string;

	enabled: boolean;
	labels: Label[];
}

@example(#{
	selector: #{},
	matchedCount: 1,
	probes: #[#{ id: "33333333-3333-3333-3333-333333333333", projectId: "22222222-2222-2222-2222-222222222222", name: "tokyo-vps-1", enabled: true, labels: #[] }],
})
model SelectorPreviewResponse {
	@doc("Backend-canonical selector AST.")
	selector: Record<unknown>;

	@minValue(0)
	matchedCount: int32;

	probes: SelectorPreviewProbe[];
}

@example(#{
	id: "77777777-7777-7777-7777-777777777777",
	projectId: "22222222-2222-2222-2222-222222222222",
	probeId: "33333333-3333-3333-3333-333333333333",
	checkId: "44444444-4444-4444-4444-444444444444",
	checkVersion: "01HXYZ7K6P4E4M3F0CB1W9DYJ5",
	selectorVersion: "01HXYZ7K6P4E4M3F0CB1W9DYJ6",
})
model ProjectAssignment {
	id: uuid;
	projectId: uuid;
	probeId: uuid;
	checkId: uuid;
	checkVersion: string;
	selectorVersion: string;
	probe?: Probe;
	check?: Check;
}

@example(#{
	assignments: #[
		#{
			id: "77777777-7777-7777-7777-777777777777",
			projectId: "22222222-2222-2222-2222-222222222222",
			probeId: "33333333-3333-3333-3333-333333333333",
			checkId: "44444444-4444-4444-4444-444444444444",
			checkVersion: "01HXYZ7K6P4E4M3F0CB1W9DYJ5",
			selectorVersion: "01HXYZ7K6P4E4M3F0CB1W9DYJ6",
		}
	],
})
model ProjectAssignmentListResponse {
	assignments: ProjectAssignment[];
}

model ProjectAssignmentQuery {
	@query
	probeId?: uuid;

	@query
	checkId?: uuid;
}
