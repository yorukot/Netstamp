package agentruntime

import (
	"net"
	"net/netip"

	"github.com/yorukot/netstamp/internal/agent/infrastructure/httpclient"
)

func agentStatus() httpclient.HeartbeatInput {
	agentVersion := AgentString
	return httpclient.HeartbeatInput{
		AgentVersion: &agentVersion,
		Addrs:        localAddrs(),
	}
}

func localAddrs() []netip.Addr {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	addrs := make([]netip.Addr, 0)
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		ifAddrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, ifAddr := range ifAddrs {
			addr, ok := prefixAddr(ifAddr)
			if !ok || !addr.IsValid() || addr.IsLoopback() || addr.IsMulticast() || addr.IsUnspecified() {
				continue
			}
			addrs = append(addrs, addr.Unmap())
		}
	}

	return addrs
}

func prefixAddr(value net.Addr) (netip.Addr, bool) {
	prefix, err := netip.ParsePrefix(value.String())
	if err == nil {
		return prefix.Addr(), true
	}
	addr, err := netip.ParseAddr(value.String())
	if err != nil {
		return netip.Addr{}, false
	}

	return addr, true
}
