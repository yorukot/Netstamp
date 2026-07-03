package check

import (
	"context"

	apptx "github.com/yorukot/netstamp/internal/controller/application/tx"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type Service struct {
	repo                Repository
	projectAccess       ProjectAccess
	labelAccess         LabelAccess
	assignmentRefresher AssignmentRefresher
	tx                  apptx.Transactor
	events              EventRecorder
}

func NewService(repo Repository, projectAccess ProjectAccess, labelAccess LabelAccess, assignmentRefresher AssignmentRefresher, events EventRecorder, transactors ...apptx.Transactor) *Service {
	tx := apptx.Transactor(apptx.NoopTransactor{})
	if len(transactors) > 0 && transactors[0] != nil {
		tx = transactors[0]
	}

	return &Service{
		repo:                repo,
		projectAccess:       projectAccess,
		labelAccess:         labelAccess,
		assignmentRefresher: assignmentRefresher,
		tx:                  tx,
		events:              events,
	}
}

func (s *Service) ListChecks(ctx context.Context, input ListChecksInput) ([]domaincheck.Check, error) {
	ctx, flow := s.startCheckFlow(ctx, "check.list", CheckActionList, input.CurrentUserID)
	defer flow.end()

	input, err := normalizeListChecksInput(input)
	if err != nil {
		return nil, flow.businessFailure(CheckEventListFailure, CheckReasonInvalidInput, err)
	}
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

	input, err := normalizeTargetCheckInput(input)
	if err != nil {
		return domaincheck.Check{}, flow.businessFailure(CheckEventGetFailure, CheckReasonInvalidInput, err)
	}
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

	normalized, err := normalizeCreateCheckInput(input)
	if err != nil {
		return domaincheck.Check{}, flow.businessFailure(CheckEventCreateFailure, CheckReasonInvalidInput, err)
	}
	flow.setProjectRef(normalized.projectRef)

	project, err := s.loadProject(ctx, flow, normalized.projectRef, input.CurrentUserID, CheckEventCreateFailure)
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

	checkInput := domaincheck.Check{
		ProjectID:        project.ID,
		Name:             normalized.name,
		Type:             normalized.checkType,
		Target:           normalized.target,
		Selector:         normalized.selector,
		Description:      normalized.description,
		IntervalSeconds:  normalized.intervalSeconds,
		PingConfig:       normalized.pingConfig,
		TCPConfig:        normalized.tcpConfig,
		TracerouteConfig: normalized.tracerouteConfig,
	}
	var check domaincheck.Check
	writeStage := "create"
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		created, err := s.repo.CreateCheck(ctx, checkInput, normalized.labelIDs)
		if err != nil {
			return err
		}
		created.Labels = labels
		flow.setCheckID(created.ID)
		writeStage = "assignment"
		if err := s.assignmentRefresher.RefreshProbeCheckAssignmentsForCheck(ctx, project.ID, created.ID); err != nil {
			return err
		}
		check = created
		return nil
	})
	if err != nil {
		if writeStage == "create" {
			return domaincheck.Check{}, flow.writeFailure(CheckEventCreateFailure, CheckReasonCheckCreateFailed, err)
		}
		return domaincheck.Check{}, flow.assignmentRefreshFailure(CheckEventCreateFailure, err)
	}
	flow.success(CheckEventCreateSuccess)

	return check, nil
}

func (s *Service) UpdateCheck(ctx context.Context, input UpdateCheckInput) (domaincheck.Check, error) {
	ctx, flow := s.startCheckFlow(ctx, "check.update", CheckActionUpdate, input.CurrentUserID)
	defer flow.end()

	normalized, err := normalizeUpdateCheckInput(input)
	if err != nil {
		return domaincheck.Check{}, flow.businessFailure(CheckEventUpdateFailure, CheckReasonInvalidInput, err)
	}
	flow.setProjectRef(normalized.projectRef)
	flow.setCheckID(normalized.checkID)

	project, err := s.loadProject(ctx, flow, normalized.projectRef, input.CurrentUserID, CheckEventUpdateFailure)
	if err != nil {
		return domaincheck.Check{}, err
	}
	err = s.requireAction(ctx, flow, project.ID, input.CurrentUserID, CheckEventUpdateFailure)
	if err != nil {
		return domaincheck.Check{}, err
	}

	current, err := s.repo.GetCheck(ctx, project.ID, normalized.checkID)
	if err != nil {
		return domaincheck.Check{}, flow.checkLookupFailure(CheckEventUpdateFailure, err)
	}

	labelIDs, resolvedLabels, err := s.resolveUpdateLabels(ctx, project.ID, current.Labels, normalized)
	if err != nil {
		return domaincheck.Check{}, flow.labelLookupFailure(CheckEventUpdateFailure, err)
	}

	updated, err := buildUpdatedCheck(project.ID, current, normalized)
	if err != nil {
		return domaincheck.Check{}, flow.businessFailure(CheckEventUpdateFailure, CheckReasonInvalidInput, err)
	}

	var check domaincheck.Check
	writeStage := "update"
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		check, err = s.repo.UpdateCheck(ctx, updated, normalized.replaceLabels, labelIDs)
		if err != nil {
			return err
		}
		if normalized.replaceLabels {
			check.Labels = resolvedLabels
		}
		writeStage = "assignment"
		if err := s.assignmentRefresher.RefreshProbeCheckAssignmentsForCheck(ctx, project.ID, check.ID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if writeStage == "update" {
			return domaincheck.Check{}, flow.writeFailure(CheckEventUpdateFailure, CheckReasonCheckUpdateFailed, err)
		}
		return domaincheck.Check{}, flow.assignmentRefreshFailure(CheckEventUpdateFailure, err)
	}
	flow.success(CheckEventUpdateSuccess)

	return check, nil
}

func (s *Service) resolveUpdateLabels(ctx context.Context, projectID string, currentLabels []domainlabel.Label, normalized normalizedUpdateCheckInput) ([]string, []domainlabel.Label, error) {
	if !normalized.replaceLabels {
		return nil, currentLabels, nil
	}

	labels, err := s.labelAccess.GetActiveLabelsByIDsForProject(ctx, projectID, normalized.labelIDs)
	if err != nil {
		return nil, nil, err
	}

	return normalized.labelIDs, labels, nil
}

func buildUpdatedCheck(projectID string, current domaincheck.Check, normalized normalizedUpdateCheckInput) (domaincheck.Check, error) {
	selector := current.Selector
	if normalized.selector != nil {
		selector = normalized.selector
	}
	updatedSelector, err := canonicalizeSelector(selector)
	if err != nil {
		return domaincheck.Check{}, err
	}

	checkType, err := mergedCheckType(current.Type, normalized.checkType)
	if err != nil {
		return domaincheck.Check{}, err
	}

	pingConfig, tcpConfig, tracerouteConfig, err := mergedTypeConfigs(current, normalized)
	if err != nil {
		return domaincheck.Check{}, err
	}

	return domaincheck.Check{
		ProjectID:        projectID,
		ID:               normalized.checkID,
		Name:             mergedString(current.Name, normalized.name),
		Type:             checkType,
		Target:           mergedString(current.Target, normalized.target),
		Selector:         updatedSelector,
		Description:      mergedDescription(current.Description, normalized.description),
		IntervalSeconds:  mergedInt32(current.IntervalSeconds, normalized.intervalSeconds),
		PingConfig:       pingConfig,
		TCPConfig:        tcpConfig,
		TracerouteConfig: tracerouteConfig,
	}, nil
}

func mergedPingConfig(current *domainping.Config, normalized normalizedUpdateCheckInput) *domainping.Config {
	config := domainping.DefaultConfig()
	if current != nil {
		config = *current
	}
	if normalized.pingConfig.packetCount != nil {
		config.PacketCount = *normalized.pingConfig.packetCount
	}
	if normalized.pingConfig.packetSizeBytes != nil {
		config.PacketSizeBytes = *normalized.pingConfig.packetSizeBytes
	}
	if normalized.pingConfig.timeoutMs != nil {
		config.TimeoutMs = *normalized.pingConfig.timeoutMs
	}
	if normalized.pingConfig.ipFamily != nil {
		config.IPFamily = normalized.pingConfig.ipFamily
	}
	return &config
}

func mergedTCPConfig(current *domaintcp.Config, normalized normalizedUpdateCheckInput) *domaintcp.Config {
	config := domaintcp.DefaultConfig()
	if current != nil {
		config = *current
	}
	if normalized.tcpConfig.port != nil {
		config.Port = *normalized.tcpConfig.port
	}
	if normalized.tcpConfig.timeoutMs != nil {
		config.TimeoutMs = *normalized.tcpConfig.timeoutMs
	}
	if normalized.tcpConfig.ipFamily != nil {
		config.IPFamily = normalized.tcpConfig.ipFamily
	}
	return &config
}

func mergedTracerouteConfig(current *domaintraceroute.Config, normalized normalizedUpdateCheckInput) *domaintraceroute.Config {
	config := domaintraceroute.DefaultConfig()
	if current != nil {
		config = *current
	}
	if normalized.tracerouteConfig.protocol != nil {
		config.Protocol = *normalized.tracerouteConfig.protocol
	}
	if normalized.tracerouteConfig.maxHops != nil {
		config.MaxHops = *normalized.tracerouteConfig.maxHops
	}
	if normalized.tracerouteConfig.timeoutMs != nil {
		config.TimeoutMs = *normalized.tracerouteConfig.timeoutMs
	}
	if normalized.tracerouteConfig.queriesPerHop != nil {
		config.QueriesPerHop = *normalized.tracerouteConfig.queriesPerHop
	}
	if normalized.tracerouteConfig.packetSizeBytes != nil {
		config.PacketSizeBytes = *normalized.tracerouteConfig.packetSizeBytes
	}
	if normalized.tracerouteConfig.port != nil {
		config.Port = *normalized.tracerouteConfig.port
	}
	if normalized.tracerouteConfig.ipFamily != nil {
		config.IPFamily = normalized.tracerouteConfig.ipFamily
	}
	return &config
}

func mergedTypeConfigs(current domaincheck.Check, normalized normalizedUpdateCheckInput) (*domainping.Config, *domaintcp.Config, *domaintraceroute.Config, error) {
	if normalized.pingConfig.hasChanges() && current.Type != domaincheck.TypePing {
		return nil, nil, nil, invalidCheckField("pingConfig", "must be omitted for non-ping checks", nil)
	}
	if normalized.tcpConfig.hasChanges() && current.Type != domaincheck.TypeTCP {
		return nil, nil, nil, invalidCheckField("tcpConfig", "must be omitted for non-tcp checks", nil)
	}
	if normalized.tracerouteConfig.hasChanges() && current.Type != domaincheck.TypeTraceroute {
		return nil, nil, nil, invalidCheckField("tracerouteConfig", "must be omitted for non-traceroute checks", nil)
	}

	switch current.Type {
	case domaincheck.TypePing:
		return mergedPingConfig(current.PingConfig, normalized), nil, nil, nil
	case domaincheck.TypeTCP:
		return nil, mergedTCPConfig(current.TCPConfig, normalized), nil, nil
	case domaincheck.TypeTraceroute:
		return nil, nil, mergedTracerouteConfig(current.TracerouteConfig, normalized), nil
	default:
		return nil, nil, nil, domaincheck.ErrInvalidInput
	}
}

func mergedString(current string, next *string) string {
	if next == nil {
		return current
	}
	return *next
}

func mergedCheckType(current domaincheck.Type, next *domaincheck.Type) (domaincheck.Type, error) {
	if next == nil {
		return current, nil
	}
	if *next != current {
		return "", invalidCheckField("type", "cannot be changed after check creation", string(*next))
	}
	return *next, nil
}

func mergedDescription(current, next *string) *string {
	if next == nil {
		return current
	}
	return next
}

func mergedInt32(current int32, next *int32) int32 {
	if next == nil {
		return current
	}
	return *next
}

func (s *Service) DeleteCheck(ctx context.Context, input GetCheckInput) error {
	ctx, flow := s.startCheckFlow(ctx, "check.delete", CheckActionDelete, input.CurrentUserID)
	defer flow.end()

	input, err := normalizeTargetCheckInput(input)
	if err != nil {
		return flow.businessFailure(CheckEventDeleteFailure, CheckReasonInvalidInput, err)
	}
	flow.setProjectRef(input.ProjectRef)
	flow.setCheckID(input.CheckID)

	project, err := s.loadProject(ctx, flow, input.ProjectRef, input.CurrentUserID, CheckEventDeleteFailure)
	if err != nil {
		return err
	}
	if err := s.requireAction(ctx, flow, project.ID, input.CurrentUserID, CheckEventDeleteFailure); err != nil {
		return err
	}

	deleteStage := "delete"
	if err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.repo.SoftDeleteCheck(ctx, project.ID, input.CheckID); err != nil {
			return err
		}
		deleteStage = "assignment"
		return s.assignmentRefresher.DeleteProbeCheckAssignmentsForCheck(ctx, project.ID, input.CheckID)
	}); err != nil {
		if deleteStage == "delete" {
			return flow.writeFailure(CheckEventDeleteFailure, CheckReasonCheckDeleteFailed, err)
		}
		return flow.assignmentDeleteFailure(CheckEventDeleteFailure, err)
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
