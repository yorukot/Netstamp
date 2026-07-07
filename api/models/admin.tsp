namespace Netstamp;

alias SMTPTLSMode = "starttls" | "implicit" | "none";

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
	backendBaseUrl: "https://app.netstamp.dev",
	publicWebBaseUrl: "https://app.netstamp.dev",
	smtp: #{ host: "smtp.example.com", port: 587, username: "netstamp", from: "alerts@example.com", tlsMode: "starttls", timeoutSeconds: 10 }
})
model UpdateAdminSettingsRequest {
	registrationEnabled?: boolean;
	backendBaseUrl?: string;
	publicWebBaseUrl?: string;
	smtp?: AdminSMTPSettingsPatch;
}

model AdminSettingsResponse {
	settings: AdminSettings;
}
