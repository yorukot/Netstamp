import type { InsightPair } from "@/features/insight/insightTypes";
import { i18n } from "@/i18n";
import { formatDateTime, formatNumber } from "@/i18n/format";
import type { LatestHttpResult } from "@/shared/api/types";
import type { BadgeTone } from "@netstamp/ui";

const dayMs = 24 * 60 * 60 * 1000;
const insightT = i18n.getFixedT(null, "insight") as (key: string, options?: Record<string, unknown>) => string;

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

export function resultStatusLabel(status: LatestHttpResult["result"]["status"] | undefined) {
	return status ? insightT(`http.results.${status}`) : insightT("http.noResult");
}

export function certificatePresentation(latest: LatestHttpResult | undefined, target: string, nowMs = Date.now()): CertificatePresentation {
	if (!latest) {
		return { label: insightT("http.noResult"), detail: insightT("http.noRetainedResult"), tone: "muted", daysRemaining: null };
	}
	const result = latest.result;
	if (result.errorCode === "tls_verification_failed") {
		return { label: insightT("http.verificationFailed"), detail: result.errorMessage || insightT("http.verificationFailedDetail"), tone: "critical", daysRemaining: null };
	}
	if (!result.certificateNotAfter) {
		if (!target.trim().toLowerCase().startsWith("https://")) {
			return { label: insightT("http.notApplicable"), detail: insightT("http.plainHttp"), tone: "muted", daysRemaining: null };
		}
		return {
			label: insightT("http.notReported"),
			detail: result.status === "successful" ? insightT("http.certificateMissing") : insightT("http.tlsUnavailable"),
			tone: result.status === "successful" ? "warning" : "critical",
			daysRemaining: null
		};
	}

	const expiresAt = Date.parse(result.certificateNotAfter);
	if (!Number.isFinite(expiresAt)) {
		return { label: insightT("http.invalidDate"), detail: insightT("http.invalidDateDetail"), tone: "warning", daysRemaining: null };
	}
	const daysRemaining = Math.floor((expiresAt - nowMs) / dayMs);
	if (daysRemaining < 0) {
		const daysExpired = Math.abs(daysRemaining);
		return { label: insightT("http.expired"), detail: insightT("http.pastExpiry", { count: daysExpired }), tone: "critical", daysRemaining };
	}
	if (daysRemaining <= 7) {
		return {
			label: insightT("http.expiresSoon"),
			detail: daysRemaining === 0 ? insightT("http.expiresToday") : insightT("http.daysRemaining", { count: daysRemaining }),
			tone: "critical",
			daysRemaining
		};
	}
	if (daysRemaining <= 30) {
		return { label: insightT("http.expiring"), detail: insightT("http.daysRemaining", { count: daysRemaining }), tone: "warning", daysRemaining };
	}
	return { label: insightT("http.valid"), detail: insightT("http.daysRemaining", { count: daysRemaining }), tone: "success", daysRemaining };
}

export function formatResultDateTime(value: string | undefined) {
	if (!value) {
		return "-";
	}
	const timestamp = Date.parse(value);
	if (!Number.isFinite(timestamp)) {
		return value;
	}
	return formatDateTime(timestamp, {
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
	return formatNumber(value, { notation: value >= 1000 ? "compact" : "standard", maximumFractionDigits: 1 }) + " B";
}

export function formatBooleanResult(value: boolean | undefined, positive: string, negative: string) {
	if (value === undefined) {
		return "-";
	}
	return value ? positive : negative;
}
