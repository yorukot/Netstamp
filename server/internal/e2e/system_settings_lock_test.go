//go:build integration

package e2e

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	pgsystem "github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/system"
)

func TestSystemSettingsResourceAdvisoryLocks(t *testing.T) {
	suite := newAPISuite(t)
	repository := pgsystem.NewRepository(suite.pool)
	transactor := postgres.NewTransactor(suite.pool)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	type lockAttempt struct {
		started  <-chan struct{}
		acquired <-chan struct{}
		done     <-chan error
	}
	startLock := func(resource string, release <-chan struct{}, resultErr error) lockAttempt {
		started := make(chan struct{})
		acquired := make(chan struct{})
		done := make(chan error, 1)
		go func() {
			close(started)
			done <- transactor.WithinTx(ctx, func(txCtx context.Context) error {
				if err := repository.LockSystemSettingsResource(txCtx, resource); err != nil {
					return err
				}
				close(acquired)
				if release != nil {
					select {
					case <-release:
					case <-txCtx.Done():
						return txCtx.Err()
					}
				}
				return resultErr
			})
		}()
		return lockAttempt{started: started, acquired: acquired, done: done}
	}
	waitForSignal := func(name string, signal <-chan struct{}) {
		t.Helper()
		select {
		case <-signal:
		case <-ctx.Done():
			t.Fatalf("wait for %s: %v", name, ctx.Err())
		}
	}
	waitForResult := func(name string, result <-chan error) error {
		t.Helper()
		select {
		case err := <-result:
			return err
		case <-ctx.Done():
			t.Fatalf("wait for %s: %v", name, ctx.Err())
			return ctx.Err()
		}
	}
	assertBlocked := func(name string, attempt lockAttempt) {
		t.Helper()
		waitForSignal(name+" start", attempt.started)
		select {
		case <-attempt.acquired:
			t.Fatalf("%s acquired a lock that should be blocked", name)
		case err := <-attempt.done:
			t.Fatalf("%s finished before the held lock was released: %v", name, err)
		case <-time.After(150 * time.Millisecond):
		}
	}

	commitRelease := make(chan struct{})
	commitHolder := startLock("access", commitRelease, nil)
	waitForSignal("commit holder acquisition", commitHolder.acquired)
	commitWaiter := startLock("access", nil, nil)
	assertBlocked("same-resource commit waiter", commitWaiter)

	independent := startLock("smtp", nil, nil)
	waitForSignal("different-resource acquisition", independent.acquired)
	if err := waitForResult("different-resource transaction", independent.done); err != nil {
		t.Fatalf("different resource lock failed: %v", err)
	}

	close(commitRelease)
	if err := waitForResult("committing holder", commitHolder.done); err != nil {
		t.Fatalf("commit holder failed: %v", err)
	}
	waitForSignal("same-resource acquisition after commit", commitWaiter.acquired)
	if err := waitForResult("same-resource transaction after commit", commitWaiter.done); err != nil {
		t.Fatalf("same resource lock after commit failed: %v", err)
	}

	rollbackErr := errors.New("force settings transaction rollback")
	rollbackRelease := make(chan struct{})
	rollbackHolder := startLock("auth.oidc", rollbackRelease, rollbackErr)
	waitForSignal("rollback holder acquisition", rollbackHolder.acquired)
	rollbackWaiter := startLock("auth.oidc", nil, nil)
	assertBlocked("same-resource rollback waiter", rollbackWaiter)

	close(rollbackRelease)
	if err := waitForResult("rolling back holder", rollbackHolder.done); !errors.Is(err, rollbackErr) {
		t.Fatalf("rollback holder error = %v, want %v", err, rollbackErr)
	}
	waitForSignal("same-resource acquisition after rollback", rollbackWaiter.acquired)
	if err := waitForResult("same-resource transaction after rollback", rollbackWaiter.done); err != nil {
		t.Fatalf("same resource lock after rollback failed: %v", err)
	}
}
