package alert

import (
	"fmt"

	"github.com/google/uuid"
)

func validateUUIDPtr(value *string) error {
	if value == nil {
		return nil
	}
	if _, err := uuid.Parse(*value); err != nil {
		return fmt.Errorf("%w: invalid uuid", ErrInvalidInput)
	}
	return nil
}
