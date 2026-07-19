import { certificatePresentation, formatBooleanResult, formatResponseBytes, formatResultDateTime, resultStatusLabel, resultStatusTone } from "@/features/insight/data/httpResultData";
import type { LatestHttpResult } from "@/shared/api/types";
import { KeyValueGrid } from "@/shared/components/KeyValueGrid";
import { formatMs } from "@/shared/utils/insightFormatters";
import { Badge, BodyCopy, Panel, Spinner } from "@netstamp/ui";
import { useTranslation } from "react-i18next";
import styles from "./HttpLatestResultPanel.module.css";

interface HttpLatestResultPanelProps {
	latestResult?: LatestHttpResult;
	target: string;
	nowMs: number;
	isLoading: boolean;
	isFetching: boolean;
}

export function HttpLatestResultPanel({ latestResult, target, nowMs, isLoading, isFetching }: HttpLatestResultPanelProps) {
	const { t } = useTranslation("insight");
	if (isLoading && !latestResult) {
		return (
			<Panel tone="deep" title={t("http.latestTitle")}>
				<Spinner label={t("http.loadingLatest")} layout="compact" size="lg" />
			</Panel>
		);
	}

	if (!latestResult) {
		return (
			<Panel tone="deep" title={t("http.latestTitle")} summary={t("http.latestSummary")}>
				<BodyCopy>{t("http.noLatest")}</BodyCopy>
			</Panel>
		);
	}

	const result = latestResult.result;
	const certificate = certificatePresentation(latestResult, target, nowMs);
	const responseStatus = result.statusCode ? String(result.statusCode) : resultStatusLabel(result.status);
	const resultItems = [
		{ label: t("http.observed"), value: formatResultDateTime(result.startedAt) },
		{ label: t("http.completed"), value: formatResultDateTime(result.finishedAt) },
		{ label: t("http.status"), value: responseStatus },
		{ label: t("http.total"), value: formatMs(result.durationMs) },
		{ label: t("http.resolvedAddress"), value: result.resolvedIp ? `${result.resolvedIp}${result.ipFamily ? ` / ${result.ipFamily}` : ""}` : "-" },
		{ label: t("http.finalUrl"), value: result.finalUrl || target },
		{ label: t("http.responseSize"), value: formatResponseBytes(result.responseBytes) },
		{ label: t("http.redirects"), value: String(result.redirectCount) },
		{ label: t("http.bodyAssertion"), value: formatBooleanResult(result.bodyMatched, t("http.matched"), t("http.notMatched")) },
		{ label: t("http.responseTruncated"), value: formatBooleanResult(result.responseTruncated, t("http.yes"), t("http.no")) }
	];
	const phaseItems = [
		{ label: "DNS", value: formatMs(result.dnsDurationMs) },
		{ label: t("http.connect"), value: formatMs(result.connectDurationMs) },
		{ label: t("http.tlsHandshake"), value: formatMs(result.tlsDurationMs) },
		{ label: "TTFB", value: formatMs(result.ttfbDurationMs) }
	];
	const certificateItems = [
		{ label: t("http.validFrom"), value: formatResultDateTime(result.certificateNotBefore) },
		{ label: t("http.validUntil"), value: formatResultDateTime(result.certificateNotAfter) },
		{ label: t("http.tlsProtocol"), value: result.tlsVersion || "-" },
		{ label: t("http.cipherSuite"), value: result.tlsCipherSuite || "-" }
	];

	return (
		<Panel
			tone="deep"
			title={t("http.latestTitle")}
			summary={t("http.latestSummary")}
			actions={
				<div className={styles.badges}>
					{isFetching ? <Badge tone="muted">{t("http.syncing")}</Badge> : null}
					<Badge tone={resultStatusTone(result.status)}>{resultStatusLabel(result.status)}</Badge>
					<Badge tone={certificate.tone}>{certificate.label}</Badge>
				</div>
			}
		>
			<div className={styles.stack}>
				<section className={styles.certificateSummary} aria-label={t("http.certificateStatus")}>
					<div>
						<span className={styles.sectionLabel}>{t("http.tlsCertificate")}</span>
						<strong>{certificate.label}</strong>
					</div>
					<p>{certificate.detail}</p>
				</section>

				{result.errorCode || result.errorMessage ? (
					<section className={styles.errorSummary} aria-label={t("http.resultError")}>
						<strong>{result.errorCode || "request_error"}</strong>
						{result.errorMessage ? <p>{result.errorMessage}</p> : null}
					</section>
				) : null}

				<section className={styles.section}>
					<h3>{t("http.certificateNegotiation")}</h3>
					<KeyValueGrid className={styles.detailsGrid} items={certificateItems} />
				</section>

				<section className={styles.section}>
					<h3>{t("http.requestPhases")}</h3>
					<KeyValueGrid className={styles.phaseGrid} items={phaseItems} />
				</section>

				<section className={styles.section}>
					<h3>{t("http.latestResponse")}</h3>
					<KeyValueGrid className={styles.detailsGrid} items={resultItems} />
				</section>
			</div>
		</Panel>
	);
}
