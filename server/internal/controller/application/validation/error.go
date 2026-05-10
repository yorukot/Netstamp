package validation

import "errors"

type FieldError struct {
	Field   string
	Message string
	Value   any
}

type Error struct {
	base   error
	fields []FieldError
}

func New(base error, field, message string, value any) error {
	return NewFields(base, FieldError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

func NewFields(base error, fields ...FieldError) error {
	copied := make([]FieldError, len(fields))
	copy(copied, fields)

	return &Error{
		base:   base,
		fields: copied,
	}
}

func FieldErrors(err error) ([]FieldError, bool) {
	var validationErr *Error
	if !errors.As(err, &validationErr) || len(validationErr.fields) == 0 {
		return nil, false
	}

	fields := make([]FieldError, len(validationErr.fields))
	copy(fields, validationErr.fields)

	return fields, true
}

func (e *Error) Error() string {
	if e.base == nil {
		return "validation failed"
	}

	return e.base.Error()
}

func (e *Error) Unwrap() error {
	return e.base
}
