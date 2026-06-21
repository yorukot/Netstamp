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
	return normalizePage(page, input.Slug, input.Title, input.Description, input.DefaultChartMode, input.DefaultChartRange)
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
	return normalizePage(page, input.Slug, input.Title, input.Description, input.DefaultChartMode, input.DefaultChartRange)
}

func normalizePage(page domainpublic.Page, slug, title string, description *string, chartMode domainpublic.ChartMode, chartRange domainpublic.ChartRange) (domainpublic.Page, error) {
	var collector appvalidation.Collector
	var err error

	page.Slug, err = domainpublic.VNSlug(slug)
	collector.AddError("slug", err, slug)
	page.Title, err = domainpublic.VNTitle(title)
	collector.AddError("title", err, title)
	page.Description, err = domainpublic.VNDescription(description)
	collector.AddError("description", err, description)
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
	return normalizeElement(element, input.ParentElementID, input.Kind, input.CheckID, input.Title, input.Description, input.SortOrder, input.ChartMode, input.ChartRange)
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
	return normalizeElement(element, input.ParentElementID, input.Kind, input.CheckID, input.Title, input.Description, input.SortOrder, input.ChartMode, input.ChartRange)
}

func normalizeElement(element domainpublic.Element, parentElementID *string, kind domainpublic.ElementKind, checkID *string, title, description *string, sortOrder int32, chartMode domainpublic.ChartMode, chartRange *domainpublic.ChartRange) (domainpublic.Element, error) {
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
	element.Title, err = domainpublic.VNDescription(title)
	collector.AddError("title", err, title)
	element.Description, err = domainpublic.VNDescription(description)
	collector.AddError("description", err, description)
	element.SortOrder, err = domainpublic.VNSortOrder(sortOrder)
	collector.AddError("sortOrder", err, sortOrder)
	element.ChartMode, err = domainpublic.VNChartMode(chartMode, true)
	collector.AddError("chartMode", err, chartMode)
	if chartRange != nil {
		normalizedRange, rangeErr := domainpublic.VNChartRange(*chartRange)
		collector.AddError("chartRange", rangeErr, *chartRange)
		if rangeErr == nil {
			element.ChartRange = &normalizedRange
		}
	}
	if kind == domainpublic.ElementKindCheck {
		if checkID == nil {
			collector.Add("checkId", "must be provided for check elements", nil)
		} else {
			normalizedCheckID, checkErr := domaincheck.VNCheckID(*checkID)
			collector.AddError("checkId", checkErr, *checkID)
			if checkErr == nil {
				element.CheckID = &normalizedCheckID
			}
		}
	}
	if kind == domainpublic.ElementKindFolder {
		if checkID != nil {
			collector.Add("checkId", "must be omitted for folder elements", *checkID)
		}
		if parentElementID != nil {
			collector.Add("parentElementId", "folder elements must be placed at the root", *parentElementID)
		}
	}

	if err := collector.Err(ErrInvalidInput); err != nil {
		return domainpublic.Element{}, err
	}
	return element, nil
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
