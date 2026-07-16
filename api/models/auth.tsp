import "@typespec/http";

using Http;

namespace Netstamp;

@example(#{ id: "11111111-1111-1111-1111-111111111111", email: "user@example.com", displayName: "Jane Doe", emailVerified: true, isSystemAdmin: true, hasPassword: true })
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

	@doc("Whether the account has a local password credential.")
	hasPassword: boolean;
}

model AuthSessionIdPathParam {
	@path
	@doc("Auth session UUID.")
	session_id: uuid;
}

@example(#{
	id: "22222222-2222-2222-2222-222222222222",
	userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/140.0 Safari/537.36",
	createdAt: utcDateTime.fromISO("2026-07-10T08:00:00Z"),
	lastUsedAt: utcDateTime.fromISO("2026-07-10T09:15:00Z"),
	idleExpiresAt: utcDateTime.fromISO("2026-07-17T09:15:00Z"),
	absoluteExpiresAt: utcDateTime.fromISO("2026-08-09T08:00:00Z"),
	authenticatedAt: utcDateTime.fromISO("2026-07-10T08:00:00Z"),
	authenticationMethod: "password",
	isCurrent: true,
})
model AuthSession {
	@doc("Server-side auth session UUID.")
	id: uuid;

	@maxLength(512)
	@doc("User-Agent captured when the session was created. Empty for sessions created before User-Agent tracking was enabled.")
	userAgent: string;

	createdAt: utcDateTime;
	lastUsedAt: utcDateTime;
	idleExpiresAt: utcDateTime;
	absoluteExpiresAt: utcDateTime;
	authenticatedAt: utcDateTime;
	authenticationMethod: "password" | "oidc";

	@doc("Whether this is the session authenticating the current request.")
	isCurrent: boolean;
}

@example(#{
	sessions: #[
		#{
			id: "22222222-2222-2222-2222-222222222222",
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/140.0 Safari/537.36",
			createdAt: utcDateTime.fromISO("2026-07-10T08:00:00Z"),
			lastUsedAt: utcDateTime.fromISO("2026-07-10T09:15:00Z"),
			idleExpiresAt: utcDateTime.fromISO("2026-07-17T09:15:00Z"),
			absoluteExpiresAt: utcDateTime.fromISO("2026-08-09T08:00:00Z"),
			authenticatedAt: utcDateTime.fromISO("2026-07-10T08:00:00Z"),
			authenticationMethod: "password",
			isCurrent: true,
		}
	],
})
model AuthSessionsResponse {
	sessions: AuthSession[];
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

@example(#{ user: #{ id: "11111111-1111-1111-1111-111111111111", email: "user@example.com", displayName: "Jane Doe", emailVerified: true, isSystemAdmin: true, hasPassword: true } })
model AuthUserResponse {
	user: User;
}

@example(#{
	authenticated: true,
	user: #{ id: "11111111-1111-1111-1111-111111111111", email: "user@example.com", displayName: "Jane Doe", emailVerified: true, isSystemAdmin: true, hasPassword: true },
})
model CurrentUserResponse {
	authenticated: true;
	user: User;
}

@example(#{ csrfToken: "YNo5Uoj64VqK5jCht8h17GO8sHJIY0ScDNef7X99w7k" })
model CSRFTokenResponse {
	@secret
	@doc("One-time readable CSRF token to send in X-CSRF-Token for state-changing session-authenticated requests.")
	csrfToken: string;
}

model AuthPasswordMethod {
	enabled: boolean;
	registrationEnabled: boolean;
}

model AuthOIDCMethod {
	enabled: boolean;
	displayName: string;
}

model AuthMethodsResponse {
	password: AuthPasswordMethod;
	oidc: AuthOIDCMethod;
}

model SudoStatusResponse {
	active: boolean;
	expiresAt: utcDateTime;
	methods: ("password" | "oidc")[];
}

model PasswordSudoRequest {
	@secret
	@minLength(8)
	@maxLength(128)
	password: string;
}

model OIDCStartQuery {
	@query intent?: "login" | "sudo" | "link";
	@query returnTo?: string;
}

model OIDCCallbackQuery {
	@query code: string;
	@query state: string;
}
