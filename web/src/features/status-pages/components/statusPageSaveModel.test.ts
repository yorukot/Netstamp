import { describe, expect, it } from "vitest";
import { statusPageSaveElementRequest, type StatusPageElementDraft } from "./statusPageSaveModel";

describe("statusPageSaveElementRequest", () => {
	it("serializes a new child with its request-local parent reference", () => {
		const element: StatusPageElementDraft = {
			localId: "child-local",
			parentLocalId: "folder-local",
			kind: "assignment_group",
			checkId: "44444444-4444-4444-4444-444444444444",
			assignmentSelectionMode: "all_check",
			assignmentIds: [],
			title: " API ",
			description: " Public endpoint ",
			sortOrder: 1,
			displayMode: "history",
			chartMode: "compact",
			chartRange: "7d"
		};

		expect(statusPageSaveElementRequest(element)).toEqual({
			clientId: "child-local",
			id: undefined,
			parentClientId: "folder-local",
			kind: "assignment_group",
			checkId: "44444444-4444-4444-4444-444444444444",
			assignmentSelectionMode: "all_check",
			assignmentIds: undefined,
			title: "API",
			description: "Public endpoint",
			sortOrder: 1,
			displayMode: "history",
			chartMode: "compact",
			chartRange: "7d"
		});
	});

	it("preserves an existing id and strips folder-only incompatible fields", () => {
		const element: StatusPageElementDraft = {
			localId: "folder-local",
			persistedId: "55555555-5555-5555-5555-555555555555",
			kind: "folder",
			checkId: "44444444-4444-4444-4444-444444444444",
			assignmentSelectionMode: "selected_assignments",
			assignmentIds: ["aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"],
			title: "Core",
			sortOrder: 0,
			displayMode: "map",
			chartMode: "compact",
			chartRange: "30d"
		};

		expect(statusPageSaveElementRequest(element)).toMatchObject({
			clientId: "folder-local",
			id: "55555555-5555-5555-5555-555555555555",
			parentClientId: undefined,
			checkId: undefined,
			assignmentSelectionMode: undefined,
			assignmentIds: undefined,
			displayMode: "status",
			chartMode: "inherit",
			chartRange: undefined
		});
	});
});
