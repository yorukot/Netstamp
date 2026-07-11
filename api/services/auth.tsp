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
	@doc("Create a user account with a normalized email address, display name, and password. When email verification is disabled, sets an HTTP-only session cookie for immediate API access. When email verification is required, sends a verification email and returns 202 without a session cookie.")
	@post
	@route("/register")
	registerUser(@body body: RegisterUserRequest):
		| SessionCreatedJson<AuthUserResponse>
		| AcceptedJson<RegisterEmailVerificationRequiredResponse>
		| BadRequestProblem
		| UnprocessableEntityProblem
		| ConflictProblem
		| ServiceUnavailableProblem
		| InternalServerProblem;

	@operationId("loginUser")
	@summary("Login user")
	@doc("Verify an email and password pair, then set an HTTP-only session cookie for the authenticated user. Invalid credentials always return the same unauthorized response. Unverified accounts return forbidden when email verification is required.")
	@post
	@route("/login")
	loginUser(@body body: LoginUserRequest): SessionOkJson<AuthUserResponse> | BadRequestProblem | UnauthorizedProblem | ForbiddenProblem | InternalServerProblem;

	@operationId("createPasswordReset")
	@summary("Create password reset")
	@doc("Create a password reset request for an account email. The response does not reveal whether the email exists.")
	@post
	@route("/password-resets")
	createPasswordReset(@body body: CreatePasswordResetRequest): Accepted | BadRequestProblem | UnprocessableEntityProblem | TooManyRequestsProblem | ServiceUnavailableProblem | InternalServerProblem;

	@operationId("confirmPasswordReset")
	@summary("Confirm password reset")
	@doc("Consume a one-time password reset token and replace the account password.")
	@patch
	@route("/password-resets")
	confirmPasswordReset(@body body: ConfirmPasswordResetRequest): NoContent | BadRequestProblem | UnauthorizedProblem | UnprocessableEntityProblem | ServiceUnavailableProblem | InternalServerProblem;

	@operationId("createEmailVerification")
	@summary("Create email verification")
	@doc("Create or resend an email verification link. The response does not reveal whether the email exists.")
	@post
	@route("/email-verifications")
	createEmailVerification(@body body: CreateEmailVerificationRequest):
		| Accepted
		| BadRequestProblem
		| UnprocessableEntityProblem
		| TooManyRequestsProblem
		| ServiceUnavailableProblem
		| InternalServerProblem;

	@operationId("confirmEmailVerification")
	@summary("Confirm email verification")
	@doc("Consume a one-time email verification token and mark the account email as verified.")
	@patch
	@route("/email-verifications")
	confirmEmailVerification(@body body: ConfirmEmailVerificationRequest):
		| NoContent
		| BadRequestProblem
		| UnauthorizedProblem
		| UnprocessableEntityProblem
		| ServiceUnavailableProblem
		| InternalServerProblem;

	@operationId("logoutUser")
	@summary("Logout user")
	@doc("Revoke the server-side session and clear the HTTP-only session cookie.")
	@post
	@route("/logout")
	logoutUser(): SessionNoContent;

	@operationId("getCSRFToken")
	@summary("Get CSRF token")
	@doc("Rotate and return the synchronizer CSRF token for the current authenticated session.")
	@useAuth(SessionCookieAuth)
	@get
	@route("/csrf")
	getCSRFToken(): OkJson<CSRFTokenResponse> | UnauthorizedProblem | InternalServerProblem;

	@operationId("getCurrentUser")
	@summary("Get current user")
	@doc("Return the user identity associated with a valid session cookie.")
	@useAuth(SessionCookieAuth)
	@get
	@route("/me")
	getCurrentUser(): OkJson<CurrentUserResponse> | UnauthorizedProblem | InternalServerProblem;

	@operationId("listAuthSessions")
	@summary("List auth sessions")
	@doc("Return the active server-side sessions owned by the current user, including the captured User-Agent and whether each session is current.")
	@useAuth(SessionCookieAuth)
	@get
	@route("/sessions")
	listAuthSessions(): OkJson<AuthSessionsResponse> | UnauthorizedProblem | InternalServerProblem;

	@operationId("revokeAuthSession")
	@summary("Revoke auth session")
	@doc("Revoke an active server-side session owned by the current user. Revoking the current session also clears its session cookie.")
	@useAuth(SessionCookieAuth)
	@delete
	@route("/sessions/{session_id}")
	revokeAuthSession(...AuthSessionIdPathParam): OptionalSessionNoContent | UnauthorizedProblem | NotFoundProblem | InternalServerProblem;
}
