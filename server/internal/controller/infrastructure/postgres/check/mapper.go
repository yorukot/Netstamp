package pgcheck

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
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
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		DeletedAt:       timePtr(row.DeletedAt),
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
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		DeletedAt:       timePtr(row.DeletedAt),
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
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
		DeletedAt:        timePtr(row.DeletedAt),
	}
}

func mapListCheck(row sqlc.ListActiveChecksForProjectRow) domaincheck.Check {
	return mapSelectedCheck(
		row.ID,
		row.ProjectID,
		row.Name,
		row.CheckType,
		row.Target,
		row.Selector,
		row.Description,
		row.IntervalSeconds,
		row.PingPacketCount,
		row.PingPacketSizeBytes,
		row.PingTimeoutMs,
		row.PingIpFamily,
		row.TcpPort,
		row.TcpTimeoutMs,
		row.TcpIpFamily,
		row.TracerouteProtocol,
		row.TracerouteMaxHops,
		row.TracerouteTimeoutMs,
		row.TracerouteQueriesPerHop,
		row.TraceroutePacketSizeBytes,
		row.TraceroutePort,
		row.TracerouteIpFamily,
		row.CreatedAt,
		row.UpdatedAt,
		row.DeletedAt,
	)
}

func mapGetCheck(row sqlc.GetActiveCheckForProjectRow) domaincheck.Check {
	return mapSelectedCheck(
		row.ID,
		row.ProjectID,
		row.Name,
		row.CheckType,
		row.Target,
		row.Selector,
		row.Description,
		row.IntervalSeconds,
		row.PingPacketCount,
		row.PingPacketSizeBytes,
		row.PingTimeoutMs,
		row.PingIpFamily,
		row.TcpPort,
		row.TcpTimeoutMs,
		row.TcpIpFamily,
		row.TracerouteProtocol,
		row.TracerouteMaxHops,
		row.TracerouteTimeoutMs,
		row.TracerouteQueriesPerHop,
		row.TraceroutePacketSizeBytes,
		row.TraceroutePort,
		row.TracerouteIpFamily,
		row.CreatedAt,
		row.UpdatedAt,
		row.DeletedAt,
	)
}

func mapSelectedCheck(
	id uuid.UUID,
	projectID uuid.UUID,
	name string,
	checkType sqlc.CheckType,
	target string,
	selector []byte,
	description *string,
	intervalSeconds int32,
	pingPacketCount *int32,
	pingPacketSizeBytes *int32,
	pingTimeoutMs *int32,
	pingIPFamily *sqlc.IpFamily,
	tcpPort *int32,
	tcpTimeoutMs *int32,
	tcpIPFamily *sqlc.IpFamily,
	tracerouteProtocol *sqlc.TracerouteProtocol,
	tracerouteMaxHops *int32,
	tracerouteTimeoutMs *int32,
	tracerouteQueriesPerHop *int32,
	traceroutePacketSizeBytes *int32,
	traceroutePort *int32,
	tracerouteIPFamily *sqlc.IpFamily,
	createdAt pgtype.Timestamptz,
	updatedAt pgtype.Timestamptz,
	deletedAt pgtype.Timestamptz,
) domaincheck.Check {
	check := domaincheck.Check{
		ID:              id.String(),
		ProjectID:       projectID.String(),
		Name:            name,
		Type:            domaincheck.Type(checkType),
		Target:          target,
		Selector:        cloneRawMessage(selector),
		Description:     description,
		IntervalSeconds: intervalSeconds,
		CreatedAt:       createdAt.Time,
		UpdatedAt:       updatedAt.Time,
		DeletedAt:       timePtr(deletedAt),
	}
	check.PingConfig = mapOptionalPingConfig(pingPacketCount, pingPacketSizeBytes, pingTimeoutMs, pingIPFamily)
	check.TCPConfig = mapOptionalTCPConfig(tcpPort, tcpTimeoutMs, tcpIPFamily)
	check.TracerouteConfig = mapOptionalTracerouteConfig(
		tracerouteProtocol,
		tracerouteMaxHops,
		tracerouteTimeoutMs,
		tracerouteQueriesPerHop,
		traceroutePacketSizeBytes,
		traceroutePort,
		tracerouteIPFamily,
	)

	return check
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
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
			DeletedAt: timePtr(row.DeletedAt),
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

func timePtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}
