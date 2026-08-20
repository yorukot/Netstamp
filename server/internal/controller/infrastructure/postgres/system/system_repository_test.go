package pgsystem

import (
	"context"
	"testing"
)

func TestLockSystemSettingsResourceRequiresTransaction(t *testing.T) {
	repo := &Repository{}

	err := repo.LockSystemSettingsResource(context.Background(), "access")
	if err == nil {
		t.Fatal("expected lock outside transaction to fail")
	}
	if got, want := err.Error(), "lock system settings resource outside transaction"; got != want {
		t.Fatalf("unexpected error: got %q, want %q", got, want)
	}
}
