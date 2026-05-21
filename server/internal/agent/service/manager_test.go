package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManagerInstallWritesEnvAndSystemdUnit(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	paths := testPaths(dir)
	if err := os.MkdirAll(paths.SystemdRuntimeDir, 0o755); err != nil {
		t.Fatalf("create systemd runtime dir: %v", err)
	}
	writeTestFile(t, paths.InstallPath, 0o755)

	runner := &fakeRunner{
		checks: map[string]error{
			commandKey("getent", "group", SystemGroup): errors.New("missing group"),
			commandKey("id", "-u", SystemUser):         errors.New("missing user"),
		},
	}
	manager := NewManager(Options{
		Paths:  paths,
		Runner: runner,
		OSName: "linux",
		EUID:   func() int { return 0 },
	})

	err := manager.Install(ctx, InstallConfig{
		ControllerURL: "https://netstamp.example.com/",
		ProbeID:       "33333333-3333-3333-3333-333333333333",
		ProbeSecret:   `secret"$value`,
	})
	if err != nil {
		t.Fatalf("install service: %v", err)
	}

	envData, err := os.ReadFile(paths.EnvFile)
	if err != nil {
		t.Fatalf("read env file: %v", err)
	}
	env := string(envData)
	for _, want := range []string{
		`NETSTAMP_PROBE_CONTROLLER_URL="https://netstamp.example.com"`,
		`NETSTAMP_PROBE_ID="33333333-3333-3333-3333-333333333333"`,
		`NETSTAMP_PROBE_SECRET="secret\"\$value"`,
	} {
		if !strings.Contains(env, want) {
			t.Fatalf("expected env file to contain %q, got %q", want, env)
		}
	}
	info, statErr := os.Stat(paths.EnvFile)
	if statErr != nil {
		t.Fatalf("stat env file: %v", statErr)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected env file mode 0600, got %o", info.Mode().Perm())
	}

	unitData, err := os.ReadFile(paths.ServiceFile)
	if err != nil {
		t.Fatalf("read service file: %v", err)
	}
	info, statErr = os.Stat(paths.ServiceFile)
	if statErr != nil {
		t.Fatalf("stat service file: %v", statErr)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected service file mode 0600, got %o", info.Mode().Perm())
	}
	unit := string(unitData)
	for _, want := range []string{
		"User=netstamp",
		"Group=netstamp",
		"ExecStart=" + paths.InstallPath + " run",
		"AmbientCapabilities=CAP_NET_RAW",
		"CapabilityBoundingSet=CAP_NET_RAW",
	} {
		if !strings.Contains(unit, want) {
			t.Fatalf("expected service file to contain %q, got %q", want, unit)
		}
	}

	for _, want := range []string{
		commandKey("groupadd", "--system", SystemGroup),
		commandKey("useradd", "--system", "--gid", SystemGroup, "--home-dir", paths.HomeDir, "--create-home", "--shell", nologinShell(), SystemUser),
		commandKey("systemctl", "daemon-reload"),
		commandKey("systemctl", "enable", "--now", ServiceName),
	} {
		if !runner.ran(want) {
			t.Fatalf("expected command %q to run, got %#v", want, runner.runs)
		}
	}
}

func TestManagerInstallRejectsMissingBinary(t *testing.T) {
	dir := t.TempDir()
	paths := testPaths(dir)
	if err := os.MkdirAll(paths.SystemdRuntimeDir, 0o755); err != nil {
		t.Fatalf("create systemd runtime dir: %v", err)
	}
	manager := NewManager(Options{
		Paths:  paths,
		Runner: &fakeRunner{},
		OSName: "linux",
		EUID:   func() int { return 0 },
	})

	err := manager.Install(context.Background(), InstallConfig{
		ControllerURL: "https://netstamp.example.com",
		ProbeID:       "33333333-3333-3333-3333-333333333333",
		ProbeSecret:   "secret",
	})
	if err == nil || !strings.Contains(err.Error(), "agent binary is not installed") {
		t.Fatalf("expected missing binary error, got %v", err)
	}
}

func TestManagerUninstallPreservesConfigWithoutPurge(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	paths := testPaths(dir)
	writeTestFile(t, paths.InstallPath, 0o755)
	writeTestFile(t, paths.ServiceFile, 0o644)
	writeTestFile(t, paths.EnvFile, 0o600)

	runner := &fakeRunner{}
	manager := NewManager(Options{
		Paths:  paths,
		Runner: runner,
		OSName: "linux",
		EUID:   func() int { return 0 },
	})

	if err := manager.Uninstall(ctx, false); err != nil {
		t.Fatalf("uninstall service: %v", err)
	}
	assertMissing(t, paths.InstallPath)
	assertMissing(t, paths.ServiceFile)
	if _, err := os.Stat(paths.EnvFile); err != nil {
		t.Fatalf("expected env file to be preserved: %v", err)
	}
	if !runner.ran(commandKey("systemctl", "disable", "--now", ServiceName)) {
		t.Fatalf("expected service disable command, got %#v", runner.runs)
	}
}

func TestManagerUninstallPurgesConfigAndHome(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	paths := testPaths(dir)
	writeTestFile(t, paths.InstallPath, 0o755)
	writeTestFile(t, paths.ServiceFile, 0o644)
	writeTestFile(t, paths.EnvFile, 0o600)
	if err := os.MkdirAll(paths.HomeDir, 0o755); err != nil {
		t.Fatalf("create home dir: %v", err)
	}
	runner := &fakeRunner{}
	manager := NewManager(Options{
		Paths:  paths,
		Runner: runner,
		OSName: "linux",
		EUID:   func() int { return 0 },
	})

	if err := manager.Uninstall(ctx, true); err != nil {
		t.Fatalf("uninstall service: %v", err)
	}
	assertMissing(t, paths.EnvFile)
	assertMissing(t, paths.HomeDir)
	if !runner.ran(commandKey("userdel", SystemUser)) {
		t.Fatalf("expected userdel command, got %#v", runner.runs)
	}
	if !runner.ran(commandKey("groupdel", SystemGroup)) {
		t.Fatalf("expected groupdel command, got %#v", runner.runs)
	}
}

func testPaths(dir string) Paths {
	return Paths{
		InstallPath:       filepath.Join(dir, "usr", "local", "bin", "netstamp-agent"),
		ConfigDir:         filepath.Join(dir, "etc", "netstamp"),
		EnvFile:           filepath.Join(dir, "etc", "netstamp", "probe.env"),
		ServiceFile:       filepath.Join(dir, "etc", "systemd", "system", "netstamp-agent.service"),
		HomeDir:           filepath.Join(dir, "var", "lib", "netstamp"),
		SystemdRuntimeDir: filepath.Join(dir, "run", "systemd", "system"),
	}
}

func writeTestFile(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("test"), mode); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected %s to be missing, got %v", path, err)
	}
}

type fakeRunner struct {
	checks map[string]error
	runs   []string
}

func (r *fakeRunner) Check(_ context.Context, name string, args ...string) error {
	if r.checks == nil {
		return nil
	}
	if err, ok := r.checks[commandKey(name, args...)]; ok {
		return err
	}
	return nil
}

func (r *fakeRunner) Run(_ context.Context, name string, args ...string) error {
	r.runs = append(r.runs, commandKey(name, args...))
	return nil
}

func (r *fakeRunner) ran(key string) bool {
	for _, run := range r.runs {
		if run == key {
			return true
		}
	}
	return false
}

func commandKey(name string, args ...string) string {
	return name + " " + strings.Join(args, " ")
}
