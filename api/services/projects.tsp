import "@typespec/http";
import "@typespec/openapi";

using Http;
using OpenAPI;

namespace Netstamp;

@tag("Projects")
@useAuth(SessionCookieAuth)
@route("/projects")
interface ProjectService {
	@operationId("createProject")
	@summary("Create project")
	@post
	createProject(@body body: CreateProjectRequest): CreatedJson<ProjectResponse> | BadRequestProblem | UnauthorizedProblem | ConflictProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("listProjects")
	@summary("List projects")
	@get
	listProjects(): OkJson<ProjectListResponse> | UnauthorizedProblem | InternalServerProblem;

	@operationId("getProject")
	@summary("Get project")
	@get
	@route("/{ref}")
	getProject(...ProjectRefParam): OkJson<ProjectResponse> | UnauthorizedProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("updateProject")
	@summary("Update project")
	@patch
	@route("/{ref}")
	updateProject(...ProjectRefParam, @body body: UpdateProjectRequest):
		| OkJson<ProjectResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| ConflictProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("deleteProject")
	@summary("Delete project")
	@delete
	@route("/{ref}")
	deleteProject(...ProjectRefParam): NoContent | UnauthorizedProblem | ForbiddenProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;
}

@tag("Project Members")
@useAuth(SessionCookieAuth)
@route("/projects/{ref}/members")
interface ProjectMemberService {
	@operationId("listProjectMembers")
	@summary("List project members")
	@get
	listProjectMembers(...ProjectRefParam): OkJson<ProjectMemberListResponse> | UnauthorizedProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("addProjectMember")
	@summary("Add project member")
	@post
	addProjectMember(...ProjectRefParam, @body body: AddProjectMemberRequest):
		| CreatedJson<ProjectMemberResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| ConflictProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("updateProjectMemberRole")
	@summary("Update project member role")
	@patch
	@route("/{user_id}")
	updateProjectMemberRole(...ProjectRefParam, ...UserIdPathParam, @body body: UpdateProjectMemberRoleRequest):
		| OkJson<ProjectMemberResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| ConflictProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("removeProjectMember")
	@summary("Remove project member")
	@delete
	@route("/{user_id}")
	removeProjectMember(...ProjectRefParam, ...UserIdPathParam):
		| NoContent
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| ConflictProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;
}
