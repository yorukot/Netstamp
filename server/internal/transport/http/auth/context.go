package auth

import (
	"github.com/danielgtaylor/huma/v2"

	appvalidation "github.com/yorukot/netstamp/internal/application/validation"
)

func invalidAuthInputError(err error) error {
	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		return huma.Error422UnprocessableEntity("invalid auth input")
	}

	details := make([]error, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		details = append(details, &huma.ErrorDetail{
			Message:  fieldErr.Message,
			Location: authErrorLocation(fieldErr.Field),
			Value:    fieldErr.Value,
		})
	}

	return huma.Error422UnprocessableEntity("invalid auth input", details...)
}

func authErrorLocation(field string) string {
	if field == "" {
		return "body"
	}

	return "body." + field
}
