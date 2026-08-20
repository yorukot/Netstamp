package updatecheck

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCacheSupportsConcurrentReadsAndWrites(t *testing.T) {
	cache := NewCache("0.0.0")
	checkedAt := time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)

	var waitGroup sync.WaitGroup
	for index := range 100 {
		waitGroup.Add(2)
		go func() {
			defer waitGroup.Done()
			cache.StoreSuccess(Status{
				CurrentVersion: "0.0.0",
				LatestVersion:  fmt.Sprintf("0.0.%d", index),
				LastCheckedAt:  checkedAt,
			})
		}()
		go func() {
			defer waitGroup.Done()
			_ = cache.Snapshot()
		}()
	}
	waitGroup.Wait()

	status := cache.Snapshot()
	if status.CurrentVersion != "0.0.0" || status.LastCheckedAt != checkedAt {
		t.Fatalf("unexpected final cache snapshot: %#v", status)
	}
}

func TestCacheFailurePreservesLastSuccessfulRelease(t *testing.T) {
	cache := NewCache("0.0.0")
	publishedAt := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)
	firstCheck := publishedAt.Add(time.Hour)
	secondCheck := firstCheck.Add(time.Hour)
	cache.StoreSuccess(Status{
		CurrentVersion:  "0.0.0",
		LatestVersion:   "v1.0.0",
		UpdateAvailable: true,
		ReleaseURL:      "https://github.com/yorukot/netstamp/releases/tag/v1.0.0",
		PublishedAt:     publishedAt,
		LastCheckedAt:   firstCheck,
	})

	cache.StoreFailure("0.0.0", secondCheck, errors.New("offline"))
	status := cache.Snapshot()
	if status.LatestVersion != "v1.0.0" || !status.UpdateAvailable || status.PublishedAt != publishedAt {
		t.Fatalf("release result was not preserved: %#v", status)
	}
	if status.LastCheckedAt != secondCheck || status.CheckError != "offline" {
		t.Fatalf("failure metadata was not recorded: %#v", status)
	}
}
