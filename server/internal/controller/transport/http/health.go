package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type rootBody struct {
	Message string `json:"message" doc:"Human-readable API status message." example:"Netstamp API is running"`
}

type rootOutput struct {
	Body rootBody
}

type healthBody struct {
	Status string `json:"status" doc:"Current health status." example:"ready"`
}

type healthOutput struct {
	Body healthBody
}

func registerSystemRoutes(api huma.API, readinessCheck func(context.Context) error) {
	huma.Register(api, huma.Operation{
		OperationID: "getAPIStatus",
		Method:      http.MethodGet,
		Path:        "/",
		Summary:     "Get API status",
		Tags:        []string{"System"},
	}, func(context.Context, *struct{}) (*rootOutput, error) {
		return &rootOutput{Body: rootBody{
			Message: "Netstamp API is running",
		}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "getLiveness",
		Method:      http.MethodGet,
		Path:        "/livez",
		Summary:     "Get liveness status",
		Tags:        []string{"System"},
	}, func(context.Context, *struct{}) (*healthOutput, error) {
		return &healthOutput{Body: healthBody{
			Status: "ok",
		}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "getReadiness",
		Method:      http.MethodGet,
		Path:        "/readyz",
		Summary:     "Get readiness status",
		Tags:        []string{"System"},
		Errors:      []int{http.StatusServiceUnavailable},
	}, func(ctx context.Context, _ *struct{}) (*healthOutput, error) {
		if readinessCheck == nil {
			return &healthOutput{Body: healthBody{
				Status: "ready",
			}}, nil
		}

		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		if err := readinessCheck(ctx); err != nil {
			return nil, huma.Error503ServiceUnavailable("readiness check failed")
		}

		return &healthOutput{Body: healthBody{
			Status: "ready",
		}}, nil
	})
}
