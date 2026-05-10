package check

import (
	"context"

	appcheck "github.com/yorukot/netstamp/internal/controller/application/check"
)

func (h *Handler) createCheck(ctx context.Context, input *createCheckInput) (*checkOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	check, err := h.service.CreateCheck(ctx, appcheck.CreateCheckInput{
		CurrentUserID:   currentUserID,
		ProjectRef:      input.Ref,
		Name:            input.Body.Name,
		Type:            input.Body.Type,
		Target:          input.Body.Target,
		Selector:        input.Body.Selector,
		Description:     input.Body.Description,
		IntervalSeconds: input.Body.IntervalSeconds,
		PacketCount:     input.Body.PacketCount,
		PacketSizeBytes: input.Body.PacketSizeBytes,
		TimeoutMs:       input.Body.TimeoutMs,
		IPFamily:        input.Body.IPFamily,
		LabelIDs:        input.Body.LabelIDs,
	})
	if err != nil {
		return nil, mapCheckError(err, "create check failed")
	}

	return &checkOutput{Body: checkOutputBody{Check: newCheckResponse(check)}}, nil
}

type createCheckInput struct {
	Ref  string `path:"ref" doc:"Project UUID or slug." example:"engineering"`
	Body createCheckInputBody
}

type createCheckInputBody struct {
	Name            string         `json:"name" doc:"Check display name." example:"api-latency"`
	Type            string         `json:"type" doc:"Check type. Only ping is supported in v1." example:"ping"`
	Target          string         `json:"target" doc:"Hostname or address to check." example:"api.netstamp.io"`
	Selector        map[string]any `json:"selector,omitempty" doc:"Selector object for later probe matching."`
	Description     *string        `json:"description,omitempty" doc:"Optional check description." example:"Latency and loss to controller API."`
	IntervalSeconds int32          `json:"intervalSeconds" doc:"Check interval in seconds." example:"30"`
	PacketCount     *int32         `json:"packetCount,omitempty" doc:"ICMP packet count. Defaults to 4." example:"4"`
	PacketSizeBytes *int32         `json:"packetSizeBytes,omitempty" doc:"ICMP payload size in bytes. Defaults to 56." example:"56"`
	TimeoutMs       *int32         `json:"timeoutMs,omitempty" doc:"Ping timeout in milliseconds. Defaults to 3000." example:"3000"`
	IPFamily        *string        `json:"ipFamily,omitempty" doc:"Optional IP family preference." example:"inet"`
	LabelIDs        []string       `json:"labelIds,omitempty" doc:"Existing project label IDs to attach to the check."`
}
