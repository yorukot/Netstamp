import type { components } from "@/shared/api/openapi";

type RuntimeConfig = components["schemas"]["PublicRuntimeConfig"];

export let demoMode = false;
export let readOnlyMode = false;

export const appFeatures = {
	registration: true,
	projectCreation: true,
	userCredentialChanges: true
};

export const applyRuntimeFeatures = (config: RuntimeConfig) => {
	demoMode = config.demoMode;
	readOnlyMode = config.demoMode;
	appFeatures.registration = !config.demoMode && config.capabilities.registrationEnabled;
	appFeatures.projectCreation = !config.demoMode && config.capabilities.projectCreationEnabled;
	appFeatures.userCredentialChanges = !config.demoMode && config.capabilities.credentialChangesEnabled;
};

const demoEmail = import.meta.env.VITE_NETSTAMP_DEMO_EMAIL?.trim();
const demoPassword = import.meta.env.VITE_NETSTAMP_DEMO_PASSWORD?.trim();

export const demoCredentials = demoEmail && demoPassword ? { email: demoEmail, password: demoPassword } : null;
