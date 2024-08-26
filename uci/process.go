package uci

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
)

type Process interface {
	Send(s string) error
	Recv() (string, error)
	Done() <-chan struct{}
	Err() error
	Kill()
}

type TracingProcessOptions struct {
	MyName      string
	ProcessName string
}

func (o *TracingProcessOptions) FillDefaults() {
	if o.MyName == "" {
		o.MyName = "me"
	}
	if o.ProcessName == "" {
		o.ProcessName = "engine"
	}
}

func (o TracingProcessOptions) Clone() TracingProcessOptions {
	return o
}

func NewTracingProcess(p Process, l Logger, o TracingProcessOptions) Process {
	o = o.Clone()
	o.FillDefaults()
	return &tracingProcess{p: p, l: l, o: o}
}

func NewCmdProcess(cmd *exec.Cmd) (Process, error) {
	if cmd.Process != nil {
		return nil, fmt.Errorf("process already started")
	}
	if cmd.Stdin != nil {
		return nil, fmt.Errorf("stdin already set")
	}
	if cmd.Stdout != nil {
		return nil, fmt.Errorf("stdout already set")
	}

	outPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("redirect stdin: %w", err)
	}
	inPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("redirect stdout: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start process: %w", err)
	}

	res := &cmdProcess{
		c:    cmd,
		done: make(chan struct{}),
		err:  nil,

		inPipe:  inPipe,
		bufIn:   bufio.NewReader(inPipe),
		outPipe: outPipe,
	}
	res.closed.Store(false)
	go res.waitLoop()
	return res, nil
}

func NewCancellableProcess(ctx context.Context, p Process) Process {
	return &cancellableProcess{ctx: ctx, p: p}
}

type tracingProcess struct {
	l Logger
	p Process
	o TracingProcessOptions
}

func (p *tracingProcess) Send(s string) error {
	p.l.Printf("%v -> %v: %v", p.o.MyName, p.o.ProcessName, s)
	if err := p.p.Send(s); err != nil {
		p.l.Printf("%v: send failed: %v", p.o.ProcessName, err)
		return err
	}
	return nil
}

func (p *tracingProcess) Recv() (string, error) {
	s, err := p.p.Recv()
	if err != nil {
		if !errors.Is(err, io.EOF) {
			p.l.Printf("%v: recv failed: %v", p.o.ProcessName, err)
		}
		return "", err
	}
	p.l.Printf("%v -> %v: %v", p.o.ProcessName, p.o.MyName, s)
	return s, nil
}

func (p *tracingProcess) Done() <-chan struct{} { return p.p.Done() }
func (p *tracingProcess) Err() error            { return p.p.Err() }

func (p *tracingProcess) Kill() {
	p.l.Printf("%v: killing", p.o.ProcessName)
	p.p.Kill()
}

type cmdProcess struct {
	c *exec.Cmd

	done chan struct{}
	err  error

	inMu  sync.Mutex
	outMu sync.Mutex

	inPipe  io.ReadCloser
	bufIn   *bufio.Reader
	outPipe io.WriteCloser
	closed  atomic.Bool
}

func (p *cmdProcess) waitLoop() {
	err := p.c.Wait()
	if err != nil {
		err = fmt.Errorf("wait: %w", err)
	}
	p.err = err
	close(p.done)
	p.closed.Store(true)
}

func (p *cmdProcess) Send(s string) error {
	if p.closed.Load() {
		<-p.done
		return fmt.Errorf("i/o pipes closed")
	}

	p.outMu.Lock()
	defer p.outMu.Unlock()
	_, err := io.WriteString(p.outPipe, s+"\n")
	if err != nil {
		p.Kill()
		<-p.done
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

func (p *cmdProcess) Recv() (string, error) {
	if p.closed.Load() {
		<-p.done
		return "", fmt.Errorf("i/o pipes closed")
	}

	p.inMu.Lock()
	defer p.inMu.Unlock()
	s, err := p.bufIn.ReadString('\n')
	if err != nil {
		p.Kill()
		<-p.done
		return "", fmt.Errorf("read: %w", err)
	}
	s = strings.TrimRight(s, "\n\r")
	return s, nil
}

func (p *cmdProcess) Done() <-chan struct{} {
	return p.done
}

func (p *cmdProcess) Err() error {
	select {
	case <-p.done:
		return p.err
	default:
		return nil
	}
}

func (p *cmdProcess) Kill() {
	if !p.closed.Swap(true) {
		_ = p.c.Process.Kill()
		_ = p.inPipe.Close()
		_ = p.outPipe.Close()
	}
}

type cancellableProcess struct {
	ctx context.Context
	p   Process
}

func (p *cancellableProcess) Send(s string) error {
	res := make(chan error, 1)
	go func() {
		res <- p.p.Send(s)
	}()
	select {
	case err := <-res:
		return err
	case <-p.ctx.Done():
		return p.ctx.Err()
	}
}

func (p *cancellableProcess) Recv() (string, error) {
	type result struct {
		s   string
		err error
	}
	res := make(chan result, 1)
	go func() {
		s, err := p.p.Recv()
		res <- result{s: s, err: err}
	}()
	select {
	case r := <-res:
		return r.s, r.err
	case <-p.ctx.Done():
		return "", p.ctx.Err()
	}
}

func (p *cancellableProcess) Done() <-chan struct{} { return p.p.Done() }
func (p *cancellableProcess) Err() error            { return p.p.Err() }
func (p *cancellableProcess) Kill()                 { p.p.Kill() }
