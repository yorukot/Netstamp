package probe

import (
	"errors"
	"net/netip"
	"regexp"
	"strings"
	"time"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	"github.com/yorukot/spvalidator"
)

var (
	ErrInvalidInput      = errors.New("probe input invalid")
	ErrProbeNotFound     = errors.New("probe not found")
	ErrProbeDisabled     = errors.New("probe disabled")
	ErrInvalidCredential = errors.New("probe credential invalid")
)

type State string

const (
	StateOnline  State = "online"
	StateOffline State = "offline"
)

var subdivisionCodeRe = regexp.MustCompile(`^[A-Z]{2}-[A-Z0-9]{1,3}$`)

type Probe struct {
	ID              string              `json:"id"`
	ProjectID       string              `json:"project_id"`
	Name            string              `json:"name"`
	Enabled         bool                `json:"enabled"`
	SubdivisionCode *string             `json:"subdivisionCode"`
	Latitude        *float64            `json:"latitude"`
	Longitude       *float64            `json:"longitude"`
	Labels          []domainlabel.Label `json:"labels"`
	Status          *Status             `json:"status"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	DeletedAt       *time.Time          `json:"deleted_at"`
}

func VNProbeID(probeID string) (string, error) {
	probeID = strings.TrimSpace(probeID)

	err := spvalidator.Required(probeID)
	if err != nil {
		return "", err
	}

	err = spvalidator.UUID(probeID)
	if err != nil {
		return "", err
	}

	return probeID, nil
}

func VNProbeProjectID(projectID string) (string, error) {
	projectID = strings.TrimSpace(projectID)

	err := spvalidator.Required(projectID)
	if err != nil {
		return "", err
	}

	err = spvalidator.UUID(projectID)
	if err != nil {
		return "", err
	}

	return projectID, nil
}

func VNProbeName(name string) (string, error) {
	name = strings.TrimSpace(name)

	err := spvalidator.Min(name, 1)
	if err != nil {
		return "", err
	}

	err = spvalidator.Max(name, 128)
	if err != nil {
		return "", err
	}

	return name, nil
}

func VNProbeOptionalName(name *string) (*string, error) {
	if name == nil {
		return nil, nil //nolint:nilnil // Nil means the caller did not provide a probe name.
	}

	normalized, err := VNProbeName(*name)
	if err != nil {
		return nil, err
	}

	return &normalized, nil
}

func VNProbeSubdivisionCode(subdivisionCode string) (string, error) {
	subdivisionCode = strings.TrimSpace(subdivisionCode)

	if !subdivisionCodeRe.MatchString(subdivisionCode) {
		return "", errors.New("invalid subdivision code")
	}

	return subdivisionCode, nil
}

func VNProbeOptionalSubdivisionCode(subdivisionCode *string) (*string, error) {
	if subdivisionCode == nil {
		return nil, nil //nolint:nilnil // Nil means the caller did not provide a subdivision code.
	}

	normalized, err := VNProbeSubdivisionCode(*subdivisionCode)
	if err != nil {
		return nil, err
	}

	return &normalized, nil
}

func VNProbeCoordinates(latitude, longitude *float64) (latitudeOutput, longitudeOutput *float64, latitudeErr, longitudeErr error) {
	if latitude != nil && longitude == nil {
		return nil, nil, nil, errors.New("longitude must be provided with latitude")
	}
	if latitude == nil && longitude != nil {
		return nil, nil, errors.New("latitude must be provided with longitude"), nil
	}
	if latitude == nil && longitude == nil {
		return nil, nil, nil, nil
	}
	if *longitude > 180 || *longitude < -180 {
		return nil, nil, nil, errors.New("longitude must be between -180 and 180")
	}
	if latitude != nil && (*latitude > 90 || *latitude < -90) {
		return nil, nil, errors.New("latitude must be between -90 and 90"), nil
	}

	lat := *latitude
	lon := *longitude
	return &lat, &lon, nil, nil
}

type Credential struct {
	ProbeID    string `json:"probe_id"`
	ProjectID  string `json:"project_id"`
	Enabled    bool   `json:"enabled"`
	SecretHash string `json:"secret_hash"`
}

type Status struct {
	ProbeID      string       `json:"probe_id"`
	State        State        `json:"state"`
	LastSeenAt   *time.Time   `json:"last_seen_at"`
	AgentVersion *string      `json:"agent_version"`
	PublicV4     *netip.Addr  `json:"public_v4"`
	PublicV6     *netip.Addr  `json:"public_v6"`
	AS           *string      `json:"as"`
	Addrs        []netip.Addr `json:"addrs"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

func VNProbeStatus(status Status) (Status, error) {
	probeID, err := VNProbeID(status.ProbeID)
	if err != nil {
		return Status{}, err
	}
	state, err := VNProbeState(status.State)
	if err != nil {
		return Status{}, err
	}
	agentVersion, err := VNProbeOptionalAgentVersion(status.AgentVersion)
	if err != nil {
		return Status{}, err
	}
	publicV4, err := VNProbePublicV4(status.PublicV4)
	if err != nil {
		return Status{}, err
	}
	publicV6, err := VNProbePublicV6(status.PublicV6)
	if err != nil {
		return Status{}, err
	}
	as, err := VNProbeOptionalAS(status.AS)
	if err != nil {
		return Status{}, err
	}
	addrs, err := VNProbeAddrs(status.Addrs)
	if err != nil {
		return Status{}, err
	}

	return Status{
		ProbeID:      probeID,
		State:        state,
		LastSeenAt:   status.LastSeenAt,
		AgentVersion: agentVersion,
		PublicV4:     publicV4,
		PublicV6:     publicV6,
		AS:           as,
		Addrs:        addrs,
		UpdatedAt:    status.UpdatedAt,
	}, nil
}

func VNProbeState(state State) (State, error) {
	state = State(strings.TrimSpace(string(state)))

	switch state {
	case StateOnline, StateOffline:
		return state, nil
	default:
		return "", errors.New("invalid probe state")
	}
}

func VNProbeAgentVersion(agentVersion string) (string, error) {
	agentVersion = strings.TrimSpace(agentVersion)

	err := spvalidator.Min(agentVersion, 1)
	if err != nil {
		return "", err
	}

	err = spvalidator.Max(agentVersion, 256)
	if err != nil {
		return "", err
	}

	return agentVersion, nil
}

func VNProbeOptionalAgentVersion(agentVersion *string) (*string, error) {
	if agentVersion == nil {
		return nil, nil //nolint:nilnil // Nil means the caller did not provide an agent version.
	}

	normalized, err := VNProbeAgentVersion(*agentVersion)
	if err != nil {
		return nil, err
	}

	return &normalized, nil
}

func VNProbeAS(as string) (string, error) {
	as = strings.TrimSpace(as)

	err := spvalidator.Min(as, 1)
	if err != nil {
		return "", err
	}

	err = spvalidator.Max(as, 32)
	if err != nil {
		return "", err
	}

	return as, nil
}

func VNProbeOptionalAS(as *string) (*string, error) {
	if as == nil {
		return nil, nil //nolint:nilnil // Nil means the caller did not provide an AS value.
	}

	normalized, err := VNProbeAS(*as)
	if err != nil {
		return nil, err
	}

	return &normalized, nil
}

func VNProbePublicV4(addr *netip.Addr) (*netip.Addr, error) {
	if addr == nil {
		return nil, nil //nolint:nilnil // Nil means the caller did not provide an IPv4 address.
	}
	if !addr.Is4() {
		return nil, errors.New("invalid public IPv4 address")
	}

	normalized := *addr
	return &normalized, nil
}

func VNProbePublicV6(addr *netip.Addr) (*netip.Addr, error) {
	if addr == nil {
		return nil, nil //nolint:nilnil // Nil means the caller did not provide an IPv6 address.
	}
	if !addr.Is6() {
		return nil, errors.New("invalid public IPv6 address")
	}

	normalized := *addr
	return &normalized, nil
}

func VNProbeAddrs(addrs []netip.Addr) ([]netip.Addr, error) {
	normalized := make([]netip.Addr, 0, len(addrs))
	for _, addr := range addrs {
		if !addr.IsValid() {
			return nil, errors.New("invalid probe address")
		}
		normalized = append(normalized, addr)
	}

	return normalized, nil
}
