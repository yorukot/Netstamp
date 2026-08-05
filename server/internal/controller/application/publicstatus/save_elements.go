package publicstatus

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
)

type pageElementChange struct {
	action    PublicStatusAction
	elementID string
}

type pageElementSavePlan struct {
	element  domainpublic.Element
	existing bool
	index    int
}

type pendingPageElement struct {
	input     SaveElementInput
	elementID string
	existing  bool
	index     int
}

func (s *Service) savePageElements(
	ctx context.Context,
	flow *publicStatusFlow,
	projectID string,
	pageID string,
	inputs *[]SaveElementInput,
	current []domainpublic.Element,
) ([]domainpublic.Element, []pageElementChange, error) {
	if inputs == nil {
		return append([]domainpublic.Element{}, current...), nil, nil
	}

	plan, err := s.normalizePageElementSavePlan(ctx, projectID, pageID, *inputs, current)
	if err != nil {
		return nil, nil, err
	}

	changes := make([]pageElementChange, 0, len(plan)+len(current))
	for _, kind := range []domainpublic.ElementKind{domainpublic.ElementKindFolder, domainpublic.ElementKindAssignmentGroup} {
		for _, planned := range plan {
			if planned.element.Kind != kind {
				continue
			}
			flow.setElementID(planned.element.ID)
			if planned.existing {
				saved, saveErr := s.repo.UpdateElement(ctx, planned.element)
				if saveErr != nil {
					return nil, nil, saveErr
				}
				changes = append(changes, pageElementChange{action: PublicStatusActionUpdateElement, elementID: saved.ID})
				continue
			}
			saved, saveErr := s.repo.CreateElement(ctx, planned.element)
			if saveErr != nil {
				return nil, nil, saveErr
			}
			changes = append(changes, pageElementChange{action: PublicStatusActionCreateElement, elementID: saved.ID})
		}
	}

	desiredIDs := make(map[string]struct{}, len(plan))
	for _, planned := range plan {
		desiredIDs[planned.element.ID] = struct{}{}
	}
	removed := append([]domainpublic.Element{}, current...)
	sort.SliceStable(removed, func(left, right int) bool {
		return removed[left].Kind == domainpublic.ElementKindAssignmentGroup && removed[right].Kind == domainpublic.ElementKindFolder
	})
	for _, element := range removed {
		if _, kept := desiredIDs[element.ID]; kept {
			continue
		}
		flow.setElementID(element.ID)
		if deleteErr := s.repo.DeleteElement(ctx, projectID, pageID, element.ID); deleteErr != nil {
			return nil, nil, deleteErr
		}
		changes = append(changes, pageElementChange{action: PublicStatusActionDeleteElement, elementID: element.ID})
	}

	saved, err := s.repo.ListElements(ctx, pageID)
	if err != nil {
		return nil, nil, err
	}
	return saved, changes, nil
}

func (s *Service) normalizePageElementSavePlan(
	ctx context.Context,
	projectID string,
	pageID string,
	inputs []SaveElementInput,
	current []domainpublic.Element,
) ([]pageElementSavePlan, error) {
	pending, elementIDByClientID, currentByID, err := indexPageElementInputs(inputs, current)
	if err != nil {
		return nil, err
	}
	plan, err := normalizePendingPageElements(projectID, pageID, pending, elementIDByClientID, currentByID)
	if err != nil {
		return nil, err
	}
	if err := s.validatePageElementSavePlan(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func indexPageElementInputs(
	inputs []SaveElementInput,
	current []domainpublic.Element,
) ([]pendingPageElement, map[string]string, map[string]domainpublic.Element, error) {
	currentByID := make(map[string]domainpublic.Element, len(current))
	for _, element := range current {
		currentByID[element.ID] = element
	}

	var collector appvalidation.Collector
	pending := make([]pendingPageElement, 0, len(inputs))
	elementIDByClientID := make(map[string]string, len(inputs))
	seenElementIDs := make(map[string]struct{}, len(inputs))
	for index, input := range inputs {
		path := pageElementField(index, "")
		clientID, clientErr := domainpublic.VNElementClientID(input.ClientID)
		collector.AddError(path+"clientId", clientErr, input.ClientID)
		if clientErr == nil {
			if _, duplicate := elementIDByClientID[clientID]; duplicate {
				collector.Add(path+"clientId", "must be unique within the element set", input.ClientID)
			}
		}

		elementID := uuid.NewString()
		existing := input.ElementID != nil
		if existing {
			normalizedID, idErr := domainpublic.VNElementID(*input.ElementID)
			collector.AddError(path+"id", idErr, *input.ElementID)
			if idErr == nil {
				elementID = normalizedID
				if _, found := currentByID[elementID]; !found {
					collector.Add(path+"id", "element does not belong to this public status page", elementID)
				}
			}
		}
		if _, duplicate := seenElementIDs[elementID]; duplicate {
			collector.Add(path+"id", "must be unique within the element set", elementID)
		}
		seenElementIDs[elementID] = struct{}{}
		if clientErr == nil {
			elementIDByClientID[clientID] = elementID
		}
		pending = append(pending, pendingPageElement{input: input, elementID: elementID, existing: existing, index: index})
	}
	if err := collector.Err(ErrInvalidInput); err != nil {
		return nil, nil, nil, err
	}
	return pending, elementIDByClientID, currentByID, nil
}

func normalizePendingPageElements(
	projectID string,
	pageID string,
	pending []pendingPageElement,
	elementIDByClientID map[string]string,
	currentByID map[string]domainpublic.Element,
) ([]pageElementSavePlan, error) {
	var collector appvalidation.Collector
	plan := make([]pageElementSavePlan, 0, len(pending))
	for _, pendingElement := range pending {
		parentElementID := resolveParentElementID(&collector, pendingElement, elementIDByClientID)

		element, elementErr := normalizeElement(
			domainpublic.Element{ID: pendingElement.elementID, ProjectID: projectID, PublicPageID: pageID},
			parentElementID,
			pendingElement.input.Kind,
			pendingElement.input.CheckID,
			pendingElement.input.AssignmentSelectionMode,
			pendingElement.input.AssignmentIDs,
			pendingElement.input.Title,
			pendingElement.input.Description,
			pendingElement.input.SortOrder,
			pendingElement.input.DisplayMode,
			pendingElement.input.ChartMode,
			pendingElement.input.ChartRange,
		)
		addPageElementValidation(&collector, pendingElement.index, elementErr)
		if elementErr != nil {
			continue
		}
		if existingElement, exists := currentByID[element.ID]; exists && existingElement.Kind != element.Kind {
			collector.Add(pageElementField(pendingElement.index, "kind"), "element kind cannot be changed", element.Kind)
		}
		plan = append(plan, pageElementSavePlan{element: element, existing: pendingElement.existing, index: pendingElement.index})
	}
	if err := collector.Err(ErrInvalidInput); err != nil {
		return nil, err
	}
	return plan, nil
}

func resolveParentElementID(collector *appvalidation.Collector, pending pendingPageElement, elementIDByClientID map[string]string) *string {
	if pending.input.ParentClientID == nil {
		return nil
	}
	parentClientID, err := domainpublic.VNElementClientID(*pending.input.ParentClientID)
	collector.AddError(pageElementField(pending.index, "parentClientId"), err, *pending.input.ParentClientID)
	if err != nil {
		return nil
	}
	resolvedID, found := elementIDByClientID[parentClientID]
	if !found {
		collector.Add(pageElementField(pending.index, "parentClientId"), "must reference an element in the same element set", parentClientID)
		return nil
	}
	return &resolvedID
}

func (s *Service) validatePageElementSavePlan(ctx context.Context, plan []pageElementSavePlan) error {
	var collector appvalidation.Collector
	desiredByID := make(map[string]domainpublic.Element, len(plan))
	for _, planned := range plan {
		desiredByID[planned.element.ID] = planned.element
	}
	for _, planned := range plan {
		element := planned.element
		if element.ParentElementID != nil {
			parent, found := desiredByID[*element.ParentElementID]
			if !found {
				collector.Add(pageElementField(planned.index, "parentClientId"), "must reference an element in the same element set", *element.ParentElementID)
			} else if parentErr := validateParent(parent, element.ID); parentErr != nil {
				addPageElementValidation(&collector, planned.index, parentErr)
			}
		}
		if element.Kind == domainpublic.ElementKindAssignmentGroup {
			if scopeErr := s.validateAssignmentGroupScope(ctx, element); scopeErr != nil {
				if !addPageElementValidation(&collector, planned.index, scopeErr) {
					return scopeErr
				}
			}
		}
	}
	return collector.Err(ErrInvalidInput)
}

func addPageElementValidation(collector *appvalidation.Collector, index int, err error) bool {
	if err == nil {
		return true
	}
	fields, ok := appvalidation.FieldErrors(err)
	if !ok {
		return false
	}
	for _, field := range fields {
		if field.Field == "parentElementId" {
			field.Field = "parentClientId"
		}
		field.Field = pageElementField(index, field.Field)
		collector.AddFields(field)
	}
	return true
}

func pageElementField(index int, field string) string {
	if field == "" {
		return fmt.Sprintf("elements.%d.", index)
	}
	return fmt.Sprintf("elements.%d.%s", index, field)
}
