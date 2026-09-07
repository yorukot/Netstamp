import { normalizeTrackerConfig } from "@netstamp/ui/tracking";

export const trackerConfig = normalizeTrackerConfig({
	umamiWebsiteId: import.meta.env.VITE_NETSTAMP_UMAMI_WEBSITE_ID,
	umamiScriptUrl: import.meta.env.VITE_NETSTAMP_UMAMI_SCRIPT_URL,
	consentMode: import.meta.env.VITE_NETSTAMP_TRACKING_CONSENT_MODE,
	consentCountries: import.meta.env.VITE_NETSTAMP_TRACKING_CONSENT_COUNTRIES
});
