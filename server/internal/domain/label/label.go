package label

import (
	"errors"
	"time"
)

var (
	ErrLabelNotFound      = errors.New("label not found")
	ErrLabelAlreadyExists = errors.New("label already exists")
	ErrInvalidInput       = errors.New("label input invalid")
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

type CreateLabelStorageInput struct {
	ProjectID string
	Key       string
	Value     string
}

type UpdateLabelStorageInput struct {
	ProjectID string
	LabelID   string
	Key       string
	Value     string
}
