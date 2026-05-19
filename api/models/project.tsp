namespace Netstamp;

alias ProjectRole = "owner" | "admin" | "editor" | "viewer";

alias AssignableProjectRole = "admin" | "editor" | "viewer";

@example(#{
	id: "22222222-2222-2222-2222-222222222222",
	name: "Engineering",
	slug: "engineering",
	createdByUserId: "11111111-1111-1111-1111-111111111111",
	createdAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
	updatedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
})
model Project {
	id: uuid;

	@minLength(1)
	@maxLength(64)
	name: string;

	@minLength(1)
	@maxLength(64)
	@pattern("^[a-z0-9-]+$")
	slug: string;

	createdByUserId: uuid;
	createdAt: utcDateTime;
	updatedAt: utcDateTime;
}

@example(#{
	project: #{
		id: "22222222-2222-2222-2222-222222222222",
		name: "Engineering",
		slug: "engineering",
		createdByUserId: "11111111-1111-1111-1111-111111111111",
		createdAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
		updatedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
	},
})
model ProjectResponse {
	project: Project;
}

@example(#{
	projects: #[
		#{
			id: "22222222-2222-2222-2222-222222222222",
			name: "Engineering",
			slug: "engineering",
			createdByUserId: "11111111-1111-1111-1111-111111111111",
			createdAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
			updatedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
		}
	],
})
model ProjectListResponse {
	projects: Project[];
}

@example(#{ id: "11111111-1111-1111-1111-111111111111", email: "user@example.com", displayName: "Jane Doe" })
model ProjectMemberUser {
	id: uuid;
	email: email;

	@minLength(1)
	@maxLength(64)
	displayName: string;
}

@example(#{
	id: "55555555-5555-5555-5555-555555555555",
	projectId: "22222222-2222-2222-2222-222222222222",
	userId: "11111111-1111-1111-1111-111111111111",
	role: "owner",
	user: #{ id: "11111111-1111-1111-1111-111111111111", email: "user@example.com", displayName: "Jane Doe" },
	createdAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
	updatedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
})
model ProjectMember {
	id: uuid;
	projectId: uuid;
	userId: uuid;
	role: ProjectRole;
	user: ProjectMemberUser;
	createdAt: utcDateTime;
	updatedAt: utcDateTime;
}

@example(#{
	member: #{
		id: "55555555-5555-5555-5555-555555555555",
		projectId: "22222222-2222-2222-2222-222222222222",
		userId: "11111111-1111-1111-1111-111111111111",
		role: "owner",
		user: #{ id: "11111111-1111-1111-1111-111111111111", email: "user@example.com", displayName: "Jane Doe" },
		createdAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
		updatedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
	},
})
model ProjectMemberResponse {
	member: ProjectMember;
}

@example(#{
	members: #[
		#{
			id: "55555555-5555-5555-5555-555555555555",
			projectId: "22222222-2222-2222-2222-222222222222",
			userId: "11111111-1111-1111-1111-111111111111",
			role: "owner",
			user: #{ id: "11111111-1111-1111-1111-111111111111", email: "user@example.com", displayName: "Jane Doe" },
			createdAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
			updatedAt: utcDateTime.fromISO("2026-05-13T10:00:00Z"),
		}
	],
})
model ProjectMemberListResponse {
	members: ProjectMember[];
}

@example(#{ name: "Engineering", slug: "engineering" })
model CreateProjectRequest {
	@minLength(1)
	@maxLength(64)
	name: string;

	@minLength(1)
	@maxLength(64)
	@pattern("^[a-z0-9-]+$")
	slug: string;
}

@example(#{ name: "Platform Engineering", slug: "platform-engineering" })
model UpdateProjectRequest {
	@minLength(1)
	@maxLength(64)
	name?: string;

	@minLength(1)
	@maxLength(64)
	@pattern("^[a-z0-9-]+$")
	slug?: string;
}

@example(#{ email: "user@example.com", role: "viewer" })
model AddProjectMemberRequest {
	email: email;
	role: AssignableProjectRole;
}

@example(#{ role: "editor" })
model UpdateProjectMemberRoleRequest {
	role: AssignableProjectRole;
}
