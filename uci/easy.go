package uci

import (
	"context"
	"fmt"
	"os/exec"
	"slices"
)

type EasyEngineOptions struct {
	Name            string
	Args            []string
	Env             []string
	Dir             string
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
			return nil, fmt.Errorf("wait for initialization: %w", err)
		}
	}
	return e, nil
}
