package check

import (
	"encoding/json"
	"errors"
	"time"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
)

var (
	ErrCheckNotFound = errors.New("check not found")
	ErrInvalidInput  = errors.New("check input invalid")
)

type Type string

const (
	TypePing Type = "ping"
)

type IPFamily string

const (
	IPFamilyIPv4 IPFamily = "ipv4"
	IPFamilyIPv6 IPFamily = "ipv6"
)

type PingConfig struct {
	PacketCount     int32
	PacketSizeBytes int32
	TimeoutMs       int32
	IPFamily        *IPFamily
}

type Check struct {
	ID              string
	ProjectID       string
	Name            string
	Type            Type
	Target          string
	Selector        json.RawMessage
	Description     *string
	IntervalSeconds int32
	PingConfig      PingConfig
	Labels          []domainlabel.Label
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type CreateCheckStorageInput struct {
	ProjectID       string
	Name            string
	Type            Type
	Target          string
	Selector        json.RawMessage
	CheckVersion    string
	SelectorVersion string
	Description     *string
	IntervalSeconds int32
	PingConfig      PingConfig
	LabelIDs        []string
}

type UpdateCheckStorageInput struct {
	ProjectID       string
	CheckID         string
	Name            string
	Type            Type
	Target          string
	Selector        json.RawMessage
	CheckVersion    string
	SelectorVersion string
	Description     *string
	IntervalSeconds int32
	PingConfig      PingConfig
	ReplaceLabels   bool
	LabelIDs        []string
}
