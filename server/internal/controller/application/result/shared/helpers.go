package shared

import (
	"time"

	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
)

func IPFamilyString(value *domainnetwork.IPFamily) *string {
	if value == nil {
		return nil
	}
	copied := string(*value)
	return &copied
}

func StringPointer(value string) *string {
	copied := value
	return &copied
}

func AveragePtr(sum float64, count int32) *float64 {
	if count == 0 {
		return nil
	}
	average := sum / float64(count)
	return &average
}

func TimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := value.UTC()
	return &copied
}

func TimePtrMillis(value *time.Time) *int64 {
	if value == nil {
		return nil
	}
	millis := value.UTC().UnixMilli()
	return &millis
}
