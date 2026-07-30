// @vitest-environment jsdom

import { apiClient } from "@/shared/api/client";
import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { applyRuntimeFeatures, getRuntimeFeaturesSnapshot, markRuntimeFeaturesFailed, resetRuntimeFeatures, subscribeRuntimeFeatures, useRuntimeFeatures } from "./features";
import { loadRuntimeConfig } from "./runtime";

vi.mock("@/shared/api/client", () => ({
	apiClient: {
		GET: vi.fn()
	}
}));

const getRuntimeConfig = vi.mocked(apiClient.GET);

const enabledRuntimeConfig = {
	demoMode: false,
	capabilities: {
		accountCreationEnabled: true,
		projectCreationEnabled: true,
		credentialChangesEnabled: true
	}
};

beforeEach(() => {
	resetRuntimeFeatures();
	getRuntimeConfig.mockReset();
});

afterEach(() => {
	vi.useRealTimers();
});

describe("loadRuntimeConfig", () => {
	it("applies runtime capabilities returned by the API", async () => {
		const { result } = renderHook(() => useRuntimeFeatures());
		getRuntimeConfig.mockResolvedValue({
			data: enabledRuntimeConfig,
			response: new Response(null, { status: 200 })
		} as never);

		await act(() => loadRuntimeConfig());

		expect(result.current).toMatchObject({
			status: "ready",
			demoMode: false,
			readOnlyMode: false
		});
		expect(result.current.appFeatures).toEqual({
			registration: true,
			projectCreation: true,
			userCredentialChanges: true
		});
	});

	it("keeps gated actions disabled in demo mode", async () => {
		getRuntimeConfig.mockResolvedValue({
			data: {
				...enabledRuntimeConfig,
				demoMode: true
			},
			response: new Response(null, { status: 200 })
		} as never);

		await loadRuntimeConfig();

		const snapshot = getRuntimeFeaturesSnapshot();
		expect(snapshot.demoMode).toBe(true);
		expect(snapshot.readOnlyMode).toBe(true);
		expect(snapshot.appFeatures).toEqual({
			registration: false,
			projectCreation: false,
			userCredentialChanges: false
		});
	});

	it("resets capabilities before a failed refresh", async () => {
		applyRuntimeFeatures(enabledRuntimeConfig as never);
		getRuntimeConfig.mockResolvedValue({
			error: { title: "Unavailable" },
			response: new Response(null, { status: 503 })
		} as never);

		await expect(loadRuntimeConfig()).rejects.toThrow("Runtime configuration is unavailable");

		const snapshot = getRuntimeFeaturesSnapshot();
		expect(snapshot.status).toBe("failed");
		expect(snapshot.readOnlyMode).toBe(true);
		expect(snapshot.appFeatures).toEqual({
			registration: false,
			projectCreation: false,
			userCredentialChanges: false
		});
	});

	it("aborts a runtime request after the configured timeout", async () => {
		vi.useFakeTimers();
		getRuntimeConfig.mockImplementation(((...args: unknown[]) => {
			const options = args[1] as { signal?: AbortSignal } | undefined;
			return new Promise((_resolve, reject) => {
				options?.signal?.addEventListener("abort", () => reject(new DOMException("The operation was aborted", "AbortError")));
			}) as never;
		}) as never);

		const result = expect(loadRuntimeConfig(50)).rejects.toMatchObject({ name: "AbortError" });
		await vi.advanceTimersByTimeAsync(50);
		await result;

		expect(getRuntimeFeaturesSnapshot()).toMatchObject({
			status: "failed",
			readOnlyMode: true,
			appFeatures: { registration: false }
		});
	});

	it("does not let an older request overwrite a newer response", async () => {
		let resolveFirst: (value: unknown) => void = () => {};
		let resolveSecond: (value: unknown) => void = () => {};
		getRuntimeConfig
			.mockImplementationOnce(
				() =>
					new Promise(resolve => {
						resolveFirst = resolve;
					}) as never
			)
			.mockImplementationOnce(
				() =>
					new Promise(resolve => {
						resolveSecond = resolve;
					}) as never
			);

		const firstRequest = loadRuntimeConfig();
		const secondRequest = loadRuntimeConfig();

		resolveSecond({
			data: { ...enabledRuntimeConfig, demoMode: true },
			response: new Response(null, { status: 200 })
		});
		await secondRequest;
		expect(getRuntimeFeaturesSnapshot().demoMode).toBe(true);

		resolveFirst({
			data: enabledRuntimeConfig,
			response: new Response(null, { status: 200 })
		});
		await firstRequest;
		expect(getRuntimeFeaturesSnapshot().demoMode).toBe(true);
	});
});

describe("runtime feature store", () => {
	it("starts fail-closed before configuration succeeds", () => {
		expect(getRuntimeFeaturesSnapshot()).toMatchObject({
			status: "unknown",
			demoMode: false,
			readOnlyMode: true,
			appFeatures: {
				registration: false,
				projectCreation: false,
				userCredentialChanges: false
			}
		});
	});

	it("publishes immutable snapshots to React subscribers", () => {
		const { result } = renderHook(() => useRuntimeFeatures());
		const initial = result.current;

		act(() => applyRuntimeFeatures(enabledRuntimeConfig as never));

		expect(result.current).not.toBe(initial);
		expect(result.current.status).toBe("ready");
		expect(Object.isFrozen(result.current)).toBe(true);
		expect(Object.isFrozen(result.current.appFeatures)).toBe(true);
	});

	it("does not notify listeners for an equivalent snapshot and supports unsubscribe", () => {
		const listener = vi.fn();
		const unsubscribe = subscribeRuntimeFeatures(listener);

		applyRuntimeFeatures(enabledRuntimeConfig as never);
		applyRuntimeFeatures(enabledRuntimeConfig as never);
		expect(listener).toHaveBeenCalledTimes(1);

		unsubscribe();
		markRuntimeFeaturesFailed();
		expect(listener).toHaveBeenCalledTimes(1);
	});
});
