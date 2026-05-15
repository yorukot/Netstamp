package check

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/yorukot/spvalidator"

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
	ID              string              `json:"id"`
	ProjectID       string              `json:"projectId"`
	Name            string              `json:"name"`
	Type            Type                `json:"type"`
	Target          string              `json:"target"`
	Selector        json.RawMessage     `json:"selector"`
	Description     *string             `json:"description"`
	IntervalSeconds int32               `json:"intervalSeconds"`
	Labels          []domainlabel.Label `json:"labels"`
	CreatedAt       time.Time           `json:"createdAt"`
	UpdatedAt       time.Time           `json:"updatedAt"`
	DeletedAt       *time.Time          `json:"-"`

	// We keep this nullable because we might have other config later and the user can only send a single config in as well as single config out.
	PingConfig *domainping.Config `json:"pingConfig"`
}

func (c Check) IntervalTime() time.Duration {
	return time.Duration(c.IntervalSeconds) * time.Second
}

func VNCheckID(checkID string) (string, error) {
	checkID = strings.TrimSpace(checkID)

	err := spvalidator.Required(checkID)
	if err != nil {
		return "", err
	}

	err = spvalidator.UUID(checkID)
	if err != nil {
		return "", err
	}

	return checkID, nil
}

func VNCheckName(name string) (string, error) {
	name = strings.TrimSpace(name)
	err := spvalidator.Required(name)
	if err != nil {
		return "", err
	}

	err = spvalidator.Max(name, 128)
	if err != nil {
		return "", err
	}

	return name, nil
}

func VNCheckType(t Type) (Type, error) {
	t = Type(strings.TrimSpace(string(t)))
	switch t {
	case TypePing:
		return TypePing, nil
	default:
		return "", errors.New("invalid check type")
	}
}

func VNCheckTarget(target string) (string, error) {
	target = strings.TrimSpace(target)
	err := spvalidator.Required(target)
	if err != nil {
		return "", err
	}

	err = spvalidator.Max(target, 128)
	if err != nil {
		return "", err
	}

	ipErr := spvalidator.IP(target)
	domainErr := spvalidator.Domain(target)

	if ipErr != nil && domainErr != nil {
		return "", errors.New("target must be a valid IP or domain")
	}

	return target, nil
}

func VNCheckSelector(selector json.RawMessage) (json.RawMessage, error) {
	if selector == nil {
		return nil, nil
	}

	err := spvalidator.Max(selector, 16384)
	if err != nil {
		return nil, err
	}

	return selector, nil
}

func VNCheckDescription(description *string) (*string, error) {
	if description == nil {
		return nil, nil //nolint:nilnil // Nil means the caller did not provide a description.
	}

	trimmed := strings.TrimSpace(*description)
	description = &trimmed

	err := spvalidator.Required(trimmed)
	if err != nil {
		return nil, err
	}

	err = spvalidator.Max(trimmed, 1024)
	if err != nil {
		return nil, err
	}

	return description, nil
}

func VNCheckInterval(interval int32) (int32, error) {
	err := spvalidator.Min(interval, 1)
	if err != nil {
		return 0, err
	}

	return interval, nil
}
