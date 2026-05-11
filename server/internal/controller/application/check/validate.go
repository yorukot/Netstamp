package check

import (
	"encoding/json"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainselector "github.com/yorukot/netstamp/internal/domain/selector"
)

const (
	maxCheckNameRunes        = 100
	maxCheckTargetRunes      = 255
	maxCheckDescriptionRunes = 500
)

var defaultSelector = json.RawMessage(`{}`)

func normalizeCreateCheckInput(input CreateCheckInput) (normalizedCreateCheckInput, error) {
	name, err := appvalidation.RequiredString(ErrInvalidInput, "name", input.Name, maxCheckNameRunes)
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}
	checkType, err := normalizeCheckType(input.Type)
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}
	target, err := appvalidation.RequiredString(ErrInvalidInput, "target", input.Target, maxCheckTargetRunes)
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}
	selector, err := normalizeCreateSelector(input.Selector)
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}
	description, err := appvalidation.OptionalString(ErrInvalidInput, "description", input.Description, maxCheckDescriptionRunes)
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}
	intervalSeconds, err := appvalidation.PositiveInt32(ErrInvalidInput, "intervalSeconds", input.IntervalSeconds)
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}
	pingConfig, err := normalizePingConfig(input.PingConfig)
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
		intervalSeconds: intervalSeconds,
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

	name, err := appvalidation.OptionalString(ErrInvalidInput, "name", input.Name, maxCheckNameRunes)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	checkType, err := normalizeOptionalCheckType(input.Type)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	target, err := appvalidation.OptionalString(ErrInvalidInput, "target", input.Target, maxCheckTargetRunes)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	selector, err := normalizeOptionalSelector(input.Selector)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	description, err := appvalidation.OptionalString(ErrInvalidInput, "description", input.Description, maxCheckDescriptionRunes)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	intervalSeconds, err := appvalidation.OptionalPositiveInt32(ErrInvalidInput, "intervalSeconds", input.IntervalSeconds)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	packetCount, err := normalizeOptionalPacketCount(input.PingConfig.packetCount())
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	packetSizeBytes, err := normalizeOptionalPacketSizeBytes(input.PingConfig.packetSizeBytes())
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	timeoutMs, err := normalizeOptionalTimeoutMs(input.PingConfig.timeoutMs())
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	ipFamily, err := normalizeOptionalIPFamily(input.PingConfig.ipFamily())
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
		input.PingConfig.hasChanges() ||
		input.LabelIDs != nil
}

func (input *PingConfigInput) hasChanges() bool {
	return input.packetCount() != nil ||
		input.packetSizeBytes() != nil ||
		input.timeoutMs() != nil ||
		input.ipFamily() != nil
}

func (input *PingConfigInput) packetCount() *int32 {
	if input == nil {
		return nil
	}

	return input.PacketCount
}

func (input *PingConfigInput) packetSizeBytes() *int32 {
	if input == nil {
		return nil
	}

	return input.PacketSizeBytes
}

func (input *PingConfigInput) timeoutMs() *int32 {
	if input == nil {
		return nil
	}

	return input.TimeoutMs
}

func (input *PingConfigInput) ipFamily() *string {
	if input == nil {
		return nil
	}

	return input.IPFamily
}

func invalidCheckField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
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

func normalizeOptionalPacketCount(value *int32) (*int32, error) {
	return appvalidation.OptionalPositiveInt32(ErrInvalidInput, "packetCount", value)
}

func normalizeOptionalPacketSizeBytes(value *int32) (*int32, error) {
	return appvalidation.OptionalInt32Range(ErrInvalidInput, "packetSizeBytes", value, 0, domainping.MaxPacketSizeBytes)
}

func normalizeOptionalTimeoutMs(value *int32) (*int32, error) {
	return appvalidation.OptionalPositiveInt32(ErrInvalidInput, "timeoutMs", value)
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
	value, err := appvalidation.RequiredString(ErrInvalidInput, "type", value, 0)
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

func normalizePingConfig(input *PingConfigInput) (domainping.Config, error) {
	config := domainping.DefaultConfig()

	normalizedPacketCount, err := normalizeOptionalPacketCount(input.packetCount())
	if err != nil {
		return domainping.Config{}, err
	}
	if normalizedPacketCount != nil {
		config.PacketCount = *normalizedPacketCount
	}

	normalizedPacketSizeBytes, err := normalizeOptionalPacketSizeBytes(input.packetSizeBytes())
	if err != nil {
		return domainping.Config{}, err
	}
	if normalizedPacketSizeBytes != nil {
		config.PacketSizeBytes = *normalizedPacketSizeBytes
	}

	normalizedTimeoutMs, err := normalizeOptionalTimeoutMs(input.timeoutMs())
	if err != nil {
		return domainping.Config{}, err
	}
	if normalizedTimeoutMs != nil {
		config.TimeoutMs = *normalizedTimeoutMs
	}

	ipFamily, err := normalizeOptionalIPFamily(input.ipFamily())
	if err != nil {
		return domainping.Config{}, err
	}
	config.IPFamily = ipFamily

	return config, nil
}

func normalizeLabelIDs(values []string) ([]string, error) {
	return appvalidation.CanonicalUUIDSet(ErrInvalidInput, "labelIds", values)
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
