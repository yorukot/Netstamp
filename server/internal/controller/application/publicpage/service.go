package publicpage

import (
	"context"
	"errors"
	"time"

	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainpublicpage "github.com/yorukot/netstamp/internal/domain/publicpage"
)

type Service struct {
	repo          Repository
	projectAccess ProjectAccess
	pings         PingInsightRepository
	events        EventRecorder
}

func NewService(repo Repository, projectAccess ProjectAccess, pings PingInsightRepository, events EventRecorder) *Service {
	return &Service{
		repo:          repo,
		projectAccess: projectAccess,
		pings:         pings,
		events:        events,
	}
}

func (s *Service) ListPages(ctx context.Context, input ListPagesInput) ([]domainpublicpage.Page, error) {
	ctx, flow := s.startPublicPageFlow(ctx, "public_page.list", PublicPageActionList, input.CurrentUserID)
	defer flow.end()

	projectRef, err := normalizeProjectInput(input.CurrentUserID, input.ProjectRef)
	if err != nil {
		return nil, flow.businessFailure(PublicPageEventListFailure, PublicPageReasonInvalidInput, err)
	}
	flow.setProjectRef(projectRef)
	project, err := s.projectAccess.GetProjectForUser(ctx, projectRef, input.CurrentUserID)
	if err != nil {
		return nil, flow.projectLookupFailure(PublicPageEventListFailure, err)
	}
	flow.setProject(project)

	pages, err := s.repo.ListPages(ctx, project.ID)
	if err != nil {
		return nil, flow.technicalFailure(PublicPageEventListFailure, PublicPageReasonPageListFailed, err)
	}

	return pages, nil
}

func (s *Service) GetPage(ctx context.Context, input GetPageInput) (domainpublicpage.Page, error) {
	ctx, flow := s.startPublicPageFlow(ctx, "public_page.get", PublicPageActionGet, input.CurrentUserID)
	defer flow.end()

	projectRef, err := normalizeProjectInput(input.CurrentUserID, input.ProjectRef)
	if err != nil {
		return domainpublicpage.Page{}, flow.businessFailure(PublicPageEventGetFailure, PublicPageReasonInvalidInput, err)
	}
	flow.setProjectRef(projectRef)
	pageID, err := normalizePageID(input.PageID)
	if err != nil {
		return domainpublicpage.Page{}, flow.businessFailure(PublicPageEventGetFailure, PublicPageReasonInvalidInput, err)
	}
	flow.setPageID(pageID)
	project, err := s.projectAccess.GetProjectForUser(ctx, projectRef, input.CurrentUserID)
	if err != nil {
		return domainpublicpage.Page{}, flow.projectLookupFailure(PublicPageEventGetFailure, err)
	}
	flow.setProject(project)
	page, err := s.repo.GetPageForProject(ctx, project.ID, pageID)
	if err != nil {
		return domainpublicpage.Page{}, flow.pageLookupFailure(PublicPageEventGetFailure, err)
	}
	flow.setPage(page)

	return s.hydratePage(ctx, flow, page, PublicPageEventGetFailure)
}

func (s *Service) CreatePage(ctx context.Context, input CreatePageInput) (domainpublicpage.Page, error) {
	ctx, flow := s.startPublicPageFlow(ctx, "public_page.create", PublicPageActionCreate, input.CurrentUserID)
	defer flow.end()

	input, err := normalizeCreatePageInput(input)
	if err != nil {
		return domainpublicpage.Page{}, flow.businessFailure(PublicPageEventCreateFailure, PublicPageReasonInvalidInput, err)
	}
	flow.setProjectRef(input.ProjectRef)
	flow.setPageSlug(input.Slug)
	project, err := s.loadProjectForWrite(ctx, flow, input.ProjectRef, input.CurrentUserID, PublicPageEventCreateFailure)
	if err != nil {
		return domainpublicpage.Page{}, err
	}

	page, err := s.repo.CreatePage(ctx, domainpublicpage.Page{
		ProjectID:   project.ID,
		Slug:        input.Slug,
		Title:       input.Title,
		Description: input.Description,
		Enabled:     input.Enabled,
	})
	if err != nil {
		return domainpublicpage.Page{}, flow.writeFailure(PublicPageEventCreateFailure, PublicPageReasonPageCreateFailed, err)
	}
	flow.setPage(page)
	flow.success(PublicPageEventCreateSuccess)

	return page, nil
}

func (s *Service) UpdatePage(ctx context.Context, input UpdatePageInput) (domainpublicpage.Page, error) {
	ctx, flow := s.startPublicPageFlow(ctx, "public_page.update", PublicPageActionUpdate, input.CurrentUserID)
	defer flow.end()

	record, err := normalizeUpdatePageInput(input)
	if err != nil {
		return domainpublicpage.Page{}, flow.businessFailure(PublicPageEventUpdateFailure, PublicPageReasonInvalidInput, err)
	}
	flow.setProjectRef(record.ProjectID)
	flow.setPageID(record.ID)
	if record.Slug != nil {
		flow.setPageSlug(*record.Slug)
	}
	project, err := s.loadProjectForWrite(ctx, flow, record.ProjectID, input.CurrentUserID, PublicPageEventUpdateFailure)
	if err != nil {
		return domainpublicpage.Page{}, err
	}
	record.ProjectID = project.ID

	page, err := s.repo.UpdatePage(ctx, record)
	if err != nil {
		return domainpublicpage.Page{}, flow.writeFailure(PublicPageEventUpdateFailure, PublicPageReasonPageUpdateFailed, err)
	}
	flow.setPage(page)
	flow.success(PublicPageEventUpdateSuccess)

	return page, nil
}

func (s *Service) DeletePage(ctx context.Context, input DeletePageInput) error {
	ctx, flow := s.startPublicPageFlow(ctx, "public_page.delete", PublicPageActionDelete, input.CurrentUserID)
	defer flow.end()

	projectRef, err := normalizeProjectInput(input.CurrentUserID, input.ProjectRef)
	if err != nil {
		return flow.businessFailure(PublicPageEventDeleteFailure, PublicPageReasonInvalidInput, err)
	}
	flow.setProjectRef(projectRef)
	pageID, err := normalizePageID(input.PageID)
	if err != nil {
		return flow.businessFailure(PublicPageEventDeleteFailure, PublicPageReasonInvalidInput, err)
	}
	flow.setPageID(pageID)
	project, err := s.loadProjectForWrite(ctx, flow, projectRef, input.CurrentUserID, PublicPageEventDeleteFailure)
	if err != nil {
		return err
	}

	if err := s.repo.SoftDeletePage(ctx, project.ID, pageID); err != nil {
		return flow.writeFailure(PublicPageEventDeleteFailure, PublicPageReasonPageDeleteFailed, err)
	}
	flow.success(PublicPageEventDeleteSuccess)

	return nil
}

func (s *Service) CreateFolder(ctx context.Context, input CreateFolderInput) (domainpublicpage.Folder, error) {
	ctx, flow := s.startPublicPageFlow(ctx, "public_page.folder.create", PublicPageActionCreateFolder, input.CurrentUserID)
	defer flow.end()

	input, err := normalizeCreateFolderInput(input)
	if err != nil {
		return domainpublicpage.Folder{}, flow.businessFailure(PublicPageEventCreateFolderFailure, PublicPageReasonInvalidInput, err)
	}
	flow.setProjectRef(input.ProjectRef)
	flow.setPageID(input.PageID)
	project, err := s.loadProjectForWrite(ctx, flow, input.ProjectRef, input.CurrentUserID, PublicPageEventCreateFolderFailure)
	if err != nil {
		return domainpublicpage.Folder{}, err
	}
	err = s.ensurePageExists(ctx, project.ID, input.PageID)
	if err != nil {
		return domainpublicpage.Folder{}, flow.pageLookupFailure(PublicPageEventCreateFolderFailure, err)
	}
	err = s.ensureParentValid(ctx, project.ID, input.PageID, "", input.ParentID)
	if err != nil {
		return domainpublicpage.Folder{}, flow.folderLookupFailure(PublicPageEventCreateFolderFailure, err)
	}

	folder, err := s.repo.CreateFolder(ctx, project.ID, domainpublicpage.Folder{
		PageID:      input.PageID,
		ParentID:    input.ParentID,
		Name:        input.Name,
		Description: input.Description,
		SortOrder:   input.SortOrder,
	})
	if err != nil {
		return domainpublicpage.Folder{}, flow.writeFailure(PublicPageEventCreateFolderFailure, PublicPageReasonFolderCreateFailed, err)
	}
	flow.setFolderID(folder.ID)
	flow.success(PublicPageEventCreateFolderSuccess)

	return folder, nil
}

func (s *Service) UpdateFolder(ctx context.Context, input UpdateFolderInput) (domainpublicpage.Folder, error) {
	ctx, flow := s.startPublicPageFlow(ctx, "public_page.folder.update", PublicPageActionUpdateFolder, input.CurrentUserID)
	defer flow.end()

	record, err := normalizeUpdateFolderInput(input)
	if err != nil {
		return domainpublicpage.Folder{}, flow.businessFailure(PublicPageEventUpdateFolderFailure, PublicPageReasonInvalidInput, err)
	}
	flow.setProjectRef(record.ProjectID)
	flow.setPageID(record.PageID)
	flow.setFolderID(record.ID)
	project, err := s.loadProjectForWrite(ctx, flow, record.ProjectID, input.CurrentUserID, PublicPageEventUpdateFolderFailure)
	if err != nil {
		return domainpublicpage.Folder{}, err
	}
	record.ProjectID = project.ID
	err = s.ensurePageExists(ctx, project.ID, record.PageID)
	if err != nil {
		return domainpublicpage.Folder{}, flow.pageLookupFailure(PublicPageEventUpdateFolderFailure, err)
	}
	if record.ParentIDSet {
		err = s.ensureParentValid(ctx, project.ID, record.PageID, record.ID, record.ParentID)
		if err != nil {
			return domainpublicpage.Folder{}, flow.folderLookupFailure(PublicPageEventUpdateFolderFailure, err)
		}
	}

	folder, err := s.repo.UpdateFolder(ctx, project.ID, record)
	if err != nil {
		return domainpublicpage.Folder{}, flow.writeFailure(PublicPageEventUpdateFolderFailure, PublicPageReasonFolderUpdateFailed, err)
	}
	flow.setFolderID(folder.ID)
	flow.success(PublicPageEventUpdateFolderSuccess)

	return folder, nil
}

func (s *Service) DeleteFolder(ctx context.Context, input DeleteFolderInput) error {
	ctx, flow := s.startPublicPageFlow(ctx, "public_page.folder.delete", PublicPageActionDeleteFolder, input.CurrentUserID)
	defer flow.end()

	projectRef, err := normalizeProjectInput(input.CurrentUserID, input.ProjectRef)
	if err != nil {
		return flow.businessFailure(PublicPageEventDeleteFolderFailure, PublicPageReasonInvalidInput, err)
	}
	flow.setProjectRef(projectRef)
	pageID, err := normalizePageID(input.PageID)
	if err != nil {
		return flow.businessFailure(PublicPageEventDeleteFolderFailure, PublicPageReasonInvalidInput, err)
	}
	flow.setPageID(pageID)
	folderID, err := normalizeFolderID(input.FolderID)
	if err != nil {
		return flow.businessFailure(PublicPageEventDeleteFolderFailure, PublicPageReasonInvalidInput, err)
	}
	flow.setFolderID(folderID)
	project, err := s.loadProjectForWrite(ctx, flow, projectRef, input.CurrentUserID, PublicPageEventDeleteFolderFailure)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteFolder(ctx, project.ID, pageID, folderID); err != nil {
		return flow.writeFailure(PublicPageEventDeleteFolderFailure, PublicPageReasonFolderDeleteFailed, err)
	}
	flow.success(PublicPageEventDeleteFolderSuccess)

	return nil
}

func (s *Service) SetFolderChecks(ctx context.Context, input SetFolderChecksInput) ([]domainpublicpage.PublishedCheck, error) {
	ctx, flow := s.startPublicPageFlow(ctx, "public_page.folder_checks.set", PublicPageActionSetFolderChecks, input.CurrentUserID)
	defer flow.end()

	input, err := normalizeSetFolderChecksInput(input)
	if err != nil {
		return nil, flow.businessFailure(PublicPageEventSetFolderChecksFailure, PublicPageReasonInvalidInput, err)
	}
	flow.setProjectRef(input.ProjectRef)
	flow.setPageID(input.PageID)
	flow.setFolderID(input.FolderID)
	flow.setCheckCount(len(input.CheckIDs))
	project, err := s.loadProjectForWrite(ctx, flow, input.ProjectRef, input.CurrentUserID, PublicPageEventSetFolderChecksFailure)
	if err != nil {
		return nil, err
	}
	err = s.ensurePageExists(ctx, project.ID, input.PageID)
	if err != nil {
		return nil, flow.pageLookupFailure(PublicPageEventSetFolderChecksFailure, err)
	}
	err = s.ensureFolderExists(ctx, project.ID, input.PageID, input.FolderID)
	if err != nil {
		return nil, flow.folderLookupFailure(PublicPageEventSetFolderChecksFailure, err)
	}

	checks, err := s.repo.SetFolderChecks(ctx, project.ID, input.PageID, input.FolderID, input.CheckIDs)
	if err != nil {
		return nil, flow.writeFailure(PublicPageEventSetFolderChecksFailure, PublicPageReasonFolderChecksUpdateFailed, err)
	}
	flow.success(PublicPageEventSetFolderChecksSuccess)

	return checks, nil
}

func (s *Service) GetPublicPage(ctx context.Context, input GetPublicPageInput) (domainpublicpage.Page, error) {
	ctx, flow := s.startPublicPageFlow(ctx, "public_page.public_get", PublicPageActionPublicGet, "")
	defer flow.end()

	slug, err := normalizePublicPageSlug(input.Slug)
	if err != nil {
		return domainpublicpage.Page{}, flow.businessFailure(PublicPageEventPublicGetFailure, PublicPageReasonInvalidInput, err)
	}
	flow.setPageSlug(slug)
	page, err := s.repo.GetEnabledPageBySlug(ctx, slug)
	if err != nil {
		return domainpublicpage.Page{}, flow.pageLookupFailure(PublicPageEventPublicGetFailure, err)
	}
	flow.setPage(page)

	return s.hydratePage(ctx, flow, page, PublicPageEventPublicGetFailure)
}

func (s *Service) QueryPublicPingInsight(ctx context.Context, input QueryPublicPingInsightInput) (PublicPingInsightOutput, error) {
	ctx, flow := s.startPublicPageFlow(ctx, "public_page.ping_insight", PublicPageActionPingInsight, "")
	defer flow.end()

	input, from, to, maxDataPoints, err := normalizeQueryPublicPingInsightInput(input)
	if err != nil {
		return PublicPingInsightOutput{}, flow.businessFailure(PublicPageEventPingInsightFailure, PublicPageReasonInvalidInput, err)
	}
	flow.setPageSlug(input.Slug)
	flow.setProbeID(input.ProbeID)
	flow.setCheckID(input.CheckID)
	projectID, err := s.repo.ResolvePublicPingPairProjectID(ctx, input.Slug, input.ProbeID, input.CheckID)
	if err != nil {
		return PublicPingInsightOutput{}, flow.publicPairLookupFailure(err)
	}
	flow.projectID = projectID
	flow.span.SetAttributes(attrProjectID.String(projectID))
	if s.pings == nil {
		return PublicPingInsightOutput{}, flow.technicalFailure(PublicPageEventPingInsightFailure, PublicPageReasonPingRepositoryNotConfigured, errors.New("ping repository is not configured"))
	}

	result, err := s.pings.ListPingInsight(ctx, domainping.InsightQuery{
		ProjectID:     projectID,
		ProbeID:       input.ProbeID,
		CheckID:       input.CheckID,
		From:          from,
		To:            to,
		MaxDataPoints: maxDataPoints,
	})
	if err != nil {
		return PublicPingInsightOutput{}, flow.technicalFailure(PublicPageEventPingInsightFailure, PublicPageReasonPingInsightQueryFailed, err)
	}

	return newPublicPingInsightOutput(result, from, to, maxDataPoints), nil
}

func (s *Service) hydratePage(ctx context.Context, flow *publicPageFlow, page domainpublicpage.Page, event PublicPageEventName) (domainpublicpage.Page, error) {
	folders, err := s.repo.ListFolders(ctx, page.ProjectID, page.ID)
	if err != nil {
		return domainpublicpage.Page{}, flow.technicalFailure(event, PublicPageReasonFolderListFailed, err)
	}
	checks, err := s.repo.ListFolderChecks(ctx, page.ProjectID, page.ID)
	if err != nil {
		return domainpublicpage.Page{}, flow.technicalFailure(event, PublicPageReasonFolderChecksListFailed, err)
	}
	folders = attachChecks(folders, checks)
	pairs, err := s.repo.ListPingPairs(ctx, page.ProjectID, page.ID)
	if err != nil {
		return domainpublicpage.Page{}, flow.technicalFailure(event, PublicPageReasonPingPairsListFailed, err)
	}
	page.Folders = folders
	page.Pairs = pairs
	return page, nil
}

func (s *Service) loadProjectForWrite(ctx context.Context, flow *publicPageFlow, projectRef, userID string, event PublicPageEventName) (domainproject.Project, error) {
	project, err := s.projectAccess.GetProjectForUser(ctx, projectRef, userID)
	if err != nil {
		return domainproject.Project{}, flow.projectLookupFailure(event, err)
	}
	flow.setProject(project)
	role, err := s.projectAccess.GetMemberRole(ctx, project.ID, userID)
	if err != nil {
		return domainproject.Project{}, flow.roleLookupFailure(event, err)
	}
	if !domainproject.Can(role, domainproject.ActionUpdateProject) {
		return domainproject.Project{}, flow.businessFailure(event, PublicPageReasonForbidden, ErrForbidden)
	}

	return project, nil
}

func (s *Service) ensurePageExists(ctx context.Context, projectID, pageID string) error {
	_, err := s.repo.GetPageForProject(ctx, projectID, pageID)
	return err
}

func (s *Service) ensureFolderExists(ctx context.Context, projectID, pageID, folderID string) error {
	folders, err := s.repo.ListFolders(ctx, projectID, pageID)
	if err != nil {
		return err
	}
	for _, folder := range folders {
		if folder.ID == folderID {
			return nil
		}
	}

	return domainpublicpage.ErrFolderNotFound
}

func (s *Service) ensureParentValid(ctx context.Context, projectID, pageID, folderID string, parentID *string) error {
	if parentID == nil {
		return nil
	}
	if folderID != "" && *parentID == folderID {
		return invalidField("parentId", "cannot be the same folder", *parentID)
	}
	folders, err := s.repo.ListFolders(ctx, projectID, pageID)
	if err != nil {
		return err
	}
	parentByID := make(map[string]*string, len(folders))
	for _, folder := range folders {
		parentByID[folder.ID] = folder.ParentID
	}
	if _, ok := parentByID[*parentID]; !ok {
		return domainpublicpage.ErrFolderNotFound
	}
	for current := parentID; current != nil; current = parentByID[*current] {
		if folderID != "" && *current == folderID {
			return invalidField("parentId", "cannot point to a descendant folder", *parentID)
		}
	}

	return nil
}

func attachChecks(folders []domainpublicpage.Folder, checks []domainpublicpage.PublishedCheck) []domainpublicpage.Folder {
	indexByID := make(map[string]int, len(folders))
	for index := range folders {
		indexByID[folders[index].ID] = index
	}
	for _, check := range checks {
		index, ok := indexByID[check.FolderID]
		if !ok {
			continue
		}
		folders[index].Checks = append(folders[index].Checks, check)
	}
	return folders
}

func newPublicPingInsightOutput(result domainping.InsightResult, from, to time.Time, maxDataPoints int32) PublicPingInsightOutput {
	return PublicPingInsightOutput{
		Summary: newPingInsightSummary(result.Summary),
		Meta: QueryMetadata{
			FromMs:        from.UTC().UnixMilli(),
			ToMs:          to.UTC().UnixMilli(),
			MaxDataPoints: maxDataPoints,
			Source:        string(result.Source),
			Resolution:    string(result.Resolution),
			TotalPoints:   result.TotalPoints,
		},
	}
}

func newPingInsightSummary(summary domainping.InsightSummary) PingInsightSummary {
	return PingInsightSummary{
		AverageRttMs: summary.AverageRttMs,
		MaxRttMs:     summary.MaxRttMs,
		LossPercent:  summary.LossPercent,
		SuccessRate:  summary.SuccessRate,
		Samples:      summary.Samples,
	}
}
