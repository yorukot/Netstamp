package install

import (
	"context"
	"embed"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

const (
	DefaultAgentBinaryDir = "/app/agents"
)

type agentBinaryAsset struct {
	filename string
}

var agentBinaryAssets = []agentBinaryAsset{
	{filename: "netstamp-agent-linux-amd64"},
	{filename: "netstamp-agent-linux-arm64"},
}

//go:embed agent.sh uninstall-agent.sh
var installerFiles embed.FS

type Handler struct {
	agentBinaryDir         string
	backendBaseURL         string
	backendBaseURLProvider BackendBaseURLProvider
	apiBasePath            string
}

type BackendBaseURLProvider interface {
	BackendBaseURL(ctx context.Context) (string, error)
}

func NewHandler(agentBinaryDir, backendBaseURL, apiBasePath string, providers ...BackendBaseURLProvider) *Handler {
	if agentBinaryDir == "" {
		agentBinaryDir = DefaultAgentBinaryDir
	}

	handler := &Handler{
		agentBinaryDir: agentBinaryDir,
		backendBaseURL: strings.TrimRight(backendBaseURL, "/"),
		apiBasePath:    "/" + strings.Trim(strings.TrimSpace(apiBasePath), "/"),
	}
	if len(providers) > 0 {
		handler.backendBaseURLProvider = providers[0]
	}
	return handler
}

func (h *Handler) RegisterRoutes(api chi.Router) {
	api.Get("/install/agent.sh", h.handleAgentScript)
	api.Get("/install/uninstall-agent.sh", h.handleAgentUninstallScript)
	for _, asset := range agentBinaryAssets {
		api.Get("/install/"+asset.filename, h.handleAgentBinary(asset))
	}
}

func (h *Handler) handleAgentScript(w http.ResponseWriter, r *http.Request) {
	h.writeScript(w, r, "agent.sh", "agent installer unavailable")
}

func (h *Handler) handleAgentUninstallScript(w http.ResponseWriter, r *http.Request) {
	h.writeScript(w, r, "uninstall-agent.sh", "agent uninstaller unavailable")
}

func (h *Handler) writeScript(
	w http.ResponseWriter,
	r *http.Request,
	name string,
	unavailableDetail string,
) {
	data, err := installerFiles.ReadFile(name)
	if err != nil {
		httpx.WriteProblem(w, r, httpx.InternalServerError(unavailableDetail))
		return
	}
	if name == "agent.sh" {
		data = h.renderAgentScript(r, data)
	}

	w.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	// #nosec G705 -- agent.sh is an embedded script; reflected origin values are sanitized before replacement.
	if _, err := w.Write(data); err != nil {
		return
	}
}

func (h *Handler) renderAgentScript(r *http.Request, data []byte) []byte {
	controllerURL := h.controllerURL(r)
	binaryBaseURL := controllerURL + h.apiBasePath + "/install"

	body := string(data)
	body = strings.ReplaceAll(body, "__NETSTAMP_CONTROLLER_URL__", controllerURL)
	body = strings.ReplaceAll(body, "__NETSTAMP_AGENT_BINARY_BASE_URL__", binaryBaseURL)
	return []byte(body)
}

func (h *Handler) controllerURL(r *http.Request) string {
	if h.backendBaseURLProvider != nil {
		if backendBaseURL, err := h.backendBaseURLProvider.BackendBaseURL(r.Context()); err == nil {
			backendBaseURL = strings.TrimRight(strings.TrimSpace(backendBaseURL), "/")
			if backendBaseURL != "" {
				return backendBaseURL
			}
		}
	}
	if h.backendBaseURL != "" {
		return h.backendBaseURL
	}

	scheme := forwardedHeaderValue(r.Header.Get("X-Forwarded-Proto"))
	if scheme != "http" && scheme != "https" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := forwardedHeaderValue(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = r.Host
	}
	host = sanitizeHost(host)
	if host == "" {
		host = "localhost"
	}

	return (&url.URL{Scheme: scheme, Host: host}).String()
}

func forwardedHeaderValue(value string) string {
	if value == "" {
		return ""
	}
	first, _, _ := strings.Cut(value, ",")
	return strings.TrimSpace(first)
}

func sanitizeHost(value string) string {
	host := strings.TrimSpace(value)
	if host == "" || strings.ContainsAny(host, "\r\n\t \"'`<>\\") {
		return ""
	}

	if hostname, port, err := net.SplitHostPort(host); err == nil {
		if hostname == "" || port == "" {
			return ""
		}
		return net.JoinHostPort(hostname, port)
	}

	return host
}

func (h *Handler) handleAgentBinary(asset agentBinaryAsset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		binaryPath := filepath.Join(h.agentBinaryDir, asset.filename)

		info, err := os.Stat(binaryPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				httpx.WriteProblem(w, r, httpx.NotFound("agent binary not found"))
				return
			}
			httpx.WriteProblem(w, r, httpx.ServiceUnavailable("agent binary unavailable"))
			return
		}
		if info.IsDir() {
			httpx.WriteProblem(w, r, httpx.NotFound("agent binary not found"))
			return
		}

		file, err := os.Open(binaryPath)
		if err != nil {
			httpx.WriteProblem(w, r, httpx.ServiceUnavailable("agent binary unavailable"))
			return
		}
		defer file.Close()

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="`+asset.filename+`"`)
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeContent(w, r, asset.filename, info.ModTime(), file)
	}
}
