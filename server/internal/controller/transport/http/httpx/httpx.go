package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

type Error struct {
	Status  int
	Code    string
	Detail  string
	Details []ErrorDetail
}

func (e *Error) Error() string {
	return e.Detail
}

type ErrorDetail struct {
	Code     string `json:"code,omitempty"`
	Message  string `json:"message,omitempty"`
	Location string `json:"location,omitempty"`
	Value    any    `json:"value,omitempty"`
}

type ProblemDetails struct {
	Type      string        `json:"type,omitempty"`
	Code      string        `json:"code,omitempty"`
	Title     string        `json:"title,omitempty"`
	Status    int           `json:"status,omitempty"`
	Detail    string        `json:"detail,omitempty"`
	Instance  string        `json:"instance,omitempty"`
	RequestID string        `json:"requestId,omitempty"`
	Errors    []ErrorDetail `json:"errors,omitempty"`
}

const (
	CodeBadRequest         = "BAD_REQUEST"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeForbidden          = "FORBIDDEN"
	CodeNotFound           = "NOT_FOUND"
	CodeConflict           = "CONFLICT"
	CodeRateLimited        = "RATE_LIMITED"
	CodeValidationFailed   = "VALIDATION_FAILED"
	CodeInternalError      = "INTERNAL_ERROR"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	CodeMethodNotAllowed   = "METHOD_NOT_ALLOWED"
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

func BadRequest(detail string) error {
	return NewError(http.StatusBadRequest, detail)
}

func BadRequestCode(code, detail string) error {
	return NewErrorCode(http.StatusBadRequest, code, detail)
}

func Unauthorized(detail string) error {
	return NewError(http.StatusUnauthorized, detail)
}

func UnauthorizedCode(code, detail string) error {
	return NewErrorCode(http.StatusUnauthorized, code, detail)
}

func Forbidden(detail string) error {
	return NewError(http.StatusForbidden, detail)
}

func ForbiddenCode(code, detail string) error {
	return NewErrorCode(http.StatusForbidden, code, detail)
}

func NotFound(detail string) error {
	return NewError(http.StatusNotFound, detail)
}

func NotFoundCode(code, detail string) error {
	return NewErrorCode(http.StatusNotFound, code, detail)
}

func Conflict(detail string) error {
	return NewError(http.StatusConflict, detail)
}

func ConflictCode(code, detail string) error {
	return NewErrorCode(http.StatusConflict, code, detail)
}

func UnprocessableEntity(detail string, details ...ErrorDetail) error {
	return UnprocessableEntityCode(CodeValidationFailed, detail, details...)
}

func UnprocessableEntityCode(code, detail string, details ...ErrorDetail) error {
	return &Error{Status: http.StatusUnprocessableEntity, Code: codeOrDefault(code, http.StatusUnprocessableEntity), Detail: detail, Details: details}
}

func InternalServerError(detail string) error {
	return NewError(http.StatusInternalServerError, detail)
}

func InternalServerErrorCode(code, detail string) error {
	return NewErrorCode(http.StatusInternalServerError, code, detail)
}

func ServiceUnavailable(detail string) error {
	return NewError(http.StatusServiceUnavailable, detail)
}

func ServiceUnavailableCode(code, detail string) error {
	return NewErrorCode(http.StatusServiceUnavailable, code, detail)
}

func NewError(status int, detail string) error {
	return NewErrorCode(status, "", detail)
}

func NewErrorCode(status int, code, detail string) error {
	return &Error{Status: status, Code: codeOrDefault(code, status), Detail: detail}
}

func codeOrDefault(code string, status int) string {
	if code != "" {
		return code
	}
	switch status {
	case http.StatusBadRequest:
		return CodeBadRequest
	case http.StatusUnauthorized:
		return CodeUnauthorized
	case http.StatusForbidden:
		return CodeForbidden
	case http.StatusNotFound:
		return CodeNotFound
	case http.StatusMethodNotAllowed:
		return CodeMethodNotAllowed
	case http.StatusConflict:
		return CodeConflict
	case http.StatusTooManyRequests:
		return CodeRateLimited
	case http.StatusUnprocessableEntity:
		return CodeValidationFailed
	case http.StatusServiceUnavailable:
		return CodeServiceUnavailable
	default:
		if status >= http.StatusInternalServerError {
			return CodeInternalError
		}
		return CodeBadRequest
	}
}

func DecodeJSON(r *http.Request, value any) error {
	if r.Body == nil {
		return BadRequest("request body is required")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return BadRequest("invalid request body")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return BadRequest("request body must contain a single JSON value")
	}

	return nil
}

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		return
	}
}

func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func WriteProblem(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	code := CodeInternalError
	detail := http.StatusText(status)
	var httpErr *Error
	if errors.As(err, &httpErr) {
		status = httpErr.Status
		code = codeOrDefault(httpErr.Code, status)
		detail = httpErr.Detail
	}

	requestID := chimw.GetReqID(r.Context())
	if requestID != "" {
		w.Header().Set("X-Request-ID", requestID)
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	body := ProblemDetails{
		Code:      code,
		Status:    status,
		Title:     http.StatusText(status),
		Detail:    detail,
		RequestID: requestID,
	}
	if httpErr != nil && len(httpErr.Details) > 0 {
		body.Errors = httpErr.Details
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		return
	}
}

func Path(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}

func QueryString(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

func QueryInt64(r *http.Request, name string) (int64, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, BadRequest("invalid query parameter " + name)
	}
	return parsed, nil
}

func QueryInt32(r *http.Request, name string) (int32, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, BadRequest("invalid query parameter " + name)
	}
	return int32(parsed), nil
}
