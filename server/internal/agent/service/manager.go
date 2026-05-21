package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

const (
	ServiceName = "netstamp-agent"
	SystemUser  = "netstamp"
	SystemGroup = "netstamp"

	configDirMode   os.FileMode = 0o750
	envFileMode     os.FileMode = 0o600
	serviceDirMode  os.FileMode = 0o750
	serviceFileMode os.FileMode = 0o600
)

type InstallConfig struct {
	ControllerURL string
	ProbeID       string
	ProbeSecret   string
}

type Paths struct {
	InstallPath       string
	ConfigDir         string
	EnvFile           string
	ServiceFile       string
	HomeDir           string
	SystemdRuntimeDir string
}

type Options struct {
	Paths  Paths
	Runner CommandRunner
	OSName string
	EUID   func() int
}

type CommandRunner interface {
	Check(ctx context.Context, name string, args ...string) error
	Run(ctx context.Context, name string, args ...string) error
}

type Manager struct {
	paths  Paths
	runner CommandRunner
	osName string
	euid   func() int
}

func NewManager(options Options) *Manager {
	paths := options.Paths.withDefaults()
	runner := options.Runner
	if runner == nil {
		runner = ExecRunner{}
	}
	osName := options.OSName
	if osName == "" {
		osName = runtime.GOOS
	}
	euid := options.EUID
	if euid == nil {
		euid = os.Geteuid
	}

	return &Manager{
		paths:  paths,
		runner: runner,
		osName: osName,
		euid:   euid,
	}
}

func (p Paths) withDefaults() Paths {
	if p.InstallPath == "" {
		p.InstallPath = "/usr/local/bin/netstamp-agent"
	}
	if p.ConfigDir == "" {
		p.ConfigDir = "/etc/netstamp"
	}
	if p.EnvFile == "" {
		p.EnvFile = p.ConfigDir + "/probe.env"
	}
	if p.ServiceFile == "" {
		p.ServiceFile = "/etc/systemd/system/" + ServiceName + ".service"
	}
	if p.HomeDir == "" {
		p.HomeDir = "/var/lib/netstamp"
	}
	if p.SystemdRuntimeDir == "" {
		p.SystemdRuntimeDir = "/run/systemd/system"
	}
	return p
}

func (m *Manager) Install(ctx context.Context, config InstallConfig) error {
	config, err := normalizeInstallConfig(config)
	if err != nil {
		return err
	}
	if err := m.requireRootSystemdHost(); err != nil {
		return err
	}
	if err := m.requireInstalledBinary(); err != nil {
		return err
	}
	if err := m.ensureGroup(ctx); err != nil {
		return err
	}
	if err := m.ensureUser(ctx); err != nil {
		return err
	}
	if err := os.MkdirAll(m.paths.ConfigDir, configDirMode); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	if err := os.WriteFile(m.paths.EnvFile, []byte(renderEnvFile(config)), envFileMode); err != nil {
		return fmt.Errorf("write env file: %w", err)
	}
	if err := os.Chmod(m.paths.EnvFile, envFileMode); err != nil {
		return fmt.Errorf("chmod env file: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(m.paths.ServiceFile), serviceDirMode); err != nil {
		return fmt.Errorf("create systemd directory: %w", err)
	}
	if err := os.WriteFile(m.paths.ServiceFile, []byte(m.renderServiceFile()), serviceFileMode); err != nil {
		return fmt.Errorf("write systemd service: %w", err)
	}
	if err := os.Chmod(m.paths.ServiceFile, serviceFileMode); err != nil {
		return fmt.Errorf("chmod systemd service: %w", err)
	}
	if err := m.runner.Run(ctx, "systemctl", "daemon-reload"); err != nil {
		return fmt.Errorf("reload systemd: %w", err)
	}
	if err := m.runner.Run(ctx, "systemctl", "enable", "--now", ServiceName); err != nil {
		return fmt.Errorf("enable service: %w", err)
	}
	return nil
}

func (m *Manager) Uninstall(ctx context.Context, purge bool) error {
	if err := m.requireRootLinuxHost(); err != nil {
		return err
	}

	m.runOptional(ctx, "systemctl", "disable", "--now", ServiceName)
	if err := removeIfExists(m.paths.ServiceFile); err != nil {
		return err
	}
	if err := removeIfExists(m.paths.InstallPath); err != nil {
		return err
	}
	if err := m.runner.Run(ctx, "systemctl", "daemon-reload"); err != nil {
		return fmt.Errorf("reload systemd: %w", err)
	}
	m.runOptional(ctx, "systemctl", "reset-failed", ServiceName)

	if purge {
		if err := removeIfExists(m.paths.EnvFile); err != nil {
			return err
		}
		_ = os.Remove(m.paths.ConfigDir)
		if err := os.RemoveAll(m.paths.HomeDir); err != nil {
			return fmt.Errorf("remove home directory: %w", err)
		}
		m.removeUser(ctx)
		m.removeGroup(ctx)
	}

	return nil
}

func (m *Manager) Status(ctx context.Context) error {
	if m.osName != "linux" {
		return errors.New("service status is supported on Linux only")
	}
	return m.runner.Run(ctx, "systemctl", "status", ServiceName)
}

func (m *Manager) requireRootSystemdHost() error {
	if err := m.requireRootLinuxHost(); err != nil {
		return err
	}
	if _, err := os.Stat(m.paths.SystemdRuntimeDir); err != nil {
		return fmt.Errorf("systemd does not appear to be running: %w", err)
	}
	return nil
}

func (m *Manager) requireRootLinuxHost() error {
	if m.osName != "linux" {
		return errors.New("service management is supported on Linux only")
	}
	if m.euid() != 0 {
		return errors.New("service management must be run as root")
	}
	return nil
}

func (m *Manager) requireInstalledBinary() error {
	info, err := os.Stat(m.paths.InstallPath)
	if err != nil {
		return fmt.Errorf("agent binary is not installed at %s: %w", m.paths.InstallPath, err)
	}
	if info.IsDir() || info.Mode().Perm()&0o111 == 0 {
		return fmt.Errorf("agent binary is not executable at %s", m.paths.InstallPath)
	}
	return nil
}

func (m *Manager) ensureGroup(ctx context.Context) error {
	if m.runner.Check(ctx, "getent", "group", SystemGroup) == nil {
		return nil
	}
	if err := m.runner.Run(ctx, "groupadd", "--system", SystemGroup); err == nil {
		return nil
	}
	if err := m.runner.Run(ctx, "addgroup", "--system", SystemGroup); err == nil {
		return nil
	}
	if err := m.runner.Run(ctx, "addgroup", "-S", SystemGroup); err == nil {
		return nil
	}
	return errors.New("create netstamp group: groupadd or addgroup is required")
}

func (m *Manager) ensureUser(ctx context.Context) error {
	if m.runner.Check(ctx, "id", "-u", SystemUser) == nil {
		return nil
	}
	shell := nologinShell()
	if err := m.runner.Run(ctx,
		"useradd",
		"--system",
		"--gid", SystemGroup,
		"--home-dir", m.paths.HomeDir,
		"--create-home",
		"--shell", shell,
		SystemUser,
	); err == nil {
		return nil
	}
	if err := m.runner.Run(ctx,
		"adduser",
		"--system",
		"--ingroup", SystemGroup,
		"--home", m.paths.HomeDir,
		"--shell", shell,
		"--disabled-login",
		"--gecos", "",
		SystemUser,
	); err == nil {
		return nil
	}
	if err := m.runner.Run(ctx,
		"adduser",
		"-S",
		"-D",
		"-h", m.paths.HomeDir,
		"-s", shell,
		"-G", SystemGroup,
		SystemUser,
	); err == nil {
		return nil
	}
	return errors.New("create netstamp user: useradd or adduser is required")
}

func (m *Manager) removeUser(ctx context.Context) {
	if m.runner.Check(ctx, "id", "-u", SystemUser) != nil {
		return
	}
	if m.runner.Run(ctx, "userdel", SystemUser) == nil {
		return
	}
	m.runOptional(ctx, "deluser", SystemUser)
}

func (m *Manager) removeGroup(ctx context.Context) {
	if m.runner.Check(ctx, "getent", "group", SystemGroup) != nil {
		return
	}
	if m.runner.Run(ctx, "groupdel", SystemGroup) == nil {
		return
	}
	m.runOptional(ctx, "delgroup", SystemGroup)
}

func (m *Manager) runOptional(ctx context.Context, name string, args ...string) bool {
	return m.runner.Run(ctx, name, args...) == nil
}

func (m *Manager) renderServiceFile() string {
	return fmt.Sprintf(`[Unit]
Description=Netstamp Probe Agent
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
User=%s
Group=%s
EnvironmentFile=%s
ExecStart=%s run
Restart=always
RestartSec=5s
AmbientCapabilities=CAP_NET_RAW
CapabilityBoundingSet=CAP_NET_RAW
NoNewPrivileges=true
PrivateTmp=true
ProtectHome=true
ProtectSystem=full

[Install]
WantedBy=multi-user.target
`, SystemUser, SystemGroup, m.paths.EnvFile, m.paths.InstallPath)
}

func normalizeInstallConfig(config InstallConfig) (InstallConfig, error) {
	config.ControllerURL = strings.TrimSpace(config.ControllerURL)
	config.ProbeID = strings.TrimSpace(config.ProbeID)
	config.ProbeSecret = strings.TrimSpace(config.ProbeSecret)

	if err := validateControllerURL(config.ControllerURL); err != nil {
		return InstallConfig{}, err
	}
	if _, err := domainprobe.VNProbeID(config.ProbeID); err != nil {
		return InstallConfig{}, fmt.Errorf("probe ID is invalid: %w", err)
	}
	if config.ProbeSecret == "" {
		return InstallConfig{}, errors.New("probe secret is required")
	}
	if strings.Contains(config.ProbeSecret, "\n") {
		return InstallConfig{}, errors.New("probe secret must not contain a newline")
	}
	return config, nil
}

func validateControllerURL(raw string) error {
	if raw == "" {
		return errors.New("controller URL is required")
	}
	controllerURL, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("controller URL is invalid: %w", err)
	}
	if controllerURL.Scheme != "http" && controllerURL.Scheme != "https" {
		return errors.New("controller URL must use http or https")
	}
	if controllerURL.Host == "" {
		return errors.New("controller URL must include a host")
	}
	if controllerURL.RawQuery != "" || controllerURL.Fragment != "" || controllerURL.User != nil {
		return errors.New("controller URL must be an origin or base URL without query, fragment, or credentials")
	}
	return nil
}

func renderEnvFile(config InstallConfig) string {
	return fmt.Sprintf(`# Managed by netstamp-agent service install.
NETSTAMP_PROBE_CONTROLLER_URL="%s"
NETSTAMP_PROBE_ID="%s"
NETSTAMP_PROBE_SECRET="%s"
NETSTAMP_PROBE_LOG_LEVEL="info"
`,
		systemdEnvEscape(strings.TrimRight(config.ControllerURL, "/")),
		systemdEnvEscape(config.ProbeID),
		systemdEnvEscape(config.ProbeSecret),
	)
}

func systemdEnvEscape(value string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
		`$`, `\$`,
		"`", "\\`",
	)
	return replacer.Replace(value)
}

func nologinShell() string {
	for _, path := range []string{"/usr/sbin/nologin", "/sbin/nologin"} {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}
	return "/bin/false"
}

func removeIfExists(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove %s: %w", path, err)
	}
	return nil
}
