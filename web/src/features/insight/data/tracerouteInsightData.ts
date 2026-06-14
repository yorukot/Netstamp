import type { HopDiagnostic, HopTone, TracerouteSummary } from "@/features/insight/insightTypes";
import type { TracerouteHop, TracerouteInsightResponse, TracerouteResult } from "@/shared/api/types";
import { formatMs, formatPercent, formatShortTime, formatTime } from "@/shared/utils/insightFormatters";
import type { RunTimelinePoint } from "@/shared/visualizations/RunTimeline";

function orderedHops(run: TracerouteResult | null | undefined) {
	return [...(run?.hops ?? [])].sort((a, b) => a.hopIndex - b.hopIndex);
}

function hopNodeId(hop: TracerouteHop) {
	return hop.address || hop.hostname || `unknown:${hop.hopIndex}`;
}

function runPathSignature(run: TracerouteResult) {
	return orderedHops(run)
		.map(hop => hopNodeId(hop))
		.join(">");
}

function lastRespondingHop(run: TracerouteResult | null | undefined) {
	return orderedHops(run)
		.reverse()
		.find(hop => hop.receivedCount > 0 || typeof hop.rttAvgMs === "number");
}

function hopLabel(hop: TracerouteHop) {
	return hop.hostname || hop.address || `unknown hop ${hop.hopIndex}`;
}

function hasMeaningfulRttJump(previousAvg: number | null, currentAvg: number | null) {
	return typeof previousAvg === "number" && typeof currentAvg === "number" && currentAvg - previousAvg > 10 && currentAvg > previousAvg * 1.2;
}

export function buildHopDiagnostics(run: TracerouteResult | null | undefined): HopDiagnostic[] {
	const hops = orderedHops(run);
	const lastResponding = [...hops].reverse().find(hop => hop.receivedCount > 0 || typeof hop.rttAvgMs === "number");
	const lastRespondingIndex = lastResponding?.hopIndex ?? null;
	let previousRespondingAvg: number | null = null;

	return hops.map(hop => {
		const avgRtt = typeof hop.rttAvgMs === "number" ? hop.rttAvgMs : null;
		const minRtt = typeof hop.rttMinMs === "number" ? hop.rttMinMs : avgRtt;
		const medianRtt = typeof hop.rttMedianMs === "number" ? hop.rttMedianMs : avgRtt;
		const maxRtt = typeof hop.rttMaxMs === "number" ? hop.rttMaxMs : avgRtt;
		const loss = typeof hop.lossPercent === "number" ? hop.lossPercent : 0;
		const noReply = hop.sentCount > 0 && hop.receivedCount === 0;
		const downstreamResponding = hops.filter(candidate => candidate.hopIndex > hop.hopIndex && (candidate.receivedCount > 0 || typeof candidate.rttAvgMs === "number"));
		const lossPropagatesDownstream = loss >= 1 && downstreamResponding.length > 0 && downstreamResponding.every(candidate => candidate.lossPercent >= 1);
		const finalHopLoss = loss >= 1 && hop.hopIndex === lastRespondingIndex;
		const propagatedLoss = lossPropagatesDownstream || finalHopLoss;
		const rttJump = hasMeaningfulRttJump(previousRespondingAvg, avgRtt);
		let tone: HopTone = "success";
		let state = "Clear";

		if (noReply) {
			tone = "muted";
			state = "No reply";
		} else if (propagatedLoss) {
			tone = "critical";
			state = finalHopLoss ? "Final loss" : "Propagated loss";
		} else if (rttJump) {
			tone = "warning";
			state = "RTT jump";
		} else if (loss >= 1) {
			tone = "warning";
			state = "Router-only loss";
		}

		if (avgRtt !== null) {
			previousRespondingAvg = avgRtt;
		}

		return {
			id: `${hop.hopIndex}-${hopNodeId(hop)}`,
			hopIndex: hop.hopIndex,
			label: hopLabel(hop),
			address: hop.address || hop.hostname || "unknown",
			sent: hop.sentCount,
			received: hop.receivedCount,
			loss,
			minRtt,
			avgRtt,
			medianRtt,
			maxRtt,
			sampleCount: hop.rttSamplesMs?.length ?? 0,
			state,
			tone,
			routerOnlyLoss: loss >= 1 && !propagatedLoss,
			propagatedLoss,
			rttJump,
			noReply,
			error: hop.errorMessage || hop.errorCode || ""
		};
	});
}

export function summarizeTraceroute(runs: TracerouteResult[], selectedRun: TracerouteResult | null): TracerouteSummary {
	const diagnostics = buildHopDiagnostics(selectedRun);
	const finalHop = lastRespondingHop(selectedRun);
	const chronologicalRuns = [...runs].sort((a, b) => Date.parse(a.startedAt) - Date.parse(b.startedAt));
	let pathChangeCount = 0;
	let previousSignature = "";

	for (const run of chronologicalRuns) {
		const signature = runPathSignature(run);

		if (previousSignature && signature && signature !== previousSignature) {
			pathChangeCount += 1;
		}

		if (signature) {
			previousSignature = signature;
		}
	}

	if (!selectedRun) {
		return {
			statusTone: "muted",
			statusLabel: "No data",
			finalRtt: null,
			finalLoss: null,
			firstPropagatedLossHop: null,
			firstRttJumpHop: null,
			pathChangeCount
		};
	}

	const statusTone = selectedRun.status === "successful" && selectedRun.destinationReached ? "success" : selectedRun.status === "error" || selectedRun.status === "timeout" ? "critical" : "warning";

	return {
		statusTone,
		statusLabel: selectedRun.destinationReached ? "Reached" : selectedRun.status,
		finalRtt: finalHop?.rttAvgMs ?? null,
		finalLoss: finalHop?.lossPercent ?? null,
		firstPropagatedLossHop: diagnostics.find(hop => hop.propagatedLoss)?.hopIndex ?? null,
		firstRttJumpHop: diagnostics.find(hop => hop.rttJump)?.hopIndex ?? null,
		pathChangeCount
	};
}

export function runFinalRtt(run: TracerouteResult) {
	return lastRespondingHop(run)?.rttAvgMs ?? null;
}

function runTimelineTone(value: number | null, hasLoss: boolean | undefined): RunTimelinePoint["tone"] {
	if (hasLoss) {
		return "critical";
	}

	if (typeof value === "number" && value >= 100) {
		return "warning";
	}

	return "normal";
}

export function tracerouteInsightTimelinePoints(data: TracerouteInsightResponse | undefined): RunTimelinePoint[] {
	return (data?.points ?? []).map(point => {
		const value = point.finalRttAvgMs ?? null;
		const labelTime = new Date(point.timestampMs).toISOString();
		const isRawRun = data?.query.resolution === "raw" && Boolean(point.runStartedAt);

		return {
			id: point.runStartedAt || `${point.bucketFromMs}:${point.bucketToMs}`,
			timestampMs: point.timestampMs,
			rangeFromMs: point.bucketFromMs,
			rangeToMs: point.bucketToMs,
			runStartedAt: point.runStartedAt,
			label: formatShortTime(labelTime),
			value,
			valueLabel: formatMs(value),
			ariaLabel: isRawRun
				? `Select traceroute run ${formatTime(point.runStartedAt || labelTime)} final RTT ${formatMs(value)} loss ${formatPercent(point.finalLossPercent)}`
				: `Narrow traceroute timeline bucket ${formatShortTime(labelTime)} final RTT ${formatMs(value)} loss ${formatPercent(point.finalLossPercent)}`,
			tone: runTimelineTone(value, point.hasLoss),
			hasLoss: point.hasLoss,
			hasChange: point.hasRouteChange
		};
	});
}

export function selectedTimelineValueLabel(selectedRun: TracerouteResult | null, points: RunTimelinePoint[]) {
	const selectedValue = selectedRun ? formatMs(runFinalRtt(selectedRun)) : points[points.length - 1]?.valueLabel || "-";

	return `${selectedValue} selected`;
}
