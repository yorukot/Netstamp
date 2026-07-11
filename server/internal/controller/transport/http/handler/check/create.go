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
		TCPConfig:        input.Body.TCPConfig.appInput(),
		TracerouteConfig: input.Body.TracerouteConfig.appInput(),
		HTTPConfig:       input.Body.HTTPConfig.appInput(),
	})
	if err != nil {
		return nil, mapCheckError(err, "create check failed")
	}

	return newCheckOutput(check, true), nil
}

type createCheckInput struct {
	Ref  string
	Body createCheckInputBody
}

type createCheckInputBody struct {
	Name            string         `json:"name"`
	Type            string         `json:"type"`
	Target          string         `json:"target"`
	Selector        map[string]any `json:"selector,omitempty"`
	Description     *string        `json:"description,omitempty"`
	IntervalSeconds int32          `json:"intervalSeconds"`
	LabelIDs        []string       `json:"labelIds,omitempty"`

	PingConfig       *checkPingConfigInput       `json:"pingConfig,omitempty"`
	TCPConfig        *checkTCPConfigInput        `json:"tcpConfig,omitempty"`
	TracerouteConfig *checkTracerouteConfigInput `json:"tracerouteConfig,omitempty"`
	HTTPConfig       *checkHTTPConfigInput       `json:"httpConfig,omitempty"`
}
