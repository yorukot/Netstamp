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
	it("maps canonical Umami variables", async () => {
		vi.stubEnv("VITE_NETSTAMP_UMAMI_WEBSITE_ID", "umami-site");

		const config = await loadTrackerConfig();

		expect(config.umamiWebsiteId).toBe("umami-site");
	});

	it("normalizes immediate consent and configured countries", async () => {
		vi.stubEnv("VITE_NETSTAMP_TRACKING_CONSENT_MODE", "immediate");
		vi.stubEnv("VITE_NETSTAMP_TRACKING_CONSENT_COUNTRIES", "tw, JP tw");

		const config = await loadTrackerConfig();

		expect(config.consentMode).toBe("immediate");
		expect(config.consentCountries).toEqual(["TW", "JP", "TW"]);
	});

	it("leaves trackers disabled when public identifiers are blank", async () => {
		vi.stubEnv("VITE_NETSTAMP_UMAMI_WEBSITE_ID", "");

		const config = await loadTrackerConfig();

		expect(config.umamiWebsiteId).toBeUndefined();
		expect(config.umamiScriptUrl).toBe("https://cloud.umami.is/script.js");
		expect(config.consentMode).toBe("regional");
	});
});
