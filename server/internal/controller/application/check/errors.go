package check

import (
	"errors"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	"github.com/yorukot/netstamp/internal/domain/identity"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

var (
	ErrCheckNotFound   = domaincheck.ErrCheckNotFound
	ErrInvalidInput    = domaincheck.ErrInvalidInput
	ErrProjectNotFound = domainproject.ErrProjectNotFound
	ErrLabelNotFound   = domainlabel.ErrLabelNotFound
	ErrUserNotFound    = identity.ErrUserNotFound
	ErrForbidden       = errors.New("check action forbidden")
)
