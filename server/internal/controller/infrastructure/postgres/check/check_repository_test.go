package pgcheck

import (
	"testing"

	"github.com/google/uuid"

	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
)

func TestNewHTTPCheckConfigParamsPreservesEmptyPostgresArrays(t *testing.T) {
	tests := []struct {
		name   string
		config domainhttp.Config
	}{
		{
			name:   "default status classes without exact codes",
			config: domainhttp.DefaultConfig(),
		},
		{
			name: "exact status code without classes",
			config: domainhttp.Config{
				Method:              domainhttp.MethodGet,
				TimeoutMs:           domainhttp.DefaultTimeoutMs,
				FollowRedirects:     true,
				ExpectedStatusCodes: []int32{204},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			params, err := newHTTPCheckConfigParams(uuid.New(), test.config)
			if err != nil {
				t.Fatalf("build HTTP check config params: %v", err)
			}
			if params.ExpectedStatusCodes == nil {
				t.Fatal("expected status codes to encode as an empty PostgreSQL array, not NULL")
			}
			if params.ExpectedStatusClasses == nil {
				t.Fatal("expected status classes to encode as an empty PostgreSQL array, not NULL")
			}
			if string(params.Headers) != "[]" {
				t.Fatalf("expected empty headers JSON array, got %s", params.Headers)
			}
		})
	}
}
