package pgprobe

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

func mapProbe(row sqlc.Probe) domainprobe.Probe {
	return mapProbeFields(
		row.ID,
		row.ProjectID,
		row.Name,
		row.Enabled,
		row.Location,
		row.SubdivisionCode,
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
				row.SubdivisionCode,
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
		first.SubdivisionCode,
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
	subdivisionCode *string,
	createdAt pgtype.Timestamptz,
	updatedAt pgtype.Timestamptz,
	deletedAt pgtype.Timestamptz,
	status *domainprobe.Status,
) domainprobe.Probe {
	latitude, longitude := coordinatesFromPoint(location)

	return domainprobe.Probe{
		ID:              id.String(),
		ProjectID:       projectID.String(),
		Name:            name,
		Enabled:         enabled,
		SubdivisionCode: subdivisionCode,
		Latitude:        latitude,
		Longitude:       longitude,
		Status:          status,
		CreatedAt:       createdAt.Time,
		UpdatedAt:       updatedAt.Time,
		DeletedAt:       timePtr(deletedAt),
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

func mapProbeStatus(row sqlc.ProbeStatus) domainprobe.Status {
	return domainprobe.Status{
		ProbeID:      row.ProbeID.String(),
		State:        domainprobe.State(row.Status),
		LastSeenAt:   timePtr(row.LastSeenAt),
		AgentVersion: row.AgentVersion,
		PublicV4:     cloneAddr(row.PublicV4),
		PublicV6:     cloneAddr(row.PublicV6),
		AS:           row.As,
		Addrs:        append([]netip.Addr(nil), row.Addrs...),
		UpdatedAt:    row.UpdatedAt.Time,
	}
}

func mapListProbeStatus(row sqlc.ListActiveProbesForProjectRow) *domainprobe.Status {
	if !row.Status.Valid {
		return nil
	}

	return &domainprobe.Status{
		ProbeID:      row.ID.String(),
		State:        domainprobe.State(row.Status.ProbeState),
		LastSeenAt:   timePtr(row.StatusLastSeenAt),
		AgentVersion: row.StatusAgentVersion,
		PublicV4:     cloneAddr(row.StatusPublicV4),
		PublicV6:     cloneAddr(row.StatusPublicV6),
		AS:           row.StatusAs,
		Addrs:        append([]netip.Addr(nil), row.StatusAddrs...),
		UpdatedAt:    row.StatusUpdatedAt.Time,
	}
}

func mapGetProbeStatus(row sqlc.GetActiveProbeRowsForProjectRow) *domainprobe.Status {
	if !row.Status.Valid {
		return nil
	}

	return &domainprobe.Status{
		ProbeID:      row.ID.String(),
		State:        domainprobe.State(row.Status.ProbeState),
		LastSeenAt:   timePtr(row.StatusLastSeenAt),
		AgentVersion: row.StatusAgentVersion,
		PublicV4:     cloneAddr(row.StatusPublicV4),
		PublicV6:     cloneAddr(row.StatusPublicV6),
		AS:           row.StatusAs,
		Addrs:        append([]netip.Addr(nil), row.StatusAddrs...),
		UpdatedAt:    row.StatusUpdatedAt.Time,
	}
}

func mapListProbeLabel(row sqlc.ListActiveProbesForProjectRow) (domainlabel.Label, bool) {
	if row.LabelID == nil || row.LabelProjectID == nil || row.LabelKey == nil || row.LabelValue == nil {
		return domainlabel.Label{}, false
	}

	return domainlabel.Label{
		ID:        row.LabelID.String(),
		ProjectID: row.LabelProjectID.String(),
		Key:       *row.LabelKey,
		Value:     *row.LabelValue,
		CreatedAt: row.LabelCreatedAt.Time,
		UpdatedAt: row.LabelUpdatedAt.Time,
		DeletedAt: timePtr(row.LabelDeletedAt),
	}, true
}

func mapGetProbeLabel(row sqlc.GetActiveProbeRowsForProjectRow) (domainlabel.Label, bool) {
	if row.LabelID == nil || row.LabelProjectID == nil || row.LabelKey == nil || row.LabelValue == nil {
		return domainlabel.Label{}, false
	}

	return domainlabel.Label{
		ID:        row.LabelID.String(),
		ProjectID: row.LabelProjectID.String(),
		Key:       *row.LabelKey,
		Value:     *row.LabelValue,
		CreatedAt: row.LabelCreatedAt.Time,
		UpdatedAt: row.LabelUpdatedAt.Time,
		DeletedAt: timePtr(row.LabelDeletedAt),
	}, true
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

func mapAssignment(row sqlc.ListActiveAssignmentsForProbeRow) domaincheck.Assignment {
	return domaincheck.Assignment{
		ID:              row.AssignmentID.String(),
		ProjectID:       row.ProjectID.String(),
		ProbeID:         row.ProbeID.String(),
		CheckID:         row.CheckID.String(),
		CheckVersion:    row.CheckVersion,
		SelectorVersion: row.SelectorVersion,
		Type:            domaincheck.Type(row.CheckType),
		Target:          row.Target,
		IntervalSeconds: row.IntervalSeconds,
		PingConfig: domainping.Config{
			PacketCount:     row.PacketCount,
			PacketSizeBytes: row.PacketSizeBytes,
			TimeoutMs:       row.TimeoutMs,
			IPFamily:        mapIPFamily(row.IpFamily),
		},
	}
}

func sqlcProbeState(value domainprobe.State) sqlc.ProbeState {
	return sqlc.ProbeState(value)
}

func mapIPFamily(value sqlc.NullIpFamily) *domainnetwork.IPFamily {
	if !value.Valid {
		return nil
	}

	ipFamily := domainnetwork.IPFamily(value.IpFamily)
	return &ipFamily
}

func cloneAddr(value *netip.Addr) *netip.Addr {
	if value == nil {
		return nil
	}

	addr := *value
	return &addr
}

func timePtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}
