import type { CheckDefinition } from "@/features/checks/data/checks";

function normalizedTargetKey(target: string) {
	return target.trim().toLocaleLowerCase();
}

export function groupChecksByTarget(checks: CheckDefinition[]): CheckDefinition[] {
	const targetOrder = new Map<string, number>();

	return checks
		.map((check, index) => {
			const targetKey = normalizedTargetKey(check.target);
			const targetIndex = targetOrder.get(targetKey) ?? targetOrder.size;
			targetOrder.set(targetKey, targetIndex);

			return { check, index, targetIndex };
		})
		.sort((a, b) => a.targetIndex - b.targetIndex || a.index - b.index)
		.map(item => item.check);
}

export function displayProbeSelection(selectedProbes: string[]) {
	if (!selectedProbes.length) {
		return "No probes assigned";
	}

	if (selectedProbes.length === 1) {
		return selectedProbes[0];
	}

	return `${selectedProbes.length} probes selected`;
}
