package label

import (
	"strings"
	"time"

	"github.com/yorukot/spvalidator"
)

type Label struct {
	ID        string
	ProjectID string
	Key       string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func VNLabelID(labelID string) (string, error) {
	labelID = strings.TrimSpace(labelID)

	err := spvalidator.Required(labelID)
	if err != nil {
		return "", err
	}
	err = spvalidator.UUID(labelID)
	if err != nil {
		return "", err
	}
	return labelID, nil
}

func VNLabelKey(key string) (string, error) {
	key = strings.TrimSpace(key)

	err := spvalidator.Min(key, 1)
	if err != nil {
		return "", err
	}

	err = spvalidator.Max(key, 64)
	if err != nil {
		return "", err
	}

	return key, nil
}

func VNLabelValue(value string) (string, error) {
	value = strings.TrimSpace(value)

	err := spvalidator.Min(value, 1)
	if err != nil {
		return "", err
	}

	err = spvalidator.Max(value, 64)
	if err != nil {
		return "", err
	}

	return value, nil
}
