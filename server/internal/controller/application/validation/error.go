package validation

import (
	"errors"
	"reflect"

	"github.com/yorukot/spvalidator"
)

const (
	FieldCodeInvalidValue  = "INVALID_VALUE"
	FieldCodeInvalidFormat = "INVALID_FORMAT"
	FieldCodeInvalidEnum   = "INVALID_ENUM"
	FieldCodeRequired      = "REQUIRED"
	FieldCodeValueTooShort = "VALUE_TOO_SHORT"
	FieldCodeValueTooLong  = "VALUE_TOO_LONG"
	FieldCodeValueTooSmall = "VALUE_TOO_SMALL"
	FieldCodeValueTooLarge = "VALUE_TOO_LARGE"
)

type FieldError struct {
	Field   string
	Code    string
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
		Code:    FieldCodeInvalidValue,
		Message: message,
		Value:   value,
	})
}

func NewCode(base error, field, code, message string, value any) error {
	return NewFields(base, FieldError{
		Field:   field,
		Code:    normalizeFieldCode(code),
		Message: message,
		Value:   value,
	})
}

func NewFields(base error, fields ...FieldError) error {
	copied := make([]FieldError, len(fields))
	copy(copied, fields)
	for i := range copied {
		copied[i].Code = normalizeFieldCode(copied[i].Code)
	}

	return &Error{
		base:   base,
		fields: copied,
	}
}

func (c *Collector) Add(field, message string, value any) {
	c.fields = append(c.fields, FieldError{
		Field:   field,
		Code:    FieldCodeInvalidValue,
		Message: message,
		Value:   value,
	})
}

func (c *Collector) AddCode(field, code, message string, value any) {
	c.fields = append(c.fields, FieldError{
		Field:   field,
		Code:    normalizeFieldCode(code),
		Message: message,
		Value:   value,
	})
}

func (c *Collector) AddFields(fields ...FieldError) {
	for _, field := range fields {
		field.Code = normalizeFieldCode(field.Code)
		c.fields = append(c.fields, field)
	}
}

func (c *Collector) AddError(field string, err error, value any) {
	if err == nil {
		return
	}

	c.AddCode(field, FieldCodeFromError(err), err.Error(), value)
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

func FieldCodeFromError(err error) string {
	var validationErr *spvalidator.ValidationError
	if !errors.As(err, &validationErr) {
		return FieldCodeInvalidValue
	}

	switch validationErr.Tag {
	case "required", "required_if", "required_unless", "required_with", "required_with_all", "required_without", "required_without_all":
		return FieldCodeRequired
	case "min":
		if isLengthValidationValue(validationErr.Value) {
			return FieldCodeValueTooShort
		}
		return FieldCodeValueTooSmall
	case "max":
		if isLengthValidationValue(validationErr.Value) {
			return FieldCodeValueTooLong
		}
		return FieldCodeValueTooLarge
	case "oneof":
		return FieldCodeInvalidEnum
	case "email", "uuid", "uuid3", "uuid3_rfc4122", "uuid4", "uuid4_rfc4122", "uuid5", "uuid5_rfc4122", "uuid_rfc4122",
		"ipv4", "ipv6", "ip", "ip_addr", "ip4_addr", "ip6_addr", "url", "url_encoded", "hostname", "hostname_rfc1123", "hostname_port",
		"alpha", "alpha_space", "alphanum", "alphanum_space", "alphanumunicode", "alphaunicode", "numeric", "mimetype", "dirpath", "filepath":
		return FieldCodeInvalidFormat
	default:
		return FieldCodeInvalidValue
	}
}

func normalizeFieldCode(code string) string {
	if code == "" {
		return FieldCodeInvalidValue
	}
	return code
}

func isLengthValidationValue(value any) bool {
	if value == nil {
		return false
	}

	v := reflect.ValueOf(value)
	for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return true
	default:
		return false
	}
}
