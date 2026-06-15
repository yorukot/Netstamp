/// <reference types="vite/client" />

interface ImportMetaEnv {
	readonly VITE_NETSTAMP_API_BASE_URL?: string;
	readonly VITE_NETSTAMP_API_PROXY_TARGET?: string;
	readonly VITE_NETSTAMP_REGISTRATION_ENABLED?: string;
	readonly VITE_NETSTAMP_PROJECT_CREATION_ENABLED?: string;
	readonly VITE_NETSTAMP_USER_CREDENTIAL_CHANGES_ENABLED?: string;
	readonly VITE_NETSTAMP_DEMO_MODE?: string;
	readonly VITE_NETSTAMP_DEMO_EMAIL?: string;
	readonly VITE_NETSTAMP_DEMO_PASSWORD?: string;
	readonly VITE_NETSTAMP_GA_MEASUREMENT_ID?: string;
	readonly VITE_NETSTAMP_GOOGLE_TAG_ID?: string;
	readonly VITE_NETSTAMP_CLARITY_PROJECT_ID?: string;
	readonly VITE_NETSTAMP_POSTHOG_KEY?: string;
	readonly VITE_NETSTAMP_POSTHOG_HOST?: string;
	readonly VITE_NETSTAMP_PLAUSIBLE_DOMAIN?: string;
	readonly VITE_NETSTAMP_PLAUSIBLE_SCRIPT_URL?: string;
	readonly VITE_NETSTAMP_UMAMI_WEBSITE_ID?: string;
	readonly VITE_NETSTAMP_UMAMI_SCRIPT_URL?: string;
	readonly VITE_NETSTAMP_TRACKING_CONSENT_MODE?: "regional" | "always" | "never";
	readonly VITE_NETSTAMP_TRACKING_CONSENT_COUNTRIES?: string;
}

interface ImportMeta {
	readonly env: ImportMetaEnv;
}
