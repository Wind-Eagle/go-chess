package uci

import (
	"context"
	"fmt"
	"os/exec"
	"slices"
	"syscall"
)

type EasyEngineOptions struct {
	Name            string
	Args            []string
	Env             []string
	Dir             string
	SysProcAttr     *syscall.SysProcAttr
	Logger          Logger
	EnableTracing   bool
	TracingOptions  TracingProcessOptions
	Options         EngineOptions
	WaitInitialized bool
}

func NewEasyEngine(ctx context.Context, o EasyEngineOptions) (*Engine, error) {
	cmd := exec.Command(o.Name, o.Args...)
	cmd.Env = slices.Clone(o.Env)
	cmd.Dir = o.Dir
	cmd.SysProcAttr = o.SysProcAttr
	p, err := NewCmdProcess(cmd)
	if err != nil {
		return nil, fmt.Errorf("create process: %w", err)
	}
	if o.EnableTracing && o.Logger != nil {
		p = NewTracingProcess(p, o.Logger, o.TracingOptions)
	}
	e := NewEngine(ctx, p, o.Logger, o.Options)
	if o.WaitInitialized {
		if err := e.WaitInitialized(ctx); err != nil {
			e.Close()
			return nil, fmt.Errorf("wait for initialization: %w", err)
		}
	}
	return e, nil
}
