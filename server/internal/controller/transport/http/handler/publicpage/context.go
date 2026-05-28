package publicpage

import (
	"context"
	"encoding/json"
	"errors"

	apppublicpage "github.com/yorukot/netstamp/internal/controller/application/publicpage"
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	"github.com/yorukot/netstamp/internal/domain/identity"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainpublicpage "github.com/yorukot/netstamp/internal/domain/publicpage"
)

type nullableString struct {
	Value *string
	Set   bool
}

func (v *nullableString) UnmarshalJSON(data []byte) error {
	v.Set = true
	if string(data) == "null" {
		v.Value = nil
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	v.Value = &value
	return nil
}

func currentUserID(ctx context.Context) (string, error) {
	claims, ok := httpmiddleware.AccessTokenClaimsFromContext(ctx)
	if !ok || claims.Subject == "" {
		return "", httpx.Unauthorized("missing auth cookie")
	}

	return claims.Subject, nil
}

func optionalInt64(value int64) *int64 {
	if value == 0 {
		return nil
	}
	return &value
}

func optionalInt32(value int32) *int32 {
	if value == 0 {
		return nil
	}
	return &value
}

func mapPublicPageError(err error, fallback string) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound),
		errors.Is(err, domainproject.ErrMemberNotFound),
		errors.Is(err, identity.ErrUserNotFound),
		errors.Is(err, domainpublicpage.ErrPageNotFound),
		errors.Is(err, domainpublicpage.ErrFolderNotFound),
		errors.Is(err, domainpublicpage.ErrCheckNotPublished),
		errors.Is(err, domaincheck.ErrCheckNotFound),
		errors.Is(err, domainprobe.ErrProbeNotFound):
		return httpx.NotFound("not found")
	case errors.Is(err, apppublicpage.ErrForbidden):
		return httpx.Forbidden("forbidden")
	case errors.Is(err, domainpublicpage.ErrDuplicateSlug):
		return httpx.Conflict("public page slug already exists")
	case errors.Is(err, domainpublicpage.ErrCheckAlreadyPublished):
		return httpx.Conflict("check is already published on this public page")
	case errors.Is(err, apppublicpage.ErrInvalidInput), errors.Is(err, domainpublicpage.ErrInvalidInput), errors.Is(err, domaincheck.ErrInvalidInput), errors.Is(err, domainprobe.ErrInvalidInput):
		return invalidPublicPageInputError(err)
	default:
		return httpx.InternalServerError(fallback)
	}
}

func invalidPublicPageInputError(err error) error {
	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		return httpx.UnprocessableEntity("invalid public page input")
	}

	details := make([]httpx.ErrorDetail, 0, len(fieldErrors))
	for _, fieldErr := range fieldErrors {
		details = append(details, httpx.ErrorDetail{
			Message:  fieldErr.Message,
			Location: publicPageErrorLocation(fieldErr.Field),
			Value:    fieldErr.Value,
		})
	}

	return httpx.UnprocessableEntity("invalid public page input", details...)
}

func publicPageErrorLocation(field string) string {
	switch field {
	case "":
		return "body"
	case "projectRef":
		return "path.ref"
	case "pageId":
		return "path.page_id"
	case "folderId":
		return "path.folder_id"
	case "slug":
		return "path.slug"
	case "probeId", "checkId", "from", "to", "maxDataPoints":
		return "query." + field
	default:
		return "body." + field
	}
}
