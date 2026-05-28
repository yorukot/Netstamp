package publicpage

import (
	"context"

	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainpublicpage "github.com/yorukot/netstamp/internal/domain/publicpage"
)

type Repository interface {
	ListPages(ctx context.Context, projectID string) ([]domainpublicpage.Page, error)
	GetPageForProject(ctx context.Context, projectID, pageID string) (domainpublicpage.Page, error)
	GetEnabledPageBySlug(ctx context.Context, slug string) (domainpublicpage.Page, error)
	CreatePage(ctx context.Context, input domainpublicpage.Page) (domainpublicpage.Page, error)
	UpdatePage(ctx context.Context, input domainpublicpage.PageUpdate) (domainpublicpage.Page, error)
	SoftDeletePage(ctx context.Context, projectID, pageID string) error
	ListFolders(ctx context.Context, projectID, pageID string) ([]domainpublicpage.Folder, error)
	ListFolderChecks(ctx context.Context, projectID, pageID string) ([]domainpublicpage.PublishedCheck, error)
	ListPingPairs(ctx context.Context, projectID, pageID string) ([]domainpublicpage.PingPair, error)
	CreateFolder(ctx context.Context, projectID string, input domainpublicpage.Folder) (domainpublicpage.Folder, error)
	UpdateFolder(ctx context.Context, projectID string, input domainpublicpage.FolderUpdate) (domainpublicpage.Folder, error)
	DeleteFolder(ctx context.Context, projectID, pageID, folderID string) error
	SetFolderChecks(ctx context.Context, projectID, pageID, folderID string, checkIDs []string) ([]domainpublicpage.PublishedCheck, error)
	ResolvePublicPingPairProjectID(ctx context.Context, slug, probeID, checkID string) (string, error)
}

type ProjectAccess interface {
	GetProjectForUser(ctx context.Context, projectRef, userID string) (domainproject.Project, error)
	GetMemberRole(ctx context.Context, projectID, userID string) (domainproject.Role, error)
}

type PingInsightRepository interface {
	ListPingInsight(ctx context.Context, input domainping.InsightQuery) (domainping.InsightResult, error)
}
