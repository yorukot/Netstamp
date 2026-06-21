package publicstatus

import (
	"time"

	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
)

type ListPagesInput struct {
	CurrentUserID string
	ProjectRef    string
}

type GetPageInput struct {
	CurrentUserID string
	ProjectRef    string
	PageID        string
}

type CreatePageInput struct {
	CurrentUserID     string
	ProjectRef        string
	Slug              string
	Title             string
	Description       *string
	Enabled           bool
	DefaultChartMode  domainpublic.ChartMode
	DefaultChartRange domainpublic.ChartRange
}

type UpdatePageInput struct {
	CurrentUserID     string
	ProjectRef        string
	PageID            string
	Slug              string
	Title             string
	Description       *string
	Enabled           bool
	DefaultChartMode  domainpublic.ChartMode
	DefaultChartRange domainpublic.ChartRange
}

type DeletePageInput struct {
	CurrentUserID string
	ProjectRef    string
	PageID        string
}

type CreateElementInput struct {
	CurrentUserID   string
	ProjectRef      string
	PageID          string
	ParentElementID *string
	Kind            domainpublic.ElementKind
	CheckID         *string
	Title           *string
	Description     *string
	SortOrder       int32
	ChartMode       domainpublic.ChartMode
	ChartRange      *domainpublic.ChartRange
}

type UpdateElementInput struct {
	CurrentUserID   string
	ProjectRef      string
	PageID          string
	ElementID       string
	ParentElementID *string
	Kind            domainpublic.ElementKind
	CheckID         *string
	Title           *string
	Description     *string
	SortOrder       int32
	ChartMode       domainpublic.ChartMode
	ChartRange      *domainpublic.ChartRange
}

type DeleteElementInput struct {
	CurrentUserID string
	ProjectRef    string
	PageID        string
	ElementID     string
}

type PublicPageInput struct {
	Slug          string
	IncludeCharts bool
	Range         *domainpublic.ChartRange
	Now           time.Time
}

type PageDetail struct {
	Page     domainpublic.Page
	Elements []domainpublic.Element
}
