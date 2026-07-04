package publicstatus

import (
	"sync"
	"time"

	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
)

const publicSnapshotTTL = 30 * time.Second

type publicSnapshot struct {
	page        domainpublic.Page
	status      domainpublic.Status
	elements    []domainpublic.RenderedElement
	incidents   []domainpublic.Incident
	generatedAt time.Time
}

type publicSnapshotCache struct {
	mu      sync.Mutex
	ttl     time.Duration
	entries map[string]publicSnapshotCacheEntry
}

type publicSnapshotCacheEntry struct {
	snapshot  publicSnapshot
	expiresAt time.Time
}

func newPublicSnapshotCache(ttl time.Duration) *publicSnapshotCache {
	return &publicSnapshotCache{ttl: ttl, entries: make(map[string]publicSnapshotCacheEntry)}
}

func (c *publicSnapshotCache) get(slug string, now time.Time) (publicSnapshot, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[slug]
	if !ok {
		return publicSnapshot{}, false
	}
	if !now.Before(entry.expiresAt) {
		delete(c.entries, slug)
		return publicSnapshot{}, false
	}
	return entry.snapshot, true
}

func (c *publicSnapshotCache) set(slug string, snapshot publicSnapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[slug] = publicSnapshotCacheEntry{
		snapshot:  snapshot,
		expiresAt: snapshot.generatedAt.Add(c.ttl),
	}
}

func (c *publicSnapshotCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]publicSnapshotCacheEntry)
}
