package check

import appcheck "github.com/yorukot/netstamp/internal/controller/application/check"

type checkPingConfigInput struct {
	PacketCount     *int32  `json:"packetCount,omitempty" minimum:"1" doc:"ICMP packet count." example:"4"`
	PacketSizeBytes *int32  `json:"packetSizeBytes,omitempty" minimum:"1" maximum:"65507" doc:"ICMP payload size in bytes." example:"56"`
	TimeoutMs       *int32  `json:"timeoutMs,omitempty" minimum:"1" doc:"Ping timeout in milliseconds." example:"3000"`
	IPFamily        *string `json:"ipFamily,omitempty" enum:"inet,inet6" doc:"Optional IP family preference." example:"inet"`
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
