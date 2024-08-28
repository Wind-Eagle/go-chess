package clock

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/alex65536/go-chess/chess"
)

type ControlItem struct {
	Time  time.Duration
	Inc   time.Duration
	Moves int
}

func doControlItemFromString(s string) (ControlItem, error) {
	c := ControlItem{
		Time:  0,
		Inc:   0,
		Moves: 0,
	}
	if pos := strings.IndexByte(s, '/'); pos >= 0 {
		l := s[:pos]
		s = s[pos+1:]
		m, err := strconv.ParseInt(l, 10, 0)
		if err != nil {
			return ControlItem{}, fmt.Errorf("parse moves: %w", err)
		}
		c.Moves = int(m)
	}
	var err error
	if pos := strings.IndexByte(s, '+'); pos >= 0 {
		r := s[pos+1:]
		s = s[:pos]
		c.Inc, err = parseDuration(r)
		if err != nil {
			return ControlItem{}, fmt.Errorf("parse inc: %w", err)
		}
	}
	c.Time, err = parseDuration(s)
	if err != nil {
		return ControlItem{}, fmt.Errorf("parse time: %w", err)
	}
	return c, nil
}

func ControlItemFromString(s string, isFinal bool) (ControlItem, error) {
	c, err := doControlItemFromString(s)
	if err != nil {
		return ControlItem{}, err
	}
	if err := c.Validate(isFinal); err != nil {
		return ControlItem{}, fmt.Errorf("validate: %w", err)
	}
	return c, nil
}

func (c ControlItem) Validate(isFinal bool) error {
	if c.Time < 0 {
		return fmt.Errorf("negative time")
	}
	if c.Inc < 0 {
		return fmt.Errorf("negative inc")
	}
	if c.Moves < 0 {
		return fmt.Errorf("negative moves")
	}
	if !isFinal && c.Moves == 0 {
		return fmt.Errorf("number of moves must be specified for non-final controls")
	}
	return nil
}

func (c ControlItem) String() string {
	var b strings.Builder
	if c.Moves != 0 {
		_, _ = b.WriteString(strconv.FormatInt(int64(c.Moves), 10))
		_ = b.WriteByte('/')
	}
	_, _ = b.WriteString(formatDuration(c.Time))
	if c.Inc != 0 {
		_ = b.WriteByte('+')
		_, _ = b.WriteString(formatDuration(c.Inc))
	}
	return b.String()
}

type ControlSide []ControlItem

func (c ControlSide) Clone() ControlSide {
	return slices.Clone(c)
}

func (c ControlSide) Eq(o ControlSide) bool {
	return slices.Equal(c, o)
}

func ControlSideFromString(s string) (ControlSide, error) {
	if s == "" {
		return ControlSide{}, fmt.Errorf("empty string")
	}
	spl := strings.Split(s, ":")
	c := make(ControlSide, len(spl))
	for i, sub := range spl {
		var err error
		c[i], err = doControlItemFromString(sub)
		if err != nil {
			return nil, fmt.Errorf("parse section #%v: %w", i+1, err)
		}
	}
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}
	return c, nil
}

func (c ControlSide) Validate() error {
	if len(c) == 0 {
		return fmt.Errorf("no time control")
	}
	for i, item := range c {
		if err := item.Validate(i == len(c)-1); err != nil {
			return fmt.Errorf("section #%v: %w", i+1, err)
		}
	}
	if c[0].Time == 0 {
		return fmt.Errorf("initial time must be positive")
	}
	return nil
}

func (c ControlSide) String() string {
	var b strings.Builder
	for i, item := range c {
		if i != 0 {
			_ = b.WriteByte(':')
		}
		_, _ = b.WriteString(item.String())
	}
	return b.String()
}

type Control struct {
	White ControlSide
	Black ControlSide
}

func (c Control) Clone() Control {
	c.White = c.White.Clone()
	c.Black = c.Black.Clone()
	return c
}

func (c Control) Eq(o Control) bool {
	return c.White.Eq(o.White) && c.Black.Eq(o.Black)
}

func (c *Control) Side(col chess.Color) *ControlSide {
	if col == chess.ColorWhite {
		return &c.White
	} else {
		return &c.Black
	}
}

func ControlFromString(s string) (Control, error) {
	var err error
	var c Control
	if pos := strings.IndexByte(s, '|'); pos >= 0 {
		c.White, err = ControlSideFromString(s[:pos])
		if err != nil {
			return Control{}, fmt.Errorf("parse white: %w", err)
		}
		c.Black, err = ControlSideFromString(s[pos+1:])
		if err != nil {
			return Control{}, fmt.Errorf("parse black: %w", err)
		}
	} else {
		c.White, err = ControlSideFromString(s)
		if err != nil {
			return Control{}, fmt.Errorf("parse control: %w", err)
		}
		c.Black = c.White.Clone()
	}
	return c, nil
}

func (c Control) Validate() error {
	if err := c.White.Validate(); err != nil {
		return fmt.Errorf("white: %w", err)
	}
	if err := c.Black.Validate(); err != nil {
		return fmt.Errorf("black: %w", err)
	}
	return nil
}

func (c Control) String() string {
	if c.White.Eq(c.Black) {
		return c.White.String()
	}
	return c.White.String() + "|" + c.Black.String()
}
