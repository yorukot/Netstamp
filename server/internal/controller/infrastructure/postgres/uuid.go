package postgres

import (
	"fmt"

	"github.com/google/uuid"
)

func ParseUUID(value string, wrap error) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse uuid: %w", wrap)
	}

	return id, nil
}
