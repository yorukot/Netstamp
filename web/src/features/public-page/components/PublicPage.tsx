import { publicPageQueries } from "@/shared/api/queries";
import type { ApiPublicPageFolder, ApiPublicPingPair } from "@/shared/api/types";
import { classNames } from "@/shared/utils/classNames";
import { formatCount, formatEpochMs } from "@/shared/utils/insightFormatters";
import { pingChartBuckets, pingSampleDensity, pingSummaryMetrics } from "@/shared/utils/pingInsightData";
import { ChartPanel } from "@/shared/visualizations/ChartPanel";
import { pingInsightChartOption } from "@/shared/visualizations/chartOptions";
import { Badge, Button, DataTable, Panel, SelectField, type BadgeTone, type DataColumn } from "@netstamp/ui";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { Helmet } from "react-helmet-async";
import { Link, useParams, useSearchParams } from "react-router-dom";
import styles from "./PublicPage.module.css";

type PublicRelativeRange = "15m" | "1h" | "6h" | "24h" | "7d" | "30d";
type TimeSelectValue = PublicRelativeRange | "custom";

interface PairRow {
	id: string;
	folder: string;
	check: string;
	checkDescription: string;
	probe: string;
	probeLocation: string;
	status: ApiPublicPingPair["probeStatus"];
	interval: string;
	source: ApiPublicPingPair;
}

const timeOptions: Array<{ value: PublicRelativeRange; label: string }> = [
	{ value: "15m", label: "Last 15 minutes" },
	{ value: "1h", label: "Last 1 hour" },
	{ value: "6h", label: "Last 6 hours" },
	{ value: "24h", label: "Last 24 hours" },
	{ value: "7d", label: "Last 7 days" },
	{ value: "30d", label: "Last 30 days" }
];

const timeRangeDurations: Record<PublicRelativeRange, number> = {
	"15m": 15 * 60 * 1000,
	"1h": 60 * 60 * 1000,
	"6h": 6 * 60 * 60 * 1000,
	"24h": 24 * 60 * 60 * 1000,
	"7d": 7 * 24 * 60 * 60 * 1000,
	"30d": 30 * 24 * 60 * 60 * 1000
};

const EMPTY_FOLDERS: ApiPublicPageFolder[] = [];
const EMPTY_PAIRS: ApiPublicPingPair[] = [];

function pairKey(pair: ApiPublicPingPair) {
	return `${pair.checkId}:${pair.probeId}`;
}

function folderLabel(folder: ApiPublicPageFolder, folders: ApiPublicPageFolder[]) {
	const names: string[] = [folder.name];
	let current = folder;
	const guard = new Set<string>([folder.id]);

	while (current.parentId) {
		const parent = folders.find(candidate => candidate.id === current.parentId);
		if (!parent || guard.has(parent.id)) {
			break;
		}

		names.unshift(parent.name);
		guard.add(parent.id);
		current = parent;
	}

	return names.join(" / ");
}

function isPublicRelativeRange(value: string | null): value is PublicRelativeRange {
	return value === "15m" || value === "1h" || value === "6h" || value === "24h" || value === "7d" || value === "30d";
}

function parseEpochMs(value: string | null) {
	if (!value) {
		return null;
	}

	const parsed = Number(value);

	return Number.isFinite(parsed) && parsed > 0 ? Math.trunc(parsed) : null;
}

function timeWindowForRange(value: PublicRelativeRange, now = Date.now()) {
	const to = now;
	const from = to - timeRangeDurations[value];

	return { from, to };
}

function formatTimeRange(from: number, to: number) {
	const options: Intl.DateTimeFormatOptions = { month: "short", day: "2-digit", hour: "2-digit", minute: "2-digit" };

	return `${new Date(from).toLocaleString(undefined, options)} - ${new Date(to).toLocaleString(undefined, options)}`;
}

function statusTone(status: ApiPublicPingPair["probeStatus"]): BadgeTone {
	return status === "online" ? "success" : "warning";
}

function countPairsByFolder(pairs: ApiPublicPingPair[]) {
	const counts = new Map<string, number>();

	for (const pair of pairs) {
		counts.set(pair.folderId, (counts.get(pair.folderId) ?? 0) + 1);
	}

	return counts;
}

function uniqueCheckOptions(pairs: ApiPublicPingPair[]) {
	const checks = new Map<string, string>();

	for (const pair of pairs) {
		checks.set(pair.checkId, pair.checkName);
	}

	return [...checks.entries()].map(([value, label]) => ({ value, label }));
}

function probeOptionsForCheck(pairs: ApiPublicPingPair[], checkId: string) {
	return pairs.filter(pair => pair.checkId === checkId).map(pair => ({ value: pair.probeId, label: pair.probeLocationName ? `${pair.probeName} · ${pair.probeLocationName}` : pair.probeName }));
}

export function PublicPage() {
	const { slug = "" } = useParams<{ slug: string }>();
	const [searchParams, setSearchParams] = useSearchParams();
	const [nowMs, setNowMs] = useState(() => Date.now());
	const publicPageQuery = useQuery({
		...publicPageQueries.detail(slug),
		enabled: Boolean(slug)
	});
	const page = publicPageQuery.data?.publicPage;
	const folders = page?.folders ?? EMPTY_FOLDERS;
	const pairs = page?.pairs ?? EMPTY_PAIRS;
	const folderCounts = useMemo(() => countPairsByFolder(pairs), [pairs]);
	const requestedFolderId = searchParams.get("folderId") || "";
	const activeFolder = folders.find(folder => folder.id === requestedFolderId) ?? null;
	const activeFolderId = activeFolder?.id ?? "";
	const visiblePairs = activeFolderId ? pairs.filter(pair => pair.folderId === activeFolderId) : pairs;
	const requestedCheckId = searchParams.get("checkId") || "";
	const requestedProbeId = searchParams.get("probeId") || "";
	const activePair = visiblePairs.find(pair => pair.checkId === requestedCheckId && pair.probeId === requestedProbeId) ?? visiblePairs[0] ?? null;
	const selectedCheckId = activePair?.checkId ?? "";
	const selectedProbeId = activePair?.probeId ?? "";
	const rawFrom = parseEpochMs(searchParams.get("from"));
	const rawTo = parseEpochMs(searchParams.get("to"));
	const hasAbsoluteWindow = rawFrom !== null && rawTo !== null && rawFrom < rawTo;
	const requestedRange = searchParams.get("range");
	const relativeRange: PublicRelativeRange = isPublicRelativeRange(requestedRange) ? requestedRange : "24h";
	const timeWindow = hasAbsoluteWindow ? { from: rawFrom, to: rawTo } : timeWindowForRange(relativeRange, nowMs);
	const timeLabel = hasAbsoluteWindow ? formatTimeRange(timeWindow.from, timeWindow.to) : (timeOptions.find(option => option.value === relativeRange)?.label ?? "Last 24 hours");
	const timeSelectValue: TimeSelectValue = hasAbsoluteWindow ? "custom" : relativeRange;
	const timeSelectOptions = hasAbsoluteWindow ? [...timeOptions, { value: "custom", label: "Custom window" }] : timeOptions;
	const insightQuery = useQuery({
		...publicPageQueries.pingInsight(slug, selectedProbeId, selectedCheckId, { from: timeWindow.from, to: timeWindow.to }),
		enabled: Boolean(slug && activePair)
	});
	const buckets = pingChartBuckets(insightQuery.data);
	const density = pingSampleDensity(insightQuery.data);
	const metrics = pingSummaryMetrics(insightQuery.data);
	const hasChartData = buckets.length > 0 || density.length > 0;
	const queryWindow = insightQuery.data?.query ? { from: insightQuery.data.query.from, to: insightQuery.data.query.to } : timeWindow;
	const checkOptions = uniqueCheckOptions(visiblePairs);
	const probeOptions = selectedCheckId ? probeOptionsForCheck(visiblePairs, selectedCheckId) : [];
	const rows: PairRow[] = visiblePairs.map(pair => ({
		id: pairKey(pair),
		folder: folderLabel(folders.find(folder => folder.id === pair.folderId) ?? { id: pair.folderId, name: "Folder", sortOrder: 0, createdAt: "", updatedAt: "" }, folders),
		check: pair.checkName,
		checkDescription: pair.checkDescription ?? "",
		probe: pair.probeName,
		probeLocation: pair.probeLocationName ?? "unlabeled location",
		status: pair.probeStatus,
		interval: `${pair.checkIntervalSeconds}s`,
		source: pair
	}));
	const rowColumns: DataColumn<PairRow>[] = [
		{
			key: "check",
			label: "Check",
			render: row => (
				<span className={styles.stackedCell}>
					<strong>{row.check}</strong>
					<span>{row.checkDescription || row.folder}</span>
				</span>
			)
		},
		{
			key: "probe",
			label: "Probe",
			render: row => (
				<span className={styles.stackedCell}>
					<strong>{row.probe}</strong>
					<span>{row.probeLocation}</span>
				</span>
			)
		},
		{ key: "folder", label: "Folder" },
		{ key: "status", label: "Probe", render: row => <Badge tone={statusTone(row.status)}>{row.status}</Badge> },
		{ key: "interval", label: "Interval" }
	];

	function updateParams(updater: (params: URLSearchParams) => void) {
		setSearchParams(current => {
			const next = new URLSearchParams(current);
			updater(next);
			return next;
		});
	}

	function selectPair(pair: ApiPublicPingPair, folderId = activeFolderId) {
		updateParams(params => {
			if (folderId) {
				params.set("folderId", folderId);
			} else {
				params.delete("folderId");
			}

			params.set("checkId", pair.checkId);
			params.set("probeId", pair.probeId);
		});
	}

	function selectFolder(folderId: string) {
		const nextPairs = folderId ? pairs.filter(pair => pair.folderId === folderId) : pairs;
		const nextPair = nextPairs[0];

		updateParams(params => {
			if (folderId) {
				params.set("folderId", folderId);
			} else {
				params.delete("folderId");
			}

			if (nextPair) {
				params.set("checkId", nextPair.checkId);
				params.set("probeId", nextPair.probeId);
			} else {
				params.delete("checkId");
				params.delete("probeId");
			}
		});
	}

	function selectCheck(checkId: string) {
		const nextPair = visiblePairs.find(pair => pair.checkId === checkId);

		if (nextPair) {
			selectPair(nextPair);
		}
	}

	function selectProbe(probeId: string) {
		const nextPair = visiblePairs.find(pair => pair.checkId === selectedCheckId && pair.probeId === probeId);

		if (nextPair) {
			selectPair(nextPair);
		}
	}

	function selectTimeWindow(value: string) {
		if (!isPublicRelativeRange(value)) {
			return;
		}

		setNowMs(Date.now());
		updateParams(params => {
			params.set("range", value);
			params.delete("from");
			params.delete("to");
		});
	}

	function selectChartTimeWindow(range: { from: number; to: number }) {
		updateParams(params => {
			params.set("from", String(Math.trunc(range.from)));
			params.set("to", String(Math.trunc(range.to)));
			params.delete("range");
		});
	}

	function refresh() {
		if (hasAbsoluteWindow) {
			void insightQuery.refetch();
			return;
		}

		setNowMs(Date.now());
	}

	if (publicPageQuery.isLoading) {
		return (
			<div className={classNames("ns-grid-shell", styles.publicShell)}>
				<Panel tone="deep" eyebrow="Public page" title="Loading Ping insight">
					<div className={styles.emptyState}>Loading public page configuration.</div>
				</Panel>
			</div>
		);
	}

	if (!page) {
		return (
			<div className={classNames("ns-grid-shell", styles.publicShell)}>
				<Helmet>
					<title>Public page not found | Netstamp</title>
				</Helmet>
				<Panel tone="deep" eyebrow="Public page" title="Page unavailable">
					<div className={styles.emptyState}>This public page is disabled or no longer exists.</div>
					<Button asChild variant="outline">
						<Link to="/login">Sign in</Link>
					</Button>
				</Panel>
			</div>
		);
	}

	return (
		<div className={classNames("ns-grid-shell", styles.publicShell)}>
			<Helmet>
				<title>{page.title} | Netstamp</title>
			</Helmet>
			<header className={styles.publicHeader}>
				<Link className={styles.brandLink} to="/login">
					<span>Netstamp</span>
					<strong>/s/{page.slug}</strong>
				</Link>
				<div className={styles.headerActions}>
					<Badge tone={page.enabled ? "success" : "warning"}>{page.enabled ? "Public" : "Paused"}</Badge>
					<Button asChild variant="outline" size="sm">
						<Link to="/login">Sign in</Link>
					</Button>
				</div>
			</header>

			<main className={styles.publicMain}>
				<section className={styles.titleBand}>
					<div>
						<p>Public Ping insight</p>
						<h1>{page.title}</h1>
						{page.description ? <span>{page.description}</span> : null}
					</div>
					<div className={styles.pageStats}>
						<span>{folders.length} folders</span>
						<span>{pairs.length} probe pairs</span>
						<span>updated {new Date(page.updatedAt).toLocaleString()}</span>
					</div>
				</section>

				<div className={styles.contentGrid}>
					<aside className={styles.folderRail} aria-label="Public page folders">
						<button type="button" className={!activeFolderId ? styles.folderTabActive : styles.folderTab} onClick={() => selectFolder("")}>
							<span>All published</span>
							<small>{pairs.length} pairs</small>
						</button>
						{folders.map(folder => (
							<button key={folder.id} type="button" className={folder.id === activeFolderId ? styles.folderTabActive : styles.folderTab} onClick={() => selectFolder(folder.id)}>
								<span>{folderLabel(folder, folders)}</span>
								<small>{folderCounts.get(folder.id) ?? 0} pairs</small>
							</button>
						))}
					</aside>

					<section className={styles.insightStack}>
						<Panel
							tone="glass"
							eyebrow="Selector"
							title={activePair ? `${activePair.checkName} from ${activePair.probeName}` : "No Ping check selected"}
							actions={
								<div className={styles.controlActions}>
									<SelectField label="Check" value={selectedCheckId} options={checkOptions} disabled={!checkOptions.length} onChange={event => selectCheck(event.currentTarget.value)} />
									<SelectField label="Probe" value={selectedProbeId} options={probeOptions} disabled={!probeOptions.length} onChange={event => selectProbe(event.currentTarget.value)} />
									<SelectField label="Window" value={timeSelectValue} options={timeSelectOptions} onChange={event => selectTimeWindow(event.currentTarget.value)} />
									<Button variant="outline" size="sm" disabled={!activePair || insightQuery.isFetching} onClick={refresh}>
										{insightQuery.isFetching ? "Syncing" : "Refresh"}
									</Button>
								</div>
							}
						>
							<DataTable
								columns={rowColumns}
								rows={rows}
								density="compact"
								minWidth="48rem"
								maxHeight="18rem"
								getRowKey={row => row.id}
								selectedKey={activePair ? pairKey(activePair) : undefined}
								onRowClick={row => selectPair(row.source, row.source.folderId)}
								emptyLabel="No Ping checks are published in this folder."
							/>
						</Panel>

						<div className={styles.summaryGrid}>
							{metrics.map(metric => (
								<div className={classNames("ns-cut-frame", styles.summaryCell)} key={metric.label}>
									<span>{metric.label}</span>
									<strong>{metric.value}</strong>
									<small>{metric.detail}</small>
								</div>
							))}
						</div>

						<Panel tone="deep" eyebrow={`${timeLabel} · ${insightQuery.data?.query.resolution || "pending"}`} title="Ping latency and loss">
							<div className={styles.chartMeta}>
								<span>{insightQuery.isFetching ? "syncing result buckets" : `${formatCount(insightQuery.data?.query.totalPoints)} results`}</span>
								<span>latest {formatEpochMs(insightQuery.data?.summary.latestStartedAtMs)}</span>
								{activePair ? <span>{activePair.probeLocationName || activePair.probeName}</span> : null}
							</div>
							{activePair && hasChartData ? (
								<ChartPanel option={pingInsightChartOption(buckets, density)} height="27rem" onTimeRangeSelect={selectChartTimeWindow} timeRangeBounds={queryWindow} />
							) : (
								<div className={styles.emptyState}>
									{activePair ? "No Ping results were recorded in the selected time range." : "Select a published Ping check to render the public insight chart."}
								</div>
							)}
						</Panel>
					</section>
				</div>
			</main>
		</div>
	);
}
