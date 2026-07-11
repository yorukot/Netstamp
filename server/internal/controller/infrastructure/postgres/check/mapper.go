package pgcheck

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

func mapStoredPingCheck(row sqlc.Check, config sqlc.PingCheckConfig) domaincheck.Check {
	return domaincheck.Check{
		ID:              row.ID.String(),
		ProjectID:       row.ProjectID.String(),
		Name:            row.Name,
		Type:            domaincheck.Type(row.CheckType),
		Target:          row.Target,
		Selector:        cloneRawMessage(row.Selector),
		Description:     row.Description,
		IntervalSeconds: row.IntervalSeconds,
		PingConfig:      pingConfigPtr(mapPingConfig(config)),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		DeletedAt:       row.DeletedAt,
	}
}

func mapStoredTCPCheck(row sqlc.Check, config sqlc.TcpCheckConfig) domaincheck.Check {
	return domaincheck.Check{
		ID:              row.ID.String(),
		ProjectID:       row.ProjectID.String(),
		Name:            row.Name,
		Type:            domaincheck.Type(row.CheckType),
		Target:          row.Target,
		Selector:        cloneRawMessage(row.Selector),
		Description:     row.Description,
		IntervalSeconds: row.IntervalSeconds,
		TCPConfig:       tcpConfigPtr(mapTCPConfig(config)),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		DeletedAt:       row.DeletedAt,
	}
}

func mapStoredTracerouteCheck(row sqlc.Check, config sqlc.TracerouteCheckConfig) domaincheck.Check {
	return domaincheck.Check{
		ID:               row.ID.String(),
		ProjectID:        row.ProjectID.String(),
		Name:             row.Name,
		Type:             domaincheck.Type(row.CheckType),
		Target:           row.Target,
		Selector:         cloneRawMessage(row.Selector),
		Description:      row.Description,
		IntervalSeconds:  row.IntervalSeconds,
		TracerouteConfig: tracerouteConfigPtr(mapTracerouteConfig(config)),
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
		DeletedAt:        row.DeletedAt,
	}
}

func mapStoredHTTPCheck(row sqlc.Check, stored sqlc.HttpCheckConfig) (domaincheck.Check, error) {
	config, err := mapHTTPConfig(stored)
	if err != nil {
		return domaincheck.Check{}, err
	}
	return domaincheck.Check{
		ID: row.ID.String(), ProjectID: row.ProjectID.String(), Name: row.Name,
		Type: domaincheck.Type(row.CheckType), Target: row.Target,
		Selector: cloneRawMessage(row.Selector), Description: row.Description,
		IntervalSeconds: row.IntervalSeconds, HTTPConfig: httpConfigPtr(config),
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, DeletedAt: row.DeletedAt,
	}, nil
}

//nolint:dupl // sqlc emits distinct list/get row types for the same projection.
func mapListCheck(row sqlc.ListActiveChecksForProjectRow) (domaincheck.Check, error) {
	return mapSelectedCheck(selectedCheckRow{
		id: row.ID, projectID: row.ProjectID, name: row.Name, checkType: row.CheckType,
		target: row.Target, selector: row.Selector, description: row.Description, intervalSeconds: row.IntervalSeconds,
		pingPacketCount: row.PingPacketCount, pingPacketSizeBytes: row.PingPacketSizeBytes,
		pingTimeoutMs: row.PingTimeoutMs, pingIPFamily: row.PingIpFamily,
		tcpPort: row.TcpPort, tcpTimeoutMs: row.TcpTimeoutMs, tcpIPFamily: row.TcpIpFamily,
		httpMethod: row.HttpMethod, httpHeaders: row.HttpHeaders, httpBody: row.HttpBody,
		httpTimeoutMs: row.HttpTimeoutMs, httpIPFamily: row.HttpIpFamily,
		httpFollowRedirects: row.HttpFollowRedirects, httpSkipTLSVerify: row.HttpSkipTlsVerify,
		httpExpectedStatusCodes: row.HttpExpectedStatusCodes, httpExpectedStatusClasses: row.HttpExpectedStatusClasses,
		httpBodyContains: row.HttpBodyContains, tracerouteProtocol: row.TracerouteProtocol,
		tracerouteMaxHops: row.TracerouteMaxHops, tracerouteTimeoutMs: row.TracerouteTimeoutMs,
		tracerouteQueriesPerHop: row.TracerouteQueriesPerHop, traceroutePacketSizeBytes: row.TraceroutePacketSizeBytes,
		traceroutePort: row.TraceroutePort, tracerouteIPFamily: row.TracerouteIpFamily,
		createdAt: row.CreatedAt, updatedAt: row.UpdatedAt, deletedAt: row.DeletedAt,
	})
}

//nolint:dupl // sqlc emits distinct list/get row types for the same projection.
func mapGetCheck(row sqlc.GetActiveCheckForProjectRow) (domaincheck.Check, error) {
	return mapSelectedCheck(selectedCheckRow{
		id: row.ID, projectID: row.ProjectID, name: row.Name, checkType: row.CheckType,
		target: row.Target, selector: row.Selector, description: row.Description, intervalSeconds: row.IntervalSeconds,
		pingPacketCount: row.PingPacketCount, pingPacketSizeBytes: row.PingPacketSizeBytes,
		pingTimeoutMs: row.PingTimeoutMs, pingIPFamily: row.PingIpFamily,
		tcpPort: row.TcpPort, tcpTimeoutMs: row.TcpTimeoutMs, tcpIPFamily: row.TcpIpFamily,
		httpMethod: row.HttpMethod, httpHeaders: row.HttpHeaders, httpBody: row.HttpBody,
		httpTimeoutMs: row.HttpTimeoutMs, httpIPFamily: row.HttpIpFamily,
		httpFollowRedirects: row.HttpFollowRedirects, httpSkipTLSVerify: row.HttpSkipTlsVerify,
		httpExpectedStatusCodes: row.HttpExpectedStatusCodes, httpExpectedStatusClasses: row.HttpExpectedStatusClasses,
		httpBodyContains: row.HttpBodyContains, tracerouteProtocol: row.TracerouteProtocol,
		tracerouteMaxHops: row.TracerouteMaxHops, tracerouteTimeoutMs: row.TracerouteTimeoutMs,
		tracerouteQueriesPerHop: row.TracerouteQueriesPerHop, traceroutePacketSizeBytes: row.TraceroutePacketSizeBytes,
		traceroutePort: row.TraceroutePort, tracerouteIPFamily: row.TracerouteIpFamily,
		createdAt: row.CreatedAt, updatedAt: row.UpdatedAt, deletedAt: row.DeletedAt,
	})
}

type selectedCheckRow struct {
	id, projectID                             uuid.UUID
	name, target                              string
	checkType                                 sqlc.CheckType
	selector                                  []byte
	description                               *string
	intervalSeconds                           int32
	pingPacketCount, pingPacketSizeBytes      *int32
	pingTimeoutMs                             *int32
	pingIPFamily                              *sqlc.IpFamily
	tcpPort, tcpTimeoutMs                     *int32
	tcpIPFamily                               *sqlc.IpFamily
	httpMethod                                *sqlc.HttpMethod
	httpHeaders                               []byte
	httpBody                                  *string
	httpTimeoutMs                             *int32
	httpIPFamily                              *sqlc.IpFamily
	httpFollowRedirects, httpSkipTLSVerify    *bool
	httpExpectedStatusCodes                   []int32
	httpExpectedStatusClasses                 []int32
	httpBodyContains                          *string
	tracerouteProtocol                        *sqlc.TracerouteProtocol
	tracerouteMaxHops, tracerouteTimeoutMs    *int32
	tracerouteQueriesPerHop                   *int32
	traceroutePacketSizeBytes, traceroutePort *int32
	tracerouteIPFamily                        *sqlc.IpFamily
	createdAt, updatedAt                      time.Time
	deletedAt                                 *time.Time
}

func mapSelectedCheck(row selectedCheckRow) (domaincheck.Check, error) {
	check := domaincheck.Check{
		ID:              row.id.String(),
		ProjectID:       row.projectID.String(),
		Name:            row.name,
		Type:            domaincheck.Type(row.checkType),
		Target:          row.target,
		Selector:        cloneRawMessage(row.selector),
		Description:     row.description,
		IntervalSeconds: row.intervalSeconds,
		CreatedAt:       row.createdAt,
		UpdatedAt:       row.updatedAt,
		DeletedAt:       row.deletedAt,
	}
	check.PingConfig = mapOptionalPingConfig(row.pingPacketCount, row.pingPacketSizeBytes, row.pingTimeoutMs, row.pingIPFamily)
	check.TCPConfig = mapOptionalTCPConfig(row.tcpPort, row.tcpTimeoutMs, row.tcpIPFamily)
	var err error
	check.HTTPConfig, err = mapOptionalHTTPConfig(row.httpMethod, row.httpHeaders, row.httpBody, row.httpTimeoutMs, row.httpIPFamily, row.httpFollowRedirects, row.httpSkipTLSVerify, row.httpExpectedStatusCodes, row.httpExpectedStatusClasses, row.httpBodyContains)
	if err != nil {
		return domaincheck.Check{}, err
	}
	check.TracerouteConfig = mapOptionalTracerouteConfig(
		row.tracerouteProtocol,
		row.tracerouteMaxHops,
		row.tracerouteTimeoutMs,
		row.tracerouteQueriesPerHop,
		row.traceroutePacketSizeBytes,
		row.traceroutePort,
		row.tracerouteIPFamily,
	)

	return check, nil
}

func mapPingConfig(row sqlc.PingCheckConfig) domainping.Config {
	return domainping.Config{
		PacketCount:     row.PacketCount,
		PacketSizeBytes: row.PacketSizeBytes,
		TimeoutMs:       row.TimeoutMs,
		IPFamily:        mapIPFamily(row.IpFamily),
	}
}

func pingConfigPtr(config domainping.Config) *domainping.Config {
	return &config
}

func mapTCPConfig(row sqlc.TcpCheckConfig) domaintcp.Config {
	return domaintcp.Config{
		Port:      row.Port,
		TimeoutMs: row.TimeoutMs,
		IPFamily:  mapIPFamily(row.IpFamily),
	}
}

func tcpConfigPtr(config domaintcp.Config) *domaintcp.Config {
	return &config
}

func mapHTTPConfig(row sqlc.HttpCheckConfig) (domainhttp.Config, error) {
	headers, err := decodeHTTPHeaders(row.Headers)
	if err != nil {
		return domainhttp.Config{}, err
	}
	return domainhttp.Config{
		Method: domainhttp.Method(row.Method), Headers: headers, Body: row.Body,
		TimeoutMs: row.TimeoutMs, IPFamily: mapIPFamily(row.IpFamily),
		FollowRedirects: row.FollowRedirects, SkipTLSVerify: row.SkipTlsVerify,
		ExpectedStatusCodes:   append([]int32(nil), row.ExpectedStatusCodes...),
		ExpectedStatusClasses: append([]int32(nil), row.ExpectedStatusClasses...),
		BodyContains:          row.BodyContains,
	}, nil
}

func httpConfigPtr(config domainhttp.Config) *domainhttp.Config { return &config }

func mapOptionalHTTPConfig(method *sqlc.HttpMethod, headers []byte, body *string, timeoutMs *int32, ipFamily *sqlc.IpFamily, followRedirects, skipTLSVerify *bool, codes, classes []int32, bodyContains *string) (*domainhttp.Config, error) {
	if method == nil || timeoutMs == nil || followRedirects == nil || skipTLSVerify == nil {
		return nil, nil //nolint:nilnil // Nil means this joined row has no HTTP config.
	}
	values, err := decodeHTTPHeaders(headers)
	if err != nil {
		return nil, err
	}
	return &domainhttp.Config{
		Method: domainhttp.Method(*method), Headers: values, Body: body, TimeoutMs: *timeoutMs,
		IPFamily: mapIPFamily(ipFamily), FollowRedirects: *followRedirects,
		SkipTLSVerify: *skipTLSVerify, ExpectedStatusCodes: append([]int32(nil), codes...),
		ExpectedStatusClasses: append([]int32(nil), classes...), BodyContains: bodyContains,
	}, nil
}

func decodeHTTPHeaders(data []byte) ([]domainhttp.Header, error) {
	var headers []domainhttp.Header
	if err := json.Unmarshal(data, &headers); err != nil {
		return nil, fmt.Errorf("decode HTTP check headers: %w", err)
	}
	return headers, nil
}

func mapTracerouteConfig(row sqlc.TracerouteCheckConfig) domaintraceroute.Config {
	return domaintraceroute.Config{
		Protocol:        domaintraceroute.Protocol(row.Protocol),
		MaxHops:         row.MaxHops,
		TimeoutMs:       row.TimeoutMs,
		QueriesPerHop:   row.QueriesPerHop,
		PacketSizeBytes: row.PacketSizeBytes,
		Port:            row.Port,
		IPFamily:        mapIPFamily(row.IpFamily),
	}
}

func tracerouteConfigPtr(config domaintraceroute.Config) *domaintraceroute.Config {
	return &config
}

func mapOptionalPingConfig(packetCount, packetSizeBytes, timeoutMs *int32, ipFamily *sqlc.IpFamily) *domainping.Config {
	if packetCount == nil || packetSizeBytes == nil || timeoutMs == nil {
		return nil
	}

	return &domainping.Config{
		PacketCount:     *packetCount,
		PacketSizeBytes: *packetSizeBytes,
		TimeoutMs:       *timeoutMs,
		IPFamily:        mapIPFamily(ipFamily),
	}
}

func mapOptionalTCPConfig(port, timeoutMs *int32, ipFamily *sqlc.IpFamily) *domaintcp.Config {
	if port == nil || timeoutMs == nil {
		return nil
	}

	return &domaintcp.Config{
		Port:      *port,
		TimeoutMs: *timeoutMs,
		IPFamily:  mapIPFamily(ipFamily),
	}
}

func mapOptionalTracerouteConfig(
	protocol *sqlc.TracerouteProtocol,
	maxHops *int32,
	timeoutMs *int32,
	queriesPerHop *int32,
	packetSizeBytes *int32,
	port *int32,
	ipFamily *sqlc.IpFamily,
) *domaintraceroute.Config {
	if protocol == nil || maxHops == nil || timeoutMs == nil || queriesPerHop == nil || packetSizeBytes == nil || port == nil {
		return nil
	}

	return &domaintraceroute.Config{
		Protocol:        domaintraceroute.Protocol(*protocol),
		MaxHops:         *maxHops,
		TimeoutMs:       *timeoutMs,
		QueriesPerHop:   *queriesPerHop,
		PacketSizeBytes: *packetSizeBytes,
		Port:            *port,
		IPFamily:        mapIPFamily(ipFamily),
	}
}

func sqlcCheckType(value domaincheck.Type) sqlc.CheckType {
	return sqlc.CheckType(value)
}

func sqlcTracerouteProtocol(value domaintraceroute.Protocol) sqlc.TracerouteProtocol {
	return sqlc.TracerouteProtocol(value)
}

func sqlcIPFamily(value *domainnetwork.IPFamily) *sqlc.IpFamily {
	if value == nil {
		return nil
	}

	ipFamily := sqlc.IpFamily(*value)
	return &ipFamily
}

func mapIPFamily(value *sqlc.IpFamily) *domainnetwork.IPFamily {
	if value == nil {
		return nil
	}

	ipFamily := domainnetwork.IPFamily(*value)
	return &ipFamily
}

func mapLabels(rows []sqlc.Label) []domainlabel.Label {
	labels := make([]domainlabel.Label, 0, len(rows))
	for _, row := range rows {
		labels = append(labels, domainlabel.Label{
			ID:        row.ID.String(),
			ProjectID: row.ProjectID.String(),
			Key:       row.Key,
			Value:     row.Value,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
			DeletedAt: row.DeletedAt,
		})
	}

	return labels
}

func cloneRawMessage(value []byte) json.RawMessage {
	if len(value) == 0 {
		return json.RawMessage(`{}`)
	}

	return append(json.RawMessage(nil), value...)
}
