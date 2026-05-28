export type RelativeTimeRange = "15m" | "1h" | "6h" | "24h" | "7d" | "30d";

export interface TimeWindow {
	from: number;
	to: number;
}

export const relativeTimeOptions: Array<{ value: RelativeTimeRange; label: string }> = [
	{ value: "15m", label: "Last 15 minutes" },
	{ value: "1h", label: "Last 1 hour" },
	{ value: "6h", label: "Last 6 hours" },
	{ value: "24h", label: "Last 24 hours" },
	{ value: "7d", label: "Last 7 days" },
	{ value: "30d", label: "Last 30 days" }
];

export const relativeTimeRangeDurations: Record<RelativeTimeRange, number> = {
	"15m": 15 * 60 * 1000,
	"1h": 60 * 60 * 1000,
	"6h": 6 * 60 * 60 * 1000,
	"24h": 24 * 60 * 60 * 1000,
	"7d": 7 * 24 * 60 * 60 * 1000,
	"30d": 30 * 24 * 60 * 60 * 1000
};

export function isRelativeTimeRange(value: string | null): value is RelativeTimeRange {
	return value === "15m" || value === "1h" || value === "6h" || value === "24h" || value === "7d" || value === "30d";
}

export function parseEpochMs(value: string | null) {
	if (!value) {
		return null;
	}

	const parsed = Number(value);

	return Number.isFinite(parsed) && parsed > 0 ? Math.trunc(parsed) : null;
}

export function timeWindowForRelativeRange(value: RelativeTimeRange, now = Date.now(), fallback: RelativeTimeRange = "24h"): TimeWindow {
	const to = now;
	const from = to - (relativeTimeRangeDurations[value] ?? relativeTimeRangeDurations[fallback]);

	return { from, to };
}

export function relativeRangeForTimeWindow(timeWindow: TimeWindow) {
	const duration = timeWindow.to - timeWindow.from;
	const option = relativeTimeOptions.find(candidate => relativeTimeRangeDurations[candidate.value] === duration);

	return option?.value ?? null;
}

export function formatAbsoluteTime(value: number) {
	return new Date(value).toLocaleString(undefined, { month: "short", day: "2-digit", hour: "2-digit", minute: "2-digit" });
}

export function formatTimeWindow(from: number, to: number, separator = " - ") {
	return `${formatAbsoluteTime(from)}${separator}${formatAbsoluteTime(to)}`;
}
