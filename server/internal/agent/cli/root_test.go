package cli

import (
	"bytes"
	"context"
	"testing"

	agentservice "github.com/yorukot/netstamp/internal/agent/service"
)

func TestServiceInstallCommandPassesFlagsToManager(t *testing.T) {
	manager := &fakeServiceManager{}
	var stdout bytes.Buffer

	code := Execute(context.Background(), Options{
		Args:           []string{"service", "install", "--url", "https://netstamp.example.com", "--probe-id", "33333333-3333-3333-3333-333333333333", "--probe-secret", "secret"},
		Stdout:         &stdout,
		Stderr:         &bytes.Buffer{},
		ServiceManager: manager,
	})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !manager.installCalled {
		t.Fatal("expected install to be called")
	}
	if manager.installConfig.ControllerURL != "https://netstamp.example.com" {
		t.Fatalf("unexpected controller URL: %#v", manager.installConfig)
	}
	if manager.installConfig.ProbeID != "33333333-3333-3333-3333-333333333333" {
		t.Fatalf("unexpected probe ID: %#v", manager.installConfig)
	}
	if manager.installConfig.ProbeSecret != "secret" {
		t.Fatalf("unexpected probe secret: %#v", manager.installConfig)
	}
	if stdout.String() != "netstamp-agent installed and started\n" {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
}

func TestServiceInstallCommandRequiresFlags(t *testing.T) {
	var stderr bytes.Buffer

	code := Execute(context.Background(), Options{
		Args:           []string{"service", "install", "--url", "https://netstamp.example.com"},
		Stdout:         &bytes.Buffer{},
		Stderr:         &stderr,
		ServiceManager: &fakeServiceManager{},
	})
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	if stderr.String() == "" {
		t.Fatal("expected missing flag error on stderr")
	}
}

func TestServiceUninstallCommandPassesPurge(t *testing.T) {
	manager := &fakeServiceManager{}

	code := Execute(context.Background(), Options{
		Args:           []string{"service", "uninstall", "--purge"},
		Stdout:         &bytes.Buffer{},
		Stderr:         &bytes.Buffer{},
		ServiceManager: manager,
	})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !manager.uninstallCalled {
		t.Fatal("expected uninstall to be called")
	}
	if !manager.purge {
		t.Fatal("expected purge to be true")
	}
}

func TestServiceStatusCommandCallsManager(t *testing.T) {
	manager := &fakeServiceManager{}

	code := Execute(context.Background(), Options{
		Args:           []string{"service", "status"},
		Stdout:         &bytes.Buffer{},
		Stderr:         &bytes.Buffer{},
		ServiceManager: manager,
	})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !manager.statusCalled {
		t.Fatal("expected status to be called")
	}
}

type fakeServiceManager struct {
	installCalled   bool
	installConfig   agentservice.InstallConfig
	uninstallCalled bool
	purge           bool
	statusCalled    bool
}

func (m *fakeServiceManager) Install(_ context.Context, config agentservice.InstallConfig) error {
	m.installCalled = true
	m.installConfig = config
	return nil
}

func (m *fakeServiceManager) Uninstall(_ context.Context, purge bool) error {
	m.uninstallCalled = true
	m.purge = purge
	return nil
}

func (m *fakeServiceManager) Status(context.Context) error {
	m.statusCalled = true
	return nil
}
