import { demoMode } from "../../config/features";
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
	if (demoMode) {
		throw new ApiError("Demo mode is read-only.", 403);
	}
}
