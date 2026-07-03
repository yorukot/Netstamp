package assignment

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yorukot/spvalidator"

	"github.com/yorukot/netstamp/internal/domain/check"
	"github.com/yorukot/netstamp/internal/domain/probe"
)

var ErrInvalidInput = errors.New("assignment input invalid")

const DefaultRefreshJobMaxAttempts int32 = 5

type RefreshTargetType string

const (
	RefreshTargetProject RefreshTargetType = "project"
	RefreshTargetProbe   RefreshTargetType = "probe"
	RefreshTargetCheck   RefreshTargetType = "check"
	RefreshTargetLabel   RefreshTargetType = "label"
)

type RefreshJobStatus string

const (
	RefreshJobStatusPending   RefreshJobStatus = "pending"
	RefreshJobStatusRunning   RefreshJobStatus = "running"
	RefreshJobStatusSucceeded RefreshJobStatus = "succeeded"
	RefreshJobStatusFailed    RefreshJobStatus = "failed"
	RefreshJobStatusDiscarded RefreshJobStatus = "discarded"
)

type RefreshTarget struct {
	ProjectID string
	Type      RefreshTargetType
	TargetID  string
}

func (target RefreshTarget) DedupeKey() string {
	return fmt.Sprintf("%s:%s:%s", target.ProjectID, target.Type, target.TargetID)
}

type RefreshJob struct {
	ID            string
	ProjectID     string
	TargetType    RefreshTargetType
	TargetID      string
	Status        RefreshJobStatus
	AttemptCount  int32
	MaxAttempts   int32
	NextAttemptAt time.Time
	LastAttemptAt *time.Time
	CompletedAt   *time.Time
	LastErrorKind *string
	LastErrorCode *string
	LastError     *string
	DedupeKey     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Assignment struct {
	ID              string       `json:"id"`
	ProjectID       string       `json:"projectId"`
	ProbeID         string       `json:"probeId"`
	CheckID         string       `json:"checkId"`
	ProbeStorageID  int64        `json:"-"`
	CheckStorageID  int64        `json:"-"`
	CheckVersion    string       `json:"checkVersion"`
	SelectorVersion string       `json:"selectorVersion"`
	Check           *check.Check `json:"check,omitempty"`
	Probe           *probe.Probe `json:"probe,omitempty"`
}

type Query struct {
	ProjectID string
	ProbeID   string
	CheckID   string
}

func VNProjectID(projectID string) (string, error) {
	return validUUID(projectID)
}

func VNProbeID(probeID string) (string, error) {
	return validUUID(probeID)
}

func VNCheckID(checkID string) (string, error) {
	return validUUID(checkID)
}

func VNLabelID(labelID string) (string, error) {
	return validUUID(labelID)
}

func validUUID(value string) (string, error) {
	value = strings.TrimSpace(value)

	err := spvalidator.Required(value)
	if err != nil {
		return "", err
	}
	err = spvalidator.UUID(value)
	if err != nil {
		return "", err
	}

	return value, nil
}
