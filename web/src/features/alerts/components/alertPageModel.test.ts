// @vitest-environment jsdom

import { changeLocale, initializeI18n } from "@/i18n";
import type { ApiAlertIncident } from "@/shared/api/types";
import { beforeAll, describe, expect, it } from "vitest";
import { formatIncidentReason } from "./alertPageModel";

const stoppedIncident = {
	resolutionReason: "target_no_longer_evaluated",
	lastSummary: {
		state: "no_data",
		metric: "ping.loss_percent",
		operator: "gte",
		threshold: 10
	}
} as ApiAlertIncident;

beforeAll(async () => {
	await initializeI18n();
});

describe("formatIncidentReason", () => {
	it("explains when a resolved target is no longer evaluated", async () => {
		await changeLocale("en");
		expect(formatIncidentReason(stoppedIncident)).toBe("Target is no longer evaluated");

		await changeLocale("zh-TW");
		expect(formatIncidentReason(stoppedIncident)).toBe("此目標已不再進行評估");
	});
});
