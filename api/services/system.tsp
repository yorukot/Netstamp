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

@example(#{ accountCreationEnabled: true, projectCreationEnabled: true, credentialChangesEnabled: true })
model PublicCapabilities {
	@doc("Whether new accounts may be created through password registration or external-provider JIT provisioning.")
	accountCreationEnabled: boolean;

	projectCreationEnabled: boolean;
	credentialChangesEnabled: boolean;
}

@example(#{ demoMode: false, capabilities: #{ accountCreationEnabled: true, projectCreationEnabled: true, credentialChangesEnabled: true } })
model PublicRuntimeConfig {
	demoMode: boolean;
	capabilities: PublicCapabilities;
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
