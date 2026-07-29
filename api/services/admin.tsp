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

	@operationId("getAdminAccessSettings")
	@summary("Get instance access settings")
	@doc("Return the current account, project, and credential-management policies.")
	@get
	@route("/settings/access")
	getAdminAccessSettings(): VersionedOkJson<AdminAccessSettingsResponse> | UnauthorizedProblem | ForbiddenProblem | InternalServerProblem;

	@operationId("updateAdminAccessSettings")
	@summary("Update instance access settings")
	@doc("Partially update account, project, and credential-management policies. If-Match must contain the strong ETag returned by getAdminAccessSettings.")
	@patch
	@route("/settings/access")
	updateAdminAccessSettings(...IfMatchHeaderParam, @body body: AdminAccessSettingsPatch):
		| VersionedOkJson<AdminAccessSettingsResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| PreconditionFailedProblem
		| PreconditionRequiredProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("getAdminSMTPSettings")
	@summary("Get SMTP settings")
	@doc("Return the current SMTP configuration. The password is never returned.")
	@get
	@route("/settings/smtp")
	getAdminSMTPSettings(): VersionedOkJson<AdminSMTPSettingsResponse> | UnauthorizedProblem | ForbiddenProblem | InternalServerProblem;

	@operationId("updateAdminSMTPSettings")
	@summary("Update SMTP settings")
	@doc("Partially update the SMTP configuration. Omit password to retain it or set it to null to clear it. If-Match must contain the strong ETag returned by getAdminSMTPSettings.")
	@patch
	@route("/settings/smtp")
	updateAdminSMTPSettings(...IfMatchHeaderParam, @body body: AdminSMTPSettingsPatch):
		| VersionedOkJson<AdminSMTPSettingsResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| PreconditionFailedProblem
		| PreconditionRequiredProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("testAdminSMTP")
	@summary("Send an SMTP test email")
	@doc("Send a test email using the persisted SMTP configuration. If-Match must contain the strong ETag returned by getAdminSMTPSettings.")
	@post
	@route("/settings/smtp/test")
	testAdminSMTP(...IfMatchHeaderParam):
		| NoContent
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| PreconditionFailedProblem
		| PreconditionRequiredProblem
		| UnprocessableEntityProblem
		| ServiceUnavailableProblem
		| InternalServerProblem;

	@operationId("getAdminOIDCProviderSettings")
	@summary("Get OIDC provider settings")
	@doc("Return the current generic OIDC provider configuration. The client secret is never returned.")
	@get
	@route("/settings/authentication-providers/oidc")
	getAdminOIDCProviderSettings(): VersionedOkJson<AdminOIDCProviderSettingsResponse> | UnauthorizedProblem | ForbiddenProblem | InternalServerProblem;

	@operationId("updateAdminOIDCProviderSettings")
	@summary("Update OIDC provider settings")
	@doc("Partially update the generic OIDC provider. Enabling the provider also validates discovery readiness. If-Match must contain the strong ETag returned by getAdminOIDCProviderSettings.")
	@patch
	@route("/settings/authentication-providers/oidc")
	updateAdminOIDCProviderSettings(...IfMatchHeaderParam, @body body: AdminOIDCProviderSettingsPatch):
		| VersionedOkJson<AdminOIDCProviderSettingsResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| PreconditionFailedProblem
		| PreconditionRequiredProblem
		| UnprocessableEntityProblem
		| ServiceUnavailableProblem
		| InternalServerProblem;

	@operationId("validateAdminOIDCProviderSettings")
	@summary("Validate OIDC provider settings")
	@doc("Validate a candidate OIDC patch against the persisted provider without saving it. This checks configuration and discovery readiness, not an end-user login.")
	@post
	@route("/settings/authentication-providers/oidc/validate")
	validateAdminOIDCProviderSettings(...IfMatchHeaderParam, @body body: AdminOIDCProviderSettingsPatch):
		| NoContent
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| PreconditionFailedProblem
		| PreconditionRequiredProblem
		| UnprocessableEntityProblem
		| ServiceUnavailableProblem
		| InternalServerProblem;

	@operationId("getAdminGoogleProviderSettings")
	@summary("Get Google provider settings")
	@doc("Return the current Google provider configuration. The client secret is never returned.")
	@get
	@route("/settings/authentication-providers/google")
	getAdminGoogleProviderSettings(): VersionedOkJson<AdminGoogleProviderSettingsResponse> | UnauthorizedProblem | ForbiddenProblem | InternalServerProblem;

	@operationId("updateAdminGoogleProviderSettings")
	@summary("Update Google provider settings")
	@doc("Partially update the Google provider. If-Match must contain the strong ETag returned by getAdminGoogleProviderSettings.")
	@patch
	@route("/settings/authentication-providers/google")
	updateAdminGoogleProviderSettings(...IfMatchHeaderParam, @body body: AdminGoogleProviderSettingsPatch):
		| VersionedOkJson<AdminGoogleProviderSettingsResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| PreconditionFailedProblem
		| PreconditionRequiredProblem
		| UnprocessableEntityProblem
		| ServiceUnavailableProblem
		| InternalServerProblem;

	@operationId("validateAdminGoogleProviderSettings")
	@summary("Validate Google provider settings")
	@doc("Validate a candidate Google patch against the persisted provider without saving it. This checks configuration readiness but cannot prove the client secret without completing OAuth authorization.")
	@post
	@route("/settings/authentication-providers/google/validate")
	validateAdminGoogleProviderSettings(...IfMatchHeaderParam, @body body: AdminGoogleProviderSettingsPatch):
		| NoContent
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| PreconditionFailedProblem
		| PreconditionRequiredProblem
		| UnprocessableEntityProblem
		| ServiceUnavailableProblem
		| InternalServerProblem;

	@operationId("getAdminGitHubProviderSettings")
	@summary("Get GitHub provider settings")
	@doc("Return the current GitHub provider configuration. The client secret is never returned.")
	@get
	@route("/settings/authentication-providers/github")
	getAdminGitHubProviderSettings(): VersionedOkJson<AdminGitHubProviderSettingsResponse> | UnauthorizedProblem | ForbiddenProblem | InternalServerProblem;

	@operationId("updateAdminGitHubProviderSettings")
	@summary("Update GitHub provider settings")
	@doc("Partially update the GitHub provider. If-Match must contain the strong ETag returned by getAdminGitHubProviderSettings.")
	@patch
	@route("/settings/authentication-providers/github")
	updateAdminGitHubProviderSettings(...IfMatchHeaderParam, @body body: AdminGitHubProviderSettingsPatch):
		| VersionedOkJson<AdminGitHubProviderSettingsResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| PreconditionFailedProblem
		| PreconditionRequiredProblem
		| UnprocessableEntityProblem
		| ServiceUnavailableProblem
		| InternalServerProblem;

	@operationId("validateAdminGitHubProviderSettings")
	@summary("Validate GitHub provider settings")
	@doc("Validate a candidate GitHub patch against the persisted provider without saving it. This checks configuration readiness but cannot prove the client secret without completing OAuth authorization.")
	@post
	@route("/settings/authentication-providers/github/validate")
	validateAdminGitHubProviderSettings(...IfMatchHeaderParam, @body body: AdminGitHubProviderSettingsPatch):
		| NoContent
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| PreconditionFailedProblem
		| PreconditionRequiredProblem
		| UnprocessableEntityProblem
		| ServiceUnavailableProblem
		| InternalServerProblem;
}
