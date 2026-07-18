import { certificatePresentation, formatResultDateTime, latestHttpResultKey, latestHttpResultMap, resultStatusTone } from "@/features/insight/data/httpResultData";
import type { InsightPair } from "@/features/insight/insightTypes";
import type { LatestHttpResult } from "@/shared/api/types";
import { Badge, DataTable, Panel, Spinner, type DataColumn } from "@netstamp/ui";
import { useMemo } from "react";
import styles from "./HttpCertificateInventory.module.css";

interface InventoryRow {
	pair: InsightPair;
	latest?: LatestHttpResult;
}

function inventoryColumns(nowMs: number): DataColumn<InventoryRow>[] {
	return [
		{
			key: "target",
			label: "HTTP check",
			render: row => (
				<span className={styles.identity}>
					<strong>{row.pair.check.name}</strong>
					<small>{row.pair.check.target}</small>
				</span>
			)
		},
		{
			key: "probe",
			label: "Probe",
			render: row => (
				<span className={styles.identity}>
					<strong>{row.pair.probe.name}</strong>
					<small>{row.pair.probe.location}</small>
				</span>
			)
		},
		{
			key: "result",
			label: "Result",
			render: row => <Badge tone={resultStatusTone(row.latest?.result.status)}>{row.latest?.result.status || "No result"}</Badge>
		},
		{
			key: "certificate",
			label: "Certificate",
			render: row => {
				const certificate = certificatePresentation(row.latest, row.pair.check.target, nowMs);
				return (
					<span className={styles.certificateState}>
						<Badge tone={certificate.tone}>{certificate.label}</Badge>
						<small>{certificate.detail}</small>
					</span>
				);
			}
		},
		{ key: "expires", label: "Valid until", render: row => formatResultDateTime(row.latest?.result.certificateNotAfter) },
		{
			key: "tls",
			label: "TLS",
			render: row => (
				<span className={styles.identity}>
					<strong>{row.latest?.result.tlsVersion || "-"}</strong>
					<small>{row.latest?.result.tlsCipherSuite || "No cipher reported"}</small>
				</span>
			)
		},
		{ key: "observed", label: "Last observed", render: row => formatResultDateTime(row.latest?.result.startedAt) }
	];
}

interface HttpCertificateInventoryProps {
	pairs: InsightPair[];
	latestResults: LatestHttpResult[] | undefined;
	nowMs: number;
	isLoading: boolean;
	isFetching: boolean;
}

export function HttpCertificateInventory({ pairs, latestResults, nowMs, isLoading, isFetching }: HttpCertificateInventoryProps) {
	const latestByPair = useMemo(() => latestHttpResultMap(latestResults), [latestResults]);
	const rows = useMemo(() => pairs.map(pair => ({ pair, latest: latestByPair.get(latestHttpResultKey(pair.probeId, pair.checkId)) })), [latestByPair, pairs]);
	const columns = useMemo(() => inventoryColumns(nowMs), [nowMs]);

	return (
		<Panel
			tone="deep"
			title={`TLS certificate inventory (${pairs.length})`}
			summary="Latest retained result for each selected HTTP assignment. This inventory is independent of the chart time range."
			actions={isFetching ? <Badge tone="muted">Syncing</Badge> : null}
		>
			{isLoading && !latestResults ? (
				<Spinner label="Loading TLS certificate inventory" layout="compact" size="lg" />
			) : (
				<DataTable
					columns={columns}
					rows={rows}
					density="compact"
					minWidth="72rem"
					getRowKey={row => row.pair.key}
					emptyLabel="No HTTP assignments selected"
					ariaLabel="Latest TLS certificate results"
				/>
			)}
		</Panel>
	);
}
