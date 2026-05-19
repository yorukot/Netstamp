package validation

import "errors"

type FieldError struct {
	Field   string
	Message string
	Value   any
}

type Collector struct {
	fields []FieldError
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

func (c *Collector) Add(field, message string, value any) {
	c.fields = append(c.fields, FieldError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

func (c *Collector) AddFields(fields ...FieldError) {
	c.fields = append(c.fields, fields...)
}

func (c *Collector) AddError(field string, err error, value any) {
	if err == nil {
		return
	}

	c.Add(field, err.Error(), value)
}

func (c *Collector) AddValidation(err error) bool {
	fields, ok := FieldErrors(err)
	if !ok {
		return false
	}

	c.AddFields(fields...)
	return true
}

func (c *Collector) HasErrors() bool {
	return len(c.fields) > 0
}

func (c *Collector) Err(base error) error {
	if !c.HasErrors() {
		return nil
	}

	return NewFields(base, c.fields...)
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
