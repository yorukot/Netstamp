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
		LabelIDs:        input.Body.LabelIDs,

		PingConfig:       input.Body.PingConfig.appInput(),
		TracerouteConfig: input.Body.TracerouteConfig.appInput(),
	})
	if err != nil {
		return nil, mapCheckError(err, "create check failed")
	}

	return &checkOutput{Body: checkOutputBody{Check: check}}, nil
}

type createCheckInput struct {
	Ref  string `path:"ref" minLength:"1" maxLength:"64" pattern:"^[a-z0-9-]+$" patternDescription:"lowercase letters, numbers, and dashes" doc:"Project slug or lowercase UUID." example:"engineering"`
	Body createCheckInputBody
}

type createCheckInputBody struct {
	Name            string         `json:"name" minLength:"1" maxLength:"128" doc:"Check display name." example:"api-latency"`
	Type            string         `json:"type" enum:"ping,traceroute" doc:"Check type." example:"ping"`
	Target          string         `json:"target" minLength:"1" maxLength:"128" doc:"Hostname or address to check." example:"api.netstamp.io"`
	Selector        map[string]any `json:"selector,omitempty" doc:"Selector object for later probe matching."`
	Description     *string        `json:"description,omitempty" minLength:"1" maxLength:"1024" doc:"Optional check description." example:"Latency and loss to controller API."`
	IntervalSeconds int32          `json:"intervalSeconds" minimum:"1" doc:"Check interval in seconds." example:"30"`
	LabelIDs        []string       `json:"labelIds,omitempty" format:"uuid" doc:"Existing project label IDs to attach to the check."`

	PingConfig       *checkPingConfigInput       `json:"pingConfig,omitempty" doc:"Ping-specific check configuration. Omitted fields use defaults."`
	TracerouteConfig *checkTracerouteConfigInput `json:"tracerouteConfig,omitempty" doc:"Traceroute-specific check configuration. Omitted fields use defaults."`
}
