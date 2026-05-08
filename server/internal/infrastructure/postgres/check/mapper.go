package pgcheck

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres/sqlc"
)

func mapStoredCheck(row sqlc.Check, config sqlc.PingCheckConfig) domaincheck.Check {
	return domaincheck.Check{
		ID:              row.ID.String(),
		ProjectID:       row.ProjectID.String(),
		Name:            row.Name,
		Type:            domaincheck.Type(row.CheckType),
		Target:          row.Target,
		Selector:        cloneRawMessage(row.Selector),
		Description:     row.Description,
		IntervalSeconds: row.IntervalSeconds,
		PingConfig:      mapPingConfig(config),
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		DeletedAt:       timePtr(row.DeletedAt),
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
		row.PacketCount,
		row.PacketSizeBytes,
		row.TimeoutMs,
		row.IpFamily,
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
		row.PacketCount,
		row.PacketSizeBytes,
		row.TimeoutMs,
		row.IpFamily,
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
	packetCount int32,
	packetSizeBytes int32,
	timeoutMs int32,
	ipFamily sqlc.NullIpFamily,
	createdAt pgtype.Timestamptz,
	updatedAt pgtype.Timestamptz,
	deletedAt pgtype.Timestamptz,
) domaincheck.Check {
	return domaincheck.Check{
		ID:              id.String(),
		ProjectID:       projectID.String(),
		Name:            name,
		Type:            domaincheck.Type(checkType),
		Target:          target,
		Selector:        cloneRawMessage(selector),
		Description:     description,
		IntervalSeconds: intervalSeconds,
		PingConfig: domainping.Config{
			PacketCount:     packetCount,
			PacketSizeBytes: packetSizeBytes,
			TimeoutMs:       timeoutMs,
			IPFamily:        mapIPFamily(ipFamily),
		},
		CreatedAt: createdAt.Time,
		UpdatedAt: updatedAt.Time,
		DeletedAt: timePtr(deletedAt),
	}
}

func mapPingConfig(row sqlc.PingCheckConfig) domainping.Config {
	return domainping.Config{
		PacketCount:     row.PacketCount,
		PacketSizeBytes: row.PacketSizeBytes,
		TimeoutMs:       row.TimeoutMs,
		IPFamily:        mapIPFamily(row.IpFamily),
	}
}

func sqlcCheckType(value domaincheck.Type) sqlc.CheckType {
	return sqlc.CheckType(value)
}

func sqlcIPFamily(value *domainnetwork.IPFamily) sqlc.NullIpFamily {
	if value == nil {
		return sqlc.NullIpFamily{}
	}

	return sqlc.NullIpFamily{
		IpFamily: sqlc.IpFamily(*value),
		Valid:    true,
	}
}

func mapIPFamily(value sqlc.NullIpFamily) *domainnetwork.IPFamily {
	if !value.Valid {
		return nil
	}

	ipFamily := domainnetwork.IPFamily(value.IpFamily)
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
