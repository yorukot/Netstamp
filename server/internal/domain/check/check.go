package check

import (
	"encoding/json"
	"errors"
	"time"

	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

var (
	ErrCheckNotFound = errors.New("check not found")
	ErrInvalidInput  = errors.New("check input invalid")
)

type Type string

const (
	TypePing Type = "ping"
)

type Check struct {
	ID              string
	ProjectID       string
	Name            string
	Type            Type
	Target          string
	Selector        json.RawMessage
	Description     *string
	IntervalSeconds int32
	PingConfig      domainping.Config
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
	PingConfig      domainping.Config
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
	PingConfig      domainping.Config
	ReplaceLabels   bool
	LabelIDs        []string
}
