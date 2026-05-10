package logger

import (
	"testing"

	"go.uber.org/zap"
)

func TestNewZapConfigDisablesImplicitStacktraces(t *testing.T) {
	for _, env := range []string{"local", "production"} {
		t.Run(env, func(t *testing.T) {
			cfg := newZapConfig(Config{Env: env}, zap.NewAtomicLevel())
			if !cfg.DisableStacktrace {
				t.Fatal("expected implicit stacktraces to be disabled")
			}
		})
	}
}
