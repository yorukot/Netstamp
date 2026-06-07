import { mapApiCheck, mapApiChecksWithAssignments, parseIntervalSeconds } from "@/features/checks/api/checkAdapters";
import {
	buildPingConfigPayload,
	buildTCPConfigPayload,
	buildTracerouteConfigPayload,
	defaultPingConfigFormState,
	defaultTCPConfigFormState,
	defaultTracerouteConfigFormState,
	pingConfigFormStateFromApi,
	tcpConfigFormStateFromApi,
	tracerouteConfigFormStateFromApi,
	type PingConfigFormState,
	type TCPConfigFormState,
	type TracerouteConfigFormState
} from "@/features/checks/data/checkConfig";
import { type CheckDefinition, type CheckType } from "@/features/checks/data/checks";
import { mapApiProbes } from "@/features/probes/api/probeAdapters";
import { pathForCheckDetail, pathForRoute } from "@/routes/routePaths";
import {
	BatchCheckDeleteError,
	useCreateProjectCheckMutation,
	useDeleteProjectCheckMutation,
	useDeleteProjectChecksMutation,
	usePreviewProjectSelectorMutation,
	useUpdateProjectCheckMutation
} from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiCheck, ApiLabel, ApiSelector, CreateCheckInput } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { ActionRow } from "@/shared/components/ActionRow";
import { useConfirm } from "@/shared/components/confirmContext";
import { EditorDrawer } from "@/shared/components/EditorDrawer";
import { FilterGrid } from "@/shared/components/FilterGrid";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { pushErrorToast, pushToast } from "@/shared/toast/toastStore";
import { classNames } from "@/shared/utils/classNames";
import { Badge, Button, Checkbox, FieldLabel, Panel, SelectField, TextAreaField, TextField } from "@netstamp/ui";
import { Trash } from "@phosphor-icons/react";
import { useQuery } from "@tanstack/react-query";
import type { RowSelectionState } from "@tanstack/react-table";
import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { CheckConfigFields } from "./CheckConfigFields";
import styles from "./ChecksPage.module.css";
import { displayProbeSelection, groupChecksByTarget } from "./checksPageData";
import { ChecksTable, type CheckTypeFilter } from "./ChecksTable";

type SelectorMode = "all-probes" | "all" | "any" | "advanced";
type SelectorLabelOp = NonNullable<ApiSelector["label"]>["op"];

interface SelectorState {
	mode: SelectorMode;
	rules: SelectorRule[];
	advancedText: string;
}

interface SelectorLabelOption {
	id: string;
	key: string;
	value: string;
}

interface SelectorRule {
	id: string;
	key: string;
	op: SelectorLabelOp;
	value: string;
	values: string;
	negated: boolean;
}

const selectorModeOptions: Array<{ value: SelectorMode; label: string }> = [
	{ value: "all-probes", label: "All probes" },
	{ value: "all", label: "Match all labels" },
	{ value: "any", label: "Match any label" },
	{ value: "advanced", label: "Advanced JSON" }
];

const selectorOpOptions: Array<{ value: SelectorLabelOp; label: string }> = [
	{ value: "eq", label: "equals" },
	{ value: "in", label: "in values" },
	{ value: "exists", label: "exists" }
];

function checkTypeFromApi(type: string): CheckType {
	switch (type) {
		case "tcp":
			return "TCP";
		case "traceroute":
			return "Traceroute";
		default:
			return "Ping";
	}
}

function copiedCheckName(name: string) {
	const base = name.trim() || "Check";
	const suffix = " copy";
	const maxBaseLength = Math.max(1, 128 - suffix.length);

	return `${base.slice(0, maxBaseLength)}${suffix}`;
}

function duplicateCheckInput(check: ApiCheck): CreateCheckInput {
	const body: CreateCheckInput = {
		intervalSeconds: check.intervalSeconds,
		name: copiedCheckName(check.name),
		target: check.target,
		type: check.type
	};

	if (check.selector) {
		body.selector = check.selector;
	}
	if (check.description) {
		body.description = check.description;
	}
	if (check.labels.length) {
		body.labelIds = check.labels.map(label => label.id);
	}
	if (check.pingConfig) {
		body.pingConfig = { ...check.pingConfig };
	}
	if (check.tcpConfig) {
		body.tcpConfig = { ...check.tcpConfig };
	}
	if (check.tracerouteConfig) {
		body.tracerouteConfig = { ...check.tracerouteConfig };
	}

	return body;
}

function selectorLabelId(key: string, value: string) {
	return `${key}\u0000${value}`;
}

function selectorLabelOptions(labels: ApiLabel[]): SelectorLabelOption[] {
	const options = new Map<string, SelectorLabelOption>();

	for (const label of labels) {
		const id = selectorLabelId(label.key, label.value);
		options.set(id, { id, key: label.key, value: label.value });
	}

	return Array.from(options.values()).sort((a, b) => a.key.localeCompare(b.key) || a.value.localeCompare(b.value));
}

function selectorValuesForKey(options: SelectorLabelOption[], key: string) {
	return options.filter(option => option.key === key).map(option => option.value);
}

function createSelectorRule(options: SelectorLabelOption[], seed?: Partial<SelectorRule>): SelectorRule {
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
		throw new Error("Selector JSON must be an object.");
	}

	return parsed as ApiSelector;
}

function buildSelector(state: SelectorState): ApiSelector {
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

function selectorStateFromApi(selector: ApiSelector | null | undefined): SelectorState {
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

function probeMatchesSelector(probeLabelTokens: string[], state: SelectorState) {
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

export function ChecksPage() {
	const { projectRef } = useCurrentProject();
	const { checkId = "" } = useParams();
	const navigate = useNavigate();
	const confirm = useConfirm();
	const createCheckMutation = useCreateProjectCheckMutation(projectRef);
	const updateCheckMutation = useUpdateProjectCheckMutation(projectRef);
	const deleteCheckMutation = useDeleteProjectCheckMutation(projectRef);
	const batchDeleteCheckMutation = useDeleteProjectChecksMutation(projectRef);
	const selectorPreviewMutation = usePreviewProjectSelectorMutation(projectRef);
	const probesQuery = useQuery({
		...projectQueries.probes(projectRef || ""),
		enabled: Boolean(projectRef),
		select: data => mapApiProbes(data.probes)
	});
	const assignmentsQuery = useQuery({
		...projectQueries.assignments(projectRef || ""),
		enabled: Boolean(projectRef),
		select: data => data.assignments
	});
	const labelsQuery = useQuery({
		...projectQueries.labels(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const checksQuery = useQuery({
		...projectQueries.checks(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const checkRows = useMemo(() => groupChecksByTarget(mapApiChecksWithAssignments(checksQuery.data?.checks, assignmentsQuery.data)), [assignmentsQuery.data, checksQuery.data?.checks]);
	const probes = probesQuery.data || [];
	const [editorMode, setEditorMode] = useState<"idle" | "create">("idle");
	const [checkSearch, setCheckSearch] = useState("");
	const [checkTypeFilter, setCheckTypeFilter] = useState<CheckTypeFilter>("all");
	const [rowSelection, setRowSelection] = useState<RowSelectionState>({});
	const [draftCheckId, setDraftCheckId] = useState("");
	const [checkName, setCheckName] = useState("");
	const [target, setTarget] = useState("");
	const [checkType, setCheckType] = useState<CheckType>("Ping");
	const [interval, setInterval] = useState("30s");
	const [selectedProbes, setSelectedProbes] = useState<string[]>([]);
	const [selectorState, setSelectorState] = useState<SelectorState>({ mode: "all-probes", rules: [], advancedText: "" });
	const [pingConfig, setPingConfig] = useState<PingConfigFormState>(defaultPingConfigFormState);
	const [tcpConfig, setTCPConfig] = useState<TCPConfigFormState>(defaultTCPConfigFormState);
	const [tracerouteConfig, setTracerouteConfig] = useState<TracerouteConfigFormState>(defaultTracerouteConfigFormState);
	const isCreating = editorMode === "create";
	const selectedId = isCreating ? "__new__" : checkId;
	const selectedListCheck = isCreating ? null : checkRows.find(check => check.id === checkId) || null;
	const selectedListApiCheck = checksQuery.data?.checks.find(check => check.id === selectedListCheck?.id) || null;
	const checkDetailQuery = useQuery({
		...projectQueries.checkDetail(projectRef || "", selectedListCheck?.id || ""),
		enabled: Boolean(projectRef && selectedListCheck && !isCreating)
	});
	const selectedApiCheck = isCreating ? null : checkDetailQuery.data?.check || selectedListApiCheck;
	const selectedAssignmentCount = (assignmentsQuery.data ?? []).filter(assignment => assignment.checkId === selectedListCheck?.id).length;
	const selectedCheck = isCreating ? null : selectedApiCheck ? mapApiCheck(selectedApiCheck, selectedAssignmentCount) : selectedListCheck;
	const selectedCheckRows = useMemo(() => checkRows.filter(check => rowSelection[check.id]), [checkRows, rowSelection]);
	const hasSelectedDraft = Boolean(selectedCheck && draftCheckId === selectedCheck.id);
	const activeCheckName = isCreating || hasSelectedDraft ? checkName : selectedCheck?.name || "";
	const activeTarget = isCreating || hasSelectedDraft ? target : selectedCheck?.target || "";
	const activeCheckType = isCreating || hasSelectedDraft ? checkType : selectedCheck?.type || checkType;
	const activeInterval = isCreating || hasSelectedDraft ? interval : selectedCheck?.interval || "30s";
	const activePingConfig = isCreating || hasSelectedDraft ? pingConfig : pingConfigFormStateFromApi(selectedApiCheck);
	const activeTCPConfig = isCreating || hasSelectedDraft ? tcpConfig : tcpConfigFormStateFromApi(selectedApiCheck);
	const activeTracerouteConfig = isCreating || hasSelectedDraft ? tracerouteConfig : tracerouteConfigFormStateFromApi(selectedApiCheck);
	const activeSelectorState = isCreating || hasSelectedDraft ? selectorState : selectorStateFromApi(selectedApiCheck?.selector);
	const assignedProbeNames = (assignmentsQuery.data ?? [])
		.filter(assignment => assignment.checkId === selectedListCheck?.id)
		.map(assignment => assignment.probe?.name || probes.find(probe => probe.id === assignment.probeId)?.name || assignment.probeId);
	const locallyMatchedProbeNames =
		activeSelectorState.mode === "advanced" ? assignedProbeNames : probes.filter(probe => probeMatchesSelector(probe.labelTokens, activeSelectorState)).map(probe => probe.name);
	const activeSelectedProbes = (isCreating || hasSelectedDraft) && selectedProbes.length ? selectedProbes : locallyMatchedProbeNames;
	const selectorOptions = selectorLabelOptions(labelsQuery.data?.labels ?? []);
	const saveCheckMutation = isCreating ? createCheckMutation : updateCheckMutation;
	const checkActionPending = createCheckMutation.isPending || deleteCheckMutation.isPending || batchDeleteCheckMutation.isPending;
	const isEditorOpen = isCreating || Boolean(selectedCheck);

	function resetEditorState() {
		setCheckName("");
		setTarget("");
		setCheckType("Ping");
		setInterval("30s");
		setSelectedProbes([]);
		setSelectorState({ mode: "all-probes", rules: [], advancedText: "" });
		setPingConfig(defaultPingConfigFormState);
		setTCPConfig(defaultTCPConfigFormState);
		setTracerouteConfig(defaultTracerouteConfigFormState);
		setDraftCheckId("");
	}

	function closeEditor() {
		setEditorMode("idle");
		resetEditorState();
		navigate(pathForRoute("checks", { projectRef }));
	}

	function startNewCheck() {
		navigate(pathForRoute("checks", { projectRef }));
		setEditorMode("create");
		resetEditorState();
		setDraftCheckId("__new__");
		setSelectedProbes(probes.map(probe => probe.name));
	}

	function loadCheckDraft(check: CheckDefinition, apiCheck = selectedApiCheck) {
		setDraftCheckId(check.id);
		setCheckName(check.name);
		setTarget(check.target);
		setCheckType(check.type);
		setInterval(check.interval);
		setSelectedProbes([]);
		setSelectorState(selectorStateFromApi(apiCheck?.selector));
		setPingConfig(pingConfigFormStateFromApi(apiCheck));
		setTCPConfig(tcpConfigFormStateFromApi(apiCheck));
		setTracerouteConfig(tracerouteConfigFormStateFromApi(apiCheck));
	}

	function selectCheck(check: CheckDefinition) {
		const apiCheck = checksQuery.data?.checks.find(item => item.id === check.id);

		setEditorMode("idle");
		loadCheckDraft(check, apiCheck);
		navigate(pathForCheckDetail(projectRef, check.id));
	}

	useEffect(() => {
		if (!projectRef || !checkId || checksQuery.isPending || checksQuery.isError || checkRows.some(check => check.id === checkId)) {
			return;
		}

		navigate(pathForRoute("checks", { projectRef }), { replace: true });
	}, [checkId, checkRows, checksQuery.isError, checksQuery.isPending, navigate, projectRef]);

	function prepareSelectedCheckEdit() {
		if (isCreating || !selectedCheck || hasSelectedDraft) {
			return;
		}

		loadCheckDraft(selectedCheck, selectedApiCheck);
	}

	function updatePingConfig(patch: Partial<PingConfigFormState>) {
		prepareSelectedCheckEdit();
		setPingConfig(current => ({ ...current, ...patch }));
	}

	function updateTCPConfig(patch: Partial<TCPConfigFormState>) {
		prepareSelectedCheckEdit();
		setTCPConfig(current => ({ ...current, ...patch }));
	}

	function updateTracerouteConfig(patch: Partial<TracerouteConfigFormState>) {
		prepareSelectedCheckEdit();
		setTracerouteConfig(current => ({ ...current, ...patch }));
	}

	function selectorDraftBase(current: SelectorState) {
		return isCreating || hasSelectedDraft ? current : activeSelectorState;
	}

	function setSelectorMode(mode: SelectorMode) {
		prepareSelectedCheckEdit();
		setSelectedProbes([]);
		setSelectorState(current => {
			const base = selectorDraftBase(current);

			return {
				mode,
				rules: mode === "all-probes" ? [] : base.rules.length ? base.rules : mode === "advanced" ? base.rules : [createSelectorRule(selectorOptions)],
				advancedText:
					mode === "advanced"
						? base.advancedText ||
							JSON.stringify(
								buildSelector({
									...base,
									mode: base.mode === "advanced" ? "all-probes" : base.mode
								}),
								null,
								2
							)
						: base.advancedText
			};
		});
	}

	function addSelectorRule() {
		prepareSelectedCheckEdit();
		setSelectedProbes([]);
		setSelectorState(current => {
			const base = selectorDraftBase(current);

			return {
				...base,
				mode: base.mode === "all-probes" || base.mode === "advanced" ? "all" : base.mode,
				rules: [...base.rules, createSelectorRule(selectorOptions)]
			};
		});
	}

	function removeSelectorRule(ruleId: string) {
		prepareSelectedCheckEdit();
		setSelectedProbes([]);
		setSelectorState(current => {
			const base = selectorDraftBase(current);
			const rules = base.rules.filter(rule => rule.id !== ruleId);
			return {
				...base,
				mode: rules.length ? base.mode : "all-probes",
				rules
			};
		});
	}

	function updateSelectorRule(ruleId: string, patch: Partial<SelectorRule>) {
		prepareSelectedCheckEdit();
		setSelectedProbes([]);
		setSelectorState(current => {
			const base = selectorDraftBase(current);

			return {
				...base,
				mode: base.mode === "all-probes" || base.mode === "advanced" ? "all" : base.mode,
				rules: base.rules.map(rule => (rule.id === ruleId ? { ...rule, ...patch } : rule))
			};
		});
	}

	function updateSelectorRuleKey(ruleId: string, key: string) {
		const values = selectorValuesForKey(selectorOptions, key);
		updateSelectorRule(ruleId, { key, value: values[0] ?? "", values: values.join(", ") });
	}

	function updateAdvancedSelectorText(value: string) {
		prepareSelectedCheckEdit();
		setSelectedProbes([]);
		setSelectorState(current => ({ ...selectorDraftBase(current), mode: "advanced", advancedText: value }));
	}

	function previewSelector() {
		let selector: ApiSelector;

		try {
			prepareSelectedCheckEdit();
			selector = buildSelector(activeSelectorState);
		} catch (error) {
			pushErrorToast(error instanceof Error ? error.message : "Selector JSON is invalid.");
			return;
		}

		selectorPreviewMutation.mutate(
			{ selector },
			{
				onSuccess: data => {
					setSelectedProbes(data.probes.map(probe => probe.name));
					setSelectorState(selectorStateFromApi(data.selector));
				}
			}
		);
	}

	function apiCheckFor(checkId: string) {
		return checksQuery.data?.checks.find(check => check.id === checkId) || null;
	}

	function handleSavedCheck(data: { check: ApiCheck }) {
		setEditorMode("idle");
		setDraftCheckId(data.check.id);
		setCheckName(data.check.name);
		setTarget(data.check.target);
		setCheckType(checkTypeFromApi(data.check.type));
		setInterval(`${data.check.intervalSeconds}s`);
		setSelectedProbes([]);
		setSelectorState(selectorStateFromApi(data.check.selector));
		setPingConfig(pingConfigFormStateFromApi(data.check));
		setTCPConfig(tcpConfigFormStateFromApi(data.check));
		setTracerouteConfig(tracerouteConfigFormStateFromApi(data.check));
		navigate(pathForCheckDetail(projectRef, data.check.id));
	}

	function clearDeletedSelection(checkIds: string[]) {
		setRowSelection(current => {
			const next = { ...current };
			for (const checkId of checkIds) {
				delete next[checkId];
			}

			return next;
		});
	}

	function closeEditorIfActiveDeleted(checkIds: string[]) {
		if (selectedCheck && checkIds.includes(selectedCheck.id)) {
			closeEditor();
		}
	}

	function duplicateCheck(check: CheckDefinition) {
		const apiCheck = apiCheckFor(check.id);
		if (!apiCheck) {
			pushErrorToast("Check details are still loading.");
			return;
		}

		createCheckMutation.mutate(duplicateCheckInput(apiCheck), { onSuccess: handleSavedCheck });
	}

	async function deleteCheck(check: CheckDefinition) {
		const confirmed = await confirm({
			title: `Delete ${check.name}?`,
			message: "This removes the check definition and stops future assignments for matching probes.",
			confirmLabel: "Delete check",
			tone: "danger"
		});

		if (!confirmed) {
			return;
		}

		deleteCheckMutation.mutate(check.id, {
			onSuccess: () => {
				clearDeletedSelection([check.id]);
				closeEditorIfActiveDeleted([check.id]);
			}
		});
	}

	async function deleteSelectedCheck() {
		if (selectedCheck) {
			await deleteCheck(selectedCheck);
		}
	}

	async function deleteSelectedChecks() {
		if (!selectedCheckRows.length) {
			return;
		}

		const previewNames = selectedCheckRows
			.slice(0, 4)
			.map(check => check.name)
			.join(", ");
		const hiddenCount = selectedCheckRows.length - 4;
		const confirmed = await confirm({
			title: `Delete ${selectedCheckRows.length} checks?`,
			message: hiddenCount > 0 ? `${previewNames}, and ${hiddenCount} more will be removed.` : `${previewNames} will be removed.`,
			confirmLabel: "Delete checks",
			tone: "danger"
		});

		if (!confirmed) {
			return;
		}

		const checkIds = selectedCheckRows.map(check => check.id);
		batchDeleteCheckMutation.mutate(checkIds, {
			onSuccess: data => {
				clearDeletedSelection(data.succeededIds);
				closeEditorIfActiveDeleted(data.succeededIds);
				pushToast({
					message: `${data.succeededIds.length} checks removed.`,
					title: "Checks deleted",
					tone: "success"
				});
			},
			onError: error => {
				if (error instanceof BatchCheckDeleteError) {
					clearDeletedSelection(error.succeededIds);
					closeEditorIfActiveDeleted(error.succeededIds);
					pushErrorToast(error.succeededIds.length ? `${error.succeededIds.length} checks were deleted, ${error.failedIds.length} failed.` : `${error.failedIds.length} checks failed to delete.`);
					return;
				}

				pushErrorToast("Selected checks failed to delete.");
			}
		});
	}

	function selectedCheckSummary() {
		const names = selectedCheckRows
			.slice(0, 3)
			.map(check => check.name)
			.join(", ");
		const hiddenCount = selectedCheckRows.length - 3;

		return hiddenCount > 0 ? `${names}, +${hiddenCount}` : names;
	}

	function saveSelectedCheck() {
		let selector: ApiSelector;

		try {
			selector = buildSelector(activeSelectorState);
		} catch (error) {
			pushErrorToast(error instanceof Error ? error.message : "Selector JSON is invalid.");
			return;
		}

		const type = activeCheckType === "Traceroute" ? "traceroute" : activeCheckType === "TCP" ? "tcp" : "ping";
		const body: CreateCheckInput = {
			intervalSeconds: parseIntervalSeconds(activeInterval),
			name: activeCheckName,
			selector,
			target: activeTarget,
			type
		};
		if (type === "traceroute") {
			body.tracerouteConfig = buildTracerouteConfigPayload(activeTracerouteConfig);
		} else if (type === "tcp") {
			body.tcpConfig = buildTCPConfigPayload(activeTCPConfig);
		} else {
			body.pingConfig = buildPingConfigPayload(activePingConfig);
		}

		const options = { onSuccess: handleSavedCheck };

		if (isCreating) {
			createCheckMutation.mutate(body, options);
			return;
		}

		updateCheckMutation.mutate({ checkId: selectedCheck?.id || "", body }, options);
	}

	return (
		<PageStack>
			<ScreenHeader title="Checks" actions={<Button onClick={startNewCheck}>New check</Button>} />

			<div className={styles.checkEditorGrid}>
				<Panel tone="glass" title="Definitions">
					<div className={styles.checkListStack}>
						<FilterGrid className={styles.checkListFilters}>
							<TextField label="Search" placeholder="check name, target, description" value={checkSearch} onChange={event => setCheckSearch(event.currentTarget.value)} />
							<SelectField
								label="Type"
								value={checkTypeFilter}
								onChange={event => setCheckTypeFilter(event.currentTarget.value as CheckTypeFilter)}
								options={[
									{ value: "all", label: "All types" },
									{ value: "ping", label: "Ping" },
									{ value: "tcp", label: "TCP" },
									{ value: "traceroute", label: "Traceroute" }
								]}
							/>
						</FilterGrid>
						{selectedCheckRows.length ? (
							<div className={classNames("ns-cut-frame", styles.batchToolbar)}>
								<div>
									<strong>{selectedCheckRows.length} selected</strong>
									<span>{selectedCheckSummary()}</span>
								</div>
								<div className={styles.batchToolbarActions}>
									<Button type="button" variant="danger" size="sm" disabled={batchDeleteCheckMutation.isPending} onClick={() => void deleteSelectedChecks()}>
										{batchDeleteCheckMutation.isPending ? "Deleting" : "Delete selected"}
									</Button>
								</div>
							</div>
						) : null}
						<ChecksTable
							actionDisabled={checkActionPending}
							checks={checkRows}
							onDeleteCheck={check => void deleteCheck(check)}
							onDuplicateCheck={duplicateCheck}
							onOpenCheck={selectCheck}
							onRowSelectionChange={setRowSelection}
							rowSelection={rowSelection}
							search={checkSearch}
							selectedKey={isCreating ? "__new__" : selectedId}
							typeFilter={checkTypeFilter}
						/>
					</div>
				</Panel>

				{isEditorOpen ? (
					<EditorDrawer open title={isCreating ? "New check" : selectedCheck?.name || "Check"} ariaLabel="Check editor" backLabel="back to checks" onClose={closeEditor}>
						<div className={classNames("ns-scrollbar", styles.checkEditorStack)}>
							<div className={styles.checkEditForm}>
								<TextField
									label="Check name"
									value={activeCheckName}
									disabled={!selectedCheck && !isCreating}
									onChange={event => {
										prepareSelectedCheckEdit();
										setCheckName(event.currentTarget.value);
									}}
								/>
								<TextField
									label="Target"
									value={activeTarget}
									disabled={!selectedCheck && !isCreating}
									onChange={event => {
										prepareSelectedCheckEdit();
										setTarget(event.currentTarget.value);
									}}
								/>
								<SelectField
									label="Check type"
									value={activeCheckType}
									disabled={!isCreating}
									onChange={event => setCheckType(event.currentTarget.value as CheckType)}
									options={[
										{ value: "Ping", label: "Ping" },
										{ value: "TCP", label: "TCP" },
										{ value: "Traceroute", label: "Traceroute" }
									]}
								/>
								<TextField
									label="Interval"
									value={activeInterval}
									onChange={event => {
										prepareSelectedCheckEdit();
										setInterval(event.currentTarget.value);
									}}
								/>
							</div>

							<CheckConfigFields
								checkType={activeCheckType}
								disabled={!selectedCheck && !isCreating}
								pingConfig={activePingConfig}
								tcpConfig={activeTCPConfig}
								tracerouteConfig={activeTracerouteConfig}
								onPingConfigChange={updatePingConfig}
								onTCPConfigChange={updateTCPConfig}
								onTracerouteConfigChange={updateTracerouteConfig}
							/>

							<div className={styles.probeMultiSelect}>
								<FieldLabel>Probe selector</FieldLabel>
								<div className={styles.selectorBuilder}>
									<SelectField label="Match mode" value={activeSelectorState.mode} onChange={event => setSelectorMode(event.currentTarget.value as SelectorMode)} options={selectorModeOptions} />
									{activeSelectorState.mode === "advanced" ? (
										<TextAreaField
											label="Selector JSON"
											rows={8}
											value={activeSelectorState.advancedText}
											onChange={event => updateAdvancedSelectorText(event.currentTarget.value)}
											spellCheck={false}
										/>
									) : null}
									{activeSelectorState.mode !== "all-probes" && activeSelectorState.mode !== "advanced" ? (
										<div className={styles.selectorRuleList}>
											{activeSelectorState.rules.map((rule, index) => {
												const valuesForKey = selectorValuesForKey(selectorOptions, rule.key);

												return (
													<div className={classNames("ns-cut-frame", styles.selectorRule)} key={rule.id}>
														<label className={styles.selectorNegation}>
															<Checkbox checked={rule.negated} onChange={event => updateSelectorRule(rule.id, { negated: event.currentTarget.checked })} />
															<span>not</span>
														</label>
														<TextField label={`Rule ${index + 1} key`} value={rule.key} onChange={event => updateSelectorRuleKey(rule.id, event.currentTarget.value)} autoComplete="off" />
														<SelectField
															label="Operator"
															value={rule.op}
															onChange={event => updateSelectorRule(rule.id, { op: event.currentTarget.value as SelectorLabelOp })}
															options={selectorOpOptions}
														/>
														{rule.op === "eq" ? <TextField label="Value" value={rule.value} onChange={event => updateSelectorRule(rule.id, { value: event.currentTarget.value })} /> : null}
														{rule.op === "in" ? (
															<TextField
																label="Values"
																value={rule.values}
																helper={valuesForKey.length ? valuesForKey.join(", ") : undefined}
																onChange={event => updateSelectorRule(rule.id, { values: event.currentTarget.value })}
															/>
														) : null}
														{rule.op === "exists" ? <div className={styles.selectorExistsValue}>any value</div> : null}
														<Button
															className={classNames(styles.iconAction, styles.iconActionDanger, styles.selectorRuleRemove)}
															type="button"
															variant="ghost"
															size="sm"
															aria-label={`Remove selector rule ${index + 1}`}
															title={`Remove selector rule ${index + 1}`}
															onClick={() => removeSelectorRule(rule.id)}
														>
															<Trash size={15} weight="bold" aria-hidden="true" focusable="false" />
														</Button>
													</div>
												);
											})}
										</div>
									) : null}
									{activeSelectorState.mode === "all-probes" ? <p className={styles.selectorNotice}>Matches every active probe.</p> : null}
									{activeSelectorState.mode !== "advanced" ? (
										<ActionRow>
											<Button type="button" variant="secondary" onClick={addSelectorRule}>
												Add rule
											</Button>
										</ActionRow>
									) : null}
								</div>
								<ActionRow>
									<Button type="button" variant="secondary" disabled={!projectRef || selectorPreviewMutation.isPending} onClick={previewSelector}>
										{selectorPreviewMutation.isPending ? "Previewing" : "Preview selector"}
									</Button>
									<Badge tone="accent">{activeSelectedProbes.length} matched</Badge>
								</ActionRow>
								<div className={classNames("ns-cut-frame", styles.probeSummary)}>{displayProbeSelection(activeSelectedProbes)}</div>
								<div className={styles.capabilityPills}>
									{activeSelectedProbes.map(probe => (
										<Badge key={probe} tone="muted">
											{probe}
										</Badge>
									))}
								</div>
							</div>

							<ActionRow>
								<Button disabled={(!selectedCheck && !isCreating) || !projectRef || !activeCheckName || !activeTarget || saveCheckMutation.isPending} onClick={saveSelectedCheck}>
									{saveCheckMutation.isPending ? "Saving" : isCreating ? "Create check" : "Save check"}
								</Button>
								<Button variant="danger" disabled={!selectedCheck || checkActionPending} onClick={() => void deleteSelectedCheck()}>
									{deleteCheckMutation.isPending ? "Deleting" : "Delete check"}
								</Button>
							</ActionRow>
						</div>
					</EditorDrawer>
				) : null}
			</div>
		</PageStack>
	);
}
