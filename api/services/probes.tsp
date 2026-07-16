import "@typespec/http";
import "@typespec/openapi";

using Http;
using OpenAPI;

namespace Netstamp;

@tag("Probes")
@useAuth(UserOrTokenAuth)
@route("/projects/{ref}/probes")
interface ProbeService {
	@operationId("listProjectProbes")
	@summary("List project probes")
	@get
	listProjectProbes(...ProjectRefParam): OkJson<ProbeListResponse> | UnauthorizedProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("createProjectProbe")
	@summary("Create project probe")
	@post
	createProjectProbe(...ProjectRefParam, @body body: CreateProbeRequest):
		| CreatedJson<ProbeSecretResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("getProjectProbe")
	@summary("Get project probe")
	@get
	@route("/{probe_id}")
	getProjectProbe(...ProjectRefParam, ...ProbeIdPathParam): OkJson<ProbeResponse> | UnauthorizedProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("updateProjectProbe")
	@summary("Update project probe")
	@patch
	@route("/{probe_id}")
	updateProjectProbe(...ProjectRefParam, ...ProbeIdPathParam, @body body: UpdateProbeRequest):
		| OkJson<ProbeResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("deleteProjectProbe")
	@summary("Delete project probe")
	@delete
	@route("/{probe_id}")
	deleteProjectProbe(...ProjectRefParam, ...ProbeIdPathParam): NoContent | UnauthorizedProblem | ForbiddenProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("rotateProjectProbeSecret")
	@summary("Rotate project probe secret")
	@post
	@route("/{probe_id}/secret-rotations")
	rotateProjectProbeSecret(...ProjectRefParam, ...ProbeIdPathParam):
		| OkJson<ProbeSecretResponse>
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;
}
