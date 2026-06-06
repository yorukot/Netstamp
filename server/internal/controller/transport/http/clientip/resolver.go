package clientip

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
)

type Resolver struct {
	trustedProxies []netip.Prefix
}

func NewResolver(trustedProxies []netip.Prefix) *Resolver {
	return &Resolver{trustedProxies: append([]netip.Prefix(nil), trustedProxies...)}
}

func (r *Resolver) PublicIP(req *http.Request) (*netip.Addr, bool) {
	if req == nil {
		return nil, false
	}

	remote, ok := parseRemoteAddr(req.RemoteAddr)
	if !ok {
		return nil, false
	}

	var addr netip.Addr
	if r.trusts(remote) {
		addr, ok = r.forwardedIP(req)
		if !ok {
			addr = remote
		}
	} else {
		addr = remote
	}

	addr = addr.Unmap()
	if !isPublicAddr(addr) {
		return nil, false
	}

	return &addr, true
}

func (r *Resolver) trusts(addr netip.Addr) bool {
	addr = addr.Unmap()
	for _, prefix := range r.trustedProxies {
		if prefix.Contains(addr) {
			return true
		}
	}

	return false
}

func (r *Resolver) forwardedIP(req *http.Request) (netip.Addr, bool) {
	for _, raw := range strings.Split(req.Header.Get("X-Forwarded-For"), ",") {
		addr, ok := parseHeaderAddr(raw)
		if !ok || !isPublicAddr(addr) || r.trusts(addr) {
			continue
		}

		return addr, true
	}

	return parseHeaderAddr(req.Header.Get("X-Real-IP"))
}

func parseRemoteAddr(value string) (netip.Addr, bool) {
	host, _, err := net.SplitHostPort(strings.TrimSpace(value))
	if err == nil {
		return parseHeaderAddr(host)
	}

	return parseHeaderAddr(value)
}

func parseHeaderAddr(value string) (netip.Addr, bool) {
	addr, err := netip.ParseAddr(strings.TrimSpace(value))
	if err != nil {
		return netip.Addr{}, false
	}

	addr = addr.Unmap()
	return addr, addr.IsValid()
}

func isPublicAddr(addr netip.Addr) bool {
	return addr.IsValid() &&
		!addr.IsUnspecified() &&
		!addr.IsLoopback() &&
		!addr.IsPrivate() &&
		!addr.IsMulticast() &&
		!addr.IsLinkLocalUnicast()
}
