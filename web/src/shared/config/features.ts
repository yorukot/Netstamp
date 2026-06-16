function booleanFeature(value: string | undefined, defaultValue: boolean) {
	const normalized = value?.trim().toLowerCase();

	if (!normalized) {
		return defaultValue;
	}

	if (["1", "true", "yes", "on"].includes(normalized)) {
		return true;
	}

	if (["0", "false", "no", "off"].includes(normalized)) {
		return false;
	}

	return defaultValue;
}

export const demoMode = booleanFeature(import.meta.env.VITE_NETSTAMP_DEMO_MODE, false);
export const readOnlyMode = demoMode;

export const appFeatures = {
	registration: !readOnlyMode && booleanFeature(import.meta.env.VITE_NETSTAMP_REGISTRATION_ENABLED, true),
	projectCreation: !readOnlyMode && booleanFeature(import.meta.env.VITE_NETSTAMP_PROJECT_CREATION_ENABLED, true),
	userCredentialChanges: !readOnlyMode && booleanFeature(import.meta.env.VITE_NETSTAMP_USER_CREDENTIAL_CHANGES_ENABLED, true)
} as const;

const demoEmail = import.meta.env.VITE_NETSTAMP_DEMO_EMAIL?.trim();
const demoPassword = import.meta.env.VITE_NETSTAMP_DEMO_PASSWORD?.trim();

export const demoCredentials =
	demoEmail && demoPassword
		? {
				email: demoEmail,
				password: demoPassword
			}
		: null;
