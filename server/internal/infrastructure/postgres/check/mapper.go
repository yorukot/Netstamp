package pgcheck

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
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
		IntervalSeconds: int(row.IntervalSeconds),
		PingConfig:      mapPingConfig(config),
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		DeletedAt:       timePtr(row.DeletedAt),
	}
}

func mapListCheck(row sqlc.ListActiveChecksForProjectRow) domaincheck.Check {
	return domaincheck.Check{
		ID:              row.ID.String(),
		ProjectID:       row.ProjectID.String(),
		Name:            row.Name,
		Type:            domaincheck.Type(row.CheckType),
		Target:          row.Target,
		Selector:        cloneRawMessage(row.Selector),
		Description:     row.Description,
		IntervalSeconds: int(row.IntervalSeconds),
		PingConfig: domaincheck.PingConfig{
			PacketCount:     int(row.PacketCount),
			PacketSizeBytes: int(row.PacketSizeBytes),
			TimeoutMs:       int(row.TimeoutMs),
			IPFamily:        mapIPFamily(row.IpFamily),
		},
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
		DeletedAt: timePtr(row.DeletedAt),
	}
}

func mapGetCheck(row sqlc.GetActiveCheckForProjectRow) domaincheck.Check {
	return domaincheck.Check{
		ID:              row.ID.String(),
		ProjectID:       row.ProjectID.String(),
		Name:            row.Name,
		Type:            domaincheck.Type(row.CheckType),
		Target:          row.Target,
		Selector:        cloneRawMessage(row.Selector),
		Description:     row.Description,
		IntervalSeconds: int(row.IntervalSeconds),
		PingConfig: domaincheck.PingConfig{
			PacketCount:     int(row.PacketCount),
			PacketSizeBytes: int(row.PacketSizeBytes),
			TimeoutMs:       int(row.TimeoutMs),
			IPFamily:        mapIPFamily(row.IpFamily),
		},
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
		DeletedAt: timePtr(row.DeletedAt),
	}
}

func mapPingConfig(row sqlc.PingCheckConfig) domaincheck.PingConfig {
	return domaincheck.PingConfig{
		PacketCount:     int(row.PacketCount),
		PacketSizeBytes: int(row.PacketSizeBytes),
		TimeoutMs:       int(row.TimeoutMs),
		IPFamily:        mapIPFamily(row.IpFamily),
	}
}

func sqlcCheckType(value domaincheck.Type) sqlc.CheckType {
	return sqlc.CheckType(value)
}

func sqlcIPFamily(value *domaincheck.IPFamily) sqlc.NullIpFamily {
	if value == nil {
		return sqlc.NullIpFamily{}
	}

	return sqlc.NullIpFamily{
		IpFamily: sqlc.IpFamily(*value),
		Valid:    true,
	}
}

func mapIPFamily(value sqlc.NullIpFamily) *domaincheck.IPFamily {
	if !value.Valid {
		return nil
	}

	ipFamily := domaincheck.IPFamily(value.IpFamily)
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
