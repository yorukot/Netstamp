package proberuntime

import (
	"encoding/json"
	"net/netip"
	"sort"
	"strconv"
	"strings"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

const (
	maxAgentVersionRunes       = 100
	maxResultErrorCodeRunes    = 100
	maxResultErrorMessageRunes = 500
)

type normalizedRuntimeAuthInput struct {
	probeID    string
	credential string
}

type normalizedResultGroup struct {
	checkID         string
	checkType       domaincheck.Type
	assignmentID    string
	checkVersion    string
	selectorVersion string
	pingResults     []PingResultInput
}

func normalizeRuntimeAuthInput(input RuntimeAuthInput) (normalizedRuntimeAuthInput, error) {
	probeID, err := appvalidation.CanonicalUUID(ErrInvalidInput, "probeId", input.ProbeID)
	if err != nil {
		return normalizedRuntimeAuthInput{}, err
	}
	credential, err := appvalidation.RequiredString(ErrInvalidCredential, "credential", input.Credential, 0)
	if err != nil {
		return normalizedRuntimeAuthInput{}, err
	}

	return normalizedRuntimeAuthInput{probeID: probeID, credential: credential}, nil
}

func normalizeRuntimeStatus(input RuntimeStatusInput, probeID string) (domainprobe.UpdateStatusInput, error) {
	agentVersion, err := appvalidation.OptionalString(ErrInvalidInput, "agentVersion", input.AgentVersion, maxAgentVersionRunes)
	if err != nil {
		return domainprobe.UpdateStatusInput{}, err
	}

	return domainprobe.UpdateStatusInput{
		ProbeID:      probeID,
		State:        domainprobe.StateOnline,
		AgentVersion: agentVersion,
		PublicV4:     input.PublicV4,
		PublicV6:     input.PublicV6,
		Addrs:        append([]netip.Addr(nil), input.Addrs...),
	}, nil
}

func validateResultBatch(input SubmitResultsInput) (int, []normalizedResultGroup, error) {
	resultCount := countResults(input.Groups)
	if resultCount == 0 || resultCount > MaxResultBatchSize {
		return resultCount, nil, invalidRuntimeField("", "must include between 1 and 500 results", nil)
	}

	groups := make([]normalizedResultGroup, 0, len(input.Groups))
	seenCheckIDs := make(map[string]struct{}, len(input.Groups))
	for index, group := range input.Groups {
		normalized, err := normalizeResultGroup(group)
		if err != nil {
			return resultCount, nil, err
		}
		if _, ok := seenCheckIDs[normalized.checkID]; ok {
			return resultCount, nil, invalidRuntimeField(resultGroupField(index, "checkId"), "must be unique", group.CheckID)
		}
		seenCheckIDs[normalized.checkID] = struct{}{}
		groups = append(groups, normalized)
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].checkID < groups[j].checkID
	})

	return resultCount, groups, nil
}

func countResults(groups []ResultGroupInput) int {
	count := 0
	for _, group := range groups {
		count += len(group.PingResults)
	}

	return count
}

func normalizeResultGroup(input ResultGroupInput) (normalizedResultGroup, error) {
	checkID, err := appvalidation.CanonicalUUID(ErrInvalidInput, "checks.checkId", input.CheckID)
	if err != nil {
		return normalizedResultGroup{}, err
	}
	checkType, err := normalizeResultType(input.Type, input.CheckID)
	if err != nil {
		return normalizedResultGroup{}, err
	}
	assignmentID, err := appvalidation.CanonicalUUID(ErrInvalidInput, resultGroupField(input.CheckID, "detail.assignmentId"), input.AssignmentID)
	if err != nil {
		return normalizedResultGroup{}, err
	}
	checkVersion, err := appvalidation.RequiredString(ErrInvalidInput, resultGroupField(input.CheckID, "detail.checkVersion"), input.CheckVersion, 0)
	if err != nil {
		return normalizedResultGroup{}, err
	}
	selectorVersion, err := appvalidation.RequiredString(ErrInvalidInput, resultGroupField(input.CheckID, "detail.selectorVersion"), input.SelectorVersion, 0)
	if err != nil {
		return normalizedResultGroup{}, err
	}
	if len(input.PingResults) == 0 {
		return normalizedResultGroup{}, invalidRuntimeField(resultGroupField(input.CheckID, "results"), "must include at least one result", input.PingResults)
	}

	return normalizedResultGroup{
		checkID:         checkID,
		checkType:       checkType,
		assignmentID:    assignmentID,
		checkVersion:    checkVersion,
		selectorVersion: selectorVersion,
		pingResults:     append([]PingResultInput(nil), input.PingResults...),
	}, nil
}

func normalizeResultType(value domaincheck.Type, checkID string) (domaincheck.Type, error) {
	field := resultGroupField(checkID, "type")
	switch value {
	case domaincheck.TypePing:
		return domaincheck.TypePing, nil
	case "":
		return "", invalidRuntimeField(field, `must be "ping"`, value)
	default:
		return "", appvalidation.New(ErrUnsupportedResult, field, "unsupported result type", value)
	}
}

func resultGroupField(checkID any, field string) string {
	switch value := checkID.(type) {
	case int:
		return "checks." + strconv.Itoa(value) + "." + field
	case string:
		if value != "" {
			return "checks." + value + "." + field
		}
	}

	return "checks." + field
}

func normalizePingResult(input PingResultInput, projectID, probeID, checkID string) (domainping.ResultStorageInput, error) {
	checkID, err := appvalidation.CanonicalUUID(ErrInvalidResult, "checks.checkId", checkID)
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	status, err := parsePingStatus(input.Status)
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	err = validatePingResultTiming(input)
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	err = validatePingResultCounts(input)
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	err = validatePingRTTs(input)
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	errorCode, err := appvalidation.OptionalString(ErrInvalidResult, "ping.errorCode", input.ErrorCode, maxResultErrorCodeRunes)
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	errorMessage, err := appvalidation.OptionalString(ErrInvalidResult, "ping.errorMessage", input.ErrorMessage, maxResultErrorMessageRunes)
	if err != nil {
		return domainping.ResultStorageInput{}, err
	}
	raw := input.Raw
	if len(raw) == 0 {
		raw = json.RawMessage(`{}`)
	}
	if !json.Valid(raw) {
		return domainping.ResultStorageInput{}, invalidResultField("ping.raw", "must be valid JSON", input.Raw)
	}

	return domainping.ResultStorageInput{
		ProjectID:     projectID,
		ProbeID:       probeID,
		CheckID:       checkID,
		StartedAt:     input.StartedAt,
		FinishedAt:    input.FinishedAt,
		DurationMs:    input.DurationMs,
		Status:        status,
		SentCount:     input.SentCount,
		ReceivedCount: input.ReceivedCount,
		LossPercent:   input.LossPercent,
		RttMinMs:      input.RttMinMs,
		RttAvgMs:      input.RttAvgMs,
		RttMedianMs:   input.RttMedianMs,
		RttMaxMs:      input.RttMaxMs,
		RttStddevMs:   input.RttStddevMs,
		RttSamplesMs:  append([]float64{}, input.RttSamplesMs...),
		ResolvedIP:    input.ResolvedIP,
		IPFamily:      input.IPFamily,
		Raw:           append(json.RawMessage(nil), raw...),
		ErrorCode:     errorCode,
		ErrorMessage:  errorMessage,
	}, nil
}

func parsePingStatus(value string) (domainping.Status, error) {
	switch domainping.Status(strings.TrimSpace(value)) {
	case domainping.StatusSuccessful:
		return domainping.StatusSuccessful, nil
	case domainping.StatusTimeout:
		return domainping.StatusTimeout, nil
	case domainping.StatusError:
		return domainping.StatusError, nil
	default:
		return "", invalidResultField("ping.status", `must be "successful", "timeout", or "error"`, value)
	}
}

func validatePingResultTiming(input PingResultInput) error {
	switch {
	case input.StartedAt.IsZero():
		return invalidResultField("ping.startedAt", "must be provided", input.StartedAt)
	case input.FinishedAt.IsZero():
		return invalidResultField("ping.finishedAt", "must be provided", input.FinishedAt)
	case input.FinishedAt.Before(input.StartedAt):
		return invalidResultField("ping.finishedAt", "must be after startedAt", input.FinishedAt)
	case input.DurationMs < 0:
		return invalidResultField("ping.durationMs", "must be greater than or equal to 0", input.DurationMs)
	default:
		return nil
	}
}

func validatePingResultCounts(input PingResultInput) error {
	switch {
	case input.SentCount < 0:
		return invalidResultField("ping.sentCount", "must be greater than or equal to 0", input.SentCount)
	case input.ReceivedCount < 0:
		return invalidResultField("ping.receivedCount", "must be greater than or equal to 0", input.ReceivedCount)
	case input.ReceivedCount > input.SentCount:
		return invalidResultField("ping.receivedCount", "must be less than or equal to sentCount", input.ReceivedCount)
	case input.LossPercent < 0 || input.LossPercent > 100:
		return invalidResultField("ping.lossPercent", "must be between 0 and 100", input.LossPercent)
	default:
		return nil
	}
}

func validatePingRTTs(input PingResultInput) error {
	if err := validateNonNegativeRTTMetric("ping.rttMinMs", input.RttMinMs); err != nil {
		return err
	}
	if err := validateNonNegativeRTTMetric("ping.rttAvgMs", input.RttAvgMs); err != nil {
		return err
	}
	if err := validateNonNegativeRTTMetric("ping.rttMedianMs", input.RttMedianMs); err != nil {
		return err
	}
	if err := validateNonNegativeRTTMetric("ping.rttMaxMs", input.RttMaxMs); err != nil {
		return err
	}
	if err := validateNonNegativeRTTMetric("ping.rttStddevMs", input.RttStddevMs); err != nil {
		return err
	}
	for _, sample := range input.RttSamplesMs {
		if sample < 0 {
			return invalidResultField("ping.rttSamplesMs", "must not contain negative values", input.RttSamplesMs)
		}
	}
	if greaterThan(input.RttMinMs, input.RttMaxMs) {
		return invalidResultField("ping.rttMinMs", "must be less than or equal to rttMaxMs", input.RttMinMs)
	}
	if greaterThan(input.RttMinMs, input.RttAvgMs) {
		return invalidResultField("ping.rttMinMs", "must be less than or equal to rttAvgMs", input.RttMinMs)
	}
	if greaterThan(input.RttAvgMs, input.RttMaxMs) {
		return invalidResultField("ping.rttAvgMs", "must be less than or equal to rttMaxMs", input.RttAvgMs)
	}
	if greaterThan(input.RttMinMs, input.RttMedianMs) {
		return invalidResultField("ping.rttMinMs", "must be less than or equal to rttMedianMs", input.RttMinMs)
	}
	if greaterThan(input.RttMedianMs, input.RttMaxMs) {
		return invalidResultField("ping.rttMedianMs", "must be less than or equal to rttMaxMs", input.RttMedianMs)
	}

	return nil
}

func validateNonNegativeRTTMetric(field string, value *float64) error {
	if value != nil && *value < 0 {
		return invalidResultField(field, "must be greater than or equal to 0", *value)
	}

	return nil
}

func greaterThan(left, right *float64) bool {
	return left != nil && right != nil && *left > *right
}

func invalidRuntimeField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}

func invalidResultField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidResult, field, message, value)
}
