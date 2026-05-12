package label

import (
	"strings"
	"time"

	"github.com/yorukot/spvalidator"
)

type Label struct {
	ID        string     `json:"id"`
	ProjectID string     `json:"projectId"`
	Key       string     `json:"key"`
	Value     string     `json:"value"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"-"`
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

func VNLabelIDs(labelIDs []string) ([]string, error) {
	normalized := make([]string, 0, len(labelIDs))
	for _, labelID := range labelIDs {
		labelID, err := VNLabelID(labelID)
		if err != nil {
			return nil, err
		}
		normalized = append(normalized, labelID)
	}

	return normalized, nil
}

func VNOptionalLabelIDs(labelIDs *[]string) (*[]string, error) {
	if labelIDs == nil {
		return nil, nil //nolint:nilnil // Nil means the caller did not provide label IDs.
	}

	normalized, err := VNLabelIDs(*labelIDs)
	if err != nil {
		return nil, err
	}

	return &normalized, nil
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
