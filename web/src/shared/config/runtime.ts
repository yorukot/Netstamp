import { apiClient } from "@/shared/api/client";
import { applyRuntimeTracking } from "@/shared/tracking/trackerConfig";
import { applyRuntimeFeatures } from "./features";

export const loadRuntimeConfig = async () => {
	const response = await apiClient.GET("/system/config");
	if (response.data) {
		applyRuntimeFeatures(response.data);
		applyRuntimeTracking(response.data.tracking);
	}
};
