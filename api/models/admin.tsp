namespace Netstamp;

alias SMTPTLSMode = "starttls" | "implicit" | "none";

@example(#{
	id: "11111111-1111-1111-1111-111111111111",
	email: "admin@example.com",
	displayName: "Jane Operator",
	emailVerified: true,
	grantedAt: utcDateTime.fromISO("2026-07-08T12:00:00Z")
})
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

@example(#{ admin: #{ id: "11111111-1111-1111-1111-111111111111", email: "admin@example.com", displayName: "Jane Operator", emailVerified: true, grantedAt: utcDateTime.fromISO("2026-07-08T12:00:00Z") } })
model SystemAdminResponse {
	admin: SystemAdminUser;
}

@example(#{ admins: #[#{ id: "11111111-1111-1111-1111-111111111111", email: "admin@example.com", displayName: "Jane Operator", emailVerified: true, grantedAt: utcDateTime.fromISO("2026-07-08T12:00:00Z") }] })
model SystemAdminsResponse {
	admins: SystemAdminUser[];
}

@example(#{
	host: "smtp.example.com",
	port: 587,
	username: "netstamp",
	passwordSet: true,
	from: "alerts@example.com",
	tlsMode: "starttls",
	timeoutSeconds: 10,
	configured: true
})
model AdminSMTPSettings {
	@doc("SMTP server hostname. Empty means email delivery is disabled unless an environment fallback is still active.")
	host: string;

	@minValue(1)
	@maxValue(65535)
	port: int32;

	@doc("SMTP username. Empty means no SMTP AUTH.")
	username: string;

	@doc("Whether a password is configured. The password value is never returned.")
	passwordSet: boolean;

	@doc("Envelope sender email address.")
	from: string;

	tlsMode: SMTPTLSMode;

	@minValue(1)
	timeoutSeconds: int32;

	@doc("Whether Netstamp has enough SMTP settings to attempt email delivery.")
	configured: boolean;
}

@example(#{
	registrationEnabled: true,
	emailVerificationRequired: false,
	backendBaseUrl: "https://app.netstamp.dev",
	publicWebBaseUrl: "https://app.netstamp.dev",
	smtp: #{
		host: "smtp.example.com",
		port: 587,
		username: "netstamp",
		passwordSet: true,
		from: "alerts@example.com",
		tlsMode: "starttls",
		timeoutSeconds: 10,
		configured: true
	}
})
model AdminSettings {
	registrationEnabled: boolean;
	emailVerificationRequired: boolean;
	backendBaseUrl: string;
	publicWebBaseUrl: string;
	smtp: AdminSMTPSettings;
}

model AdminSMTPSettingsPatch {
	host?: string;

	@minValue(1)
	@maxValue(65535)
	port?: int32;

	username?: string;

	@secret
	@doc("New SMTP password. Omit to keep the stored password unchanged.")
	password?: string;

	@doc("Clear the stored SMTP password and override any environment fallback with no password.")
	clearPassword?: boolean;

	from?: string;
	tlsMode?: SMTPTLSMode;

	@minValue(1)
	timeoutSeconds?: int32;
}

@example(#{
	registrationEnabled: false,
	emailVerificationRequired: true,
	backendBaseUrl: "https://app.netstamp.dev",
	publicWebBaseUrl: "https://app.netstamp.dev",
	smtp: #{ host: "smtp.example.com", port: 587, username: "netstamp", from: "alerts@example.com", tlsMode: "starttls", timeoutSeconds: 10 }
})
model UpdateAdminSettingsRequest {
	registrationEnabled?: boolean;
	emailVerificationRequired?: boolean;
	backendBaseUrl?: string;
	publicWebBaseUrl?: string;
	smtp?: AdminSMTPSettingsPatch;
}

model AdminSettingsResponse {
	settings: AdminSettings;
}
