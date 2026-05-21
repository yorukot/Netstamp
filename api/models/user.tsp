namespace Netstamp;

@example(#{ displayName: "Jane Operator" })
@doc("Patch payload. At least one field must be provided.")
model UpdateCurrentUserRequest {
	@minLength(1)
	@maxLength(64)
	@doc("Name shown in the app.")
	displayName?: string;
}

@example(#{ newEmail: "jane.operator@example.com", password: "correct-horse-battery-staple" })
model ChangeCurrentUserEmailRequest {
	@doc("New normalized email address to use for sign-in.")
	newEmail: email;

	@secret
	@minLength(8)
	@maxLength(128)
	@doc("Current password used to confirm this sensitive account change.")
	password: string;
}

@example(#{ currentPassword: "correct-horse-battery-staple", newPassword: "new-correct-horse-battery-staple" })
model ChangeCurrentUserPasswordRequest {
	@secret
	@minLength(8)
	@maxLength(128)
	@doc("Current password to verify before changing credentials.")
	currentPassword: string;

	@secret
	@minLength(8)
	@maxLength(128)
	@doc("New password. It is stored only as an Argon2id hash.")
	newPassword: string;
}
