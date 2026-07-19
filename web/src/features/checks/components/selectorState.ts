import { i18n } from "@/i18n";
import type { ApiLabel, ApiSelector } from "@/shared/api/types";

const checkT = i18n.getFixedT(null, "checks") as (key: string) => string;

export type SelectorMode = "all-probes" | "all" | "any" | "advanced";
export type SelectorLabelOp = NonNullable<ApiSelector["label"]>["op"];

export interface SelectorState {
	mode: SelectorMode;
	rules: SelectorRule[];
	advancedText: string;
}

export interface SelectorLabelOption {
	id: string;
	key: string;
	value: string;
}

export interface SelectorRule {
	id: string;
	key: string;
	op: SelectorLabelOp;
	value: string;
	values: string;
	negated: boolean;
}

function selectorLabelId(key: string, value: string) {
	return `${key}\u0000${value}`;
}

export function selectorLabelOptions(labels: ApiLabel[]): SelectorLabelOption[] {
	const options = new Map<string, SelectorLabelOption>();

	for (const label of labels) {
		const id = selectorLabelId(label.key, label.value);
		options.set(id, { id, key: label.key, value: label.value });
	}

	return Array.from(options.values()).sort((a, b) => a.key.localeCompare(b.key) || a.value.localeCompare(b.value));
}

export function selectorValuesForKey(options: SelectorLabelOption[], key: string) {
	return options.filter(option => option.key === key).map(option => option.value);
}

export function createSelectorRule(options: SelectorLabelOption[], seed?: Partial<SelectorRule>): SelectorRule {
	const firstOption = options[0];
	return {
		id: globalThis.crypto.randomUUID(),
		key: seed?.key ?? firstOption?.key ?? "",
		op: seed?.op ?? "eq",
		value: seed?.value ?? firstOption?.value ?? "",
		values: seed?.values ?? firstOption?.value ?? "",
		negated: seed?.negated ?? false
	};
}

function splitSelectorValues(value: string) {
	return Array.from(
		new Set(
			value
				.split(",")
				.map(item => item.trim())
				.filter(Boolean)
		)
	);
}

function selectorRuleLabel(rule: SelectorRule): NonNullable<ApiSelector["label"]> | null {
	const key = rule.key.trim();
	if (!key) {
		return null;
	}

	if (rule.op === "exists") {
		return { key, op: "exists" };
	}

	if (rule.op === "in") {
		const values = splitSelectorValues(rule.values);
		return values.length ? { key, op: "in", values } : null;
	}

	const value = rule.value.trim();
	return value ? { key, op: "eq", value } : null;
}

function selectorRuleNode(rule: SelectorRule): ApiSelector | null {
	const label = selectorRuleLabel(rule);
	if (!label) {
		return null;
	}

	const node: ApiSelector = { label };
	return rule.negated ? { not: node } : node;
}

function parseAdvancedSelector(value: string): ApiSelector {
	const trimmed = value.trim();
	if (!trimmed) {
		return {};
	}

	const parsed: unknown = JSON.parse(trimmed);
	if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
		throw new Error(checkT("selector.object"));
	}

	return parsed as ApiSelector;
}

export function buildSelector(state: SelectorState): ApiSelector {
	if (state.mode === "advanced") {
		return parseAdvancedSelector(state.advancedText);
	}

	if (state.mode === "all-probes") {
		return {};
	}

	const children = state.rules.map(selectorRuleNode).filter((selector): selector is ApiSelector => Boolean(selector));
	if (!children.length) {
		return {};
	}

	if (children.length === 1) {
		return children[0];
	}

	return state.mode === "any" ? { any: children } : { all: children };
}

function isEmptySelector(selector: ApiSelector | null | undefined) {
	return !selector || Object.keys(selector).length === 0;
}

function selectorRuleFromNode(selector: ApiSelector | null | undefined): SelectorRule | null {
	const negated = Boolean(selector?.not);
	const node = negated ? selector?.not : selector;
	const label = node?.label;

	if (!label) {
		return null;
	}

	if (label.op === "exists") {
		return createSelectorRule([], { key: label.key, op: "exists", negated });
	}

	if (label.op === "in" && label.values?.length) {
		return createSelectorRule([], { key: label.key, op: "in", values: label.values.join(", "), negated });
	}

	if (label.op === "eq" && label.value) {
		return createSelectorRule([], { key: label.key, op: "eq", value: label.value, values: label.value, negated });
	}

	return null;
}

export function selectorStateFromApi(selector: ApiSelector | null | undefined): SelectorState {
	if (isEmptySelector(selector)) {
		return { mode: "all-probes", rules: [], advancedText: "" };
	}

	const directRule = selectorRuleFromNode(selector);
	if (directRule) {
		return { mode: "all", rules: [directRule], advancedText: "" };
	}

	for (const mode of ["all", "any"] as const) {
		const children = selector?.[mode];
		if (!children?.length) {
			continue;
		}

		const rules = children.map(selectorRuleFromNode);
		if (rules.every((rule): rule is SelectorRule => Boolean(rule))) {
			return { mode, rules, advancedText: "" };
		}
	}

	return { mode: "advanced", rules: [], advancedText: JSON.stringify(selector, null, 2) };
}

export function probeMatchesSelector(probeLabelTokens: string[], state: SelectorState) {
	if (state.mode === "all-probes") {
		return true;
	}

	if (state.mode === "advanced") {
		return false;
	}

	const matches = state.rules.map(rule => {
		const key = rule.key.trim();
		const matched =
			rule.op === "exists"
				? probeLabelTokens.some(labelToken => labelToken.startsWith(`${key}:`))
				: rule.op === "in"
					? splitSelectorValues(rule.values).some(value => probeLabelTokens.includes(`${key}:${value}`))
					: probeLabelTokens.includes(`${key}:${rule.value.trim()}`);

		return rule.negated ? !matched : matched;
	});

	return state.mode === "any" ? matches.some(Boolean) : matches.every(Boolean);
}
