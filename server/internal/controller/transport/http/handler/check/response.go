package check

import (
	"encoding/json"
	"time"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domaintcp "github.com/yorukot/netstamp/internal/domain/tcp"
	domaintraceroute "github.com/yorukot/netstamp/internal/domain/traceroute"
)

type checkOutput struct {
	Body checkOutputBody
}

type checkOutputBody struct {
	Check           checkBody `json:"check"`
	CanManageChecks bool      `json:"canManageChecks"`
}

type checkBody struct {
	ID               string                   `json:"id"`
	ProjectID        string                   `json:"projectId"`
	Name             string                   `json:"name"`
	Type             domaincheck.Type         `json:"type"`
	Target           string                   `json:"target"`
	Selector         json.RawMessage          `json:"selector,omitempty"`
	Description      *string                  `json:"description,omitempty"`
	IntervalSeconds  int32                    `json:"intervalSeconds"`
	Labels           []domainlabel.Label      `json:"labels"`
	CreatedAt        time.Time                `json:"createdAt"`
	UpdatedAt        time.Time                `json:"updatedAt"`
	PingConfig       *domainping.Config       `json:"pingConfig,omitempty"`
	TCPConfig        *domaintcp.Config        `json:"tcpConfig,omitempty"`
	TracerouteConfig *domaintraceroute.Config `json:"tracerouteConfig,omitempty"`
	HTTPConfig       *checkHTTPConfigBody     `json:"httpConfig,omitempty"`
}

type checkHTTPConfigBody struct {
	Method                  domainhttp.Method           `json:"method"`
	Headers                 []domainhttp.Header         `json:"headers"`
	Body                    *string                     `json:"body,omitempty"`
	TimeoutMs               int32                       `json:"timeoutMs"`
	IPFamily                *domainnetwork.IPFamily     `json:"ipFamily,omitempty"`
	FollowRedirects         bool                        `json:"followRedirects"`
	SkipTLSVerify           bool                        `json:"skipTlsVerify"`
	ExpectedStatuses        []domainhttp.StatusSelector `json:"expectedStatuses"`
	BodyContains            *string                     `json:"bodyContains,omitempty"`
	SensitiveFieldsRedacted bool                        `json:"sensitiveFieldsRedacted"`
}

func newCheckOutput(check domaincheck.Check, canManageChecks bool) *checkOutput {
	return &checkOutput{Body: checkOutputBody{
		Check:           newCheckBody(check, canManageChecks),
		CanManageChecks: canManageChecks,
	}}
}

func newCheckBody(check domaincheck.Check, includeSensitiveHTTPConfig bool) checkBody {
	target := check.Target
	if check.Type == domaincheck.TypeHTTP && !includeSensitiveHTTPConfig {
		target = domainhttp.RedactTarget(target)
	}

	return checkBody{
		ID:               check.ID,
		ProjectID:        check.ProjectID,
		Name:             check.Name,
		Type:             check.Type,
		Target:           target,
		Selector:         append(json.RawMessage(nil), check.Selector...),
		Description:      check.Description,
		IntervalSeconds:  check.IntervalSeconds,
		Labels:           check.Labels,
		CreatedAt:        check.CreatedAt,
		UpdatedAt:        check.UpdatedAt,
		PingConfig:       check.PingConfig,
		TCPConfig:        check.TCPConfig,
		TracerouteConfig: check.TracerouteConfig,
		HTTPConfig:       newCheckHTTPConfigBody(check.HTTPConfig, includeSensitiveHTTPConfig),
	}
}

func newCheckHTTPConfigBody(config *domainhttp.Config, includeSensitive bool) *checkHTTPConfigBody {
	if config == nil {
		return nil
	}

	headers := append([]domainhttp.Header(nil), config.Headers...)
	body := config.Body
	bodyContains := config.BodyContains
	if !includeSensitive {
		headers = []domainhttp.Header{}
		body = nil
		bodyContains = nil
	}

	return &checkHTTPConfigBody{
		Method:                  config.Method,
		Headers:                 headers,
		Body:                    body,
		TimeoutMs:               config.TimeoutMs,
		IPFamily:                config.IPFamily,
		FollowRedirects:         config.FollowRedirects,
		SkipTLSVerify:           config.SkipTLSVerify,
		ExpectedStatuses:        config.StatusSelectors(),
		BodyContains:            bodyContains,
		SensitiveFieldsRedacted: !includeSensitive,
	}
}
