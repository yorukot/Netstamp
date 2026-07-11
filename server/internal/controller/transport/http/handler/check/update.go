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
		TCPConfig:        input.Body.TCPConfig.appInput(),
		TracerouteConfig: input.Body.TracerouteConfig.appInput(),
		HTTPConfig:       input.Body.HTTPConfig.appInput(),
		LabelIDs:         input.Body.LabelIDs,
	})
	if err != nil {
		return nil, mapCheckError(err, "update check failed")
	}

	return newCheckOutput(check, true), nil
}

type updateCheckInput struct {
	Ref     string
	CheckID string
	Body    updateCheckInputBody
}

type updateCheckInputBody struct {
	Name             *string                     `json:"name,omitempty"`
	Type             *string                     `json:"type,omitempty"`
	Target           *string                     `json:"target,omitempty"`
	Selector         map[string]any              `json:"selector,omitempty"`
	Description      *string                     `json:"description,omitempty"`
	IntervalSeconds  *int32                      `json:"intervalSeconds,omitempty"`
	PingConfig       *checkPingConfigInput       `json:"pingConfig,omitempty"`
	TCPConfig        *checkTCPConfigInput        `json:"tcpConfig,omitempty"`
	TracerouteConfig *checkTracerouteConfigInput `json:"tracerouteConfig,omitempty"`
	HTTPConfig       *checkHTTPConfigInput       `json:"httpConfig,omitempty"`
	LabelIDs         *[]string                   `json:"labelIds,omitempty"`
}
