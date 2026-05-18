import { mapApiChecks } from "@/features/checks/api/checkAdapters";
import { type CheckDefinition } from "@/features/checks/data/checks";
import { dnsData, latencyData, routeDiffData } from "@/features/insight/data/series";
import { mapApiProbes } from "@/features/probes/api/probeAdapters";
import { type Probe } from "@/features/probes/data/probes";
import { projectQueries } from "@/shared/api/queries";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { BodyCopy } from "@/shared/components/BodyCopy";
import { ChartPanel } from "@/shared/components/ChartPanel";
import { KeyValueGrid } from "@/shared/components/KeyValueGrid";
import { PageStack } from "@/shared/components/PageStack";
import { ResponsiveGrid } from "@/shared/components/ResponsiveGrid";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { lineChartOption } from "@/shared/utils/chartOptions";
import { toneForStatus } from "@/shared/utils/statusTone";
import { Badge, DataTable, Panel, SelectField, Surface, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
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
	return probeName && checkId ? "available" : "unselected";
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
	const { projectRef } = useCurrentProject();
	const probesQuery = useQuery({
		...projectQueries.probes(projectRef || ""),
		enabled: Boolean(projectRef),
		select: data => mapApiProbes(data.probes)
	});
	const checksQuery = useQuery({
		...projectQueries.checks(projectRef || ""),
		enabled: Boolean(projectRef),
		select: data => mapApiChecks(data.checks, probesQuery.data)
	});
	const probes = probesQuery.data || [];
	const checks = checksQuery.data || [];
	const [timeRange, setTimeRange] = useState("24h");
	const [view, setView] = useState<InsightView>("probe");
	const [selectedProbeId, setSelectedProbeId] = useState("");
	const [selectedTargetId, setSelectedTargetId] = useState("");

	const selectedProbe = probes.find(probe => probe.id === selectedProbeId) || probes[0] || null;
	const selectedTarget = checks.find(check => check.id === selectedTargetId) || checks[0] || null;
	const pingSeriesQuery = useQuery({
		...projectQueries.pingSeries(projectRef || "", selectedProbe?.id || "", selectedTarget?.id || ""),
		enabled: Boolean(projectRef && selectedProbe && selectedTarget)
	});
	const resultRows: ResultRow[] =
		pingSeriesQuery.data?.series?.flatMap(series =>
			(series.points ?? []).flatMap(point => {
				if (!point) {
					return [];
				}

				const [time, value] = point;

				return {
					time: new Date(time).toLocaleString(),
					probe: selectedProbe?.name || "-",
					check: selectedTarget?.name || "-",
					status: "success",
					latency: `${Math.round(value)}${series.unit}`,
					loss: "-",
					metadata: series.name
				};
			})
		) ?? [];
	const selectedTitle = view === "probe" ? selectedProbe?.name || "No probe selected" : selectedTarget?.target || "No target selected";
	const selectedDetails = view === "probe" ? (selectedProbe ? detailsForProbe(selectedProbe) : []) : selectedTarget ? detailsForTarget(selectedTarget) : [];
	const pickerOptions =
		view === "probe" ? probes.map(probe => ({ value: probe.id, label: `${probe.name} · ${probe.location}` })) : checks.map(check => ({ value: check.id, label: `${check.target} · ${check.type}` }));

	const graphCards: GraphCard[] =
		view === "probe"
			? checks.map((check, index) => ({
					key: check.id,
					eyebrow: `${timeLabel(timeRange)} · Illustration`,
					title: `${selectedProbe?.name || "probe"} → ${check.target}`,
					copy: `${check.type} insight for ${assignedLabel(selectedProbe?.name || "", check.id)} probe-target measurement.`,
					metric: check.type.toLowerCase(),
					values: shiftSeries(checkSeries(check), index),
					baseline: checkSeries(check)
				}))
			: probes.map((probe, index) => ({
					key: probe.id,
					eyebrow: `${timeLabel(timeRange)} · Illustration`,
					title: `${probe.name} → ${selectedTarget?.target || "target"}`,
					copy: `${selectedTarget?.type || "Ping"} insight from ${probe.location}; ${assignedLabel(probe.name, selectedTarget?.id || "")} path.`,
					metric: (selectedTarget?.type || "Ping").toLowerCase(),
					values: selectedTarget ? shiftSeries(checkSeries(selectedTarget), index) : [],
					baseline: selectedTarget ? checkSeries(selectedTarget) : []
				}));

	return (
		<PageStack>
			<ScreenHeader eyebrow="Measurement insight" title="Insight" copy="Pick a time window, then switch between probe-first and target-first views to compare every matching measurement graph." />

			<div className={styles.filters}>
				<SelectField label="Time" value={timeRange} onChange={event => setTimeRange(event.currentTarget.value)} options={timeOptions} />
				<SelectField label="View" value={view} onChange={event => setView(event.currentTarget.value as InsightView)} options={viewOptions} />
				<SelectField
					label={view === "probe" ? "Probe" : "Target"}
					value={view === "probe" ? selectedProbe?.id || "" : selectedTarget?.id || ""}
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
