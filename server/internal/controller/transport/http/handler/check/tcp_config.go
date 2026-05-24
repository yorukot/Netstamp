package check

import appcheck "github.com/yorukot/netstamp/internal/controller/application/check"

type checkTCPConfigInput struct {
	Port      *int32  `json:"port,omitempty"`
	TimeoutMs *int32  `json:"timeoutMs,omitempty"`
	IPFamily  *string `json:"ipFamily,omitempty"`
}

func (config *checkTCPConfigInput) appInput() *appcheck.TCPConfigInput {
	if config == nil {
		return nil
	}

	return &appcheck.TCPConfigInput{
		Port:      config.Port,
		TimeoutMs: config.TimeoutMs,
		IPFamily:  config.IPFamily,
	}
}
