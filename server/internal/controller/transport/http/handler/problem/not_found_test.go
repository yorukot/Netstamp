package problem

import (
	"errors"
	"net/http"
	"testing"

	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	"github.com/yorukot/netstamp/internal/domain/identity"
	"github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
)

func TestNotFoundMapsDomainErrorsToSpecificDetails(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		detail string
	}{
		{name: "project", err: domainproject.ErrProjectNotFound, detail: "project not found"},
		{name: "project member", err: domainproject.ErrMemberNotFound, detail: "project member not found"},
		{name: "project invite", err: domainproject.ErrInviteNotFound, detail: "project invite not found"},
		{name: "user", err: identity.ErrUserNotFound, detail: "user not found"},
		{name: "label", err: label.ErrLabelNotFound, detail: "label not found"},
		{name: "check", err: domaincheck.ErrCheckNotFound, detail: "check not found"},
		{name: "probe", err: domainprobe.ErrProbeNotFound, detail: "probe not found"},
		{name: "alert rule", err: domainalert.ErrRuleNotFound, detail: "alert rule not found"},
		{name: "alert incident", err: domainalert.ErrIncidentNotFound, detail: "alert incident not found"},
		{name: "alert notification", err: domainalert.ErrNotificationNotFound, detail: "alert notification not found"},
		{name: "public status page", err: domainpublic.ErrPageNotFound, detail: "public status page not found"},
		{name: "public status element", err: domainpublic.ErrElementNotFound, detail: "public status page element not found"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ok, err := NotFound(test.err)
			if !ok {
				t.Fatal("expected not-found mapping")
			}

			var httpErr *httpx.Error
			if !errors.As(err, &httpErr) {
				t.Fatalf("expected http error, got %T", err)
			}
			if httpErr.Status != http.StatusNotFound {
				t.Fatalf("status = %d, want %d", httpErr.Status, http.StatusNotFound)
			}
			if httpErr.Detail != test.detail {
				t.Fatalf("detail = %q, want %q", httpErr.Detail, test.detail)
			}
		})
	}
}

func TestNotFoundIgnoresUnmappedErrors(t *testing.T) {
	ok, err := NotFound(errors.New("other failure"))
	if ok {
		t.Fatalf("expected no mapping, got %v", err)
	}
}
