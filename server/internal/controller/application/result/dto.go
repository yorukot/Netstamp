package result

import (
	apphttp "github.com/yorukot/netstamp/internal/controller/application/result/httpcheck"
	"github.com/yorukot/netstamp/internal/controller/application/result/latest"
	"github.com/yorukot/netstamp/internal/controller/application/result/ping"
	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
	"github.com/yorukot/netstamp/internal/controller/application/result/tcp"
	"github.com/yorukot/netstamp/internal/controller/application/result/traceroute"
)

type (
	PingSeriesKey = ping.SeriesKey
	TCPSeriesKey  = tcp.SeriesKey
)

const (
	PingSeriesLatencyAvg  = ping.SeriesLatencyAvg
	PingSeriesLatencyMin  = ping.SeriesLatencyMin
	PingSeriesLatencyMax  = ping.SeriesLatencyMax
	PingSeriesLossPercent = ping.SeriesLossPercent

	TCPSeriesConnectAvg     = tcp.SeriesConnectAvg
	TCPSeriesConnectMin     = tcp.SeriesConnectMin
	TCPSeriesConnectMax     = tcp.SeriesConnectMax
	TCPSeriesFailurePercent = tcp.SeriesFailurePercent
)

type (
	QueryPingSeriesInput         = ping.QuerySeriesInput
	QueryPingInsightInput        = ping.QueryInsightInput
	QueryTCPSeriesInput          = tcp.QuerySeriesInput
	QueryTCPInsightInput         = tcp.QueryInsightInput
	QueryHTTPSeriesInput         = apphttp.QuerySeriesInput
	QueryHTTPInsightInput        = apphttp.QueryInsightInput
	QueryLatestHTTPResultsInput  = apphttp.QueryLatestInput
	QueryTracerouteRunsInput     = traceroute.QueryRunsInput
	QueryTracerouteInsightInput  = traceroute.QueryInsightInput
	QueryTracerouteTopologyInput = traceroute.QueryTopologyInput
	QueryLatestResultsInput      = latest.QueryInput
)

type (
	PingSeriesOutput         = ping.SeriesOutput
	PingInsightOutput        = ping.InsightOutput
	TCPSeriesOutput          = tcp.SeriesOutput
	TCPInsightOutput         = tcp.InsightOutput
	HTTPSeriesOutput         = apphttp.SeriesOutput
	HTTPInsightOutput        = apphttp.InsightOutput
	LatestHTTPResultsOutput  = apphttp.LatestResultsOutput
	TracerouteRunsOutput     = traceroute.RunsOutput
	TracerouteInsightOutput  = traceroute.InsightOutput
	TracerouteTopologyOutput = traceroute.TopologyOutput
	LatestResultsOutput      = latest.Output
)

type (
	LatestResult                    = latest.Result
	TracerouteRun                   = traceroute.Run
	TracerouteHop                   = traceroute.Hop
	TracerouteInsightPoint          = traceroute.InsightPoint
	TracerouteTopologyNode          = traceroute.TopologyNode
	TracerouteTopologyEdge          = traceroute.TopologyEdge
	PingSeries                      = ping.Series
	PingSeriesLabels                = ping.SeriesLabels
	PingSeriesPoint                 = ping.SeriesPoint
	TCPSeries                       = tcp.Series
	TCPSeriesLabels                 = tcp.SeriesLabels
	TCPSeriesPoint                  = tcp.SeriesPoint
	Series                          = ping.Series
	SeriesLabels                    = ping.SeriesLabels
	SeriesPoint                     = ping.SeriesPoint
	PingInsightSummary              = ping.InsightSummary
	TCPInsightSummary               = tcp.InsightSummary
	HTTPSeries                      = apphttp.Series
	HTTPSeriesPoint                 = apphttp.SeriesPoint
	HTTPInsightSummary              = apphttp.InsightSummary
	LatestHTTPResult                = apphttp.LatestResult
	QueryMetadata                   = resultshared.QueryMetadata
	TracerouteRunsQueryMetadata     = traceroute.RunsQueryMetadata
	TracerouteInsightQueryMetadata  = traceroute.InsightQueryMetadata
	TracerouteTopologyQueryMetadata = traceroute.TopologyQueryMetadata
)
