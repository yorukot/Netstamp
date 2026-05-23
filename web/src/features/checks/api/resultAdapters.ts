import type { CheckDefinition } from "@/features/checks/data/checks";
import type { AssignedRow } from "@/features/probes/components/types";
import type { Probe } from "@/features/probes/data/probes";
import type { ApiMeasurement, ApiProjectAssignment } from "@/shared/api/types";
import { formatInterval } from "./checkAdapters";

export interface LogRow {
	time: string;
	check: string;
	probe: string;
	status: string;
	latency: string;
	event: string;
}

function statusLabel(status: ApiMeasurement["status"]) {
	if (status === "successful") {
		return "success";
	}

	return status;
}

function measurementLatency(measurement: ApiMeasurement) {
	if (typeof measurement.latencyMs === "number") {
		return `${Math.round(measurement.latencyMs)}ms`;
	}

	if (typeof measurement.durationMs === "number") {
		return `${Math.round(measurement.durationMs)}ms`;
	}

	return "-";
}

function measurementEvent(measurement: ApiMeasurement) {
	return measurement.errorMessage || measurement.errorCode || measurement.metadata || measurement.type;
}

function namesById<TItem extends { id: string; name: string }>(items: TItem[]) {
	return new Map(items.map(item => [item.id, item.name]));
}

export function mapApiMeasurements(measurements: ApiMeasurement[] | null | undefined, probes: Probe[], checks: CheckDefinition[]): LogRow[] {
	const probeNames = namesById(probes);
	const checkNames = namesById(checks);

	return (measurements ?? []).map(measurement => ({
		time: new Date(measurement.startedAt).toLocaleString(),
		check: checkNames.get(measurement.checkId) || measurement.checkId,
		probe: probeNames.get(measurement.probeId) || measurement.probeId,
		status: statusLabel(measurement.status),
		latency: measurementLatency(measurement),
		event: measurementEvent(measurement)
	}));
}

export function mapApiAssignments(assignments: ApiProjectAssignment[] | null | undefined, probes: Probe[], checks: CheckDefinition[]): AssignedRow[] {
	const probeNames = namesById(probes);
	const checksById = new Map(checks.map(check => [check.id, check]));

	return (assignments ?? []).map(assignment => {
		const check = checksById.get(assignment.checkId);

		return {
			probe: assignment.probe?.name || probeNames.get(assignment.probeId) || assignment.probeId,
			check: assignment.check?.name || check?.name || assignment.checkId,
			type: check?.type || "Ping",
			interval: check?.interval || (assignment.check ? formatInterval(assignment.check.intervalSeconds) : "-"),
			latest: check?.latest || "-"
		};
	});
}
