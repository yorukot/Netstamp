package publicstatus

import (
	"context"

	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
)

func (s *Service) ListPages(ctx context.Context, input ListPagesInput) ([]domainpublic.Page, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListPages(ctx, project.ID)
}

func (s *Service) GetPage(ctx context.Context, input GetPageInput) (PageDetail, error) {
	project, err := s.loadProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return PageDetail{}, err
	}
	pageID, err := domainpublic.VNPageID(input.PageID)
	if err != nil {
		return PageDetail{}, invalidField("pageId", err.Error(), input.PageID)
	}
	page, err := s.repo.GetPage(ctx, project.ID, pageID)
	if err != nil {
		return PageDetail{}, err
	}
	elements, err := s.repo.ListElements(ctx, page.ID)
	if err != nil {
		return PageDetail{}, err
	}
	return PageDetail{Page: page, Elements: elements}, nil
}

func (s *Service) CreatePage(ctx context.Context, input CreatePageInput) (domainpublic.Page, error) {
	project, err := s.loadWritableProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainpublic.Page{}, err
	}
	page, err := normalizeCreatePageInput(project.ID, input)
	if err != nil {
		return domainpublic.Page{}, err
	}
	created, err := s.repo.CreatePage(ctx, page)
	if err != nil {
		return domainpublic.Page{}, err
	}
	s.clearPublicSnapshots()
	return created, nil
}

func (s *Service) UpdatePage(ctx context.Context, input UpdatePageInput) (domainpublic.Page, error) {
	project, err := s.loadWritableProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainpublic.Page{}, err
	}
	page, err := normalizeUpdatePageInput(project.ID, input)
	if err != nil {
		return domainpublic.Page{}, err
	}
	updated, err := s.repo.UpdatePage(ctx, page)
	if err != nil {
		return domainpublic.Page{}, err
	}
	s.clearPublicSnapshots()
	return updated, nil
}

func (s *Service) DeletePage(ctx context.Context, input DeletePageInput) error {
	project, err := s.loadWritableProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return err
	}
	pageID, err := domainpublic.VNPageID(input.PageID)
	if err != nil {
		return invalidField("pageId", err.Error(), input.PageID)
	}
	if err := s.repo.DeletePage(ctx, project.ID, pageID); err != nil {
		return err
	}
	s.clearPublicSnapshots()
	return nil
}

func (s *Service) CreateElement(ctx context.Context, input CreateElementInput) (domainpublic.Element, error) {
	return s.saveElement(ctx, input.ProjectRef, input.CurrentUserID, func(projectID string) (domainpublic.Element, error) {
		return normalizeCreateElementInput(projectID, input.PageID, input)
	}, s.repo.CreateElement)
}

func (s *Service) UpdateElement(ctx context.Context, input UpdateElementInput) (domainpublic.Element, error) {
	return s.saveElement(ctx, input.ProjectRef, input.CurrentUserID, func(projectID string) (domainpublic.Element, error) {
		return normalizeUpdateElementInput(projectID, input.PageID, input)
	}, s.repo.UpdateElement)
}

func (s *Service) DeleteElement(ctx context.Context, input DeleteElementInput) error {
	project, err := s.loadWritableProject(ctx, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return err
	}
	pageID, err := domainpublic.VNPageID(input.PageID)
	if err != nil {
		return invalidField("pageId", err.Error(), input.PageID)
	}
	elementID, err := domainpublic.VNElementID(input.ElementID)
	if err != nil {
		return invalidField("elementId", err.Error(), input.ElementID)
	}
	if err := s.repo.DeleteElement(ctx, project.ID, pageID, elementID); err != nil {
		return err
	}
	s.clearPublicSnapshots()
	return nil
}

func (s *Service) validateElementReferences(ctx context.Context, element domainpublic.Element) error {
	if _, err := s.repo.GetPage(ctx, element.ProjectID, element.PublicPageID); err != nil {
		return err
	}
	if element.ID != "" {
		current, err := s.repo.GetElement(ctx, element.ProjectID, element.PublicPageID, element.ID)
		if err != nil {
			return err
		}
		if current.Kind != element.Kind {
			return invalidField("kind", "element kind cannot be changed", element.Kind)
		}
	}
	if element.ParentElementID != nil {
		parent, err := s.repo.GetElement(ctx, element.ProjectID, element.PublicPageID, *element.ParentElementID)
		if err != nil {
			return err
		}
		if err := validateParent(parent, element.ID); err != nil {
			return err
		}
	}
	if element.Kind == domainpublic.ElementKindAssignmentGroup {
		if err := s.validateAssignmentGroupScope(ctx, element); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) saveElement(
	ctx context.Context,
	projectRef string,
	userID string,
	normalize func(projectID string) (domainpublic.Element, error),
	save func(context.Context, domainpublic.Element) (domainpublic.Element, error),
) (domainpublic.Element, error) {
	project, err := s.loadWritableProject(ctx, projectRef, userID)
	if err != nil {
		return domainpublic.Element{}, err
	}
	element, err := normalize(project.ID)
	if err != nil {
		return domainpublic.Element{}, err
	}
	err = s.validateElementReferences(ctx, element)
	if err != nil {
		return domainpublic.Element{}, err
	}
	saved, err := save(ctx, element)
	if err != nil {
		return domainpublic.Element{}, err
	}
	s.clearPublicSnapshots()
	return saved, nil
}

func (s *Service) clearPublicSnapshots() {
	if s.snapshots != nil {
		s.snapshots.clear()
	}
}

func (s *Service) validateAssignmentGroupScope(ctx context.Context, element domainpublic.Element) error {
	if element.AssignmentSelectionMode == nil {
		return invalidField("assignmentSelectionMode", "must be provided for assignment groups", nil)
	}
	switch *element.AssignmentSelectionMode {
	case domainpublic.AssignmentSelectionModeAllCheck:
		if element.CheckID == nil {
			return invalidField("checkId", "must be provided for all-check assignment groups", nil)
		}
		ok, err := s.repo.HasAssignableCheck(ctx, element.ProjectID, *element.CheckID)
		if err != nil {
			return err
		}
		if !ok {
			return invalidField("checkId", "check must be an active ping or tcp check", *element.CheckID)
		}
	case domainpublic.AssignmentSelectionModeSelectedAssignments:
		count, err := s.repo.CountAssignableAssignments(ctx, element.ProjectID, element.AssignmentIDs)
		if err != nil {
			return err
		}
		if count != int64(len(element.AssignmentIDs)) {
			return invalidField("assignmentIds", "assignments must be active ping or tcp assignments in this project", element.AssignmentIDs)
		}
	}
	return nil
}
