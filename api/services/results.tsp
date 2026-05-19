import "@typespec/http";
import "@typespec/openapi";

using Http;
using OpenAPI;

namespace Netstamp;

model PingSeriesQuery {
	@query
	probeId: uuid;

	@query
	checkId: uuid;

	@query
	from?: int64;

	@query
	to?: int64;

	@query
	metric: PingMetric;

	@query
	maxDataPoints?: int32;
}

model TracerouteRunsQuery {
	@query
	probeId: uuid;

	@query
	checkId: uuid;

	@query
	from?: int64;

	@query
	to?: int64;

	@query
	limit?: int32;

	@query
	cursor?: int64;
}

@tag("Results")
@useAuth(SessionCookieAuth)
@route("/projects/{ref}/results")
interface ResultService {
	@operationId("queryProjectPingResultSeries")
	@summary("Query project ping result series")
	@get
	@route("/ping/series")
	queryProjectPingResultSeries(...ProjectRefParam, ...PingSeriesQuery): OkJson<PingSeriesResponse> | UnauthorizedProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("queryProjectTracerouteResultRuns")
	@summary("Query project traceroute result runs")
	@get
	@route("/traceroute/runs")
	queryProjectTracerouteResultRuns(...ProjectRefParam, ...TracerouteRunsQuery):
		| OkJson<TracerouteRunsResponse>
		| UnauthorizedProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;
}

@tag("Measurements")
@useAuth(SessionCookieAuth)
@route("/projects/{ref}/measurements")
interface MeasurementService {
	@operationId("listProjectMeasurements")
	@summary("List project measurements")
	@doc("List recent raw-ish measurement rows for dashboard activity, check logs, and measurement tables. Aggregated chart series remain under the result series endpoints.")
	@get
	listProjectMeasurements(...ProjectRefParam, ...MeasurementQuery): OkJson<MeasurementListResponse> | UnauthorizedProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;
}
