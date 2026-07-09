import { mapApiCheck, mapApiChecksWithAssignments, parseIntervalSeconds, validateIntervalSeconds } from "@/features/checks/api/checkAdapters";
import {
	buildPingConfigPayload,
	buildTCPConfigPayload,
	buildTracerouteConfigPayload,
	defaultPingConfigFormState,
	defaultTCPConfigFormState,
	defaultTracerouteConfigFormState,
	firstConfigValidationError,
	pingConfigFormStateFromApi,
	tcpConfigFormStateFromApi,
	tracerouteConfigFormStateFromApi,
	validatePingConfig,
	validateTCPConfig,
	validateTracerouteConfig,
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
import type { ApiCheck, ApiSelector, CreateCheckInput } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { useConfirm } from "@/shared/components/confirmContext";
import { EditorDrawer } from "@/shared/components/EditorDrawer";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { pushErrorToast, pushToast } from "@/shared/toast/toastStore";
import { classNames } from "@/shared/utils/classNames";
import { ActionRow, Badge, Button, Checkbox, FieldLabel, FilterGrid, IconButton, SelectField, TextAreaField, TextField } from "@netstamp/ui";
import { TrashIcon } from "@phosphor-icons/react/dist/csr/Trash";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { CheckConfigFields } from "./CheckConfigFields";
import { checkTypeFromApi, duplicateCheckInput, isCheckTypeFilter, pathWithSearch } from "./checkPageHelpers";
import styles from "./ChecksPage.module.css";
import { displayProbeSelection, groupChecksByTarget } from "./checksPageData";
import { ChecksTable, type CheckRowSelectionState, type CheckTypeFilter } from "./ChecksTable";
import {
	buildSelector,
	createSelectorRule,
	probeMatchesSelector,
	selectorLabelOptions,
	selectorModeOptions,
	selectorOpOptions,
	selectorStateFromApi,
	selectorValuesForKey,
	type SelectorLabelOp,
	type SelectorMode,
	type SelectorRule,
	type SelectorState
} from "./selectorState";

export function ChecksPage() {
	const { projectRef } = useCurrentProject();
	const { checkId = "" } = useParams();
	const [searchParams, setSearchParams] = useSearchParams();
	const navigate = useNavigate();
	const confirm = useConfirm();
	const createCheckMutation = useCreateProjectCheckMutation(projectRef);
	const updateCheckMutation = useUpdateProjectCheckMutation(projectRef);
	const deleteCheckMutation = useDeleteProjectCheckMutation(projectRef);
	const batchDeleteCheckMutation = useDeleteProjectChecksMutation(projectRef, { suppressGlobalErrorToast: true });
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
	const searchParamString = searchParams.toString();
	const checkSearch = searchParams.get("q") ?? "";
	const rawCheckTypeFilter = searchParams.get("type");
	const checkTypeFilter: CheckTypeFilter = isCheckTypeFilter(rawCheckTypeFilter) ? rawCheckTypeFilter : "all";
	const [rowSelection, setRowSelection] = useState<CheckRowSelectionState>({});
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
	const [copiedCheckFields, setCopiedCheckFields] = useState<Pick<CreateCheckInput, "description" | "labelIds">>({});
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
	const activeIntervalValidation = validateIntervalSeconds(activeInterval);
	const activePingValidation = validatePingConfig(activePingConfig);
	const activeTCPValidation = validateTCPConfig(activeTCPConfig);
	const activeTracerouteValidation = validateTracerouteConfig(activeTracerouteConfig);
	const activeConfigError =
		activeCheckType === "Traceroute"
			? firstConfigValidationError(activeTracerouteConfig.protocol === "udp" ? activeTracerouteValidation : { ...activeTracerouteValidation, port: { value: 1, error: "" } })
			: activeCheckType === "TCP"
				? firstConfigValidationError(activeTCPValidation)
				: firstConfigValidationError(activePingValidation);
	const activeFormError = activeIntervalValidation.error || activeConfigError;
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
		setCopiedCheckFields({});
		setDraftCheckId("");
	}

	function closeEditor() {
		setEditorMode("idle");
		resetEditorState();
		navigate(pathWithSearch(pathForRoute("checks", { projectRef }), searchParamString));
	}

	function startNewCheck() {
		navigate(pathWithSearch(pathForRoute("checks", { projectRef }), searchParamString));
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
		navigate(pathWithSearch(pathForCheckDetail(projectRef, check.id), searchParamString));
	}

	useEffect(() => {
		if (!projectRef || !checkId || checksQuery.isPending || checksQuery.isError || checkRows.some(check => check.id === checkId)) {
			return;
		}

		navigate(pathWithSearch(pathForRoute("checks", { projectRef }), searchParamString), { replace: true });
	}, [checkId, checkRows, checksQuery.isError, checksQuery.isPending, navigate, projectRef, searchParamString]);

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

	function assignedProbeNamesFor(checkId: string) {
		return (assignmentsQuery.data ?? [])
			.filter(assignment => assignment.checkId === checkId)
			.map(assignment => assignment.probe?.name || probes.find(probe => probe.id === assignment.probeId)?.name || assignment.probeId);
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
		navigate(pathWithSearch(pathForCheckDetail(projectRef, data.check.id), searchParamString));
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

		const body = duplicateCheckInput(apiCheck);
		const copiedSelectorState = selectorStateFromApi(body.selector);

		navigate(pathWithSearch(pathForRoute("checks", { projectRef }), searchParamString));
		setEditorMode("create");
		setDraftCheckId("__new__");
		setCheckName(body.name);
		setTarget(body.target);
		setCheckType(checkTypeFromApi(body.type));
		setInterval(`${body.intervalSeconds}s`);
		setSelectorState(copiedSelectorState);
		setSelectedProbes(copiedSelectorState.mode === "advanced" ? assignedProbeNamesFor(check.id) : []);
		setPingConfig(pingConfigFormStateFromApi(apiCheck));
		setTCPConfig(tcpConfigFormStateFromApi(apiCheck));
		setTracerouteConfig(tracerouteConfigFormStateFromApi(apiCheck));
		setCopiedCheckFields({
			...(body.description ? { description: body.description } : {}),
			...(body.labelIds?.length ? { labelIds: body.labelIds } : {})
		});
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

	function updateCheckSearchParam(key: "q" | "type", value: string) {
		const next = new URLSearchParams(searchParamString);

		if ((key === "q" && !value.trim()) || (key === "type" && value === "all")) {
			next.delete(key);
		} else {
			next.set(key, value);
		}

		setSearchParams(next, { replace: true });
	}

	function saveSelectedCheck() {
		let selector: ApiSelector;

		try {
			if (activeFormError) {
				throw new Error(activeFormError);
			}

			selector = buildSelector(activeSelectorState);
		} catch (error) {
			pushErrorToast(error instanceof Error ? error.message : "Check form is invalid.");
			return;
		}

		const type = activeCheckType === "Traceroute" ? "traceroute" : activeCheckType === "TCP" ? "tcp" : "ping";
		let body: CreateCheckInput;

		try {
			body = {
				intervalSeconds: parseIntervalSeconds(activeInterval),
				name: activeCheckName,
				selector,
				target: activeTarget,
				type
			};
			if (isCreating) {
				Object.assign(body, copiedCheckFields);
			}
			if (type === "traceroute") {
				body.tracerouteConfig = buildTracerouteConfigPayload(activeTracerouteConfig);
			} else if (type === "tcp") {
				body.tcpConfig = buildTCPConfigPayload(activeTCPConfig);
			} else {
				body.pingConfig = buildPingConfigPayload(activePingConfig);
			}
		} catch (error) {
			pushErrorToast(error instanceof Error ? error.message : "Check form is invalid.");
			return;
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
				<div className={styles.checkListStack}>
					<FilterGrid className={styles.checkListFilters}>
						<TextField label="Search" placeholder="check name, target, description" value={checkSearch} onChange={event => updateCheckSearchParam("q", event.currentTarget.value)} />
						<SelectField
							label="Type"
							value={checkTypeFilter}
							onChange={event => updateCheckSearchParam("type", event.currentTarget.value)}
							options={[
								{ value: "all", label: "All types" },
								{ value: "ping", label: "Ping" },
								{ value: "tcp", label: "TCP" },
								{ value: "traceroute", label: "Traceroute" }
							]}
						/>
					</FilterGrid>
					<ChecksTable
						actionDisabled={checkActionPending}
						batchDeleteDisabled={!selectedCheckRows.length}
						batchDeletePending={batchDeleteCheckMutation.isPending}
						checks={checkRows}
						onDeleteCheck={check => void deleteCheck(check)}
						onDeleteSelectedChecks={() => void deleteSelectedChecks()}
						onDuplicateCheck={duplicateCheck}
						onOpenCheck={selectCheck}
						onRowSelectionChange={setRowSelection}
						rowSelection={rowSelection}
						search={checkSearch}
						selectedKey={isCreating ? "__new__" : selectedId}
						selectedSummary={
							<>
								<strong>{selectedCheckRows.length} selected</strong>
								<span>{selectedCheckSummary()}</span>
							</>
						}
						typeFilter={checkTypeFilter}
					/>
				</div>

				{isEditorOpen ? (
					<EditorDrawer open title={isCreating ? "New check" : selectedCheck?.name || "Check"} ariaLabel="Check editor" onClose={closeEditor}>
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
									inputMode="numeric"
									helper="Whole seconds, for example 30s."
									error={activeIntervalValidation.error || undefined}
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
								pingValidation={activePingValidation}
								tcpConfig={activeTCPConfig}
								tcpValidation={activeTCPValidation}
								tracerouteConfig={activeTracerouteConfig}
								tracerouteValidation={activeTracerouteValidation}
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
													<div className={styles.selectorRule} key={rule.id}>
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
														<IconButton
															className={classNames(styles.iconAction, styles.iconActionDanger, styles.selectorRuleRemove)}
															aria-label={`Remove selector rule ${index + 1}`}
															danger
															onClick={() => removeSelectorRule(rule.id)}
														>
															<TrashIcon size={15} weight="bold" aria-hidden="true" focusable="false" />
														</IconButton>
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
								<div className={styles.probeSummary}>{displayProbeSelection(activeSelectedProbes)}</div>
								<div className={styles.capabilityPills}>
									{activeSelectedProbes.map(probe => (
										<Badge key={probe} tone="muted">
											{probe}
										</Badge>
									))}
								</div>
							</div>

							<ActionRow>
								<Button
									disabled={(!selectedCheck && !isCreating) || !projectRef || !activeCheckName || !activeTarget || Boolean(activeFormError) || saveCheckMutation.isPending}
									onClick={saveSelectedCheck}
								>
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
