package pglabel

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	"github.com/yorukot/netstamp/internal/infrastructure/postgres/sqlc"
)

func mapLabel(row sqlc.Label) domainlabel.Label {
	return domainlabel.Label{
		ID:        row.ID.String(),
		ProjectID: row.ProjectID.String(),
		Key:       row.Key,
		Value:     row.Value,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
		DeletedAt: timePtr(row.DeletedAt),
	}
}

func mapLabels(rows []sqlc.Label) []domainlabel.Label {
	labels := make([]domainlabel.Label, 0, len(rows))
	for _, row := range rows {
		labels = append(labels, mapLabel(row))
	}

	return labels
}

func timePtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}
