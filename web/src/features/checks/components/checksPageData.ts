export function displayProbeSelection(selectedProbes: string[]) {
	if (!selectedProbes.length) {
		return "No probes assigned";
	}

	if (selectedProbes.length === 1) {
		return selectedProbes[0];
	}

	return `${selectedProbes.length} probes selected`;
}
