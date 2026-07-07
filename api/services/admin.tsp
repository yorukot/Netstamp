import "@typespec/http";
import "@typespec/openapi";

using Http;
using OpenAPI;

namespace Netstamp;

@tag("Admin")
@useAuth(SessionCookieAuth)
@route("/admin")
interface AdminService {
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
