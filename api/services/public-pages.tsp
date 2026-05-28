import "@typespec/http";
import "@typespec/openapi";

using Http;
using OpenAPI;

namespace Netstamp;

model PublicPageIdPathParam {
	@path
	page_id: uuid;
}

model PublicPageFolderIdPathParam {
	@path
	folder_id: uuid;
}

model PublicPageSlugPathParam {
	@path
	@minLength(1)
	@maxLength(64)
	@pattern("^[a-z0-9-]+$")
	slug: string;
}

model PublicPingInsightQuery {
	@query
	probeId: uuid;

	@query
	checkId: uuid;

	@query
	@minValue(1)
	from?: int64;

	@query
	@minValue(1)
	to?: int64;

	@query
	@minValue(1)
	@maxValue(2000)
	maxDataPoints?: int32 = 600;
}

@tag("Public Pages")
@useAuth(SessionCookieAuth)
@route("/projects/{ref}/public-pages")
interface ProjectPublicPageService {
	@operationId("listProjectPublicPages")
	@summary("List project public pages")
	@get
	listProjectPublicPages(...ProjectRefParam):
		| OkJson<PublicPageListResponse>
		| UnauthorizedProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("createProjectPublicPage")
	@summary("Create project public page")
	@post
	createProjectPublicPage(...ProjectRefParam, @body body: CreatePublicPageRequest):
		| CreatedJson<PublicPageResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| ConflictProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("getProjectPublicPage")
	@summary("Get project public page")
	@get
	@route("/{page_id}")
	getProjectPublicPage(...ProjectRefParam, ...PublicPageIdPathParam):
		| OkJson<PublicPageResponse>
		| UnauthorizedProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("updateProjectPublicPage")
	@summary("Update project public page")
	@patch
	@route("/{page_id}")
	updateProjectPublicPage(...ProjectRefParam, ...PublicPageIdPathParam, @body body: UpdatePublicPageRequest):
		| OkJson<PublicPageResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| ConflictProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("deleteProjectPublicPage")
	@summary("Delete project public page")
	@delete
	@route("/{page_id}")
	deleteProjectPublicPage(...ProjectRefParam, ...PublicPageIdPathParam):
		| NoContent
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("createProjectPublicPageFolder")
	@summary("Create project public page folder")
	@post
	@route("/{page_id}/folders")
	createProjectPublicPageFolder(...ProjectRefParam, ...PublicPageIdPathParam, @body body: CreatePublicPageFolderRequest):
		| CreatedJson<PublicPageFolderResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("updateProjectPublicPageFolder")
	@summary("Update project public page folder")
	@patch
	@route("/{page_id}/folders/{folder_id}")
	updateProjectPublicPageFolder(...ProjectRefParam, ...PublicPageIdPathParam, ...PublicPageFolderIdPathParam, @body body: UpdatePublicPageFolderRequest):
		| OkJson<PublicPageFolderResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("deleteProjectPublicPageFolder")
	@summary("Delete project public page folder")
	@delete
	@route("/{page_id}/folders/{folder_id}")
	deleteProjectPublicPageFolder(...ProjectRefParam, ...PublicPageIdPathParam, ...PublicPageFolderIdPathParam):
		| NoContent
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("setProjectPublicPageFolderChecks")
	@summary("Set project public page folder checks")
	@put
	@route("/{page_id}/folders/{folder_id}/checks")
	setProjectPublicPageFolderChecks(...ProjectRefParam, ...PublicPageIdPathParam, ...PublicPageFolderIdPathParam, @body body: SetPublicPageFolderChecksRequest):
		| OkJson<PublicPageFolderChecksResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| ConflictProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;
}

@tag("Public Pages")
@route("/public-pages/{slug}")
interface PublicPageService {
	@operationId("getPublicPage")
	@summary("Get public page")
	@get
	getPublicPage(...PublicPageSlugPathParam): OkJson<PublicPageResponse> | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("queryPublicPagePingInsight")
	@summary("Query public page ping insight")
	@get
	@route("/results/ping/insight")
	queryPublicPagePingInsight(...PublicPageSlugPathParam, ...PublicPingInsightQuery):
		| OkJson<PublicPingInsightResponse>
		| BadRequestProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;
}
