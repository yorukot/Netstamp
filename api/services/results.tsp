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
	series?: string = "latency_avg,latency_min,latency_max,loss_percent";

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

model TcpSeriesQuery {
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
	series?: string = "connect_avg,connect_min,connect_max,failure_percent";

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

model HttpSeriesQuery {
	@query probeId: uuid;
	@query checkId: uuid;
	@query @minValue(1) from?: int64;
	@query @minValue(1) to?: int64;
	@query series?: string = "dns_avg,connect_avg,tls_avg,ttfb_avg,total_avg,failure_percent";

	@query
	@minValue(1)
	@maxValue(2000)
	maxDataPoints?: int32 = 600;
}

model HttpInsightQuery {
	@query probeId: uuid;
	@query checkId: uuid;
	@query @minValue(1) from?: int64;
	@query @minValue(1) to?: int64;

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

	@operationId("queryProjectLatestResults")
	@summary("Query project latest results")
	@doc("List latest all-time result timestamps and statuses for active assignment pairs. Chart series and detailed result rows remain under type-specific result endpoints.")
	@get
	@route("/latest")
	queryProjectLatestResults(...ProjectRefParam, ...LatestResultsQuery):
		| OkJson<LatestResultsResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("queryProjectTcpResultSeries")
	@summary("Query project TCP result series")
	@get
	@route("/tcp/series")
	queryProjectTcpResultSeries(...ProjectRefParam, ...TcpSeriesQuery):
		| OkJson<TcpSeriesResponse>
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

	@operationId("queryProjectHttpResultSeries")
	@summary("Query project HTTP result series")
	@get
	@route("/http/series")
	queryProjectHttpResultSeries(...ProjectRefParam, ...HttpSeriesQuery):
		| OkJson<HttpSeriesResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("queryProjectHttpResultInsight")
	@summary("Query project HTTP result insight")
	@get
	@route("/http/insight")
	queryProjectHttpResultInsight(...ProjectRefParam, ...HttpInsightQuery):
		| OkJson<HttpInsightResponse>
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
