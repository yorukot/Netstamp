package pgprobe

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

func mapCreateProbeRow(row sqlc.CreateProbeRow) domainprobe.Probe {
	return mapProbeFields(
		row.ID,
		row.ProjectID,
		row.Name,
		row.Enabled,
		row.Location,
		row.LocationName,
		row.CreatedAt,
		row.UpdatedAt,
		row.DeletedAt,
		nil,
	)
}

func mapListProbeRows(rows []sqlc.ListActiveProbesForProjectRow) []domainprobe.Probe {
	probeIndex := make(map[uuid.UUID]int)
	probes := make([]domainprobe.Probe, 0)
	for _, row := range rows {
		index, ok := probeIndex[row.ID]
		if !ok {
			index = len(probes)
			probeIndex[row.ID] = index
			probes = append(probes, mapProbeFields(
				row.ID,
				row.ProjectID,
				row.Name,
				row.Enabled,
				row.Location,
				row.LocationName,
				row.CreatedAt,
				row.UpdatedAt,
				row.DeletedAt,
				mapListProbeStatus(row),
			))
		}
		if label, ok := mapListProbeLabel(row); ok {
			probes[index].Labels = append(probes[index].Labels, label)
		}
	}

	return probes
}

func mapGetProbeRows(rows []sqlc.GetActiveProbeRowsForProjectRow) (domainprobe.Probe, bool) {
	if len(rows) == 0 {
		return domainprobe.Probe{}, false
	}

	first := rows[0]
	probe := mapProbeFields(
		first.ID,
		first.ProjectID,
		first.Name,
		first.Enabled,
		first.Location,
		first.LocationName,
		first.CreatedAt,
		first.UpdatedAt,
		first.DeletedAt,
		mapGetProbeStatus(first),
	)
	for _, row := range rows {
		if label, ok := mapGetProbeLabel(row); ok {
			probe.Labels = append(probe.Labels, label)
		}
	}

	return probe, true
}

func mapProbeFields(
	id uuid.UUID,
	projectID uuid.UUID,
	name string,
	enabled bool,
	location pgtype.Point,
	locationName *string,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
	status *domainprobe.Status,
) domainprobe.Probe {
	latitude, longitude := coordinatesFromPoint(location)

	return domainprobe.Probe{
		ID:           id.String(),
		ProjectID:    projectID.String(),
		Name:         name,
		Enabled:      enabled,
		LocationName: locationName,
		Latitude:     latitude,
		Longitude:    longitude,
		Labels:       []domainlabel.Label{},
		Status:       status,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		DeletedAt:    deletedAt,
	}
}

func coordinatesFromPoint(point pgtype.Point) (*float64, *float64) {
	if !point.Valid {
		return nil, nil
	}

	lon := point.P.X
	lat := point.P.Y
	return &lat, &lon
}

func mapProbeCredential(row sqlc.GetActiveProbeCredentialRow) domainprobe.Credential {
	return domainprobe.Credential{
		ProbeID:    row.ID.String(),
		ProjectID:  row.ProjectID.String(),
		Enabled:    row.Enabled,
		SecretHash: row.SecretHash,
	}
}

func mapCreateProbeStatus(row sqlc.CreateProbeStatusRow) domainprobe.Status {
	return mapProbeStatusFields(
		row.ProbeID,
		row.Status,
		row.LastSeenAt,
		row.OnlineSince,
		row.AgentVersion,
		row.PublicV4,
		row.PublicV6,
		row.As,
		row.Addrs,
		row.UpdatedAt,
	)
}

func mapUpdateProbeStatus(row sqlc.UpdateProbeStatusRow) domainprobe.Status {
	return mapProbeStatusFields(
		row.ProbeID,
		row.Status,
		row.LastSeenAt,
		row.OnlineSince,
		row.AgentVersion,
		row.PublicV4,
		row.PublicV6,
		row.As,
		row.Addrs,
		row.UpdatedAt,
	)
}

func mapUpdateProbeIPFamilyStatus(row sqlc.UpdateProbeIPFamilyCapabilitiesRow) domainprobe.Status {
	return mapProbeStatusFields(
		row.ProbeID,
		row.Status,
		row.LastSeenAt,
		row.OnlineSince,
		row.AgentVersion,
		row.PublicV4,
		row.PublicV6,
		row.As,
		row.Addrs,
		row.UpdatedAt,
	)
}

func mapListProbeStatus(row sqlc.ListActiveProbesForProjectRow) *domainprobe.Status {
	status := mapProbeStatusFields(
		row.ID,
		row.Status,
		row.StatusLastSeenAt,
		row.StatusOnlineSince,
		row.StatusAgentVersion,
		row.StatusPublicV4,
		row.StatusPublicV6,
		row.StatusAs,
		row.StatusAddrs,
		timeValue(row.StatusUpdatedAt),
	)

	return &status
}

func mapGetProbeStatus(row sqlc.GetActiveProbeRowsForProjectRow) *domainprobe.Status {
	status := mapProbeStatusFields(
		row.ID,
		row.Status,
		row.StatusLastSeenAt,
		row.StatusOnlineSince,
		row.StatusAgentVersion,
		row.StatusPublicV4,
		row.StatusPublicV6,
		row.StatusAs,
		row.StatusAddrs,
		timeValue(row.StatusUpdatedAt),
	)

	return &status
}

func mapProbeStatusFields(
	probeID uuid.UUID,
	state sqlc.ProbeState,
	lastSeenAt *time.Time,
	onlineSince *time.Time,
	agentVersion *string,
	publicV4 *netip.Addr,
	publicV6 *netip.Addr,
	as *string,
	addrs []netip.Addr,
	updatedAt time.Time,
) domainprobe.Status {
	domainState := domainprobe.State(state)
	activeOnlineSince := onlineSinceForState(domainState, onlineSince)

	return domainprobe.Status{
		ProbeID:       probeID.String(),
		State:         domainState,
		LastSeenAt:    lastSeenAt,
		OnlineSince:   activeOnlineSince,
		UptimeSeconds: uptimeSecondsFromOnlineSince(activeOnlineSince),
		AgentVersion:  agentVersion,
		PublicV4:      cloneAddr(publicV4),
		PublicV6:      cloneAddr(publicV6),
		AS:            as,
		Addrs:         append([]netip.Addr(nil), addrs...),
		UpdatedAt:     updatedAt,
	}
}

func onlineSinceForState(state domainprobe.State, onlineSince *time.Time) *time.Time {
	if state != domainprobe.StateOnline || onlineSince == nil {
		return nil
	}

	return onlineSince
}

func uptimeSecondsFromOnlineSince(onlineSince *time.Time) *int64 {
	if onlineSince == nil {
		return nil
	}

	seconds := int64(time.Since(*onlineSince).Seconds())
	if seconds < 0 {
		seconds = 0
	}

	return &seconds
}

func mapListProbeLabel(row sqlc.ListActiveProbesForProjectRow) (domainlabel.Label, bool) {
	if row.LabelID == nil || row.LabelProjectID == nil || row.LabelKey == nil || row.LabelValue == nil || row.LabelCreatedAt == nil || row.LabelUpdatedAt == nil {
		return domainlabel.Label{}, false
	}

	return domainlabel.Label{
		ID:        row.LabelID.String(),
		ProjectID: row.LabelProjectID.String(),
		Key:       *row.LabelKey,
		Value:     *row.LabelValue,
		CreatedAt: *row.LabelCreatedAt,
		UpdatedAt: *row.LabelUpdatedAt,
		DeletedAt: row.LabelDeletedAt,
	}, true
}

func mapGetProbeLabel(row sqlc.GetActiveProbeRowsForProjectRow) (domainlabel.Label, bool) {
	if row.LabelID == nil || row.LabelProjectID == nil || row.LabelKey == nil || row.LabelValue == nil || row.LabelCreatedAt == nil || row.LabelUpdatedAt == nil {
		return domainlabel.Label{}, false
	}

	return domainlabel.Label{
		ID:        row.LabelID.String(),
		ProjectID: row.LabelProjectID.String(),
		Key:       *row.LabelKey,
		Value:     *row.LabelValue,
		CreatedAt: *row.LabelCreatedAt,
		UpdatedAt: *row.LabelUpdatedAt,
		DeletedAt: row.LabelDeletedAt,
	}, true
}

func mapAssignment(row sqlc.ListActiveAssignmentsForProbeRow) (domainassignment.Assignment, error) {
	httpConfig, err := mapOptionalHTTPConfig(
		row.HttpMethod,
		row.HttpHeaders,
		row.HttpBody,
		row.HttpTimeoutMs,
		row.HttpIpFamily,
		row.HttpFollowRedirects,
		row.HttpSkipTlsVerify,
		row.HttpExpectedStatusCodes,
		row.HttpExpectedStatusClasses,
		row.HttpBodyContains,
	)
	if err != nil {
		return domainassignment.Assignment{}, err
	}
	return domainassignment.Assignment{
		ID:              row.AssignmentID.String(),
		ProjectID:       row.ProjectID.String(),
		ProbeID:         row.ProbeID.String(),
		CheckID:         row.CheckID.String(),
		ProbeStorageID:  row.ProbeInternalID,
		CheckStorageID:  row.CheckInternalID,
		CheckVersion:    row.CheckVersion,
		SelectorVersion: row.SelectorVersion,
		Check: &domaincheck.Check{
			ID:              row.CheckID.String(),
			ProjectID:       row.ProjectID.String(),
			Name:            row.CheckName,
			Type:            domaincheck.Type(row.CheckType),
			Target:          row.Target,
			IntervalSeconds: row.IntervalSeconds,
			PingConfig:      mapOptionalPingConfig(row.PingPacketCount, row.PingPacketSizeBytes, row.PingTimeoutMs, row.PingIpFamily),
			TCPConfig:       mapOptionalTCPConfig(row.TcpPort, row.TcpTimeoutMs, row.TcpIpFamily),
			HTTPConfig:      httpConfig,
			TracerouteConfig: mapOptionalTracerouteConfig(
				row.TracerouteProtocol,
				row.TracerouteMaxHops,
				row.TracerouteTimeoutMs,
				row.TracerouteQueriesPerHop,
				row.TraceroutePacketSizeBytes,
				row.TraceroutePort,
				row.TracerouteIpFamily,
			),
		},
	}, nil
}

func mapAssignmentForProbeChecks(row sqlc.ListActiveAssignmentsForProbeChecksRow) (domainassignment.Assignment, error) {
	assignment, err := mapAssignment(sqlc.ListActiveAssignmentsForProbeRow(row))
	if err != nil {
		return domainassignment.Assignment{}, err
	}
	assignment.Probe = &domainprobe.Probe{
		ID:        row.ProbeID.String(),
		ProjectID: row.ProjectID.String(),
		Name:      row.ProbeName,
	}
	return assignment, nil
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

func mapOptionalHTTPConfig(method *sqlc.HttpMethod, headers []byte, body *string, timeoutMs *int32, ipFamily *sqlc.IpFamily, followRedirects, skipTLSVerify *bool, codes, classes []int32, bodyContains *string) (*domainhttp.Config, error) {
	if method == nil || timeoutMs == nil || followRedirects == nil || skipTLSVerify == nil {
		return nil, nil //nolint:nilnil // Nil means this joined row has no HTTP config.
	}
	var values []domainhttp.Header
	if err := json.Unmarshal(headers, &values); err != nil {
		return nil, fmt.Errorf("decode HTTP check headers: %w", err)
	}
	return &domainhttp.Config{
		Method: domainhttp.Method(*method), Headers: values, Body: body, TimeoutMs: *timeoutMs,
		IPFamily: mapIPFamily(ipFamily), FollowRedirects: *followRedirects,
		SkipTLSVerify: *skipTLSVerify, ExpectedStatusCodes: append([]int32(nil), codes...),
		ExpectedStatusClasses: append([]int32(nil), classes...), BodyContains: bodyContains,
	}, nil
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

func sqlcProbeState(value domainprobe.State) sqlc.ProbeState {
	return sqlc.ProbeState(value)
}

func mapIPFamily(value *sqlc.IpFamily) *domainnetwork.IPFamily {
	if value == nil {
		return nil
	}

	ipFamily := domainnetwork.IPFamily(*value)
	return &ipFamily
}

func cloneAddr(value *netip.Addr) *netip.Addr {
	if value == nil {
		return nil
	}

	addr := *value
	return &addr
}

func timeValue(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}
