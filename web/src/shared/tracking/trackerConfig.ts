import type { components } from "@/shared/api/openapi";
import { normalizeTrackerConfig } from "@netstamp/ui/tracking";

type TrackingSettings = components["schemas"]["TrackingSettings"];

export let trackerConfig = normalizeTrackerConfig({ consentMode: "regional" });

export const applyRuntimeTracking = (tracking: TrackingSettings) => {
	trackerConfig = normalizeTrackerConfig({
		...tracking,
		posthogKey: tracking.postHogKey,
		posthogHost: tracking.postHogHost
	});
};
