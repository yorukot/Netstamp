import { describe, expect, it } from "vitest";
import { applyRuntimeTracking, trackerConfig } from "./trackerConfig";

describe("applyRuntimeTracking", () => {
	it("maps the API PostHog settings to the tracker config", () => {
		applyRuntimeTracking({
			googleTagId: "",
			clarityProjectId: "",
			metaPixelId: "",
			postHogKey: "phc_test",
			postHogHost: "https://posthog.example.com",
			plausibleDomain: "",
			plausibleScriptUrl: "",
			umamiWebsiteId: "",
			umamiScriptUrl: "",
			consentMode: "regional",
			consentCountries: ""
		});

		expect(trackerConfig.posthogKey).toBe("phc_test");
		expect(trackerConfig.posthogHost).toBe("https://posthog.example.com");
	});
});
