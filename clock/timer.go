package clock

import (
	"time"

	"github.com/alex65536/go-chess/chess"
)

type subController struct {
	control ControlSide
	left    int
}

func (c *subController) init(control ControlSide, d *time.Duration) {
	c.control = control.Clone()
	c.left = control[0].Moves
	*d = control[0].Time
}

func (c *subController) flip(d *time.Duration) {
	*d += c.control[0].Inc
	if c.control[0].Moves != 0 {
		c.left--
		if c.left == 0 {
			if len(c.control) > 1 {
				c.control = c.control[1:]
			}
			c.left = c.control[0].Moves
			*d += c.control[0].Time
		}
	}
}

type controller struct {
	sub [chess.ColorMax]subController
}

func (c *controller) init(control Control, clock *SimpleClock) {
	for color := range chess.ColorMax {
		c.sub[color].init(*control.Side(color), clock.Side(color))
	}
}

func (c *controller) flip(who chess.Color, clock *SimpleClock) {
	c.sub[who].flip(clock.Side(who))
}

func (c *controller) uciTimeSpec(who chess.Color, clock SimpleClock) UCITimeSpec {
	return UCITimeSpec{
		Wtime:     clock.White,
		Btime:     clock.Black,
		Winc:      c.sub[chess.ColorWhite].control[0].Inc,
		Binc:      c.sub[chess.ColorBlack].control[0].Inc,
		MovesToGo: c.sub[who].left,
	}
}

type Timer struct {
	side    chess.Color
	outcome chess.Outcome
	cur     time.Time
	nowFn   func() time.Time
	ctrl    controller
	clock   SimpleClock
}

type TimerOptions struct {
	NumFlips int
	Outcome  chess.Outcome
	Now      func() time.Time
}

func NewTimer(side chess.Color, control Control, o TimerOptions) *Timer {
	nowFn := o.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	t := &Timer{
		side:    side,
		outcome: chess.RunningOutcome(),
		cur:     nowFn(),
		nowFn:   nowFn,
	}
	t.ctrl.init(control, &t.clock)
	t.doCheckForfeit()
	for range o.NumFlips {
		if t.outcome.IsFinished() {
			break
		}
		t.doFlip()
		t.doCheckForfeit()
	}
	if o.Outcome.IsFinished() && !t.outcome.IsFinished() {
		t.outcome = o.Outcome
	}
	return t
}

func (t *Timer) Side() chess.Color      { return t.side }
func (t *Timer) Outcome() chess.Outcome { return t.outcome }

func (t *Timer) Clock() Clock {
	c := Clock{
		White:        t.clock.White,
		Black:        t.clock.Black,
		WhiteTicking: false,
		BlackTicking: false,
	}
	if t.outcome.IsFinished() {
		return c
	}
	*c.SideTicking(t.side) = true
	now := t.nowFn()
	if now.After(t.cur) {
		*c.Side(t.side) -= now.Sub(t.cur)
	}
	return c
}

func (t *Timer) Deadline() (time.Time, bool) {
	if t.outcome.IsFinished() {
		return time.Time{}, false
	}
	return t.cur.Add(*t.clock.Side(t.side)), true
}

func (t *Timer) doCheckForfeit() {
	if !t.outcome.IsFinished() && *t.clock.Side(t.side) <= 0 {
		t.outcome = chess.MustWinOutcome(chess.VerdictTimeForfeit, t.side.Inv())
	}
}

func (t *Timer) Update() {
	if t.outcome.IsFinished() {
		return
	}
	now := t.nowFn()
	if now.After(t.cur) {
		*t.clock.Side(t.side) -= now.Sub(t.cur)
		t.cur = now
	}
	t.doCheckForfeit()
}

func (t *Timer) doFlip() {
	t.ctrl.flip(t.side, &t.clock)
	t.side = t.side.Inv()
}

func (t *Timer) Flip() {
	if t.outcome.IsFinished() {
		return
	}
	t.Update()
	if t.outcome.IsFinished() {
		return
	}
	t.doFlip()
}

func (t *Timer) Stop(outcome chess.Outcome) {
	if !outcome.IsFinished() || t.outcome.IsFinished() {
		return
	}
	t.Update()
	if t.outcome.IsFinished() {
		return
	}
	t.outcome = outcome
}

func (t *Timer) UCITimeSpec() UCITimeSpec {
	return t.ctrl.uciTimeSpec(t.side, t.clock)
}
