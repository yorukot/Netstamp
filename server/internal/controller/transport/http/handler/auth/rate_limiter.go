package auth

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/yorukot/netstamp/internal/controller/transport/http/clientip"
	"github.com/yorukot/netstamp/internal/platform/normalize"
)

type PasswordResetRateLimitConfig struct {
	Window     time.Duration
	IPLimit    int32
	EmailLimit int32
}

type PasswordResetRateLimiter struct {
	window     time.Duration
	ipLimit    int32
	emailLimit int32
	now        func() time.Time

	mu      sync.Mutex
	buckets map[string]rateBucket
}

type rateBucket struct {
	count    int32
	resetAt  time.Time
	lastSeen time.Time
}

func NewPasswordResetRateLimiter(cfg PasswordResetRateLimitConfig) *PasswordResetRateLimiter {
	if cfg.Window <= 0 {
		cfg.Window = time.Hour
	}
	if cfg.IPLimit <= 0 {
		cfg.IPLimit = 10
	}
	if cfg.EmailLimit <= 0 {
		cfg.EmailLimit = 3
	}

	return &PasswordResetRateLimiter{
		window:     cfg.Window,
		ipLimit:    cfg.IPLimit,
		emailLimit: cfg.EmailLimit,
		now:        func() time.Time { return time.Now().UTC() },
		buckets:    make(map[string]rateBucket),
	}
}

func (l *PasswordResetRateLimiter) Allow(ctx context.Context, clientKey, email string) bool {
	if l == nil {
		return true
	}

	now := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()

	l.prune(now)

	if !l.allowKey("ip:"+clientKey, l.ipLimit, now) {
		return false
	}

	email = normalize.Email(email)
	if email == "" {
		return true
	}

	if !l.allowKey("email:"+email, l.emailLimit, now) {
		return false
	}

	_ = ctx
	return true
}

func (l *PasswordResetRateLimiter) allowKey(key string, limit int32, now time.Time) bool {
	bucket := l.buckets[key]
	if bucket.resetAt.IsZero() || !bucket.resetAt.After(now) {
		bucket = rateBucket{resetAt: now.Add(l.window)}
	}
	if bucket.count >= limit {
		bucket.lastSeen = now
		l.buckets[key] = bucket
		return false
	}

	bucket.count++
	bucket.lastSeen = now
	l.buckets[key] = bucket
	return true
}

func (l *PasswordResetRateLimiter) prune(now time.Time) {
	for key, bucket := range l.buckets {
		if bucket.resetAt.Before(now) && bucket.lastSeen.Add(l.window).Before(now) {
			delete(l.buckets, key)
		}
	}
}

func resetLimiterClientKey(r *http.Request) string {
	if addr, ok := clientip.FromContext(r.Context()); ok {
		return addr.String()
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	if trimmed := strings.TrimSpace(r.RemoteAddr); trimmed != "" {
		return trimmed
	}
	return "unknown"
}
