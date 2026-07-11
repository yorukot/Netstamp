package check

import (
	"testing"

	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
)

func TestCheckHashIncludesHTTPConfig(t *testing.T) {
	body := "request body"
	contains := "healthy"
	inet := domainnetwork.IPFamilyInet
	config := domainhttp.DefaultConfig()
	config.Headers = []domainhttp.Header{{Name: "Authorization", Value: "Bearer secret"}}
	config.Body = &body
	config.IPFamily = &inet
	config.BodyContains = &contains
	base := Check{Type: TypeHTTP, Target: "https://example.com/health", IntervalSeconds: 30, HTTPConfig: &config}
	baseHash := base.Hash()

	tests := []struct {
		name   string
		mutate func(*domainhttp.Config)
	}{
		{name: "method", mutate: func(config *domainhttp.Config) { config.Method = domainhttp.MethodPut }},
		{name: "headers", mutate: func(config *domainhttp.Config) { config.Headers[0].Value = "Bearer changed" }},
		{name: "body", mutate: func(config *domainhttp.Config) { value := "changed"; config.Body = &value }},
		{name: "timeout", mutate: func(config *domainhttp.Config) { config.TimeoutMs++ }},
		{name: "ip family", mutate: func(config *domainhttp.Config) { value := domainnetwork.IPFamilyInet6; config.IPFamily = &value }},
		{name: "follow redirects", mutate: func(config *domainhttp.Config) { config.FollowRedirects = !config.FollowRedirects }},
		{name: "skip tls verify", mutate: func(config *domainhttp.Config) { config.SkipTLSVerify = !config.SkipTLSVerify }},
		{name: "status codes", mutate: func(config *domainhttp.Config) { config.ExpectedStatusCodes = []int32{204} }},
		{name: "status classes", mutate: func(config *domainhttp.Config) { config.ExpectedStatusClasses = []int32{4} }},
		{name: "body assertion", mutate: func(config *domainhttp.Config) { value := "changed"; config.BodyContains = &value }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			changedConfig := config
			changedConfig.Headers = append([]domainhttp.Header(nil), config.Headers...)
			changedConfig.ExpectedStatusCodes = append([]int32(nil), config.ExpectedStatusCodes...)
			changedConfig.ExpectedStatusClasses = append([]int32(nil), config.ExpectedStatusClasses...)
			test.mutate(&changedConfig)
			changed := base
			changed.HTTPConfig = &changedConfig
			if changed.Hash() == baseHash {
				t.Fatal("expected HTTP config change to update check hash")
			}
		})
	}
}
