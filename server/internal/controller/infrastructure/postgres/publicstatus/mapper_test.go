package pgpublicstatus

import (
	"testing"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

func TestPublicCheckTargetRedactsHTTPQuery(t *testing.T) {
	const target = "https://status.example.com/health?signature=secret&expires=123"
	if got := publicCheckTarget(domaincheck.TypeHTTP, target); got != "https://status.example.com/health" {
		t.Fatalf("expected redacted HTTP target, got %q", got)
	}
	if got := publicCheckTarget(domaincheck.TypeTCP, "example.com"); got != "example.com" {
		t.Fatalf("expected non-HTTP target to remain unchanged, got %q", got)
	}
}
