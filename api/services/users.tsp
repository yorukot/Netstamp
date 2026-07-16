import "@typespec/http";
import "@typespec/openapi";

using Http;
using OpenAPI;

namespace Netstamp;

@tag("Users")
@useAuth(SessionCookieAuth)
@route("/users")
interface UserService {
	@operationId("updateCurrentUser")
	@summary("Update current user")
	@doc("Update profile fields for the user authenticated by the session cookie.")
	@patch
	@route("/me")
	updateCurrentUser(@body body: UpdateCurrentUserRequest): OkJson<AuthUserResponse> | BadRequestProblem | UnauthorizedProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("changeCurrentUserEmail")
	@summary("Change current user email")
	@doc("Change the sign-in email after recent authentication.")
	@post
	@route("/me/email-change")
	changeCurrentUserEmail(@body body: ChangeCurrentUserEmailRequest):
		| OkJson<AuthUserResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| ConflictProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("setCurrentUserPassword")
	@summary("Set current user password")
	@doc("Create or replace the current user's local password after recent authentication. A passwordless GitHub account may set its first password from the same GitHub-authenticated session during the external authorization flow window; this exception does not grant sudo for other operations.")
	@put
	@route("/me/password")
	changeCurrentUserPassword(@body body: ChangeCurrentUserPasswordRequest): NoContent | BadRequestProblem | UnauthorizedProblem | ForbiddenProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("removeCurrentUserPassword")
	@summary("Remove current user password")
	@doc("Remove the local password when another authentication identity remains.")
	@delete
	@route("/me/password")
	removeCurrentUserPassword(): NoContent | UnauthorizedProblem | ForbiddenProblem | ConflictProblem | InternalServerProblem;

	@operationId("listCurrentUserAuthenticationMethods")
	@summary("List current user authentication methods")
	@get
	@route("/me/authentication-methods")
	listCurrentUserAuthenticationMethods(): OkJson<AuthenticationMethodsResponse> | UnauthorizedProblem | InternalServerProblem;

	@operationId("removeCurrentUserIdentity")
	@summary("Remove current user identity")
	@delete
	@route("/me/identities/{identity_id}")
	removeCurrentUserIdentity(...IdentityIdPathParam): NoContent | UnauthorizedProblem | ForbiddenProblem | NotFoundProblem | ConflictProblem | InternalServerProblem;

	@operationId("deactivateCurrentUser")
	@summary("Deactivate current user")
	@doc("Disable the authenticated user's account. Disabled accounts cannot sign in or access protected routes until a system administrator re-enables them.")
	@post
	@route("/me/deactivation")
	deactivateCurrentUser(): NoContent | UnauthorizedProblem | ForbiddenProblem | ConflictProblem | InternalServerProblem;
}
