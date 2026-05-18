import "@typespec/http";
import "@typespec/openapi";

using Http;
using OpenAPI;

namespace Netstamp;

@tag("Checks")
@useAuth(SessionCookieAuth)
@route("/projects/{ref}/checks")
interface CheckService {
	@operationId("listProjectChecks")
	@summary("List project checks")
	@get
	listProjectChecks(...ProjectRefParam): OkJson<CheckListResponse> | UnauthorizedProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("createProjectCheck")
	@summary("Create project check")
	@post
	createProjectCheck(...ProjectRefParam, @body body: CreateCheckRequest):
		| CreatedJson<CheckResponse>
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("getProjectCheck")
	@summary("Get project check")
	@get
	@route("/{check_id}")
	getProjectCheck(...ProjectRefParam, ...CheckIdPathParam): OkJson<CheckResponse> | UnauthorizedProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("updateProjectCheck")
	@summary("Update project check")
	@patch
	@route("/{check_id}")
	updateProjectCheck(...ProjectRefParam, ...CheckIdPathParam, @body body: UpdateCheckRequest):
		| OkJson<CheckResponse>
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("deleteProjectCheck")
	@summary("Delete project check")
	@delete
	@route("/{check_id}")
	deleteProjectCheck(...ProjectRefParam, ...CheckIdPathParam): NoContent | UnauthorizedProblem | ForbiddenProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;
}
