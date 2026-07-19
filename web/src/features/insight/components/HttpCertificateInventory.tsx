import { certificatePresentation, formatResultDateTime, latestHttpResultKey, latestHttpResultMap, resultStatusLabel, resultStatusTone } from "@/features/insight/data/httpResultData";
import type { InsightPair } from "@/features/insight/insightTypes";
import type { LatestHttpResult } from "@/shared/api/types";
import { Badge, DataTable, Panel, Spinner, type DataColumn } from "@netstamp/ui";
import type { TFunction } from "i18next";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import styles from "./HttpCertificateInventory.module.css";

interface InventoryRow {
	pair: InsightPair;
	latest?: LatestHttpResult;
}

function inventoryColumns(nowMs: number, t: TFunction<"insight">): DataColumn<InventoryRow>[] {
	return [
		{
			key: "target",
			label: t("http.httpCheck"),
			render: row => (
				<span className={styles.identity}>
					<strong>{row.pair.check.name}</strong>
					<small>{row.pair.check.target}</small>
				</span>
			)
		},
		{
			key: "probe",
			label: t("http.probe"),
			render: row => (
				<span className={styles.identity}>
					<strong>{row.pair.probe.name}</strong>
					<small>{row.pair.probe.location}</small>
				</span>
			)
		},
		{
			key: "result",
			label: t("http.result"),
			render: row => <Badge tone={resultStatusTone(row.latest?.result.status)}>{resultStatusLabel(row.latest?.result.status)}</Badge>
		},
		{
			key: "certificate",
			label: t("http.certificate"),
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
		{ key: "expires", label: t("http.validUntil"), render: row => formatResultDateTime(row.latest?.result.certificateNotAfter) },
		{
			key: "tls",
			label: t("http.tls"),
			render: row => (
				<span className={styles.identity}>
					<strong>{row.latest?.result.tlsVersion || "-"}</strong>
					<small>{row.latest?.result.tlsCipherSuite || t("http.noCipher")}</small>
				</span>
			)
		},
		{ key: "observed", label: t("http.lastObserved"), render: row => formatResultDateTime(row.latest?.result.startedAt) }
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
	const { t } = useTranslation("insight");
	const latestByPair = useMemo(() => latestHttpResultMap(latestResults), [latestResults]);
	const rows = useMemo(() => pairs.map(pair => ({ pair, latest: latestByPair.get(latestHttpResultKey(pair.probeId, pair.checkId)) })), [latestByPair, pairs]);
	const columns = useMemo(() => inventoryColumns(nowMs, t), [nowMs, t]);

	return (
		<Panel tone="deep" title={t("http.inventory", { count: pairs.length })} summary={t("http.inventorySummary")} actions={isFetching ? <Badge tone="muted">{t("http.syncing")}</Badge> : null}>
			{isLoading && !latestResults ? (
				<Spinner label={t("http.loadingInventory")} layout="compact" size="lg" />
			) : (
				<DataTable columns={columns} rows={rows} density="compact" minWidth="72rem" getRowKey={row => row.pair.key} emptyLabel={t("http.noAssignments")} ariaLabel={t("http.inventoryAria")} />
			)}
		</Panel>
	);
}
