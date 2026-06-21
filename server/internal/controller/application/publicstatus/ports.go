package publicstatus

import (
	"context"

	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
)

type Repository interface {
	ListPages(ctx context.Context, projectID string) ([]domainpublic.Page, error)
	GetPage(ctx context.Context, projectID, pageID string) (domainpublic.Page, error)
	GetPageBySlug(ctx context.Context, slug string) (domainpublic.Page, error)
	CreatePage(ctx context.Context, input domainpublic.Page) (domainpublic.Page, error)
	UpdatePage(ctx context.Context, input domainpublic.Page) (domainpublic.Page, error)
	DeletePage(ctx context.Context, projectID, pageID string) error

	ListElements(ctx context.Context, pageID string) ([]domainpublic.Element, error)
	GetElement(ctx context.Context, projectID, pageID, elementID string) (domainpublic.Element, error)
	CreateElement(ctx context.Context, input domainpublic.Element) (domainpublic.Element, error)
	UpdateElement(ctx context.Context, input domainpublic.Element) (domainpublic.Element, error)
	DeleteElement(ctx context.Context, projectID, pageID, elementID string) error

	ListAssignments(ctx context.Context, pageID string) ([]domainpublic.Assignment, error)
	ListIncidents(ctx context.Context, pageID string, limit int32) ([]domainpublic.Incident, error)
}

type ProjectAccess interface {
	GetProjectForUser(ctx context.Context, projectRef, userID string) (domainproject.Project, error)
	GetMemberRole(ctx context.Context, projectID, userID string) (domainproject.Role, error)
}

type PingSeriesRepository interface {
	CountPingSeriesPoints(ctx context.Context, input domainping.SeriesPointCountQuery) (int64, error)
	ListPingSeries(ctx context.Context, input domainping.SeriesReadQuery) (map[string]domainping.SeriesData, error)
}

type TCPSeriesRepository interface {
	CountTCPSeriesPoints(ctx context.Context, input domaintcp.SeriesPointCountQuery) (int64, error)
	ListTCPSeries(ctx context.Context, input domaintcp.SeriesReadQuery) (map[string]domaintcp.SeriesData, error)
}
