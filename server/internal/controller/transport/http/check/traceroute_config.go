package check

import appcheck "github.com/yorukot/netstamp/internal/controller/application/check"

type checkTracerouteConfigInput struct {
	Protocol        *string `json:"protocol,omitempty" enum:"icmp,udp" doc:"Traceroute protocol." example:"icmp"`
	MaxHops         *int32  `json:"maxHops,omitempty" minimum:"1" maximum:"64" doc:"Maximum hop count." example:"30"`
	TimeoutMs       *int32  `json:"timeoutMs,omitempty" minimum:"1" maximum:"60000" doc:"Traceroute timeout in milliseconds." example:"3000"`
	QueriesPerHop   *int32  `json:"queriesPerHop,omitempty" minimum:"1" maximum:"10" doc:"Probe attempts per hop." example:"3"`
	PacketSizeBytes *int32  `json:"packetSizeBytes,omitempty" minimum:"1" maximum:"65507" doc:"Probe payload size in bytes." example:"56"`
	Port            *int32  `json:"port,omitempty" minimum:"1" maximum:"65535" doc:"Base destination port for UDP traceroute." example:"33434"`
	IPFamily        *string `json:"ipFamily,omitempty" enum:"inet,inet6" doc:"Optional IP family preference." example:"inet"`
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
