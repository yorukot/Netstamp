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
	ctx, flow := s.startPublicStatusFlow(ctx, "publicstatus.page.create", PublicStatusActionCreatePage, input.CurrentUserID)
	defer flow.end()
	flow.setProjectRef(input.ProjectRef)

	project, err := s.loadProjectForFlow(ctx, flow, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainpublic.Page{}, err
	}
	if err := s.requireProjectWriteForFlow(ctx, flow, project.ID, input.CurrentUserID); err != nil {
		return domainpublic.Page{}, err
	}
	page, err := normalizeCreatePageInput(project.ID, input)
	if err != nil {
		return domainpublic.Page{}, flow.writeFailure(PublicStatusReasonPageCreateFailed, err)
	}
	created, err := s.repo.CreatePage(ctx, page)
	if err != nil {
		return domainpublic.Page{}, flow.writeFailure(PublicStatusReasonPageCreateFailed, err)
	}
	flow.setPageID(created.ID)
	flow.success()
	s.clearPublicSnapshots()
	return created, nil
}

func (s *Service) UpdatePage(ctx context.Context, input UpdatePageInput) (domainpublic.Page, error) {
	ctx, flow := s.startPublicStatusFlow(ctx, "publicstatus.page.update", PublicStatusActionUpdatePage, input.CurrentUserID)
	defer flow.end()
	flow.setProjectRef(input.ProjectRef)

	project, err := s.loadProjectForFlow(ctx, flow, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return domainpublic.Page{}, err
	}
	if err := s.requireProjectWriteForFlow(ctx, flow, project.ID, input.CurrentUserID); err != nil {
		return domainpublic.Page{}, err
	}
	page, err := normalizeUpdatePageInput(project.ID, input)
	if err != nil {
		return domainpublic.Page{}, flow.writeFailure(PublicStatusReasonPageUpdateFailed, err)
	}
	flow.setPageID(page.ID)
	updated, err := s.repo.UpdatePage(ctx, page)
	if err != nil {
		return domainpublic.Page{}, flow.writeFailure(PublicStatusReasonPageUpdateFailed, err)
	}
	flow.setPageID(updated.ID)
	flow.success()
	s.clearPublicSnapshots()
	return updated, nil
}

func (s *Service) DeletePage(ctx context.Context, input DeletePageInput) error {
	ctx, flow := s.startPublicStatusFlow(ctx, "publicstatus.page.delete", PublicStatusActionDeletePage, input.CurrentUserID)
	defer flow.end()
	flow.setProjectRef(input.ProjectRef)

	project, err := s.loadProjectForFlow(ctx, flow, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return err
	}
	if err := s.requireProjectWriteForFlow(ctx, flow, project.ID, input.CurrentUserID); err != nil {
		return err
	}
	pageID, err := domainpublic.VNPageID(input.PageID)
	if err != nil {
		return flow.writeFailure(PublicStatusReasonPageDeleteFailed, invalidField("pageId", err.Error(), input.PageID))
	}
	flow.setPageID(pageID)
	if err := s.repo.DeletePage(ctx, project.ID, pageID); err != nil {
		return flow.writeFailure(PublicStatusReasonPageDeleteFailed, err)
	}
	flow.success()
	s.clearPublicSnapshots()
	return nil
}

func (s *Service) CreateElement(ctx context.Context, input CreateElementInput) (domainpublic.Element, error) {
	ctx, flow := s.startPublicStatusFlow(ctx, "publicstatus.element.create", PublicStatusActionCreateElement, input.CurrentUserID)
	defer flow.end()
	flow.setProjectRef(input.ProjectRef)

	element, err := s.saveElement(ctx, flow, input.ProjectRef, input.CurrentUserID, PublicStatusReasonElementCreateFailed, func(projectID string) (domainpublic.Element, error) {
		return normalizeCreateElementInput(projectID, input.PageID, input)
	}, s.repo.CreateElement)
	if err != nil {
		return domainpublic.Element{}, err
	}
	flow.success()
	return element, nil
}

func (s *Service) UpdateElement(ctx context.Context, input UpdateElementInput) (domainpublic.Element, error) {
	ctx, flow := s.startPublicStatusFlow(ctx, "publicstatus.element.update", PublicStatusActionUpdateElement, input.CurrentUserID)
	defer flow.end()
	flow.setProjectRef(input.ProjectRef)

	element, err := s.saveElement(ctx, flow, input.ProjectRef, input.CurrentUserID, PublicStatusReasonElementUpdateFailed, func(projectID string) (domainpublic.Element, error) {
		return normalizeUpdateElementInput(projectID, input.PageID, input)
	}, s.repo.UpdateElement)
	if err != nil {
		return domainpublic.Element{}, err
	}
	flow.success()
	return element, nil
}

func (s *Service) DeleteElement(ctx context.Context, input DeleteElementInput) error {
	ctx, flow := s.startPublicStatusFlow(ctx, "publicstatus.element.delete", PublicStatusActionDeleteElement, input.CurrentUserID)
	defer flow.end()
	flow.setProjectRef(input.ProjectRef)

	project, err := s.loadProjectForFlow(ctx, flow, input.ProjectRef, input.CurrentUserID)
	if err != nil {
		return err
	}
	if err := s.requireProjectWriteForFlow(ctx, flow, project.ID, input.CurrentUserID); err != nil {
		return err
	}
	pageID, err := domainpublic.VNPageID(input.PageID)
	if err != nil {
		return flow.writeFailure(PublicStatusReasonElementDeleteFailed, invalidField("pageId", err.Error(), input.PageID))
	}
	flow.setPageID(pageID)
	elementID, err := domainpublic.VNElementID(input.ElementID)
	if err != nil {
		return flow.writeFailure(PublicStatusReasonElementDeleteFailed, invalidField("elementId", err.Error(), input.ElementID))
	}
	flow.setElementID(elementID)
	if err := s.repo.DeleteElement(ctx, project.ID, pageID, elementID); err != nil {
		return flow.writeFailure(PublicStatusReasonElementDeleteFailed, err)
	}
	flow.success()
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
	flow *publicStatusFlow,
	projectRef string,
	userID string,
	technicalReason PublicStatusReason,
	normalize func(projectID string) (domainpublic.Element, error),
	save func(context.Context, domainpublic.Element) (domainpublic.Element, error),
) (domainpublic.Element, error) {
	project, err := s.loadProjectForFlow(ctx, flow, projectRef, userID)
	if err != nil {
		return domainpublic.Element{}, err
	}
	if err := s.requireProjectWriteForFlow(ctx, flow, project.ID, userID); err != nil {
		return domainpublic.Element{}, err
	}
	element, err := normalize(project.ID)
	if err != nil {
		return domainpublic.Element{}, flow.writeFailure(technicalReason, err)
	}
	flow.setPageID(element.PublicPageID)
	flow.setElementID(element.ID)
	err = s.validateElementReferences(ctx, element)
	if err != nil {
		return domainpublic.Element{}, flow.writeFailure(technicalReason, err)
	}
	saved, err := save(ctx, element)
	if err != nil {
		return domainpublic.Element{}, flow.writeFailure(technicalReason, err)
	}
	flow.setElementID(saved.ID)
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
