namespace Netstamp;

@example(#{ id: "11111111-1111-1111-1111-111111111111", email: "user@example.com", displayName: "Jane Doe", emailVerified: true, isSystemAdmin: true })
model User {
	@doc("User UUID.")
	id: uuid;

	@doc("Normalized email address used to sign in.")
	email: email;

	@minLength(1)
	@maxLength(64)
	@doc("Name shown in the app.")
	displayName: string;

	@doc("Whether the sign-in email address has been verified.")
	emailVerified: boolean;

	@doc("Whether the user has instance-level administrator access.")
	isSystemAdmin: boolean;
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

@example(#{ email: "user@example.com" })
model CreatePasswordResetRequest {
	@doc("Email address for the account that should receive a password reset link.")
	email: email;
}

@example(#{ token: "YNo5Uoj64VqK5jCht8h17GO8sHJIY0ScDNef7X99w7k", newPassword: "correct-horse-battery-staple" })
model ConfirmPasswordResetRequest {
	@secret
	@minLength(1)
	@maxLength(512)
	@doc("One-time reset token from the password reset email.")
	token: string;

	@secret
	@minLength(8)
	@maxLength(128)
	@doc("New password. It is stored only as an Argon2id hash.")
	newPassword: string;
}

@example(#{ email: "user@example.com" })
model CreateEmailVerificationRequest {
	@doc("Email address for the account that should receive a new verification link.")
	email: email;
}

@example(#{ token: "YNo5Uoj64VqK5jCht8h17GO8sHJIY0ScDNef7X99w7k" })
model ConfirmEmailVerificationRequest {
	@secret
	@minLength(1)
	@maxLength(512)
	@doc("One-time verification token from the email verification message.")
	token: string;
}

@example(#{ emailVerificationRequired: true })
model RegisterEmailVerificationRequiredResponse {
	emailVerificationRequired: true;
}

@example(#{ user: #{ id: "11111111-1111-1111-1111-111111111111", email: "user@example.com", displayName: "Jane Doe", emailVerified: true, isSystemAdmin: true } })
model AuthUserResponse {
	user: User;
}

@example(#{ authenticated: true, user: #{ id: "11111111-1111-1111-1111-111111111111", email: "user@example.com", displayName: "Jane Doe", emailVerified: true, isSystemAdmin: true } })
model CurrentUserResponse {
	authenticated: true;
	user: User;
}
