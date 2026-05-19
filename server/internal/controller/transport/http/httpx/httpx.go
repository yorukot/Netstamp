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
	Detail  string
	Details []ErrorDetail
}

func (e *Error) Error() string {
	return e.Detail
}

type ErrorDetail struct {
	Message  string `json:"message,omitempty"`
	Location string `json:"location,omitempty"`
	Value    any    `json:"value,omitempty"`
}

type ProblemDetails struct {
	Type     string        `json:"type,omitempty"`
	Title    string        `json:"title,omitempty"`
	Status   int           `json:"status,omitempty"`
	Detail   string        `json:"detail,omitempty"`
	Instance string        `json:"instance,omitempty"`
	Errors   []ErrorDetail `json:"errors,omitempty"`
}

func BadRequest(detail string) error {
	return NewError(http.StatusBadRequest, detail)
}

func Unauthorized(detail string) error {
	return NewError(http.StatusUnauthorized, detail)
}

func Forbidden(detail string) error {
	return NewError(http.StatusForbidden, detail)
}

func NotFound(detail string) error {
	return NewError(http.StatusNotFound, detail)
}

func Conflict(detail string) error {
	return NewError(http.StatusConflict, detail)
}

func UnprocessableEntity(detail string, details ...ErrorDetail) error {
	return &Error{Status: http.StatusUnprocessableEntity, Detail: detail, Details: details}
}

func InternalServerError(detail string) error {
	return NewError(http.StatusInternalServerError, detail)
}

func ServiceUnavailable(detail string) error {
	return NewError(http.StatusServiceUnavailable, detail)
}

func NewError(status int, detail string) error {
	return &Error{Status: status, Detail: detail}
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
	detail := http.StatusText(status)
	var httpErr *Error
	if errors.As(err, &httpErr) {
		status = httpErr.Status
		detail = httpErr.Detail
	}

	if requestID := chimw.GetReqID(r.Context()); requestID != "" {
		w.Header().Set("X-Request-ID", requestID)
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	body := ProblemDetails{
		Status: status,
		Title:  http.StatusText(status),
		Detail: detail,
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
