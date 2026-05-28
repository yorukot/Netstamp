export function optionalTrimmedText(value: string) {
	const trimmed = value.trim();
	return trimmed ? trimmed : undefined;
}

export function nullableTrimmedText(value: string) {
	const trimmed = value.trim();
	return trimmed ? trimmed : null;
}
