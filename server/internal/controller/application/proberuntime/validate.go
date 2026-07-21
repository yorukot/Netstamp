package proberuntime

import (
	"fmt"
	"net/netip"
	"strings"
	"time"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
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
	fieldFamilies     = "families"
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

type normalizedIPFamilyCapabilitiesInput struct {
	capabilities domainprobe.IPFamilyCapabilities
	hasUpdate    bool
}

type normalizedResultGroup struct {
	checkID    string
	checkType  domaincheck.Type
	ping       []domainping.ResultStorageInput
	tcp        []domaintcp.ResultStorageInput
	traceroute []domaintraceroute.ResultStorageInput
	http       []domainhttp.ResultStorageInput
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

		validateUniqueGroupResults(normalized, i, seenResults, &validation)

		if _, ok := seenChecks[normalized.checkID]; !ok {
			checkIDs = append(checkIDs, normalized.checkID)
			seenChecks[normalized.checkID] = struct{}{}
		}
		accepted += len(normalized.ping) + len(normalized.tcp) + len(normalized.traceroute) + len(normalized.http)
		groups = append(groups, normalized)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return normalizedSubmitResultsInput{}, err
	}

	return normalizedSubmitResultsInput{groups: groups, checkIDs: checkIDs, accepted: accepted}, nil
}

func validateUniqueGroupResults(group normalizedResultGroup, groupIndex int, seen map[string]struct{}, validation *appvalidation.Collector) {
	for index, result := range group.ping {
		validateUniqueResultTime(group, groupIndex, "ping", index, result.StartedAt, seen, validation)
	}
	for index, result := range group.tcp {
		validateUniqueResultTime(group, groupIndex, "tcp", index, result.StartedAt, seen, validation)
	}
	for index, result := range group.traceroute {
		validateUniqueResultTime(group, groupIndex, "traceroute", index, result.StartedAt, seen, validation)
	}
	for index, result := range group.http {
		validateUniqueResultTime(group, groupIndex, "http", index, result.StartedAt, seen, validation)
	}
}

func validateUniqueResultTime(group normalizedResultGroup, groupIndex int, resultType string, resultIndex int, startedAt time.Time, seen map[string]struct{}, validation *appvalidation.Collector) {
	key := group.checkID + "\x00" + string(group.checkType) + "\x00" + startedAt.Format(timeKeyLayout)
	if _, exists := seen[key]; exists {
		validation.Add(resultGroupField(groupIndex, fmt.Sprintf("%s[%d].startedAt", resultType, resultIndex)), "duplicate result startedAt for check", startedAt)
		return
	}
	seen[key] = struct{}{}
}

const timeKeyLayout = "2006-01-02T15:04:05.999999999Z07:00"

func normalizeResultGroup(input RuntimeResultGroupInput, index int) (normalizedResultGroup, error) {
	var validation appvalidation.Collector

	checkID, err := domaincheck.VNCheckID(input.CheckID)
	if err != nil {
		validation.AddError(resultGroupField(index, "checkId"), err, input.CheckID)
	}

	checkType := domaincheck.Type(strings.TrimSpace(input.Type))
	if !isSupportedResultGroupType(checkType) {
		validation.Add(resultGroupField(index, "type"), "unsupported result type", input.Type)
	}
	if validation.HasErrors() && !isSupportedResultGroupType(checkType) {
		return normalizedResultGroup{}, validation.Err(ErrInvalidInput)
	}

	switch checkType {
	case domaincheck.TypePing:
		return normalizePingResultGroup(input, index, checkID, &validation)
	case domaincheck.TypeTCP:
		return normalizeTCPResultGroup(input, index, checkID, &validation)
	case domaincheck.TypeTraceroute:
		return normalizeTracerouteResultGroup(input, index, checkID, &validation)
	case domaincheck.TypeHTTP:
		return normalizeHTTPResultGroup(input, index, checkID, &validation)
	default:
		return normalizedResultGroup{}, invalidRuntimeField(resultGroupField(index, "type"), "unsupported result type", input.Type)
	}
}

func isSupportedResultGroupType(checkType domaincheck.Type) bool {
	return checkType == domaincheck.TypePing || checkType == domaincheck.TypeTCP || checkType == domaincheck.TypeTraceroute || checkType == domaincheck.TypeHTTP
}

func normalizePingResultGroup(input RuntimeResultGroupInput, index int, checkID string, validation *appvalidation.Collector) (normalizedResultGroup, error) {
	validateResultGroupShape(validation, index, "ping", input.Ping, map[string]int{
		"tcp":        len(input.TCP),
		"traceroute": len(input.Traceroute),
		"http":       len(input.HTTP),
	})
	pingResults := normalizeRuntimeResults(input.Ping, index, "ping", normalizePingResult, validation)
	if err := validation.Err(ErrInvalidInput); err != nil {
		return normalizedResultGroup{}, err
	}

	return normalizedResultGroup{
		checkID:   checkID,
		checkType: domaincheck.TypePing,
		ping:      pingResults,
		index:     index,
	}, nil
}

func normalizeTCPResultGroup(input RuntimeResultGroupInput, index int, checkID string, validation *appvalidation.Collector) (normalizedResultGroup, error) {
	validateResultGroupShape(validation, index, "tcp", input.TCP, map[string]int{
		"ping":       len(input.Ping),
		"traceroute": len(input.Traceroute),
		"http":       len(input.HTTP),
	})
	tcpResults := normalizeRuntimeResults(input.TCP, index, "tcp", normalizeTCPResult, validation)
	if err := validation.Err(ErrInvalidInput); err != nil {
		return normalizedResultGroup{}, err
	}

	return normalizedResultGroup{
		checkID:   checkID,
		checkType: domaincheck.TypeTCP,
		tcp:       tcpResults,
		index:     index,
	}, nil
}

func normalizeTracerouteResultGroup(input RuntimeResultGroupInput, index int, checkID string, validation *appvalidation.Collector) (normalizedResultGroup, error) {
	validateResultGroupShape(validation, index, "traceroute", input.Traceroute, map[string]int{
		"ping": len(input.Ping),
		"tcp":  len(input.TCP),
		"http": len(input.HTTP),
	})
	tracerouteResults := normalizeRuntimeResults(input.Traceroute, index, "traceroute", normalizeTracerouteResult, validation)
	if err := validation.Err(ErrInvalidInput); err != nil {
		return normalizedResultGroup{}, err
	}

	return normalizedResultGroup{
		checkID:    checkID,
		checkType:  domaincheck.TypeTraceroute,
		traceroute: tracerouteResults,
		index:      index,
	}, nil
}

func normalizeHTTPResultGroup(input RuntimeResultGroupInput, index int, checkID string, validation *appvalidation.Collector) (normalizedResultGroup, error) {
	validateResultGroupShape(validation, index, "http", input.HTTP, map[string]int{
		"ping": len(input.Ping), "tcp": len(input.TCP), "traceroute": len(input.Traceroute),
	})
	results := normalizeRuntimeResults(input.HTTP, index, "http", normalizeHTTPResult, validation)
	if err := validation.Err(ErrInvalidInput); err != nil {
		return normalizedResultGroup{}, err
	}
	return normalizedResultGroup{checkID: checkID, checkType: domaincheck.TypeHTTP, http: results, index: index}, nil
}

func validateResultGroupShape[T any](validation *appvalidation.Collector, index int, requiredField string, required []T, omitted map[string]int) {
	if len(required) == 0 {
		validation.Add(resultGroupField(index, requiredField), "must include at least one "+requiredField+" result", required)
	}
	for omittedField, omittedCount := range omitted {
		if omittedCount > 0 {
			validation.Add(resultGroupField(index, omittedField), "must be omitted for "+requiredField+" results", omittedCount)
		}
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
		Addrs:        append(make([]netip.Addr, 0, len(addrs)), addrs...),
	}, nil
}

func normalizeIPFamilyCapabilities(input IPFamilyCapabilitiesInput, probeID string) (normalizedIPFamilyCapabilitiesInput, error) {
	if !input.BodyPresent {
		capabilities, ok := inferredIPFamilyCapabilities(probeID, input.ObservedIP)
		if !ok {
			return normalizedIPFamilyCapabilitiesInput{}, nil
		}

		return normalizedIPFamilyCapabilitiesInput{capabilities: capabilities, hasUpdate: true}, nil
	}

	var validation appvalidation.Collector
	if len(input.Families) == 0 {
		validation.Add(fieldFamilies, "must contain at least one family", input.Families)
	}

	seen := make(map[domainnetwork.IPFamily]struct{}, len(input.Families))
	capabilities := domainprobe.IPFamilyCapabilities{
		ProbeID:  probeID,
		UpdateV4: true,
		UpdateV6: true,
	}
	for i, familyInput := range input.Families {
		family, err := domainnetwork.ParseIPFamily(familyInput)
		if err != nil {
			validation.Add(ipFamilyCapabilityField(i), `must be "inet" or "inet6"`, familyInput)
			continue
		}
		if _, ok := seen[family]; ok {
			validation.Add(ipFamilyCapabilityField(i), "must not be duplicated", familyInput)
			continue
		}
		seen[family] = struct{}{}

		if family == domainnetwork.IPFamilyInet {
			capabilities.PublicV4 = observedPublicIPForFamily(input.ObservedIP, family)
		} else {
			capabilities.PublicV6 = observedPublicIPForFamily(input.ObservedIP, family)
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return normalizedIPFamilyCapabilitiesInput{}, err
	}

	return normalizedIPFamilyCapabilitiesInput{capabilities: capabilities, hasUpdate: true}, nil
}

func inferredIPFamilyCapabilities(probeID string, observedIP *netip.Addr) (domainprobe.IPFamilyCapabilities, bool) {
	if observedIP == nil || !observedIP.IsValid() {
		return domainprobe.IPFamilyCapabilities{}, false
	}

	addr := observedIP.Unmap()
	if addr.Is4() {
		return domainprobe.IPFamilyCapabilities{
			ProbeID:  probeID,
			PublicV4: &addr,
			UpdateV4: true,
		}, true
	}

	return domainprobe.IPFamilyCapabilities{
		ProbeID:  probeID,
		PublicV6: &addr,
		UpdateV6: true,
	}, true
}

func observedPublicIPForFamily(observedIP *netip.Addr, family domainnetwork.IPFamily) *netip.Addr {
	if observedIP == nil || !observedIP.IsValid() {
		return nil
	}

	addr := observedIP.Unmap()
	switch family {
	case domainnetwork.IPFamilyInet:
		if !addr.Is4() {
			return nil
		}
	case domainnetwork.IPFamilyInet6:
		if !addr.Is6() {
			return nil
		}
	default:
		return nil
	}

	return &addr
}

func ipFamilyCapabilityField(index int) string {
	return fmt.Sprintf("%s[%d]", fieldFamilies, index)
}

func invalidRuntimeField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}
