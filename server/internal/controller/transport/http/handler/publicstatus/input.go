package publicstatus

import domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"

type pageInputBody struct {
	Slug                string                  `json:"slug"`
	Title               string                  `json:"title"`
	Description         *string                 `json:"description"`
	Enabled             *bool                   `json:"enabled"`
	FooterText          *string                 `json:"footerText"`
	BannerImageURL      *string                 `json:"bannerImageUrl"`
	Theme               domainpublic.Theme      `json:"theme"`
	ShowTargets         *bool                   `json:"showTargets"`
	ShowProbeNames      *bool                   `json:"showProbeNames"`
	ShowProbeLocations  *bool                   `json:"showProbeLocations"`
	ShowIncidentHistory *bool                   `json:"showIncidentHistory"`
	ShowGeneratedAt     *bool                   `json:"showGeneratedAt"`
	CustomCSS           *string                 `json:"customCss"`
	DefaultChartMode    domainpublic.ChartMode  `json:"defaultChartMode"`
	DefaultChartRange   domainpublic.ChartRange `json:"defaultChartRange"`
}

type elementInputBody struct {
	ParentElementID         *string                              `json:"parentElementId"`
	Kind                    domainpublic.ElementKind             `json:"kind"`
	CheckID                 *string                              `json:"checkId"`
	AssignmentSelectionMode domainpublic.AssignmentSelectionMode `json:"assignmentSelectionMode"`
	AssignmentIDs           []string                             `json:"assignmentIds"`
	Title                   *string                              `json:"title"`
	Description             *string                              `json:"description"`
	SortOrder               int32                                `json:"sortOrder"`
	DisplayMode             domainpublic.ElementDisplayMode      `json:"displayMode"`
	ChartMode               domainpublic.ChartMode               `json:"chartMode"`
	ChartRange              *domainpublic.ChartRange             `json:"chartRange"`
}

func defaultBool(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func defaultPageChartMode(value domainpublic.ChartMode) domainpublic.ChartMode {
	if value == "" {
		return domainpublic.ChartModeOff
	}
	return value
}

func defaultPageTheme(value domainpublic.Theme) domainpublic.Theme {
	if value == "" {
		return domainpublic.ThemeAuto
	}
	return value
}

func defaultElementDisplayMode(value domainpublic.ElementDisplayMode) domainpublic.ElementDisplayMode {
	if value == "" {
		return domainpublic.ElementDisplayModeStatus
	}
	return value
}

func defaultElementChartMode(value domainpublic.ChartMode) domainpublic.ChartMode {
	if value == "" {
		return domainpublic.ChartModeInherit
	}
	return value
}

func defaultChartRange(value domainpublic.ChartRange) domainpublic.ChartRange {
	if value == "" {
		return domainpublic.ChartRange24h
	}
	return value
}
