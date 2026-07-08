import { normalizeTrackerConfig } from "@netstamp/ui/tracking";

export const trackerConfig = normalizeTrackerConfig({
	googleTagId: import.meta.env.VITE_NETSTAMP_GOOGLE_TAG_ID || import.meta.env.VITE_NETSTAMP_GA_MEASUREMENT_ID,
	clarityProjectId: import.meta.env.VITE_NETSTAMP_CLARITY_PROJECT_ID,
	metaPixelId: import.meta.env.VITE_NETSTAMP_META_PIXEL_ID || import.meta.env.VITE_NETSTAMP_FACEBOOK_PIXEL_ID,
	posthogKey: import.meta.env.VITE_NETSTAMP_POSTHOG_KEY,
	posthogHost: import.meta.env.VITE_NETSTAMP_POSTHOG_HOST,
	plausibleDomain: import.meta.env.VITE_NETSTAMP_PLAUSIBLE_DOMAIN,
	plausibleScriptUrl: import.meta.env.VITE_NETSTAMP_PLAUSIBLE_SCRIPT_URL,
	umamiWebsiteId: import.meta.env.VITE_NETSTAMP_UMAMI_WEBSITE_ID,
	umamiScriptUrl: import.meta.env.VITE_NETSTAMP_UMAMI_SCRIPT_URL,
	consentMode: import.meta.env.VITE_NETSTAMP_TRACKING_CONSENT_MODE,
	consentCountries: import.meta.env.VITE_NETSTAMP_TRACKING_CONSENT_COUNTRIES
});
