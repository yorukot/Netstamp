import type { ApiProjectAssignment } from "@/shared/api/types";
import { describe, expect, it } from "vitest";
import {
	assignmentsForScope,
	defaultSelectedScopeTitle,
	inferSingleCheckId,
	statusAssignmentOptions,
	statusAssignmentRequestScope,
	statusAssignmentScopeForMode,
	statusCheckOptions,
	unavailableAssignmentIds
} from "./statusAssignmentModel";

const assignment = ({ id, checkId, checkName, checkType = "ping", probeId, probeName }: { id: string; checkId: string; checkName: string; checkType?: string; probeId: string; probeName: string }) =>
	({
		id,
		projectId: "project-1",
		probeId,
		checkId,
		checkVersion: "check-version",
		selectorVersion: "selector-version",
		check: {
			id: checkId,
			projectId: "project-1",
			name: checkName,
			type: checkType,
			target: `${checkName.toLocaleLowerCase()}.example.com`
		},
		probe: {
			id: probeId,
			projectId: "project-1",
			name: probeName,
			locationName: `${probeName} location`,
			latitude: 25,
			longitude: 121
		}
	}) as unknown as ApiProjectAssignment;

const apiAssignments = [
	assignment({ id: "assignment-a", checkId: "check-a", checkName: "GitHub Raw", probeId: "probe-tpe", probeName: "Taipei" }),
	assignment({ id: "assignment-b", checkId: "check-a", checkName: "GitHub Raw", probeId: "probe-fra", probeName: "Frankfurt" }),
	assignment({ id: "assignment-c", checkId: "check-b", checkName: "Public API", checkType: "http", probeId: "probe-tpe", probeName: "Taipei" }),
	assignment({ id: "assignment-trace", checkId: "check-trace", checkName: "Trace", checkType: "traceroute", probeId: "probe-tpe", probeName: "Taipei" })
];

describe("status assignment model", () => {
	it("normalizes supported assignments and groups them into check options", () => {
		const assignments = statusAssignmentOptions(apiAssignments);
		const checks = statusCheckOptions(assignments);

		expect(assignments.map(item => item.id)).toEqual(["assignment-b", "assignment-a", "assignment-c"]);
		expect(checks).toHaveLength(2);
		expect(checks.find(check => check.id === "check-a")?.assignmentIds).toEqual(["assignment-b", "assignment-a"]);
		expect(checks.find(check => check.id === "check-a")?.viewpoints.map(viewpoint => viewpoint.id)).toEqual(["probe-fra", "probe-tpe"]);
	});

	it("resolves dynamic check scope and exact selected assignment scope", () => {
		const assignments = statusAssignmentOptions(apiAssignments);

		expect(assignmentsForScope({ checkId: "check-a", assignmentSelectionMode: "all_check" }, assignments).map(item => item.id)).toEqual(["assignment-b", "assignment-a"]);
		expect(assignmentsForScope({ assignmentSelectionMode: "selected_assignments", assignmentIds: ["assignment-b", "assignment-c"] }, assignments).map(item => item.id)).toEqual([
			"assignment-b",
			"assignment-c"
		]);
	});

	it("infers a single check without collapsing cross-check selections", () => {
		const assignments = statusAssignmentOptions(apiAssignments);

		expect(inferSingleCheckId(["assignment-a", "assignment-b"], assignments)).toBe("check-a");
		expect(inferSingleCheckId(["assignment-a", "assignment-c"], assignments)).toBeUndefined();
	});

	it("changes selection mode while preserving reversible scope state", () => {
		const assignments = statusAssignmentOptions(apiAssignments);
		const checks = statusCheckOptions(assignments);

		expect(statusAssignmentScopeForMode({ checkId: "check-a", assignmentIds: [] }, "selected_assignments", checks, assignments)).toEqual({
			checkId: "check-a",
			assignmentSelectionMode: "selected_assignments",
			assignmentIds: ["assignment-b", "assignment-a"]
		});
		expect(statusAssignmentScopeForMode({ assignmentIds: ["assignment-a", "assignment-b"] }, "all_check", checks, assignments)).toEqual({
			checkId: "check-a",
			assignmentSelectionMode: "all_check",
			assignmentIds: ["assignment-a", "assignment-b"]
		});
		expect(statusAssignmentScopeForMode({ assignmentIds: ["assignment-a", "assignment-c"] }, "all_check", checks, assignments).checkId).toBeUndefined();
	});

	it("reports unavailable IDs and derives safe default titles", () => {
		const assignments = statusAssignmentOptions(apiAssignments);

		expect(unavailableAssignmentIds(["assignment-a", "assignment-missing"], assignments)).toEqual(["assignment-missing"]);
		expect(defaultSelectedScopeTitle(["assignment-a", "assignment-b"], assignments, "Selected services")).toBe("GitHub Raw");
		expect(defaultSelectedScopeTitle(["assignment-a", "assignment-c"], assignments, "Selected services")).toBe("Selected services");
	});

	it("serializes only the fields accepted by each backend selection mode", () => {
		expect(
			statusAssignmentRequestScope({
				assignmentSelectionMode: "all_check",
				checkId: "check-a",
				assignmentIds: ["assignment-a"]
			})
		).toEqual({ assignmentSelectionMode: "all_check", checkId: "check-a", assignmentIds: undefined });
		expect(
			statusAssignmentRequestScope({
				assignmentSelectionMode: "selected_assignments",
				checkId: "check-a",
				assignmentIds: ["assignment-a"]
			})
		).toEqual({ assignmentSelectionMode: "selected_assignments", checkId: undefined, assignmentIds: ["assignment-a"] });
	});
});
