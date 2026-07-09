package problem

import (
	"errors"

	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	domainalert "github.com/yorukot/netstamp/internal/domain/alert"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	"github.com/yorukot/netstamp/internal/domain/identity"
	"github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
)

func NotFound(err error) (error, bool) {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound):
		return httpx.NotFound("project not found"), true
	case errors.Is(err, domainproject.ErrMemberNotFound):
		return httpx.NotFound("project member not found"), true
	case errors.Is(err, domainproject.ErrInviteNotFound):
		return httpx.NotFound("project invite not found"), true
	case errors.Is(err, identity.ErrUserNotFound):
		return httpx.NotFound("user not found"), true
	case errors.Is(err, label.ErrLabelNotFound):
		return httpx.NotFound("label not found"), true
	case errors.Is(err, domaincheck.ErrCheckNotFound):
		return httpx.NotFound("check not found"), true
	case errors.Is(err, domainprobe.ErrProbeNotFound):
		return httpx.NotFound("probe not found"), true
	case errors.Is(err, domainalert.ErrRuleNotFound):
		return httpx.NotFound("alert rule not found"), true
	case errors.Is(err, domainalert.ErrIncidentNotFound):
		return httpx.NotFound("alert incident not found"), true
	case errors.Is(err, domainalert.ErrNotificationNotFound):
		return httpx.NotFound("alert notification not found"), true
	case errors.Is(err, domainpublic.ErrPageNotFound):
		return httpx.NotFound("public status page not found"), true
	case errors.Is(err, domainpublic.ErrElementNotFound):
		return httpx.NotFound("public status page element not found"), true
	default:
		return nil, false
	}
}
