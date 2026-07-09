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
		return true, httpx.NotFoundCode(httpx.CodeProjectNotFound, "project not found")
	case errors.Is(err, domainproject.ErrMemberNotFound):
		return true, httpx.NotFoundCode(httpx.CodeProjectMemberNotFound, "project member not found")
	case errors.Is(err, domainproject.ErrInviteNotFound):
		return true, httpx.NotFoundCode(httpx.CodeProjectInviteNotFound, "project invite not found")
	case errors.Is(err, identity.ErrUserNotFound):
		return true, httpx.NotFoundCode(httpx.CodeUserNotFound, "user not found")
	case errors.Is(err, label.ErrLabelNotFound):
		return true, httpx.NotFoundCode(httpx.CodeLabelNotFound, "label not found")
	case errors.Is(err, domaincheck.ErrCheckNotFound):
		return true, httpx.NotFoundCode(httpx.CodeCheckNotFound, "check not found")
	case errors.Is(err, domainprobe.ErrProbeNotFound):
		return true, httpx.NotFoundCode(httpx.CodeProbeNotFound, "probe not found")
	case errors.Is(err, domainalert.ErrRuleNotFound):
		return true, httpx.NotFoundCode(httpx.CodeAlertRuleNotFound, "alert rule not found")
	case errors.Is(err, domainalert.ErrIncidentNotFound):
		return true, httpx.NotFoundCode(httpx.CodeAlertIncidentNotFound, "alert incident not found")
	case errors.Is(err, domainalert.ErrNotificationNotFound):
		return true, httpx.NotFoundCode(httpx.CodeAlertNotificationNotFound, "alert notification not found")
	case errors.Is(err, domainpublic.ErrPageNotFound):
		return true, httpx.NotFoundCode(httpx.CodePublicStatusPageNotFound, "public status page not found")
	case errors.Is(err, domainpublic.ErrElementNotFound):
		return true, httpx.NotFoundCode(httpx.CodePublicStatusElementNotFound, "public status page element not found")
	default:
		return false, nil
	}
}
