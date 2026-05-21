package install

import (
	"embed"
	"errors"
	"net/http"
	"net/url"
	"os"
	pathpkg "path"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

const (
	DefaultAgentBinaryPath = "/app/agent"
	agentBinaryFilename    = "netstamp-agent-linux-amd64"
)

//go:embed agent.sh uninstall-agent.sh
var installerFiles embed.FS

type Handler struct {
	agentBinaryPath string
}

func NewHandler(agentBinaryPath string) *Handler {
	if agentBinaryPath == "" {
		agentBinaryPath = DefaultAgentBinaryPath
	}

	return &Handler{agentBinaryPath: agentBinaryPath}
}

func (h *Handler) RegisterRoutes(api chi.Router) {
	api.Get("/install/agent.sh", h.handleAgentScript)
	api.Get("/install/uninstall-agent.sh", h.handleAgentUninstallScript)
	api.Get("/install/"+agentBinaryFilename, h.handleAgentBinary)
}

func (h *Handler) handleAgentScript(w http.ResponseWriter, r *http.Request) {
	h.writeScript(w, r, "agent.sh", "agent installer unavailable", h.renderAgentScript)
}

func (h *Handler) handleAgentUninstallScript(w http.ResponseWriter, r *http.Request) {
	h.writeScript(w, r, "uninstall-agent.sh", "agent uninstaller unavailable", nil)
}

func (h *Handler) writeScript(
	w http.ResponseWriter,
	r *http.Request,
	name string,
	unavailableDetail string,
	render func(*http.Request, []byte) []byte,
) {
	data, err := installerFiles.ReadFile(name)
	if err != nil {
		httpx.WriteProblem(w, r, httpx.InternalServerError(unavailableDetail))
		return
	}
	if render != nil {
		data = render(r, data)
	}

	w.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		return
	}
}

func (h *Handler) renderAgentScript(r *http.Request, data []byte) []byte {
	script := string(data)
	script = strings.ReplaceAll(script, "__NETSTAMP_AGENT_BINARY_URL__", installAssetURL(r, agentBinaryFilename))
	script = strings.ReplaceAll(script, "__NETSTAMP_CONTROLLER_URL__", requestOrigin(r))
	return []byte(script)
}

func installAssetURL(r *http.Request, filename string) string {
	assetURL, err := url.Parse(requestOrigin(r))
	if err != nil {
		return pathpkg.Dir(r.URL.Path) + "/" + filename
	}
	assetURL.Path = pathpkg.Dir(r.URL.Path) + "/" + filename
	return assetURL.String()
}

func requestOrigin(r *http.Request) string {
	scheme := firstHeaderValue(r.Header.Get("X-Forwarded-Proto"))
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := firstHeaderValue(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = r.Host
	}

	return scheme + "://" + host
}

func firstHeaderValue(value string) string {
	value, _, _ = strings.Cut(value, ",")
	return strings.TrimSpace(value)
}

func (h *Handler) handleAgentBinary(w http.ResponseWriter, r *http.Request) {
	info, err := os.Stat(h.agentBinaryPath)
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

	file, err := os.Open(h.agentBinaryPath)
	if err != nil {
		httpx.WriteProblem(w, r, httpx.ServiceUnavailable("agent binary unavailable"))
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="`+agentBinaryFilename+`"`)
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, agentBinaryFilename, info.ModTime(), file)
}
