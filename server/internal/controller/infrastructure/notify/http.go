package notify

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	appnotification "github.com/yorukot/netstamp/internal/controller/application/notification"
)

type JSONPoster struct {
	client *http.Client
}

func NewJSONPoster(timeout time.Duration) *JSONPoster {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &JSONPoster{
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, _ []*http.Request) error {
				return validateWebhookTarget(req.Context(), req.URL.String())
			},
		},
	}
}

func (p *JSONPoster) PostJSON(ctx context.Context, endpoint string, body []byte, targetName string) appnotification.DeliveryResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return permanent("request", "invalid_request", "invalid "+targetName+" request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "netstamp-alerts/1")

	resp, err := p.client.Do(req)
	if err != nil {
		return retryable("network", "request_failed", targetName+" request failed")
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return appnotification.DeliveryResult{Delivered: true}
	case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusRequestTimeout || resp.StatusCode >= 500:
		return retryable("http", fmt.Sprintf("status_%d", resp.StatusCode), targetName+" returned retryable status")
	default:
		return permanent("http", fmt.Sprintf("status_%d", resp.StatusCode), targetName+" returned permanent status")
	}
}
