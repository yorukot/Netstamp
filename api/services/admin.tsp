import "@typespec/http";
import "@typespec/openapi";

using Http;
using OpenAPI;

namespace Netstamp;

@tag("Admin")
@useAuth(SessionCookieAuth)
@route("/admin")
interface AdminService {
	@operationId("listSystemAdmins")
	@summary("List system administrators")
	@doc("Return users with instance-level administrator access. This does not include project membership or project permissions.")
	@get
	@route("/system-admins")
	listSystemAdmins(): OkJson<SystemAdminsResponse> | UnauthorizedProblem | ForbiddenProblem | InternalServerProblem;

	@operationId("grantSystemAdmin")
	@summary("Grant system administrator")
	@doc("Grant instance-level administrator access to an existing user by email. This does not add the user to any project.")
	@post
	@route("/system-admins")
	grantSystemAdmin(@body body: GrantSystemAdminRequest): CreatedJson<SystemAdminResponse> | BadRequestProblem | UnauthorizedProblem | ForbiddenProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("revokeSystemAdmin")
	@summary("Revoke system administrator")
	@doc("Revoke instance-level administrator access from a user. The current administrator cannot revoke themselves, and the system must keep at least one administrator.")
	@delete
	@route("/system-admins/{user_id}")
	revokeSystemAdmin(...UserIdPathParam): NoContent | UnauthorizedProblem | ForbiddenProblem | NotFoundProblem | ConflictProblem | InternalServerProblem;

	@operationId("getAdminSettings")
	@summary("Get admin settings")
	@doc("Return instance-level settings visible to global administrators. Secret values are redacted.")
	@get
	@route("/settings")
	getAdminSettings(): OkJson<AdminSettingsResponse> | UnauthorizedProblem | ForbiddenProblem | InternalServerProblem;

	@operationId("updateAdminSettings")
	@summary("Update admin settings")
	@doc("Update instance-level settings. Only global administrators can call this operation. SMTP passwords are write-only.")
	@patch
	@route("/settings")
	updateAdminSettings(@body body: UpdateAdminSettingsRequest): OkJson<AdminSettingsResponse> | BadRequestProblem | UnauthorizedProblem | ForbiddenProblem | UnprocessableEntityProblem | InternalServerProblem;
}
