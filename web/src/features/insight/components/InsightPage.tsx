import { assignments, checks, results, type CheckDefinition } from "@/features/checks/data/checks";
import { dnsData, latencyData, routeDiffData } from "@/features/insight/data/series";
import { probes, type Probe } from "@/features/probes/data/probes";
import { BodyCopy } from "@/shared/components/BodyCopy";
import { ChartPanel } from "@/shared/components/ChartPanel";
import { KeyValueGrid } from "@/shared/components/KeyValueGrid";
import { PageStack } from "@/shared/components/PageStack";
import { ResponsiveGrid } from "@/shared/components/ResponsiveGrid";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { lineChartOption } from "@/shared/utils/chartOptions";
import { toneForStatus } from "@/shared/utils/statusTone";
import { Badge, DataTable, Panel, SelectField, Surface, type DataColumn } from "@netstamp/ui";
import { useState } from "react";
import styles from "./InsightPage.module.css";

type InsightView = "probe" | "target";

interface ResultRow {
	time: string;
	probe: string;
	check: string;
	status: string;
	latency: string;
	loss: string;
	metadata: string;
}

interface EntityDetail {
	label: string;
	value: string;
}

interface GraphCard {
	key: string;
	eyebrow: string;
	title: string;
	copy: string;
	metric: string;
	values: number[];
	baseline: number[];
}

const timeOptions = [
	{ value: "1h", label: "Last 1 hour" },
	{ value: "24h", label: "Last 24 hours" },
	{ value: "7d", label: "Last 7 days" }
];

const viewOptions = [
	{ value: "probe", label: "Choose a probe" },
	{ value: "target", label: "Choose a target" }
];

const resultRows: ResultRow[] = results.map(([time, probe, check, status, latency, loss, metadata]) => ({
	time,
	probe,
	check,
	status,
	latency,
	loss,
	metadata
}));

const resultColumns: DataColumn<ResultRow>[] = [
	{ key: "time", label: "Time" },
	{ key: "probe", label: "Probe" },
	{ key: "check", label: "Check" },
	{ key: "status", label: "Status", render: row => <Badge tone={toneForStatus(row.status)}>{row.status}</Badge> },
	{ key: "latency", label: "Latency" },
	{ key: "loss", label: "Loss" },
	{ key: "metadata", label: "Raw metadata" }
];

function checkSeries(check: CheckDefinition) {
	if (check.type === "DNS") {
		return dnsData;
	}

	if (check.type === "Traceroute") {
		return routeDiffData.map((value, index) => 54 + value * 9 + index * 2);
	}

	return latencyData;
}

function shiftSeries(values: number[], seed: number) {
	return values.map((value, index) => Math.max(0, Math.round(value + (((seed + 1) * 4 + index * 3) % 18) - 7)));
}

function timeLabel(value: string) {
	return timeOptions.find(option => option.value === value)?.label || value;
}

function assignedLabel(probeName: string, checkId: string) {
	return assignments.some(([probe, check]) => probe === probeName && check === checkId) ? "assigned" : "available";
}

function detailsForProbe(probe: Probe): EntityDetail[] {
	return [
		{ label: "Status", value: probe.status },
		{ label: "Location", value: probe.location },
		{ label: "Network", value: probe.asn },
		{ label: "Last heartbeat", value: probe.lastHeartbeat }
	];
}

function detailsForTarget(check: CheckDefinition): EntityDetail[] {
	return [
		{ label: "Target", value: check.target },
		{ label: "Family", value: check.type },
		{ label: "Interval", value: check.interval },
		{ label: "Latest", value: check.latest }
	];
}

export function InsightPage() {
	const [timeRange, setTimeRange] = useState("24h");
	const [view, setView] = useState<InsightView>("probe");
	const [selectedProbeId, setSelectedProbeId] = useState(probes[0]?.id || "");
	const [selectedTargetId, setSelectedTargetId] = useState(checks[0]?.id || "");

	const selectedProbe = probes.find(probe => probe.id === selectedProbeId) || probes[0];
	const selectedTarget = checks.find(check => check.id === selectedTargetId) || checks[0];
	const selectedTitle = view === "probe" ? selectedProbe.name : selectedTarget.target;
	const selectedDetails = view === "probe" ? detailsForProbe(selectedProbe) : detailsForTarget(selectedTarget);
	const pickerOptions =
		view === "probe" ? probes.map(probe => ({ value: probe.id, label: `${probe.name} · ${probe.location}` })) : checks.map(check => ({ value: check.id, label: `${check.target} · ${check.type}` }));

	const graphCards: GraphCard[] =
		view === "probe"
			? checks.map((check, index) => ({
					key: check.id,
					eyebrow: `${timeLabel(timeRange)} · Illustration`,
					title: `${selectedProbe.name} → ${check.target}`,
					copy: `${check.type} insight for ${assignedLabel(selectedProbe.name, check.id)} probe-target measurement.`,
					metric: check.type.toLowerCase(),
					values: shiftSeries(checkSeries(check), index),
					baseline: checkSeries(check)
				}))
			: probes.map((probe, index) => ({
					key: probe.id,
					eyebrow: `${timeLabel(timeRange)} · Illustration`,
					title: `${probe.name} → ${selectedTarget.target}`,
					copy: `${selectedTarget.type} insight from ${probe.location}; ${assignedLabel(probe.name, selectedTarget.id)} path.`,
					metric: selectedTarget.type.toLowerCase(),
					values: shiftSeries(checkSeries(selectedTarget), index),
					baseline: checkSeries(selectedTarget)
				}));

	return (
		<PageStack>
			<ScreenHeader eyebrow="Measurement insight" title="Insight" copy="Pick a time window, then switch between probe-first and target-first views to compare every matching measurement graph." />

			<div className={styles.filters}>
				<SelectField label="Time" value={timeRange} onChange={event => setTimeRange(event.currentTarget.value)} options={timeOptions} />
				<SelectField label="View" value={view} onChange={event => setView(event.currentTarget.value as InsightView)} options={viewOptions} />
				<SelectField
					label={view === "probe" ? "Probe" : "Target"}
					value={view === "probe" ? selectedProbeId : selectedTargetId}
					onChange={event => {
						if (view === "probe") {
							setSelectedProbeId(event.currentTarget.value);
							return;
						}

						setSelectedTargetId(event.currentTarget.value);
					}}
					options={pickerOptions}
				/>
			</div>

			<ResponsiveGrid>
				<Panel tone="glass" eyebrow={view === "probe" ? "Probe" : "Target"} title={selectedTitle}>
					<KeyValueGrid items={selectedDetails} />
				</Panel>
				<Panel tone="glass" eyebrow={view === "probe" ? "Targets" : "Probes"} title={view === "probe" ? "Target list" : "Probe list"}>
					<div className={styles.entityList}>
						{graphCards.map(graph => (
							<Surface as="article" tone="flat" cut="sm" padding="sm" key={graph.key}>
								<span>{graph.title}</span>
								<strong>{graph.metric}</strong>
							</Surface>
						))}
					</div>
				</Panel>
			</ResponsiveGrid>

			<ResponsiveGrid>
				{graphCards.map(graph => (
					<Panel key={graph.key} tone="deep" eyebrow={graph.eyebrow} title={graph.title}>
						<BodyCopy>{graph.copy}</BodyCopy>
						<ChartPanel option={lineChartOption(graph.metric, graph.values, graph.baseline)} height="11rem" />
					</Panel>
				))}
			</ResponsiveGrid>

			<Panel tone="glass" eyebrow="Measurement table" title="Recent measurements">
				<DataTable columns={resultColumns} rows={resultRows} />
			</Panel>
		</PageStack>
	);
}
