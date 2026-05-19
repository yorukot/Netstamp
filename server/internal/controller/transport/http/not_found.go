package httpserver

import (
	"net/http"
	"strings"

	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

func writeNotFound(w http.ResponseWriter, r *http.Request) {
	if wantsHTML(r) {
		http.NotFound(w, r)
		return
	}
	httpx.WriteProblem(w, r, httpx.NotFound("route not found"))
}

func writeMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	if wantsHTML(r) {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}
	httpx.WriteProblem(w, r, httpx.NewError(http.StatusMethodNotAllowed, "method not allowed"))
}

func wantsHTML(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "text/html")
}
