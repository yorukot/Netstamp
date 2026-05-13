package runtime

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	Version     = "0.1.0"
	AgentString = "netstamp-probe/" + Version
)

func EnsureMinimumVersion(current, minimum string) error {
	currentParts, err := parseVersion(current)
	if err != nil {
		return err
	}
	minimumParts, err := parseVersion(minimum)
	if err != nil {
		return err
	}

	for i := range currentParts {
		if currentParts[i] > minimumParts[i] {
			return nil
		}
		if currentParts[i] < minimumParts[i] {
			return fmt.Errorf("%w: current=%s minimum=%s", ErrVersionUnsupported, current, minimum)
		}
	}

	return nil
}

func parseVersion(value string) ([3]int, error) {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "netstamp-probe/")
	value = strings.TrimPrefix(value, "v")

	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return [3]int{}, fmt.Errorf("invalid semantic version %q", value)
	}

	var out [3]int
	for i, part := range parts {
		parsed, err := strconv.Atoi(part)
		if err != nil || parsed < 0 {
			return [3]int{}, fmt.Errorf("invalid semantic version %q", value)
		}
		out[i] = parsed
	}

	return out, nil
}
