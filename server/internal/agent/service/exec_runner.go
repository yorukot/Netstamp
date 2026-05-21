package service

import (
	"context"
	"io"
	"os"
	"os/exec"
)

type ExecRunner struct {
	Stdout io.Writer
	Stderr io.Writer
}

func (r ExecRunner) Check(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}

func (r ExecRunner) Run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = r.stdout()
	cmd.Stderr = r.stderr()
	return cmd.Run()
}

func (r ExecRunner) stdout() io.Writer {
	if r.Stdout != nil {
		return r.Stdout
	}
	return os.Stdout
}

func (r ExecRunner) stderr() io.Writer {
	if r.Stderr != nil {
		return r.Stderr
	}
	return os.Stderr
}
