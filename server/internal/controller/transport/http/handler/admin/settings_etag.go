package admin

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

const settingsCacheControl = "private, no-store"

var errSettingsVersionConflict = errors.New("settings version conflict")

type settingsVersionConflictError struct {
	resource        string
	currentRevision int64
}

func (e *settingsVersionConflictError) Error() string {
	return errSettingsVersionConflict.Error()
}

func (e *settingsVersionConflictError) Unwrap() error {
	return errSettingsVersionConflict
}

func formatSettingsETag(resource string, revision int64) string {
	return `"` + resource + "-" + strconv.FormatInt(revision, 10) + `"`
}

func parseSettingsIfMatch(r *http.Request, resource string) (int64, error) {
	values := r.Header.Values("If-Match")
	if len(values) == 0 {
		return 0, httpx.NewErrorCode(http.StatusPreconditionRequired, httpx.CodePreconditionRequired, "If-Match header is required")
	}
	if len(values) != 1 {
		return 0, httpx.BadRequest("If-Match must contain one strong entity tag")
	}

	value := strings.TrimSpace(values[0])
	if value == "" {
		return 0, httpx.NewErrorCode(http.StatusPreconditionRequired, httpx.CodePreconditionRequired, "If-Match header is required")
	}
	if strings.Contains(value, ",") {
		return 0, httpx.BadRequest("If-Match must contain one strong entity tag")
	}
	if value == "*" || strings.HasPrefix(value, "W/") {
		return 0, errSettingsVersionConflict
	}
	if len(value) < 2 || value[0] != '"' || value[len(value)-1] != '"' {
		return 0, httpx.BadRequest("If-Match must contain a quoted strong entity tag")
	}

	tag := value[1 : len(value)-1]
	prefix := resource + "-"
	if !strings.HasPrefix(tag, prefix) {
		return 0, errSettingsVersionConflict
	}

	revisionText := strings.TrimPrefix(tag, prefix)
	if !canonicalSettingsRevision(revisionText) {
		return 0, httpx.BadRequest("If-Match contains an invalid settings version")
	}
	revision, err := strconv.ParseInt(revisionText, 10, 64)
	if err != nil {
		return 0, httpx.BadRequest("If-Match contains an invalid settings version")
	}
	return revision, nil
}

func canonicalSettingsRevision(value string) bool {
	if value == "" || value[0] < '1' || value[0] > '9' {
		return false
	}
	for i := 1; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}
	return true
}

func resolveSettingsIfMatch(r *http.Request, resource string, currentRevision func() (int64, error)) (int64, error) {
	revision, err := parseSettingsIfMatch(r, resource)
	if !errors.Is(err, errSettingsVersionConflict) {
		return revision, err
	}

	current, readErr := currentRevision()
	if readErr != nil {
		return 0, readErr
	}
	return 0, &settingsVersionConflictError{resource: resource, currentRevision: current}
}

func setVersionedSettingsHeaders(w http.ResponseWriter, resource string, revision int64) {
	w.Header().Set("ETag", formatSettingsETag(resource, revision))
	w.Header().Set("Cache-Control", settingsCacheControl)
}
