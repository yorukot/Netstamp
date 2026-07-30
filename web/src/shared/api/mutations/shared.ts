import { getRuntimeFeaturesSnapshot } from "../../config/features";
import { ApiError } from "../client";

export interface AppMutationOptions {
	suppressGlobalErrorToast?: boolean;
}

const localErrorToastMeta = { suppressGlobalErrorToast: true } as const;

export function mutationToastOptions(options?: AppMutationOptions) {
	return options?.suppressGlobalErrorToast ? { meta: localErrorToastMeta } : {};
}

export function requireProjectRef(projectRef: string | null | undefined) {
	if (!projectRef) {
		throw new ApiError("No project selected.", 400);
	}

	return projectRef;
}

export function projectCacheRef(project: { id: string; slug?: string }) {
	return project.slug || project.id;
}

export function requireWritableAccess() {
	const runtimeFeatures = getRuntimeFeaturesSnapshot();

	if (runtimeFeatures.status !== "ready") {
		throw new ApiError("Runtime configuration is unavailable.", 503);
	}

	if (runtimeFeatures.readOnlyMode) {
		throw new ApiError("Demo mode is read-only.", 403);
	}
}
