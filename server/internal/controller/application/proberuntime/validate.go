package proberuntime

import (
	"fmt"
	"net/netip"
	"strings"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
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
	checkID string
	type_   domaincheck.Type
	ping    []domainping.ResultStorageInput
	index   int
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
	if len(input.Results) > MaxResultGroupBatchSize {
		return normalizedSubmitResultsInput{}, invalidRuntimeField(fieldResults, fmt.Sprintf("must include at most %d result groups", MaxResultGroupBatchSize), len(input.Results))
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

		groupKey := normalized.checkID + "\x00" + string(normalized.type_)
		if _, ok := seenGroups[groupKey]; ok {
			return normalizedSubmitResultsInput{}, invalidRuntimeField(resultGroupField(i, "checkId"), "duplicate result group", group.CheckID)
		}
		seenGroups[groupKey] = struct{}{}

		for j, result := range normalized.ping {
			resultKey := normalized.checkID + "\x00" + string(normalized.type_) + "\x00" + result.StartedAt.Format(timeKeyLayout)
			if _, ok := seenResults[resultKey]; ok {
				return normalizedSubmitResultsInput{}, invalidRuntimeField(resultGroupField(i, fmt.Sprintf("ping[%d].startedAt", j)), "duplicate result startedAt for check", result.StartedAt)
			}
			seenResults[resultKey] = struct{}{}
		}

		if _, ok := seenChecks[normalized.checkID]; !ok {
			checkIDs = append(checkIDs, normalized.checkID)
			seenChecks[normalized.checkID] = struct{}{}
		}
		accepted += len(normalized.ping)
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
	if checkType != domaincheck.TypePing {
		return normalizedResultGroup{}, invalidRuntimeField(resultGroupField(index, "type"), "unsupported result type", input.Type)
	}
	if len(input.Ping) == 0 {
		return normalizedResultGroup{}, invalidRuntimeField(resultGroupField(index, "ping"), "must include at least one ping result", input.Ping)
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
		checkID: checkID,
		type_:   checkType,
		ping:    pingResults,
		index:   index,
	}, nil
}

func normalizePingResult(input PingResultInput, fieldPrefix string) (domainping.ResultStorageInput, error) {
	startedAt, err := domainping.VNResultTime(input.StartedAt)
	if err != nil {
		return domainping.ResultStorageInput{}, invalidRuntimeField(resultField(fieldPrefix, "startedAt"), err.Error(), input.StartedAt)
	}
	finishedAt, err := domainping.VNResultTime(input.FinishedAt)
	if err != nil {
		return domainping.ResultStorageInput{}, invalidRuntimeField(resultField(fieldPrefix, "finishedAt"), err.Error(), input.FinishedAt)
	}
	if finishedAt.Before(startedAt) {
		return domainping.ResultStorageInput{}, invalidRuntimeField(resultField(fieldPrefix, "finishedAt"), "must be greater than or equal to startedAt", input.FinishedAt)
	}
	durationMs, err := domainping.VNResultDurationMs(input.DurationMs)
	if err != nil {
		return domainping.ResultStorageInput{}, invalidRuntimeField(resultField(fieldPrefix, "durationMs"), err.Error(), input.DurationMs)
	}
	status, err := domainping.VNResultStatus(domainping.Status(input.Status))
	if err != nil {
		return domainping.ResultStorageInput{}, invalidRuntimeField(resultField(fieldPrefix, "status"), err.Error(), input.Status)
	}
	sentCount, err := domainping.VNResultSentCount(input.SentCount)
	if err != nil {
		return domainping.ResultStorageInput{}, invalidRuntimeField(resultField(fieldPrefix, "sentCount"), err.Error(), input.SentCount)
	}
	receivedCount, err := domainping.VNResultReceivedCount(input.ReceivedCount, sentCount)
	if err != nil {
		return domainping.ResultStorageInput{}, invalidRuntimeField(resultField(fieldPrefix, "receivedCount"), err.Error(), input.ReceivedCount)
	}
	lossPercent, err := domainping.VNResultLossPercent(input.LossPercent)
	if err != nil {
		return domainping.ResultStorageInput{}, invalidRuntimeField(resultField(fieldPrefix, "lossPercent"), err.Error(), input.LossPercent)
	}

	rttMin, err := normalizeOptionalResultRTT(input.RttMinMs, resultField(fieldPrefix, "rttMinMs"))
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	rttAvg, err := normalizeOptionalResultRTT(input.RttAvgMs, resultField(fieldPrefix, "rttAvgMs"))
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	rttMedian, err := normalizeOptionalResultRTT(input.RttMedianMs, resultField(fieldPrefix, "rttMedianMs"))
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	rttMax, err := normalizeOptionalResultRTT(input.RttMaxMs, resultField(fieldPrefix, "rttMaxMs"))
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	rttStddev, err := normalizeOptionalResultRTT(input.RttStddevMs, resultField(fieldPrefix, "rttStddevMs"))
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	if err := validateRTTOrder(rttMin, rttAvg, rttMax, fieldPrefix); err != nil {
		return domainping.ResultStorageInput{}, err
	}
	rttSamples, err := domainping.VNResultRTTSamples(input.RttSamplesMs)
	if err != nil {
		return domainping.ResultStorageInput{}, invalidRuntimeField(resultField(fieldPrefix, "rttSamplesMs"), err.Error(), input.RttSamplesMs)
	}

	ipFamily, err := domainnetwork.ParseOptionalIPFamily(input.IPFamily)
	if err != nil {
		return domainping.ResultStorageInput{}, invalidRuntimeField(resultField(fieldPrefix, "ipFamily"), `must be "inet" or "inet6"`, input.IPFamily)
	}
	raw, err := domainping.VNResultRaw(input.Raw)
	if err != nil {
		return domainping.ResultStorageInput{}, invalidRuntimeField(resultField(fieldPrefix, "raw"), err.Error(), input.Raw)
	}
	errorCode, err := normalizeOptionalResultText(input.ErrorCode, resultField(fieldPrefix, "errorCode"))
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	errorMessage, err := normalizeOptionalResultText(input.ErrorMessage, resultField(fieldPrefix, "errorMessage"))
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}

	return domainping.ResultStorageInput{
		StartedAt:     startedAt,
		FinishedAt:    finishedAt,
		DurationMs:    durationMs,
		Status:        status,
		SentCount:     sentCount,
		ReceivedCount: receivedCount,
		LossPercent:   lossPercent,
		RttMinMs:      rttMin,
		RttAvgMs:      rttAvg,
		RttMedianMs:   rttMedian,
		RttMaxMs:      rttMax,
		RttStddevMs:   rttStddev,
		RttSamplesMs:  rttSamples,
		ResolvedIP:    cloneAddr(input.ResolvedIP),
		IPFamily:      ipFamily,
		Raw:           raw,
		ErrorCode:     errorCode,
		ErrorMessage:  errorMessage,
	}, nil
}

func normalizeOptionalResultRTT(input *float64, field string) (*float64, error) {
	value, err := domainping.VNResultOptionalRTT(input)
	if err != nil {
		return nil, invalidRuntimeField(field, err.Error(), input)
	}
	return value, nil
}

func normalizeOptionalResultText(input *string, field string) (*string, error) {
	value, err := domainping.VNResultOptionalText(input)
	if err != nil {
		return nil, invalidRuntimeField(field, err.Error(), input)
	}
	return value, nil
}

func validateRTTOrder(minValue, avgValue, maxValue *float64, fieldPrefix string) error {
	if minValue != nil && maxValue != nil && *minValue > *maxValue {
		return invalidRuntimeField(resultField(fieldPrefix, "rttMinMs"), "must be less than or equal to rttMaxMs", minValue)
	}
	if minValue != nil && avgValue != nil && *minValue > *avgValue {
		return invalidRuntimeField(resultField(fieldPrefix, "rttMinMs"), "must be less than or equal to rttAvgMs", minValue)
	}
	if avgValue != nil && maxValue != nil && *avgValue > *maxValue {
		return invalidRuntimeField(resultField(fieldPrefix, "rttAvgMs"), "must be less than or equal to rttMaxMs", avgValue)
	}

	return nil
}

func resultGroupField(index int, field string) string {
	if field == "" {
		return fmt.Sprintf("results[%d]", index)
	}

	return fmt.Sprintf("results[%d].%s", index, field)
}

func resultField(prefix, field string) string {
	return prefix + "." + field
}

func cloneAddr(value *netip.Addr) *netip.Addr {
	if value == nil {
		return nil
	}

	copied := *value
	return &copied
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
