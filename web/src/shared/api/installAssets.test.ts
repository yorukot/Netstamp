// @vitest-environment jsdom

import { describe, expect, it } from "vitest";
import { controllerInstallTarget, probeServiceInstallCommand, probeUpgradeCommand } from "./installAssets";

describe("probe agent commands", () => {
	it("uses the browser origin as the controller install target", () => {
		expect(controllerInstallTarget()).toBe(window.location.origin);
	});

	it("installs a probe against the same-origin controller", () => {
		const command = probeServiceInstallCommand({ probeId: "probe-1", probeSecret: "secret-1" });

		expect(command).toContain(`--url '${window.location.origin}'`);
		expect(command).toContain("--probe-id 'probe-1'");
		expect(command).toContain("--probe-secret 'secret-1'");
	});

	it("updates without the removed API version flag", () => {
		const command = probeUpgradeCommand();

		expect(command).toBe(`sudo netstamp-agent update \\\n  --url '${window.location.origin}'`);
		expect(command).not.toContain("--api-version");
	});
});
