package publicstatus

import (
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
)

func invalidField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}

func normalizeCreatePageInput(projectID string, input CreatePageInput) (domainpublic.Page, error) {
	page := domainpublic.Page{
		ProjectID:       projectID,
		CreatedByUserID: input.CurrentUserID,
		Enabled:         input.Enabled,
	}
	return normalizePage(page, input.Slug, input.Title, input.Description, input.FooterText, input.BannerImageURL, input.Theme, input.ShowTargets, input.ShowProbeNames, input.ShowProbeLocations, input.ShowIncidentHistory, input.ShowGeneratedAt, input.CustomCSS, input.DefaultChartMode, input.DefaultChartRange)
}

func normalizeUpdatePageInput(projectID string, input UpdatePageInput) (domainpublic.Page, error) {
	pageID, err := domainpublic.VNPageID(input.PageID)
	if err != nil {
		return domainpublic.Page{}, invalidField("pageId", err.Error(), input.PageID)
	}
	page := domainpublic.Page{
		ID:              pageID,
		ProjectID:       projectID,
		CreatedByUserID: input.CurrentUserID,
		Enabled:         input.Enabled,
	}
	return normalizePage(page, input.Slug, input.Title, input.Description, input.FooterText, input.BannerImageURL, input.Theme, input.ShowTargets, input.ShowProbeNames, input.ShowProbeLocations, input.ShowIncidentHistory, input.ShowGeneratedAt, input.CustomCSS, input.DefaultChartMode, input.DefaultChartRange)
}

func normalizePage(
	page domainpublic.Page,
	slug, title string,
	description, footerText, bannerImageURL *string,
	theme domainpublic.Theme,
	showTargets, showProbeNames, showProbeLocations, showIncidentHistory, showGeneratedAt bool,
	customCSS *string,
	chartMode domainpublic.ChartMode,
	chartRange domainpublic.ChartRange,
) (domainpublic.Page, error) {
	var collector appvalidation.Collector
	var err error

	page.Slug, err = domainpublic.VNSlug(slug)
	collector.AddError("slug", err, slug)
	page.Title, err = domainpublic.VNTitle(title)
	collector.AddError("title", err, title)
	page.Description, err = domainpublic.VNDescription(description)
	collector.AddError("description", err, description)
	page.FooterText, err = domainpublic.VNFooterText(footerText)
	collector.AddError("footerText", err, footerText)
	page.BannerImageURL, err = domainpublic.VNBannerImageURL(bannerImageURL)
	collector.AddError("bannerImageUrl", err, bannerImageURL)
	if theme == "" {
		theme = domainpublic.ThemeAuto
	}
	page.Theme, err = domainpublic.VNTheme(theme)
	collector.AddError("theme", err, theme)
	page.ShowTargets = showTargets
	page.ShowProbeNames = showProbeNames
	page.ShowProbeLocations = showProbeLocations
	page.ShowIncidentHistory = showIncidentHistory
	page.ShowGeneratedAt = showGeneratedAt
	page.CustomCSS, err = domainpublic.VNCustomCSS(customCSS)
	collector.AddError("customCss", err, customCSS)
	page.DefaultChartMode, err = domainpublic.VNChartMode(chartMode, false)
	collector.AddError("defaultChartMode", err, chartMode)
	page.DefaultChartRange, err = domainpublic.VNChartRange(chartRange)
	collector.AddError("defaultChartRange", err, chartRange)

	if err := collector.Err(ErrInvalidInput); err != nil {
		return domainpublic.Page{}, err
	}
	return page, nil
}

func normalizeCreateElementInput(projectID, pageID string, input CreateElementInput) (domainpublic.Element, error) {
	pageID, err := domainpublic.VNPageID(pageID)
	if err != nil {
		return domainpublic.Element{}, invalidField("pageId", err.Error(), pageID)
	}
	element := domainpublic.Element{
		ProjectID:    projectID,
		PublicPageID: pageID,
	}
	return normalizeElement(element, input.ParentElementID, input.Kind, input.CheckID, input.AssignmentSelectionMode, input.AssignmentIDs, input.Title, input.Description, input.SortOrder, input.DisplayMode, input.ChartMode, input.ChartRange)
}

func normalizeUpdateElementInput(projectID, pageID string, input UpdateElementInput) (domainpublic.Element, error) {
	pageID, err := domainpublic.VNPageID(pageID)
	if err != nil {
		return domainpublic.Element{}, invalidField("pageId", err.Error(), pageID)
	}
	elementID, err := domainpublic.VNElementID(input.ElementID)
	if err != nil {
		return domainpublic.Element{}, invalidField("elementId", err.Error(), input.ElementID)
	}
	element := domainpublic.Element{
		ID:           elementID,
		ProjectID:    projectID,
		PublicPageID: pageID,
	}
	return normalizeElement(element, input.ParentElementID, input.Kind, input.CheckID, input.AssignmentSelectionMode, input.AssignmentIDs, input.Title, input.Description, input.SortOrder, input.DisplayMode, input.ChartMode, input.ChartRange)
}

func normalizeElement(
	element domainpublic.Element,
	parentElementID *string,
	kind domainpublic.ElementKind,
	checkID *string,
	selectionMode domainpublic.AssignmentSelectionMode,
	assignmentIDs []string,
	title *string,
	description *string,
	sortOrder int32,
	displayMode domainpublic.ElementDisplayMode,
	chartMode domainpublic.ChartMode,
	chartRange *domainpublic.ChartRange,
) (domainpublic.Element, error) {
	var collector appvalidation.Collector
	var err error

	if parentElementID != nil {
		normalizedParentID, parentErr := domainpublic.VNElementID(*parentElementID)
		collector.AddError("parentElementId", parentErr, *parentElementID)
		if parentErr == nil {
			element.ParentElementID = &normalizedParentID
		}
	}
	element.Kind, err = domainpublic.VNElementKind(kind)
	collector.AddError("kind", err, kind)
	element.AssignmentIDs = normalizeAssignmentIDs(&collector, assignmentIDs)
	element.Title, err = domainpublic.VNDescription(title)
	collector.AddError("title", err, title)
	element.Description, err = domainpublic.VNDescription(description)
	collector.AddError("description", err, description)
	element.SortOrder, err = domainpublic.VNSortOrder(sortOrder)
	collector.AddError("sortOrder", err, sortOrder)
	if displayMode == "" {
		displayMode = domainpublic.ElementDisplayModeStatus
	}
	element.DisplayMode, err = domainpublic.VNElementDisplayMode(displayMode)
	collector.AddError("displayMode", err, displayMode)
	element.ChartMode, err = domainpublic.VNChartMode(chartMode, true)
	collector.AddError("chartMode", err, chartMode)
	if chartRange != nil {
		normalizedRange, rangeErr := domainpublic.VNChartRange(*chartRange)
		collector.AddError("chartRange", rangeErr, *chartRange)
		if rangeErr == nil {
			element.ChartRange = &normalizedRange
		}
	}
	if kind == domainpublic.ElementKindFolder {
		if checkID != nil {
			collector.Add("checkId", "must be omitted for folder elements", *checkID)
		}
		if selectionMode != "" {
			collector.Add("assignmentSelectionMode", "must be omitted for folder elements", selectionMode)
		}
		if len(assignmentIDs) > 0 {
			collector.Add("assignmentIds", "must be omitted for folder elements", assignmentIDs)
		}
		if parentElementID != nil {
			collector.Add("parentElementId", "folder elements must be placed at the root", *parentElementID)
		}
	}
	if kind == domainpublic.ElementKindAssignmentGroup {
		mode := selectionMode
		if mode == "" {
			mode = domainpublic.AssignmentSelectionModeAllCheck
		}
		normalizedMode, modeErr := domainpublic.VNAssignmentSelectionMode(mode)
		collector.AddError("assignmentSelectionMode", modeErr, selectionMode)
		if modeErr == nil {
			element.AssignmentSelectionMode = &normalizedMode
		}
		normalizeAssignmentGroupScope(&collector, &element, checkID, normalizedMode)
	}

	if err := collector.Err(ErrInvalidInput); err != nil {
		return domainpublic.Element{}, err
	}
	return element, nil
}

func normalizeAssignmentIDs(collector *appvalidation.Collector, assignmentIDs []string) []string {
	seen := make(map[string]struct{}, len(assignmentIDs))
	normalized := make([]string, 0, len(assignmentIDs))
	for _, assignmentID := range assignmentIDs {
		value, err := domainpublic.VNAssignmentID(assignmentID)
		collector.AddError("assignmentIds", err, assignmentID)
		if err != nil {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}

func normalizeAssignmentGroupScope(collector *appvalidation.Collector, element *domainpublic.Element, checkID *string, mode domainpublic.AssignmentSelectionMode) {
	switch mode {
	case domainpublic.AssignmentSelectionModeAllCheck:
		if checkID == nil {
			collector.Add("checkId", "must be provided for all-check assignment groups", nil)
			return
		}
		normalizedCheckID, checkErr := domaincheck.VNCheckID(*checkID)
		collector.AddError("checkId", checkErr, *checkID)
		if checkErr == nil {
			element.CheckID = &normalizedCheckID
		}
		if len(element.AssignmentIDs) > 0 {
			collector.Add("assignmentIds", "must be omitted for all-check assignment groups", element.AssignmentIDs)
		}
	case domainpublic.AssignmentSelectionModeSelectedAssignments:
		if checkID != nil {
			collector.Add("checkId", "must be omitted for selected-assignment groups", *checkID)
		}
		if len(element.AssignmentIDs) == 0 {
			collector.Add("assignmentIds", "must include at least one assignment", nil)
		}
	}
}

func validateParent(parent domainpublic.Element, elementID string) error {
	if parent.Kind != domainpublic.ElementKindFolder {
		return invalidField("parentElementId", "parent element must be a folder", parent.ID)
	}
	if parent.ParentElementID != nil {
		return invalidField("parentElementId", "nested folders are not supported", parent.ID)
	}
	if elementID != "" && parent.ID == elementID {
		return invalidField("parentElementId", "element cannot be its own parent", parent.ID)
	}
	return nil
}
