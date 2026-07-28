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

model PublicCapabilities {
	registrationEnabled: boolean;
	projectCreationEnabled: boolean;
	credentialChangesEnabled: boolean;
}

model PublicRuntimeConfig {
	demoMode: boolean;
	capabilities: PublicCapabilities;
	tracking: TrackingSettings;
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

	@operationId("getRuntimeConfig")
	@summary("Get public runtime configuration")
	@get
	@route("/system/config")
	getRuntimeConfig(): OkJson<PublicRuntimeConfig> | InternalServerProblem;
}
