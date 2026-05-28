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
	type IPFamilyFormValue,
	type PingConfigFormState,
	type TCPConfigFormState,
	type TracerouteConfigFormState,
	type TracerouteProtocolFormValue
} from "@/features/checks/data/checkConfig";
import { type CheckDefinition, type CheckType } from "@/features/checks/data/checks";
import { mapApiProbes } from "@/features/probes/api/probeAdapters";
import { useCreateProjectCheckMutation, useDeleteProjectCheckMutation, usePreviewProjectSelectorMutation, useUpdateProjectCheckMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiLabel, ApiSelector, CreateCheckInput } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { ActionRow } from "@/shared/components/ActionRow";
import { useConfirm } from "@/shared/components/confirmContext";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { pushErrorToast } from "@/shared/toast/toastStore";
import { classNames } from "@/shared/utils/classNames";
import { Badge, Button, Checkbox, DataTable, FieldLabel, Panel, SelectField, TextAreaField, TextField, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import styles from "./ChecksPage.module.css";
import { displayProbeSelection } from "./checksPageData";

const checkColumns: DataColumn<CheckDefinition>[] = [
	{ key: "name", label: "Check name" },
	{ key: "type", label: "Type", render: row => <Badge tone="accent">{row.type}</Badge> },
	{ key: "target", label: "Target" },
	{ key: "interval", label: "Interval" },
	{ key: "assigned", label: "Assigned probes" }
];

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

const ipFamilyOptions: Array<{ value: IPFamilyFormValue; label: string }> = [
	{ value: "", label: "Auto" },
	{ value: "inet", label: "IPv4" },
	{ value: "inet6", label: "IPv6" }
];

const tracerouteProtocolOptions: Array<{ value: TracerouteProtocolFormValue; label: string }> = [
	{ value: "icmp", label: "ICMP" },
	{ value: "udp", label: "UDP" }
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

function selectorKeyOptions(options: SelectorLabelOption[], rules: SelectorRule[]) {
	return Array.from(new Set([...options.map(option => option.key), ...rules.map(rule => rule.key.trim()).filter(Boolean)]))
		.sort((a, b) => a.localeCompare(b))
		.map(key => ({ value: key, label: key }));
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
	const confirm = useConfirm();
	const createCheckMutation = useCreateProjectCheckMutation(projectRef);
	const updateCheckMutation = useUpdateProjectCheckMutation(projectRef);
	const deleteCheckMutation = useDeleteProjectCheckMutation(projectRef);
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
	const checkRows = mapApiChecksWithAssignments(checksQuery.data?.checks, assignmentsQuery.data);
	const probes = probesQuery.data || [];
	const [selectedId, setSelectedId] = useState("");
	const [checkName, setCheckName] = useState("");
	const [target, setTarget] = useState("");
	const [checkType, setCheckType] = useState<CheckType>("Ping");
	const [interval, setInterval] = useState("30s");
	const [selectedProbes, setSelectedProbes] = useState<string[]>([]);
	const [selectorState, setSelectorState] = useState<SelectorState>({ mode: "all-probes", rules: [], advancedText: "" });
	const [pingConfig, setPingConfig] = useState<PingConfigFormState>(defaultPingConfigFormState);
	const [tcpConfig, setTCPConfig] = useState<TCPConfigFormState>(defaultTCPConfigFormState);
	const [tracerouteConfig, setTracerouteConfig] = useState<TracerouteConfigFormState>(defaultTracerouteConfigFormState);
	const isCreating = selectedId === "__new__";
	const selectedListCheck = checkRows.find(check => check.id === selectedId) || null;
	const selectedListApiCheck = checksQuery.data?.checks.find(check => check.id === selectedListCheck?.id) || null;
	const checkDetailQuery = useQuery({
		...projectQueries.checkDetail(projectRef || "", selectedListCheck?.id || ""),
		enabled: Boolean(projectRef && selectedListCheck && !isCreating)
	});
	const selectedApiCheck = isCreating ? null : checkDetailQuery.data?.check || selectedListApiCheck;
	const selectedAssignmentCount = (assignmentsQuery.data ?? []).filter(assignment => assignment.checkId === selectedListCheck?.id).length;
	const selectedCheck = isCreating ? null : selectedApiCheck ? mapApiCheck(selectedApiCheck, selectedAssignmentCount) : selectedListCheck;
	const activeCheckName = isCreating || selectedId ? checkName : selectedCheck?.name || "";
	const activeTarget = isCreating || selectedId ? target : selectedCheck?.target || "";
	const activeCheckType = isCreating || selectedId ? checkType : selectedCheck?.type || checkType;
	const activeInterval = isCreating || selectedId ? interval : selectedCheck?.interval || "30s";
	const activePingConfig = selectedId || isCreating ? pingConfig : pingConfigFormStateFromApi(selectedApiCheck);
	const activeTCPConfig = selectedId || isCreating ? tcpConfig : tcpConfigFormStateFromApi(selectedApiCheck);
	const activeTracerouteConfig = selectedId || isCreating ? tracerouteConfig : tracerouteConfigFormStateFromApi(selectedApiCheck);
	const assignedProbeNames = (assignmentsQuery.data ?? [])
		.filter(assignment => assignment.checkId === selectedListCheck?.id)
		.map(assignment => assignment.probe?.name || probes.find(probe => probe.id === assignment.probeId)?.name || assignment.probeId);
	const locallyMatchedProbeNames = selectorState.mode === "advanced" ? assignedProbeNames : probes.filter(probe => probeMatchesSelector(probe.labelTokens, selectorState)).map(probe => probe.name);
	const activeSelectedProbes = selectedProbes.length ? selectedProbes : locallyMatchedProbeNames;
	const selectorOptions = selectorLabelOptions(labelsQuery.data?.labels ?? []);
	const selectorKeys = selectorKeyOptions(selectorOptions, selectorState.rules);
	const saveCheckMutation = isCreating ? createCheckMutation : updateCheckMutation;
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
	}

	function closeEditor() {
		setSelectedId("");
		resetEditorState();
	}

	function startNewCheck() {
		setSelectedId("__new__");
		resetEditorState();
		setSelectedProbes(probes.map(probe => probe.name));
	}

	function selectCheck(check: CheckDefinition) {
		const apiCheck = checksQuery.data?.checks.find(item => item.id === check.id);
		const nextSelectorState = selectorStateFromApi(apiCheck?.selector);

		setSelectedId(check.id);
		setCheckName(check.name);
		setTarget(check.target);
		setCheckType(check.type);
		setInterval(check.interval);
		setSelectedProbes([]);
		setSelectorState(nextSelectorState);
		setPingConfig(pingConfigFormStateFromApi(apiCheck));
		setTCPConfig(tcpConfigFormStateFromApi(apiCheck));
		setTracerouteConfig(tracerouteConfigFormStateFromApi(apiCheck));
	}

	function prepareSelectedCheckEdit() {
		if (isCreating || selectedId || !selectedCheck) {
			return;
		}

		setSelectedId(selectedCheck.id);
		setCheckName(selectedCheck.name);
		setTarget(selectedCheck.target);
		setCheckType(selectedCheck.type);
		setInterval(selectedCheck.interval);
		setSelectedProbes([]);
		setSelectorState(selectorStateFromApi(selectedApiCheck?.selector));
		setPingConfig(pingConfigFormStateFromApi(selectedApiCheck));
		setTCPConfig(tcpConfigFormStateFromApi(selectedApiCheck));
		setTracerouteConfig(tracerouteConfigFormStateFromApi(selectedApiCheck));
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

	function setSelectorMode(mode: SelectorMode) {
		setSelectedProbes([]);
		setSelectorState(current => ({
			mode,
			rules: mode === "all-probes" ? [] : current.rules.length ? current.rules : mode === "advanced" ? current.rules : [createSelectorRule(selectorOptions)],
			advancedText:
				mode === "advanced" ? current.advancedText || JSON.stringify(buildSelector({ ...current, mode: current.mode === "advanced" ? "all-probes" : current.mode }), null, 2) : current.advancedText
		}));
	}

	function addSelectorRule() {
		setSelectedProbes([]);
		setSelectorState(current => ({
			...current,
			mode: current.mode === "all-probes" || current.mode === "advanced" ? "all" : current.mode,
			rules: [...current.rules, createSelectorRule(selectorOptions)]
		}));
	}

	function removeSelectorRule(ruleId: string) {
		setSelectedProbes([]);
		setSelectorState(current => {
			const rules = current.rules.filter(rule => rule.id !== ruleId);
			return {
				...current,
				mode: rules.length ? current.mode : "all-probes",
				rules
			};
		});
	}

	function updateSelectorRule(ruleId: string, patch: Partial<SelectorRule>) {
		setSelectedProbes([]);
		setSelectorState(current => ({
			...current,
			mode: current.mode === "all-probes" || current.mode === "advanced" ? "all" : current.mode,
			rules: current.rules.map(rule => (rule.id === ruleId ? { ...rule, ...patch } : rule))
		}));
	}

	function updateSelectorRuleKey(ruleId: string, key: string) {
		const values = selectorValuesForKey(selectorOptions, key);
		updateSelectorRule(ruleId, { key, value: values[0] ?? "", values: values.join(", ") });
	}

	function updateAdvancedSelectorText(value: string) {
		setSelectedProbes([]);
		setSelectorState(current => ({ ...current, mode: "advanced", advancedText: value }));
	}

	function previewSelector() {
		let selector: ApiSelector;

		try {
			selector = buildSelector(selectorState);
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

	async function deleteSelectedCheck() {
		if (!selectedCheck) {
			return;
		}

		const confirmed = await confirm({
			title: `Delete ${selectedCheck.name}?`,
			message: "This removes the check definition and stops future assignments for matching probes.",
			confirmLabel: "Delete check",
			tone: "danger"
		});

		if (!confirmed) {
			return;
		}

		deleteCheckMutation.mutate(selectedCheck.id, {
			onSuccess: () => {
				closeEditor();
			}
		});
	}

	function saveSelectedCheck() {
		let selector: ApiSelector;

		try {
			selector = buildSelector(selectorState);
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

		const options = {
			onSuccess: (data: Awaited<ReturnType<typeof createCheckMutation.mutateAsync>>) => {
				setSelectedId(data.check.id);
				setCheckName(data.check.name);
				setTarget(data.check.target);
				setCheckType(checkTypeFromApi(data.check.type));
				setInterval(`${data.check.intervalSeconds}s`);
				setSelectedProbes([]);
				setSelectorState(selectorStateFromApi(data.check.selector));
				setPingConfig(pingConfigFormStateFromApi(data.check));
				setTCPConfig(tcpConfigFormStateFromApi(data.check));
				setTracerouteConfig(tracerouteConfigFormStateFromApi(data.check));
			}
		};

		if (isCreating) {
			createCheckMutation.mutate(body, options);
			return;
		}

		updateCheckMutation.mutate({ checkId: selectedCheck?.id || "", body }, options);
	}

	return (
		<PageStack>
			<ScreenHeader
				eyebrow="Check management"
				title="Checks"
				copy="Select a check, edit its schedule and type, assign probes, and review the latest measurement logs."
				actions={<Button onClick={startNewCheck}>New check</Button>}
			/>

			<div className={classNames(styles.checkEditorGrid, !isEditorOpen && styles.checkEditorGridCollapsed)}>
				<Panel tone="glass" eyebrow="Checks list" title="Definitions">
					<div className={styles.checkListStack}>
						<div className={styles.checkListFilters}>
							<TextField label="Search" placeholder="check name, target, description" />
							<SelectField
								label="Type"
								defaultValue="all"
								options={[
									{ value: "all", label: "All types" },
									{ value: "ping", label: "Ping" },
									{ value: "tcp", label: "TCP" },
									{ value: "traceroute", label: "Traceroute" }
								]}
							/>
						</div>
						<DataTable columns={checkColumns} rows={checkRows} getRowKey={row => String(row.id)} selectedKey={isCreating ? "__new__" : selectedId} onRowClick={selectCheck} />
					</div>
				</Panel>

				{isEditorOpen ? (
					<Panel
						className={styles.stickyCheckPanel}
						tone="glass"
						eyebrow={isCreating ? "Create check" : "Edit check"}
						title={isCreating ? "New check" : selectedCheck?.name}
						actions={
							<Button type="button" variant="secondary" onClick={closeEditor}>
								Close
							</Button>
						}
					>
						<div className={classNames("ns-scrollbar", styles.checkEditorStack)}>
							<div className={styles.checkEditForm}>
								<TextField label="Check name" value={activeCheckName} disabled={!selectedCheck && !isCreating} onChange={event => setCheckName(event.currentTarget.value)} />
								<TextField label="Target" value={activeTarget} disabled={!selectedCheck && !isCreating} onChange={event => setTarget(event.currentTarget.value)} />
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
								<TextField label="Interval" value={activeInterval} onChange={event => setInterval(event.currentTarget.value)} />
							</div>

							<div className={styles.checkConfigSection}>
								<FieldLabel>{activeCheckType} config</FieldLabel>
								{activeCheckType === "Traceroute" ? (
									<div className={styles.checkConfigGrid}>
										<SelectField
											label="Protocol"
											value={activeTracerouteConfig.protocol}
											disabled={!selectedCheck && !isCreating}
											onChange={event => updateTracerouteConfig({ protocol: event.currentTarget.value as TracerouteProtocolFormValue })}
											options={tracerouteProtocolOptions}
										/>
										<TextField
											label="Max hops"
											type="number"
											min={1}
											max={64}
											step={1}
											inputMode="numeric"
											value={activeTracerouteConfig.maxHops}
											disabled={!selectedCheck && !isCreating}
											onChange={event => updateTracerouteConfig({ maxHops: event.currentTarget.value })}
										/>
										<TextField
											label="Timeout ms"
											type="number"
											min={1}
											max={60000}
											step={1}
											inputMode="numeric"
											value={activeTracerouteConfig.timeoutMs}
											disabled={!selectedCheck && !isCreating}
											onChange={event => updateTracerouteConfig({ timeoutMs: event.currentTarget.value })}
										/>
										<TextField
											label="Queries per hop"
											type="number"
											min={1}
											max={10}
											step={1}
											inputMode="numeric"
											value={activeTracerouteConfig.queriesPerHop}
											disabled={!selectedCheck && !isCreating}
											onChange={event => updateTracerouteConfig({ queriesPerHop: event.currentTarget.value })}
										/>
										<TextField
											label="Packet size bytes"
											type="number"
											min={1}
											max={65507}
											step={1}
											inputMode="numeric"
											value={activeTracerouteConfig.packetSizeBytes}
											disabled={!selectedCheck && !isCreating}
											onChange={event => updateTracerouteConfig({ packetSizeBytes: event.currentTarget.value })}
										/>
										{activeTracerouteConfig.protocol === "udp" ? (
											<TextField
												label="Port"
												type="number"
												min={1}
												max={65535}
												step={1}
												inputMode="numeric"
												value={activeTracerouteConfig.port}
												disabled={!selectedCheck && !isCreating}
												onChange={event => updateTracerouteConfig({ port: event.currentTarget.value })}
											/>
										) : null}
										<SelectField
											label="IP family"
											value={activeTracerouteConfig.ipFamily}
											disabled={!selectedCheck && !isCreating}
											onChange={event => updateTracerouteConfig({ ipFamily: event.currentTarget.value as IPFamilyFormValue })}
											options={ipFamilyOptions}
										/>
									</div>
								) : activeCheckType === "TCP" ? (
									<div className={styles.checkConfigGrid}>
										<TextField
											label="Port"
											type="number"
											min={1}
											max={65535}
											step={1}
											inputMode="numeric"
											value={activeTCPConfig.port}
											disabled={!selectedCheck && !isCreating}
											onChange={event => updateTCPConfig({ port: event.currentTarget.value })}
										/>
										<TextField
											label="Timeout ms"
											type="number"
											min={1}
											step={1}
											inputMode="numeric"
											value={activeTCPConfig.timeoutMs}
											disabled={!selectedCheck && !isCreating}
											onChange={event => updateTCPConfig({ timeoutMs: event.currentTarget.value })}
										/>
										<SelectField
											label="IP family"
											value={activeTCPConfig.ipFamily}
											disabled={!selectedCheck && !isCreating}
											onChange={event => updateTCPConfig({ ipFamily: event.currentTarget.value as IPFamilyFormValue })}
											options={ipFamilyOptions}
										/>
									</div>
								) : (
									<div className={styles.checkConfigGrid}>
										<TextField
											label="Packet count"
											type="number"
											min={1}
											step={1}
											inputMode="numeric"
											value={activePingConfig.packetCount}
											disabled={!selectedCheck && !isCreating}
											onChange={event => updatePingConfig({ packetCount: event.currentTarget.value })}
										/>
										<TextField
											label="Packet size bytes"
											type="number"
											min={1}
											max={65507}
											step={1}
											inputMode="numeric"
											value={activePingConfig.packetSizeBytes}
											disabled={!selectedCheck && !isCreating}
											onChange={event => updatePingConfig({ packetSizeBytes: event.currentTarget.value })}
										/>
										<TextField
											label="Timeout ms"
											type="number"
											min={1}
											step={1}
											inputMode="numeric"
											value={activePingConfig.timeoutMs}
											disabled={!selectedCheck && !isCreating}
											onChange={event => updatePingConfig({ timeoutMs: event.currentTarget.value })}
										/>
										<SelectField
											label="IP family"
											value={activePingConfig.ipFamily}
											disabled={!selectedCheck && !isCreating}
											onChange={event => updatePingConfig({ ipFamily: event.currentTarget.value as IPFamilyFormValue })}
											options={ipFamilyOptions}
										/>
									</div>
								)}
							</div>

							<div className={styles.probeMultiSelect}>
								<FieldLabel>Probe selector</FieldLabel>
								<div className={styles.selectorBuilder}>
									<SelectField label="Match mode" value={selectorState.mode} onChange={event => setSelectorMode(event.currentTarget.value as SelectorMode)} options={selectorModeOptions} />
									{selectorState.mode === "advanced" ? (
										<TextAreaField label="Selector JSON" rows={8} value={selectorState.advancedText} onChange={event => updateAdvancedSelectorText(event.currentTarget.value)} spellCheck={false} />
									) : null}
									{selectorState.mode !== "all-probes" && selectorState.mode !== "advanced" ? (
										<div className={styles.selectorRuleList}>
											{selectorState.rules.map((rule, index) => {
												const valuesForKey = selectorValuesForKey(selectorOptions, rule.key);

												return (
													<div className={classNames("ns-cut-frame", styles.selectorRule)} key={rule.id}>
														<label className={styles.selectorNegation}>
															<Checkbox checked={rule.negated} onChange={event => updateSelectorRule(rule.id, { negated: event.currentTarget.checked })} />
															<span>not</span>
														</label>
														<SelectField
															label={`Rule ${index + 1} key`}
															value={rule.key}
															onChange={event => updateSelectorRuleKey(rule.id, event.currentTarget.value)}
															options={selectorKeys.length ? selectorKeys : [{ value: "", label: "No labels" }]}
														/>
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
														<Button type="button" variant="secondary" onClick={() => removeSelectorRule(rule.id)}>
															Remove
														</Button>
													</div>
												);
											})}
										</div>
									) : null}
									{selectorState.mode === "all-probes" ? <p className={styles.selectorNotice}>Matches every active probe.</p> : null}
									{selectorState.mode !== "advanced" ? (
										<ActionRow>
											<Button type="button" variant="secondary" disabled={!selectorOptions.length} onClick={addSelectorRule}>
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
								<Button variant="danger" disabled={!selectedCheck || deleteCheckMutation.isPending} onClick={() => void deleteSelectedCheck()}>
									{deleteCheckMutation.isPending ? "Deleting" : "Delete check"}
								</Button>
							</ActionRow>
						</div>
					</Panel>
				) : null}
			</div>
		</PageStack>
	);
}
