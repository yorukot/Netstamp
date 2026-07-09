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

@tag("Project Invites")
@useAuth(SessionCookieAuth)
@route("/projects/{ref}/invites")
interface ProjectInviteService {
	@operationId("listProjectInvites")
	@summary("List pending project invites")
	@get
	listProjectInvites(...ProjectRefParam):
		| OkJson<ProjectInviteListResponse>
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("createProjectInvite")
	@summary("Create project invite")
	@post
	createProjectInvite(...ProjectRefParam, @body body: CreateProjectInviteRequest):
		| CreatedJson<ProjectInviteResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| ConflictProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("cancelProjectInvite")
	@summary("Cancel project invite")
	@delete
	@route("/{invite_id}")
	cancelProjectInvite(...ProjectRefParam, ...ProjectInviteIdPathParam):
		| NoContent
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;
}

@tag("Project Invites")
@useAuth(SessionCookieAuth)
@route("/me/project-invites")
interface CurrentUserProjectInviteService {
	@operationId("listCurrentUserProjectInvites")
	@summary("List current user's pending project invites")
	@get
	listCurrentUserProjectInvites(): OkJson<ProjectInviteListResponse> | UnauthorizedProblem | InternalServerProblem;

	@operationId("acceptProjectInvite")
	@summary("Accept project invite")
	@post
	@route("/{invite_id}/accept")
	acceptProjectInvite(...ProjectInviteIdPathParam):
		| OkJson<ProjectInviteResponse>
		| UnauthorizedProblem
		| NotFoundProblem
		| ConflictProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("rejectProjectInvite")
	@summary("Reject project invite")
	@post
	@route("/{invite_id}/reject")
	rejectProjectInvite(...ProjectInviteIdPathParam):
		| OkJson<ProjectInviteResponse>
		| UnauthorizedProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;
}
