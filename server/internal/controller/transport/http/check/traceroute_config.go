package check

import appcheck "github.com/yorukot/netstamp/internal/controller/application/check"

type checkTracerouteConfigInput struct {
	Protocol        *string `json:"protocol,omitempty"`
	MaxHops         *int32  `json:"maxHops,omitempty"`
	TimeoutMs       *int32  `json:"timeoutMs,omitempty"`
	QueriesPerHop   *int32  `json:"queriesPerHop,omitempty"`
	PacketSizeBytes *int32  `json:"packetSizeBytes,omitempty"`
	Port            *int32  `json:"port,omitempty"`
	IPFamily        *string `json:"ipFamily,omitempty"`
}

func (config *checkTracerouteConfigInput) appInput() *appcheck.TracerouteConfigInput {
	if config == nil {
		return nil
	}

	return &appcheck.TracerouteConfigInput{
		Protocol:        config.Protocol,
		MaxHops:         config.MaxHops,
		TimeoutMs:       config.TimeoutMs,
		QueriesPerHop:   config.QueriesPerHop,
		PacketSizeBytes: config.PacketSizeBytes,
		Port:            config.Port,
		IPFamily:        config.IPFamily,
	}
}
