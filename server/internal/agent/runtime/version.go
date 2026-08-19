package agentruntime

import (
	"fmt"

	appversion "github.com/yorukot/netstamp/internal/version"
)

func EnsureMinimumVersion(current, minimum string) error {
	comparison, err := appversion.Compare(current, minimum)
	if err != nil {
		return err
	}

	if comparison < 0 {
		return fmt.Errorf("%w: current=%s minimum=%s", ErrVersionUnsupported, current, minimum)
	}

	return nil
}
