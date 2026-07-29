import { afterEach, describe, expect, it, vi } from "vitest";

const loadTrackerConfig = async () => {
	vi.resetModules();
	return (await import("./trackerConfig")).trackerConfig;
};

afterEach(() => {
	vi.unstubAllEnvs();
	vi.resetModules();
});

describe("trackerConfig", () => {
	it("maps build-time PostHog values to the tracker config casing", async () => {
		vi.stubEnv("VITE_NETSTAMP_POSTHOG_KEY", "phc_test");
		vi.stubEnv("VITE_NETSTAMP_POSTHOG_HOST", "https://posthog.example.com");

		const config = await loadTrackerConfig();

		expect(config.posthogKey).toBe("phc_test");
		expect(config.posthogHost).toBe("https://posthog.example.com");
	});

	it("prefers canonical Google and Meta variables over their aliases", async () => {
		vi.stubEnv("VITE_NETSTAMP_GOOGLE_TAG_ID", "G-CANONICAL");
		vi.stubEnv("VITE_NETSTAMP_GA_MEASUREMENT_ID", "G-ALIAS");
		vi.stubEnv("VITE_NETSTAMP_META_PIXEL_ID", "meta-canonical");
		vi.stubEnv("VITE_NETSTAMP_FACEBOOK_PIXEL_ID", "meta-alias");

		const config = await loadTrackerConfig();

		expect(config.googleTagId).toBe("G-CANONICAL");
		expect(config.metaPixelId).toBe("meta-canonical");
	});

	it("uses aliases when the canonical variables are blank or whitespace", async () => {
		vi.stubEnv("VITE_NETSTAMP_GOOGLE_TAG_ID", " ");
		vi.stubEnv("VITE_NETSTAMP_GA_MEASUREMENT_ID", "G-ALIAS");
		vi.stubEnv("VITE_NETSTAMP_META_PIXEL_ID", "");
		vi.stubEnv("VITE_NETSTAMP_FACEBOOK_PIXEL_ID", "meta-alias");

		const config = await loadTrackerConfig();

		expect(config.googleTagId).toBe("G-ALIAS");
		expect(config.metaPixelId).toBe("meta-alias");
	});

	it("normalizes immediate consent and configured countries", async () => {
		vi.stubEnv("VITE_NETSTAMP_TRACKING_CONSENT_MODE", "immediate");
		vi.stubEnv("VITE_NETSTAMP_TRACKING_CONSENT_COUNTRIES", "tw, JP tw");

		const config = await loadTrackerConfig();

		expect(config.consentMode).toBe("immediate");
		expect(config.consentCountries).toEqual(["TW", "JP", "TW"]);
	});

	it("leaves trackers disabled when public identifiers are blank", async () => {
		vi.stubEnv("VITE_NETSTAMP_GOOGLE_TAG_ID", "");
		vi.stubEnv("VITE_NETSTAMP_CLARITY_PROJECT_ID", "");
		vi.stubEnv("VITE_NETSTAMP_META_PIXEL_ID", "");
		vi.stubEnv("VITE_NETSTAMP_POSTHOG_KEY", "");
		vi.stubEnv("VITE_NETSTAMP_PLAUSIBLE_DOMAIN", "");
		vi.stubEnv("VITE_NETSTAMP_UMAMI_WEBSITE_ID", "");

		const config = await loadTrackerConfig();

		expect(config.googleTagId).toBeUndefined();
		expect(config.clarityProjectId).toBeUndefined();
		expect(config.metaPixelId).toBeUndefined();
		expect(config.posthogKey).toBeUndefined();
		expect(config.plausibleDomain).toBeUndefined();
		expect(config.umamiWebsiteId).toBeUndefined();
		expect(config.posthogHost).toBe("https://us.i.posthog.com");
		expect(config.plausibleScriptUrl).toBe("https://plausible.io/js/script.js");
		expect(config.umamiScriptUrl).toBe("https://cloud.umami.is/script.js");
		expect(config.consentMode).toBe("regional");
	});
});
