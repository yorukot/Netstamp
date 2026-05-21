package install

import (
	"embed"
	"errors"
	"net/http"
	"os"

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

	w.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		return
	}
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
