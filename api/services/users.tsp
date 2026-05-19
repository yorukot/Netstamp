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
	updateCurrentUser(@body body: UpdateCurrentUserRequest): OkJson<AuthUserResponse> | UnauthorizedProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("changeCurrentUserEmail")
	@summary("Change current user email")
	@doc("Change the sign-in email for the current user after confirming the current password.")
	@post
	@route("/me/email-change")
	changeCurrentUserEmail(@body body: ChangeCurrentUserEmailRequest): OkJson<AuthUserResponse> | UnauthorizedProblem | ConflictProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("changeCurrentUserPassword")
	@summary("Change current user password")
	@doc("Change the current user's password after verifying the existing password.")
	@post
	@route("/me/password-change")
	changeCurrentUserPassword(@body body: ChangeCurrentUserPasswordRequest): NoContent | UnauthorizedProblem | UnprocessableEntityProblem | InternalServerProblem;
}
