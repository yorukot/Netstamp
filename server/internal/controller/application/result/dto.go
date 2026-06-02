package result

import (
	"github.com/yorukot/netstamp/internal/controller/application/result/measurement"
	"github.com/yorukot/netstamp/internal/controller/application/result/ping"
	resultshared "github.com/yorukot/netstamp/internal/controller/application/result/shared"
	"github.com/yorukot/netstamp/internal/controller/application/result/tcp"
	"github.com/yorukot/netstamp/internal/controller/application/result/traceroute"
)

type PingSeriesKey = ping.SeriesKey

const (
	PingSeriesLatencyAvg  = ping.SeriesLatencyAvg
	PingSeriesLatencyMin  = ping.SeriesLatencyMin
	PingSeriesLatencyMax  = ping.SeriesLatencyMax
	PingSeriesLossPercent = ping.SeriesLossPercent
)

type QueryPingSeriesInput = ping.QuerySeriesInput
type QueryPingInsightInput = ping.QueryInsightInput
type QueryTCPInsightInput = tcp.QueryInsightInput
type QueryTracerouteRunsInput = traceroute.QueryRunsInput
type QueryTracerouteInsightInput = traceroute.QueryInsightInput
type QueryTracerouteTopologyInput = traceroute.QueryTopologyInput
type QueryMeasurementsInput = measurement.QueryInput

type PingSeriesOutput = ping.SeriesOutput
type PingInsightOutput = ping.InsightOutput
type TCPInsightOutput = tcp.InsightOutput
type TracerouteRunsOutput = traceroute.RunsOutput
type TracerouteInsightOutput = traceroute.InsightOutput
type TracerouteTopologyOutput = traceroute.TopologyOutput
type MeasurementsOutput = measurement.Output

type Measurement = measurement.Measurement
type TracerouteRun = traceroute.Run
type TracerouteHop = traceroute.Hop
type TracerouteInsightPoint = traceroute.InsightPoint
type TracerouteTopologyNode = traceroute.TopologyNode
type TracerouteTopologyEdge = traceroute.TopologyEdge
type Series = ping.Series
type SeriesLabels = ping.SeriesLabels
type SeriesPoint = ping.SeriesPoint
type PingInsightSummary = ping.InsightSummary
type TCPInsightBucket = tcp.InsightBucket
type TCPInsightSummary = tcp.InsightSummary
type QueryMetadata = resultshared.QueryMetadata
type TracerouteRunsQueryMetadata = traceroute.RunsQueryMetadata
type TracerouteInsightQueryMetadata = traceroute.InsightQueryMetadata
type TracerouteTopologyQueryMetadata = traceroute.TopologyQueryMetadata
type MeasurementQueryMetadata = measurement.QueryMetadata
