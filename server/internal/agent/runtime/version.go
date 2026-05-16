package agentruntime

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
	currentVersion, err := parseVersion(current)
	if err != nil {
		return err
	}
	minimumVersion, err := parseVersion(minimum)
	if err != nil {
		return err
	}

	if currentVersion.compare(minimumVersion) < 0 {
		return fmt.Errorf("%w: current=%s minimum=%s", ErrVersionUnsupported, current, minimum)
	}

	return nil
}

type semanticVersion struct {
	major int
	minor int
	patch int
}

func (v semanticVersion) compare(other semanticVersion) int {
	switch {
	case v.major != other.major:
		return compareInt(v.major, other.major)
	case v.minor != other.minor:
		return compareInt(v.minor, other.minor)
	default:
		return compareInt(v.patch, other.patch)
	}
}

func compareInt(left, right int) int {
	switch {
	case left > right:
		return 1
	case left < right:
		return -1
	default:
		return 0
	}
}

func parseVersion(value string) (semanticVersion, error) {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "netstamp-probe/")
	value = strings.TrimPrefix(value, "v")

	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return semanticVersion{}, fmt.Errorf("invalid semantic version %q", value)
	}

	var parsed [3]int
	for i, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return semanticVersion{}, fmt.Errorf("invalid semantic version %q", value)
		}
		parsed[i] = number
	}

	return semanticVersion{
		major: parsed[0],
		minor: parsed[1],
		patch: parsed[2],
	}, nil
}
