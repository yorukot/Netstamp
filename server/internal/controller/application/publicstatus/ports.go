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
	HasAssignableCheck(ctx context.Context, projectID, checkID string) (bool, error)
	CountAssignableAssignments(ctx context.Context, projectID string, assignmentIDs []string) (int64, error)

	ListAssignments(ctx context.Context, pageID string) ([]domainpublic.Assignment, error)
	ListElementAssignments(ctx context.Context, pageID, elementID string) ([]domainpublic.Assignment, error)
	ListIncidents(ctx context.Context, pageID string, limit int32) ([]domainpublic.Incident, error)
}

type ProjectAccess interface {
	GetProjectForUser(ctx context.Context, projectRef, userID string) (domainproject.Project, error)
	GetMemberRole(ctx context.Context, projectID, userID string) (domainproject.Role, error)
}

type EventRecorder interface {
	RecordPublicStatusEvent(ctx context.Context, event PublicStatusEvent)
}

type PingSeriesRepository interface {
	CountPingSeriesPoints(ctx context.Context, input domainping.SeriesPointCountQuery) (int64, error)
	ListPingSeries(ctx context.Context, input domainping.SeriesReadQuery) (map[string]domainping.SeriesData, error)
}

type TCPSeriesRepository interface {
	CountTCPSeriesPoints(ctx context.Context, input domaintcp.SeriesPointCountQuery) (int64, error)
	ListTCPSeries(ctx context.Context, input domaintcp.SeriesReadQuery) (map[string]domaintcp.SeriesData, error)
}

type PublicStatusAction string

const (
	PublicStatusActionCreatePage    PublicStatusAction = "page.create"
	PublicStatusActionUpdatePage    PublicStatusAction = "page.update"
	PublicStatusActionDeletePage    PublicStatusAction = "page.delete"
	PublicStatusActionCreateElement PublicStatusAction = "element.create"
	PublicStatusActionUpdateElement PublicStatusAction = "element.update"
	PublicStatusActionDeleteElement PublicStatusAction = "element.delete"
)

type PublicStatusOutcome string

const (
	PublicStatusOutcomeSuccess PublicStatusOutcome = "success"
	PublicStatusOutcomeFailure PublicStatusOutcome = "failure"
)

type PublicStatusReason string

const (
	PublicStatusReasonInvalidInput        PublicStatusReason = "invalid_input"
	PublicStatusReasonForbidden           PublicStatusReason = "forbidden"
	PublicStatusReasonProjectNotFound     PublicStatusReason = "project_not_found"
	PublicStatusReasonUserNotFound        PublicStatusReason = "user_not_found"
	PublicStatusReasonPageNotFound        PublicStatusReason = "page_not_found"
	PublicStatusReasonElementNotFound     PublicStatusReason = "element_not_found"
	PublicStatusReasonSlugAlreadyExists   PublicStatusReason = "slug_already_exists"
	PublicStatusReasonProjectLookupFailed PublicStatusReason = "project_lookup_failed"
	PublicStatusReasonRoleLookupFailed    PublicStatusReason = "role_lookup_failed"
	PublicStatusReasonPageCreateFailed    PublicStatusReason = "page_create_failed"
	PublicStatusReasonPageUpdateFailed    PublicStatusReason = "page_update_failed"
	PublicStatusReasonPageDeleteFailed    PublicStatusReason = "page_delete_failed"
	PublicStatusReasonElementCreateFailed PublicStatusReason = "element_create_failed"
	PublicStatusReasonElementUpdateFailed PublicStatusReason = "element_update_failed"
	PublicStatusReasonElementDeleteFailed PublicStatusReason = "element_delete_failed"
)

type PublicStatusEventName string

const (
	PublicStatusEventCreatePageSuccess    PublicStatusEventName = "public_status.page.create.success"
	PublicStatusEventCreatePageFailure    PublicStatusEventName = "public_status.page.create.failure"
	PublicStatusEventUpdatePageSuccess    PublicStatusEventName = "public_status.page.update.success"
	PublicStatusEventUpdatePageFailure    PublicStatusEventName = "public_status.page.update.failure"
	PublicStatusEventDeletePageSuccess    PublicStatusEventName = "public_status.page.delete.success"
	PublicStatusEventDeletePageFailure    PublicStatusEventName = "public_status.page.delete.failure"
	PublicStatusEventCreateElementSuccess PublicStatusEventName = "public_status.element.create.success"
	PublicStatusEventCreateElementFailure PublicStatusEventName = "public_status.element.create.failure"
	PublicStatusEventUpdateElementSuccess PublicStatusEventName = "public_status.element.update.success"
	PublicStatusEventUpdateElementFailure PublicStatusEventName = "public_status.element.update.failure"
	PublicStatusEventDeleteElementSuccess PublicStatusEventName = "public_status.element.delete.success"
	PublicStatusEventDeleteElementFailure PublicStatusEventName = "public_status.element.delete.failure"
)

type PublicStatusEvent struct {
	Name        PublicStatusEventName
	Action      PublicStatusAction
	Outcome     PublicStatusOutcome
	Reason      PublicStatusReason
	ActorUserID string
	ProjectID   string
	ProjectRef  string
	ProjectSlug string
	PageID      string
	ElementID   string
	Err         error
}
