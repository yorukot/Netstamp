import type { ApiPublicStatusElement, ApiPublicStatusPublicElement, ApiSeries, PublicStatusState } from "@/shared/api/types";
import type { ChartOption } from "@/shared/visualizations/chartOptions";
import { chartAxisLabel, chartTheme, chartTooltipTextStyle } from "@/shared/visualizations/chartTheme";
import type { BadgeTone } from "@netstamp/ui";

export interface ElementTreeNode extends ApiPublicStatusElement {
	children: ElementTreeNode[];
}

export function statusLabel(status: PublicStatusState | string | undefined) {
	switch (status) {
		case "operational":
			return "Operational";
		case "degraded":
			return "Degraded";
		case "down":
			return "Down";
		default:
			return "Unknown";
	}
}

export function statusTone(status: PublicStatusState | string | undefined): BadgeTone {
	switch (status) {
		case "operational":
			return "success";
		case "degraded":
			return "warning";
		case "down":
			return "critical";
		default:
			return "neutral";
	}
}

export function severityTone(severity: string | undefined): BadgeTone {
	switch (severity) {
		case "critical":
			return "critical";
		case "warning":
			return "warning";
		case "info":
			return "accent";
		default:
			return "neutral";
	}
}

export function checkTypeLabel(type: string | undefined) {
	switch (type) {
		case "http":
			return "HTTP";
		case "tcp":
			return "TCP";
		case "traceroute":
			return "Traceroute";
		case "ping":
			return "Ping";
		default:
			return "Check";
	}
}

export function chartModeLabel(value: string | undefined) {
	switch (value) {
		case "compact":
			return "Compact";
		case "inherit":
			return "Inherit";
		default:
			return "Off";
	}
}

export function chartRangeLabel(value: string | undefined) {
	switch (value) {
		case "30d":
			return "30 days";
		case "7d":
			return "7 days";
		default:
			return "24 hours";
	}
}

export function formatDateTime(value?: string | null) {
	if (!value) {
		return "-";
	}
	const date = new Date(value);

	if (Number.isNaN(date.getTime())) {
		return "-";
	}

	return date.toLocaleString([], {
		month: "short",
		day: "2-digit",
		hour: "2-digit",
		minute: "2-digit"
	});
}

export function formatMetric(value: number | undefined, unit: string) {
	if (typeof value !== "number" || !Number.isFinite(value)) {
		return "-";
	}

	return `${value.toFixed(value >= 100 ? 0 : 1)}${unit}`;
}

export function publicStatusPath(slug: string) {
	return `/status/${encodeURIComponent(slug)}`;
}

export function buildElementTree(elements: ApiPublicStatusElement[]): ElementTreeNode[] {
	const byID = new Map<string, ElementTreeNode>();
	const roots: ElementTreeNode[] = [];

	for (const element of elements) {
		byID.set(element.id, { ...element, children: [] });
	}

	for (const node of byID.values()) {
		if (!node.parentElementId) {
			roots.push(node);
			continue;
		}

		const parent = byID.get(node.parentElementId);
		if (parent) {
			parent.children.push(node);
		} else {
			roots.push(node);
		}
	}

	sortElementTree(roots);
	return roots;
}

export function rootFolderOptions(elements: ApiPublicStatusElement[]) {
	return elements
		.filter(element => element.kind === "folder" && !element.parentElementId)
		.sort(compareElements)
		.map(element => ({ value: element.id, label: element.title || "Untitled folder" }));
}

function sortElementTree(nodes: ElementTreeNode[]) {
	nodes.sort(compareElements);
	for (const node of nodes) {
		sortElementTree(node.children);
	}
}

function compareElements(a: Pick<ApiPublicStatusElement, "sortOrder" | "createdAt" | "id">, b: Pick<ApiPublicStatusElement, "sortOrder" | "createdAt" | "id">) {
	return a.sortOrder - b.sortOrder || Date.parse(a.createdAt) - Date.parse(b.createdAt) || a.id.localeCompare(b.id);
}

function seriesDisplayName(series: ApiSeries) {
	const checkName = series.labels.checkName;
	const probeName = series.labels.probeName;
	if (checkName && probeName) {
		return `${checkName} / ${probeName}`;
	}
	if (probeName) {
		return `${series.name} / ${probeName}`;
	}
	return checkName || series.name;
}

function seriesPoints(series: ApiSeries) {
	return series.points.filter(([timestamp, value]) => Number.isFinite(timestamp) && Number.isFinite(value));
}

function tooltipTimestamp(value: unknown) {
	if (!Array.isArray(value) || typeof value[0] !== "number") {
		return "";
	}
	return formatDateTime(new Date(value[0]).toISOString());
}

export function publicStatusChartOption(element: ApiPublicStatusPublicElement): ChartOption {
	const series = element.chart?.series ?? [];
	const theme = chartTheme();
	const splitLine = { lineStyle: { color: theme.splitLine } };

	return {
		backgroundColor: "transparent",
		color: theme.seriesPalette,
		tooltip: {
			trigger: "axis",
			backgroundColor: theme.tooltipBackground,
			borderColor: theme.tooltipBorder,
			textStyle: chartTooltipTextStyle(),
			formatter: (params: unknown) => {
				const items = (Array.isArray(params) ? params : [params]) as Array<{ marker?: string; seriesName?: string; value?: unknown }>;
				const lines = [tooltipTimestamp(items[0]?.value)].filter(Boolean);

				for (const item of items) {
					if (!Array.isArray(item.value) || typeof item.value[1] !== "number") {
						continue;
					}
					lines.push(`${item.marker || ""}${item.seriesName}: ${item.value[1].toFixed(1)}`);
				}

				return lines.join("<br />");
			}
		},
		grid: { top: 12, right: 14, bottom: 24, left: 42 },
		xAxis: { type: "time", axisLabel: chartAxisLabel(), axisTick: { show: false }, axisLine: { lineStyle: { color: theme.axisLine } } },
		yAxis: { type: "value", axisLabel: chartAxisLabel(), splitLine, axisTick: { show: false }, axisLine: { show: false } },
		series: series.map(item => ({
			name: seriesDisplayName(item),
			type: "line",
			data: seriesPoints(item),
			smooth: true,
			showSymbol: false,
			lineStyle: { width: 1.8 }
		}))
	};
}
