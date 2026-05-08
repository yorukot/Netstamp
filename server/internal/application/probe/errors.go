package probe

import (
	"errors"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

var (
	ErrProjectNotFound = domainproject.ErrProjectNotFound
	ErrLabelNotFound   = domainlabel.ErrLabelNotFound
	ErrInvalidInput    = domainprobe.ErrInvalidInput
	ErrForbidden       = errors.New("probe action forbidden")
)
