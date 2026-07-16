import "@typespec/http";

using Http;

namespace Netstamp;

@example(#{ displayName: "Jane Operator" })
@doc("Patch payload. At least one field must be provided.")
model UpdateCurrentUserRequest {
	@minLength(1)
	@maxLength(64)
	@doc("Name shown in the app.")
	displayName?: string;
}

@example(#{ newEmail: "jane.operator@example.com" })
model ChangeCurrentUserEmailRequest {
	@doc("New normalized email address to use for sign-in.")
	newEmail: email;
}

@example(#{ newPassword: "new-correct-horse-battery-staple" })
model ChangeCurrentUserPasswordRequest {
	@secret
	@minLength(8)
	@maxLength(128)
	@doc("New password. It is stored only as an Argon2id hash.")
	newPassword: string;
}

model ExternalIdentity {
	id: uuid;
	provider: "google" | "github" | "oidc";
	issuer: string;
	email?: email;
	emailVerified: boolean;
	displayName?: string;
	username?: string;
	avatarUrl?: url;
	createdAt: utcDateTime;
	lastLoginAt?: utcDateTime;
}

model AuthenticationMethodsResponse {
	hasPassword: boolean;
	identities: ExternalIdentity[];
}

model IdentityIdPathParam {
	@path("identity_id") identityId: uuid;
}
