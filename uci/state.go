package uci

import (
	"container/list"
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/alex65536/go-chess/chess"
	"github.com/alex65536/go-chess/util/maybe"
)

type EngineInfo struct {
	Name   string
	Author string
}

type engineState struct {
	o EngineOptions
	l Logger

	initedCh chan struct{}

	mu      sync.RWMutex
	inited  bool
	exiting bool
	exited  bool
	search  *searchState
	pongs   *list.List
	board   *chess.Board
	info    EngineInfo
	debug   bool
	opts    map[string]optPair
}

func newEngineState(o EngineOptions, l Logger) *engineState {
	return &engineState{
		o: o,
		l: l,

		initedCh: make(chan struct{}),

		inited:  false,
		exiting: false,
		exited:  false,
		search:  nil,
		pongs:   list.New(),
		board:   nil,
		info:    EngineInfo{},
		debug:   false,
		opts:    make(map[string]optPair),
	}
}

func (s *engineState) Start(_ context.Context, p Process) error {
	if err := p.Send("uci"); err != nil {
		return fmt.Errorf("send \"uci\": %w", err)
	}
	return nil
}

func (s *engineState) ProcessCommand(scmd command) (command, any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.exited || !s.inited {
		panic("must not happen")
	}

	if _, ok := scmd.(cmdQuit); ok && s.exiting {
		return cmdEmpty{}, nil, nil
	}
	if s.exiting {
		return nil, nil, fmt.Errorf("engine terminating")
	}

	switch cmd := scmd.(type) {
	case cmdDebug:
		s.debug = cmd.val
		return cmd, nil, nil
	case cmdIsReady:
		ch := make(chan error, 1)
		s.pongs.PushFront(ch)
		return cmd, cmdIsReadyRes(ch), nil
	case cmdSetOption:
		opt, ok := s.opts[caseFold(cmd.name)]
		if !ok {
			return nil, nil, fmt.Errorf("unknown option %q", cmd.name)
		}
		if err := opt.value.setValue(cmd.value, s.o.coderOptions()); err != nil {
			return nil, nil, fmt.Errorf("set option %q", cmd.name)
		}
		cmd = cmdSetOption{
			name:  opt.name,
			value: opt.value.Value(),
		}
		return cmd, nil, nil
	case cmdUCINewGame:
		if s.search != nil {
			return nil, nil, fmt.Errorf("engine must not be searching")
		}
		s.board = nil
		return cmd, nil, nil
	case cmdPosition:
		if s.search != nil {
			return nil, nil, fmt.Errorf("engine must not be searching")
		}
		s.board = cmd.board
		return cmd, nil, nil
	case cmdGo:
		if s.search != nil {
			return nil, nil, fmt.Errorf("engine must not be searching")
		}
		if s.board == nil {
			return nil, nil, fmt.Errorf("no position specified")
		}
		if !s.doPonderEnabled() && cmd.opts.Ponder {
			return nil, nil, fmt.Errorf("pondering is not allowed")
		}
		if err := cmd.opts.Validate(s.board); err != nil {
			return nil, nil, fmt.Errorf("invalid options: %w", err)
		}
		s.search = newSearchState(cmd.c, s.l, s.board, cmd.opts.Ponder)
		return cmd, cmdGoRes(s.search), nil
	case cmdStop:
		if cmd.s == nil {
			return nil, nil, fmt.Errorf("nil search state")
		}
		if cmd.s != s.search {
			return cmdEmpty{}, nil, nil
		}
		if err := s.search.OnStop(); err != nil {
			return nil, nil, err
		}
		return cmd, nil, nil
	case cmdPonderHit:
		if cmd.s == nil {
			return nil, nil, fmt.Errorf("nil search state")
		}
		if cmd.s != s.search {
			return nil, nil, fmt.Errorf("search stopped")
		}
		if err := s.search.OnPonderHit(); err != nil {
			return nil, nil, err
		}
		return cmd, nil, nil
	case cmdQuit:
		s.exiting = true
		return cmd, nil, nil
	default:
		return nil, nil, fmt.Errorf("unrecognized command")
	}
}

func (s *engineState) ProcessMessage(msg string) error {
	tok, err := newTokenizer(msg, s.o.coderOptions())
	if err != nil {
		return fmt.Errorf("tokenize %q: %w", msg, err)
	}

restart:
	name, ok := tok.Next()
	if !ok {
		return nil
	}

	switch name {
	case "id":
		sub, ok := tok.Next()
		if !ok {
			return fmt.Errorf("parse \"id\": incomplete message")
		}
		switch sub {
		case "name":
			return s.onIdName(tok.NextUntilEnd())
		case "author":
			return s.onIdAuthor(tok.NextUntilEnd())
		default:
			return fmt.Errorf("parse \"id\": bad submessage %q", sub)
		}
	case "uciok":
		if tok.More() {
			s.l.Printf("parse \"uciok\": extra data")
		}
		return s.onUCIOk()
	case "readyok":
		if tok.More() {
			s.l.Printf("parse \"readyok\": extra data")
		}
		return s.onReadyOk()
	case "bestmove":
		ms, err := parseBestMove(tok, s.l)
		if err != nil {
			s.l.Printf("parse \"bestmove\": %v", err)
			return s.onSearchCancel(fmt.Errorf("parse: %w", err))
		}
		ponder := maybe.None[chess.UCIMove]()
		if len(ms) >= 2 {
			ponder = maybe.Some(ms[1])
		}
		return s.onBestMove(ms[0], ponder)
	case "copyprotection":
		return fmt.Errorf("\"copyprotection\" not implemented")
	case "registration":
		return fmt.Errorf("\"registration\" not implemented")
	case "info":
		i, err := parseInfo(tok, s.l)
		if err != nil {
			return fmt.Errorf("parse \"info\": %v", err)
		}
		return s.onInfo(i)
	case "option":
		o, err := parseOption(tok, s.l)
		if err != nil {
			return fmt.Errorf("parse \"option\": %v", err)
		}
		return s.onOption(o)
	default:
		goto restart
	}
}

func (s *engineState) Finish() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.search != nil {
		s.search.Cancel(errTerminated)
		s.search = nil
	}
	s.exited = true
	s.exiting = false
	for s.pongs.Len() != 0 {
		pong := s.pongs.Remove(s.pongs.Front()).(chan error)
		pong <- errTerminated
	}
	s.board = nil
}

func (s *engineState) onIdName(val string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.inited {
		return fmt.Errorf("cannot process \"id name\" after initialization")
	}
	s.info.Name = val
	return nil
}

func (s *engineState) onIdAuthor(val string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.inited {
		return fmt.Errorf("cannot process \"id author\" after initialization")
	}
	s.info.Author = val
	return nil
}

func (s *engineState) onUCIOk() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.inited {
		return fmt.Errorf("duplicate \"uciok\"")
	}
	s.inited = true
	close(s.initedCh)
	return nil
}

func (s *engineState) onReadyOk() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.inited {
		return fmt.Errorf("engine not initialized")
	}
	front := s.pongs.Front()
	if front == nil {
		return fmt.Errorf("unmatched \"readyok\"")
	}
	pong := s.pongs.Remove(front).(chan error)
	pong <- nil
	return nil
}

func (s *engineState) onOption(o optPair) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.inited {
		return fmt.Errorf("cannot process \"option\" after initialization")
	}
	folded := caseFold(o.name)
	if _, ok := s.opts[folded]; ok {
		return fmt.Errorf("duplicate option %q", o.name)
	}
	s.opts[folded] = o
	return nil
}

func (s *engineState) onInfo(info Info) error {
	strOnly := reflect.DeepEqual(info, Info{String: info.String})
	if str, ok := info.String.TryGet(); ok && s.o.LogEngineString {
		s.l.Printf("engine: %v", str)
	}

	search, err := func() (*searchState, error) {
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.search != nil {
			return s.search, nil
		}
		if strOnly {
			return nil, nil
		}
		return nil, fmt.Errorf("no search in progress")
	}()
	if err != nil {
		return err
	}

	if search != nil {
		return search.OnInfo(info, strOnly)
	}
	return nil
}

func (s *engineState) onSearchCancel(err error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.search == nil {
		return fmt.Errorf("no search in progress")
	}
	s.search.Cancel(err)
	s.search = nil
	return nil
}

func (s *engineState) onBestMove(best chess.UCIMove, ponder maybe.Maybe[chess.UCIMove]) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.search == nil {
		return fmt.Errorf("no search in progress")
	}
	s.search.OnBestMove(best, ponder)
	s.search = nil
	return nil
}

func (s *engineState) doPonderEnabled() bool {
	if !s.inited {
		return false
	}
	opt, ok := s.opts[ponderOptName]
	if !ok {
		return false
	}
	b, ok := opt.value.Value().(OptValueBool)
	if !ok {
		return false
	}
	return bool(b)
}

func (s *engineState) Info() (EngineInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.inited {
		return EngineInfo{}, false
	}
	return s.info, true
}

func (s *engineState) InitializedChan() <-chan struct{} { return s.initedCh }

func (s *engineState) Initialized() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.inited
}

func (s *engineState) Terminating() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.exiting
}

func (s *engineState) Debug() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.debug
}

func (s *engineState) GetOpt(name string) Option {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.inited {
		return nil
	}
	opt, ok := s.opts[caseFold(name)]
	if !ok {
		return nil
	}
	return opt.value.Clone()
}

func (s *engineState) ListOpts() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.inited {
		return nil
	}
	res := make([]string, 0, len(s.opts))
	for _, v := range s.opts {
		res = append(res, v.name)
	}
	return res
}

func (s *engineState) CurSearch() *searchState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.search
}

func (s *engineState) PonderSupported() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.inited {
		return false
	}
	opt, ok := s.opts[ponderOptName]
	if !ok {
		return false
	}
	_, ok = opt.value.Value().(OptValueBool)
	return ok
}

func (s *engineState) Ponder() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.doPonderEnabled()
}
