package check

import (
	"encoding/json"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainselector "github.com/yorukot/netstamp/internal/domain/selector"
)

var defaultSelector = json.RawMessage(`{}`)

func normalizeListChecksInput(input ListChecksInput) (ListChecksInput, error) {
	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		return ListChecksInput{}, invalidCheckField("projectRef", err.Error(), input.ProjectRef)
	}

	return ListChecksInput{CurrentUserID: input.CurrentUserID, ProjectRef: projectRef}, nil
}

func normalizeTargetCheckInput(input GetCheckInput) (GetCheckInput, error) {
	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		return GetCheckInput{}, invalidCheckField("projectRef", err.Error(), input.ProjectRef)
	}
	checkID, err := domaincheck.VNCheckID(input.CheckID)
	if err != nil {
		return GetCheckInput{}, invalidCheckField("checkId", err.Error(), input.CheckID)
	}

	return GetCheckInput{CurrentUserID: input.CurrentUserID, ProjectRef: projectRef, CheckID: checkID}, nil
}

func normalizeCreateCheckInput(input CreateCheckInput) (normalizedCreateCheckInput, error) {
	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		return normalizedCreateCheckInput{}, invalidCheckField("projectRef", err.Error(), input.ProjectRef)
	}
	name, err := domaincheck.VNCheckName(input.Name)
	if err != nil {
		return normalizedCreateCheckInput{}, invalidCheckField("name", err.Error(), input.Name)
	}
	checkType, err := domaincheck.VNCheckType(domaincheck.Type(input.Type))
	if err != nil {
		return normalizedCreateCheckInput{}, invalidCheckField("type", err.Error(), input.Type)
	}
	target, err := domaincheck.VNCheckTarget(input.Target)
	if err != nil {
		return normalizedCreateCheckInput{}, invalidCheckField("target", err.Error(), input.Target)
	}
	selector, err := normalizeCreateSelector(input.Selector)
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}
	description, err := domaincheck.VNCheckDescription(input.Description)
	if err != nil {
		return normalizedCreateCheckInput{}, invalidCheckField("description", err.Error(), input.Description)
	}
	interval, err := domaincheck.VNCheckInterval(input.IntervalSeconds)
	if err != nil {
		return normalizedCreateCheckInput{}, invalidCheckField("intervalSeconds", err.Error(), input.IntervalSeconds)
	}
	pingConfig, err := normalizePingConfig(input.PingConfig)
	if err != nil {
		return normalizedCreateCheckInput{}, err
	}
	labelIDs, err := domainlabel.VNLabelIDs(input.LabelIDs)
	if err != nil {
		return normalizedCreateCheckInput{}, invalidCheckField("labelIds", err.Error(), input.LabelIDs)
	}

	return normalizedCreateCheckInput{
		projectRef:      projectRef,
		name:            name,
		checkType:       checkType,
		target:          target,
		selector:        selector,
		description:     description,
		intervalSeconds: interval,
		pingConfig:      pingConfig,
		labelIDs:        labelIDs,
	}, nil
}

func normalizeUpdateCheckInput(input UpdateCheckInput) (normalizedUpdateCheckInput, error) {
	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		return normalizedUpdateCheckInput{}, invalidCheckField("projectRef", err.Error(), input.ProjectRef)
	}
	checkID, err := domaincheck.VNCheckID(input.CheckID)
	if err != nil {
		return normalizedUpdateCheckInput{}, invalidCheckField("checkId", err.Error(), input.CheckID)
	}

	if !hasUpdateCheckChanges(input) {
		return normalizedUpdateCheckInput{}, invalidCheckField("", "at least one field must be provided", nil)
	}

	name, err := normalizeOptionalCheckName(input.Name)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	checkType, err := normalizeOptionalCheckType(input.Type)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	target, err := normalizeOptionalCheckTarget(input.Target)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	selector, err := normalizeOptionalSelector(input.Selector)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	description, err := domaincheck.VNCheckDescription(input.Description)
	if err != nil {
		return normalizedUpdateCheckInput{}, invalidCheckField("description", err.Error(), input.Description)
	}

	intervalSeconds, err := normalizeOptionalCheckInterval(input.IntervalSeconds)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	pingConfig, err := normalizeUpdatePingConfig(input.PingConfig)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}

	labelIDs, replaceLabels, err := normalizeUpdateLabelIDs(input.LabelIDs)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}

	return normalizedUpdateCheckInput{
		projectRef:      projectRef,
		checkID:         checkID,
		name:            name,
		checkType:       checkType,
		target:          target,
		selector:        selector,
		description:     description,
		intervalSeconds: intervalSeconds,
		packetCount:     pingConfig.packetCount,
		packetSizeBytes: pingConfig.packetSizeBytes,
		timeoutMs:       pingConfig.timeoutMs,
		ipFamily:        pingConfig.ipFamily,
		replaceLabels:   replaceLabels,
		labelIDs:        labelIDs,
	}, nil
}

type updatePingConfigPatch struct {
	packetCount     *int32
	packetSizeBytes *int32
	timeoutMs       *int32
	ipFamily        *domainnetwork.IPFamily
}

func normalizeOptionalCheckName(input *string) (*string, error) {
	if input == nil {
		return nil, nil //nolint:nilnil // Nil means the update did not include a name.
	}
	name, err := domaincheck.VNCheckName(*input)
	if err != nil {
		return nil, invalidCheckField("name", err.Error(), input)
	}
	return &name, nil
}

func normalizeOptionalCheckType(input *string) (*domaincheck.Type, error) {
	if input == nil {
		return nil, nil //nolint:nilnil // Nil means the update did not include a check type.
	}
	checkType, err := domaincheck.VNCheckType(domaincheck.Type(*input))
	if err != nil {
		return nil, invalidCheckField("type", err.Error(), input)
	}
	return &checkType, nil
}

func normalizeOptionalCheckTarget(input *string) (*string, error) {
	if input == nil {
		return nil, nil //nolint:nilnil // Nil means the update did not include a target.
	}
	target, err := domaincheck.VNCheckTarget(*input)
	if err != nil {
		return nil, invalidCheckField("target", err.Error(), input)
	}
	return &target, nil
}

func normalizeOptionalCheckInterval(input *int32) (*int32, error) {
	if input == nil {
		return nil, nil //nolint:nilnil // Nil means the update did not include an interval.
	}
	interval, err := domaincheck.VNCheckInterval(*input)
	if err != nil {
		return nil, invalidCheckField("intervalSeconds", err.Error(), input)
	}
	return &interval, nil
}

func normalizeUpdatePingConfig(input *PingConfigInput) (updatePingConfigPatch, error) {
	if input == nil {
		return updatePingConfigPatch{}, nil
	}

	packetCount, err := normalizeOptionalPacketCount(input.PacketCount)
	if err != nil {
		return updatePingConfigPatch{}, err
	}
	packetSizeBytes, err := normalizeOptionalPacketSizeBytes(input.PacketSizeBytes)
	if err != nil {
		return updatePingConfigPatch{}, err
	}
	timeoutMs, err := normalizeOptionalTimeoutMs(input.TimeoutMs)
	if err != nil {
		return updatePingConfigPatch{}, err
	}
	ipFamily, err := normalizeOptionalIPFamily(input.IPFamily)
	if err != nil {
		return updatePingConfigPatch{}, err
	}

	return updatePingConfigPatch{
		packetCount:     packetCount,
		packetSizeBytes: packetSizeBytes,
		timeoutMs:       timeoutMs,
		ipFamily:        ipFamily,
	}, nil
}

func normalizeOptionalPacketCount(input *int32) (*int32, error) {
	if input == nil {
		return nil, nil //nolint:nilnil // Nil means the update did not include a packet count.
	}
	packetCount, err := domainping.VNConfigPacketCount(*input)
	if err != nil {
		return nil, invalidCheckField("packetCount", err.Error(), input)
	}
	return &packetCount, nil
}

func normalizeOptionalPacketSizeBytes(input *int32) (*int32, error) {
	if input == nil {
		return nil, nil //nolint:nilnil // Nil means the update did not include a packet size.
	}
	packetSizeBytes, err := domainping.VNConfigPacketSizeBytes(*input)
	if err != nil {
		return nil, invalidCheckField("packetSizeBytes", err.Error(), input)
	}
	return &packetSizeBytes, nil
}

func normalizeOptionalTimeoutMs(input *int32) (*int32, error) {
	if input == nil {
		return nil, nil //nolint:nilnil // Nil means the update did not include a timeout.
	}
	timeoutMs, err := domainping.VNConfigTimeoutMs(*input)
	if err != nil {
		return nil, invalidCheckField("timeoutMs", err.Error(), input)
	}
	return &timeoutMs, nil
}

func normalizeOptionalIPFamily(input *string) (*domainnetwork.IPFamily, error) {
	ipFamily, err := domainnetwork.ParseOptionalIPFamily(input)
	if err != nil {
		return nil, invalidCheckField("ipFamily", `must be "inet" or "inet6"`, input)
	}
	ipFamily, err = domainping.VNConfigIPFamily(ipFamily)
	if err != nil {
		return nil, invalidCheckField("ipFamily", err.Error(), input)
	}
	return ipFamily, nil
}

func normalizeUpdateLabelIDs(input *[]string) ([]string, bool, error) {
	if input == nil {
		return nil, false, nil
	}
	labelIDs, err := domainlabel.VNLabelIDs(*input)
	if err != nil {
		return nil, false, invalidCheckField("labelIds", err.Error(), input)
	}
	return labelIDs, true, nil
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
	if input == nil {
		return false
	}
	return input.PacketCount != nil ||
		input.PacketSizeBytes != nil ||
		input.TimeoutMs != nil ||
		input.IPFamily != nil
}

func invalidCheckField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
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
	canonical, err = domaincheck.VNCheckSelector(canonical)
	if err != nil {
		return nil, invalidCheckField("selector", err.Error(), selector)
	}

	return canonical, nil
}

func normalizeCreateSelector(selector map[string]any) (json.RawMessage, error) {
	if selector == nil {
		canonical, err := canonicalizeSelector(nil)
		if err != nil {
			return nil, err
		}
		canonical, err = domaincheck.VNCheckSelector(canonical)
		if err != nil {
			return nil, invalidCheckField("selector", err.Error(), selector)
		}

		return canonical, nil
	}

	raw, err := normalizeSelector(selector)
	if err != nil {
		return nil, err
	}

	canonical, err := canonicalizeSelector(raw)
	if err != nil {
		return nil, invalidCheckField("selector", "must be a valid selector", selector)
	}
	canonical, err = domaincheck.VNCheckSelector(canonical)
	if err != nil {
		return nil, invalidCheckField("selector", err.Error(), selector)
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

	if input == nil {
		return config, nil
	}

	if input.PacketCount != nil {
		packetCount, err := domainping.VNConfigPacketCount(*input.PacketCount)
		if err != nil {
			return domainping.Config{}, invalidCheckField("packetCount", err.Error(), input.PacketCount)
		}
		config.PacketCount = packetCount
	}
	if input.PacketSizeBytes != nil {
		packetSizeBytes, err := domainping.VNConfigPacketSizeBytes(*input.PacketSizeBytes)
		if err != nil {
			return domainping.Config{}, invalidCheckField("packetSizeBytes", err.Error(), input.PacketSizeBytes)
		}
		config.PacketSizeBytes = packetSizeBytes
	}
	if input.TimeoutMs != nil {
		timeoutMs, err := domainping.VNConfigTimeoutMs(*input.TimeoutMs)
		if err != nil {
			return domainping.Config{}, invalidCheckField("timeoutMs", err.Error(), input.TimeoutMs)
		}
		config.TimeoutMs = timeoutMs
	}

	ipFamily, err := domainnetwork.ParseOptionalIPFamily(input.IPFamily)
	if err != nil {
		return domainping.Config{}, invalidCheckField("ipFamily", `must be "inet" or "inet6"`, input.IPFamily)
	}
	ipFamily, err = domainping.VNConfigIPFamily(ipFamily)
	if err != nil {
		return domainping.Config{}, invalidCheckField("ipFamily", err.Error(), input.IPFamily)
	}
	config.IPFamily = ipFamily

	return config, nil
}
