package probe

import (
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

var (
	ErrProjectNotFound = domainproject.ErrProjectNotFound
	ErrLabelNotFound   = domainlabel.ErrLabelNotFound
	ErrInvalidInput    = domainprobe.ErrInvalidInput
)
