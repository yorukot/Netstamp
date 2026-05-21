import { mapApiChecks, mapApiChecksWithAssignments, parseIntervalSeconds } from "@/features/checks/api/checkAdapters";
import { mapApiMeasurements, type LogRow } from "@/features/checks/api/resultAdapters";
import { type CheckDefinition, type CheckType } from "@/features/checks/data/checks";
import { mapApiProbes } from "@/features/probes/api/probeAdapters";
import { useCreateProjectCheckMutation, useDeleteProjectCheckMutation, usePreviewProjectSelectorMutation, useUpdateProjectCheckMutation } from "@/shared/api/mutations";
import { projectQueries } from "@/shared/api/queries";
import type { ApiSelector, CreateCheckInput } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { ActionRow } from "@/shared/components/ActionRow";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { classNames } from "@/shared/utils/classNames";
import { toneForStatus } from "@/shared/utils/statusTone";
import { Badge, Button, Checkbox, DataTable, FieldLabel, Panel, SelectField, TextAreaField, TextField, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import styles from "./ChecksPage.module.css";
import { displayProbeSelection } from "./checksPageData";

const checkColumns: DataColumn<CheckDefinition>[] = [
	{ key: "name", label: "Check name" },
	{ key: "type", label: "Type", render: row => <Badge tone="accent">{row.type}</Badge> },
	{ key: "target", label: "Target" },
	{ key: "status", label: "Latest status", render: row => <Badge tone={toneForStatus(row.status)}>{row.status}</Badge> },
	{ key: "interval", label: "Interval" },
	{ key: "assigned", label: "Assigned probes" }
];

const logColumns: DataColumn<LogRow>[] = [
	{ key: "time", label: "Time" },
	{ key: "check", label: "Check" },
	{ key: "probe", label: "Probe" },
	{ key: "status", label: "Status", render: row => <Badge tone={toneForStatus(row.status)}>{row.status}</Badge> },
	{ key: "latency", label: "Latency" },
	{ key: "event", label: "Event" }
];

function formatSelector(selector: ApiSelector | null | undefined) {
	return JSON.stringify(selector ?? {}, null, 2);
}

function isSelector(value: unknown): value is ApiSelector {
	return typeof value === "object" && value !== null && !Array.isArray(value);
}

export function ChecksPage() {
	const { projectRef } = useCurrentProject();
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
	const [jitter, setJitter] = useState("4s");
	const [enabled, setEnabled] = useState("enabled");
	const [selectedProbes, setSelectedProbes] = useState<string[]>([]);
	const [selectorDraft, setSelectorDraft] = useState("{}");
	const [selectorError, setSelectorError] = useState("");
	const isCreating = selectedId === "__new__";
	const selectedListCheck = checkRows.find(check => check.id === selectedId) || checkRows[0] || null;
	const checkDetailQuery = useQuery({
		...projectQueries.checkDetail(projectRef || "", selectedListCheck?.id || ""),
		enabled: Boolean(projectRef && selectedListCheck && !isCreating),
		select: data => mapApiChecks([data.check], probesQuery.data)[0]
	});
	const selectedCheck = isCreating ? null : checkDetailQuery.data || selectedListCheck;
	const measurementsQuery = useQuery({
		...projectQueries.measurements(projectRef || "", { checkId: selectedListCheck?.id || "", limit: 10 }),
		enabled: Boolean(projectRef && selectedListCheck && !isCreating),
		select: data => mapApiMeasurements(data.measurements, probes, checkRows)
	});
	const activeCheckName = checkName || selectedCheck?.name || "";
	const activeTarget = target || selectedCheck?.target || "";
	const activeCheckType = isCreating || selectedId ? checkType : selectedCheck?.type || checkType;
	const activeInterval = interval || selectedCheck?.interval || "30s";
	const activeJitter = jitter || selectedCheck?.jitter || "4s";
	const activeEnabled = selectedId ? enabled : selectedCheck?.status.toLowerCase().includes("disabled") ? "disabled" : enabled;
	const assignedProbeNames = (assignmentsQuery.data ?? [])
		.filter(assignment => assignment.checkId === selectedListCheck?.id)
		.map(assignment => assignment.probe?.name || probes.find(probe => probe.id === assignment.probeId)?.name || assignment.probeId);
	const activeSelectedProbes = selectedProbes.length ? selectedProbes : assignedProbeNames.length ? assignedProbeNames : probes.map(probe => probe.name);
	const saveCheckMutation = isCreating ? createCheckMutation : updateCheckMutation;
	const selectedLogs = selectedCheck ? (measurementsQuery.data ?? []) : [];

	function startNewCheck() {
		setSelectedId("__new__");
		setCheckName("");
		setTarget("");
		setCheckType("Ping");
		setInterval("30s");
		setJitter("4s");
		setEnabled("enabled");
		setSelectedProbes(probes.map(probe => probe.name));
		setSelectorDraft("{}");
		setSelectorError("");
	}

	function selectCheck(check: CheckDefinition) {
		const apiCheck = checksQuery.data?.checks.find(item => item.id === check.id);

		setSelectedId(check.id);
		setCheckName(check.name);
		setTarget(check.target);
		setCheckType(check.type);
		setInterval(check.interval);
		setJitter(check.jitter);
		setEnabled(check.status.toLowerCase().includes("disabled") ? "disabled" : "enabled");
		setSelectedProbes(probes.map(probe => probe.name));
		setSelectorDraft(formatSelector(apiCheck?.selector));
		setSelectorError("");
	}

	function parseSelectorDraft() {
		try {
			const parsed: unknown = JSON.parse(selectorDraft || "{}");

			if (!isSelector(parsed)) {
				setSelectorError("Selector must be a JSON object.");
				return null;
			}

			setSelectorError("");
			return parsed;
		} catch {
			setSelectorError("Selector must be valid JSON.");
			return null;
		}
	}

	function previewSelector() {
		const selector = parseSelectorDraft();

		if (!selector) {
			return;
		}

		selectorPreviewMutation.mutate(
			{ selector },
			{
				onSuccess: data => {
					setSelectorDraft(formatSelector(data.selector));
					setSelectedProbes(data.probes.map(probe => probe.name));
				}
			}
		);
	}

	function deleteSelectedCheck() {
		if (!selectedCheck || !window.confirm(`Delete check ${selectedCheck.name}?`)) {
			return;
		}

		deleteCheckMutation.mutate(selectedCheck.id, {
			onSuccess: () => {
				setSelectedId("");
				setCheckName("");
				setTarget("");
			}
		});
	}

	function saveSelectedCheck() {
		const selector = parseSelectorDraft();

		if (!selector) {
			return;
		}

		const body: CreateCheckInput = {
			intervalSeconds: parseIntervalSeconds(activeInterval),
			name: activeCheckName,
			selector,
			target: activeTarget,
			type: activeCheckType === "Traceroute" ? "traceroute" : "ping"
		};
		const options = {
			onSuccess: (data: Awaited<ReturnType<typeof createCheckMutation.mutateAsync>>) => {
				setSelectedId(data.check.id);
				setCheckName(data.check.name);
				setTarget(data.check.target);
				setInterval(`${data.check.intervalSeconds}s`);
				setSelectorDraft(formatSelector(data.check.selector));
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

			<div className={styles.checkEditorGrid}>
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
									{ value: "traceroute", label: "Traceroute" },
									{ value: "dns", label: "DNS" }
								]}
							/>
							<SelectField
								label="Enabled"
								defaultValue="all"
								options={[
									{ value: "all", label: "All states" },
									{ value: "enabled", label: "Enabled" },
									{ value: "disabled", label: "Disabled" }
								]}
							/>
						</div>
						<DataTable columns={checkColumns} rows={checkRows} getRowKey={row => String(row.id)} selectedKey={isCreating ? "__new__" : selectedCheck?.id || ""} onRowClick={selectCheck} />
					</div>
				</Panel>

				<Panel className={styles.stickyCheckPanel} tone="glass" eyebrow={isCreating ? "Create check" : "Edit check"} title={isCreating ? "New check" : selectedCheck?.name || "No check selected"}>
					<div className={classNames("ns-scrollbar", styles.checkEditorStack)}>
						<div className={styles.checkEditForm}>
							<TextField label="Check name" value={activeCheckName} disabled={!selectedCheck && !isCreating} onChange={event => setCheckName(event.currentTarget.value)} />
							<TextField label="Target" value={activeTarget} disabled={!selectedCheck && !isCreating} onChange={event => setTarget(event.currentTarget.value)} />
							<SelectField label="Check type" value={activeCheckType} onChange={event => setCheckType(event.currentTarget.value as CheckType)} options={[{ value: "Ping", label: "Ping" }]} />
							<TextField label="Interval" value={activeInterval} onChange={event => setInterval(event.currentTarget.value)} />
							<TextField label="Jitter" value={activeJitter} onChange={event => setJitter(event.currentTarget.value)} />
							<SelectField
								label="Enabled"
								value={activeEnabled}
								onChange={event => setEnabled(event.currentTarget.value)}
								options={[
									{ value: "enabled", label: "Enabled" },
									{ value: "disabled", label: "Disabled" }
								]}
							/>
						</div>

						<div className={styles.probeMultiSelect}>
							<FieldLabel>Selector preview</FieldLabel>
							<TextAreaField
								label="Selector JSON"
								value={selectorDraft}
								rows={6}
								error={selectorError}
								helper="Use {}, label, all, any, and not selector nodes."
								onChange={event => {
									setSelectorDraft(event.currentTarget.value);
									setSelectorError("");
								}}
							/>
							<ActionRow>
								<Button type="button" variant="secondary" disabled={!projectRef || selectorPreviewMutation.isPending} onClick={previewSelector}>
									{selectorPreviewMutation.isPending ? "Previewing" : "Preview selector"}
								</Button>
								{selectorPreviewMutation.data ? <Badge tone="accent">{selectorPreviewMutation.data.matchedCount} matched</Badge> : null}
							</ActionRow>
							<details>
								<summary className={classNames("ns-cut-frame", styles.probeSummary)}>{displayProbeSelection(activeSelectedProbes)}</summary>
								<div className={classNames("ns-scrollbar", styles.probeOptionList)}>
									{probes.map(probe => (
										<label key={probe.id}>
											<Checkbox checked={activeSelectedProbes.includes(probe.name)} disabled />
											<span>{probe.name}</span>
											<small>{probe.location}</small>
										</label>
									))}
								</div>
							</details>
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
							<Button variant="danger" disabled={!selectedCheck || deleteCheckMutation.isPending} onClick={deleteSelectedCheck}>
								{deleteCheckMutation.isPending ? "Deleting" : "Delete check"}
							</Button>
						</ActionRow>

						<DataTable columns={logColumns} rows={selectedLogs} />
					</div>
				</Panel>
			</div>
		</PageStack>
	);
}
