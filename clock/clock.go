package clock

import (
	"time"

	"github.com/alex65536/go-chess/chess"
)

// TODO: use UCITimeSpec in "go" command in "uci" module
// TODO: validate
type UCITimeSpec struct {
	Wtime     time.Duration
	Btime     time.Duration
	Winc      time.Duration
	Binc      time.Duration
	MovesToGo int
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
