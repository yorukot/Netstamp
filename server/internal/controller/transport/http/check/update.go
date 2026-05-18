package check

import (
	"context"

	appcheck "github.com/yorukot/netstamp/internal/controller/application/check"
)

func (h *Handler) updateCheck(ctx context.Context, input *updateCheckInput) (*checkOutput, error) {
	currentUserID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	check, err := h.service.UpdateCheck(ctx, appcheck.UpdateCheckInput{
		CurrentUserID:    currentUserID,
		ProjectRef:       input.Ref,
		CheckID:          input.CheckID,
		Name:             input.Body.Name,
		Type:             input.Body.Type,
		Target:           input.Body.Target,
		Selector:         input.Body.Selector,
		Description:      input.Body.Description,
		IntervalSeconds:  input.Body.IntervalSeconds,
		PingConfig:       input.Body.PingConfig.appInput(),
		TracerouteConfig: input.Body.TracerouteConfig.appInput(),
		LabelIDs:         input.Body.LabelIDs,
	})
	if err != nil {
		return nil, mapCheckError(err, "update check failed")
	}

	return &checkOutput{Body: checkOutputBody{Check: check}}, nil
}

type updateCheckInput struct {
	Ref     string `path:"ref" minLength:"1" maxLength:"64" pattern:"^[a-z0-9-]+$" patternDescription:"lowercase letters, numbers, and dashes" doc:"Project slug or lowercase UUID." example:"engineering"`
	CheckID string `path:"check_id" minLength:"1" format:"uuid" doc:"Check ID." example:"33333333-3333-3333-3333-333333333333"`
	Body    updateCheckInputBody
}

type updateCheckInputBody struct {
	Name             *string                     `json:"name,omitempty" minLength:"1" maxLength:"128" doc:"Check display name." example:"api-latency"`
	Type             *string                     `json:"type,omitempty" enum:"ping,traceroute" doc:"Check type. Cannot be changed after creation in v1." example:"ping"`
	Target           *string                     `json:"target,omitempty" minLength:"1" maxLength:"128" doc:"Hostname or address to check." example:"api.netstamp.io"`
	Selector         map[string]any              `json:"selector,omitempty" doc:"Selector object for later probe matching."`
	Description      *string                     `json:"description,omitempty" minLength:"1" maxLength:"1024" doc:"Optional check description." example:"Latency and loss to controller API."`
	IntervalSeconds  *int32                      `json:"intervalSeconds,omitempty" minimum:"1" doc:"Check interval in seconds." example:"30"`
	PingConfig       *checkPingConfigInput       `json:"pingConfig,omitempty" doc:"Ping-specific check configuration fields to update."`
	TracerouteConfig *checkTracerouteConfigInput `json:"tracerouteConfig,omitempty" doc:"Traceroute-specific check configuration fields to update."`
	LabelIDs         *[]string                   `json:"labelIds,omitempty" format:"uuid" doc:"Replacement project label IDs for the check."`
}
