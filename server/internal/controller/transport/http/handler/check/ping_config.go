package check

import appcheck "github.com/yorukot/netstamp/internal/controller/application/check"

type checkPingConfigInput struct {
	PacketCount     *int32  `json:"packetCount,omitempty"`
	PacketSizeBytes *int32  `json:"packetSizeBytes,omitempty"`
	TimeoutMs       *int32  `json:"timeoutMs,omitempty"`
	IPFamily        *string `json:"ipFamily,omitempty"`
}

func (config *checkPingConfigInput) appInput() *appcheck.PingConfigInput {
	if config == nil {
		return nil
	}

	return &appcheck.PingConfigInput{
		PacketCount:     config.PacketCount,
		PacketSizeBytes: config.PacketSizeBytes,
		TimeoutMs:       config.TimeoutMs,
		IPFamily:        config.IPFamily,
	}
}
