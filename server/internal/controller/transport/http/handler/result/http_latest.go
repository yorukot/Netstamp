package result

import (
	"context"
	"net/netip"
	"time"

	appresult "github.com/yorukot/netstamp/internal/controller/application/result"
	domainhttp "github.com/yorukot/netstamp/internal/domain/httpcheck"
)

func (h *Handler) queryLatestHTTPResults(ctx context.Context, input *queryLatestHTTPResultsInput) (*queryLatestHTTPResultsOutput, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	output, err := h.service.QueryLatestHTTPResults(ctx, appresult.QueryLatestHTTPResultsInput{
		CurrentUserID: userID,
		ProjectRef:    input.Ref,
		ProbeID:       input.ProbeID,
		CheckID:       input.CheckID,
	})
	if err != nil {
		return nil, mapResultError(err, "query latest HTTP results failed")
	}

	return &queryLatestHTTPResultsOutput{Body: newQueryLatestHTTPResultsBody(output)}, nil
}

type queryLatestHTTPResultsInput struct {
	Ref     string
	ProbeID string
	CheckID string
}

type queryLatestHTTPResultsOutput struct {
	Body queryLatestHTTPResultsBody
}

type queryLatestHTTPResultsBody struct {
	Results []latestHTTPResultBody `json:"results"`
}

type latestHTTPResultBody struct {
	ProbeID string         `json:"probeId"`
	CheckID string         `json:"checkId"`
	Result  httpResultBody `json:"result"`
}

type httpResultBody struct {
	StartedAt            time.Time   `json:"startedAt"`
	FinishedAt           time.Time   `json:"finishedAt"`
	DurationMs           int32       `json:"durationMs"`
	Status               string      `json:"status"`
	DNSDurationMs        *float64    `json:"dnsDurationMs,omitempty"`
	ConnectDurationMs    *float64    `json:"connectDurationMs,omitempty"`
	TLSDurationMs        *float64    `json:"tlsDurationMs,omitempty"`
	TTFBDurationMs       *float64    `json:"ttfbDurationMs,omitempty"`
	ResolvedIP           *netip.Addr `json:"resolvedIp,omitempty"`
	IPFamily             *string     `json:"ipFamily,omitempty"`
	StatusCode           *int32      `json:"statusCode,omitempty"`
	FinalURL             *string     `json:"finalUrl,omitempty"`
	RedirectCount        int32       `json:"redirectCount"`
	ResponseBytes        *int64      `json:"responseBytes,omitempty"`
	ResponseTruncated    bool        `json:"responseTruncated"`
	BodyMatched          *bool       `json:"bodyMatched,omitempty"`
	TLSVersion           *string     `json:"tlsVersion,omitempty"`
	TLSCipherSuite       *string     `json:"tlsCipherSuite,omitempty"`
	CertificateNotBefore *time.Time  `json:"certificateNotBefore,omitempty"`
	CertificateNotAfter  *time.Time  `json:"certificateNotAfter,omitempty"`
	ErrorCode            *string     `json:"errorCode,omitempty"`
	ErrorMessage         *string     `json:"errorMessage,omitempty"`
}

func newQueryLatestHTTPResultsBody(output appresult.LatestHTTPResultsOutput) queryLatestHTTPResultsBody {
	results := make([]latestHTTPResultBody, 0, len(output.Results))
	for _, latest := range output.Results {
		results = append(results, latestHTTPResultBody{
			ProbeID: latest.ProbeID,
			CheckID: latest.CheckID,
			Result:  newHTTPResultBody(latest.Result),
		})
	}
	return queryLatestHTTPResultsBody{Results: results}
}

func newHTTPResultBody(result domainhttp.Result) httpResultBody {
	return httpResultBody{
		StartedAt:            result.StartedAt,
		FinishedAt:           result.FinishedAt,
		DurationMs:           result.DurationMs,
		Status:               string(result.Status),
		DNSDurationMs:        result.DNSDurationMs,
		ConnectDurationMs:    result.ConnectDurationMs,
		TLSDurationMs:        result.TLSDurationMs,
		TTFBDurationMs:       result.TTFBDurationMs,
		ResolvedIP:           result.ResolvedIP,
		IPFamily:             stringPointer(result.IPFamily),
		StatusCode:           result.StatusCode,
		FinalURL:             result.FinalURL,
		RedirectCount:        result.RedirectCount,
		ResponseBytes:        result.ResponseBytes,
		ResponseTruncated:    result.ResponseTruncated,
		BodyMatched:          result.BodyMatched,
		TLSVersion:           result.TLSVersion,
		TLSCipherSuite:       result.TLSCipherSuite,
		CertificateNotBefore: result.CertificateNotBefore,
		CertificateNotAfter:  result.CertificateNotAfter,
		ErrorCode:            result.ErrorCode,
		ErrorMessage:         result.ErrorMessage,
	}
}

func stringPointer[T ~string](value *T) *string {
	if value == nil {
		return nil
	}
	output := string(*value)
	return &output
}
