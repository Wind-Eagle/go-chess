package clock

import (
	"fmt"
	"time"

	"github.com/alex65536/go-chess/chess"
	"github.com/alex65536/go-chess/util/maybe"
)

type GameOptions struct {
	OutcomeFilter maybe.Maybe[chess.VerdictFilter]
	Now           func() time.Time
}

type Game struct {
	filter chess.VerdictFilter
	game   *chess.Game
	timer  *Timer
}

func NewGame(game *chess.Game, control maybe.Maybe[Control], o GameOptions) *Game {
	filter := o.OutcomeFilter.GetOr(chess.VerdictFilterStrict)
	game = game.Clone()
	game.SetAutoOutcome(filter)
	var timer *Timer
	if control.IsSome() {
		to := TimerOptions{
			NumFlips: game.Len(),
			Outcome:  game.Outcome(),
			Now:      o.Now,
		}
		if to.Outcome.IsFinished() && to.NumFlips != 0 {
			to.NumFlips--
		}
		timer = NewTimer(game.StartPos().Side, control.Get(), to)
	}
	return &Game{
		filter: filter,
		game:   game,
		timer:  timer,
	}
}

func (g *Game) Inner() *chess.Game     { return g.game }
func (g *Game) CurBoard() *chess.Board { return g.game.CurBoard() }
func (g *Game) CurSide() chess.Color   { return g.game.CurBoard().Side() }
func (g *Game) Outcome() chess.Outcome { return g.game.Outcome() }
func (g *Game) IsFinished() bool       { return g.game.IsFinished() }
func (g *Game) HasTimer() bool         { return g.timer != nil }

func (g *Game) UCITimeSpec() (UCITimeSpec, bool) {
	if g.timer == nil {
		return UCITimeSpec{}, false
	}
	return g.timer.UCITimeSpec(), true
}

func (g *Game) Clock() (Clock, bool) {
	if g.timer == nil {
		return Clock{}, false
	}
	return g.timer.Clock(), true
}

func (g *Game) Deadline() (time.Time, bool) {
	if g.timer == nil {
		return time.Time{}, false
	}
	return g.timer.Deadline()
}

func (g *Game) UpdateTimer() {
	if g.timer == nil || g.IsFinished() {
		return
	}
	g.timer.Update()
	if o := g.timer.Outcome(); o.IsFinished() {
		g.game.SetOutcome(o)
	}
}

func (g *Game) Finish(o chess.Outcome) error {
	if g.IsFinished() {
		return fmt.Errorf("game already finished")
	}
	if !o.IsFinished() {
		return fmt.Errorf("outcome must finish the game")
	}
	g.game.SetOutcome(o)
	g.timer.Stop(o)
	if to := g.timer.Outcome(); o != to {
		g.game.SetOutcome(to)
	}
	return nil
}

func (g *Game) Push(mv chess.Move) error {
	if g.IsFinished() {
		return fmt.Errorf("game already finished")
	}
	if err := g.game.PushMove(mv); err != nil {
		return fmt.Errorf("add move: %w", err)
	}
	g.game.SetAutoOutcome(g.filter)
	if g.timer != nil {
		o := g.Outcome()
		if !o.IsFinished() {
			g.timer.Flip()
		} else {
			g.timer.Stop(o)
		}
		if to := g.timer.Outcome(); o != to {
			g.game.Pop()
			g.game.SetOutcome(to)
		}
	}
	return nil
}
