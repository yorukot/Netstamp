export const updateSettingsIntent = <TSettings extends object, TKey extends keyof TSettings>(
	current: Partial<TSettings> | null,
	serverSettings: TSettings,
	key: TKey,
	value: TSettings[TKey]
): Partial<TSettings> | null => {
	const next: Partial<TSettings> = { ...(current ?? {}), [key]: value };

	if (Object.is(value, serverSettings[key])) {
		delete next[key];
	}

	return Object.keys(next).length > 0 ? next : null;
};

export const applySettingsIntent = <TSettings extends object>(serverSettings: TSettings, intent: Partial<TSettings> | null): TSettings => ({
	...serverSettings,
	...intent
});

export const hasSettingsIntent = (intent: object | null) => Boolean(intent && Object.keys(intent).length > 0);

export const secretValueFromForm = (value: string, clear: boolean): string | null | undefined => {
	if (clear) {
		return null;
	}
	return value || undefined;
};

export const googleAllowedDomainsFromText = (value: string): string[] => {
	const rawDomains = value.trim();
	const allowedDomains = [
		...new Set(
			rawDomains
				.split(/[,\n]/)
				.map(domain => domain.trim().toLowerCase())
				.filter(Boolean)
		)
	];

	return rawDomains && allowedDomains.length === 0 ? [""] : allowedDomains;
};
