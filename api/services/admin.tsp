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

	@operationId("getAdminAccessSettings")
	@summary("Get instance access settings")
	@doc("Return the current account, project, and credential-management policies.")
	@get
	@route("/settings/access")
	getAdminAccessSettings(): PrivateNoStoreOkJson<AdminAccessSettingsResponse> | UnauthorizedProblem | ForbiddenProblem | InternalServerProblem;

	@operationId("updateAdminAccessSettings")
	@summary("Update instance access settings")
	@doc("Partially update account, project, and credential-management policies.")
	@patch
	@route("/settings/access")
	updateAdminAccessSettings(@body body: AdminAccessSettingsPatch):
		| PrivateNoStoreOkJson<AdminAccessSettingsResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("getAdminSMTPSettings")
	@summary("Get SMTP settings")
	@doc("Return the current SMTP configuration. The password is never returned.")
	@get
	@route("/settings/smtp")
	getAdminSMTPSettings(): PrivateNoStoreOkJson<AdminSMTPSettingsResponse> | UnauthorizedProblem | ForbiddenProblem | InternalServerProblem;

	@operationId("updateAdminSMTPSettings")
	@summary("Update SMTP settings")
	@doc("Partially update the SMTP configuration. Omit password to retain it or set it to null to clear it.")
	@patch
	@route("/settings/smtp")
	updateAdminSMTPSettings(@body body: AdminSMTPSettingsPatch):
		| PrivateNoStoreOkJson<AdminSMTPSettingsResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("testAdminSMTP")
	@summary("Send an SMTP test email")
	@doc("Send a test email using the persisted SMTP configuration.")
	@post
	@route("/settings/smtp/test")
	testAdminSMTP(): NoContent | BadRequestProblem | UnauthorizedProblem | ForbiddenProblem | UnprocessableEntityProblem | ServiceUnavailableProblem | InternalServerProblem;

	@operationId("getAdminOIDCProviderSettings")
	@summary("Get OIDC provider settings")
	@doc("Return the current generic OIDC provider configuration. The client secret is never returned.")
	@get
	@route("/settings/authentication-providers/oidc")
	getAdminOIDCProviderSettings(): PrivateNoStoreOkJson<AdminOIDCProviderSettingsResponse> | UnauthorizedProblem | ForbiddenProblem | InternalServerProblem;

	@operationId("updateAdminOIDCProviderSettings")
	@summary("Update OIDC provider settings")
	@doc("Partially update the generic OIDC provider. Enabling the provider also validates discovery readiness.")
	@patch
	@route("/settings/authentication-providers/oidc")
	updateAdminOIDCProviderSettings(@body body: AdminOIDCProviderSettingsPatch):
		| PrivateNoStoreOkJson<AdminOIDCProviderSettingsResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| UnprocessableEntityProblem
		| ServiceUnavailableProblem
		| InternalServerProblem;

	@operationId("validateAdminOIDCProviderSettings")
	@summary("Validate OIDC provider settings")
	@doc("Validate a candidate OIDC patch against the persisted provider without saving it. This checks configuration and discovery readiness, not an end-user login.")
	@post
	@route("/settings/authentication-providers/oidc/validate")
	validateAdminOIDCProviderSettings(@body body: AdminOIDCProviderSettingsPatch):
		| NoContent
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| UnprocessableEntityProblem
		| ServiceUnavailableProblem
		| InternalServerProblem;

	@operationId("getAdminGoogleProviderSettings")
	@summary("Get Google provider settings")
	@doc("Return the current Google provider configuration. The client secret is never returned.")
	@get
	@route("/settings/authentication-providers/google")
	getAdminGoogleProviderSettings(): PrivateNoStoreOkJson<AdminGoogleProviderSettingsResponse> | UnauthorizedProblem | ForbiddenProblem | InternalServerProblem;

	@operationId("updateAdminGoogleProviderSettings")
	@summary("Update Google provider settings")
	@doc("Partially update the Google provider.")
	@patch
	@route("/settings/authentication-providers/google")
	updateAdminGoogleProviderSettings(@body body: AdminGoogleProviderSettingsPatch):
		| PrivateNoStoreOkJson<AdminGoogleProviderSettingsResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| UnprocessableEntityProblem
		| ServiceUnavailableProblem
		| InternalServerProblem;

	@operationId("validateAdminGoogleProviderSettings")
	@summary("Validate Google provider settings")
	@doc("Validate a candidate Google patch against the persisted provider without saving it. This checks configuration readiness but cannot prove the client secret without completing OAuth authorization.")
	@post
	@route("/settings/authentication-providers/google/validate")
	validateAdminGoogleProviderSettings(@body body: AdminGoogleProviderSettingsPatch):
		| NoContent
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| UnprocessableEntityProblem
		| ServiceUnavailableProblem
		| InternalServerProblem;

	@operationId("getAdminGitHubProviderSettings")
	@summary("Get GitHub provider settings")
	@doc("Return the current GitHub provider configuration. The client secret is never returned.")
	@get
	@route("/settings/authentication-providers/github")
	getAdminGitHubProviderSettings(): PrivateNoStoreOkJson<AdminGitHubProviderSettingsResponse> | UnauthorizedProblem | ForbiddenProblem | InternalServerProblem;

	@operationId("updateAdminGitHubProviderSettings")
	@summary("Update GitHub provider settings")
	@doc("Partially update the GitHub provider.")
	@patch
	@route("/settings/authentication-providers/github")
	updateAdminGitHubProviderSettings(@body body: AdminGitHubProviderSettingsPatch):
		| PrivateNoStoreOkJson<AdminGitHubProviderSettingsResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| UnprocessableEntityProblem
		| ServiceUnavailableProblem
		| InternalServerProblem;

	@operationId("validateAdminGitHubProviderSettings")
	@summary("Validate GitHub provider settings")
	@doc("Validate a candidate GitHub patch against the persisted provider without saving it. This checks configuration readiness but cannot prove the client secret without completing OAuth authorization.")
	@post
	@route("/settings/authentication-providers/github/validate")
	validateAdminGitHubProviderSettings(@body body: AdminGitHubProviderSettingsPatch):
		| NoContent
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| UnprocessableEntityProblem
		| ServiceUnavailableProblem
		| InternalServerProblem;
}
