package updatecheck

import (
	"sync"
	"time"
)

type Status struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	ReleaseURL      string
	PublishedAt     time.Time
	LastCheckedAt   time.Time
	CheckError      string
}

type Cache struct {
	mu     sync.RWMutex
	status Status
}

func NewCache(currentVersion string) *Cache {
	return &Cache{status: Status{CurrentVersion: currentVersion}}
}

func (c *Cache) Snapshot() Status {
	if c == nil {
		return Status{}
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

func (c *Cache) StoreSuccess(status Status) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	status.CheckError = ""
	c.status = status
}

func (c *Cache) StoreFailure(currentVersion string, checkedAt time.Time, err error) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status.CurrentVersion = currentVersion
	c.status.LastCheckedAt = checkedAt
	if err == nil {
		c.status.CheckError = ""
		return
	}
	c.status.CheckError = err.Error()
}

func (c *Cache) Clear(currentVersion string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status = Status{CurrentVersion: currentVersion}
}
