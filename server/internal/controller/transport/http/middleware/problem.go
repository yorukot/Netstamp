package middleware

import (
	"net/http"

	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

func WriteProblem(w http.ResponseWriter, r *http.Request, status int, detail string) {
	httpx.WriteProblem(w, r, httpx.NewError(status, detail))
}

func WriteProblemCode(w http.ResponseWriter, r *http.Request, status int, code, detail string) {
	httpx.WriteProblem(w, r, httpx.NewErrorCode(status, code, detail))
}
