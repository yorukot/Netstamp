package proberuntime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/netip"
	"time"

	appproberuntime "github.com/yorukot/netstamp/internal/controller/application/proberuntime"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

const maxIPFamilyCapabilitiesBodyBytes = 4096

func (h *Handler) handleUpdateIPFamilyCapabilities(w http.ResponseWriter, r *http.Request) {
	body, bodyPresent, err := decodeOptionalIPFamilyCapabilitiesBody(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}

	observedIP, _ := h.clientIPResolver.PublicIP(r)
	output, err := h.updateIPFamilyCapabilities(r.Context(), &ipFamilyCapabilitiesInput{
		ProbeID:     httpx.Path(r, "probe_id"),
		Body:        body,
		BodyPresent: bodyPresent,
		ObservedIP:  observedIP,
	})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) updateIPFamilyCapabilities(ctx context.Context, input *ipFamilyCapabilitiesInput) (*ipFamilyCapabilitiesOutput, error) {
	auth, err := requireRuntimeAuthInput(ctx)
	if err != nil {
		return nil, err
	}
	runtimeInput := newIPFamilyCapabilitiesInput(auth, input)
	output, err := h.service.UpdateIPFamilyCapabilities(ctx, runtimeInput)
	if err != nil {
		return nil, mapRuntimeError(err, "update probe IP family capabilities failed")
	}

	return &ipFamilyCapabilitiesOutput{Body: ipFamilyCapabilitiesOutputBody{ServerTime: output.ServerTime}}, nil
}

type ipFamilyCapabilitiesInput struct {
	ProbeID     string
	Body        ipFamilyCapabilitiesBody
	BodyPresent bool
	ObservedIP  *netip.Addr
}

type ipFamilyCapabilitiesBody struct {
	Families []string `json:"families,omitempty"`
}

type ipFamilyCapabilitiesOutput struct {
	Body ipFamilyCapabilitiesOutputBody
}

type ipFamilyCapabilitiesOutputBody struct {
	ServerTime time.Time `json:"serverTime"`
}

func newIPFamilyCapabilitiesInput(auth appproberuntime.RuntimeAuthInput, input *ipFamilyCapabilitiesInput) appproberuntime.IPFamilyCapabilitiesInput {
	return appproberuntime.IPFamilyCapabilitiesInput{
		RuntimeAuthInput: auth,
		BodyPresent:      input.BodyPresent,
		ObservedIP:       input.ObservedIP,
		Families:         append([]string(nil), input.Body.Families...),
	}
}

func decodeOptionalIPFamilyCapabilitiesBody(r *http.Request) (ipFamilyCapabilitiesBody, bool, error) {
	if r.Body == nil {
		return ipFamilyCapabilitiesBody{}, false, nil
	}

	data, err := io.ReadAll(io.LimitReader(r.Body, maxIPFamilyCapabilitiesBodyBytes+1))
	if err != nil {
		return ipFamilyCapabilitiesBody{}, false, httpx.BadRequest("invalid request body")
	}
	if len(data) > maxIPFamilyCapabilitiesBodyBytes {
		return ipFamilyCapabilitiesBody{}, false, httpx.BadRequest("request body is too large")
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return ipFamilyCapabilitiesBody{}, false, nil
	}

	var body ipFamilyCapabilitiesBody
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		return ipFamilyCapabilitiesBody{}, false, httpx.BadRequest("invalid request body")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return ipFamilyCapabilitiesBody{}, false, httpx.BadRequest("request body must contain a single JSON value")
	}

	return body, true, nil
}
