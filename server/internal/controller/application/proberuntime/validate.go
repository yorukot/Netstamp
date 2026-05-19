package proberuntime

import (
	"fmt"
	"net/netip"
	"strings"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

const (
	fieldProbeID      = "probeId"
	fieldCredential   = "credential"
	fieldAgentVersion = "agentVersion"
	fieldPublicV4     = "publicV4"
	fieldPublicV6     = "publicV6"
	fieldAS           = "as"
	fieldAddrs        = "addrs"
	fieldResults      = "results"
)

type normalizedRuntimeAuthInput struct {
	probeID    string
	credential string
}

type normalizedSubmitResultsInput struct {
	groups   []normalizedResultGroup
	checkIDs []string
	accepted int
}

type normalizedResultGroup struct {
	checkID    string
	checkType  domaincheck.Type
	ping       []domainping.ResultStorageInput
	traceroute []domaintraceroute.ResultStorageInput
	index      int
}

func normalizeRuntimeAuthInput(input RuntimeAuthInput) (normalizedRuntimeAuthInput, error) {
	probeID, err := domainprobe.VNProbeID(input.ProbeID)
	if err != nil {
		return normalizedRuntimeAuthInput{}, invalidRuntimeField(fieldProbeID, "must be a valid UUID", input.ProbeID)
	}
	credential := strings.TrimSpace(input.Credential)
	if credential == "" {
		return normalizedRuntimeAuthInput{}, domainprobe.ErrInvalidCredential
	}

	return normalizedRuntimeAuthInput{probeID: probeID, credential: credential}, nil
}

func normalizeSubmitResults(input SubmitResultsInput) (normalizedSubmitResultsInput, error) {
	if len(input.Results) == 0 {
		return normalizedSubmitResultsInput{}, invalidRuntimeField(fieldResults, "must include at least one result group", input.Results)
	}

	var validation appvalidation.Collector
	groups := make([]normalizedResultGroup, 0, len(input.Results))
	checkIDs := make([]string, 0, len(input.Results))
	seenGroups := map[string]struct{}{}
	seenChecks := map[string]struct{}{}
	seenResults := map[string]struct{}{}
	accepted := 0
	for i, group := range input.Results {
		normalized, err := normalizeResultGroup(group, i)
		if err != nil {
			if !validation.AddValidation(err) {
				return normalizedSubmitResultsInput{}, err
			}
			continue
		}

		groupKey := normalized.checkID + "\x00" + string(normalized.checkType)
		if _, ok := seenGroups[groupKey]; ok {
			validation.Add(resultGroupField(i, "checkId"), "duplicate result group", group.CheckID)
			continue
		}
		seenGroups[groupKey] = struct{}{}

		for j, result := range normalized.ping {
			resultKey := normalized.checkID + "\x00" + string(normalized.checkType) + "\x00" + result.StartedAt.Format(timeKeyLayout)
			if _, ok := seenResults[resultKey]; ok {
				validation.Add(resultGroupField(i, fmt.Sprintf("ping[%d].startedAt", j)), "duplicate result startedAt for check", result.StartedAt)
				continue
			}
			seenResults[resultKey] = struct{}{}
		}
		for j, result := range normalized.traceroute {
			resultKey := normalized.checkID + "\x00" + string(normalized.checkType) + "\x00" + result.StartedAt.Format(timeKeyLayout)
			if _, ok := seenResults[resultKey]; ok {
				validation.Add(resultGroupField(i, fmt.Sprintf("traceroute[%d].startedAt", j)), "duplicate result startedAt for check", result.StartedAt)
				continue
			}
			seenResults[resultKey] = struct{}{}
		}

		if _, ok := seenChecks[normalized.checkID]; !ok {
			checkIDs = append(checkIDs, normalized.checkID)
			seenChecks[normalized.checkID] = struct{}{}
		}
		accepted += len(normalized.ping) + len(normalized.traceroute)
		groups = append(groups, normalized)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return normalizedSubmitResultsInput{}, err
	}

	return normalizedSubmitResultsInput{groups: groups, checkIDs: checkIDs, accepted: accepted}, nil
}

const timeKeyLayout = "2006-01-02T15:04:05.999999999Z07:00"

func normalizeResultGroup(input RuntimeResultGroupInput, index int) (normalizedResultGroup, error) {
	var validation appvalidation.Collector

	checkID, err := domaincheck.VNCheckID(input.CheckID)
	if err != nil {
		validation.AddError(resultGroupField(index, "checkId"), err, input.CheckID)
	}

	checkType := domaincheck.Type(strings.TrimSpace(input.Type))
	if checkType != domaincheck.TypePing && checkType != domaincheck.TypeTraceroute {
		validation.Add(resultGroupField(index, "type"), "unsupported result type", input.Type)
	}
	if validation.HasErrors() && checkType != domaincheck.TypePing && checkType != domaincheck.TypeTraceroute {
		return normalizedResultGroup{}, validation.Err(ErrInvalidInput)
	}

	switch checkType {
	case domaincheck.TypePing:
		validateResultGroupShape(&validation, index, "ping", input.Ping, "traceroute", input.Traceroute)
		pingResults := normalizeRuntimeResults(input.Ping, index, "ping", normalizePingResult, &validation)
		if err := validation.Err(ErrInvalidInput); err != nil {
			return normalizedResultGroup{}, err
		}

		return normalizedResultGroup{
			checkID:   checkID,
			checkType: checkType,
			ping:      pingResults,
			index:     index,
		}, nil
	case domaincheck.TypeTraceroute:
		validateResultGroupShape(&validation, index, "traceroute", input.Traceroute, "ping", input.Ping)
		tracerouteResults := normalizeRuntimeResults(input.Traceroute, index, "traceroute", normalizeTracerouteResult, &validation)
		if err := validation.Err(ErrInvalidInput); err != nil {
			return normalizedResultGroup{}, err
		}

		return normalizedResultGroup{
			checkID:    checkID,
			checkType:  checkType,
			traceroute: tracerouteResults,
			index:      index,
		}, nil
	default:
		return normalizedResultGroup{}, invalidRuntimeField(resultGroupField(index, "type"), "unsupported result type", input.Type)
	}
}

func validateResultGroupShape[T, O any](validation *appvalidation.Collector, index int, requiredField string, required []T, omittedField string, omitted []O) {
	if len(required) == 0 {
		validation.Add(resultGroupField(index, requiredField), "must include at least one "+requiredField+" result", required)
	}
	if len(omitted) > 0 {
		validation.Add(resultGroupField(index, omittedField), "must be omitted for "+requiredField+" results", omitted)
	}
}

func normalizeRuntimeResults[T, R any](inputs []T, groupIndex int, field string, normalize func(T, string) (R, error), validation *appvalidation.Collector) []R {
	outputs := make([]R, 0, len(inputs))
	for i, input := range inputs {
		normalized, err := normalize(input, resultGroupField(groupIndex, fmt.Sprintf("%s[%d]", field, i)))
		if err != nil {
			if !validation.AddValidation(err) {
				validation.Add(resultGroupField(groupIndex, fmt.Sprintf("%s[%d]", field, i)), err.Error(), input)
			}
			continue
		}
		outputs = append(outputs, normalized)
	}
	return outputs
}

func resultGroupField(index int, field string) string {
	if field == "" {
		return fmt.Sprintf("results[%d]", index)
	}

	return fmt.Sprintf("results[%d].%s", index, field)
}

func normalizeRuntimeStatus(input RuntimeStatusInput, probeID string) (domainprobe.Status, error) {
	var validation appvalidation.Collector

	agentVersion, err := domainprobe.VNProbeOptionalAgentVersion(input.AgentVersion)
	if err != nil {
		validation.AddError(fieldAgentVersion, err, input.AgentVersion)
	}
	publicV4, err := domainprobe.VNProbePublicV4(input.PublicV4)
	if err != nil {
		validation.AddError(fieldPublicV4, err, input.PublicV4)
	}
	publicV6, err := domainprobe.VNProbePublicV6(input.PublicV6)
	if err != nil {
		validation.AddError(fieldPublicV6, err, input.PublicV6)
	}
	as, err := domainprobe.VNProbeOptionalAS(input.AS)
	if err != nil {
		validation.AddError(fieldAS, err, input.AS)
	}
	addrs, err := domainprobe.VNProbeAddrs(input.Addrs)
	if err != nil {
		validation.AddError(fieldAddrs, err, input.Addrs)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return domainprobe.Status{}, err
	}

	return domainprobe.Status{
		ProbeID:      probeID,
		State:        domainprobe.StateOnline,
		AgentVersion: agentVersion,
		PublicV4:     publicV4,
		PublicV6:     publicV6,
		AS:           as,
		Addrs:        append([]netip.Addr(nil), addrs...),
	}, nil
}

func invalidRuntimeField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}
