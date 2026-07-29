import { apiClient } from "@/shared/api/client";
import { applyRuntimeFeatures, markRuntimeFeaturesFailed, markRuntimeFeaturesLoading } from "./features";

export const DEFAULT_RUNTIME_CONFIG_TIMEOUT_MS = 3_000;

let currentRuntimeConfigRequest = 0;

export const loadRuntimeConfig = async (timeoutMs = DEFAULT_RUNTIME_CONFIG_TIMEOUT_MS) => {
	const requestId = ++currentRuntimeConfigRequest;
	markRuntimeFeaturesLoading();

	const controller = new AbortController();
	const timeout = window.setTimeout(() => controller.abort(), timeoutMs);

	try {
		const response = await apiClient.GET("/system/config", {
			signal: controller.signal
		});

		if (!response.response.ok || !response.data) {
			throw new Error("Runtime configuration is unavailable");
		}

		if (requestId === currentRuntimeConfigRequest) {
			applyRuntimeFeatures(response.data);
		}
	} catch (error) {
		if (requestId === currentRuntimeConfigRequest) {
			markRuntimeFeaturesFailed();
		}
		throw error;
	} finally {
		window.clearTimeout(timeout);
	}
};
