package check

import (
	"context"
	"encoding/json"

	appvalidation "github.com/yorukot/netstamp/internal/application/validation"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainselector "github.com/yorukot/netstamp/internal/domain/selector"
	"github.com/yorukot/netstamp/internal/normalize"
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
	err = s.requireAction(ctx, flow, project.ID, input.CurrentUserID, CheckEventCreateFailure)
	if err != nil {
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
		CheckVersion:    domaincheck.CheckVersion(normalized.executionSpec()),
		SelectorVersion: domaincheck.SelectorVersion(normalized.selector),
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
	err = s.requireAction(ctx, flow, project.ID, input.CurrentUserID, CheckEventUpdateFailure)
	if err != nil {
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

	updatedPingConfig := domainping.Config{
		PacketCount:     chooseInt32(current.PingConfig.PacketCount, normalized.packetCount),
		PacketSizeBytes: chooseInt32(current.PingConfig.PacketSizeBytes, normalized.packetSizeBytes),
		TimeoutMs:       chooseInt32(current.PingConfig.TimeoutMs, normalized.timeoutMs),
		IPFamily:        chooseIPFamily(current.PingConfig.IPFamily, normalized.ipFamily),
	}
	updatedSelector, err := canonicalizeSelector(chooseRawMessage(current.Selector, normalized.selector))
	if err != nil {
		return domaincheck.Check{}, flow.businessFailure(CheckEventUpdateFailure, CheckReasonInvalidInput, err)
	}

	updated := domaincheck.UpdateCheckStorageInput{
		ProjectID:       project.ID,
		CheckID:         input.CheckID,
		Name:            chooseString(current.Name, normalized.name),
		Type:            chooseCheckType(current.Type, normalized.checkType),
		Target:          chooseString(current.Target, normalized.target),
		Selector:        updatedSelector,
		Description:     chooseOptionalString(current.Description, normalized.description),
		IntervalSeconds: chooseInt32(current.IntervalSeconds, normalized.intervalSeconds),
		PingConfig:      updatedPingConfig,
		ReplaceLabels:   normalized.replaceLabels,
		LabelIDs:        labelIDs,
	}
	updated.CheckVersion = domaincheck.CheckVersion(domaincheck.ExecutionSpec{
		Type:            updated.Type,
		Target:          updated.Target,
		IntervalSeconds: updated.IntervalSeconds,
		PingConfig:      updated.PingConfig,
	})
	updated.SelectorVersion = domaincheck.SelectorVersion(updated.Selector)

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

func (s *Service) loadProject(ctx context.Context, flow *checkFlow, projectRef, userID string, failureEvent CheckEventName) (domainproject.Project, error) {
	project, err := s.projectAccess.GetProjectForUser(ctx, projectRef, userID)
	if err != nil {
		return domainproject.Project{}, flow.projectLookupFailure(failureEvent, err)
	}
	flow.setProject(project)

	return project, nil
}

func (s *Service) requireAction(ctx context.Context, flow *checkFlow, projectID, userID string, failureEvent CheckEventName) error {
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
	name, err := normalizeRequiredStringField(input.Name, "name")
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}
	checkType, err := normalizeCheckType(input.Type)
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}
	target, err := normalizeRequiredStringField(input.Target, "target")
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}
	selector, err := normalizeCreateSelector(input.Selector)
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}
	description, err := normalizeOptionalStringField(input.Description, "description")
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}
	if input.IntervalSeconds <= 0 {
		return normalizedCreateCheckInput{}, invalidCheckField("intervalSeconds", "must be greater than 0", input.IntervalSeconds)
	}
	pingConfig, err := normalizePingConfig(input.PacketCount, input.PacketSizeBytes, input.TimeoutMs, input.IPFamily)
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}
	labelIDs, err := normalizeLabelIDs(input.LabelIDs)
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

func (input normalizedCreateCheckInput) executionSpec() domaincheck.ExecutionSpec {
	return domaincheck.ExecutionSpec{
		Type:            input.checkType,
		Target:          input.target,
		IntervalSeconds: input.intervalSeconds,
		PingConfig:      input.pingConfig,
	}
}

func normalizeUpdateCheckInput(input UpdateCheckInput) (normalizedUpdateCheckInput, error) {
	if !hasUpdateCheckChanges(input) {
		return normalizedUpdateCheckInput{}, invalidCheckField("", "at least one field must be provided", nil)
	}

	name, err := normalizeOptionalRequiredString(input.Name, "name")
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	checkType, err := normalizeOptionalCheckType(input.Type)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	target, err := normalizeOptionalRequiredString(input.Target, "target")
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	selector, err := normalizeOptionalSelector(input.Selector)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	description, err := normalizeOptionalStringField(input.Description, "description")
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	intervalSeconds, err := normalizeOptionalPositiveInt(input.IntervalSeconds, "intervalSeconds")
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	packetCount, err := normalizeOptionalPacketCount(input.PacketCount)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	packetSizeBytes, err := normalizeOptionalPacketSizeBytes(input.PacketSizeBytes)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	timeoutMs, err := normalizeOptionalTimeoutMs(input.TimeoutMs)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	ipFamily, err := normalizeOptionalIPFamily(input.IPFamily)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	replaceLabels, labelIDs, err := normalizeOptionalLabelIDs(input.LabelIDs)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}

	return normalizedUpdateCheckInput{
		name:            name,
		checkType:       checkType,
		target:          target,
		selector:        selector,
		description:     description,
		intervalSeconds: intervalSeconds,
		packetCount:     packetCount,
		packetSizeBytes: packetSizeBytes,
		timeoutMs:       timeoutMs,
		ipFamily:        ipFamily,
		replaceLabels:   replaceLabels,
		labelIDs:        labelIDs,
	}, nil
}

func hasUpdateCheckChanges(input UpdateCheckInput) bool {
	return input.Name != nil ||
		input.Type != nil ||
		input.Target != nil ||
		input.Selector != nil ||
		input.Description != nil ||
		input.IntervalSeconds != nil ||
		input.PacketCount != nil ||
		input.PacketSizeBytes != nil ||
		input.TimeoutMs != nil ||
		input.IPFamily != nil ||
		input.LabelIDs != nil
}

func invalidCheckField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}

func normalizeRequiredStringField(value, field string) (string, error) {
	normalized, err := normalize.RequiredString(value, ErrInvalidInput)
	if err != nil {
		return "", invalidCheckField(field, "must not be blank", value)
	}

	return normalized, nil
}

func normalizeOptionalRequiredString(value *string, field string) (*string, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // A nil pointer is the explicit representation of an omitted optional field.
	}

	normalized, err := normalizeRequiredStringField(*value, field)
	if err != nil {
		return nil, err
	}

	return &normalized, nil
}

func normalizeOptionalStringField(value *string, field string) (*string, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // A nil pointer is the explicit representation of an omitted optional field.
	}

	normalized, err := normalize.OptionalString(value, ErrInvalidInput)
	if err != nil {
		return nil, invalidCheckField(field, "must not be blank", *value)
	}

	return normalized, nil
}

func normalizeOptionalCheckType(value *string) (*domaincheck.Type, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // A nil pointer is the explicit representation of an omitted optional field.
	}

	checkType, err := normalizeCheckType(*value)
	if err != nil {
		return nil, err
	}

	return &checkType, nil
}

func normalizeOptionalSelector(selector map[string]any) (json.RawMessage, error) {
	if selector == nil {
		return nil, nil
	}

	raw, err := normalizeSelector(selector)
	if err != nil {
		return nil, err
	}

	canonical, err := canonicalizeSelector(raw)
	if err != nil {
		return nil, invalidCheckField("selector", "must be a valid selector", selector)
	}

	return canonical, nil
}

func normalizeCreateSelector(selector map[string]any) (json.RawMessage, error) {
	if selector == nil {
		return canonicalizeSelector(nil)
	}

	raw, err := normalizeSelector(selector)
	if err != nil {
		return nil, err
	}

	canonical, err := canonicalizeSelector(raw)
	if err != nil {
		return nil, invalidCheckField("selector", "must be a valid selector", selector)
	}

	return canonical, nil
}

func canonicalizeSelector(raw json.RawMessage) (json.RawMessage, error) {
	parsed, err := domainselector.Parse(raw)
	if err != nil {
		return nil, ErrInvalidInput
	}

	return parsed.CanonicalJSON(), nil
}

func normalizeOptionalPositiveInt(value *int32, field string) (*int32, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // A nil pointer is the explicit representation of an omitted optional field.
	}
	if *value <= 0 {
		return nil, invalidCheckField(field, "must be greater than 0", *value)
	}

	normalized := *value
	return &normalized, nil
}

func normalizeOptionalPacketCount(value *int32) (*int32, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // A nil pointer is the explicit representation of an omitted optional field.
	}
	if err := domainping.ValidatePacketCount(*value); err != nil {
		return nil, invalidCheckField("packetCount", "must be greater than 0", *value)
	}

	normalized := *value
	return &normalized, nil
}

func normalizeOptionalPacketSizeBytes(value *int32) (*int32, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // A nil pointer is the explicit representation of an omitted optional field.
	}
	if err := domainping.ValidatePacketSizeBytes(*value); err != nil {
		return nil, invalidCheckField("packetSizeBytes", "must be between 0 and 65507", *value)
	}

	normalized := *value
	return &normalized, nil
}

func normalizeOptionalTimeoutMs(value *int32) (*int32, error) {
	if value == nil {
		return nil, nil //nolint:nilnil // A nil pointer is the explicit representation of an omitted optional field.
	}
	if err := domainping.ValidateTimeoutMs(*value); err != nil {
		return nil, invalidCheckField("timeoutMs", "must be greater than 0", *value)
	}

	normalized := *value
	return &normalized, nil
}

func normalizeOptionalIPFamily(value *string) (*domainnetwork.IPFamily, error) {
	ipFamily, err := domainnetwork.ParseOptionalIPFamily(value)
	if err != nil {
		return nil, invalidCheckField("ipFamily", `must be "inet" or "inet6"`, *value)
	}

	return ipFamily, nil
}

func normalizeOptionalLabelIDs(value *[]string) (bool, []string, error) {
	if value == nil {
		return false, nil, nil
	}

	labelIDs, err := normalizeLabelIDs(*value)
	if err != nil {
		return false, nil, err
	}

	return true, labelIDs, nil
}

func normalizeCheckType(value string) (domaincheck.Type, error) {
	raw := value
	value, err := normalizeRequiredStringField(value, "type")
	if err != nil {
		return "", err
	}
	if domaincheck.Type(value) != domaincheck.TypePing {
		return "", invalidCheckField("type", `must be "ping"`, raw)
	}

	return domaincheck.TypePing, nil
}

func normalizeSelector(selector map[string]any) (json.RawMessage, error) {
	if selector == nil {
		return defaultSelector, nil
	}

	raw, err := json.Marshal(selector)
	if err != nil {
		return nil, invalidCheckField("selector", "must be a valid selector", selector)
	}

	return raw, nil
}

func normalizePingConfig(packetCount, packetSizeBytes, timeoutMs *int32, ipFamilyValue *string) (domainping.Config, error) {
	config := domainping.DefaultConfig()

	normalizedPacketCount, err := normalizeOptionalPacketCount(packetCount)
	if err != nil {
		return domainping.Config{}, err
	}
	if normalizedPacketCount != nil {
		config.PacketCount = *normalizedPacketCount
	}

	normalizedPacketSizeBytes, err := normalizeOptionalPacketSizeBytes(packetSizeBytes)
	if err != nil {
		return domainping.Config{}, err
	}
	if normalizedPacketSizeBytes != nil {
		config.PacketSizeBytes = *normalizedPacketSizeBytes
	}

	normalizedTimeoutMs, err := normalizeOptionalTimeoutMs(timeoutMs)
	if err != nil {
		return domainping.Config{}, err
	}
	if normalizedTimeoutMs != nil {
		config.TimeoutMs = *normalizedTimeoutMs
	}

	ipFamily, err := normalizeOptionalIPFamily(ipFamilyValue)
	if err != nil {
		return domainping.Config{}, err
	}
	config.IPFamily = ipFamily

	return config, nil
}

func normalizeLabelIDs(values []string) ([]string, error) {
	labelIDs, err := normalize.CanonicalUUIDSet(values, ErrInvalidInput)
	if err != nil {
		return nil, invalidCheckField("labelIds", "must contain valid UUIDs", values)
	}

	return labelIDs, nil
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

func chooseRawMessage(current, next json.RawMessage) json.RawMessage {
	if next == nil {
		return current
	}

	return next
}

func chooseOptionalString(current, next *string) *string {
	if next == nil {
		return current
	}

	return next
}

func chooseInt32(current int32, next *int32) int32 {
	if next == nil {
		return current
	}

	return *next
}

func chooseIPFamily(current, next *domainnetwork.IPFamily) *domainnetwork.IPFamily {
	if next == nil {
		return current
	}

	return next
}
