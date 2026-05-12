package pgassignment

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
	domainselector "github.com/yorukot/netstamp/internal/domain/selector"
)

type activeProbeLabels struct {
	probeID uuid.UUID
	enabled bool
	labels  []domainlabel.Label
}

func activeProbeFromRows(rows []sqlc.GetActiveProbeRowsForProjectRow) (activeProbeLabels, bool) {
	if len(rows) == 0 {
		return activeProbeLabels{}, false
	}

	probe := activeProbeLabels{
		probeID: rows[0].ID,
		enabled: rows[0].Enabled,
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
				probeID: row.ProbeID,
				enabled: true,
			})
		}
		if label, ok := mapEnabledProbeLabel(row); ok {
			probes[index].labels = append(probes[index].labels, label)
		}
	}

	return probes
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

func mapEnabledProbeLabel(row sqlc.ListActiveEnabledProbeLabelsForProjectRow) (domainlabel.Label, bool) {
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
	config := domainping.Config{
		PacketCount:     row.PacketCount,
		PacketSizeBytes: row.PacketSizeBytes,
		TimeoutMs:       row.TimeoutMs,
		IPFamily:        mapIPFamily(row.IpFamily),
	}
	return domaincheck.Check{
		Type:            domaincheck.Type(row.CheckType),
		Target:          row.Target,
		IntervalSeconds: row.IntervalSeconds,
		PingConfig:      &config,
	}.CheckHash()
}

func listCheckVersion(row sqlc.ListActiveChecksForProjectRow) string {
	config := domainping.Config{
		PacketCount:     row.PacketCount,
		PacketSizeBytes: row.PacketSizeBytes,
		TimeoutMs:       row.TimeoutMs,
		IPFamily:        mapIPFamily(row.IpFamily),
	}
	return domaincheck.Check{
		Type:            domaincheck.Type(row.CheckType),
		Target:          row.Target,
		IntervalSeconds: row.IntervalSeconds,
		PingConfig:      &config,
	}.CheckHash()
}

func mapIPFamily(value sqlc.NullIpFamily) *domainnetwork.IPFamily {
	if !value.Valid {
		return nil
	}

	ipFamily := domainnetwork.IPFamily(value.IpFamily)
	return &ipFamily
}

func timePtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}
