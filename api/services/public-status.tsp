import "@typespec/http";
import "@typespec/openapi";

using Http;
using OpenAPI;

namespace Netstamp;

@tag("Public Status Pages")
@useAuth(SessionCookieAuth)
@route("/projects/{ref}/status-pages")
interface ProjectPublicStatusPageService {
	@operationId("listProjectPublicStatusPages")
	@summary("List project public status pages")
	@get
	listProjectPublicStatusPages(...ProjectRefParam): OkJson<PublicStatusPageListResponse> | UnauthorizedProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("createProjectPublicStatusPage")
	@summary("Create project public status page")
	@post
	createProjectPublicStatusPage(...ProjectRefParam, @body body: CreatePublicStatusPageRequest):
		| CreatedJson<PublicStatusPageResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| ConflictProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("getProjectPublicStatusPage")
	@summary("Get project public status page")
	@get
	@route("/{page_id}")
	getProjectPublicStatusPage(...ProjectRefParam, ...PublicStatusPageIdPathParam):
		| OkJson<PublicStatusPageDetailResponse>
		| UnauthorizedProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("updateProjectPublicStatusPage")
	@summary("Update project public status page")
	@patch
	@route("/{page_id}")
	updateProjectPublicStatusPage(...ProjectRefParam, ...PublicStatusPageIdPathParam, @body body: UpdatePublicStatusPageRequest):
		| OkJson<PublicStatusPageResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| ConflictProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("deleteProjectPublicStatusPage")
	@summary("Delete project public status page")
	@delete
	@route("/{page_id}")
	deleteProjectPublicStatusPage(...ProjectRefParam, ...PublicStatusPageIdPathParam):
		| NoContent
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("createProjectPublicStatusElement")
	@summary("Create public status page element")
	@post
	@route("/{page_id}/elements")
	createProjectPublicStatusElement(...ProjectRefParam, ...PublicStatusPageIdPathParam, @body body: CreatePublicStatusElementRequest):
		| CreatedJson<PublicStatusElementResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("updateProjectPublicStatusElement")
	@summary("Update public status page element")
	@patch
	@route("/{page_id}/elements/{element_id}")
	updateProjectPublicStatusElement(...ProjectRefParam, ...PublicStatusPageIdPathParam, ...PublicStatusElementIdPathParam, @body body: UpdatePublicStatusElementRequest):
		| OkJson<PublicStatusElementResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("deleteProjectPublicStatusElement")
	@summary("Delete public status page element")
	@delete
	@route("/{page_id}/elements/{element_id}")
	deleteProjectPublicStatusElement(...ProjectRefParam, ...PublicStatusPageIdPathParam, ...PublicStatusElementIdPathParam):
		| NoContent
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;
}

@tag("Public Status Pages")
@route("/public/status-pages")
interface PublicStatusPageService {
	@operationId("getPublicStatusPage")
	@summary("Get public status page")
	@get
	@route("/{slug}")
	getPublicStatusPage(...PublicStatusSlugPathParam, ...PublicStatusQuery):
		| OkJson<PublicStatusPublicResponse>
		| BadRequestProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("getPublicStatusSummary")
	@summary("Get public status page summary")
	@get
	@route("/{slug}/summary")
	getPublicStatusSummary(...PublicStatusSlugPathParam): OkJson<PublicStatusSummaryResponse> | BadRequestProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("getPublicStatusElements")
	@summary("Get public status page elements")
	@get
	@route("/{slug}/elements")
	getPublicStatusElements(...PublicStatusSlugPathParam): OkJson<PublicStatusElementsResponse> | BadRequestProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("getPublicStatusIncidents")
	@summary("Get public status page incidents")
	@get
	@route("/{slug}/incidents")
	getPublicStatusIncidents(...PublicStatusSlugPathParam, ...PublicStatusIncidentsQuery):
		| OkJson<PublicStatusIncidentsResponse>
		| BadRequestProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("getPublicStatusElementChart")
	@summary("Get public status page element chart")
	@get
	@route("/{slug}/elements/{element_id}/chart")
	getPublicStatusElementChart(...PublicStatusSlugPathParam, ...PublicStatusElementIdPathParam, ...PublicStatusChartQuery):
		| OkJson<PublicStatusElementChartResponse>
		| BadRequestProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;
}
