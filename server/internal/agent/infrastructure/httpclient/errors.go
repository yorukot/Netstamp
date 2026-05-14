package httpclient

import "errors"

var (
	ErrAuthFailed          = errors.New("probe runtime authentication failed")
	ErrPermanentRuntimeAPI = errors.New("probe runtime api permanent failure")
)
