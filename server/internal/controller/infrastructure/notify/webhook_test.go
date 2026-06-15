package notify

import (
	"bytes"
	"testing"
)

func TestRenderWebhookBodyKeepsPayloadUnchanged(t *testing.T) {
	payload := []byte(`{"hello":"world"}`)
	if got := renderWebhookBody(payload); !bytes.Equal(got, payload) {
		t.Fatalf("expected webhook payload to stay unchanged, got %q", string(got))
	}
}
