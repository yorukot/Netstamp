namespace Netstamp;

alias ProjectRole = "owner" | "admin" | "editor" | "viewer";

alias AssignableProjectRole = "admin" | "editor" | "viewer";

alias ProjectInviteStatus = "pending" | "accepted" | "rejected";

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

@example(#{ id: "33333333-3333-3333-3333-333333333333", email: "member@example.com", displayName: "Pat Chen" })
model ProjectInviteUser {
	id?: uuid;
	email: email;

	@maxLength(64)
	displayName?: string;
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

@example(#{ id: "22222222-2222-2222-2222-222222222222", name: "Engineering", slug: "engineering" })
model ProjectInviteProject {
	id: uuid;

	@minLength(1)
	@maxLength(64)
	name: string;

	@minLength(1)
	@maxLength(64)
	@pattern("^[a-z0-9-]+$")
	slug: string;
}

@example(#{
	id: "77777777-7777-7777-7777-777777777777",
	projectId: "22222222-2222-2222-2222-222222222222",
	invitedEmail: "member@example.com",
	invitedUserId: "33333333-3333-3333-3333-333333333333",
	invitedByUserId: "11111111-1111-1111-1111-111111111111",
	role: "viewer",
	status: "pending",
	project: #{ id: "22222222-2222-2222-2222-222222222222", name: "Engineering", slug: "engineering" },
	invitedUser: #{ id: "33333333-3333-3333-3333-333333333333", email: "member@example.com", displayName: "Pat Chen" },
	invitedByUser: #{ id: "11111111-1111-1111-1111-111111111111", email: "owner@example.com", displayName: "Jane Doe" },
	createdAt: utcDateTime.fromISO("2026-05-24T10:00:00Z"),
	updatedAt: utcDateTime.fromISO("2026-05-24T10:00:00Z"),
})
model ProjectInvite {
	id: uuid;
	projectId: uuid;
	invitedEmail: email;
	invitedUserId?: uuid;
	invitedByUserId: uuid;
	role: AssignableProjectRole;
	status: ProjectInviteStatus;
	project: ProjectInviteProject;
	invitedUser: ProjectInviteUser;
	invitedByUser: ProjectMemberUser;
	createdAt: utcDateTime;
	updatedAt: utcDateTime;
	resolvedAt?: utcDateTime;
}

@example(#{
	invite: #{
		id: "77777777-7777-7777-7777-777777777777",
		projectId: "22222222-2222-2222-2222-222222222222",
		invitedEmail: "member@example.com",
		invitedUserId: "33333333-3333-3333-3333-333333333333",
		invitedByUserId: "11111111-1111-1111-1111-111111111111",
		role: "viewer",
		status: "pending",
		project: #{ id: "22222222-2222-2222-2222-222222222222", name: "Engineering", slug: "engineering" },
		invitedUser: #{ id: "33333333-3333-3333-3333-333333333333", email: "member@example.com", displayName: "Pat Chen" },
		invitedByUser: #{ id: "11111111-1111-1111-1111-111111111111", email: "owner@example.com", displayName: "Jane Doe" },
		createdAt: utcDateTime.fromISO("2026-05-24T10:00:00Z"),
		updatedAt: utcDateTime.fromISO("2026-05-24T10:00:00Z"),
	},
})
model ProjectInviteResponse {
	invite: ProjectInvite;
}

@example(#{
	invites: #[
		#{
			id: "77777777-7777-7777-7777-777777777777",
			projectId: "22222222-2222-2222-2222-222222222222",
			invitedEmail: "member@example.com",
			invitedUserId: "33333333-3333-3333-3333-333333333333",
			invitedByUserId: "11111111-1111-1111-1111-111111111111",
			role: "viewer",
			status: "pending",
			project: #{ id: "22222222-2222-2222-2222-222222222222", name: "Engineering", slug: "engineering" },
			invitedUser: #{ id: "33333333-3333-3333-3333-333333333333", email: "member@example.com", displayName: "Pat Chen" },
			invitedByUser: #{ id: "11111111-1111-1111-1111-111111111111", email: "owner@example.com", displayName: "Jane Doe" },
			createdAt: utcDateTime.fromISO("2026-05-24T10:00:00Z"),
			updatedAt: utcDateTime.fromISO("2026-05-24T10:00:00Z"),
		}
	],
})
model ProjectInviteListResponse {
	invites: ProjectInvite[];
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
@doc("Patch payload. At least one field must be provided.")
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
model CreateProjectInviteRequest {
	email: email;
	role: AssignableProjectRole;
}

@example(#{ role: "editor" })
model UpdateProjectMemberRoleRequest {
	role: AssignableProjectRole;
}
