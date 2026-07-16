import "@typespec/http";
import "@typespec/openapi";

using Http;
using OpenAPI;

namespace Netstamp;

@tag("Assignments")
@useAuth(UserOrTokenAuth)
@route("/projects/{ref}")
interface AssignmentService {
	@operationId("previewProjectSelector")
	@summary("Preview project selector")
	@doc("Validate a selector AST, return its canonical form, and list the project probes it currently matches.")
	@post
	@route("/selector-previews")
	previewProjectSelector(...ProjectRefParam, @body body: SelectorPreviewRequest):
		| OkJson<SelectorPreviewResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("listProjectAssignments")
	@summary("List project assignments")
	@doc("List effective probe-check assignments for authenticated console users. Probe runtime polling uses the separate runtime endpoint.")
	@get
	@route("/assignments")
	listProjectAssignments(...ProjectRefParam, ...ProjectAssignmentQuery):
		| OkJson<ProjectAssignmentListResponse>
		| UnauthorizedProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;
}
