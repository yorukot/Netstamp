package pgprobe

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres/sqlc"
)

func mapProbe(row sqlc.Probe, labels []domainprobe.Label) domainprobe.Probe {
	if labels == nil {
		labels = []domainprobe.Label{}
	}

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
		Labels:    labels,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
		DeletedAt: timePtr(row.DeletedAt),
	}
}

func mapLabels(rows []sqlc.Label) []domainprobe.Label {
	labels := make([]domainprobe.Label, 0, len(rows))
	for _, row := range rows {
		labels = append(labels, domainprobe.Label{
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

func timePtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}
