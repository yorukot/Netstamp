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

func NotFound(err error) (bool, error) {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound):
		return true, httpx.NotFound("project not found")
	case errors.Is(err, domainproject.ErrMemberNotFound):
		return true, httpx.NotFound("project member not found")
	case errors.Is(err, domainproject.ErrInviteNotFound):
		return true, httpx.NotFound("project invite not found")
	case errors.Is(err, identity.ErrUserNotFound):
		return true, httpx.NotFound("user not found")
	case errors.Is(err, label.ErrLabelNotFound):
		return true, httpx.NotFound("label not found")
	case errors.Is(err, domaincheck.ErrCheckNotFound):
		return true, httpx.NotFound("check not found")
	case errors.Is(err, domainprobe.ErrProbeNotFound):
		return true, httpx.NotFound("probe not found")
	case errors.Is(err, domainalert.ErrRuleNotFound):
		return true, httpx.NotFound("alert rule not found")
	case errors.Is(err, domainalert.ErrIncidentNotFound):
		return true, httpx.NotFound("alert incident not found")
	case errors.Is(err, domainalert.ErrNotificationNotFound):
		return true, httpx.NotFound("alert notification not found")
	case errors.Is(err, domainpublic.ErrPageNotFound):
		return true, httpx.NotFound("public status page not found")
	case errors.Is(err, domainpublic.ErrElementNotFound):
		return true, httpx.NotFound("public status page element not found")
	default:
		return false, nil
	}
}
