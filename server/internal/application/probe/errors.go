package probe

import (
	"errors"

	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

var (
	ErrProjectNotFound = domainproject.ErrProjectNotFound
	ErrLabelNotFound   = domainprobe.ErrLabelNotFound
	ErrInvalidInput    = errors.New("probe input invalid")
)
