import type { CheckDefinition } from "@/features/checks/data/checks";
import type { Probe } from "@/features/probes/data/probes";
import type { BadgeTone } from "@netstamp/ui";

export type InsightCheckTypeFilter = "all" | "ping" | "traceroute";
export type InsightGroupBy = "check" | "probe";
export type InsightRefreshInterval = "off" | "10s" | "30s" | "1m" | "5m";
export type InsightRelativeRange = "15m" | "1h" | "6h" | "24h" | "7d" | "30d";
export type InsightTimeMode = "relative" | "absolute";
export type HopTone = Extract<BadgeTone, "success" | "warning" | "critical" | "muted">;

export interface EntityDetail {
	label: string;
	value: string;
}

export interface InsightPair {
	key: string;
	probeId: string;
	checkId: string;
	probe: Probe;
	check: CheckDefinition;
}

export interface SummaryMetric {
	label: string;
	value: string;
	detail: string;
}

export interface HopDiagnostic {
	id: string;
	hopIndex: number;
	label: string;
	address: string;
	sent: number;
	received: number;
	loss: number;
	minRtt: number | null;
	avgRtt: number | null;
	medianRtt: number | null;
	maxRtt: number | null;
	sampleCount: number;
	state: string;
	tone: HopTone;
	routerOnlyLoss: boolean;
	propagatedLoss: boolean;
	rttJump: boolean;
	noReply: boolean;
	error: string;
}

export interface TracerouteSummary {
	statusTone: BadgeTone;
	statusLabel: string;
	finalRtt: number | null;
	finalLoss: number | null;
	firstPropagatedLossHop: number | null;
	firstRttJumpHop: number | null;
	pathChangeCount: number;
}

export interface TimeWindow {
	from: number;
	to: number;
}

export interface ParsedInsightUrlState {
	checkType: InsightCheckTypeFilter;
	hasValidCheckType: boolean;
	groupBy: InsightGroupBy;
	hasValidGroupBy: boolean;
	timeMode: InsightTimeMode;
	hasValidTimeMode: boolean;
	timeRange: InsightRelativeRange;
	hasValidTimeRange: boolean;
	timeWindow: TimeWindow;
	hasValidTimeWindow: boolean;
	refresh: InsightRefreshInterval;
	hasValidRefresh: boolean;
	probeId: string;
	checkId: string;
	runStartedAt: string;
}
