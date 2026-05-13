package runtime

import (
	"errors"
	"testing"
)

func TestEnsureMinimumVersion(t *testing.T) {
	t.Parallel()

	if err := EnsureMinimumVersion("0.1.0", "0.1.0"); err != nil {
		t.Fatalf("expected equal versions to pass: %v", err)
	}
	if err := EnsureMinimumVersion("0.2.0", "0.1.0"); err != nil {
		t.Fatalf("expected newer version to pass: %v", err)
	}
	if err := EnsureMinimumVersion("0.1.0", "0.2.0"); !errors.Is(err, ErrVersionUnsupported) {
		t.Fatalf("expected unsupported version error, got %v", err)
	}
}
