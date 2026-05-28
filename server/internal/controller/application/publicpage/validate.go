package publicpage

import (
	"time"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainpublicpage "github.com/yorukot/netstamp/internal/domain/publicpage"
)

const (
	defaultRange         = 24 * time.Hour
	defaultMaxDataPoints = int32(600)
	maxDataPointsLimit   = int32(2000)
)

func normalizeProjectInput(userID, projectRef string) (string, error) {
	ref, err := domainproject.VNProjectRef(projectRef)
	if err != nil {
		return "", invalidField("projectRef", err.Error(), projectRef)
	}
	return ref, nil
}

func normalizePageID(pageID string) (string, error) {
	id, err := domainpublicpage.VNPageID(pageID)
	if err != nil {
		return "", invalidField("pageId", err.Error(), pageID)
	}
	return id, nil
}

func normalizeFolderID(folderID string) (string, error) {
	id, err := domainpublicpage.VNFolderID(folderID)
	if err != nil {
		return "", invalidField("folderId", err.Error(), folderID)
	}
	return id, nil
}

func normalizeCreatePageInput(input CreatePageInput) (CreatePageInput, error) {
	var validation appvalidation.Collector
	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	slug, err := domainpublicpage.VNSlug(input.Slug)
	if err != nil {
		validation.AddError("slug", err, input.Slug)
	}
	title, err := domainpublicpage.VNTitle(input.Title)
	if err != nil {
		validation.AddError("title", err, input.Title)
	}
	description, err := domainpublicpage.VNDescription(input.Description)
	if err != nil {
		validation.AddError("description", err, input.Description)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return CreatePageInput{}, err
	}

	input.ProjectRef = projectRef
	input.Slug = slug
	input.Title = title
	input.Description = description
	return input, nil
}

func normalizeUpdatePageInput(input UpdatePageInput) (domainpublicpage.PageUpdate, error) {
	var validation appvalidation.Collector
	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	pageID, err := domainpublicpage.VNPageID(input.PageID)
	if err != nil {
		validation.AddError("pageId", err, input.PageID)
	}
	if input.Slug == nil && input.Title == nil && !input.DescriptionSet && input.Enabled == nil {
		validation.Add("", "at least one field must be provided", nil)
	}
	slug, err := optionalSlug(input.Slug)
	if err != nil {
		validation.AddError("slug", err, input.Slug)
	}
	title, err := optionalTitle(input.Title)
	if err != nil {
		validation.AddError("title", err, input.Title)
	}
	description, err := optionalDescription(input.Description, input.DescriptionSet)
	if err != nil {
		validation.AddError("description", err, input.Description)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return domainpublicpage.PageUpdate{}, err
	}

	return domainpublicpage.PageUpdate{
		ProjectID:      projectRef,
		ID:             pageID,
		Slug:           slug,
		Title:          title,
		Description:    description,
		DescriptionSet: input.DescriptionSet,
		Enabled:        input.Enabled,
	}, nil
}

func normalizeCreateFolderInput(input CreateFolderInput) (CreateFolderInput, error) {
	var validation appvalidation.Collector
	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	pageID, err := domainpublicpage.VNPageID(input.PageID)
	if err != nil {
		validation.AddError("pageId", err, input.PageID)
	}
	parentID, err := optionalFolderID(input.ParentID)
	if err != nil {
		validation.AddError("parentId", err, input.ParentID)
	}
	name, err := domainpublicpage.VNFolderName(input.Name)
	if err != nil {
		validation.AddError("name", err, input.Name)
	}
	description, err := domainpublicpage.VNDescription(input.Description)
	if err != nil {
		validation.AddError("description", err, input.Description)
	}
	sortOrder, err := domainpublicpage.VNSortOrder(input.SortOrder)
	if err != nil {
		validation.AddError("sortOrder", err, input.SortOrder)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return CreateFolderInput{}, err
	}

	input.ProjectRef = projectRef
	input.PageID = pageID
	input.ParentID = parentID
	input.Name = name
	input.Description = description
	input.SortOrder = sortOrder
	return input, nil
}

func normalizeUpdateFolderInput(input UpdateFolderInput) (domainpublicpage.FolderUpdate, error) {
	var validation appvalidation.Collector
	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	pageID, err := domainpublicpage.VNPageID(input.PageID)
	if err != nil {
		validation.AddError("pageId", err, input.PageID)
	}
	folderID, err := domainpublicpage.VNFolderID(input.FolderID)
	if err != nil {
		validation.AddError("folderId", err, input.FolderID)
	}
	if !input.ParentIDSet && input.Name == nil && !input.DescriptionSet && input.SortOrder == nil {
		validation.Add("", "at least one field must be provided", nil)
	}
	parentID, err := optionalDescriptionID(input.ParentID, input.ParentIDSet)
	if err != nil {
		validation.AddError("parentId", err, input.ParentID)
	}
	name, err := optionalFolderName(input.Name)
	if err != nil {
		validation.AddError("name", err, input.Name)
	}
	description, err := optionalDescription(input.Description, input.DescriptionSet)
	if err != nil {
		validation.AddError("description", err, input.Description)
	}
	sortOrder, err := optionalSortOrder(input.SortOrder)
	if err != nil {
		validation.AddError("sortOrder", err, input.SortOrder)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return domainpublicpage.FolderUpdate{}, err
	}

	return domainpublicpage.FolderUpdate{
		ProjectID:      projectRef,
		PageID:         pageID,
		ID:             folderID,
		ParentID:       parentID,
		ParentIDSet:    input.ParentIDSet,
		Name:           name,
		Description:    description,
		DescriptionSet: input.DescriptionSet,
		SortOrder:      sortOrder,
	}, nil
}

func normalizeSetFolderChecksInput(input SetFolderChecksInput) (SetFolderChecksInput, error) {
	var validation appvalidation.Collector
	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	pageID, err := domainpublicpage.VNPageID(input.PageID)
	if err != nil {
		validation.AddError("pageId", err, input.PageID)
	}
	folderID, err := domainpublicpage.VNFolderID(input.FolderID)
	if err != nil {
		validation.AddError("folderId", err, input.FolderID)
	}
	checkIDs := make([]string, 0, len(input.CheckIDs))
	seen := make(map[string]struct{}, len(input.CheckIDs))
	for index, checkIDValue := range input.CheckIDs {
		checkID, err := domaincheck.VNCheckID(checkIDValue)
		if err != nil {
			validation.AddError("checkIds", err, checkIDValue)
			continue
		}
		if _, ok := seen[checkID]; ok {
			validation.Add("checkIds", "duplicate check id", checkIDValue)
			continue
		}
		seen[checkID] = struct{}{}
		_ = index
		checkIDs = append(checkIDs, checkID)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return SetFolderChecksInput{}, err
	}

	input.ProjectRef = projectRef
	input.PageID = pageID
	input.FolderID = folderID
	input.CheckIDs = checkIDs
	return input, nil
}

func normalizePublicPageSlug(slug string) (string, error) {
	normalized, err := domainpublicpage.VNSlug(slug)
	if err != nil {
		return "", invalidField("slug", err.Error(), slug)
	}
	return normalized, nil
}

func normalizeQueryPublicPingInsightInput(input QueryPublicPingInsightInput) (QueryPublicPingInsightInput, time.Time, time.Time, int32, error) {
	var validation appvalidation.Collector
	slug, err := domainpublicpage.VNSlug(input.Slug)
	if err != nil {
		validation.AddError("slug", err, input.Slug)
	}
	probeID, err := domainprobe.VNProbeID(input.ProbeID)
	if err != nil {
		validation.AddError("probeId", err, input.ProbeID)
	}
	checkID, err := domaincheck.VNCheckID(input.CheckID)
	if err != nil {
		validation.AddError("checkId", err, input.CheckID)
	}
	now := input.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	from, to, err := normalizeRange(input.FromMs, input.ToMs, now)
	if err != nil {
		if !validation.AddValidation(err) {
			return QueryPublicPingInsightInput{}, time.Time{}, time.Time{}, 0, err
		}
	}
	maxDataPoints, err := normalizeMaxDataPoints(input.MaxDataPoints)
	if err != nil {
		if !validation.AddValidation(err) {
			return QueryPublicPingInsightInput{}, time.Time{}, time.Time{}, 0, err
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return QueryPublicPingInsightInput{}, time.Time{}, time.Time{}, 0, err
	}

	input.Slug = slug
	input.ProbeID = probeID
	input.CheckID = checkID
	return input, from, to, maxDataPoints, nil
}

func optionalSlug(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	normalized, err := domainpublicpage.VNSlug(*value)
	if err != nil {
		return nil, err
	}
	return &normalized, nil
}

func optionalTitle(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	normalized, err := domainpublicpage.VNTitle(*value)
	if err != nil {
		return nil, err
	}
	return &normalized, nil
}

func optionalDescription(value *string, set bool) (*string, error) {
	if !set {
		return nil, nil
	}
	return domainpublicpage.VNDescription(value)
}

func optionalFolderID(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	normalized, err := domainpublicpage.VNFolderID(*value)
	if err != nil {
		return nil, err
	}
	return &normalized, nil
}

func optionalDescriptionID(value *string, set bool) (*string, error) {
	if !set {
		return nil, nil
	}
	if value == nil {
		return nil, nil
	}
	return optionalFolderID(value)
}

func optionalFolderName(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	normalized, err := domainpublicpage.VNFolderName(*value)
	if err != nil {
		return nil, err
	}
	return &normalized, nil
}

func optionalSortOrder(value *int32) (*int32, error) {
	if value == nil {
		return nil, nil
	}
	normalized, err := domainpublicpage.VNSortOrder(*value)
	if err != nil {
		return nil, err
	}
	return &normalized, nil
}

func normalizeRange(fromMs, toMs *int64, now time.Time) (time.Time, time.Time, error) {
	to := now
	if toMs != nil {
		if *toMs <= 0 {
			return time.Time{}, time.Time{}, invalidField("to", "must be greater than 0", *toMs)
		}
		to = time.UnixMilli(*toMs).UTC()
	}
	from := to.Add(-defaultRange)
	if fromMs != nil {
		if *fromMs <= 0 {
			return time.Time{}, time.Time{}, invalidField("from", "must be greater than 0", *fromMs)
		}
		from = time.UnixMilli(*fromMs).UTC()
	}
	if !from.Before(to) {
		return time.Time{}, time.Time{}, invalidField("from", "must be before to", fromMs)
	}

	return from, to, nil
}

func normalizeMaxDataPoints(value *int32) (int32, error) {
	if value == nil || *value == 0 {
		return defaultMaxDataPoints, nil
	}
	if *value < 1 {
		return 0, invalidField("maxDataPoints", "must be greater than 0", *value)
	}
	if *value > maxDataPointsLimit {
		return 0, invalidField("maxDataPoints", "must be less than or equal to 2000", *value)
	}
	return *value, nil
}

func invalidField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}
