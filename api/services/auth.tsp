import "@typespec/http";
import "@typespec/openapi";

using Http;
using OpenAPI;

namespace Netstamp;

@tag("Auth")
@route("/auth")
interface AuthService {
	@operationId("registerUser")
	@summary("Register user")
	@doc("Create a user account with a normalized email address, display name, and password. On success, sets an HTTP-only session cookie for immediate API access.")
	@post
	@route("/register")
	registerUser(@body body: RegisterUserRequest): SessionCreatedJson<AuthUserResponse> | BadRequestProblem | UnprocessableEntityProblem | ConflictProblem | InternalServerProblem;

	@operationId("loginUser")
	@summary("Login user")
	@doc("Verify an email and password pair, then set an HTTP-only session cookie for the authenticated user. Invalid credentials always return the same unauthorized response.")
	@post
	@route("/login")
	loginUser(@body body: LoginUserRequest): SessionOkJson<AuthUserResponse> | BadRequestProblem | UnauthorizedProblem | InternalServerProblem;

	@operationId("logoutUser")
	@summary("Logout user")
	@doc("Clear the HTTP-only session cookie.")
	@post
	@route("/logout")
	logoutUser(): SessionNoContent;

	@operationId("getCurrentUser")
	@summary("Get current user")
	@doc("Return the user identity embedded in a valid session cookie.")
	@useAuth(SessionCookieAuth)
	@get
	@route("/me")
	getCurrentUser(): OkJson<CurrentUserResponse> | UnauthorizedProblem | InternalServerProblem;
}
