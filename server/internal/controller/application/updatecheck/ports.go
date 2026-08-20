package updatecheck

import (
	"context"
	"time"
)

type Release struct {
	TagName     string
	URL         string
	PublishedAt time.Time
}

type ReleaseSource interface {
	LatestRelease(ctx context.Context) (Release, error)
}

type SettingsReader interface {
	UpdateCheckEnabled(ctx context.Context) (bool, error)
}
