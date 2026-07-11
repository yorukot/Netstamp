package pgassignment

import (
	"encoding/json"
	"fmt"
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
	domainselector "github.com/yorukot/netstamp/internal/domain/selector"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type activeProbeLabels struct {
	probeID   uuid.UUID
	projectID uuid.UUID
	name      string
	enabled   bool
	labels    []domainlabel.Label
}

func activeProbeFromRows(rows []sqlc.GetActiveProbeRowsForProjectRow) (activeProbeLabels, bool) {
	if len(rows) == 0 {
		return activeProbeLabels{}, false
	}

	probe := activeProbeLabels{
		probeID:   rows[0].ID,
		projectID: rows[0].ProjectID,
		name:      rows[0].Name,
		enabled:   rows[0].Enabled,
	}
	for _, row := range rows {
		if label, ok := mapGetProbeLabel(row); ok {
			probe.labels = append(probe.labels, label)
		}
	}

	return probe, true
}

func activeProbeLabelsFromRows(rows []sqlc.ListActiveEnabledProbeLabelsForProjectRow) []activeProbeLabels {
	probeIndex := make(map[uuid.UUID]int)
	probes := make([]activeProbeLabels, 0)
	for _, row := range rows {
		index, ok := probeIndex[row.ProbeID]
		if !ok {
			index = len(probes)
			probeIndex[row.ProbeID] = index
			probes = append(probes, activeProbeLabels{
				probeID:   row.ProbeID,
				projectID: row.ProbeProjectID,
				name:      row.ProbeName,
				enabled:   row.ProbeEnabled,
			})
		}
		if label, ok := mapEnabledProbeLabel(row); ok {
			probes[index].labels = append(probes[index].labels, label)
		}
	}

	return probes
}

func activeProbeLabelsFromProjectRows(rows []sqlc.ListActiveProbesForProjectRow) []activeProbeLabels {
	probeIndex := make(map[uuid.UUID]int)
	probes := make([]activeProbeLabels, 0)
	for _, row := range rows {
		index, ok := probeIndex[row.ID]
		if !ok {
			index = len(probes)
			probeIndex[row.ID] = index
			probes = append(probes, activeProbeLabels{
				probeID:   row.ID,
				projectID: row.ProjectID,
				name:      row.Name,
				enabled:   row.Enabled,
			})
		}
		if label, ok := mapListProbeLabel(row); ok {
			probes[index].labels = append(probes[index].labels, label)
		}
	}

	return probes
}

func probeCandidatesFromActive(probes []activeProbeLabels) []domainassignment.ProbeAssignmentCandidate {
	candidates := make([]domainassignment.ProbeAssignmentCandidate, 0, len(probes))
	for _, probe := range probes {
		candidates = append(candidates, probeCandidateFromActive(probe))
	}
	return candidates
}

func probeCandidateFromActive(probe activeProbeLabels) domainassignment.ProbeAssignmentCandidate {
	return domainassignment.ProbeAssignmentCandidate{
		ProbeID:   probe.probeID.String(),
		ProjectID: probe.projectID.String(),
		Name:      probe.name,
		Enabled:   probe.enabled,
		Labels:    append([]domainlabel.Label(nil), probe.labels...),
	}
}

func listCheckCandidates(rows []sqlc.ListActiveChecksForProjectRow) ([]domainassignment.CheckAssignmentCandidate, error) {
	candidates := make([]domainassignment.CheckAssignmentCandidate, 0, len(rows))
	for _, row := range rows {
		selector, selectorRaw, err := listCheckSelector(row)
		if err != nil {
			return nil, err
		}
		check, err := listCheck(row)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, domainassignment.CheckAssignmentCandidate{
			Check:           check,
			Selector:        selector,
			SelectorVersion: domaincheck.SelectorVersion(selectorRaw),
			CheckVersion:    check.Hash(),
		})
	}
	return candidates, nil
}

func checkCandidate(row sqlc.GetActiveCheckForProjectRow) (domainassignment.CheckAssignmentCandidate, error) {
	selector, selectorRaw, err := checkSelector(row)
	if err != nil {
		return domainassignment.CheckAssignmentCandidate{}, err
	}
	check, err := getCheck(row)
	if err != nil {
		return domainassignment.CheckAssignmentCandidate{}, err
	}
	return domainassignment.CheckAssignmentCandidate{
		Check:           check,
		Selector:        selector,
		SelectorVersion: domaincheck.SelectorVersion(selectorRaw),
		CheckVersion:    check.Hash(),
	}, nil
}

//nolint:dupl // sqlc generates distinct row types with the same check columns.
func listCheck(row sqlc.ListActiveChecksForProjectRow) (domaincheck.Check, error) {
	return newCheck(row.ID, row.ProjectID, row.Name, row.CheckType, row.Target, row.Selector, row.Description, row.IntervalSeconds, row.CreatedAt, row.UpdatedAt, row.DeletedAt, checkConfigs{
		pingPacketCount:         row.PingPacketCount,
		pingPacketSizeBytes:     row.PingPacketSizeBytes,
		pingTimeoutMs:           row.PingTimeoutMs,
		pingIPFamily:            row.PingIpFamily,
		tcpPort:                 row.TcpPort,
		tcpTimeoutMs:            row.TcpTimeoutMs,
		tcpIPFamily:             row.TcpIpFamily,
		httpMethod:              row.HttpMethod,
		httpHeaders:             row.HttpHeaders,
		httpBody:                row.HttpBody,
		httpTimeoutMs:           row.HttpTimeoutMs,
		httpIPFamily:            row.HttpIpFamily,
		httpFollowRedirects:     row.HttpFollowRedirects,
		httpSkipTLSVerify:       row.HttpSkipTlsVerify,
		httpExpectedCodes:       row.HttpExpectedStatusCodes,
		httpExpectedClasses:     row.HttpExpectedStatusClasses,
		httpBodyContains:        row.HttpBodyContains,
		tracerouteProtocol:      row.TracerouteProtocol,
		tracerouteMaxHops:       row.TracerouteMaxHops,
		tracerouteTimeoutMs:     row.TracerouteTimeoutMs,
		tracerouteQueriesPerHop: row.TracerouteQueriesPerHop,
		traceroutePacketSize:    row.TraceroutePacketSizeBytes,
		traceroutePort:          row.TraceroutePort,
		tracerouteIPFamily:      row.TracerouteIpFamily,
	})
}

//nolint:dupl // sqlc generates distinct row types with the same check columns.
func getCheck(row sqlc.GetActiveCheckForProjectRow) (domaincheck.Check, error) {
	return newCheck(row.ID, row.ProjectID, row.Name, row.CheckType, row.Target, row.Selector, row.Description, row.IntervalSeconds, row.CreatedAt, row.UpdatedAt, row.DeletedAt, checkConfigs{
		pingPacketCount:         row.PingPacketCount,
		pingPacketSizeBytes:     row.PingPacketSizeBytes,
		pingTimeoutMs:           row.PingTimeoutMs,
		pingIPFamily:            row.PingIpFamily,
		tcpPort:                 row.TcpPort,
		tcpTimeoutMs:            row.TcpTimeoutMs,
		tcpIPFamily:             row.TcpIpFamily,
		httpMethod:              row.HttpMethod,
		httpHeaders:             row.HttpHeaders,
		httpBody:                row.HttpBody,
		httpTimeoutMs:           row.HttpTimeoutMs,
		httpIPFamily:            row.HttpIpFamily,
		httpFollowRedirects:     row.HttpFollowRedirects,
		httpSkipTLSVerify:       row.HttpSkipTlsVerify,
		httpExpectedCodes:       row.HttpExpectedStatusCodes,
		httpExpectedClasses:     row.HttpExpectedStatusClasses,
		httpBodyContains:        row.HttpBodyContains,
		tracerouteProtocol:      row.TracerouteProtocol,
		tracerouteMaxHops:       row.TracerouteMaxHops,
		tracerouteTimeoutMs:     row.TracerouteTimeoutMs,
		tracerouteQueriesPerHop: row.TracerouteQueriesPerHop,
		traceroutePacketSize:    row.TraceroutePacketSizeBytes,
		traceroutePort:          row.TraceroutePort,
		tracerouteIPFamily:      row.TracerouteIpFamily,
	})
}

type checkConfigs struct {
	pingPacketCount         *int32
	pingPacketSizeBytes     *int32
	pingTimeoutMs           *int32
	pingIPFamily            *sqlc.IpFamily
	tcpPort                 *int32
	tcpTimeoutMs            *int32
	tcpIPFamily             *sqlc.IpFamily
	httpMethod              *sqlc.HttpMethod
	httpHeaders             []byte
	httpBody                *string
	httpTimeoutMs           *int32
	httpIPFamily            *sqlc.IpFamily
	httpFollowRedirects     *bool
	httpSkipTLSVerify       *bool
	httpExpectedCodes       []int32
	httpExpectedClasses     []int32
	httpBodyContains        *string
	tracerouteProtocol      *sqlc.TracerouteProtocol
	tracerouteMaxHops       *int32
	tracerouteTimeoutMs     *int32
	tracerouteQueriesPerHop *int32
	traceroutePacketSize    *int32
	traceroutePort          *int32
	tracerouteIPFamily      *sqlc.IpFamily
}

func newCheck(id, projectID uuid.UUID, name string, checkType sqlc.CheckType, target string, selector []byte, description *string, intervalSeconds int32, createdAt, updatedAt time.Time, deletedAt *time.Time, configs checkConfigs) (domaincheck.Check, error) {
	httpConfig, err := mapOptionalHTTPConfig(configs.httpMethod, configs.httpHeaders, configs.httpBody, configs.httpTimeoutMs, configs.httpIPFamily, configs.httpFollowRedirects, configs.httpSkipTLSVerify, configs.httpExpectedCodes, configs.httpExpectedClasses, configs.httpBodyContains)
	if err != nil {
		return domaincheck.Check{}, err
	}
	return domaincheck.Check{
		ID:               id.String(),
		ProjectID:        projectID.String(),
		Name:             name,
		Type:             domaincheck.Type(checkType),
		Target:           target,
		Selector:         cloneRawMessage(selector),
		Description:      description,
		IntervalSeconds:  intervalSeconds,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		DeletedAt:        deletedAt,
		PingConfig:       mapOptionalPingConfig(configs.pingPacketCount, configs.pingPacketSizeBytes, configs.pingTimeoutMs, configs.pingIPFamily),
		TCPConfig:        mapOptionalTCPConfig(configs.tcpPort, configs.tcpTimeoutMs, configs.tcpIPFamily),
		HTTPConfig:       httpConfig,
		TracerouteConfig: mapOptionalTracerouteConfig(configs.tracerouteProtocol, configs.tracerouteMaxHops, configs.tracerouteTimeoutMs, configs.tracerouteQueriesPerHop, configs.traceroutePacketSize, configs.traceroutePort, configs.tracerouteIPFamily),
	}, nil
}

func matchingProbeIDs(selector domainselector.Selector, probes []activeProbeLabels) []uuid.UUID {
	probeIDs := make([]uuid.UUID, 0, len(probes))
	for _, probe := range probes {
		if probe.enabled && selector.Matches(probe.labels) {
			probeIDs = append(probeIDs, probe.probeID)
		}
	}

	return probeIDs
}

func mapProjectAssignments(rows []sqlc.ListProjectAssignmentsRow) []domainassignment.Assignment {
	assignments := make([]domainassignment.Assignment, 0, len(rows))
	for _, row := range rows {
		latitude, longitude := coordinatesFromPoint(row.ProbeLocation)
		checkType := domaincheck.Type(row.CheckType)
		target := row.Target
		if checkType == domaincheck.TypeHTTP {
			target = domainhttp.RedactTarget(target)
		}
		assignments = append(assignments, domainassignment.Assignment{
			ID:              row.AssignmentID.String(),
			ProjectID:       row.ProjectID.String(),
			ProbeID:         row.ProbeID.String(),
			CheckID:         row.CheckID.String(),
			ProbeStorageID:  row.ProbeInternalID,
			CheckStorageID:  row.CheckInternalID,
			CheckVersion:    row.CheckVersion,
			SelectorVersion: row.SelectorVersion,
			Probe: &domainprobe.Probe{
				ID:           row.ProbeID.String(),
				ProjectID:    row.ProjectID.String(),
				Name:         row.ProbeName,
				Enabled:      row.ProbeEnabled,
				LocationName: row.ProbeLocationName,
				Latitude:     latitude,
				Longitude:    longitude,
				Labels:       []domainlabel.Label{},
				CreatedAt:    row.ProbeCreatedAt,
				UpdatedAt:    row.ProbeUpdatedAt,
				DeletedAt:    row.ProbeDeletedAt,
			},
			Check: &domaincheck.Check{
				ID:               row.CheckID.String(),
				ProjectID:        row.ProjectID.String(),
				Name:             row.CheckName,
				Type:             checkType,
				Target:           target,
				Selector:         cloneRawMessage(row.Selector),
				Description:      row.Description,
				IntervalSeconds:  row.IntervalSeconds,
				Labels:           []domainlabel.Label{},
				CreatedAt:        row.CheckCreatedAt,
				UpdatedAt:        row.CheckUpdatedAt,
				DeletedAt:        row.CheckDeletedAt,
				PingConfig:       mapOptionalPingConfig(row.PingPacketCount, row.PingPacketSizeBytes, row.PingTimeoutMs, row.PingIpFamily),
				TCPConfig:        mapOptionalTCPConfig(row.TcpPort, row.TcpTimeoutMs, row.TcpIpFamily),
				TracerouteConfig: mapOptionalTracerouteConfig(row.TracerouteProtocol, row.TracerouteMaxHops, row.TracerouteTimeoutMs, row.TracerouteQueriesPerHop, row.TraceroutePacketSizeBytes, row.TraceroutePort, row.TracerouteIpFamily),
			},
		})
	}
	return assignments
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

func mapEnabledProbeLabel(row sqlc.ListActiveEnabledProbeLabelsForProjectRow) (domainlabel.Label, bool) {
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

func checkSelector(row sqlc.GetActiveCheckForProjectRow) (domainselector.Selector, json.RawMessage, error) {
	raw := json.RawMessage(row.Selector)
	selector, err := domainselector.Parse(raw)
	if err != nil {
		return domainselector.Selector{}, nil, err
	}

	return selector, raw, nil
}

func listCheckSelector(row sqlc.ListActiveChecksForProjectRow) (domainselector.Selector, json.RawMessage, error) {
	raw := json.RawMessage(row.Selector)
	selector, err := domainselector.Parse(raw)
	if err != nil {
		return domainselector.Selector{}, nil, err
	}

	return selector, raw, nil
}

func cloneRawMessage(value []byte) json.RawMessage {
	if value == nil {
		return nil
	}
	return append(json.RawMessage(nil), value...)
}

func coordinatesFromPoint(point pgtype.Point) (*float64, *float64) {
	if !point.Valid {
		return nil, nil
	}

	lon := point.P.X
	lat := point.P.Y
	return &lat, &lon
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

func mapOptionalHTTPConfig(method *sqlc.HttpMethod, headers []byte, body *string, timeoutMs *int32, family *sqlc.IpFamily, followRedirects, skipTLSVerify *bool, codes, classes []int32, bodyContains *string) (*domainhttp.Config, error) {
	if method == nil || timeoutMs == nil || followRedirects == nil || skipTLSVerify == nil {
		return nil, nil //nolint:nilnil // Nil means this joined row has no HTTP config.
	}
	var values []domainhttp.Header
	if err := json.Unmarshal(headers, &values); err != nil {
		return nil, fmt.Errorf("decode HTTP check headers: %w", err)
	}
	return &domainhttp.Config{
		Method: domainhttp.Method(*method), Headers: values, Body: body, TimeoutMs: *timeoutMs,
		IPFamily: mapIPFamily(family), FollowRedirects: *followRedirects,
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

func mapIPFamily(value *sqlc.IpFamily) *domainnetwork.IPFamily {
	if value == nil {
		return nil
	}

	ipFamily := domainnetwork.IPFamily(*value)
	return &ipFamily
}
