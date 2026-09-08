package version

import (
	"fmt"
	"strings"

	"github.com/coreos/go-semver/semver"
)

const (
	Product = "0.1.0"
	API     = "v1"
	// MinimumAgent is the oldest probe agent build that still speaks this controller's runtime
	// contract. Agents below it refuse to start, so raise it only for a breaking runtime change.
	MinimumAgent = "0.0.0"
	GitHubOwner  = "yorukot"
	GitHubRepo   = "netstamp"
	agentPrefix  = "netstamp-probe/"
)

func Agent() string {
	return agentPrefix + Product
}

func Compare(left, right string) (int, error) {
	leftVersion, err := parse(left)
	if err != nil {
		return 0, err
	}
	rightVersion, err := parse(right)
	if err != nil {
		return 0, err
	}
	return leftVersion.Compare(*rightVersion), nil
}

func parse(value string) (*semver.Version, error) {
	normalized := strings.TrimSpace(value)
	normalized = strings.TrimPrefix(normalized, agentPrefix)
	normalized = strings.TrimPrefix(normalized, "v")

	parsed, err := semver.NewVersion(normalized)
	if err != nil {
		return nil, fmt.Errorf("invalid semantic version %q: %w", value, err)
	}
	return parsed, nil
}
