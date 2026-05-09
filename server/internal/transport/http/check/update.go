package check

import (
	"context"

	appcheck "github.com/yorukot/netstamp/internal/application/check"
)

func (h *Handler) updateCheck(ctx context.Context, input *updateCheckInput) (*checkOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	check, err := h.service.UpdateCheck(ctx, appcheck.UpdateCheckInput{
		CurrentUserID:   currentUserID,
		ProjectRef:      input.Ref,
		CheckID:         input.CheckID,
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
		return nil, mapCheckError(err, "update check failed")
	}

	return &checkOutput{Body: checkOutputBody{Check: newCheckResponse(check)}}, nil
}

type updateCheckInput struct {
	Ref     string `path:"ref" doc:"Project UUID or slug." example:"engineering"`
	CheckID string `path:"check_id" doc:"Check ID." example:"33333333-3333-3333-3333-333333333333"`
	Body    updateCheckInputBody
}

type updateCheckInputBody struct {
	Name            *string        `json:"name,omitempty" doc:"Check display name." example:"api-latency"`
	Type            *string        `json:"type,omitempty" doc:"Check type. Only ping is supported in v1." example:"ping"`
	Target          *string        `json:"target,omitempty" doc:"Hostname or address to check." example:"api.netstamp.io"`
	Selector        map[string]any `json:"selector,omitempty" doc:"Selector object for later probe matching."`
	Description     *string        `json:"description,omitempty" doc:"Optional check description." example:"Latency and loss to controller API."`
	IntervalSeconds *int32         `json:"intervalSeconds,omitempty" doc:"Check interval in seconds." example:"30"`
	PacketCount     *int32         `json:"packetCount,omitempty" doc:"ICMP packet count." example:"4"`
	PacketSizeBytes *int32         `json:"packetSizeBytes,omitempty" doc:"ICMP payload size in bytes." example:"56"`
	TimeoutMs       *int32         `json:"timeoutMs,omitempty" doc:"Ping timeout in milliseconds." example:"3000"`
	IPFamily        *string        `json:"ipFamily,omitempty" doc:"Optional IP family preference." example:"inet"`
	LabelIDs        *[]string      `json:"labelIds,omitempty" doc:"Replacement project label IDs for the check."`
}
