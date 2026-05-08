package check

import (
	"context"
	"encoding/json"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	"github.com/yorukot/netstamp/internal/normalize"
)

const (
	defaultPacketCount     = 4
	defaultPacketSizeBytes = 56
	defaultTimeoutMs       = 3000
	maxPacketSizeBytes     = 65507
)

var defaultSelector = json.RawMessage(`{}`)

type Service struct {
	repo          Repository
	projectAccess ProjectAccess
	labelAccess   LabelAccess
	events        EventRecorder
}

func NewService(repo Repository, projectAccess ProjectAccess, labelAccess LabelAccess, events EventRecorder) *Service {
	return &Service{
		repo:          repo,
		projectAccess: projectAccess,
		labelAccess:   labelAccess,
		events:        events,
	}
}

func (s *Service) ListChecks(ctx context.Context, input ListChecksInput) ([]domaincheck.Check, error) {
	ctx, flow := s.startCheckFlow(ctx, "check.list", CheckActionList, input.CurrentUserID)
	defer flow.end()
	flow.setProjectRef(input.ProjectRef)

	project, err := s.loadProject(ctx, flow, input.ProjectRef, input.CurrentUserID, CheckEventListFailure)
	if err != nil {
		return nil, err
	}

	checks, err := s.repo.ListChecks(ctx, project.ID)
	if err != nil {
		return nil, flow.technicalFailure(CheckEventListFailure, CheckReasonCheckListFailed, err)
	}

	return checks, nil
}

func (s *Service) GetCheck(ctx context.Context, input GetCheckInput) (domaincheck.Check, error) {
	ctx, flow := s.startCheckFlow(ctx, "check.get", CheckActionGet, input.CurrentUserID)
	defer flow.end()
	flow.setProjectRef(input.ProjectRef)
	flow.setCheckID(input.CheckID)

	project, err := s.loadProject(ctx, flow, input.ProjectRef, input.CurrentUserID, CheckEventGetFailure)
	if err != nil {
		return domaincheck.Check{}, err
	}

	check, err := s.repo.GetCheck(ctx, project.ID, input.CheckID)
	if err != nil {
		return domaincheck.Check{}, flow.checkLookupFailure(CheckEventGetFailure, err)
	}

	return check, nil
}

func (s *Service) CreateCheck(ctx context.Context, input CreateCheckInput) (domaincheck.Check, error) {
	ctx, flow := s.startCheckFlow(ctx, "check.create", CheckActionCreate, input.CurrentUserID)
	defer flow.end()
	flow.setProjectRef(input.ProjectRef)

	normalized, err := normalizeCreateCheckInput(input)
	if err != nil {
		return domaincheck.Check{}, flow.businessFailure(CheckEventCreateFailure, CheckReasonInvalidInput, err)
	}

	project, err := s.loadProject(ctx, flow, input.ProjectRef, input.CurrentUserID, CheckEventCreateFailure)
	if err != nil {
		return domaincheck.Check{}, err
	}
	if err := s.requireAction(ctx, flow, project.ID, input.CurrentUserID, CheckEventCreateFailure); err != nil {
		return domaincheck.Check{}, err
	}

	labels, err := s.labelAccess.GetActiveLabelsByIDsForProject(ctx, project.ID, normalized.labelIDs)
	if err != nil {
		return domaincheck.Check{}, flow.labelLookupFailure(CheckEventCreateFailure, err)
	}

	check, err := s.repo.CreateCheck(ctx, domaincheck.CreateCheckStorageInput{
		ProjectID:       project.ID,
		Name:            normalized.name,
		Type:            normalized.checkType,
		Target:          normalized.target,
		Selector:        normalized.selector,
		Description:     normalized.description,
		IntervalSeconds: normalized.intervalSeconds,
		PingConfig:      normalized.pingConfig,
		LabelIDs:        normalized.labelIDs,
	})
	if err != nil {
		return domaincheck.Check{}, flow.writeFailure(CheckEventCreateFailure, CheckReasonCheckCreateFailed, err)
	}
	check.Labels = labels
	flow.setCheckID(check.ID)
	flow.success(CheckEventCreateSuccess)

	return check, nil
}

func (s *Service) UpdateCheck(ctx context.Context, input UpdateCheckInput) (domaincheck.Check, error) {
	ctx, flow := s.startCheckFlow(ctx, "check.update", CheckActionUpdate, input.CurrentUserID)
	defer flow.end()
	flow.setProjectRef(input.ProjectRef)
	flow.setCheckID(input.CheckID)

	project, err := s.loadProject(ctx, flow, input.ProjectRef, input.CurrentUserID, CheckEventUpdateFailure)
	if err != nil {
		return domaincheck.Check{}, err
	}
	if err := s.requireAction(ctx, flow, project.ID, input.CurrentUserID, CheckEventUpdateFailure); err != nil {
		return domaincheck.Check{}, err
	}

	normalized, err := normalizeUpdateCheckInput(input)
	if err != nil {
		return domaincheck.Check{}, flow.businessFailure(CheckEventUpdateFailure, CheckReasonInvalidInput, err)
	}

	current, err := s.repo.GetCheck(ctx, project.ID, input.CheckID)
	if err != nil {
		return domaincheck.Check{}, flow.checkLookupFailure(CheckEventUpdateFailure, err)
	}

	labelIDs := []string(nil)
	resolvedLabels := current.Labels
	if normalized.replaceLabels {
		labelIDs = normalized.labelIDs
		resolvedLabels, err = s.labelAccess.GetActiveLabelsByIDsForProject(ctx, project.ID, labelIDs)
		if err != nil {
			return domaincheck.Check{}, flow.labelLookupFailure(CheckEventUpdateFailure, err)
		}
	}

	updated := domaincheck.UpdateCheckStorageInput{
		ProjectID:       project.ID,
		CheckID:         input.CheckID,
		Name:            chooseString(current.Name, normalized.name),
		Type:            chooseCheckType(current.Type, normalized.checkType),
		Target:          chooseString(current.Target, normalized.target),
		Selector:        chooseRawMessage(current.Selector, normalized.selector),
		Description:     chooseOptionalString(current.Description, normalized.description),
		IntervalSeconds: chooseInt(current.IntervalSeconds, normalized.intervalSeconds),
		PingConfig: domaincheck.PingConfig{
			PacketCount:     chooseInt(current.PingConfig.PacketCount, normalized.packetCount),
			PacketSizeBytes: chooseInt(current.PingConfig.PacketSizeBytes, normalized.packetSizeBytes),
			TimeoutMs:       chooseInt(current.PingConfig.TimeoutMs, normalized.timeoutMs),
			IPFamily:        chooseIPFamily(current.PingConfig.IPFamily, normalized.ipFamily),
		},
		ReplaceLabels: normalized.replaceLabels,
		LabelIDs:      labelIDs,
	}

	check, err := s.repo.UpdateCheck(ctx, updated)
	if err != nil {
		return domaincheck.Check{}, flow.writeFailure(CheckEventUpdateFailure, CheckReasonCheckUpdateFailed, err)
	}
	if normalized.replaceLabels {
		check.Labels = resolvedLabels
	}
	flow.success(CheckEventUpdateSuccess)

	return check, nil
}

func (s *Service) DeleteCheck(ctx context.Context, input GetCheckInput) error {
	ctx, flow := s.startCheckFlow(ctx, "check.delete", CheckActionDelete, input.CurrentUserID)
	defer flow.end()
	flow.setProjectRef(input.ProjectRef)
	flow.setCheckID(input.CheckID)

	project, err := s.loadProject(ctx, flow, input.ProjectRef, input.CurrentUserID, CheckEventDeleteFailure)
	if err != nil {
		return err
	}
	if err := s.requireAction(ctx, flow, project.ID, input.CurrentUserID, CheckEventDeleteFailure); err != nil {
		return err
	}

	if err := s.repo.SoftDeleteCheck(ctx, project.ID, input.CheckID); err != nil {
		return flow.writeFailure(CheckEventDeleteFailure, CheckReasonCheckDeleteFailed, err)
	}
	flow.success(CheckEventDeleteSuccess)

	return nil
}

func (s *Service) loadProject(ctx context.Context, flow *checkFlow, projectRef string, userID string, failureEvent CheckEventName) (domainproject.Project, error) {
	project, err := s.projectAccess.GetProjectForUser(ctx, projectRef, userID)
	if err != nil {
		return domainproject.Project{}, flow.projectLookupFailure(failureEvent, err)
	}
	flow.setProject(project)

	return project, nil
}

func (s *Service) requireAction(ctx context.Context, flow *checkFlow, projectID string, userID string, failureEvent CheckEventName) error {
	role, err := s.projectAccess.GetMemberRole(ctx, projectID, userID)
	if err != nil {
		return flow.roleLookupFailure(failureEvent, err)
	}
	if !domainproject.Can(role, domainproject.ActionManageChecks) {
		return flow.businessFailure(failureEvent, CheckReasonForbidden, ErrForbidden)
	}

	return nil
}

func normalizeCreateCheckInput(input CreateCheckInput) (normalizedCreateCheckInput, error) {
	name, err := normalize.RequiredString(input.Name, ErrInvalidInput)
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}
	checkType, err := normalizeCheckType(input.Type)
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}
	target, err := normalize.RequiredString(input.Target, ErrInvalidInput)
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}
	selector, err := normalizeSelector(input.Selector)
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}
	description, err := normalize.OptionalString(input.Description, ErrInvalidInput)
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}
	if input.IntervalSeconds <= 0 {
		return normalizedCreateCheckInput{}, ErrInvalidInput
	}
	pingConfig, err := normalizePingConfig(input.PacketCount, input.PacketSizeBytes, input.TimeoutMs, input.IPFamily)
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}
	labelIDs, err := normalize.CanonicalUUIDSet(input.LabelIDs, ErrInvalidInput)
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}

	return normalizedCreateCheckInput{
		name:            name,
		checkType:       checkType,
		target:          target,
		selector:        selector,
		description:     description,
		intervalSeconds: input.IntervalSeconds,
		pingConfig:      pingConfig,
		labelIDs:        labelIDs,
	}, nil
}

func normalizeUpdateCheckInput(input UpdateCheckInput) (normalizedUpdateCheckInput, error) {
	if input.Name == nil &&
		input.Type == nil &&
		input.Target == nil &&
		input.Selector == nil &&
		input.Description == nil &&
		input.IntervalSeconds == nil &&
		input.PacketCount == nil &&
		input.PacketSizeBytes == nil &&
		input.TimeoutMs == nil &&
		input.IPFamily == nil &&
		input.LabelIDs == nil {
		return normalizedUpdateCheckInput{}, ErrInvalidInput
	}

	var normalized normalizedUpdateCheckInput
	var err error
	if input.Name != nil {
		normalized.name = new(string)
		*normalized.name, err = normalize.RequiredString(*input.Name, ErrInvalidInput)
		if err != nil {
			return normalizedUpdateCheckInput{}, err
		}
	}
	if input.Type != nil {
		checkType, err := normalizeCheckType(*input.Type)
		if err != nil {
			return normalizedUpdateCheckInput{}, err
		}
		normalized.checkType = &checkType
	}
	if input.Target != nil {
		normalized.target = new(string)
		*normalized.target, err = normalize.RequiredString(*input.Target, ErrInvalidInput)
		if err != nil {
			return normalizedUpdateCheckInput{}, err
		}
	}
	if input.Selector != nil {
		normalized.selector, err = normalizeSelector(input.Selector)
		if err != nil {
			return normalizedUpdateCheckInput{}, err
		}
	}
	if input.Description != nil {
		normalized.description, err = normalize.OptionalString(input.Description, ErrInvalidInput)
		if err != nil {
			return normalizedUpdateCheckInput{}, err
		}
	}
	if input.IntervalSeconds != nil {
		if *input.IntervalSeconds <= 0 {
			return normalizedUpdateCheckInput{}, ErrInvalidInput
		}
		intervalSeconds := *input.IntervalSeconds
		normalized.intervalSeconds = &intervalSeconds
	}
	if input.PacketCount != nil {
		if *input.PacketCount <= 0 {
			return normalizedUpdateCheckInput{}, ErrInvalidInput
		}
		packetCount := *input.PacketCount
		normalized.packetCount = &packetCount
	}
	if input.PacketSizeBytes != nil {
		if *input.PacketSizeBytes < 0 || *input.PacketSizeBytes > maxPacketSizeBytes {
			return normalizedUpdateCheckInput{}, ErrInvalidInput
		}
		packetSizeBytes := *input.PacketSizeBytes
		normalized.packetSizeBytes = &packetSizeBytes
	}
	if input.TimeoutMs != nil {
		if *input.TimeoutMs <= 0 {
			return normalizedUpdateCheckInput{}, ErrInvalidInput
		}
		timeoutMs := *input.TimeoutMs
		normalized.timeoutMs = &timeoutMs
	}
	if input.IPFamily != nil {
		ipFamily, err := normalizeIPFamily(*input.IPFamily)
		if err != nil {
			return normalizedUpdateCheckInput{}, err
		}
		normalized.ipFamily = &ipFamily
	}
	if input.LabelIDs != nil {
		normalized.replaceLabels = true
		normalized.labelIDs, err = normalize.CanonicalUUIDSet(*input.LabelIDs, ErrInvalidInput)
		if err != nil {
			return normalizedUpdateCheckInput{}, err
		}
	}

	return normalized, nil
}

func normalizeCheckType(value string) (domaincheck.Type, error) {
	value, err := normalize.RequiredString(value, ErrInvalidInput)
	if err != nil {
		return "", err
	}
	if domaincheck.Type(value) != domaincheck.TypePing {
		return "", ErrInvalidInput
	}

	return domaincheck.TypePing, nil
}

func normalizeSelector(selector map[string]any) (json.RawMessage, error) {
	if selector == nil {
		return defaultSelector, nil
	}

	raw, err := json.Marshal(selector)
	if err != nil {
		return nil, ErrInvalidInput
	}

	return raw, nil
}

func normalizePingConfig(packetCount *int, packetSizeBytes *int, timeoutMs *int, ipFamilyValue *string) (domaincheck.PingConfig, error) {
	config := domaincheck.PingConfig{
		PacketCount:     defaultPacketCount,
		PacketSizeBytes: defaultPacketSizeBytes,
		TimeoutMs:       defaultTimeoutMs,
	}

	if packetCount != nil {
		if *packetCount <= 0 {
			return domaincheck.PingConfig{}, ErrInvalidInput
		}
		config.PacketCount = *packetCount
	}
	if packetSizeBytes != nil {
		if *packetSizeBytes < 0 || *packetSizeBytes > maxPacketSizeBytes {
			return domaincheck.PingConfig{}, ErrInvalidInput
		}
		config.PacketSizeBytes = *packetSizeBytes
	}
	if timeoutMs != nil {
		if *timeoutMs <= 0 {
			return domaincheck.PingConfig{}, ErrInvalidInput
		}
		config.TimeoutMs = *timeoutMs
	}
	if ipFamilyValue != nil {
		ipFamily, err := normalizeIPFamily(*ipFamilyValue)
		if err != nil {
			return domaincheck.PingConfig{}, err
		}
		config.IPFamily = &ipFamily
	}

	return config, nil
}

func normalizeIPFamily(value string) (domaincheck.IPFamily, error) {
	value, err := normalize.RequiredString(value, ErrInvalidInput)
	if err != nil {
		return "", err
	}
	switch domaincheck.IPFamily(value) {
	case domaincheck.IPFamilyIPv4, domaincheck.IPFamilyIPv6:
		return domaincheck.IPFamily(value), nil
	default:
		return "", ErrInvalidInput
	}
}

func chooseString(current string, next *string) string {
	if next == nil {
		return current
	}

	return *next
}

func chooseCheckType(current domaincheck.Type, next *domaincheck.Type) domaincheck.Type {
	if next == nil {
		return current
	}

	return *next
}

func chooseRawMessage(current json.RawMessage, next json.RawMessage) json.RawMessage {
	if next == nil {
		return current
	}

	return next
}

func chooseOptionalString(current *string, next *string) *string {
	if next == nil {
		return current
	}

	return next
}

func chooseInt(current int, next *int) int {
	if next == nil {
		return current
	}

	return *next
}

func chooseIPFamily(current *domaincheck.IPFamily, next *domaincheck.IPFamily) *domaincheck.IPFamily {
	if next == nil {
		return current
	}

	return next
}
