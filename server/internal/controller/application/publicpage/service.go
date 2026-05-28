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
}

func NewService(repo Repository, projectAccess ProjectAccess, pings PingInsightRepository) *Service {
	return &Service{
		repo:          repo,
		projectAccess: projectAccess,
		pings:         pings,
	}
}

func (s *Service) ListPages(ctx context.Context, input ListPagesInput) ([]domainpublicpage.Page, error) {
	projectRef, err := normalizeProjectInput(input.CurrentUserID, input.ProjectRef)
	if err != nil {
		return nil, err
	}
	project, err := s.projectAccess.GetProjectForUser(ctx, projectRef, input.CurrentUserID)
	if err != nil {
		return nil, err
	}

	return s.repo.ListPages(ctx, project.ID)
}

func (s *Service) GetPage(ctx context.Context, input GetPageInput) (domainpublicpage.Page, error) {
	projectRef, err := normalizeProjectInput(input.CurrentUserID, input.ProjectRef)
	if err != nil {
		return domainpublicpage.Page{}, err
	}
	pageID, err := normalizePageID(input.PageID)
	if err != nil {
		return domainpublicpage.Page{}, err
	}
	project, err := s.projectAccess.GetProjectForUser(ctx, projectRef, input.CurrentUserID)
	if err != nil {
		return domainpublicpage.Page{}, err
	}
	page, err := s.repo.GetPageForProject(ctx, project.ID, pageID)
	if err != nil {
		return domainpublicpage.Page{}, err
	}

	return s.hydratePage(ctx, page)
}

func (s *Service) CreatePage(ctx context.Context, input CreatePageInput) (domainpublicpage.Page, error) {
	input, err := normalizeCreatePageInput(input)
	if err != nil {
		return domainpublicpage.Page{}, err
	}
	project, err := s.loadProjectForWrite(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainpublicpage.Page{}, err
	}

	return s.repo.CreatePage(ctx, domainpublicpage.Page{
		ProjectID:   project.ID,
		Slug:        input.Slug,
		Title:       input.Title,
		Description: input.Description,
		Enabled:     input.Enabled,
	})
}

func (s *Service) UpdatePage(ctx context.Context, input UpdatePageInput) (domainpublicpage.Page, error) {
	record, err := normalizeUpdatePageInput(input)
	if err != nil {
		return domainpublicpage.Page{}, err
	}
	project, err := s.loadProjectForWrite(ctx, record.ProjectID, input.CurrentUserID)
	if err != nil {
		return domainpublicpage.Page{}, err
	}
	record.ProjectID = project.ID

	return s.repo.UpdatePage(ctx, record)
}

func (s *Service) DeletePage(ctx context.Context, input DeletePageInput) error {
	projectRef, err := normalizeProjectInput(input.CurrentUserID, input.ProjectRef)
	if err != nil {
		return err
	}
	pageID, err := normalizePageID(input.PageID)
	if err != nil {
		return err
	}
	project, err := s.loadProjectForWrite(ctx, projectRef, input.CurrentUserID)
	if err != nil {
		return err
	}

	return s.repo.SoftDeletePage(ctx, project.ID, pageID)
}

func (s *Service) CreateFolder(ctx context.Context, input CreateFolderInput) (domainpublicpage.Folder, error) {
	input, err := normalizeCreateFolderInput(input)
	if err != nil {
		return domainpublicpage.Folder{}, err
	}
	project, err := s.loadProjectForWrite(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainpublicpage.Folder{}, err
	}
	if err := s.ensurePageExists(ctx, project.ID, input.PageID); err != nil {
		return domainpublicpage.Folder{}, err
	}
	if err := s.ensureParentValid(ctx, project.ID, input.PageID, "", input.ParentID); err != nil {
		return domainpublicpage.Folder{}, err
	}

	return s.repo.CreateFolder(ctx, project.ID, domainpublicpage.Folder{
		PageID:      input.PageID,
		ParentID:    input.ParentID,
		Name:        input.Name,
		Description: input.Description,
		SortOrder:   input.SortOrder,
	})
}

func (s *Service) UpdateFolder(ctx context.Context, input UpdateFolderInput) (domainpublicpage.Folder, error) {
	record, err := normalizeUpdateFolderInput(input)
	if err != nil {
		return domainpublicpage.Folder{}, err
	}
	project, err := s.loadProjectForWrite(ctx, record.ProjectID, input.CurrentUserID)
	if err != nil {
		return domainpublicpage.Folder{}, err
	}
	record.ProjectID = project.ID
	if err := s.ensurePageExists(ctx, project.ID, record.PageID); err != nil {
		return domainpublicpage.Folder{}, err
	}
	if record.ParentIDSet {
		if err := s.ensureParentValid(ctx, project.ID, record.PageID, record.ID, record.ParentID); err != nil {
			return domainpublicpage.Folder{}, err
		}
	}

	return s.repo.UpdateFolder(ctx, project.ID, record)
}

func (s *Service) DeleteFolder(ctx context.Context, input DeleteFolderInput) error {
	projectRef, err := normalizeProjectInput(input.CurrentUserID, input.ProjectRef)
	if err != nil {
		return err
	}
	pageID, err := normalizePageID(input.PageID)
	if err != nil {
		return err
	}
	folderID, err := normalizeFolderID(input.FolderID)
	if err != nil {
		return err
	}
	project, err := s.loadProjectForWrite(ctx, projectRef, input.CurrentUserID)
	if err != nil {
		return err
	}

	return s.repo.DeleteFolder(ctx, project.ID, pageID, folderID)
}

func (s *Service) SetFolderChecks(ctx context.Context, input SetFolderChecksInput) ([]domainpublicpage.PublishedCheck, error) {
	input, err := normalizeSetFolderChecksInput(input)
	if err != nil {
		return nil, err
	}
	project, err := s.loadProjectForWrite(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePageExists(ctx, project.ID, input.PageID); err != nil {
		return nil, err
	}
	if err := s.ensureFolderExists(ctx, project.ID, input.PageID, input.FolderID); err != nil {
		return nil, err
	}

	return s.repo.SetFolderChecks(ctx, project.ID, input.PageID, input.FolderID, input.CheckIDs)
}

func (s *Service) GetPublicPage(ctx context.Context, input GetPublicPageInput) (domainpublicpage.Page, error) {
	slug, err := normalizePublicPageSlug(input.Slug)
	if err != nil {
		return domainpublicpage.Page{}, err
	}
	page, err := s.repo.GetEnabledPageBySlug(ctx, slug)
	if err != nil {
		return domainpublicpage.Page{}, err
	}

	return s.hydratePage(ctx, page)
}

func (s *Service) QueryPublicPingInsight(ctx context.Context, input QueryPublicPingInsightInput) (PublicPingInsightOutput, error) {
	input, from, to, maxDataPoints, err := normalizeQueryPublicPingInsightInput(input)
	if err != nil {
		return PublicPingInsightOutput{}, err
	}
	projectID, err := s.repo.ResolvePublicPingPairProjectID(ctx, input.Slug, input.ProbeID, input.CheckID)
	if err != nil {
		return PublicPingInsightOutput{}, err
	}
	if s.pings == nil {
		return PublicPingInsightOutput{}, errors.New("ping repository is not configured")
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
		return PublicPingInsightOutput{}, err
	}

	return newPublicPingInsightOutput(result, from, to, maxDataPoints), nil
}

func (s *Service) hydratePage(ctx context.Context, page domainpublicpage.Page) (domainpublicpage.Page, error) {
	folders, err := s.repo.ListFolders(ctx, page.ProjectID, page.ID)
	if err != nil {
		return domainpublicpage.Page{}, err
	}
	checks, err := s.repo.ListFolderChecks(ctx, page.ProjectID, page.ID)
	if err != nil {
		return domainpublicpage.Page{}, err
	}
	folders = attachChecks(folders, checks)
	pairs, err := s.repo.ListPingPairs(ctx, page.ProjectID, page.ID)
	if err != nil {
		return domainpublicpage.Page{}, err
	}
	page.Folders = folders
	page.Pairs = pairs
	return page, nil
}

func (s *Service) loadProjectForWrite(ctx context.Context, projectRef, userID string) (domainproject.Project, error) {
	project, err := s.projectAccess.GetProjectForUser(ctx, projectRef, userID)
	if err != nil {
		return domainproject.Project{}, err
	}
	role, err := s.projectAccess.GetMemberRole(ctx, project.ID, userID)
	if err != nil {
		return domainproject.Project{}, err
	}
	if !domainproject.Can(role, domainproject.ActionUpdateProject) {
		return domainproject.Project{}, ErrForbidden
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
		Buckets:       newPingInsightBuckets(result.Buckets),
		SampleDensity: newPingSampleDensity(result.SampleDensity),
		Summary:       newPingInsightSummary(result.Summary),
		Query: QueryMetadata{
			FromMs:        from.UTC().UnixMilli(),
			ToMs:          to.UTC().UnixMilli(),
			MaxDataPoints: maxDataPoints,
			Resolution:    string(result.Resolution),
			TotalPoints:   result.TotalPoints,
		},
	}
}

func newPingInsightBuckets(buckets []domainping.InsightBucket) []PingInsightBucket {
	values := make([]PingInsightBucket, 0, len(buckets))
	for _, bucket := range buckets {
		values = append(values, PingInsightBucket{
			TimestampMs:   bucket.Timestamp.UTC().UnixMilli(),
			ResultCount:   bucket.ResultCount,
			DurationAvgMs: bucket.DurationAvgMs,
			RttMinMs:      bucket.RttMinMs,
			RttAvgMs:      bucket.RttAvgMs,
			RttMedianMs:   bucket.RttMedianMs,
			RttMaxMs:      bucket.RttMaxMs,
			RttStddevMs:   bucket.RttStddevMs,
			LossPercent:   bucket.LossPercent,
			SuccessRate:   bucket.SuccessRate,
			SentCount:     bucket.SentCount,
			ReceivedCount: bucket.ReceivedCount,
			TimeoutCount:  bucket.TimeoutCount,
			ErrorCount:    bucket.ErrorCount,
		})
	}
	return values
}

func newPingSampleDensity(cells []domainping.SampleDensityCell) []PingSampleDensityCell {
	values := make([]PingSampleDensityCell, 0, len(cells))
	for _, cell := range cells {
		values = append(values, PingSampleDensityCell{
			TimestampMs:      cell.Timestamp.UTC().UnixMilli(),
			RttBucketStartMs: cell.RttBucketStartMs,
			RttBucketEndMs:   cell.RttBucketEndMs,
			SampleCount:      cell.SampleCount,
		})
	}
	return values
}

func newPingInsightSummary(summary domainping.InsightSummary) PingInsightSummary {
	return PingInsightSummary{
		TotalResults:      summary.TotalResults,
		SuccessfulCount:   summary.SuccessfulCount,
		TimeoutCount:      summary.TimeoutCount,
		ErrorCount:        summary.ErrorCount,
		SentCount:         summary.SentCount,
		ReceivedCount:     summary.ReceivedCount,
		AvgLossPercent:    summary.AvgLossPercent,
		AvgRttMs:          summary.AvgRttMs,
		MedianRttMs:       summary.MedianRttMs,
		MaxRttMs:          summary.MaxRttMs,
		P95RttMs:          summary.P95RttMs,
		P99RttMs:          summary.P99RttMs,
		LatestStatus:      pingStatusString(summary.LatestStatus),
		LatestStartedAtMs: timePtrMillis(summary.LatestStartedAt),
		LatestRttAvgMs:    summary.LatestRttAvgMs,
		LatestLossPercent: summary.LatestLossPercent,
	}
}

func pingStatusString(value *domainping.Status) *string {
	if value == nil {
		return nil
	}
	copied := string(*value)
	return &copied
}

func timePtrMillis(value *time.Time) *int64 {
	if value == nil {
		return nil
	}
	millis := value.UTC().UnixMilli()
	return &millis
}
