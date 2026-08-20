namespace Netstamp;

alias SMTPTLSMode = "starttls" | "implicit" | "none";

@minLength(1)
scalar NonEmptyAdminSecret extends string;

@minLength(1)
@maxLength(253)
@pattern("^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)*$")
scalar GoogleHostedDomain extends string;

@example(#{ id: "11111111-1111-1111-1111-111111111111", email: "admin@example.com", displayName: "Jane Operator", emailVerified: true, grantedAt: utcDateTime.fromISO("2026-07-08T12:00:00Z") })
model SystemAdminUser {
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

	@doc("When the user received instance-level administrator access.")
	grantedAt: utcDateTime;
}

@example(#{ email: "admin@example.com" })
model GrantSystemAdminRequest {
	@doc("Existing user email to grant instance-level administrator access.")
	email: email;
}

@example(#{
	admin: #{ id: "11111111-1111-1111-1111-111111111111", email: "admin@example.com", displayName: "Jane Operator", emailVerified: true, grantedAt: utcDateTime.fromISO("2026-07-08T12:00:00Z") },
})
model SystemAdminResponse {
	admin: SystemAdminUser;
}

@example(#{
	admins: #[#{ id: "11111111-1111-1111-1111-111111111111", email: "admin@example.com", displayName: "Jane Operator", emailVerified: true, grantedAt: utcDateTime.fromISO("2026-07-08T12:00:00Z") }],
})
model SystemAdminsResponse {
	admins: SystemAdminUser[];
}

@example(#{
	id: "11111111-1111-1111-1111-111111111111",
	email: "operator@example.com",
	displayName: "Jane Operator",
	emailVerified: true,
	isSystemAdmin: false,
	hasPassword: true,
	createdAt: utcDateTime.fromISO("2026-07-08T12:00:00Z"),
	updatedAt: utcDateTime.fromISO("2026-07-08T12:00:00Z"),
})
model ManagedUser {
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

	@doc("When the user account was disabled. Omitted for active users.")
	disabledAt?: utcDateTime;

	@doc("Whether the user has instance-level administrator access.")
	isSystemAdmin: boolean;

	hasPassword: boolean;
	createdAt: utcDateTime;
	updatedAt: utcDateTime;

	@doc("When the user received instance-level administrator access.")
	grantedAt?: utcDateTime;
}

@example(#{
	users: #[
		#{
			id: "11111111-1111-1111-1111-111111111111",
			email: "operator@example.com",
			displayName: "Jane Operator",
			emailVerified: true,
			isSystemAdmin: false,
			hasPassword: true,
			createdAt: utcDateTime.fromISO("2026-07-08T12:00:00Z"),
			updatedAt: utcDateTime.fromISO("2026-07-08T12:00:00Z"),
		}
	],
})
model ManagedUsersResponse {
	users: ManagedUser[];
}

@example(#{
	user: #{
		id: "11111111-1111-1111-1111-111111111111",
		email: "operator@example.com",
		displayName: "Jane Operator",
		emailVerified: true,
		isSystemAdmin: false,
		hasPassword: true,
		createdAt: utcDateTime.fromISO("2026-07-08T12:00:00Z"),
		updatedAt: utcDateTime.fromISO("2026-07-08T12:00:00Z"),
	},
})
model ManagedUserResponse {
	user: ManagedUser;
}

@example(#{ disabled: true, systemAdmin: false })
@doc("Patch payload. At least one field should be provided.")
model UpdateManagedUserRequest {
	@doc("Disable or re-enable the account. Disabled accounts cannot sign in or access protected routes.")
	disabled?: boolean;

	@doc("Grant or revoke instance-level administrator access.")
	systemAdmin?: boolean;
}

@example(#{ password: "correct-horse-battery-staple" })
model SetManagedUserPasswordRequest {
	@secret
	@minLength(8)
	@maxLength(128)
	@doc("New password. It is stored only as an Argon2id hash.")
	password: string;
}

@example(#{ checkForUpdates: true })
model AdminUpdateSettings {
	@doc("Whether the controller checks the latest Netstamp GitHub release on startup and every six hours.")
	checkForUpdates: boolean;
}

@example(#{ checkForUpdates: false })
@doc("Partial update-check settings update. Omitted properties retain their current values.")
model AdminUpdateSettingsPatch {
	checkForUpdates?: boolean;
}

@example(#{ settings: #{ checkForUpdates: true } })
model AdminUpdateSettingsResponse {
	settings: AdminUpdateSettings;
}

@example(#{
	currentVersion: "0.1.0",
	latestVersion: "v0.2.0",
	updateAvailable: true,
	releaseUrl: "https://github.com/yorukot/netstamp/releases/tag/v0.2.0",
	publishedAt: utcDateTime.fromISO("2026-08-20T00:00:00Z"),
	lastCheckedAt: utcDateTime.fromISO("2026-08-20T06:00:00Z"),
	checkError: null,
})
model AdminUpdateStatus {
	currentVersion: string;
	latestVersion: string | null;
	updateAvailable: boolean;

	@format("uri")
	releaseUrl: string | null;

	publishedAt: utcDateTime | null;
	lastCheckedAt: utcDateTime | null;
	checkError: string | null;
}

@example(#{ accountCreationEnabled: true, emailVerificationRequired: true, projectCreationEnabled: true, credentialChangesEnabled: true })
model AdminAccessSettings {
	@doc("Whether new accounts may be created through password registration or external-provider JIT provisioning.")
	accountCreationEnabled: boolean;

	@doc("Whether newly created accounts must verify their email address.")
	emailVerificationRequired: boolean;

	@doc("Whether regular users may create projects.")
	projectCreationEnabled: boolean;

	@doc("Whether regular users may change account credentials or link and unlink external identities.")
	credentialChangesEnabled: boolean;
}

@example(#{ accountCreationEnabled: false, projectCreationEnabled: false })
@doc("Partial access-policy update. Omitted properties retain their current values.")
model AdminAccessSettingsPatch {
	accountCreationEnabled?: boolean;
	emailVerificationRequired?: boolean;
	projectCreationEnabled?: boolean;
	credentialChangesEnabled?: boolean;
}

@example(#{ settings: #{ accountCreationEnabled: true, emailVerificationRequired: true, projectCreationEnabled: true, credentialChangesEnabled: true } })
model AdminAccessSettingsResponse {
	settings: AdminAccessSettings;
}

@example(#{ host: "smtp.example.com", port: 587, username: "netstamp", passwordSet: true, from: "Netstamp <notifications@example.com>", tlsMode: "starttls", timeoutSeconds: 10, configured: true })
model AdminSMTPSettings {
	host: string;

	@minValue(1)
	@maxValue(65535)
	port: int32;

	username: string;
	passwordSet: boolean;
	from: string;
	tlsMode: SMTPTLSMode;

	@minValue(1)
	timeoutSeconds: int32;

	configured: boolean;
}

@example(#{ host: "smtp.example.com", port: 587, password: "replace-with-a-secret", from: "Netstamp <notifications@example.com>", tlsMode: "starttls", timeoutSeconds: 10 })
@doc("Partial SMTP update. Omitted properties retain their current values. Set password to null to clear it; an empty string is invalid.")
model AdminSMTPSettingsPatch {
	host?: string;

	@minValue(1)
	@maxValue(65535)
	port?: int32;

	username?: string;
	@secret password?: NonEmptyAdminSecret | null;
	from?: string;
	tlsMode?: SMTPTLSMode;

	@minValue(1)
	timeoutSeconds?: int32;
}

@example(#{
	settings: #{ host: "smtp.example.com", port: 587, username: "netstamp", passwordSet: true, from: "Netstamp <notifications@example.com>", tlsMode: "starttls", timeoutSeconds: 10, configured: true },
})
model AdminSMTPSettingsResponse {
	settings: AdminSMTPSettings;
}

model AdminExternalProviderSettingsBase {
	enabled: boolean;
	clientId: string;
	clientSecretSet: boolean;
	displayName: string;
	jitEnabled: boolean;

	@format("uri")
	@doc("OAuth callback URL derived from the server's configured public backend URL.")
	callbackUrl?: string;
}

@example(#{
	enabled: true,
	issuerUrl: "https://identity.example.com",
	clientId: "netstamp",
	clientSecretSet: true,
	displayName: "Company SSO",
	jitEnabled: true,
	callbackUrl: "https://netstamp.example.com/api/v1/auth/external/oidc/callback",
})
model AdminOIDCProviderSettings extends AdminExternalProviderSettingsBase {
	@format("uri")
	issuerUrl: string;
}

@example(#{
	enabled: true,
	clientId: "1234567890.apps.googleusercontent.com",
	clientSecretSet: true,
	displayName: "Google",
	jitEnabled: true,
	callbackUrl: "https://netstamp.example.com/api/v1/auth/external/google/callback",
	allowedDomains: #["example.com"],
})
model AdminGoogleProviderSettings extends AdminExternalProviderSettingsBase {
	@doc("Normalized Google Workspace domains allowed to authenticate. An empty list imposes no domain restriction.")
	allowedDomains: GoogleHostedDomain[];
}

@example(#{
	enabled: true,
	clientId: "github-client-id",
	clientSecretSet: true,
	displayName: "GitHub",
	jitEnabled: true,
	callbackUrl: "https://netstamp.example.com/api/v1/auth/external/github/callback",
	allowSignup: false,
})
model AdminGitHubProviderSettings extends AdminExternalProviderSettingsBase {
	@doc("Whether GitHub may display its account-signup option on the authorization page. This does not override Netstamp account-creation policy.")
	allowSignup: boolean;
}

model AdminExternalProviderSettingsPatchBase {
	enabled?: boolean;
	clientId?: string;

	@secret
	@doc("Replacement client secret. Omit to retain the current secret, or set to null to clear it. An empty string is invalid.")
	clientSecret?: NonEmptyAdminSecret | null;

	displayName?: string;
	jitEnabled?: boolean;
}

@example(#{ enabled: true, issuerUrl: "https://identity.example.com", clientId: "netstamp", clientSecret: "replace-with-a-secret", displayName: "Company SSO", jitEnabled: true })
@doc("Partial OIDC-provider update. Omitted properties retain their current values.")
model AdminOIDCProviderSettingsPatch extends AdminExternalProviderSettingsPatchBase {
	@format("uri")
	issuerUrl?: string;
}

@example(#{ enabled: true, clientId: "1234567890.apps.googleusercontent.com", clientSecret: "replace-with-a-secret", displayName: "Google", jitEnabled: true, allowedDomains: #["example.com"] })
@doc("Partial Google-provider update. Omitted properties retain their current values.")
model AdminGoogleProviderSettingsPatch extends AdminExternalProviderSettingsPatchBase {
	allowedDomains?: GoogleHostedDomain[];
}

@example(#{ enabled: true, clientId: "github-client-id", clientSecret: "replace-with-a-secret", displayName: "GitHub", jitEnabled: true, allowSignup: false })
@doc("Partial GitHub-provider update. Omitted properties retain their current values.")
model AdminGitHubProviderSettingsPatch extends AdminExternalProviderSettingsPatchBase {
	allowSignup?: boolean;
}

@example(#{
	settings: #{
		enabled: true,
		issuerUrl: "https://identity.example.com",
		clientId: "netstamp",
		clientSecretSet: true,
		displayName: "Company SSO",
		jitEnabled: true,
		callbackUrl: "https://netstamp.example.com/api/v1/auth/external/oidc/callback",
	},
})
model AdminOIDCProviderSettingsResponse {
	settings: AdminOIDCProviderSettings;
}

@example(#{
	settings: #{
		enabled: true,
		clientId: "1234567890.apps.googleusercontent.com",
		clientSecretSet: true,
		displayName: "Google",
		jitEnabled: true,
		callbackUrl: "https://netstamp.example.com/api/v1/auth/external/google/callback",
		allowedDomains: #["example.com"],
	},
})
model AdminGoogleProviderSettingsResponse {
	settings: AdminGoogleProviderSettings;
}

@example(#{
	settings: #{
		enabled: true,
		clientId: "github-client-id",
		clientSecretSet: true,
		displayName: "GitHub",
		jitEnabled: true,
		callbackUrl: "https://netstamp.example.com/api/v1/auth/external/github/callback",
		allowSignup: false,
	},
})
model AdminGitHubProviderSettingsResponse {
	settings: AdminGitHubProviderSettings;
}
