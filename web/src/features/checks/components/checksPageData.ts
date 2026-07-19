import type { CheckDefinition } from "@/features/checks/data/checks";
import { i18n } from "@/i18n";

const checkT = i18n.getFixedT(null, "checks") as (key: string, options?: Record<string, unknown>) => string;

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
		return checkT("selector.noneAssigned");
	}

	if (selectedProbes.length === 1) {
		return selectedProbes[0];
	}

	return checkT("selector.selected", { count: selectedProbes.length });
}
