package result

import (
	"github.com/yorukot/netstamp/internal/controller/application/result/measurement"
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
	QueryTracerouteRunsInput     = traceroute.QueryRunsInput
	QueryTracerouteInsightInput  = traceroute.QueryInsightInput
	QueryTracerouteTopologyInput = traceroute.QueryTopologyInput
	QueryMeasurementsInput       = measurement.QueryInput
)

type (
	PingSeriesOutput         = ping.SeriesOutput
	PingInsightOutput        = ping.InsightOutput
	TCPSeriesOutput          = tcp.SeriesOutput
	TCPInsightOutput         = tcp.InsightOutput
	TracerouteRunsOutput     = traceroute.RunsOutput
	TracerouteInsightOutput  = traceroute.InsightOutput
	TracerouteTopologyOutput = traceroute.TopologyOutput
	MeasurementsOutput       = measurement.Output
)

type (
	Measurement                     = measurement.Measurement
	TracerouteRun                   = traceroute.Run
	TracerouteHop                   = traceroute.Hop
	TracerouteInsightPoint          = traceroute.InsightPoint
	TracerouteTopologyNode          = traceroute.TopologyNode
	TracerouteTopologyEdge          = traceroute.TopologyEdge
	Series                          = ping.Series
	SeriesLabels                    = ping.SeriesLabels
	SeriesPoint                     = ping.SeriesPoint
	PingInsightSummary              = ping.InsightSummary
	TCPInsightSummary               = tcp.InsightSummary
	QueryMetadata                   = resultshared.QueryMetadata
	TracerouteRunsQueryMetadata     = traceroute.RunsQueryMetadata
	TracerouteInsightQueryMetadata  = traceroute.InsightQueryMetadata
	TracerouteTopologyQueryMetadata = traceroute.TopologyQueryMetadata
	MeasurementQueryMetadata        = measurement.QueryMetadata
)
