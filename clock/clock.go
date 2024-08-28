package clock

import (
	"fmt"
	"time"

	"github.com/alex65536/go-chess/chess"
)

type UCITimeSpec struct {
	Wtime     time.Duration
	Btime     time.Duration
	Winc      time.Duration
	Binc      time.Duration
	MovesToGo int
}

func (s *UCITimeSpec) Validate() error {
	if s.Wtime <= 0 {
		return fmt.Errorf("non-positive wtime")
	}
	if s.Btime <= 0 {
		return fmt.Errorf("non-positive btime")
	}
	if s.Winc < 0 {
		return fmt.Errorf("negative winc")
	}
	if s.Binc < 0 {
		return fmt.Errorf("negative binc")
	}
	if s.MovesToGo < 0 {
		return fmt.Errorf("negative movestogo")
	}
	return nil
}

type Clock struct {
	White time.Duration
	Black time.Duration
}

func (c *Clock) Side(col chess.Color) *time.Duration {
	if col == chess.ColorWhite {
		return &c.White
	} else {
		return &c.Black
	}
}
