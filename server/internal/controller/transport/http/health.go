package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

type rootBody struct {
	Message string `json:"message"`
}

type healthBody struct {
	Status string `json:"status"`
}

func registerSystemRoutes(api chi.Router, readinessCheck func(context.Context) error) {
	api.Get("/", func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, rootBody{
			Message: "Netstamp API is running",
		})
	})

	api.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if readinessCheck == nil {
			httpx.WriteJSON(w, http.StatusOK, healthBody{
				Status: "ok",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := readinessCheck(ctx); err != nil {
			httpx.WriteProblem(w, r, httpx.ServiceUnavailable("readiness check failed"))
			return
		}

		httpx.WriteJSON(w, http.StatusOK, healthBody{
			Status: "ok",
		})
	})
}
