package httpserver

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const immutableAssetCache = "public, max-age=31536000, immutable"

type frontendHandler struct {
	dir        string
	fileServer http.Handler
}

func routeFrontend(next http.Handler, dep Dependencies) http.Handler {
	webDir := strings.TrimSpace(dep.WebDir)
	if webDir == "" {
		return next
	}

	frontend := &frontendHandler{
		dir:        webDir,
		fileServer: http.FileServer(http.Dir(webDir)),
	}
	apiBasePath := dep.basePath()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			r = cloneRequestPath(r, apiBasePath+"/healthz")
			next.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}
		if isAPIPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		frontend.ServeHTTP(w, r)
	})
}

func cloneRequestPath(r *http.Request, requestPath string) *http.Request {
	clone := r.Clone(r.Context())
	clone.URL.Path = requestPath
	clone.URL.RawPath = ""
	return clone
}

func isAPIPath(requestPath string) bool {
	return requestPath == "/api" || strings.HasPrefix(requestPath, "/api/")
}

func (h *frontendHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}

	requestPath := path.Clean("/" + r.URL.Path)
	if requestPath == "/" || isSPARoute(requestPath) {
		h.serveIndex(w, r)
		return
	}

	h.serveStatic(w, r, requestPath, strings.HasPrefix(requestPath, "/assets/"))
}

func isSPARoute(requestPath string) bool {
	if strings.HasPrefix(requestPath, "/assets/") {
		return false
	}
	return path.Ext(requestPath) == ""
}

func (h *frontendHandler) serveIndex(w http.ResponseWriter, r *http.Request) {
	// #nosec G703 -- h.dir is trusted configuration and index.html is a fixed file name.
	http.ServeFile(w, r, filepath.Join(h.dir, "index.html"))
}

func (h *frontendHandler) serveStatic(w http.ResponseWriter, r *http.Request, requestPath string, immutableCache bool) {
	name := strings.TrimPrefix(requestPath, "/")
	filePath := filepath.Join(h.dir, filepath.FromSlash(name))

	// #nosec G703 -- requestPath is path.Clean-normalized and resolved under the configured WEB_DIR.
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}

	if immutableCache {
		w.Header().Set("Cache-Control", immutableAssetCache)
	}
	r = cloneRequestPath(r, requestPath)
	h.fileServer.ServeHTTP(w, r)
}
