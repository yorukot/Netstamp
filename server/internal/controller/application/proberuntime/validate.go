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

	groups := make([]normalizedResultGroup, 0, len(input.Results))
	checkIDs := make([]string, 0, len(input.Results))
	seenGroups := map[string]struct{}{}
	seenChecks := map[string]struct{}{}
	seenResults := map[string]struct{}{}
	accepted := 0
	for i, group := range input.Results {
		normalized, err := normalizeResultGroup(group, i)
		if err != nil {
			return normalizedSubmitResultsInput{}, err
		}

		groupKey := normalized.checkID + "\x00" + string(normalized.checkType)
		if _, ok := seenGroups[groupKey]; ok {
			return normalizedSubmitResultsInput{}, invalidRuntimeField(resultGroupField(i, "checkId"), "duplicate result group", group.CheckID)
		}
		seenGroups[groupKey] = struct{}{}

		for j, result := range normalized.ping {
			resultKey := normalized.checkID + "\x00" + string(normalized.checkType) + "\x00" + result.StartedAt.Format(timeKeyLayout)
			if _, ok := seenResults[resultKey]; ok {
				return normalizedSubmitResultsInput{}, invalidRuntimeField(resultGroupField(i, fmt.Sprintf("ping[%d].startedAt", j)), "duplicate result startedAt for check", result.StartedAt)
			}
			seenResults[resultKey] = struct{}{}
		}
		for j, result := range normalized.traceroute {
			resultKey := normalized.checkID + "\x00" + string(normalized.checkType) + "\x00" + result.StartedAt.Format(timeKeyLayout)
			if _, ok := seenResults[resultKey]; ok {
				return normalizedSubmitResultsInput{}, invalidRuntimeField(resultGroupField(i, fmt.Sprintf("traceroute[%d].startedAt", j)), "duplicate result startedAt for check", result.StartedAt)
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

	return normalizedSubmitResultsInput{groups: groups, checkIDs: checkIDs, accepted: accepted}, nil
}

const timeKeyLayout = "2006-01-02T15:04:05.999999999Z07:00"

func normalizeResultGroup(input RuntimeResultGroupInput, index int) (normalizedResultGroup, error) {
	checkID, err := domaincheck.VNCheckID(input.CheckID)
	if err != nil {
		return normalizedResultGroup{}, invalidRuntimeField(resultGroupField(index, "checkId"), err.Error(), input.CheckID)
	}

	checkType := domaincheck.Type(strings.TrimSpace(input.Type))
	if checkType != domaincheck.TypePing && checkType != domaincheck.TypeTraceroute {
		return normalizedResultGroup{}, invalidRuntimeField(resultGroupField(index, "type"), "unsupported result type", input.Type)
	}

	switch checkType {
	case domaincheck.TypePing:
		if len(input.Ping) == 0 {
			return normalizedResultGroup{}, invalidRuntimeField(resultGroupField(index, "ping"), "must include at least one ping result", input.Ping)
		}
		if len(input.Traceroute) > 0 {
			return normalizedResultGroup{}, invalidRuntimeField(resultGroupField(index, "traceroute"), "must be omitted for ping results", input.Traceroute)
		}
		pingResults := make([]domainping.ResultStorageInput, 0, len(input.Ping))
		for i, result := range input.Ping {
			normalized, err := normalizePingResult(result, resultGroupField(index, fmt.Sprintf("ping[%d]", i)))
			if err != nil {
				return normalizedResultGroup{}, err
			}
			pingResults = append(pingResults, normalized)
		}

		return normalizedResultGroup{
			checkID:   checkID,
			checkType: checkType,
			ping:      pingResults,
			index:     index,
		}, nil
	case domaincheck.TypeTraceroute:
		if len(input.Traceroute) == 0 {
			return normalizedResultGroup{}, invalidRuntimeField(resultGroupField(index, "traceroute"), "must include at least one traceroute result", input.Traceroute)
		}
		if len(input.Ping) > 0 {
			return normalizedResultGroup{}, invalidRuntimeField(resultGroupField(index, "ping"), "must be omitted for traceroute results", input.Ping)
		}
		tracerouteResults := make([]domaintraceroute.ResultStorageInput, 0, len(input.Traceroute))
		for i, result := range input.Traceroute {
			normalized, err := normalizeTracerouteResult(result, resultGroupField(index, fmt.Sprintf("traceroute[%d]", i)))
			if err != nil {
				return normalizedResultGroup{}, err
			}
			tracerouteResults = append(tracerouteResults, normalized)
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

func resultGroupField(index int, field string) string {
	if field == "" {
		return fmt.Sprintf("results[%d]", index)
	}

	return fmt.Sprintf("results[%d].%s", index, field)
}

func normalizeRuntimeStatus(input RuntimeStatusInput, probeID string) (domainprobe.Status, error) {
	agentVersion, err := domainprobe.VNProbeOptionalAgentVersion(input.AgentVersion)
	if err != nil {
		return domainprobe.Status{}, invalidRuntimeField(fieldAgentVersion, err.Error(), input.AgentVersion)
	}
	publicV4, err := domainprobe.VNProbePublicV4(input.PublicV4)
	if err != nil {
		return domainprobe.Status{}, invalidRuntimeField(fieldPublicV4, err.Error(), input.PublicV4)
	}
	publicV6, err := domainprobe.VNProbePublicV6(input.PublicV6)
	if err != nil {
		return domainprobe.Status{}, invalidRuntimeField(fieldPublicV6, err.Error(), input.PublicV6)
	}
	as, err := domainprobe.VNProbeOptionalAS(input.AS)
	if err != nil {
		return domainprobe.Status{}, invalidRuntimeField(fieldAS, err.Error(), input.AS)
	}
	addrs, err := domainprobe.VNProbeAddrs(input.Addrs)
	if err != nil {
		return domainprobe.Status{}, invalidRuntimeField(fieldAddrs, err.Error(), input.Addrs)
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
