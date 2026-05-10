package proberegistry

import (
	"errors"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

var (
	ErrProjectNotFound = domainproject.ErrProjectNotFound
	ErrLabelNotFound   = domainlabel.ErrLabelNotFound
	ErrProbeNotFound   = domainprobe.ErrProbeNotFound
	ErrInvalidInput    = domainprobe.ErrInvalidInput
	ErrForbidden       = errors.New("probe action forbidden")
)
