package probe

import (
	"errors"
	"time"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
)

var (
	ErrInvalidInput = errors.New("probe input invalid")
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
