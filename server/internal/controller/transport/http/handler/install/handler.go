package install

import (
	"embed"
	"errors"
	"net/http"
	"os"
	"path/filepath"

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
	agentBinaryDir string
}

func NewHandler(agentBinaryDir string) *Handler {
	if agentBinaryDir == "" {
		agentBinaryDir = DefaultAgentBinaryDir
	}

	return &Handler{agentBinaryDir: agentBinaryDir}
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

	w.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		return
	}
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
