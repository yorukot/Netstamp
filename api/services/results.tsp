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
	@minValue(1)
	from?: int64;

	@query
	@minValue(1)
	to?: int64;

	@query
	metric?: PingMetric = "rttAvgMs";

	@query
	@minValue(1)
	@maxValue(2000)
	maxDataPoints?: int32 = 600;
}

model PingInsightQuery {
	@query
	probeId: uuid;

	@query
	checkId: uuid;

	@query
	@minValue(1)
	from?: int64;

	@query
	@minValue(1)
	to?: int64;

	@query
	@minValue(1)
	@maxValue(2000)
	maxDataPoints?: int32 = 600;
}

model TcpInsightQuery {
	@query
	probeId: uuid;

	@query
	checkId: uuid;

	@query
	@minValue(1)
	from?: int64;

	@query
	@minValue(1)
	to?: int64;

	@query
	@minValue(1)
	@maxValue(2000)
	maxDataPoints?: int32 = 600;
}

model TracerouteRunsQuery {
	@query
	probeId: uuid;

	@query
	checkId: uuid;

	@query
	@minValue(1)
	from?: int64;

	@query
	@minValue(1)
	to?: int64;

	@query
	@minValue(1)
	@maxValue(500)
	limit?: int32 = 100;

	@query
	@minValue(1)
	cursor?: int64;
}

model TracerouteInsightQuery {
	@query
	probeId: uuid;

	@query
	checkId: uuid;

	@query
	@minValue(1)
	from?: int64;

	@query
	@minValue(1)
	to?: int64;

	@query
	@minValue(1)
	@maxValue(2000)
	maxDataPoints?: int32 = 600;
}

model TracerouteTopologyQuery {
	@query
	probeId?: uuid;

	@query
	checkId?: uuid;

	@query
	@minValue(1)
	from?: int64;

	@query
	@minValue(1)
	to?: int64;

	@query
	@minValue(1)
	@maxValue(500)
	limit?: int32 = 100;
}

@tag("Results")
@useAuth(SessionCookieAuth)
@route("/projects/{ref}/results")
interface ResultService {
	@operationId("queryProjectPingResultSeries")
	@summary("Query project ping result series")
	@get
	@route("/ping/series")
	queryProjectPingResultSeries(...ProjectRefParam, ...PingSeriesQuery):
		| OkJson<PingSeriesResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("queryProjectPingResultInsight")
	@summary("Query project ping result insight")
	@get
	@route("/ping/insight")
	queryProjectPingResultInsight(...ProjectRefParam, ...PingInsightQuery):
		| OkJson<PingInsightResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("queryProjectTcpResultInsight")
	@summary("Query project TCP result insight")
	@get
	@route("/tcp/insight")
	queryProjectTcpResultInsight(...ProjectRefParam, ...TcpInsightQuery):
		| OkJson<TcpInsightResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("queryProjectTracerouteResultRuns")
	@summary("Query project traceroute result runs")
	@get
	@route("/traceroute/runs")
	queryProjectTracerouteResultRuns(...ProjectRefParam, ...TracerouteRunsQuery):
		| OkJson<TracerouteRunsResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("queryProjectTracerouteResultInsight")
	@summary("Query project traceroute result insight")
	@get
	@route("/traceroute/insight")
	queryProjectTracerouteResultInsight(...ProjectRefParam, ...TracerouteInsightQuery):
		| OkJson<TracerouteInsightResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("queryProjectTracerouteTopology")
	@summary("Query project traceroute topology")
	@get
	@route("/traceroute/topology")
	queryProjectTracerouteTopology(...ProjectRefParam, ...TracerouteTopologyQuery):
		| OkJson<TracerouteTopologyResponse>
		| BadRequestProblem
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
	listProjectMeasurements(...ProjectRefParam, ...MeasurementQuery):
		| OkJson<MeasurementListResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;
}
