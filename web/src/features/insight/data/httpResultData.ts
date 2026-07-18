import type { InsightPair } from "@/features/insight/insightTypes";
import type { LatestHttpResult } from "@/shared/api/types";
import type { BadgeTone } from "@netstamp/ui";

const dayMs = 24 * 60 * 60 * 1000;

export interface CertificatePresentation {
	label: string;
	detail: string;
	tone: BadgeTone;
	daysRemaining: number | null;
}

export function latestHttpResultKey(probeId: string, checkId: string) {
	return `${probeId}:${checkId}`;
}

export function latestHttpResultMap(results: LatestHttpResult[] | undefined) {
	return new Map((results ?? []).map(result => [latestHttpResultKey(result.probeId, result.checkId), result]));
}

export function latestHTTPResultForPair(results: Map<string, LatestHttpResult>, pair: InsightPair | null) {
	if (!pair) {
		return undefined;
	}
	return results.get(latestHttpResultKey(pair.probeId, pair.checkId));
}

export function resultStatusTone(status: LatestHttpResult["result"]["status"] | undefined): BadgeTone {
	switch (status) {
		case "successful":
			return "success";
		case "timeout":
			return "warning";
		case "error":
			return "critical";
		default:
			return "muted";
	}
}

export function certificatePresentation(latest: LatestHttpResult | undefined, target: string, nowMs = Date.now()): CertificatePresentation {
	if (!latest) {
		return { label: "No result", detail: "No retained HTTP result", tone: "muted", daysRemaining: null };
	}
	const result = latest.result;
	if (result.errorCode === "tls_verification_failed") {
		return { label: "Verification failed", detail: result.errorMessage || "TLS certificate verification failed", tone: "critical", daysRemaining: null };
	}
	if (!result.certificateNotAfter) {
		if (!target.trim().toLowerCase().startsWith("https://")) {
			return { label: "Not applicable", detail: "Plain HTTP target", tone: "muted", daysRemaining: null };
		}
		return {
			label: "Not reported",
			detail: result.status === "successful" ? "The latest HTTPS result did not include a certificate" : "TLS metadata was unavailable for the failed result",
			tone: result.status === "successful" ? "warning" : "critical",
			daysRemaining: null
		};
	}

	const expiresAt = Date.parse(result.certificateNotAfter);
	if (!Number.isFinite(expiresAt)) {
		return { label: "Invalid date", detail: "The certificate expiry timestamp could not be parsed", tone: "warning", daysRemaining: null };
	}
	const daysRemaining = Math.floor((expiresAt - nowMs) / dayMs);
	if (daysRemaining < 0) {
		const daysExpired = Math.abs(daysRemaining);
		return { label: "Expired", detail: `${daysExpired}d past expiry`, tone: "critical", daysRemaining };
	}
	if (daysRemaining <= 7) {
		return { label: "Expires soon", detail: daysRemaining === 0 ? "Expires today" : `${daysRemaining}d remaining`, tone: "critical", daysRemaining };
	}
	if (daysRemaining <= 30) {
		return { label: "Expiring", detail: `${daysRemaining}d remaining`, tone: "warning", daysRemaining };
	}
	return { label: "Valid", detail: `${daysRemaining}d remaining`, tone: "success", daysRemaining };
}

export function formatResultDateTime(value: string | undefined) {
	if (!value) {
		return "-";
	}
	const timestamp = Date.parse(value);
	if (!Number.isFinite(timestamp)) {
		return value;
	}
	return new Date(timestamp).toLocaleString(undefined, {
		year: "numeric",
		month: "short",
		day: "2-digit",
		hour: "2-digit",
		minute: "2-digit",
		second: "2-digit",
		timeZoneName: "short"
	});
}

export function formatResponseBytes(value: number | undefined) {
	if (typeof value !== "number" || !Number.isFinite(value)) {
		return "-";
	}
	return new Intl.NumberFormat(undefined, { notation: value >= 1000 ? "compact" : "standard", maximumFractionDigits: 1 }).format(value) + " B";
}

export function formatBooleanResult(value: boolean | undefined, positive = "Yes", negative = "No") {
	if (value === undefined) {
		return "-";
	}
	return value ? positive : negative;
}
