package probe

import (
	"errors"
	"time"
)

var (
	ErrLabelNotFound = errors.New("probe label not found")
	ErrInvalidInput  = errors.New("probe input invalid")
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
	Labels    []Label
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Label struct {
	ID        string
	ProjectID string
	Key       string
	Value     string
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
