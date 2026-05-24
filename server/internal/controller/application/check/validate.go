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
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
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
	var validation appvalidation.Collector

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	checkID, err := domaincheck.VNCheckID(input.CheckID)
	if err != nil {
		validation.AddError("checkId", err, input.CheckID)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return GetCheckInput{}, err
	}

	return GetCheckInput{CurrentUserID: input.CurrentUserID, ProjectRef: projectRef, CheckID: checkID}, nil
}

func normalizeCreateCheckInput(input CreateCheckInput) (normalizedCreateCheckInput, error) {
	var validation appvalidation.Collector

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	name, err := domaincheck.VNCheckName(input.Name)
	if err != nil {
		validation.AddError("name", err, input.Name)
	}
	checkType, err := domaincheck.VNCheckType(domaincheck.Type(input.Type))
	if err != nil {
		validation.AddError("type", err, input.Type)
	}
	target, err := domaincheck.VNCheckTarget(input.Target)
	if err != nil {
		validation.AddError("target", err, input.Target)
	}
	selector, err := normalizeCreateSelector(input.Selector)
	if err != nil {
		if !validation.AddValidation(err) {
			return normalizedCreateCheckInput{}, err
		}
	}
	description, err := domaincheck.VNCheckDescription(input.Description)
	if err != nil {
		validation.AddError("description", err, input.Description)
	}
	interval, err := domaincheck.VNCheckInterval(input.IntervalSeconds)
	if err != nil {
		validation.AddError("intervalSeconds", err, input.IntervalSeconds)
	}
	var pingConfig *domainping.Config
	var tcpConfig *domaintcp.Config
	var tracerouteConfig *domaintraceroute.Config
	if checkType != "" {
		pingConfig, tcpConfig, tracerouteConfig, err = normalizeCreateTypeConfig(checkType, input.PingConfig, input.TCPConfig, input.TracerouteConfig)
		if err != nil {
			if !validation.AddValidation(err) {
				return normalizedCreateCheckInput{}, err
			}
		}
	}
	labelIDs, err := domainlabel.VNLabelIDs(input.LabelIDs)
	if err != nil {
		validation.AddError("labelIds", err, input.LabelIDs)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return normalizedCreateCheckInput{}, err
	}

	return normalizedCreateCheckInput{
		projectRef:       projectRef,
		name:             name,
		checkType:        checkType,
		target:           target,
		selector:         selector,
		description:      description,
		intervalSeconds:  interval,
		pingConfig:       pingConfig,
		tcpConfig:        tcpConfig,
		tracerouteConfig: tracerouteConfig,
		labelIDs:         labelIDs,
	}, nil
}

func normalizeUpdateCheckInput(input UpdateCheckInput) (normalizedUpdateCheckInput, error) {
	var validation appvalidation.Collector

	projectRef, err := domainproject.VNProjectRef(input.ProjectRef)
	if err != nil {
		validation.AddError("projectRef", err, input.ProjectRef)
	}
	checkID, err := domaincheck.VNCheckID(input.CheckID)
	if err != nil {
		validation.AddError("checkId", err, input.CheckID)
	}

	if !hasUpdateCheckChanges(input) {
		validation.Add("", "at least one field must be provided", nil)
	}
	if validation.HasErrors() && !hasUpdateCheckChanges(input) {
		return normalizedUpdateCheckInput{}, validation.Err(ErrInvalidInput)
	}

	fields := normalizeUpdateCheckFields(input, &validation)
	pingConfig, tcpConfig, tracerouteConfig, labelIDs, replaceLabels := normalizeUpdateCheckCollections(input, &validation)
	if err := validation.Err(ErrInvalidInput); err != nil {
		return normalizedUpdateCheckInput{}, err
	}

	return normalizedUpdateCheckInput{
		projectRef:       projectRef,
		checkID:          checkID,
		name:             fields.name,
		checkType:        fields.checkType,
		target:           fields.target,
		selector:         fields.selector,
		description:      fields.description,
		intervalSeconds:  fields.intervalSeconds,
		pingConfig:       pingConfig,
		tcpConfig:        tcpConfig,
		tracerouteConfig: tracerouteConfig,
		replaceLabels:    replaceLabels,
		labelIDs:         labelIDs,
	}, nil
}

type normalizedUpdateCheckFields struct {
	name            *string
	checkType       *domaincheck.Type
	target          *string
	selector        json.RawMessage
	description     *string
	intervalSeconds *int32
}

func normalizeUpdateCheckFields(input UpdateCheckInput, validation *appvalidation.Collector) normalizedUpdateCheckFields {
	var fields normalizedUpdateCheckFields

	name, err := normalizeOptionalCheckName(input.Name)
	if err != nil {
		if !validation.AddValidation(err) {
			validation.AddError("name", err, input.Name)
		}
	} else {
		fields.name = name
	}
	checkType, err := normalizeOptionalCheckType(input.Type)
	if err != nil {
		if !validation.AddValidation(err) {
			validation.AddError("type", err, input.Type)
		}
	} else {
		fields.checkType = checkType
	}
	target, err := normalizeOptionalCheckTarget(input.Target)
	if err != nil {
		if !validation.AddValidation(err) {
			validation.AddError("target", err, input.Target)
		}
	} else {
		fields.target = target
	}
	selector, err := normalizeOptionalSelector(input.Selector)
	if err != nil {
		if !validation.AddValidation(err) {
			validation.AddError("selector", err, input.Selector)
		}
	} else {
		fields.selector = selector
	}
	description, err := domaincheck.VNCheckDescription(input.Description)
	if err != nil {
		validation.AddError("description", err, input.Description)
	} else {
		fields.description = description
	}

	intervalSeconds, err := normalizeOptionalCheckInterval(input.IntervalSeconds)
	if err != nil {
		if !validation.AddValidation(err) {
			validation.AddError("intervalSeconds", err, input.IntervalSeconds)
		}
	} else {
		fields.intervalSeconds = intervalSeconds
	}

	return fields
}

func normalizeUpdateCheckCollections(input UpdateCheckInput, validation *appvalidation.Collector) (updatePingConfigPatch, updateTCPConfigPatch, updateTracerouteConfigPatch, []string, bool) {
	pingConfig, err := normalizeUpdatePingConfig(input.PingConfig)
	if err != nil {
		if !validation.AddValidation(err) {
			validation.AddError("pingConfig", err, input.PingConfig)
		}
	}
	tcpConfig, err := normalizeUpdateTCPConfig(input.TCPConfig)
	if err != nil {
		if !validation.AddValidation(err) {
			validation.AddError("tcpConfig", err, input.TCPConfig)
		}
	}
	tracerouteConfig, err := normalizeUpdateTracerouteConfig(input.TracerouteConfig)
	if err != nil {
		if !validation.AddValidation(err) {
			validation.AddError("tracerouteConfig", err, input.TracerouteConfig)
		}
	}

	labelIDs, replaceLabels, err := normalizeUpdateLabelIDs(input.LabelIDs)
	if err != nil {
		if !validation.AddValidation(err) {
			validation.AddError("labelIds", err, input.LabelIDs)
		}
	}

	return pingConfig, tcpConfig, tracerouteConfig, labelIDs, replaceLabels
}

type updatePingConfigPatch struct {
	packetCount     *int32
	packetSizeBytes *int32
	timeoutMs       *int32
	ipFamily        *domainnetwork.IPFamily
}

type updateTCPConfigPatch struct {
	port      *int32
	timeoutMs *int32
	ipFamily  *domainnetwork.IPFamily
}

type updateTracerouteConfigPatch struct {
	protocol        *domaintraceroute.Protocol
	maxHops         *int32
	timeoutMs       *int32
	queriesPerHop   *int32
	packetSizeBytes *int32
	port            *int32
	ipFamily        *domainnetwork.IPFamily
}

func (patch updatePingConfigPatch) hasChanges() bool {
	return patch.packetCount != nil ||
		patch.packetSizeBytes != nil ||
		patch.timeoutMs != nil ||
		patch.ipFamily != nil
}

func (patch updateTCPConfigPatch) hasChanges() bool {
	return patch.port != nil ||
		patch.timeoutMs != nil ||
		patch.ipFamily != nil
}

func (patch updateTracerouteConfigPatch) hasChanges() bool {
	return patch.protocol != nil ||
		patch.maxHops != nil ||
		patch.timeoutMs != nil ||
		patch.queriesPerHop != nil ||
		patch.packetSizeBytes != nil ||
		patch.port != nil ||
		patch.ipFamily != nil
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

	var validation appvalidation.Collector
	var patch updatePingConfigPatch

	packetCount, err := normalizeOptionalPacketCount(input.PacketCount)
	if err != nil {
		if !validation.AddValidation(err) {
			return updatePingConfigPatch{}, err
		}
	} else {
		patch.packetCount = packetCount
	}
	packetSizeBytes, err := normalizeOptionalPacketSizeBytes(input.PacketSizeBytes)
	if err != nil {
		if !validation.AddValidation(err) {
			return updatePingConfigPatch{}, err
		}
	} else {
		patch.packetSizeBytes = packetSizeBytes
	}
	timeoutMs, err := normalizeOptionalTimeoutMs(input.TimeoutMs)
	if err != nil {
		if !validation.AddValidation(err) {
			return updatePingConfigPatch{}, err
		}
	} else {
		patch.timeoutMs = timeoutMs
	}
	ipFamily, err := normalizeOptionalIPFamily(input.IPFamily)
	if err != nil {
		if !validation.AddValidation(err) {
			return updatePingConfigPatch{}, err
		}
	} else {
		patch.ipFamily = ipFamily
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return updatePingConfigPatch{}, err
	}

	return patch, nil
}

func normalizeUpdateTCPConfig(input *TCPConfigInput) (updateTCPConfigPatch, error) {
	if input == nil {
		return updateTCPConfigPatch{}, nil
	}

	var validation appvalidation.Collector
	var patch updateTCPConfigPatch

	port, err := normalizeOptionalTCPPort(input.Port)
	if err != nil {
		if !validation.AddValidation(err) {
			return updateTCPConfigPatch{}, err
		}
	} else {
		patch.port = port
	}
	timeoutMs, err := normalizeOptionalTCPTimeoutMs(input.TimeoutMs)
	if err != nil {
		if !validation.AddValidation(err) {
			return updateTCPConfigPatch{}, err
		}
	} else {
		patch.timeoutMs = timeoutMs
	}
	ipFamily, err := normalizeOptionalTCPIPFamily(input.IPFamily)
	if err != nil {
		if !validation.AddValidation(err) {
			return updateTCPConfigPatch{}, err
		}
	} else {
		patch.ipFamily = ipFamily
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return updateTCPConfigPatch{}, err
	}

	return patch, nil
}

func normalizeUpdateTracerouteConfig(input *TracerouteConfigInput) (updateTracerouteConfigPatch, error) {
	if input == nil {
		return updateTracerouteConfigPatch{}, nil
	}

	var validation appvalidation.Collector
	var patch updateTracerouteConfigPatch

	normalizeUpdateTracerouteRouteConfig(input, &patch, &validation)
	normalizeUpdateTracerouteProbeConfig(input, &patch, &validation)
	if err := validation.Err(ErrInvalidInput); err != nil {
		return updateTracerouteConfigPatch{}, err
	}

	return patch, nil
}

func normalizeUpdateTracerouteRouteConfig(input *TracerouteConfigInput, patch *updateTracerouteConfigPatch, validation *appvalidation.Collector) {
	protocol, err := normalizeOptionalTracerouteProtocol(input.Protocol)
	if err != nil {
		if !validation.AddValidation(err) {
			validation.AddError("tracerouteConfig.protocol", err, input.Protocol)
		}
	} else {
		patch.protocol = protocol
	}
	maxHops, err := normalizeOptionalTracerouteMaxHops(input.MaxHops)
	if err != nil {
		if !validation.AddValidation(err) {
			validation.AddError("tracerouteConfig.maxHops", err, input.MaxHops)
		}
	} else {
		patch.maxHops = maxHops
	}
	timeoutMs, err := normalizeOptionalTracerouteTimeoutMs(input.TimeoutMs)
	if err != nil {
		if !validation.AddValidation(err) {
			validation.AddError("tracerouteConfig.timeoutMs", err, input.TimeoutMs)
		}
	} else {
		patch.timeoutMs = timeoutMs
	}
}

func normalizeUpdateTracerouteProbeConfig(input *TracerouteConfigInput, patch *updateTracerouteConfigPatch, validation *appvalidation.Collector) {
	queriesPerHop, err := normalizeOptionalTracerouteQueriesPerHop(input.QueriesPerHop)
	if err != nil {
		if !validation.AddValidation(err) {
			validation.AddError("tracerouteConfig.queriesPerHop", err, input.QueriesPerHop)
		}
	} else {
		patch.queriesPerHop = queriesPerHop
	}
	packetSizeBytes, err := normalizeOptionalTraceroutePacketSizeBytes(input.PacketSizeBytes)
	if err != nil {
		if !validation.AddValidation(err) {
			validation.AddError("tracerouteConfig.packetSizeBytes", err, input.PacketSizeBytes)
		}
	} else {
		patch.packetSizeBytes = packetSizeBytes
	}
	port, err := normalizeOptionalTraceroutePort(input.Port)
	if err != nil {
		if !validation.AddValidation(err) {
			validation.AddError("tracerouteConfig.port", err, input.Port)
		}
	} else {
		patch.port = port
	}
	ipFamily, err := normalizeOptionalTracerouteIPFamily(input.IPFamily)
	if err != nil {
		if !validation.AddValidation(err) {
			validation.AddError("tracerouteConfig.ipFamily", err, input.IPFamily)
		}
	} else {
		patch.ipFamily = ipFamily
	}
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

func normalizeOptionalTCPPort(input *int32) (*int32, error) {
	if input == nil {
		return nil, nil //nolint:nilnil // Nil means the update did not include a TCP port.
	}
	port, err := domaintcp.VNConfigPort(*input)
	if err != nil {
		return nil, invalidCheckField("tcpConfig.port", err.Error(), input)
	}
	return &port, nil
}

func normalizeOptionalTCPTimeoutMs(input *int32) (*int32, error) {
	if input == nil {
		return nil, nil //nolint:nilnil // Nil means the update did not include a TCP timeout.
	}
	timeoutMs, err := domaintcp.VNConfigTimeoutMs(*input)
	if err != nil {
		return nil, invalidCheckField("tcpConfig.timeoutMs", err.Error(), input)
	}
	return &timeoutMs, nil
}

func normalizeOptionalTCPIPFamily(input *string) (*domainnetwork.IPFamily, error) {
	ipFamily, err := domainnetwork.ParseOptionalIPFamily(input)
	if err != nil {
		return nil, invalidCheckField("tcpConfig.ipFamily", `must be "inet" or "inet6"`, input)
	}
	ipFamily, err = domaintcp.VNConfigIPFamily(ipFamily)
	if err != nil {
		return nil, invalidCheckField("tcpConfig.ipFamily", err.Error(), input)
	}
	return ipFamily, nil
}

func normalizeOptionalTracerouteProtocol(input *string) (*domaintraceroute.Protocol, error) {
	if input == nil {
		return nil, nil //nolint:nilnil // Nil means the update did not include a protocol.
	}
	protocol, err := domaintraceroute.VNConfigProtocol(domaintraceroute.Protocol(*input))
	if err != nil {
		return nil, invalidCheckField("tracerouteConfig.protocol", err.Error(), input)
	}
	return &protocol, nil
}

func normalizeOptionalTracerouteMaxHops(input *int32) (*int32, error) {
	if input == nil {
		return nil, nil //nolint:nilnil // Nil means the update did not include max hops.
	}
	maxHops, err := domaintraceroute.VNConfigMaxHops(*input)
	if err != nil {
		return nil, invalidCheckField("tracerouteConfig.maxHops", err.Error(), input)
	}
	return &maxHops, nil
}

func normalizeOptionalTracerouteTimeoutMs(input *int32) (*int32, error) {
	if input == nil {
		return nil, nil //nolint:nilnil // Nil means the update did not include a timeout.
	}
	timeoutMs, err := domaintraceroute.VNConfigTimeoutMs(*input)
	if err != nil {
		return nil, invalidCheckField("tracerouteConfig.timeoutMs", err.Error(), input)
	}
	return &timeoutMs, nil
}

func normalizeOptionalTracerouteQueriesPerHop(input *int32) (*int32, error) {
	if input == nil {
		return nil, nil //nolint:nilnil // Nil means the update did not include queries per hop.
	}
	queriesPerHop, err := domaintraceroute.VNConfigQueriesPerHop(*input)
	if err != nil {
		return nil, invalidCheckField("tracerouteConfig.queriesPerHop", err.Error(), input)
	}
	return &queriesPerHop, nil
}

func normalizeOptionalTraceroutePacketSizeBytes(input *int32) (*int32, error) {
	if input == nil {
		return nil, nil //nolint:nilnil // Nil means the update did not include a packet size.
	}
	packetSizeBytes, err := domaintraceroute.VNConfigPacketSizeBytes(*input)
	if err != nil {
		return nil, invalidCheckField("tracerouteConfig.packetSizeBytes", err.Error(), input)
	}
	return &packetSizeBytes, nil
}

func normalizeOptionalTraceroutePort(input *int32) (*int32, error) {
	if input == nil {
		return nil, nil //nolint:nilnil // Nil means the update did not include a port.
	}
	port, err := domaintraceroute.VNConfigPort(*input)
	if err != nil {
		return nil, invalidCheckField("tracerouteConfig.port", err.Error(), input)
	}
	return &port, nil
}

func normalizeOptionalTracerouteIPFamily(input *string) (*domainnetwork.IPFamily, error) {
	ipFamily, err := domainnetwork.ParseOptionalIPFamily(input)
	if err != nil {
		return nil, invalidCheckField("tracerouteConfig.ipFamily", `must be "inet" or "inet6"`, input)
	}
	ipFamily, err = domaintraceroute.VNConfigIPFamily(ipFamily)
	if err != nil {
		return nil, invalidCheckField("tracerouteConfig.ipFamily", err.Error(), input)
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
		input.TCPConfig.hasChanges() ||
		input.TracerouteConfig.hasChanges() ||
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

func (input *TCPConfigInput) hasChanges() bool {
	if input == nil {
		return false
	}
	return input.Port != nil ||
		input.TimeoutMs != nil ||
		input.IPFamily != nil
}

func (input *TracerouteConfigInput) hasChanges() bool {
	if input == nil {
		return false
	}
	return input.Protocol != nil ||
		input.MaxHops != nil ||
		input.TimeoutMs != nil ||
		input.QueriesPerHop != nil ||
		input.PacketSizeBytes != nil ||
		input.Port != nil ||
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

func normalizeCreateTypeConfig(checkType domaincheck.Type, pingInput *PingConfigInput, tcpInput *TCPConfigInput, tracerouteInput *TracerouteConfigInput) (*domainping.Config, *domaintcp.Config, *domaintraceroute.Config, error) {
	switch checkType {
	case domaincheck.TypePing:
		var validation appvalidation.Collector
		if tcpInput != nil {
			validation.Add("tcpConfig", "must be omitted for ping checks", tcpInput)
		}
		if tracerouteInput != nil {
			validation.Add("tracerouteConfig", "must be omitted for ping checks", tracerouteInput)
		}
		config, err := normalizePingConfig(pingInput)
		if err != nil {
			if !validation.AddValidation(err) {
				return nil, nil, nil, err
			}
		}
		if err := validation.Err(ErrInvalidInput); err != nil {
			return nil, nil, nil, err
		}

		return &config, nil, nil, nil
	case domaincheck.TypeTCP:
		var validation appvalidation.Collector
		if pingInput != nil {
			validation.Add("pingConfig", "must be omitted for tcp checks", pingInput)
		}
		if tracerouteInput != nil {
			validation.Add("tracerouteConfig", "must be omitted for tcp checks", tracerouteInput)
		}
		config, err := normalizeTCPConfig(tcpInput)
		if err != nil {
			if !validation.AddValidation(err) {
				return nil, nil, nil, err
			}
		}
		if err := validation.Err(ErrInvalidInput); err != nil {
			return nil, nil, nil, err
		}

		return nil, &config, nil, nil
	case domaincheck.TypeTraceroute:
		var validation appvalidation.Collector
		if pingInput != nil {
			validation.Add("pingConfig", "must be omitted for traceroute checks", pingInput)
		}
		if tcpInput != nil {
			validation.Add("tcpConfig", "must be omitted for traceroute checks", tcpInput)
		}
		config, err := normalizeTracerouteConfig(tracerouteInput)
		if err != nil {
			if !validation.AddValidation(err) {
				return nil, nil, nil, err
			}
		}
		if err := validation.Err(ErrInvalidInput); err != nil {
			return nil, nil, nil, err
		}

		return nil, nil, &config, nil
	default:
		return nil, nil, nil, invalidCheckField("type", "unsupported check type", string(checkType))
	}
}

func normalizePingConfig(input *PingConfigInput) (domainping.Config, error) {
	config := domainping.DefaultConfig()

	if input == nil {
		return config, nil
	}

	var validation appvalidation.Collector

	if input.PacketCount != nil {
		packetCount, err := domainping.VNConfigPacketCount(*input.PacketCount)
		if err != nil {
			validation.AddError("packetCount", err, input.PacketCount)
		} else {
			config.PacketCount = packetCount
		}
	}
	if input.PacketSizeBytes != nil {
		packetSizeBytes, err := domainping.VNConfigPacketSizeBytes(*input.PacketSizeBytes)
		if err != nil {
			validation.AddError("packetSizeBytes", err, input.PacketSizeBytes)
		} else {
			config.PacketSizeBytes = packetSizeBytes
		}
	}
	if input.TimeoutMs != nil {
		timeoutMs, err := domainping.VNConfigTimeoutMs(*input.TimeoutMs)
		if err != nil {
			validation.AddError("timeoutMs", err, input.TimeoutMs)
		} else {
			config.TimeoutMs = timeoutMs
		}
	}

	ipFamily, err := domainnetwork.ParseOptionalIPFamily(input.IPFamily)
	if err != nil {
		validation.Add("ipFamily", `must be "inet" or "inet6"`, input.IPFamily)
	} else {
		ipFamily, err = domainping.VNConfigIPFamily(ipFamily)
		if err != nil {
			validation.AddError("ipFamily", err, input.IPFamily)
		} else {
			config.IPFamily = ipFamily
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return domainping.Config{}, err
	}

	return config, nil
}

func normalizeTCPConfig(input *TCPConfigInput) (domaintcp.Config, error) {
	config := domaintcp.DefaultConfig()

	if input == nil {
		return config, nil
	}

	var validation appvalidation.Collector

	if input.Port != nil {
		port, err := domaintcp.VNConfigPort(*input.Port)
		if err != nil {
			validation.AddError("tcpConfig.port", err, input.Port)
		} else {
			config.Port = port
		}
	}
	if input.TimeoutMs != nil {
		timeoutMs, err := domaintcp.VNConfigTimeoutMs(*input.TimeoutMs)
		if err != nil {
			validation.AddError("tcpConfig.timeoutMs", err, input.TimeoutMs)
		} else {
			config.TimeoutMs = timeoutMs
		}
	}

	ipFamily, err := normalizeTCPConfigIPFamily(input.IPFamily)
	if err != nil {
		if !validation.AddValidation(err) {
			validation.AddError("tcpConfig.ipFamily", err, input.IPFamily)
		}
	} else {
		config.IPFamily = ipFamily
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return domaintcp.Config{}, err
	}

	return config, nil
}

func normalizeTCPConfigIPFamily(input *string) (*domainnetwork.IPFamily, error) {
	ipFamily, err := domainnetwork.ParseOptionalIPFamily(input)
	if err != nil {
		return nil, invalidCheckField("tcpConfig.ipFamily", `must be "inet" or "inet6"`, input)
	}
	ipFamily, err = domaintcp.VNConfigIPFamily(ipFamily)
	if err != nil {
		return nil, invalidCheckField("tcpConfig.ipFamily", err.Error(), input)
	}
	return ipFamily, nil
}

func normalizeTracerouteConfig(input *TracerouteConfigInput) (domaintraceroute.Config, error) {
	config := domaintraceroute.DefaultConfig()

	if input == nil {
		return config, nil
	}

	var validation appvalidation.Collector

	normalizeTracerouteRouteConfig(input, &config, &validation)
	normalizeTracerouteProbeConfig(input, &config, &validation)
	if err := validation.Err(ErrInvalidInput); err != nil {
		return domaintraceroute.Config{}, err
	}

	return config, nil
}

func normalizeTracerouteRouteConfig(input *TracerouteConfigInput, config *domaintraceroute.Config, validation *appvalidation.Collector) {
	if input.Protocol != nil {
		protocol, err := domaintraceroute.VNConfigProtocol(domaintraceroute.Protocol(*input.Protocol))
		if err != nil {
			validation.AddError("tracerouteConfig.protocol", err, input.Protocol)
		} else {
			config.Protocol = protocol
		}
	}
	if input.MaxHops != nil {
		maxHops, err := domaintraceroute.VNConfigMaxHops(*input.MaxHops)
		if err != nil {
			validation.AddError("tracerouteConfig.maxHops", err, input.MaxHops)
		} else {
			config.MaxHops = maxHops
		}
	}
	if input.TimeoutMs != nil {
		timeoutMs, err := domaintraceroute.VNConfigTimeoutMs(*input.TimeoutMs)
		if err != nil {
			validation.AddError("tracerouteConfig.timeoutMs", err, input.TimeoutMs)
		} else {
			config.TimeoutMs = timeoutMs
		}
	}
}

func normalizeTracerouteProbeConfig(input *TracerouteConfigInput, config *domaintraceroute.Config, validation *appvalidation.Collector) {
	if input.QueriesPerHop != nil {
		queriesPerHop, err := domaintraceroute.VNConfigQueriesPerHop(*input.QueriesPerHop)
		if err != nil {
			validation.AddError("tracerouteConfig.queriesPerHop", err, input.QueriesPerHop)
		} else {
			config.QueriesPerHop = queriesPerHop
		}
	}
	if input.PacketSizeBytes != nil {
		packetSizeBytes, err := domaintraceroute.VNConfigPacketSizeBytes(*input.PacketSizeBytes)
		if err != nil {
			validation.AddError("tracerouteConfig.packetSizeBytes", err, input.PacketSizeBytes)
		} else {
			config.PacketSizeBytes = packetSizeBytes
		}
	}
	if input.Port != nil {
		port, err := domaintraceroute.VNConfigPort(*input.Port)
		if err != nil {
			validation.AddError("tracerouteConfig.port", err, input.Port)
		} else {
			config.Port = port
		}
	}

	ipFamily, err := normalizeTracerouteConfigIPFamily(input.IPFamily)
	if err != nil {
		if !validation.AddValidation(err) {
			validation.AddError("tracerouteConfig.ipFamily", err, input.IPFamily)
		}
	} else {
		config.IPFamily = ipFamily
	}
}

func normalizeTracerouteConfigIPFamily(input *string) (*domainnetwork.IPFamily, error) {
	ipFamily, err := domainnetwork.ParseOptionalIPFamily(input)
	if err != nil {
		return nil, invalidCheckField("tracerouteConfig.ipFamily", `must be "inet" or "inet6"`, input)
	}
	ipFamily, err = domaintraceroute.VNConfigIPFamily(ipFamily)
	if err != nil {
		return nil, invalidCheckField("tracerouteConfig.ipFamily", err.Error(), input)
	}
	return ipFamily, nil
}
