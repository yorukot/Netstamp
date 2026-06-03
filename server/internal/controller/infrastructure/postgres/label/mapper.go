package pglabel

import (
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
)

func mapLabel(row sqlc.Label) domainlabel.Label {
	return domainlabel.Label{
		ID:        row.ID.String(),
		ProjectID: row.ProjectID.String(),
		Key:       row.Key,
		Value:     row.Value,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		DeletedAt: row.DeletedAt,
	}
}

func mapLabels(rows []sqlc.Label) []domainlabel.Label {
	labels := make([]domainlabel.Label, 0, len(rows))
	for _, row := range rows {
		labels = append(labels, mapLabel(row))
	}

	return labels
}
