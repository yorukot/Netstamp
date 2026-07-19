import { formatDateTime, formatNumber } from "@/i18n/format";

export function formatTime(value: string) {
	return formatDateTime(value, { month: "short", day: "2-digit", hour: "2-digit", minute: "2-digit" });
}

export function formatShortTime(value: string) {
	return formatDateTime(value, { hour: "2-digit", minute: "2-digit" });
}

export function formatMs(value: number | null | undefined) {
	if (typeof value !== "number" || !Number.isFinite(value)) {
		return "-";
	}

	return `${formatNumber(value, value >= 100 ? { maximumFractionDigits: 0 } : { minimumFractionDigits: 1, maximumFractionDigits: 1 })}ms`;
}

export function formatPercent(value: number | null | undefined) {
	if (typeof value !== "number" || !Number.isFinite(value)) {
		return "-";
	}

	return `${formatNumber(value, value >= 10 ? { maximumFractionDigits: 0 } : { minimumFractionDigits: 1, maximumFractionDigits: 1 })}%`;
}

export function formatCount(value: number | null | undefined) {
	if (typeof value !== "number" || !Number.isFinite(value)) {
		return "-";
	}

	return formatNumber(value);
}

export function formatEpochMs(value: number | null | undefined) {
	if (typeof value !== "number" || !Number.isFinite(value)) {
		return "-";
	}

	return formatDateTime(value, { month: "short", day: "2-digit", hour: "2-digit", minute: "2-digit" });
}
