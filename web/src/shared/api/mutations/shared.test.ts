import { ApiError } from "@/shared/api/client";
import { applyRuntimeFeatures, markRuntimeFeaturesFailed, markRuntimeFeaturesLoading, resetRuntimeFeatures } from "@/shared/config/features";
import { beforeEach, describe, expect, it } from "vitest";
import { requireWritableAccess } from "./shared";

const runtimeConfig = (demoMode: boolean) => ({
	demoMode,
	capabilities: {
		accountCreationEnabled: true,
		projectCreationEnabled: true,
		credentialChangesEnabled: true
	}
});

const writableAccessError = (): ApiError => {
	try {
		requireWritableAccess();
	} catch (error) {
		if (error instanceof ApiError) {
			return error;
		}

		throw error;
	}

	throw new Error("Expected writable access to be rejected");
};

beforeEach(() => {
	resetRuntimeFeatures();
});

describe("requireWritableAccess", () => {
	it("rejects before runtime configuration has loaded", () => {
		expect(writableAccessError().status).toBe(503);
	});

	it("rejects while runtime configuration is loading or failed", () => {
		markRuntimeFeaturesLoading();
		expect(writableAccessError().status).toBe(503);

		markRuntimeFeaturesFailed();
		expect(writableAccessError().status).toBe(503);
	});

	it("rejects a successfully loaded demo configuration", () => {
		applyRuntimeFeatures(runtimeConfig(true) as never);
		expect(writableAccessError().status).toBe(403);
	});

	it("allows mutations only after a writable configuration succeeds", () => {
		applyRuntimeFeatures(runtimeConfig(false) as never);
		expect(() => requireWritableAccess()).not.toThrow();
	});
});
