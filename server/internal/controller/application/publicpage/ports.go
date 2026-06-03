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
	CountPingSeriesPoints(ctx context.Context, input domainping.SeriesPointCountQuery) (int64, error)
	GetPingInsightSummary(ctx context.Context, input domainping.InsightSummaryQuery) (domainping.InsightSummary, error)
}

type EventRecorder interface {
	RecordPublicPageEvent(ctx context.Context, event PublicPageEvent)
}

type PublicPageEventName string

const (
	PublicPageEventListFailure            PublicPageEventName = "public_page.list.failure"
	PublicPageEventGetFailure             PublicPageEventName = "public_page.get.failure"
	PublicPageEventCreateSuccess          PublicPageEventName = "public_page.create.success"
	PublicPageEventCreateFailure          PublicPageEventName = "public_page.create.failure"
	PublicPageEventUpdateSuccess          PublicPageEventName = "public_page.update.success"
	PublicPageEventUpdateFailure          PublicPageEventName = "public_page.update.failure"
	PublicPageEventDeleteSuccess          PublicPageEventName = "public_page.delete.success"
	PublicPageEventDeleteFailure          PublicPageEventName = "public_page.delete.failure"
	PublicPageEventCreateFolderSuccess    PublicPageEventName = "public_page.folder.create.success"
	PublicPageEventCreateFolderFailure    PublicPageEventName = "public_page.folder.create.failure"
	PublicPageEventUpdateFolderSuccess    PublicPageEventName = "public_page.folder.update.success"
	PublicPageEventUpdateFolderFailure    PublicPageEventName = "public_page.folder.update.failure"
	PublicPageEventDeleteFolderSuccess    PublicPageEventName = "public_page.folder.delete.success"
	PublicPageEventDeleteFolderFailure    PublicPageEventName = "public_page.folder.delete.failure"
	PublicPageEventSetFolderChecksSuccess PublicPageEventName = "public_page.folder_checks.set.success"
	PublicPageEventSetFolderChecksFailure PublicPageEventName = "public_page.folder_checks.set.failure"
	PublicPageEventPublicGetFailure       PublicPageEventName = "public_page.public_get.failure"
	PublicPageEventPingInsightFailure     PublicPageEventName = "public_page.ping_insight.failure"
)

type PublicPageEventAction string

const (
	PublicPageActionList            PublicPageEventAction = "list"
	PublicPageActionGet             PublicPageEventAction = "get"
	PublicPageActionCreate          PublicPageEventAction = "create"
	PublicPageActionUpdate          PublicPageEventAction = "update"
	PublicPageActionDelete          PublicPageEventAction = "delete"
	PublicPageActionCreateFolder    PublicPageEventAction = "create_folder"
	PublicPageActionUpdateFolder    PublicPageEventAction = "update_folder"
	PublicPageActionDeleteFolder    PublicPageEventAction = "delete_folder"
	PublicPageActionSetFolderChecks PublicPageEventAction = "set_folder_checks"
	PublicPageActionPublicGet       PublicPageEventAction = "public_get"
	PublicPageActionPingInsight     PublicPageEventAction = "ping_insight"
)

type PublicPageEventOutcome string

const (
	PublicPageOutcomeSuccess PublicPageEventOutcome = "success"
	PublicPageOutcomeFailure PublicPageEventOutcome = "failure"
)

type PublicPageEventReason string

const (
	PublicPageReasonInvalidInput                PublicPageEventReason = "invalid_input"
	PublicPageReasonForbidden                   PublicPageEventReason = "forbidden"
	PublicPageReasonProjectNotFound             PublicPageEventReason = "project_not_found"
	PublicPageReasonUserNotFound                PublicPageEventReason = "user_not_found"
	PublicPageReasonPageNotFound                PublicPageEventReason = "page_not_found"
	PublicPageReasonFolderNotFound              PublicPageEventReason = "folder_not_found"
	PublicPageReasonCheckNotFound               PublicPageEventReason = "check_not_found"
	PublicPageReasonProbeNotFound               PublicPageEventReason = "probe_not_found"
	PublicPageReasonCheckNotPublished           PublicPageEventReason = "check_not_published"
	PublicPageReasonDuplicateSlug               PublicPageEventReason = "duplicate_slug"
	PublicPageReasonCheckAlreadyPublished       PublicPageEventReason = "check_already_published"
	PublicPageReasonProjectLookupFailed         PublicPageEventReason = "project_lookup_failed"
	PublicPageReasonRoleLookupFailed            PublicPageEventReason = "role_lookup_failed"
	PublicPageReasonPageListFailed              PublicPageEventReason = "page_list_failed"
	PublicPageReasonPageLookupFailed            PublicPageEventReason = "page_lookup_failed"
	PublicPageReasonPageCreateFailed            PublicPageEventReason = "page_create_failed"
	PublicPageReasonPageUpdateFailed            PublicPageEventReason = "page_update_failed"
	PublicPageReasonPageDeleteFailed            PublicPageEventReason = "page_delete_failed"
	PublicPageReasonFolderListFailed            PublicPageEventReason = "folder_list_failed"
	PublicPageReasonFolderChecksListFailed      PublicPageEventReason = "folder_checks_list_failed"
	PublicPageReasonPingPairsListFailed         PublicPageEventReason = "ping_pairs_list_failed"
	PublicPageReasonFolderCreateFailed          PublicPageEventReason = "folder_create_failed"
	PublicPageReasonFolderUpdateFailed          PublicPageEventReason = "folder_update_failed"
	PublicPageReasonFolderDeleteFailed          PublicPageEventReason = "folder_delete_failed"
	PublicPageReasonFolderChecksUpdateFailed    PublicPageEventReason = "folder_checks_update_failed"
	PublicPageReasonPublicPairLookupFailed      PublicPageEventReason = "public_pair_lookup_failed"
	PublicPageReasonPingRepositoryNotConfigured PublicPageEventReason = "ping_repository_not_configured"
	PublicPageReasonPingInsightQueryFailed      PublicPageEventReason = "ping_insight_query_failed"
)

type PublicPageEvent struct {
	Name        PublicPageEventName
	Action      PublicPageEventAction
	Outcome     PublicPageEventOutcome
	Reason      PublicPageEventReason
	ActorUserID string
	ProjectID   string
	ProjectRef  string
	ProjectSlug string
	PageID      string
	PageSlug    string
	FolderID    string
	CheckID     string
	ProbeID     string
	CheckCount  int
	Err         error
}
