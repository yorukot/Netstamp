package pgprobe

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

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

func timePtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}
