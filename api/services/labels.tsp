import "@typespec/http";
import "@typespec/openapi";

using Http;
using OpenAPI;

namespace Netstamp;

@tag("Labels")
@useAuth(SessionCookieAuth)
@route("/projects/{ref}/labels")
interface LabelService {
	@operationId("listProjectLabels")
	@summary("List project labels")
	@get
	listProjectLabels(...ProjectRefParam): OkJson<LabelListResponse> | UnauthorizedProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("createProjectLabel")
	@summary("Create project label")
	@post
	createProjectLabel(...ProjectRefParam, @body body: CreateLabelRequest):
		| CreatedJson<LabelResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| ConflictProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("updateProjectLabel")
	@summary("Update project label")
	@patch
	@route("/{label_id}")
	updateProjectLabel(...ProjectRefParam, ...LabelIdPathParam, @body body: UpdateLabelRequest):
		| OkJson<LabelResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| ConflictProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("deleteProjectLabel")
	@summary("Delete project label")
	@delete
	@route("/{label_id}")
	deleteProjectLabel(...ProjectRefParam, ...LabelIdPathParam): NoContent | UnauthorizedProblem | ForbiddenProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;
}
