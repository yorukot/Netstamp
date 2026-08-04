import type { ApiProjectAssignment, CreatePublicStatusElementInput } from "@/shared/api/types";

export interface StatusAssignmentOption {
	id: string;
	checkId: string;
	checkName: string;
	checkType: string;
	checkTarget: string;
	probeId: string;
	probeName: string;
	probeLocationName?: string;
	latitude?: number;
	longitude?: number;
}

export interface StatusCheckOption {
	id: string;
	name: string;
	type: string;
	target: string;
	assignmentIds: string[];
	viewpoints: Array<{
		id: string;
		name: string;
		locationName?: string;
		latitude?: number;
		longitude?: number;
	}>;
}

export type StatusAssignmentSelectionMode = NonNullable<CreatePublicStatusElementInput["assignmentSelectionMode"]>;

export interface StatusAssignmentScope {
	checkId?: string;
	assignmentSelectionMode?: StatusAssignmentSelectionMode;
	assignmentIds?: string[];
}

export const supportedStatusCheckTypes = new Set(["http", "ping", "tcp"]);

export const statusAssignmentOptions = (assignments: ApiProjectAssignment[]): StatusAssignmentOption[] =>
	assignments
		.flatMap(assignment => {
			if (!assignment.check || !supportedStatusCheckTypes.has(assignment.check.type)) {
				return [];
			}

			return [
				{
					id: assignment.id,
					checkId: assignment.checkId,
					checkName: assignment.check.name,
					checkType: assignment.check.type,
					checkTarget: assignment.check.target,
					probeId: assignment.probeId,
					probeName: assignment.probe?.name ?? assignment.probeId,
					probeLocationName: assignment.probe?.locationName,
					latitude: assignment.probe?.latitude,
					longitude: assignment.probe?.longitude
				}
			];
		})
		.sort((left, right) => left.checkName.localeCompare(right.checkName) || left.probeName.localeCompare(right.probeName) || left.id.localeCompare(right.id));

export const statusCheckOptions = (assignments: StatusAssignmentOption[]): StatusCheckOption[] => {
	const options = new Map<string, StatusCheckOption>();

	for (const assignment of assignments) {
		const option = options.get(assignment.checkId) ?? {
			id: assignment.checkId,
			name: assignment.checkName,
			type: assignment.checkType,
			target: assignment.checkTarget,
			assignmentIds: [],
			viewpoints: []
		};
		option.assignmentIds.push(assignment.id);
		if (!option.viewpoints.some(viewpoint => viewpoint.id === assignment.probeId)) {
			option.viewpoints.push({
				id: assignment.probeId,
				name: assignment.probeName,
				locationName: assignment.probeLocationName,
				latitude: assignment.latitude,
				longitude: assignment.longitude
			});
		}
		options.set(assignment.checkId, option);
	}

	return [...options.values()].sort((left, right) => left.name.localeCompare(right.name) || left.id.localeCompare(right.id));
};

export const assignmentsForScope = (scope: StatusAssignmentScope, assignments: StatusAssignmentOption[]) => {
	if (scope.assignmentSelectionMode === "selected_assignments") {
		const selected = new Set(scope.assignmentIds ?? []);
		return assignments.filter(assignment => selected.has(assignment.id));
	}

	return assignments.filter(assignment => assignment.checkId === scope.checkId);
};

export const inferSingleCheckId = (assignmentIds: string[] | undefined, assignments: StatusAssignmentOption[]) => {
	const selected = new Set(assignmentIds ?? []);
	const checkIds = new Set(assignments.filter(assignment => selected.has(assignment.id)).map(assignment => assignment.checkId));
	return checkIds.size === 1 ? [...checkIds][0] : undefined;
};

export const statusAssignmentScopeForMode = (
	scope: StatusAssignmentScope,
	assignmentSelectionMode: StatusAssignmentSelectionMode,
	checks: StatusCheckOption[],
	assignments: StatusAssignmentOption[]
): StatusAssignmentScope => {
	if (assignmentSelectionMode === "selected_assignments") {
		return {
			checkId: scope.checkId,
			assignmentSelectionMode,
			assignmentIds: scope.assignmentIds?.length ? scope.assignmentIds : (checks.find(check => check.id === scope.checkId)?.assignmentIds ?? [])
		};
	}

	return {
		assignmentIds: scope.assignmentIds,
		assignmentSelectionMode,
		checkId: scope.checkId ?? inferSingleCheckId(scope.assignmentIds, assignments)
	};
};

export const unavailableAssignmentIds = (assignmentIds: string[] | undefined, assignments: StatusAssignmentOption[]) => {
	const available = new Set(assignments.map(assignment => assignment.id));
	return (assignmentIds ?? []).filter(assignmentId => !available.has(assignmentId));
};

export const statusAssignmentRequestScope = (scope: StatusAssignmentScope): Pick<CreatePublicStatusElementInput, "assignmentIds" | "assignmentSelectionMode" | "checkId"> => {
	const assignmentSelectionMode = scope.assignmentSelectionMode ?? "all_check";
	return {
		assignmentSelectionMode,
		checkId: assignmentSelectionMode === "all_check" ? scope.checkId : undefined,
		assignmentIds: assignmentSelectionMode === "selected_assignments" ? scope.assignmentIds : undefined
	};
};

export const defaultSelectedScopeTitle = (assignmentIds: string[], assignments: StatusAssignmentOption[], fallback: string) => {
	const selected = new Set(assignmentIds);
	const selectedAssignments = assignments.filter(assignment => selected.has(assignment.id));
	const names = [...new Set(selectedAssignments.map(assignment => assignment.checkName))];
	return names.length === 1 ? names[0] : fallback;
};
