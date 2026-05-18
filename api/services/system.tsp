import "@typespec/http";
import "@typespec/openapi";

using Http;
using OpenAPI;

namespace Netstamp;

@example(#{ message: "Netstamp API is running" })
model APIStatusResponse {
	message: "Netstamp API is running";
}

@example(#{ status: "ok" })
model HealthResponse {
	status: "ok";
}

@tag("System")
interface SystemService {
	@operationId("getAPIStatus")
	@summary("Get API status")
	@get
	@route("/")
	getAPIStatus(): OkJson<APIStatusResponse>;

	@operationId("getHealth")
	@summary("Get health status")
	@get
	@route("/healthz")
	getHealth(): OkJson<HealthResponse> | ServiceUnavailableProblem;
}
