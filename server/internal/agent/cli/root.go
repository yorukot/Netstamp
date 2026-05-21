package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/spf13/cobra"

	agentapp "github.com/yorukot/netstamp/internal/agent/app"
	agentconfig "github.com/yorukot/netstamp/internal/agent/config"
	agentservice "github.com/yorukot/netstamp/internal/agent/service"
)

type Options struct {
	Args           []string
	Stdout         io.Writer
	Stderr         io.Writer
	ServiceManager serviceManager
}

type serviceManager interface {
	Install(ctx context.Context, config agentservice.InstallConfig) error
	Uninstall(ctx context.Context, purge bool) error
	Status(ctx context.Context) error
}

func Execute(ctx context.Context, options Options) int {
	cmd := NewRootCommand(options)
	if err := cmd.ExecuteContext(ctx); err != nil {
		_, _ = fmt.Fprintln(options.stderr(), err)
		return 1
	}
	return 0
}

func NewRootCommand(options Options) *cobra.Command {
	root := &cobra.Command{
		Use:           "netstamp-agent",
		Short:         "Run and manage the Netstamp probe agent",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAgent(cmd.Context(), options.stderr())
		},
	}
	root.SetOut(options.stdout())
	root.SetErr(options.stderr())
	root.SetArgs(options.Args)

	root.AddCommand(newRunCommand(options))
	root.AddCommand(newServiceCommand(options))

	return root
}

func newRunCommand(options Options) *cobra.Command {
	return &cobra.Command{
		Use:           "run",
		Short:         "Run the probe agent runtime",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAgent(cmd.Context(), options.stderr())
		},
	}
}

func newServiceCommand(options Options) *cobra.Command {
	serviceCmd := &cobra.Command{
		Use:           "service",
		Short:         "Manage the probe agent system service",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
	}

	serviceCmd.AddCommand(newServiceInstallCommand(options))
	serviceCmd.AddCommand(newServiceUninstallCommand(options))
	serviceCmd.AddCommand(newServiceStatusCommand(options))

	return serviceCmd
}

func newServiceInstallCommand(options Options) *cobra.Command {
	var config agentservice.InstallConfig
	cmd := &cobra.Command{
		Use:           "install",
		Short:         "Install and start the probe agent systemd service",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := options.manager().Install(cmd.Context(), config); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(options.stdout(), "%s installed and started\n", agentservice.ServiceName)
			return nil
		},
	}
	cmd.Flags().StringVar(&config.ControllerURL, "url", "", "Netstamp controller URL")
	cmd.Flags().StringVar(&config.ProbeID, "probe-id", "", "Probe ID")
	cmd.Flags().StringVar(&config.ProbeSecret, "probe-secret", "", "Probe secret")
	markRequired(cmd, "url")
	markRequired(cmd, "probe-id")
	markRequired(cmd, "probe-secret")
	return cmd
}

func newServiceUninstallCommand(options Options) *cobra.Command {
	var purge bool
	cmd := &cobra.Command{
		Use:           "uninstall",
		Short:         "Uninstall the probe agent systemd service",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := options.manager().Uninstall(cmd.Context(), purge); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(options.stdout(), "%s uninstalled\n", agentservice.ServiceName)
			return nil
		},
	}
	cmd.Flags().BoolVar(&purge, "purge", false, "Remove probe config, data directory, and system user/group")
	return cmd
}

func newServiceStatusCommand(options Options) *cobra.Command {
	return &cobra.Command{
		Use:           "status",
		Short:         "Show the probe agent systemd service status",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return options.manager().Status(cmd.Context())
		},
	}
}

func runAgent(ctx context.Context, errw io.Writer) error {
	config, err := agentconfig.LoadConfig()
	if err != nil {
		return fmt.Errorf("load probe config: %w", err)
	}

	log := agentapp.NewLogger(config.LogLevel)
	log.LogAttrs(ctx, slog.LevelInfo, "probe agent starting", config.SafeLogAttrs()...)

	app := agentapp.New(config, log)
	if err := app.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Error("probe agent stopped with error", "error", err)
		return err
	}

	log.Info("probe agent stopped")
	return nil
}

func (o Options) manager() serviceManager {
	if o.ServiceManager != nil {
		return o.ServiceManager
	}
	return agentservice.NewManager(agentservice.Options{
		Runner: agentservice.ExecRunner{
			Stdout: o.stdout(),
			Stderr: o.stderr(),
		},
	})
}

func (o Options) stdout() io.Writer {
	if o.Stdout != nil {
		return o.Stdout
	}
	return io.Discard
}

func (o Options) stderr() io.Writer {
	if o.Stderr != nil {
		return o.Stderr
	}
	return io.Discard
}

func markRequired(cmd *cobra.Command, name string) {
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(err)
	}
}
