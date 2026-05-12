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
	interval, err := domaincheck.VNCheckInterval(int(input.IntervalSeconds))
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
		intervalSeconds: int32(interval),
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

	var name *string
	if input.Name != nil {
		normalizedName, err := domaincheck.VNCheckName(*input.Name)
		if err != nil {
			return normalizedUpdateCheckInput{}, invalidCheckField("name", err.Error(), input.Name)
		}
		name = &normalizedName
	}

	var checkType *domaincheck.Type
	if input.Type != nil {
		normalizedType, err := domaincheck.VNCheckType(domaincheck.Type(*input.Type))
		if err != nil {
			return normalizedUpdateCheckInput{}, invalidCheckField("type", err.Error(), input.Type)
		}
		checkType = &normalizedType
	}

	var target *string
	if input.Target != nil {
		normalizedTarget, err := domaincheck.VNCheckTarget(*input.Target)
		if err != nil {
			return normalizedUpdateCheckInput{}, invalidCheckField("target", err.Error(), input.Target)
		}
		target = &normalizedTarget
	}

	selector, err := normalizeOptionalSelector(input.Selector)
	if err != nil {
		return normalizedUpdateCheckInput{}, err
	}
	description, err := domaincheck.VNCheckDescription(input.Description)
	if err != nil {
		return normalizedUpdateCheckInput{}, invalidCheckField("description", err.Error(), input.Description)
	}

	var intervalSeconds *int32
	if input.IntervalSeconds != nil {
		interval, err := domaincheck.VNCheckInterval(int(*input.IntervalSeconds))
		if err != nil {
			return normalizedUpdateCheckInput{}, invalidCheckField("intervalSeconds", err.Error(), input.IntervalSeconds)
		}
		normalizedInterval := int32(interval)
		intervalSeconds = &normalizedInterval
	}

	var packetCount *int32
	var packetSizeBytes *int32
	var timeoutMs *int32
	var ipFamily *domainnetwork.IPFamily
	if input.PingConfig != nil {
		if input.PingConfig.PacketCount != nil {
			normalizedPacketCount, err := domainping.VNConfigPacketCount(*input.PingConfig.PacketCount)
			if err != nil {
				return normalizedUpdateCheckInput{}, invalidCheckField("packetCount", err.Error(), input.PingConfig.PacketCount)
			}
			packetCount = &normalizedPacketCount
		}
		if input.PingConfig.PacketSizeBytes != nil {
			normalizedPacketSizeBytes, err := domainping.VNConfigPacketSizeBytes(*input.PingConfig.PacketSizeBytes)
			if err != nil {
				return normalizedUpdateCheckInput{}, invalidCheckField("packetSizeBytes", err.Error(), input.PingConfig.PacketSizeBytes)
			}
			packetSizeBytes = &normalizedPacketSizeBytes
		}
		if input.PingConfig.TimeoutMs != nil {
			normalizedTimeoutMs, err := domainping.VNConfigTimeoutMs(*input.PingConfig.TimeoutMs)
			if err != nil {
				return normalizedUpdateCheckInput{}, invalidCheckField("timeoutMs", err.Error(), input.PingConfig.TimeoutMs)
			}
			timeoutMs = &normalizedTimeoutMs
		}
		ipFamily, err = domainnetwork.ParseOptionalIPFamily(input.PingConfig.IPFamily)
		if err != nil {
			return normalizedUpdateCheckInput{}, invalidCheckField("ipFamily", `must be "inet" or "inet6"`, input.PingConfig.IPFamily)
		}
		ipFamily, err = domainping.VNConfigIPFamily(ipFamily)
		if err != nil {
			return normalizedUpdateCheckInput{}, invalidCheckField("ipFamily", err.Error(), input.PingConfig.IPFamily)
		}
	}

	replaceLabels := input.LabelIDs != nil
	var labelIDs []string
	if input.LabelIDs != nil {
		labelIDs, err = domainlabel.VNLabelIDs(*input.LabelIDs)
		if err != nil {
			return normalizedUpdateCheckInput{}, invalidCheckField("labelIds", err.Error(), input.LabelIDs)
		}
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
