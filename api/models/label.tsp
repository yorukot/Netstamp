namespace Netstamp;

@example(#{
	id: "66666666-6666-6666-6666-666666666666",
	projectId: "22222222-2222-2222-2222-222222222222",
	key: "region",
	value: "tokyo",
	createdAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
	updatedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
})
model Label {
	id: uuid;
	projectId: uuid;

	@minLength(1)
	@maxLength(64)
	key: string;

	@minLength(1)
	@maxLength(64)
	value: string;

	createdAt: utcDateTime;
	updatedAt: utcDateTime;
}

@example(#{
	label: #{
		id: "66666666-6666-6666-6666-666666666666",
		projectId: "22222222-2222-2222-2222-222222222222",
		key: "region",
		value: "tokyo",
		createdAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
		updatedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
	},
})
model LabelResponse {
	label: Label;
}

@example(#{
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
model LabelListResponse {
	labels: Label[];
}

@example(#{ key: "region", value: "tokyo" })
model CreateLabelRequest {
	@minLength(1)
	@maxLength(64)
	key: string;

	@minLength(1)
	@maxLength(64)
	value: string;
}

@example(#{ key: "metro", value: "tokyo" })
model UpdateLabelRequest {
	@minLength(1)
	@maxLength(64)
	key?: string;

	@minLength(1)
	@maxLength(64)
	value?: string;
}
