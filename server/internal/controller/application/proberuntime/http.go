package proberuntime

import (
	"errors"
	"strings"
	"time"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
)

func normalizeHTTPResult(input HTTPResultInput, fieldPrefix string) (domainhttp.ResultStorageInput, error) {
	var validation appvalidation.Collector
	timing, err := normalizeResultTiming(input.StartedAt, input.FinishedAt, input.DurationMs, fieldPrefix, func(value int32) (int32, error) {
		if value < 0 {
			return 0, errors.New("must be non-negative")
		}
		return value, nil
	})
	if err != nil {
		validation.AddValidation(err)
	}
	status, err := normalizeHTTPStatus(input.Status)
	if err != nil {
		validation.AddError(resultField(fieldPrefix, "status"), err, input.Status)
	}
	dns := normalizeOptionalHTTPDuration(input.DNSDurationMs, resultField(fieldPrefix, "dnsDurationMs"), &validation)
	connect := normalizeOptionalHTTPDuration(input.ConnectDurationMs, resultField(fieldPrefix, "connectDurationMs"), &validation)
	tlsDuration := normalizeOptionalHTTPDuration(input.TLSDurationMs, resultField(fieldPrefix, "tlsDurationMs"), &validation)
	ttfb := normalizeOptionalHTTPDuration(input.TTFBDurationMs, resultField(fieldPrefix, "ttfbDurationMs"), &validation)
	if input.StatusCode != nil && (*input.StatusCode < 100 || *input.StatusCode > 599) {
		validation.Add(resultField(fieldPrefix, "statusCode"), "must be between 100 and 599", input.StatusCode)
	}
	if input.RedirectCount < 0 || input.RedirectCount > domainhttp.MaxRedirects {
		validation.Add(resultField(fieldPrefix, "redirectCount"), "must be between 0 and 10", input.RedirectCount)
	}
	if input.ResponseBytes != nil && *input.ResponseBytes < 0 {
		validation.Add(resultField(fieldPrefix, "responseBytes"), "must be non-negative", input.ResponseBytes)
	}
	if input.FinalURL != nil {
		value, targetErr := domainhttp.VNTarget(*input.FinalURL)
		if targetErr != nil {
			validation.AddError(resultField(fieldPrefix, "finalUrl"), targetErr, input.FinalURL)
		} else {
			input.FinalURL = &value
		}
	}
	metadata, metadataErr := normalizeResultMetadata(input.IPFamily, input.ErrorCode, input.ErrorMessage, fieldPrefix, normalizeOptionalHTTPText)
	if metadataErr != nil {
		validation.AddValidation(metadataErr)
	}
	validateCertificateTimes(input.CertificateNotBefore, input.CertificateNotAfter, fieldPrefix, &validation)
	if err := validation.Err(ErrInvalidInput); err != nil {
		return domainhttp.ResultStorageInput{}, err
	}
	return domainhttp.ResultStorageInput{Result: domainhttp.Result{
		StartedAt: timing.startedAt, FinishedAt: timing.finishedAt, DurationMs: timing.durationMs,
		Status: status, DNSDurationMs: dns, ConnectDurationMs: connect, TLSDurationMs: tlsDuration,
		TTFBDurationMs: ttfb, ResolvedIP: cloneAddr(input.ResolvedIP), IPFamily: metadata.ipFamily,
		StatusCode: input.StatusCode, FinalURL: input.FinalURL, RedirectCount: input.RedirectCount,
		ResponseBytes: input.ResponseBytes, ResponseTruncated: input.ResponseTruncated,
		BodyMatched: input.BodyMatched, TLSVersion: input.TLSVersion, TLSCipherSuite: input.TLSCipherSuite,
		CertificateNotBefore: utcTimePtr(input.CertificateNotBefore), CertificateNotAfter: utcTimePtr(input.CertificateNotAfter),
		ErrorCode: metadata.errorCode, ErrorMessage: metadata.errorMessage,
	}}, nil
}

func normalizeHTTPStatus(value string) (domainhttp.Status, error) {
	switch domainhttp.Status(strings.TrimSpace(value)) {
	case domainhttp.StatusSuccessful:
		return domainhttp.StatusSuccessful, nil
	case domainhttp.StatusTimeout:
		return domainhttp.StatusTimeout, nil
	case domainhttp.StatusError:
		return domainhttp.StatusError, nil
	default:
		return "", errors.New("invalid http status")
	}
}

func normalizeOptionalHTTPDuration(value *float64, field string, validation *appvalidation.Collector) *float64 {
	if value == nil {
		return nil
	}
	if *value < 0 {
		validation.Add(field, "must be non-negative", value)
		return nil
	}
	normalized := *value
	return &normalized
}

func normalizeOptionalHTTPText(input *string, field string) (*string, error) {
	if input == nil {
		return nil, nil //nolint:nilnil // Nil input means the optional result metadata is absent.
	}
	value := strings.TrimSpace(*input)
	if value == "" {
		return nil, invalidRuntimeField(field, "must not be empty", input)
	}
	return &value, nil
}

func validateCertificateTimes(notBefore, notAfter *time.Time, prefix string, validation *appvalidation.Collector) {
	if notBefore != nil && notBefore.IsZero() {
		validation.Add(resultField(prefix, "certificateNotBefore"), "must be a valid timestamp", notBefore)
	}
	if notAfter != nil && notAfter.IsZero() {
		validation.Add(resultField(prefix, "certificateNotAfter"), "must be a valid timestamp", notAfter)
	}
	if notBefore != nil && notAfter != nil && notAfter.Before(*notBefore) {
		validation.Add(resultField(prefix, "certificateNotAfter"), "must not be before certificateNotBefore", notAfter)
	}
}

func utcTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}
