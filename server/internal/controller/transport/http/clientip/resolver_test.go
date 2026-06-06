package clientip

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
)

func TestResolverUsesDirectRemotePublicIP(t *testing.T) {
	req := requestWithRemote("203.0.113.10:44321")
	got, ok := NewResolver(nil).PublicIP(req)
	if !ok || got == nil || got.String() != "203.0.113.10" {
		t.Fatalf("expected direct remote IP, got %#v ok=%v", got, ok)
	}
}

func TestResolverIgnoresForwardedHeadersFromUntrustedRemote(t *testing.T) {
	req := requestWithRemote("203.0.113.10:44321")
	req.Header.Set("X-Forwarded-For", "198.51.100.20")

	got, ok := NewResolver(nil).PublicIP(req)
	if !ok || got == nil || got.String() != "203.0.113.10" {
		t.Fatalf("expected untrusted forwarded header to be ignored, got %#v ok=%v", got, ok)
	}
}

func TestResolverUsesForwardedIPFromTrustedProxy(t *testing.T) {
	req := requestWithRemote("10.0.0.5:44321")
	req.Header.Set("X-Forwarded-For", "198.51.100.20, 10.0.0.5")

	resolver := NewResolver([]netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")})
	got, ok := resolver.PublicIP(req)
	if !ok || got == nil || got.String() != "198.51.100.20" {
		t.Fatalf("expected forwarded public IP, got %#v ok=%v", got, ok)
	}
}

func TestResolverFallsBackToXRealIPForTrustedProxy(t *testing.T) {
	req := requestWithRemote("10.0.0.5:44321")
	req.Header.Set("X-Real-IP", "2001:db8::10")

	resolver := NewResolver([]netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")})
	got, ok := resolver.PublicIP(req)
	if !ok || got == nil || got.String() != "2001:db8::10" {
		t.Fatalf("expected x-real-ip public IP, got %#v ok=%v", got, ok)
	}
}

func TestResolverRejectsPrivateSourceIP(t *testing.T) {
	req := requestWithRemote("10.0.0.5:44321")
	got, ok := NewResolver(nil).PublicIP(req)
	if ok || got != nil {
		t.Fatalf("expected private remote to be ignored, got %#v ok=%v", got, ok)
	}
}

func TestMiddlewareStoresForwardedIPFromTrustedProxy(t *testing.T) {
	req := requestWithRemote("10.0.0.5:44321")
	req.Header.Set("X-Forwarded-For", "198.51.100.20, 10.0.0.5")

	var got netip.Addr
	var ok bool
	handler := Middleware([]netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")})(http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		got, ok = FromContext(req.Context())
	}))

	handler.ServeHTTP(httptest.NewRecorder(), req)
	if !ok || got.String() != "198.51.100.20" {
		t.Fatalf("expected forwarded IP in context, got %q ok=%v", got.String(), ok)
	}
}

func TestMiddlewareIgnoresForwardedIPFromUntrustedRemote(t *testing.T) {
	req := requestWithRemote("203.0.113.10:44321")
	req.Header.Set("X-Forwarded-For", "198.51.100.20")

	var got netip.Addr
	var ok bool
	handler := Middleware(nil)(http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		got, ok = FromContext(req.Context())
	}))

	handler.ServeHTTP(httptest.NewRecorder(), req)
	if !ok || got.String() != "203.0.113.10" {
		t.Fatalf("expected direct remote IP in context, got %q ok=%v", got.String(), ok)
	}
}

func requestWithRemote(remoteAddr string) *http.Request {
	req, _ := http.NewRequest(http.MethodPut, "/runtime/probes/probe/ip-family-capabilities", http.NoBody)
	req.RemoteAddr = remoteAddr
	return req
}
