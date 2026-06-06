import "@typespec/http";
import "@typespec/openapi";

using Http;
using OpenAPI;

namespace Netstamp;

@tag("Probe Runtime")
@useAuth(ProbeAuth)
@route("/runtime/probes/{probe_id}")
interface ProbeRuntimeService {
	@operationId("probeRuntimeHello")
	@summary("Start probe runtime session")
	@post
	@route("/hello")
	probeRuntimeHello(...ProbeIdPathParam): OkJson<RuntimeHelloResponse> | UnauthorizedProblem | ForbiddenProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("probeRuntimeHeartbeat")
	@summary("Update probe runtime status")
	@post
	@route("/heartbeat")
	probeRuntimeHeartbeat(...ProbeIdPathParam, @body body: RuntimeStatusRequest):
		| OkJson<RuntimeHeartbeatResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("updateProbeRuntimeIPFamilyCapabilities")
	@summary("Update probe runtime IP family capabilities")
	@put
	@route("/ip-family-capabilities")
	updateProbeRuntimeIPFamilyCapabilities(...ProbeIdPathParam, @body body?: RuntimeIPFamilyCapabilitiesRequest):
		| OkJson<RuntimeIPFamilyCapabilitiesResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("listProbeRuntimeAssignments")
	@summary("List probe runtime assignments")
	@get
	@route("/assignments")
	listProbeRuntimeAssignments(...ProbeIdPathParam): OkJson<RuntimeAssignmentsResponse> | UnauthorizedProblem | ForbiddenProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("submitProbeRuntimeResults")
	@summary("Submit probe runtime results")
	@post
	@route("/results")
	submitProbeRuntimeResults(...ProbeIdPathParam, @body body: SubmitResultsRequest):
		| OkJson<SubmitResultsResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;
}
