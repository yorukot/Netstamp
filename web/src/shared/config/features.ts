import type { components } from "@/shared/api/openapi";
import { useSyncExternalStore } from "react";

type RuntimeConfig = components["schemas"]["PublicRuntimeConfig"];

export type RuntimeFeatureStatus = "unknown" | "loading" | "ready" | "failed";

interface AppFeatures {
	readonly registration: boolean;
	readonly projectCreation: boolean;
	readonly userCredentialChanges: boolean;
}

export interface RuntimeFeatureSnapshot {
	readonly status: RuntimeFeatureStatus;
	readonly demoMode: boolean;
	readonly readOnlyMode: boolean;
	readonly appFeatures: AppFeatures;
}

const disabledAppFeatures = Object.freeze<AppFeatures>({
	registration: false,
	projectCreation: false,
	userCredentialChanges: false
});

const failClosedSnapshot = (status: Exclude<RuntimeFeatureStatus, "ready">): RuntimeFeatureSnapshot =>
	Object.freeze({
		status,
		demoMode: false,
		readOnlyMode: true,
		appFeatures: disabledAppFeatures
	});

const initialSnapshot = failClosedSnapshot("unknown");
let runtimeFeatureSnapshot = initialSnapshot;
const runtimeFeatureListeners = new Set<() => void>();

const runtimeFeaturesEqual = (left: RuntimeFeatureSnapshot, right: RuntimeFeatureSnapshot) =>
	left.status === right.status &&
	left.demoMode === right.demoMode &&
	left.readOnlyMode === right.readOnlyMode &&
	left.appFeatures.registration === right.appFeatures.registration &&
	left.appFeatures.projectCreation === right.appFeatures.projectCreation &&
	left.appFeatures.userCredentialChanges === right.appFeatures.userCredentialChanges;

const publishRuntimeFeatures = (snapshot: RuntimeFeatureSnapshot) => {
	if (runtimeFeaturesEqual(runtimeFeatureSnapshot, snapshot)) {
		return;
	}

	runtimeFeatureSnapshot = snapshot;
	[...runtimeFeatureListeners].forEach(listener => listener());
};

export const getRuntimeFeaturesSnapshot = () => runtimeFeatureSnapshot;

export const subscribeRuntimeFeatures = (listener: () => void) => {
	runtimeFeatureListeners.add(listener);
	return () => runtimeFeatureListeners.delete(listener);
};

export const useRuntimeFeatures = () => useSyncExternalStore(subscribeRuntimeFeatures, getRuntimeFeaturesSnapshot, getRuntimeFeaturesSnapshot);

export const applyRuntimeFeatures = (config: RuntimeConfig) => {
	const demoMode = config.demoMode === true;
	const writable = config.demoMode === false;

	publishRuntimeFeatures(
		Object.freeze({
			status: "ready",
			demoMode,
			readOnlyMode: !writable,
			appFeatures: Object.freeze({
				registration: writable && config.capabilities.accountCreationEnabled === true,
				projectCreation: writable && config.capabilities.projectCreationEnabled === true,
				userCredentialChanges: writable && config.capabilities.credentialChangesEnabled === true
			})
		})
	);
};

export const resetRuntimeFeatures = () => {
	publishRuntimeFeatures(initialSnapshot);
};

export const markRuntimeFeaturesLoading = () => {
	publishRuntimeFeatures(failClosedSnapshot("loading"));
};

export const markRuntimeFeaturesFailed = () => {
	publishRuntimeFeatures(failClosedSnapshot("failed"));
};

const demoEmail = import.meta.env.VITE_NETSTAMP_DEMO_EMAIL?.trim();
const demoPassword = import.meta.env.VITE_NETSTAMP_DEMO_PASSWORD?.trim();

export const demoCredentials = demoEmail && demoPassword ? { email: demoEmail, password: demoPassword } : null;
