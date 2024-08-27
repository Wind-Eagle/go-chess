package uci

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
)

var errTerminated = errors.New("process terminated")

type command interface {
	Serialize() string
	uciCommandMarker()
}

type state interface {
	Start(ctx context.Context, p Process) error
	ProcessCommand(cmd command) (command, any, error)
	ProcessMessage(msg string) error
	Initialized() bool
	Finish()
}

type commandReply struct {
	res any
	err error
}

type commandExt struct {
	cmd   command
	reply chan<- commandReply
}

func newCommandExt(cmd command) (commandExt, <-chan commandReply) {
	reply := make(chan commandReply, 1)
	return commandExt{cmd: cmd, reply: reply}, reply
}

type core struct {
	ctx      context.Context
	p        Process
	l        Logger
	s        state
	cmdCh    chan commandExt
	msgCh    chan string
	done     atomic.Bool
	cancel   func()
	procDone chan struct{}
	commDone chan struct{}
}

func newCore(p Process, l Logger, s state) *core {
	ctx, cancel := context.WithCancel(context.Background())
	c := &core{
		ctx:      ctx,
		p:        NewCancellableProcess(ctx, p),
		l:        l,
		s:        s,
		cmdCh:    make(chan commandExt),
		msgCh:    make(chan string),
		cancel:   cancel,
		procDone: make(chan struct{}),
		commDone: make(chan struct{}),
	}
	c.done.Store(false)
	go c.commLoop()
	go c.processLoop()
	return c
}

func (c *core) Done() <-chan struct{} {
	return c.commDone
}

func (c *core) doCancel() {
	if !c.done.Swap(true) {
		c.cancel()
	}
}

func (c *core) Cancel(wait bool) {
	c.doCancel()
	if wait {
		<-c.commDone
	}
}

func (c *core) Send(ctx context.Context, cmd command) (any, error) {
	cmdEx, reply := newCommandExt(cmd)
	select {
	case c.cmdCh <- cmdEx:
		r := <-reply
		return r.res, r.err
	case <-ctx.Done():
		return nil, fmt.Errorf("wait: %w", ctx.Err())
	case <-c.commDone:
		return nil, errTerminated
	}
}

func (c *core) commLoop() {
	defer close(c.commDone)
loop:
	for {
		ln, err := c.p.Recv()
		if err != nil {
			select {
			case <-c.ctx.Done():
			default:
				if !errors.Is(err, io.EOF) {
					c.l.Printf("cannot receive line from engine: %v", err)
				}
			}
			break
		}
		select {
		case c.msgCh <- ln:
		case <-c.ctx.Done():
			break loop
		}
	}
	c.doCancel()
	select {
	case <-c.p.Done():
		if err := c.p.Err(); err != nil {
			c.l.Printf("engine terminated badly: %v", err)
		}
	default:
		c.l.Printf("killing engine")
		c.p.Kill()
	}
	<-c.procDone
}

func (c *core) processLoop() {
	defer close(c.procDone)
	defer c.s.Finish()

	if err := c.s.Start(c.ctx, c.p); err != nil {
		c.l.Printf("cannot start: %v", err)
		c.doCancel()
		return
	}

	cmdCh := (chan commandExt)(nil)
	for {
		if cmdCh == nil && c.s.Initialized() {
			cmdCh = c.cmdCh
		}
		select {
		case msg := <-c.msgCh:
			if err := c.s.ProcessMessage(msg); err != nil {
				c.l.Printf("bad line: %v", err)
			}
		case cmd := <-cmdCh:
			realCmd, res, err := c.s.ProcessCommand(cmd.cmd)
			cmd.reply <- commandReply{res: res, err: err}
			canSend := false
			if err == nil {
				canSend = true
				if _, ok := realCmd.(cmdEmpty); ok {
					canSend = false
				}
			}
			if canSend {
				if err := c.p.Send(realCmd.Serialize()); err != nil {
					c.l.Printf("cannot send command: %v", err)
					c.doCancel()
					return
				}
			}
		case <-c.ctx.Done():
			return
		}
	}
}
