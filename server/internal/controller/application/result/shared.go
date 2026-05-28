package result

import (
	"time"

	domainnetwork "github.com/yorukot/netstamp/internal/domain/network"
)

func ipFamilyString(value *domainnetwork.IPFamily) *string {
	if value == nil {
		return nil
	}
	copied := string(*value)
	return &copied
}

func stringPointer(value string) *string {
	copied := value
	return &copied
}

func averagePtr(sum float64, count int32) *float64 {
	if count == 0 {
		return nil
	}
	average := sum / float64(count)
	return &average
}

func timePtrMillis(value *time.Time) *int64 {
	if value == nil {
		return nil
	}
	millis := value.UTC().UnixMilli()
	return &millis
}
