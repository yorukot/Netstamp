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
	grantSystemAdmin(@body body: GrantSystemAdminRequest):
		| CreatedJson<SystemAdminResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("revokeSystemAdmin")
	@summary("Revoke system administrator")
	@doc("Revoke instance-level administrator access from a user. The current administrator cannot revoke themselves, and the system must keep at least one administrator.")
	@delete
	@route("/system-admins/{user_id}")
	revokeSystemAdmin(...UserIdPathParam): NoContent | UnauthorizedProblem | ForbiddenProblem | NotFoundProblem | ConflictProblem | InternalServerProblem;

	@operationId("listManagedUsers")
	@summary("List managed users")
	@doc("Return all users visible to system administrators, including disabled accounts.")
	@get
	@route("/users")
	listManagedUsers(): OkJson<ManagedUsersResponse> | UnauthorizedProblem | ForbiddenProblem | InternalServerProblem;

	@operationId("updateManagedUser")
	@summary("Update managed user")
	@doc("Disable or re-enable a user account and grant or revoke instance-level administrator access.")
	@patch
	@route("/users/{user_id}")
	updateManagedUser(...UserIdPathParam, @body body: UpdateManagedUserRequest):
		| OkJson<ManagedUserResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| ConflictProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("setManagedUserPassword")
	@summary("Set managed user password")
	@doc("Replace a user's password with a system administrator supplied value.")
	@post
	@route("/users/{user_id}/password")
	setManagedUserPassword(...UserIdPathParam, @body body: SetManagedUserPasswordRequest):
		| OkJson<ManagedUserResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("clearManagedUserPassword")
	@summary("Clear managed user password")
	@delete
	@route("/users/{user_id}/password")
	clearManagedUserPassword(...UserIdPathParam): OkJson<ManagedUserResponse> | UnauthorizedProblem | ForbiddenProblem | ConflictProblem | NotFoundProblem | InternalServerProblem;

	@operationId("exportAdminData")
	@summary("Export admin data")
	@doc("Export Netstamp application data as a JSON backup. The export includes sensitive account and credential material such as password hashes and stored secrets.")
	@get
	@route("/data-export")
	exportAdminData(): OkJson<AdminDataExport> | UnauthorizedProblem | ForbiddenProblem | InternalServerProblem;

	@operationId("importAdminData")
	@summary("Import admin data")
	@doc("Import a JSON backup produced by exportAdminData. Existing application data in the managed tables is replaced.")
	@post
	@route("/data-import")
	importAdminData(@body body: AdminDataExport): OkJson<AdminDataImportResponse> | BadRequestProblem | UnauthorizedProblem | ForbiddenProblem | UnprocessableEntityProblem | InternalServerProblem;

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
	updateAdminSettings(@body body: UpdateAdminSettingsRequest):
		| OkJson<AdminSettingsResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;
}
