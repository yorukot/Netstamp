/// <reference types="vite/client" />

interface ImportMetaEnv {
	readonly VITE_NETSTAMP_API_BASE_URL?: string;
	readonly VITE_NETSTAMP_API_PROXY_TARGET?: string;
	readonly VITE_NETSTAMP_DEMO_EMAIL?: string;
	readonly VITE_NETSTAMP_DEMO_PASSWORD?: string;
	readonly VITE_NETSTAMP_GA_MEASUREMENT_ID?: string;
	readonly VITE_NETSTAMP_GOOGLE_TAG_ID?: string;
	readonly VITE_NETSTAMP_CLARITY_PROJECT_ID?: string;
	readonly VITE_NETSTAMP_META_PIXEL_ID?: string;
	readonly VITE_NETSTAMP_FACEBOOK_PIXEL_ID?: string;
	readonly VITE_NETSTAMP_POSTHOG_KEY?: string;
	readonly VITE_NETSTAMP_POSTHOG_HOST?: string;
	readonly VITE_NETSTAMP_PLAUSIBLE_DOMAIN?: string;
	readonly VITE_NETSTAMP_PLAUSIBLE_SCRIPT_URL?: string;
	readonly VITE_NETSTAMP_UMAMI_WEBSITE_ID?: string;
	readonly VITE_NETSTAMP_UMAMI_SCRIPT_URL?: string;
	readonly VITE_NETSTAMP_TRACKING_CONSENT_MODE?: "regional" | "always" | "immediate";
	readonly VITE_NETSTAMP_TRACKING_CONSENT_COUNTRIES?: string;
}

interface ImportMeta {
	readonly env: ImportMetaEnv;
}
