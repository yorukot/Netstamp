package publicstatus

import (
	"context"
	"sort"
	"time"

	"github.com/yorukot/netstamp/internal/controller/application/pingquery"
	"github.com/yorukot/netstamp/internal/controller/application/tcpquery"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
)

const publicMaxDataPoints int32 = 600

func (s *Service) chartForElement(ctx context.Context, page domainpublic.Page, assignments []domainpublic.Assignment, chartRange domainpublic.ChartRange, now time.Time) *domainpublic.Chart {
	from := chartFrom(now, chartRange)
	var series []domainpublic.Series
	for _, assignment := range assignments {
		switch assignment.CheckType {
		case domaincheck.TypePing:
			series = append(series, s.pingChartSeries(ctx, page.ProjectID, assignment, assignment.CheckID, from, now)...)
		case domaincheck.TypeTCP:
			series = append(series, s.tcpChartSeries(ctx, page.ProjectID, assignment, assignment.CheckID, from, now)...)
		}
	}
	if len(series) == 0 {
		return nil
	}
	return &domainpublic.Chart{Range: chartRange, Series: series}
}

func chartFrom(now time.Time, chartRange domainpublic.ChartRange) time.Time {
	switch chartRange {
	case domainpublic.ChartRange30d:
		return now.Add(-30 * 24 * time.Hour)
	case domainpublic.ChartRange7d:
		return now.Add(-7 * 24 * time.Hour)
	default:
		return now.Add(-24 * time.Hour)
	}
}

type chartSeriesScope struct {
	projectID string
	probeID   string
	checkID   string
	checkName string
	from      time.Time
	to        time.Time
	probeName string
}

func (s *Service) pingChartSeries(ctx context.Context, projectID string, assignment domainpublic.Assignment, checkID string, from, to time.Time) []domainpublic.Series {
	if s.pings == nil {
		return nil
	}
	scope := newChartSeriesScope(projectID, assignment, checkID, from, to)
	reader := pingChartReader{repo: s.pings}
	return readChartSeries(ctx, scope, reader.count, reader.list, reader.series)
}

func (s *Service) tcpChartSeries(ctx context.Context, projectID string, assignment domainpublic.Assignment, checkID string, from, to time.Time) []domainpublic.Series {
	if s.tcps == nil {
		return nil
	}
	scope := newChartSeriesScope(projectID, assignment, checkID, from, to)
	reader := tcpChartReader{repo: s.tcps}
	return readChartSeries(ctx, scope, reader.count, reader.list, reader.series)
}

func newChartSeriesScope(projectID string, assignment domainpublic.Assignment, checkID string, from, to time.Time) chartSeriesScope {
	return chartSeriesScope{
		projectID: projectID,
		probeID:   assignment.ProbeID,
		checkID:   checkID,
		checkName: assignment.CheckName,
		from:      from,
		to:        to,
		probeName: assignment.ProbeName,
	}
}

func readChartSeries[D any](
	ctx context.Context,
	scope chartSeriesScope,
	count func(context.Context, chartSeriesScope) (int64, error),
	list func(context.Context, chartSeriesScope, int64) (map[string]D, error),
	series func(map[string]D, chartSeriesScope) []domainpublic.Series,
) []domainpublic.Series {
	rawPoints, err := count(ctx, scope)
	if err != nil {
		return nil
	}
	values, err := list(ctx, scope, rawPoints)
	if err != nil {
		return nil
	}
	return series(values, scope)
}

type pingChartReader struct {
	repo PingSeriesRepository
}

func (r pingChartReader) count(ctx context.Context, scope chartSeriesScope) (int64, error) {
	return r.repo.CountPingSeriesPoints(ctx, domainping.SeriesPointCountQuery{ProjectID: scope.projectID, ProbeID: scope.probeID, CheckID: scope.checkID, From: scope.from, To: scope.to})
}

func (r pingChartReader) list(ctx context.Context, scope chartSeriesScope, rawPoints int64) (map[string]domainping.SeriesData, error) {
	plan := pingquery.SelectReadPlan(rawPoints, scope.from, scope.to, publicMaxDataPoints)
	return r.repo.ListPingSeries(ctx, domainping.SeriesReadQuery{
		ProjectID:     scope.projectID,
		ProbeID:       scope.probeID,
		CheckID:       scope.checkID,
		From:          scope.from,
		To:            scope.to,
		Series:        []string{"latency_avg"},
		MaxDataPoints: publicMaxDataPoints,
		Mode:          plan.Mode,
	})
}

func (r pingChartReader) series(values map[string]domainping.SeriesData, scope chartSeriesScope) []domainpublic.Series {
	return pingDomainSeries(values, scope.probeName, scope.checkName, scope.checkID, map[string]string{
		"latency_avg": "ms",
	})
}

type tcpChartReader struct {
	repo TCPSeriesRepository
}

func (r tcpChartReader) count(ctx context.Context, scope chartSeriesScope) (int64, error) {
	return r.repo.CountTCPSeriesPoints(ctx, domaintcp.SeriesPointCountQuery{ProjectID: scope.projectID, ProbeID: scope.probeID, CheckID: scope.checkID, From: scope.from, To: scope.to})
}

func (r tcpChartReader) list(ctx context.Context, scope chartSeriesScope, rawPoints int64) (map[string]domaintcp.SeriesData, error) {
	plan := tcpquery.SelectReadPlan(rawPoints, scope.from, scope.to, publicMaxDataPoints)
	return r.repo.ListTCPSeries(ctx, domaintcp.SeriesReadQuery{
		ProjectID:     scope.projectID,
		ProbeID:       scope.probeID,
		CheckID:       scope.checkID,
		From:          scope.from,
		To:            scope.to,
		Series:        []string{"connect_avg"},
		MaxDataPoints: publicMaxDataPoints,
		Mode:          plan.Mode,
	})
}

func (r tcpChartReader) series(values map[string]domaintcp.SeriesData, scope chartSeriesScope) []domainpublic.Series {
	return tcpDomainSeries(values, scope.probeName, scope.checkName, scope.checkID, map[string]string{
		"connect_avg": "ms",
	})
}

func pingDomainSeries(values map[string]domainping.SeriesData, probeName, checkName, checkID string, units map[string]string) []domainpublic.Series {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	series := make([]domainpublic.Series, 0, len(keys))
	for _, key := range keys {
		series = append(series, domainpublic.Series{
			Name: key,
			Labels: map[string]string{
				"checkId":   checkID,
				"checkName": checkName,
				"checkType": "ping",
				"probeName": probeName,
			},
			Unit:   units[key],
			Points: pingSeriesPoints(values[key].Points),
		})
	}
	return series
}

func tcpDomainSeries(values map[string]domaintcp.SeriesData, probeName, checkName, checkID string, units map[string]string) []domainpublic.Series {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	series := make([]domainpublic.Series, 0, len(keys))
	for _, key := range keys {
		series = append(series, domainpublic.Series{
			Name: key,
			Labels: map[string]string{
				"checkId":   checkID,
				"checkName": checkName,
				"checkType": "tcp",
				"probeName": probeName,
			},
			Unit:   units[key],
			Points: tcpSeriesPoints(values[key].Points),
		})
	}
	return series
}

func pingSeriesPoints(points []domainping.SeriesPoint) []domainpublic.SeriesPoint {
	values := make([]domainpublic.SeriesPoint, 0, len(points))
	for _, point := range points {
		values = append(values, domainpublic.SeriesPoint{TimestampMs: point.Timestamp.UTC().UnixMilli(), Value: point.Value})
	}
	return values
}

func tcpSeriesPoints(points []domaintcp.SeriesPoint) []domainpublic.SeriesPoint {
	values := make([]domainpublic.SeriesPoint, 0, len(points))
	for _, point := range points {
		values = append(values, domainpublic.SeriesPoint{TimestampMs: point.Timestamp.UTC().UnixMilli(), Value: point.Value})
	}
	return values
}
