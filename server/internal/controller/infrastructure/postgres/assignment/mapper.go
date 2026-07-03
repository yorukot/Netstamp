package pgassignment

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
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
		check := listCheck(row)
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
	check := getCheck(row)
	return domainassignment.CheckAssignmentCandidate{
		Check:           check,
		Selector:        selector,
		SelectorVersion: domaincheck.SelectorVersion(selectorRaw),
		CheckVersion:    check.Hash(),
	}, nil
}

func listCheck(row sqlc.ListActiveChecksForProjectRow) domaincheck.Check {
	return domaincheck.Check{
		ID:               row.ID.String(),
		ProjectID:        row.ProjectID.String(),
		Name:             row.Name,
		Type:             domaincheck.Type(row.CheckType),
		Target:           row.Target,
		Selector:         cloneRawMessage(row.Selector),
		Description:      row.Description,
		IntervalSeconds:  row.IntervalSeconds,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
		DeletedAt:        row.DeletedAt,
		PingConfig:       mapOptionalPingConfig(row.PingPacketCount, row.PingPacketSizeBytes, row.PingTimeoutMs, row.PingIpFamily),
		TCPConfig:        mapOptionalTCPConfig(row.TcpPort, row.TcpTimeoutMs, row.TcpIpFamily),
		TracerouteConfig: mapOptionalTracerouteConfig(row.TracerouteProtocol, row.TracerouteMaxHops, row.TracerouteTimeoutMs, row.TracerouteQueriesPerHop, row.TraceroutePacketSizeBytes, row.TraceroutePort, row.TracerouteIpFamily),
	}
}

func getCheck(row sqlc.GetActiveCheckForProjectRow) domaincheck.Check {
	return domaincheck.Check{
		ID:               row.ID.String(),
		ProjectID:        row.ProjectID.String(),
		Name:             row.Name,
		Type:             domaincheck.Type(row.CheckType),
		Target:           row.Target,
		Selector:         cloneRawMessage(row.Selector),
		Description:      row.Description,
		IntervalSeconds:  row.IntervalSeconds,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
		DeletedAt:        row.DeletedAt,
		PingConfig:       mapOptionalPingConfig(row.PingPacketCount, row.PingPacketSizeBytes, row.PingTimeoutMs, row.PingIpFamily),
		TCPConfig:        mapOptionalTCPConfig(row.TcpPort, row.TcpTimeoutMs, row.TcpIpFamily),
		TracerouteConfig: mapOptionalTracerouteConfig(row.TracerouteProtocol, row.TracerouteMaxHops, row.TracerouteTimeoutMs, row.TracerouteQueriesPerHop, row.TraceroutePacketSizeBytes, row.TraceroutePort, row.TracerouteIpFamily),
	}
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

func matchingPreviewProbes(selector domainselector.Selector, probes []activeProbeLabels) []domainprobe.Probe {
	matches := make([]domainprobe.Probe, 0, len(probes))
	for _, probe := range probes {
		if !probe.enabled || !selector.Matches(probe.labels) {
			continue
		}
		matches = append(matches, domainprobe.Probe{
			ID:        probe.probeID.String(),
			ProjectID: probe.projectID.String(),
			Name:      probe.name,
			Enabled:   probe.enabled,
			Labels:    append([]domainlabel.Label(nil), probe.labels...),
		})
	}
	return matches
}

func mapProjectAssignments(rows []sqlc.ListProjectAssignmentsRow) []domainassignment.Assignment {
	assignments := make([]domainassignment.Assignment, 0, len(rows))
	for _, row := range rows {
		latitude, longitude := coordinatesFromPoint(row.ProbeLocation)
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
				Type:             domaincheck.Type(row.CheckType),
				Target:           row.Target,
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

func checkVersion(row sqlc.GetActiveCheckForProjectRow) string {
	return domaincheck.Check{
		Type:             domaincheck.Type(row.CheckType),
		Target:           row.Target,
		IntervalSeconds:  row.IntervalSeconds,
		PingConfig:       mapOptionalPingConfig(row.PingPacketCount, row.PingPacketSizeBytes, row.PingTimeoutMs, row.PingIpFamily),
		TCPConfig:        mapOptionalTCPConfig(row.TcpPort, row.TcpTimeoutMs, row.TcpIpFamily),
		TracerouteConfig: mapOptionalTracerouteConfig(row.TracerouteProtocol, row.TracerouteMaxHops, row.TracerouteTimeoutMs, row.TracerouteQueriesPerHop, row.TraceroutePacketSizeBytes, row.TraceroutePort, row.TracerouteIpFamily),
	}.Hash()
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

func listCheckVersion(row sqlc.ListActiveChecksForProjectRow) string {
	return domaincheck.Check{
		Type:             domaincheck.Type(row.CheckType),
		Target:           row.Target,
		IntervalSeconds:  row.IntervalSeconds,
		PingConfig:       mapOptionalPingConfig(row.PingPacketCount, row.PingPacketSizeBytes, row.PingTimeoutMs, row.PingIpFamily),
		TCPConfig:        mapOptionalTCPConfig(row.TcpPort, row.TcpTimeoutMs, row.TcpIpFamily),
		TracerouteConfig: mapOptionalTracerouteConfig(row.TracerouteProtocol, row.TracerouteMaxHops, row.TracerouteTimeoutMs, row.TracerouteQueriesPerHop, row.TraceroutePacketSizeBytes, row.TraceroutePort, row.TracerouteIpFamily),
	}.Hash()
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

func mapIPFamily(value *sqlc.IpFamily) *domainnetwork.IPFamily {
	if value == nil {
		return nil
	}

	ipFamily := domainnetwork.IPFamily(*value)
	return &ipFamily
}
