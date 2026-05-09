package probe

import (
	"errors"
	"net/netip"
	"time"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
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

type Probe struct {
	ID        string
	ProjectID string
	Name      string
	Enabled   bool
	City      *string
	Latitude  *float64
	Longitude *float64
	Labels    []domainlabel.Label
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Credential struct {
	ProbeID    string
	ProjectID  string
	Enabled    bool
	SecretHash string
}

type Status struct {
	ProbeID      string
	State        State
	LastSeenAt   *time.Time
	AgentVersion *string
	PublicV4     *netip.Addr
	PublicV6     *netip.Addr
	Addrs        []netip.Addr
	UpdatedAt    time.Time
}

type CreateProbeStorageInput struct {
	ProjectID  string
	Name       string
	Enabled    bool
	City       *string
	Latitude   *float64
	Longitude  *float64
	LabelIDs   []string
	SecretHash string
}

type UpdateStatusInput struct {
	ProbeID      string
	State        State
	AgentVersion *string
	PublicV4     *netip.Addr
	PublicV6     *netip.Addr
	Addrs        []netip.Addr
}
