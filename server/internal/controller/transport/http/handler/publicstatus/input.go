package publicstatus

import domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"

type pageInputBody struct {
	Slug              string                  `json:"slug"`
	Title             string                  `json:"title"`
	Description       *string                 `json:"description"`
	Enabled           *bool                   `json:"enabled"`
	DefaultChartMode  domainpublic.ChartMode  `json:"defaultChartMode"`
	DefaultChartRange domainpublic.ChartRange `json:"defaultChartRange"`
}

type elementInputBody struct {
	ParentElementID *string                  `json:"parentElementId"`
	Kind            domainpublic.ElementKind `json:"kind"`
	CheckID         *string                  `json:"checkId"`
	Title           *string                  `json:"title"`
	Description     *string                  `json:"description"`
	SortOrder       int32                    `json:"sortOrder"`
	ChartMode       domainpublic.ChartMode   `json:"chartMode"`
	ChartRange      *domainpublic.ChartRange `json:"chartRange"`
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
