package runtime

import "errors"

var (
	ErrAuthFailed          = errors.New("probe runtime authentication failed")
	ErrPermanentRuntimeAPI = errors.New("probe runtime api permanent failure")
	ErrVersionUnsupported  = errors.New("probe agent version unsupported")
)
