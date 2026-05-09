package proberuntime

import (
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

var (
	ErrInvalidInput      = domainprobe.ErrInvalidInput
	ErrProbeNotFound     = domainprobe.ErrProbeNotFound
	ErrProbeDisabled     = domainprobe.ErrProbeDisabled
	ErrInvalidCredential = domainprobe.ErrInvalidCredential
	ErrInvalidResult     = domainping.ErrInvalidResult
)
