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
	CurrentUserID       string
	ProjectRef          string
	Slug                string
	Title               string
	Description         *string
	Enabled             bool
	FooterText          *string
	BannerImageURL      *string
	Theme               domainpublic.Theme
	ShowTargets         bool
	ShowProbeNames      bool
	ShowProbeLocations  bool
	ShowIncidentHistory bool
	ShowGeneratedAt     bool
	CustomCSS           *string
	DefaultChartMode    domainpublic.ChartMode
	DefaultChartRange   domainpublic.ChartRange
}

type UpdatePageInput struct {
	CurrentUserID       string
	ProjectRef          string
	PageID              string
	Slug                string
	Title               string
	Description         *string
	Enabled             bool
	FooterText          *string
	BannerImageURL      *string
	Theme               domainpublic.Theme
	ShowTargets         bool
	ShowProbeNames      bool
	ShowProbeLocations  bool
	ShowIncidentHistory bool
	ShowGeneratedAt     bool
	CustomCSS           *string
	DefaultChartMode    domainpublic.ChartMode
	DefaultChartRange   domainpublic.ChartRange
}

type DeletePageInput struct {
	CurrentUserID string
	ProjectRef    string
	PageID        string
}

type CreateElementInput struct {
	CurrentUserID           string
	ProjectRef              string
	PageID                  string
	ParentElementID         *string
	Kind                    domainpublic.ElementKind
	CheckID                 *string
	AssignmentSelectionMode domainpublic.AssignmentSelectionMode
	AssignmentIDs           []string
	Title                   *string
	Description             *string
	SortOrder               int32
	DisplayMode             domainpublic.ElementDisplayMode
	ChartMode               domainpublic.ChartMode
	ChartRange              *domainpublic.ChartRange
}

type UpdateElementInput struct {
	CurrentUserID           string
	ProjectRef              string
	PageID                  string
	ElementID               string
	ParentElementID         *string
	Kind                    domainpublic.ElementKind
	CheckID                 *string
	AssignmentSelectionMode domainpublic.AssignmentSelectionMode
	AssignmentIDs           []string
	Title                   *string
	Description             *string
	SortOrder               int32
	DisplayMode             domainpublic.ElementDisplayMode
	ChartMode               domainpublic.ChartMode
	ChartRange              *domainpublic.ChartRange
}

type DeleteElementInput struct {
	CurrentUserID string
	ProjectRef    string
	PageID        string
	ElementID     string
}

type PublicSummaryInput struct {
	Slug string
	Now  time.Time
}

type EditorContextInput struct {
	CurrentUserID string
	Slug          string
}

type EditorContext struct {
	ProjectRef string
	PageID     string
}

type PublicElementsInput struct {
	Slug string
	Now  time.Time
}

type PublicIncidentsInput struct {
	Slug  string
	Limit int32
	Now   time.Time
}

type PublicElementChartInput struct {
	Slug      string
	ElementID string
	Range     *domainpublic.ChartRange
	Now       time.Time
}

type PublicElementDailyStatusInput struct {
	Slug      string
	ElementID string
	Range     *domainpublic.ChartRange
	Now       time.Time
}

type PublicSummary struct {
	Page        domainpublic.Page
	Status      domainpublic.Status
	GeneratedAt time.Time
}

type PublicElements struct {
	Page        domainpublic.Page
	Elements    []domainpublic.RenderedElement
	GeneratedAt time.Time
}

type PublicIncidents struct {
	ActiveIncidents   []domainpublic.Incident
	ResolvedIncidents []domainpublic.Incident
	GeneratedAt       time.Time
}

type PublicElementChart struct {
	Chart       *domainpublic.Chart
	GeneratedAt time.Time
}

type PublicElementDailyStatus struct {
	Range       domainpublic.ChartRange
	Days        []domainpublic.DailyStatusDay
	GeneratedAt time.Time
}

type PageDetail struct {
	Page     domainpublic.Page
	Elements []domainpublic.Element
}
