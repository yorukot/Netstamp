namespace Netstamp;

@example(#{ id: "11111111-1111-1111-1111-111111111111", email: "user@example.com", displayName: "Jane Doe" })
model User {
	@doc("User UUID.")
	id: uuid;

	@doc("Normalized email address used to sign in.")
	email: email;

	@minLength(1)
	@maxLength(64)
	@doc("Name shown in the app.")
	displayName: string;
}

@example(#{ email: "user@example.com", displayName: "Jane Doe", password: "correct-horse-battery-staple" })
model RegisterUserRequest {
	@doc("Email address used to sign in.")
	email: email;

	@minLength(1)
	@maxLength(64)
	@doc("Name shown in the app.")
	displayName: string;

	@secret
	@minLength(8)
	@maxLength(128)
	@doc("Plain-text password. It is stored only as an Argon2id hash.")
	password: string;
}

@example(#{ email: "user@example.com", password: "correct-horse-battery-staple" })
model LoginUserRequest {
	@doc("Email address used to sign in. It is normalized before lookup.")
	email: email;

	@secret
	@minLength(8)
	@maxLength(128)
	@doc("Plain-text password to verify. It is never returned by the API.")
	password: string;
}

@example(#{ user: #{ id: "11111111-1111-1111-1111-111111111111", email: "user@example.com", displayName: "Jane Doe" } })
model AuthUserResponse {
	user: User;
}

@example(#{ authenticated: true, user: #{ id: "11111111-1111-1111-1111-111111111111", email: "user@example.com", displayName: "Jane Doe" } })
model CurrentUserResponse {
	authenticated: true;
	user: User;
}
