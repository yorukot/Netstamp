import { type CheckDefinition, type CheckType } from "@/features/checks/data/checks";
import { probes } from "@/features/probes/data/probes";
import { ActionRow } from "@/shared/components/ActionRow";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { classNames } from "@/shared/utils/classNames";
import { toneForStatus } from "@/shared/utils/statusTone";
import { Badge, Button, Checkbox, DataTable, FieldLabel, Panel, SelectField, TextField, type DataColumn } from "@netstamp/ui";
import { useState } from "react";
import styles from "./ChecksPage.module.css";
import { assignedProbeNames, checkRows, displayProbeSelection, logsForCheck, type LogRow } from "./checksPageData";

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

export function ChecksPage() {
	const [selectedId, setSelectedId] = useState("api-latency");
	const [checkType, setCheckType] = useState<CheckType>("Ping");
	const [interval, setInterval] = useState("30s");
	const [jitter, setJitter] = useState("4s");
	const [enabled, setEnabled] = useState("enabled");
	const [selectedProbes, setSelectedProbes] = useState(() => assignedProbeNames("api-latency"));
	const selectedCheck = checkRows.find(check => check.id === selectedId) || checkRows[0];
	const selectedLogs = logsForCheck(selectedCheck, selectedProbes);

	function selectCheck(check: CheckDefinition) {
		setSelectedId(check.id);
		setCheckType(check.type);
		setInterval(check.interval);
		setJitter(check.jitter);
		setEnabled(check.status.toLowerCase().includes("disabled") ? "disabled" : "enabled");
		setSelectedProbes(assignedProbeNames(check.id));
	}

	function toggleProbe(probeName: string) {
		setSelectedProbes(current => (current.includes(probeName) ? current.filter(value => value !== probeName) : [...current, probeName]));
	}

	return (
		<PageStack>
			<ScreenHeader
				eyebrow="Check management"
				title="Checks"
				copy="Select a check, edit its schedule and type, assign probes, and review the latest measurement logs."
				actions={<Button>New check</Button>}
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
						<DataTable columns={checkColumns} rows={checkRows} getRowKey={row => String(row.id)} selectedKey={selectedId} onRowClick={selectCheck} />
					</div>
				</Panel>

				<Panel className={styles.stickyCheckPanel} tone="glass" eyebrow="Edit check" title={selectedCheck.name}>
					<div className={classNames("ns-scrollbar", styles.checkEditorStack)}>
						<div className={styles.checkEditForm}>
							<TextField label="Check name" defaultValue={selectedCheck.name} />
							<TextField label="Target" defaultValue={selectedCheck.target} />
							<SelectField
								label="Check type"
								value={checkType}
								onChange={event => setCheckType(event.currentTarget.value as CheckType)}
								options={[
									{ value: "Ping", label: "Ping" },
									{ value: "Traceroute", label: "Traceroute" },
									{ value: "DNS", label: "DNS" }
								]}
							/>
							<TextField label="Interval" value={interval} onChange={event => setInterval(event.currentTarget.value)} />
							<TextField label="Jitter" value={jitter} onChange={event => setJitter(event.currentTarget.value)} />
							<SelectField
								label="Enabled"
								value={enabled}
								onChange={event => setEnabled(event.currentTarget.value)}
								options={[
									{ value: "enabled", label: "Enabled" },
									{ value: "disabled", label: "Disabled" }
								]}
							/>
						</div>

						<div className={styles.probeMultiSelect}>
							<FieldLabel>Assign probes</FieldLabel>
							<details>
								<summary className={classNames("ns-cut-frame", styles.probeSummary)}>{displayProbeSelection(selectedProbes)}</summary>
								<div className={classNames("ns-scrollbar", styles.probeOptionList)}>
									{probes.map(probe => (
										<label key={probe.id}>
											<Checkbox checked={selectedProbes.includes(probe.name)} onChange={() => toggleProbe(probe.name)} />
											<span>{probe.name}</span>
											<small>{probe.location}</small>
										</label>
									))}
								</div>
							</details>
							<div className={styles.capabilityPills}>
								{selectedProbes.map(probe => (
									<Badge key={probe} tone="muted">
										{probe}
									</Badge>
								))}
							</div>
						</div>

						<ActionRow>
							<Button>Save check</Button>
							<Button variant="outline">Run now</Button>
						</ActionRow>

						<DataTable columns={logColumns} rows={selectedLogs} />
					</div>
				</Panel>
			</div>
		</PageStack>
	);
}
