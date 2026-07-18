import { certificatePresentation, formatBooleanResult, formatResponseBytes, formatResultDateTime, resultStatusTone } from "@/features/insight/data/httpResultData";
import type { LatestHttpResult } from "@/shared/api/types";
import { KeyValueGrid } from "@/shared/components/KeyValueGrid";
import { formatMs } from "@/shared/utils/insightFormatters";
import { Badge, BodyCopy, Panel, Spinner } from "@netstamp/ui";
import styles from "./HttpLatestResultPanel.module.css";

interface HttpLatestResultPanelProps {
	latestResult?: LatestHttpResult;
	target: string;
	nowMs: number;
	isLoading: boolean;
	isFetching: boolean;
}

export function HttpLatestResultPanel({ latestResult, target, nowMs, isLoading, isFetching }: HttpLatestResultPanelProps) {
	if (isLoading && !latestResult) {
		return (
			<Panel tone="deep" title="Latest HTTP and TLS result">
				<Spinner label="Loading latest HTTP and TLS result" layout="compact" size="lg" />
			</Panel>
		);
	}

	if (!latestResult) {
		return (
			<Panel tone="deep" title="Latest HTTP and TLS result" summary="Most recent retained measurement, independent of the selected chart range.">
				<BodyCopy>No retained HTTP result is available for this active assignment.</BodyCopy>
			</Panel>
		);
	}

	const result = latestResult.result;
	const certificate = certificatePresentation(latestResult, target, nowMs);
	const responseStatus = result.statusCode ? String(result.statusCode) : result.status;
	const resultItems = [
		{ label: "Observed", value: formatResultDateTime(result.startedAt) },
		{ label: "Completed", value: formatResultDateTime(result.finishedAt) },
		{ label: "Status", value: responseStatus },
		{ label: "Total", value: formatMs(result.durationMs) },
		{ label: "Resolved address", value: result.resolvedIp ? `${result.resolvedIp}${result.ipFamily ? ` / ${result.ipFamily}` : ""}` : "-" },
		{ label: "Final URL", value: result.finalUrl || target },
		{ label: "Response size", value: formatResponseBytes(result.responseBytes) },
		{ label: "Redirects", value: String(result.redirectCount) },
		{ label: "Body assertion", value: formatBooleanResult(result.bodyMatched, "Matched", "Did not match") },
		{ label: "Response truncated", value: formatBooleanResult(result.responseTruncated) }
	];
	const phaseItems = [
		{ label: "DNS", value: formatMs(result.dnsDurationMs) },
		{ label: "Connect", value: formatMs(result.connectDurationMs) },
		{ label: "TLS handshake", value: formatMs(result.tlsDurationMs) },
		{ label: "TTFB", value: formatMs(result.ttfbDurationMs) }
	];
	const certificateItems = [
		{ label: "Valid from", value: formatResultDateTime(result.certificateNotBefore) },
		{ label: "Valid until", value: formatResultDateTime(result.certificateNotAfter) },
		{ label: "TLS protocol", value: result.tlsVersion || "-" },
		{ label: "Cipher suite", value: result.tlsCipherSuite || "-" }
	];

	return (
		<Panel
			tone="deep"
			title="Latest HTTP and TLS result"
			summary="Most recent retained measurement, independent of the selected chart range."
			actions={
				<div className={styles.badges}>
					{isFetching ? <Badge tone="muted">Syncing</Badge> : null}
					<Badge tone={resultStatusTone(result.status)}>{result.status}</Badge>
					<Badge tone={certificate.tone}>{certificate.label}</Badge>
				</div>
			}
		>
			<div className={styles.stack}>
				<section className={styles.certificateSummary} aria-label="Certificate status">
					<div>
						<span className={styles.sectionLabel}>TLS certificate</span>
						<strong>{certificate.label}</strong>
					</div>
					<p>{certificate.detail}</p>
				</section>

				{result.errorCode || result.errorMessage ? (
					<section className={styles.errorSummary} aria-label="Latest result error">
						<strong>{result.errorCode || "request_error"}</strong>
						{result.errorMessage ? <p>{result.errorMessage}</p> : null}
					</section>
				) : null}

				<section className={styles.section}>
					<h3>Certificate and negotiation</h3>
					<KeyValueGrid className={styles.detailsGrid} items={certificateItems} />
				</section>

				<section className={styles.section}>
					<h3>Request phases</h3>
					<KeyValueGrid className={styles.phaseGrid} items={phaseItems} />
				</section>

				<section className={styles.section}>
					<h3>Latest response</h3>
					<KeyValueGrid className={styles.detailsGrid} items={resultItems} />
				</section>
			</div>
		</Panel>
	);
}
