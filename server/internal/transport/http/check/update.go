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
	Ref     string `path:"ref" minLength:"1" maxLength:"100" doc:"Project UUID or slug." example:"engineering"`
	CheckID string `path:"check_id" format:"uuid" doc:"Check ID." example:"33333333-3333-3333-3333-333333333333"`
	Body    updateCheckInputBody
}

type updateCheckInputBody struct {
	Name            *string        `json:"name,omitempty" minLength:"1" maxLength:"100" doc:"Check display name." example:"api-latency"`
	Type            *string        `json:"type,omitempty" enum:"ping" doc:"Check type. Only ping is supported in v1." example:"ping"`
	Target          *string        `json:"target,omitempty" minLength:"1" maxLength:"255" doc:"Hostname or address to check." example:"api.netstamp.io"`
	Selector        map[string]any `json:"selector,omitempty" doc:"Selector object for later probe matching."`
	Description     *string        `json:"description,omitempty" minLength:"1" maxLength:"500" doc:"Optional check description." example:"Latency and loss to controller API."`
	IntervalSeconds *int32         `json:"intervalSeconds,omitempty" minimum:"1" doc:"Check interval in seconds." example:"30"`
	PacketCount     *int32         `json:"packetCount,omitempty" minimum:"1" doc:"ICMP packet count." example:"4"`
	PacketSizeBytes *int32         `json:"packetSizeBytes,omitempty" minimum:"0" maximum:"65507" doc:"ICMP payload size in bytes." example:"56"`
	TimeoutMs       *int32         `json:"timeoutMs,omitempty" minimum:"1" doc:"Ping timeout in milliseconds." example:"3000"`
	IPFamily        *string        `json:"ipFamily,omitempty" enum:"inet,inet6" doc:"Optional IP family preference." example:"inet"`
	LabelIDs        *[]string      `json:"labelIds,omitempty" doc:"Replacement project label IDs for the check."`
}
