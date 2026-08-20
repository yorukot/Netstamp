package updatecheck

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWorkerRunOnceStoresAvailableRelease(t *testing.T) {
	checkedAt := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)
	publishedAt := checkedAt.Add(-time.Hour)
	releases := &fakeReleaseSource{release: Release{
		TagName:     "v1.2.3",
		URL:         "https://github.com/yorukot/netstamp/releases/tag/v1.2.3",
		PublishedAt: publishedAt,
	}}
	cache := NewCache("0.0.0")
	worker := NewWorker(cache, releases, fakeSettingsReader{enabled: true}, nil)
	worker.now = func() time.Time { return checkedAt }

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("run update check: %v", err)
	}
	status := cache.Snapshot()
	if status.CurrentVersion != "0.0.0" || status.LatestVersion != "v1.2.3" || !status.UpdateAvailable {
		t.Fatalf("unexpected update status: %#v", status)
	}
	if status.ReleaseURL != releases.release.URL || status.PublishedAt != publishedAt || status.LastCheckedAt != checkedAt || status.CheckError != "" {
		t.Fatalf("unexpected release metadata: %#v", status)
	}
}

func TestWorkerRunOnceStoresNoUpdateForCurrentRelease(t *testing.T) {
	cache := NewCache("0.0.0")
	worker := NewWorker(cache, &fakeReleaseSource{release: Release{
		TagName: "v0.0.0", URL: "https://github.com/yorukot/netstamp/releases/tag/v0.0.0",
		PublishedAt: time.Date(2026, 8, 20, 9, 0, 0, 0, time.UTC),
	}}, fakeSettingsReader{enabled: true}, nil)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("run update check: %v", err)
	}
	if status := cache.Snapshot(); status.UpdateAvailable || status.LatestVersion != "v0.0.0" {
		t.Fatalf("unexpected update status: %#v", status)
	}
}

func TestWorkerRunOnceFailurePreservesSuccessfulResult(t *testing.T) {
	firstCheck := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)
	secondCheck := firstCheck.Add(time.Hour)
	cache := NewCache("0.0.0")
	releases := &fakeReleaseSource{release: Release{
		TagName: "v1.0.0", URL: "https://github.com/yorukot/netstamp/releases/tag/v1.0.0",
		PublishedAt: firstCheck.Add(-time.Hour),
	}}
	worker := NewWorker(cache, releases, fakeSettingsReader{enabled: true}, nil)
	worker.now = func() time.Time { return firstCheck }
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("run successful update check: %v", err)
	}

	releases.err = errors.New("github unavailable")
	worker.now = func() time.Time { return secondCheck }
	if err := worker.RunOnce(context.Background()); err == nil {
		t.Fatal("expected update check failure")
	}
	status := cache.Snapshot()
	if status.LatestVersion != "v1.0.0" || !status.UpdateAvailable {
		t.Fatalf("successful release result was not preserved: %#v", status)
	}
	if status.LastCheckedAt != secondCheck || status.CheckError == "" {
		t.Fatalf("failure metadata was not recorded: %#v", status)
	}
}

func TestWorkerRunOnceDisabledClearsStatusWithoutFetching(t *testing.T) {
	cache := NewCache("0.0.0")
	cache.StoreSuccess(Status{CurrentVersion: "0.0.0", LatestVersion: "v1.0.0", UpdateAvailable: true})
	releases := &fakeReleaseSource{}
	worker := NewWorker(cache, releases, fakeSettingsReader{enabled: false}, nil)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("run disabled update check: %v", err)
	}
	if releases.calls != 0 {
		t.Fatalf("disabled update check fetched %d releases", releases.calls)
	}
	if status := cache.Snapshot(); status != (Status{CurrentVersion: "0.0.0"}) {
		t.Fatalf("disabled update check did not clear status: %#v", status)
	}
}

func TestWorkerRunOnceSettingFailurePreservesStatus(t *testing.T) {
	checkedAt := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)
	cache := NewCache("0.0.0")
	cache.StoreSuccess(Status{CurrentVersion: "0.0.0", LatestVersion: "v1.0.0", UpdateAvailable: true})
	worker := NewWorker(cache, &fakeReleaseSource{}, fakeSettingsReader{err: errors.New("database unavailable")}, nil)
	worker.now = func() time.Time { return checkedAt }

	if err := worker.RunOnce(context.Background()); err == nil {
		t.Fatal("expected settings failure")
	}
	status := cache.Snapshot()
	if status.LatestVersion != "v1.0.0" || !status.UpdateAvailable || status.LastCheckedAt != checkedAt || status.CheckError == "" {
		t.Fatalf("unexpected status after settings failure: %#v", status)
	}
}

type fakeReleaseSource struct {
	release Release
	err     error
	calls   int
}

func (f *fakeReleaseSource) LatestRelease(context.Context) (Release, error) {
	f.calls++
	return f.release, f.err
}

type fakeSettingsReader struct {
	enabled bool
	err     error
}

func (f fakeSettingsReader) UpdateCheckEnabled(context.Context) (bool, error) {
	return f.enabled, f.err
}
