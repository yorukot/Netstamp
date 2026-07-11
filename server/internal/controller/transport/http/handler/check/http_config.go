package check

import (
	"bytes"
	"encoding/json"

	appcheck "github.com/yorukot/netstamp/internal/controller/application/check"
	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
)

type checkHTTPStatusSelectorInput struct {
	Kind  string  `json:"kind"`
	Code  *int32  `json:"code,omitempty"`
	Class *string `json:"class,omitempty"`
}

type checkHTTPConfigInput struct {
	Method           *string                         `json:"method,omitempty"`
	Headers          *[]domainhttp.Header            `json:"headers,omitempty"`
	Body             *string                         `json:"body,omitempty"`
	TimeoutMs        *int32                          `json:"timeoutMs,omitempty"`
	IPFamily         nullableStringInput             `json:"ipFamily"`
	FollowRedirects  *bool                           `json:"followRedirects,omitempty"`
	SkipTLSVerify    *bool                           `json:"skipTlsVerify,omitempty"`
	ExpectedStatuses *[]checkHTTPStatusSelectorInput `json:"expectedStatuses,omitempty"`
	BodyContains     *string                         `json:"bodyContains,omitempty"`
}

type nullableStringInput struct {
	Value   *string
	Present bool
}

func (input *nullableStringInput) UnmarshalJSON(data []byte) error {
	input.Present = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		input.Value = nil
		return nil
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	input.Value = &value
	return nil
}

func (config *checkHTTPConfigInput) appInput() *appcheck.HTTPConfigInput {
	if config == nil {
		return nil
	}
	var selectors *[]appcheck.HTTPStatusSelectorInput
	if config.ExpectedStatuses != nil {
		values := make([]appcheck.HTTPStatusSelectorInput, len(*config.ExpectedStatuses))
		for i, value := range *config.ExpectedStatuses {
			values[i] = appcheck.HTTPStatusSelectorInput{Kind: value.Kind, Code: value.Code, Class: value.Class}
		}
		selectors = &values
	}
	return &appcheck.HTTPConfigInput{
		Method: config.Method, Headers: config.Headers, Body: config.Body, TimeoutMs: config.TimeoutMs,
		IPFamily: config.IPFamily.Value, IPFamilySet: config.IPFamily.Present, FollowRedirects: config.FollowRedirects,
		SkipTLSVerify: config.SkipTLSVerify, ExpectedStatuses: selectors, BodyContains: config.BodyContains,
	}
}
