package pgpublicpage

import (
	"time"

	"github.com/google/uuid"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	domainpublicpage "github.com/yorukot/netstamp/internal/domain/publicpage"
)

func mapPage(row sqlc.PublicPage) domainpublicpage.Page {
	return domainpublicpage.Page{
		ID:          row.ID.String(),
		ProjectID:   row.ProjectID.String(),
		Slug:        row.Slug,
		Title:       row.Title,
		Description: row.Description,
		Enabled:     row.Enabled,
		CreatedAt:   row.CreatedAt.UTC(),
		UpdatedAt:   row.UpdatedAt.UTC(),
		DeletedAt:   utcTimePtr(row.DeletedAt),
	}
}

func mapPages(rows []sqlc.PublicPage) []domainpublicpage.Page {
	pages := make([]domainpublicpage.Page, 0, len(rows))
	for _, row := range rows {
		pages = append(pages, mapPage(row))
	}
	return pages
}

func mapFolder(row sqlc.PublicPageFolder) domainpublicpage.Folder {
	return domainpublicpage.Folder{
		ID:          row.ID.String(),
		PageID:      row.PublicPageID.String(),
		ParentID:    uuidPtrString(row.ParentID),
		Name:        row.Name,
		Description: row.Description,
		SortOrder:   row.SortOrder,
		CreatedAt:   row.CreatedAt.UTC(),
		UpdatedAt:   row.UpdatedAt.UTC(),
	}
}

func mapFolders(rows []sqlc.PublicPageFolder) []domainpublicpage.Folder {
	folders := make([]domainpublicpage.Folder, 0, len(rows))
	for _, row := range rows {
		folders = append(folders, mapFolder(row))
	}
	return folders
}

func mapPublishedChecks(rows []sqlc.ListPublicPageFolderChecksForProjectPageRow) []domainpublicpage.PublishedCheck {
	checks := make([]domainpublicpage.PublishedCheck, 0, len(rows))
	for _, row := range rows {
		checks = append(checks, domainpublicpage.PublishedCheck{
			ID:              row.CheckID.String(),
			FolderID:        row.FolderID.String(),
			Name:            row.CheckName,
			Description:     row.CheckDescription,
			IntervalSeconds: row.CheckIntervalSeconds,
			SortOrder:       row.SortOrder,
			CreatedAt:       row.CheckCreatedAt.UTC(),
			UpdatedAt:       row.CheckUpdatedAt.UTC(),
		})
	}
	return checks
}

func mapPingPairs(rows []sqlc.ListPublicPagePingPairsRow) []domainpublicpage.PingPair {
	pairs := make([]domainpublicpage.PingPair, 0, len(rows))
	for _, row := range rows {
		pairs = append(pairs, domainpublicpage.PingPair{
			FolderID:             row.FolderID.String(),
			ProbeID:              row.ProbeID.String(),
			ProbeName:            row.ProbeName,
			ProbeLocationName:    row.ProbeLocationName,
			ProbeStatus:          string(row.ProbeStatus),
			CheckID:              row.CheckID.String(),
			CheckName:            row.CheckName,
			CheckDescription:     row.CheckDescription,
			CheckIntervalSeconds: row.CheckIntervalSeconds,
		})
	}
	return pairs
}

func uuidPtrString(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}
	copied := value.String()
	return &copied
}

func utcTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := value.UTC()
	return &copied
}
