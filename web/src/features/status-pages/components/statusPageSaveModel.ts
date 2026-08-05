import type { CreatePublicStatusElementInput, SavePublicStatusPageElementInput } from "@/shared/api/types";
import { statusAssignmentRequestScope } from "./statusAssignmentModel";

export interface StatusPageElementDraft extends Omit<CreatePublicStatusElementInput, "parentElementId"> {
	localId: string;
	persistedId?: string;
	parentLocalId?: string;
}

export const statusPageSaveElementRequest = (element: StatusPageElementDraft): SavePublicStatusPageElementInput => ({
	clientId: element.localId,
	id: element.persistedId,
	parentClientId: element.kind === "assignment_group" ? element.parentLocalId : undefined,
	kind: element.kind,
	...(element.kind === "assignment_group" ? statusAssignmentRequestScope(element) : { checkId: undefined, assignmentSelectionMode: undefined, assignmentIds: undefined }),
	title: element.title?.trim() || undefined,
	description: element.description?.trim() || undefined,
	sortOrder: element.sortOrder,
	displayMode: element.kind === "folder" ? "status" : element.displayMode,
	chartMode: element.kind === "folder" ? "inherit" : element.chartMode,
	chartRange: element.kind === "folder" ? undefined : element.chartRange
});
