import "@typespec/http";
import "@typespec/openapi";

using Http;
using OpenAPI;

namespace Netstamp;

@tag("API Tokens")
@useAuth(SessionCookieAuth)
@route("/auth/api-tokens")
interface APITokenService {
	@operationId("listAPITokens")
	@summary("List personal API tokens")
	@get
	listAPITokens(): OkJson<APITokenListResponse> | UnauthorizedProblem | InternalServerProblem;

	@operationId("createAPIToken")
	@summary("Create personal API token")
	@doc("Create a scoped personal access token. The plaintext value is returned once in this response.")
	@post
	createAPIToken(@body body: CreateAPITokenRequest):
		| CreatedJson<CreateAPITokenResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ConflictProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("revokeAPIToken")
	@summary("Revoke personal API token")
	@delete
	@route("/{token_id}")
	revokeAPIToken(...APITokenIdPathParam): NoContent | UnauthorizedProblem | NotFoundProblem | InternalServerProblem;
}
