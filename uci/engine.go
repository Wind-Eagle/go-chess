package uci

import (
	"context"
	"fmt"
	"time"

	"github.com/alex65536/go-chess/chess"
)

type EngineOptions struct {
	// Reject all the lines containing non-ASCII characters, both from and to engine.
	SanitizeUTF8 bool

	// Put "info string" output into the logger.
	LogEngineString bool

	// Allow "name" and "value" substings in "setoption".
	AllowBadSubstringsInOptions bool

	// If true, then the engine is killed immediately when the context is cancelled. No mercy.
	NoWaitOnCancel bool

	// Maximum time to wait until the engine is initialized.
	//
	// Zero means default.
	InitTimeout time.Duration

	// Maximum time to wait until the engine terminates gracefully when the context is cancelled.
	//
	// Zero means default.
	WaitOnCancelTimeout time.Duration
}

func (o EngineOptions) Clone() EngineOptions {
	return o
}

func (o *EngineOptions) FillDefaults() {
	if o.InitTimeout == 0 {
		o.InitTimeout = 5 * time.Second
	}
	if o.WaitOnCancelTimeout == 0 {
		o.WaitOnCancelTimeout = 500 * time.Millisecond
	}
}

func (o *EngineOptions) coderOptions() coderOptions {
	return coderOptions{
		SanitizeUTF8:                o.SanitizeUTF8,
		AllowBadSubstringsInOptions: o.AllowBadSubstringsInOptions,
	}
}

type Engine struct {
	o EngineOptions
	l Logger
	s *engineState
	c *core
}

type Search struct {
	s *searchState
	e *Engine
}

func NewEngine(ctx context.Context, p Process, l Logger, o EngineOptions) *Engine {
	if l == nil {
		l = NewNullLogger()
	}
	o = o.Clone()
	o.FillDefaults()

	s := newEngineState(o, l)

	e := &Engine{
		o: o,
		l: l,
		s: s,
		c: newCore(p, l, s),
	}

	go e.waitInitializedThread()
	go e.watchCtxThread(ctx)

	return e
}

func (e *Engine) waitInitializedThread() {
	ctx, cancel := context.WithTimeout(context.Background(), e.o.InitTimeout)
	defer cancel()
	if err := e.WaitInitialized(ctx); err != nil {
		e.l.Printf("wait initialized failed: %v", err)
		e.Cancel()
		return
	}
}

func (e *Engine) watchCtxThread(watchedCtx context.Context) {
	select {
	case <-watchedCtx.Done():
	case <-e.Done():
		return
	}
	if e.o.NoWaitOnCancel {
		e.Cancel()
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), e.o.WaitOnCancelTimeout)
	defer cancel()
	if err := e.Quit(ctx, true); err != nil {
		e.l.Printf("engine has not terminated gracefully: %v", err)
		return
	}
}

func (e *Engine) Cancel() {
	e.c.Cancel()
}

func (e *Engine) Close() {
	e.c.Close()
}

func (e *Engine) Done() <-chan struct{} {
	return e.c.Done()
}

func (e *Engine) WaitInitialized(ctx context.Context) error {
	select {
	case <-e.s.InitializedChan():
		return nil
	case <-ctx.Done():
		return fmt.Errorf("wait: %w", ctx.Err())
	case <-e.Done():
		return errTerminated
	}
}

func (s *Search) Wait(ctx context.Context) error {
	select {
	case <-s.s.Done():
		return s.s.Err()
	case <-ctx.Done():
		return fmt.Errorf("wait: %w", ctx.Err())
	case <-s.e.Done():
		return errTerminated
	}
}

func (e *Engine) SetDebug(ctx context.Context, val bool) error {
	if _, err := e.c.Send(ctx, cmdDebug{val: val}); err != nil {
		return fmt.Errorf("send \"debug\": %w", err)
	}
	return nil
}

func (e *Engine) Ping(ctx context.Context) error {
	res, err := e.c.Send(ctx, cmdIsReady{})
	if err != nil {
		return fmt.Errorf("send \"isready\": %w", err)
	}
	ch := res.(cmdIsReadyRes)
	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		return fmt.Errorf("wait: %w", ctx.Err())
	case <-e.Done():
		return errTerminated
	}
}

func (e *Engine) SetOption(ctx context.Context, name string, value OptValue) error {
	if _, err := e.c.Send(ctx, cmdSetOption{name: name, value: value}); err != nil {
		return fmt.Errorf("send \"setoption\": %w", err)
	}
	return nil
}

func (e *Engine) UCINewGame(ctx context.Context, wait bool) error {
	if _, err := e.c.Send(ctx, cmdUCINewGame{}); err != nil {
		return fmt.Errorf("send \"ucinewgame\": %w", err)
	}
	if wait {
		if err := e.Ping(ctx); err != nil {
			return fmt.Errorf("ping after \"ucinewgame\": %w", err)
		}
	}
	return nil
}

func (e *Engine) SetPosition(ctx context.Context, g *chess.Game) error {
	cmd := cmdPosition{
		start: g.StartPos(),
		moves: make([]chess.Move, g.Len()),
		board: g.CurBoard().Clone(),
	}
	for i := range g.Len() {
		cmd.moves[i] = g.MoveAt(i)
	}

	if _, err := e.c.Send(ctx, cmd); err != nil {
		return fmt.Errorf("send \"position\": %w", err)
	}
	return nil
}

func (e *Engine) Go(ctx context.Context, opts GoOptions, c InfoConsumer) (*Search, error) {
	res, err := e.c.Send(ctx, cmdGo{opts: opts.Clone(), c: c})
	if err != nil {
		return nil, fmt.Errorf("send \"go\": %w", err)
	}
	return &Search{
		s: res.(cmdGoRes),
		e: e,
	}, nil
}

func (s *Search) Stop(ctx context.Context, wait bool) error {
	if _, err := s.e.c.Send(ctx, cmdStop{s: s.s}); err != nil {
		return fmt.Errorf("send \"stop\": %w", err)
	}
	if wait {
		if err := s.Wait(ctx); err != nil {
			return fmt.Errorf("wait for stop: %w", err)
		}
	}
	return nil
}

func (s *Search) PonderHit(ctx context.Context) error {
	if _, err := s.e.c.Send(ctx, cmdPonderHit{s: s.s}); err != nil {
		return fmt.Errorf("send \"stop\": %w", err)
	}
	return nil
}

func (e *Engine) Quit(ctx context.Context, wait bool) error {
	defer func() {
		if wait {
			e.Close()
		}
	}()

	if _, err := e.c.Send(ctx, cmdQuit{}); err != nil {
		select {
		case <-e.Done():
			return nil
		default:
		}
		return fmt.Errorf("send \"quit\": %w", err)
	}

	if !wait {
		return nil
	}
	select {
	case <-ctx.Done():
		return fmt.Errorf("wait: %w", ctx.Err())
	case <-e.Done():
		return nil
	}
}

func (e *Engine) SetPonder(ctx context.Context, val bool) error {
	return e.SetOption(ctx, ponderOptName, OptValueBool(val))
}

func (e *Engine) Terminated() bool {
	select {
	case <-e.Done():
		return true
	default:
		return false
	}
}

func (e *Engine) CurSearch() *Search {
	s := e.s.CurSearch()
	if s == nil {
		return nil
	}
	return &Search{s: s, e: e}
}

func (e *Engine) Info() (EngineInfo, bool)  { return e.s.Info() }
func (e *Engine) Initialized() bool         { return e.s.Initialized() }
func (e *Engine) Terminating() bool         { return e.s.Terminating() }
func (e *Engine) Debug() bool               { return e.s.Debug() }
func (e *Engine) GetOpt(name string) Option { return e.s.GetOpt(name) }
func (e *Engine) ListOpts() []string        { return e.s.ListOpts() }
func (e *Engine) PonderSupported() bool     { return e.s.PonderSupported() }
func (e *Engine) Ponder() bool              { return e.s.Ponder() }

func (s *Search) Done() <-chan struct{}                 { return s.s.Done() }
func (s *Search) Err() error                            { return s.s.Err() }
func (s *Search) Status() SearchStatus                  { return s.s.Status() }
func (s *Search) Ponder() bool                          { return s.s.Ponder() }
func (s *Search) Stopping() bool                        { return s.s.Stopping() }
func (s *Search) Stopped() bool                         { return s.s.Stopped() }
func (s *Search) BestMove() (chess.Move, error)         { return s.s.BestMove() }
func (s *Search) PonderMove() (chess.Move, bool, error) { return s.s.PonderMove() }
