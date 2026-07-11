package check

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
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
	target, err := domaincheck.VNCheckTargetForType(checkType, input.Target)
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
	configs, err := normalizeCreateCheckConfigs(checkType, input)
	if err != nil && !validation.AddValidation(err) {
		return normalizedCreateCheckInput{}, err
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
		pingConfig:       configs.ping,
		tcpConfig:        configs.tcp,
		tracerouteConfig: configs.traceroute,
		httpConfig:       configs.http,
		labelIDs:         labelIDs,
	}, nil
}

type normalizedCreateCheckConfigs struct {
	ping       *domainping.Config
	tcp        *domaintcp.Config
	traceroute *domaintraceroute.Config
	http       *domainhttp.Config
}

func normalizeCreateCheckConfigs(checkType domaincheck.Type, input CreateCheckInput) (normalizedCreateCheckConfigs, error) {
	if checkType == "" {
		return normalizedCreateCheckConfigs{}, nil
	}
	if checkType == domaincheck.TypeHTTP {
		config, err := normalizeCreateHTTPTypeConfig(input.HTTPConfig, input.PingConfig, input.TCPConfig, input.TracerouteConfig)
		return normalizedCreateCheckConfigs{http: config}, err
	}
	ping, tcp, traceroute, err := normalizeCreateTypeConfig(checkType, input.PingConfig, input.TCPConfig, input.TracerouteConfig, input.HTTPConfig)
	return normalizedCreateCheckConfigs{ping: ping, tcp: tcp, traceroute: traceroute}, err
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
	collections := normalizeUpdateCheckCollections(input, &validation)
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
		pingConfig:       collections.pingConfig,
		tcpConfig:        collections.tcpConfig,
		tracerouteConfig: collections.tracerouteConfig,
		httpConfig:       collections.httpConfig,
		replaceLabels:    collections.replaceLabels,
		labelIDs:         collections.labelIDs,
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

type normalizedUpdateCheckCollections struct {
	pingConfig       updatePingConfigPatch
	tcpConfig        updateTCPConfigPatch
	tracerouteConfig updateTracerouteConfigPatch
	httpConfig       updateHTTPConfigPatch
	labelIDs         []string
	replaceLabels    bool
}

func normalizeUpdateCheckCollections(input UpdateCheckInput, validation *appvalidation.Collector) normalizedUpdateCheckCollections {
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
	httpConfig, err := normalizeUpdateHTTPConfig(input.HTTPConfig)
	if err != nil {
		if !validation.AddValidation(err) {
			validation.AddError("httpConfig", err, input.HTTPConfig)
		}
	}

	labelIDs, replaceLabels, err := normalizeUpdateLabelIDs(input.LabelIDs)
	if err != nil {
		if !validation.AddValidation(err) {
			validation.AddError("labelIds", err, input.LabelIDs)
		}
	}

	return normalizedUpdateCheckCollections{
		pingConfig:       pingConfig,
		tcpConfig:        tcpConfig,
		tracerouteConfig: tracerouteConfig,
		httpConfig:       httpConfig,
		labelIDs:         labelIDs,
		replaceLabels:    replaceLabels,
	}
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

type updateHTTPConfigPatch struct {
	method                *domainhttp.Method
	headers               *[]domainhttp.Header
	body                  *string
	timeoutMs             *int32
	ipFamily              *domainnetwork.IPFamily
	ipFamilySet           bool
	followRedirects       *bool
	skipTLSVerify         *bool
	expectedStatusCodes   *[]int32
	expectedStatusClasses *[]int32
	bodyContains          *string
	bodyContainsSet       bool
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

func (patch updateHTTPConfigPatch) hasChanges() bool {
	return patch.method != nil || patch.headers != nil || patch.body != nil || patch.timeoutMs != nil ||
		patch.ipFamilySet || patch.followRedirects != nil || patch.skipTLSVerify != nil ||
		patch.expectedStatusCodes != nil || patch.expectedStatusClasses != nil || patch.bodyContainsSet
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
	target := strings.TrimSpace(*input)
	if target == "" || len(target) > domainhttp.MaxTargetLength {
		return nil, invalidCheckField("target", "must be between 1 and 2048 characters", input)
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
		input.HTTPConfig.hasChanges() ||
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

func (input *HTTPConfigInput) hasChanges() bool {
	if input == nil {
		return false
	}
	return input.Method != nil || input.Headers != nil || input.Body != nil || input.TimeoutMs != nil ||
		input.IPFamilySet || input.IPFamily != nil || input.FollowRedirects != nil || input.SkipTLSVerify != nil ||
		input.ExpectedStatuses != nil || input.BodyContains != nil
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

func normalizeCreateTypeConfig(checkType domaincheck.Type, pingInput *PingConfigInput, tcpInput *TCPConfigInput, tracerouteInput *TracerouteConfigInput, httpInput *HTTPConfigInput) (*domainping.Config, *domaintcp.Config, *domaintraceroute.Config, error) {
	switch checkType {
	case domaincheck.TypePing:
		config, err := normalizeCreatePingTypeConfig(pingInput, tcpInput, tracerouteInput, httpInput)
		if err != nil {
			return nil, nil, nil, err
		}

		return config, nil, nil, nil
	case domaincheck.TypeTCP:
		config, err := normalizeCreateTCPTypeConfig(pingInput, tcpInput, tracerouteInput, httpInput)
		if err != nil {
			return nil, nil, nil, err
		}

		return nil, config, nil, nil
	case domaincheck.TypeTraceroute:
		config, err := normalizeCreateTracerouteTypeConfig(pingInput, tcpInput, tracerouteInput, httpInput)
		if err != nil {
			return nil, nil, nil, err
		}

		return nil, nil, config, nil
	default:
		return nil, nil, nil, invalidCheckField("type", "unsupported check type", string(checkType))
	}
}

func normalizeCreatePingTypeConfig(pingInput *PingConfigInput, tcpInput *TCPConfigInput, tracerouteInput *TracerouteConfigInput, httpInput *HTTPConfigInput) (*domainping.Config, error) {
	var validation appvalidation.Collector
	if tcpInput != nil {
		validation.Add("tcpConfig", "must be omitted for ping checks", tcpInput)
	}
	if tracerouteInput != nil {
		validation.Add("tracerouteConfig", "must be omitted for ping checks", tracerouteInput)
	}
	if httpInput != nil {
		validation.Add("httpConfig", "must be omitted for ping checks", httpInput)
	}
	config, err := normalizePingConfig(pingInput)
	if err != nil {
		if !validation.AddValidation(err) {
			return nil, err
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return nil, err
	}

	return &config, nil
}

func normalizeCreateTCPTypeConfig(pingInput *PingConfigInput, tcpInput *TCPConfigInput, tracerouteInput *TracerouteConfigInput, httpInput *HTTPConfigInput) (*domaintcp.Config, error) {
	var validation appvalidation.Collector
	if pingInput != nil {
		validation.Add("pingConfig", "must be omitted for tcp checks", pingInput)
	}
	if tracerouteInput != nil {
		validation.Add("tracerouteConfig", "must be omitted for tcp checks", tracerouteInput)
	}
	if httpInput != nil {
		validation.Add("httpConfig", "must be omitted for tcp checks", httpInput)
	}
	config, err := normalizeTCPConfig(tcpInput)
	if err != nil {
		if !validation.AddValidation(err) {
			return nil, err
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return nil, err
	}

	return &config, nil
}

func normalizeCreateTracerouteTypeConfig(pingInput *PingConfigInput, tcpInput *TCPConfigInput, tracerouteInput *TracerouteConfigInput, httpInput *HTTPConfigInput) (*domaintraceroute.Config, error) {
	var validation appvalidation.Collector
	if pingInput != nil {
		validation.Add("pingConfig", "must be omitted for traceroute checks", pingInput)
	}
	if tcpInput != nil {
		validation.Add("tcpConfig", "must be omitted for traceroute checks", tcpInput)
	}
	if httpInput != nil {
		validation.Add("httpConfig", "must be omitted for traceroute checks", httpInput)
	}
	config, err := normalizeTracerouteConfig(tracerouteInput)
	if err != nil {
		if !validation.AddValidation(err) {
			return nil, err
		}
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return nil, err
	}

	return &config, nil
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

func normalizeCreateHTTPTypeConfig(httpInput *HTTPConfigInput, pingInput *PingConfigInput, tcpInput *TCPConfigInput, tracerouteInput *TracerouteConfigInput) (*domainhttp.Config, error) {
	var validation appvalidation.Collector
	if pingInput != nil {
		validation.Add("pingConfig", "must be omitted for http checks", pingInput)
	}
	if tcpInput != nil {
		validation.Add("tcpConfig", "must be omitted for http checks", tcpInput)
	}
	if tracerouteInput != nil {
		validation.Add("tracerouteConfig", "must be omitted for http checks", tracerouteInput)
	}
	config, err := normalizeHTTPConfig(httpInput)
	if err != nil && !validation.AddValidation(err) {
		return nil, err
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return nil, err
	}
	return &config, nil
}

func normalizeHTTPConfig(input *HTTPConfigInput) (domainhttp.Config, error) {
	config := domainhttp.DefaultConfig()
	if input == nil {
		return config, nil
	}
	patch, err := normalizeUpdateHTTPConfig(input)
	if err != nil {
		return domainhttp.Config{}, err
	}
	applyHTTPConfigPatch(&config, patch)
	if _, err := domainhttp.VNBody(config.Method, config.Body); err != nil {
		return domainhttp.Config{}, invalidCheckField("httpConfig.body", err.Error(), input.Body)
	}
	return config, nil
}

func normalizeUpdateHTTPConfig(input *HTTPConfigInput) (updateHTTPConfigPatch, error) {
	if input == nil {
		return updateHTTPConfigPatch{}, nil
	}
	var validation appvalidation.Collector
	patch := updateHTTPConfigPatch{
		method:          normalizeHTTPMethodPatch(input.Method, &validation),
		headers:         normalizeHTTPHeadersPatch(input.Headers, &validation),
		body:            normalizeHTTPBodyPatch(input.Body, &validation),
		timeoutMs:       normalizeHTTPTimeoutPatch(input.TimeoutMs, &validation),
		followRedirects: input.FollowRedirects,
		skipTLSVerify:   input.SkipTLSVerify,
	}
	patch.ipFamily, patch.ipFamilySet = normalizeHTTPIPFamilyPatch(input.IPFamily, input.IPFamilySet, &validation)
	patch.expectedStatusCodes, patch.expectedStatusClasses = normalizeHTTPExpectedStatusesPatch(input.ExpectedStatuses, &validation)
	patch.bodyContains, patch.bodyContainsSet = normalizeHTTPBodyContainsPatch(input.BodyContains, &validation)
	if err := validation.Err(ErrInvalidInput); err != nil {
		return updateHTTPConfigPatch{}, err
	}
	return patch, nil
}

func normalizeHTTPMethodPatch(input *string, validation *appvalidation.Collector) *domainhttp.Method {
	if input == nil {
		return nil
	}
	method, err := domainhttp.VNMethod(domainhttp.Method(*input))
	if err != nil {
		validation.AddError("httpConfig.method", err, input)
		return nil
	}
	return &method
}

func normalizeHTTPHeadersPatch(input *[]domainhttp.Header, validation *appvalidation.Collector) *[]domainhttp.Header {
	if input == nil {
		return nil
	}
	headers, err := domainhttp.VNHeaders(*input)
	if err != nil {
		validation.AddError("httpConfig.headers", err, input)
		return nil
	}
	return &headers
}

func normalizeHTTPBodyPatch(input *string, validation *appvalidation.Collector) *string {
	if input == nil {
		return nil
	}
	if len([]byte(*input)) > domainhttp.MaxBodyBytes {
		validation.Add("httpConfig.body", fmt.Sprintf("must be at most %d bytes", domainhttp.MaxBodyBytes), input)
		return nil
	}
	body := *input
	return &body
}

func normalizeHTTPTimeoutPatch(input *int32, validation *appvalidation.Collector) *int32 {
	if input == nil {
		return nil
	}
	timeout, err := domainhttp.VNTimeoutMs(*input)
	if err != nil {
		validation.AddError("httpConfig.timeoutMs", err, input)
		return nil
	}
	return &timeout
}

func normalizeHTTPIPFamilyPatch(input *string, fieldSet bool, validation *appvalidation.Collector) (*domainnetwork.IPFamily, bool) {
	if !fieldSet && input == nil {
		return nil, false
	}
	if input == nil {
		return nil, true
	}
	family, err := domainnetwork.ParseOptionalIPFamily(input)
	if err != nil {
		validation.Add("httpConfig.ipFamily", `must be "inet" or "inet6"`, input)
		return nil, true
	}
	family, err = domainhttp.VNIPFamily(family)
	if err != nil {
		validation.AddError("httpConfig.ipFamily", err, input)
		return nil, true
	}
	return family, true
}

func normalizeHTTPExpectedStatusesPatch(input *[]HTTPStatusSelectorInput, validation *appvalidation.Collector) (*[]int32, *[]int32) {
	if input == nil {
		return nil, nil
	}
	codes, classes, err := normalizeExpectedStatuses(*input)
	if err != nil {
		validation.AddError("httpConfig.expectedStatuses", err, input)
		return nil, nil
	}
	return &codes, &classes
}

func normalizeHTTPBodyContainsPatch(input *string, validation *appvalidation.Collector) (*string, bool) {
	if input == nil {
		return nil, false
	}
	if *input == "" {
		return nil, true
	}
	value, err := domainhttp.VNBodyContains(input)
	if err != nil {
		validation.AddError("httpConfig.bodyContains", err, input)
		return nil, true
	}
	return value, true
}

func normalizeExpectedStatuses(selectors []HTTPStatusSelectorInput) ([]int32, []int32, error) {
	codes := make([]int32, 0, len(selectors))
	classes := make([]int32, 0, len(selectors))
	for _, selector := range selectors {
		switch selector.Kind {
		case "code":
			if selector.Code == nil || selector.Class != nil {
				return nil, nil, errors.New("code selector must include only code")
			}
			codes = append(codes, *selector.Code)
		case "class":
			if selector.Class == nil || selector.Code != nil || len(*selector.Class) != 3 || !strings.HasSuffix(*selector.Class, "xx") {
				return nil, nil, errors.New("class selector must include a status class")
			}
			classes = append(classes, int32((*selector.Class)[0]-'0'))
		default:
			return nil, nil, errors.New("selector kind must be code or class")
		}
	}
	return domainhttp.VNExpectedStatuses(codes, classes)
}

func applyHTTPConfigPatch(config *domainhttp.Config, patch updateHTTPConfigPatch) {
	if patch.method != nil {
		config.Method = *patch.method
		if config.Method == domainhttp.MethodGet || config.Method == domainhttp.MethodHead {
			config.Body = nil
		}
	}
	if patch.headers != nil {
		config.Headers = append([]domainhttp.Header(nil), (*patch.headers)...)
	}
	if patch.body != nil {
		value := *patch.body
		config.Body = &value
	}
	if patch.timeoutMs != nil {
		config.TimeoutMs = *patch.timeoutMs
	}
	if patch.ipFamilySet {
		config.IPFamily = patch.ipFamily
	}
	if patch.followRedirects != nil {
		config.FollowRedirects = *patch.followRedirects
	}
	if patch.skipTLSVerify != nil {
		config.SkipTLSVerify = *patch.skipTLSVerify
	}
	if patch.expectedStatusCodes != nil {
		config.ExpectedStatusCodes = append([]int32(nil), (*patch.expectedStatusCodes)...)
	}
	if patch.expectedStatusClasses != nil {
		config.ExpectedStatusClasses = append([]int32(nil), (*patch.expectedStatusClasses)...)
	}
	if patch.bodyContainsSet {
		config.BodyContains = nil
		if patch.bodyContains != nil {
			value := *patch.bodyContains
			config.BodyContains = &value
		}
	}
}
