package pgprobe

import (
	"net/netip"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres/sqlc"
)

func mapProbe(row sqlc.Probe) domainprobe.Probe {
	var latitude *float64
	var longitude *float64
	if row.Location.Valid {
		lon := row.Location.P.X
		lat := row.Location.P.Y
		latitude = &lat
		longitude = &lon
	}

	return domainprobe.Probe{
		ID:        row.ID.String(),
		ProjectID: row.ProjectID.String(),
		Name:      row.Name,
		Enabled:   row.Enabled,
		City:      row.City,
		Latitude:  latitude,
		Longitude: longitude,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
		DeletedAt: timePtr(row.DeletedAt),
	}
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
		Addrs:        append([]netip.Addr(nil), row.Addrs...),
		UpdatedAt:    row.UpdatedAt.Time,
	}
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
