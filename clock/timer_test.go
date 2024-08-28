package clock

import (
	"testing"
	"time"

	"github.com/alex65536/go-chess/chess"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimerSimple(t *testing.T) {
	now, err := time.Parse(time.RFC3339, "2012-09-28T12:34:56Z")
	require.NoError(t, err)

	c, err := ControlFromString("4/60+2")
	require.NoError(t, err)

	timer := NewTimer(chess.ColorWhite, c, TimerOptions{
		Now: func() time.Time { return now },
	})

	assert.False(t, timer.Outcome().IsFinished())
	assert.Equal(t, Clock{White: 1 * time.Minute, Black: 1 * time.Minute}, timer.Clock())
	now = now.Add(6 * time.Second)
	assert.Equal(t, Clock{White: 54 * time.Second, Black: 1 * time.Minute}, timer.Clock())
	assert.Equal(t, chess.ColorWhite, timer.Side())
	timer.Flip()
	assert.Equal(t, chess.ColorBlack, timer.Side())
	assert.Equal(t, Clock{White: 56 * time.Second, Black: 1 * time.Minute}, timer.Clock())
	now = now.Add(3 * time.Second)
	dl, ok := timer.Deadline()
	assert.True(t, ok)
	assert.Equal(t, now.Add(57*time.Second), dl)
	assert.Equal(t, Clock{White: 56 * time.Second, Black: 57 * time.Second}, timer.Clock())
	timer.Flip()
	assert.Equal(t, Clock{White: 56 * time.Second, Black: 59 * time.Second}, timer.Clock())
	now = now.Add(9 * time.Second)
	assert.Equal(t, Clock{White: 47 * time.Second, Black: 59 * time.Second}, timer.Clock())
	timer.Flip()
	uci := UCITimeSpec{
		Wtime:     49 * time.Second,
		Btime:     59 * time.Second,
		Winc:      2 * time.Second,
		Binc:      2 * time.Second,
		MovesToGo: 3,
	}
	assert.Equal(t, uci, timer.UCITimeSpec())
	assert.Equal(t, Clock{White: 49 * time.Second, Black: 59 * time.Second}, timer.Clock())
	timer.Flip()
	assert.Equal(t, Clock{White: 49 * time.Second, Black: 61 * time.Second}, timer.Clock())
	timer.Flip()
	assert.Equal(t, Clock{White: 51 * time.Second, Black: 61 * time.Second}, timer.Clock())
	timer.Flip()
	assert.Equal(t, Clock{White: 51 * time.Second, Black: 63 * time.Second}, timer.Clock())
	timer.Flip()
	assert.Equal(t, Clock{White: 113 * time.Second, Black: 63 * time.Second}, timer.Clock())
	timer.Flip()
	assert.Equal(t, Clock{White: 113 * time.Second, Black: 125 * time.Second}, timer.Clock())
	timer.Flip()
	assert.Equal(t, Clock{White: 115 * time.Second, Black: 125 * time.Second}, timer.Clock())
	timer.Flip()
	assert.Equal(t, Clock{White: 115 * time.Second, Black: 127 * time.Second}, timer.Clock())
	timer.Flip()
	assert.Equal(t, Clock{White: 117 * time.Second, Black: 127 * time.Second}, timer.Clock())
	timer.Flip()
	assert.Equal(t, Clock{White: 117 * time.Second, Black: 129 * time.Second}, timer.Clock())
	timer.Flip()
	assert.Equal(t, Clock{White: 119 * time.Second, Black: 129 * time.Second}, timer.Clock())
	timer.Flip()
	assert.Equal(t, Clock{White: 119 * time.Second, Black: 131 * time.Second}, timer.Clock())
	timer.Flip()
	assert.Equal(t, Clock{White: 181 * time.Second, Black: 131 * time.Second}, timer.Clock())
	timer.Flip()
	assert.Equal(t, Clock{White: 181 * time.Second, Black: 193 * time.Second}, timer.Clock())
	require.False(t, timer.Outcome().IsFinished())
	now = now.Add(180 * time.Second)
	assert.Equal(t, Clock{White: 1 * time.Second, Black: 193 * time.Second}, timer.Clock())
	now = now.Add(2 * time.Second)
	assert.Equal(t, Clock{White: -1 * time.Second, Black: 193 * time.Second}, timer.Clock())
	timer.Flip()
	assert.Equal(t, Clock{White: -1 * time.Second, Black: 193 * time.Second}, timer.Clock())
	assert.Equal(t, chess.MustWinOutcome(chess.VerdictTimeForfeit, chess.ColorBlack), timer.Outcome())
	timer.Flip()
	assert.Equal(t, Clock{White: -1 * time.Second, Black: 193 * time.Second}, timer.Clock())
	assert.Equal(t, chess.MustWinOutcome(chess.VerdictTimeForfeit, chess.ColorBlack), timer.Outcome())
	timer.Stop(chess.MustDrawOutcome(chess.VerdictDrawAgreement))
	assert.Equal(t, Clock{White: -1 * time.Second, Black: 193 * time.Second}, timer.Clock())
	assert.Equal(t, chess.MustWinOutcome(chess.VerdictTimeForfeit, chess.ColorBlack), timer.Outcome())
	assert.Equal(t, chess.ColorWhite, timer.Side())
	_, ok = timer.Deadline()
	assert.False(t, ok)
}

func TestMultiControl(t *testing.T) {
	now, err := time.Parse(time.RFC3339, "2012-09-28T12:34:56Z")
	require.NoError(t, err)

	c, err := ControlFromString("3/60+2:2/20+5:120")
	require.NoError(t, err)
	require.NoError(t, c.Validate())

	timer := NewTimer(chess.ColorWhite, c, TimerOptions{
		Now: func() time.Time { return now },
	})

	assert.Equal(t, UCITimeSpec{
		Wtime:     60 * time.Second,
		Btime:     60 * time.Second,
		Winc:      2 * time.Second,
		Binc:      2 * time.Second,
		MovesToGo: 3,
	}, timer.UCITimeSpec())

	timer.Flip()
	assert.Equal(t, Clock{White: 62 * time.Second, Black: 60 * time.Second}, timer.Clock())
	timer.Flip()
	assert.Equal(t, Clock{White: 62 * time.Second, Black: 62 * time.Second}, timer.Clock())
	timer.Flip()
	assert.Equal(t, Clock{White: 64 * time.Second, Black: 62 * time.Second}, timer.Clock())
	timer.Flip()
	assert.Equal(t, Clock{White: 64 * time.Second, Black: 64 * time.Second}, timer.Clock())
	assert.Equal(t, UCITimeSpec{
		Wtime:     64 * time.Second,
		Btime:     64 * time.Second,
		Winc:      2 * time.Second,
		Binc:      2 * time.Second,
		MovesToGo: 1,
	}, timer.UCITimeSpec())

	timer.Flip()
	assert.Equal(t, Clock{White: 86 * time.Second, Black: 64 * time.Second}, timer.Clock())
	assert.Equal(t, UCITimeSpec{
		Wtime:     86 * time.Second,
		Btime:     64 * time.Second,
		Winc:      5 * time.Second,
		Binc:      2 * time.Second,
		MovesToGo: 1,
	}, timer.UCITimeSpec())

	timer.Flip()
	assert.Equal(t, Clock{White: 86 * time.Second, Black: 86 * time.Second}, timer.Clock())
	assert.Equal(t, UCITimeSpec{
		Wtime:     86 * time.Second,
		Btime:     86 * time.Second,
		Winc:      5 * time.Second,
		Binc:      5 * time.Second,
		MovesToGo: 2,
	}, timer.UCITimeSpec())

	timer.Flip()
	assert.Equal(t, Clock{White: 91 * time.Second, Black: 86 * time.Second}, timer.Clock())
	timer.Flip()
	assert.Equal(t, Clock{White: 91 * time.Second, Black: 91 * time.Second}, timer.Clock())

	timer.Flip()
	assert.Equal(t, Clock{White: 216 * time.Second, Black: 91 * time.Second}, timer.Clock())
	assert.Equal(t, UCITimeSpec{
		Wtime:     216 * time.Second,
		Btime:     91 * time.Second,
		Winc:      0,
		Binc:      5 * time.Second,
		MovesToGo: 1,
	}, timer.UCITimeSpec())

	timer.Flip()
	assert.Equal(t, Clock{White: 216 * time.Second, Black: 216 * time.Second}, timer.Clock())
	assert.Equal(t, UCITimeSpec{
		Wtime:     216 * time.Second,
		Btime:     216 * time.Second,
		Winc:      0,
		Binc:      0,
		MovesToGo: 0,
	}, timer.UCITimeSpec())

	for range 100 {
		timer.Flip()
		assert.Equal(t, Clock{White: 216 * time.Second, Black: 216 * time.Second}, timer.Clock())
		assert.Equal(t, UCITimeSpec{
			Wtime:     216 * time.Second,
			Btime:     216 * time.Second,
			Winc:      0,
			Binc:      0,
			MovesToGo: 0,
		}, timer.UCITimeSpec())
	}

	assert.False(t, timer.Outcome().IsFinished())
}

func TestStop(t *testing.T) {
	now, err := time.Parse(time.RFC3339, "2012-09-28T12:34:56Z")
	require.NoError(t, err)

	c, err := ControlFromString("60+1")
	require.NoError(t, err)
	require.NoError(t, c.Validate())

	timer := NewTimer(chess.ColorWhite, c, TimerOptions{
		Now: func() time.Time { return now },
	})

	timer.Flip()
	timer.Flip()
	now = now.Add(6 * time.Second)
	assert.Equal(t, Clock{White: 55 * time.Second, Black: 61 * time.Second}, timer.Clock())
	assert.Equal(t, chess.ColorWhite, timer.Side())
	timer.Stop(chess.MustDrawOutcome(chess.VerdictDrawAgreement))
	assert.Equal(t, chess.MustDrawOutcome(chess.VerdictDrawAgreement), timer.Outcome())

	now = now.Add(10 * time.Second)
	assert.Equal(t, Clock{White: 55 * time.Second, Black: 61 * time.Second}, timer.Clock())
	timer.Flip()
	assert.Equal(t, chess.ColorWhite, timer.Side())
	timer.Flip()
	assert.Equal(t, chess.ColorWhite, timer.Side())
	timer.Flip()
	assert.Equal(t, chess.ColorWhite, timer.Side())
	assert.Equal(t, Clock{White: 55 * time.Second, Black: 61 * time.Second}, timer.Clock())
	assert.Equal(t, chess.MustDrawOutcome(chess.VerdictDrawAgreement), timer.Outcome())

	now = now.Add(5 * time.Second)
	timer.Update()
	assert.Equal(t, Clock{White: 55 * time.Second, Black: 61 * time.Second}, timer.Clock())
}

func TestRacyStop(t *testing.T) {
	now, err := time.Parse(time.RFC3339, "2012-09-28T12:34:56Z")
	require.NoError(t, err)

	c, err := ControlFromString("60+1")
	require.NoError(t, err)
	require.NoError(t, c.Validate())

	timer := NewTimer(chess.ColorWhite, c, TimerOptions{
		Now: func() time.Time { return now },
	})

	timer.Flip()
	timer.Flip()
	timer.Flip()
	now = now.Add(64 * time.Second)
	assert.Equal(t, Clock{White: 62 * time.Second, Black: -3 * time.Second}, timer.Clock())
	timer.Stop(chess.MustDrawOutcome(chess.VerdictDrawAgreement))
	assert.Equal(t, chess.MustWinOutcome(chess.VerdictTimeForfeit, chess.ColorWhite), timer.Outcome())

	timer.Flip()
	assert.Equal(t, chess.ColorBlack, timer.Side())
	timer.Flip()
	assert.Equal(t, chess.ColorBlack, timer.Side())
	now = now.Add(64 * time.Second)
	assert.Equal(t, Clock{White: 62 * time.Second, Black: -3 * time.Second}, timer.Clock())
}

func TestUpdate(t *testing.T) {
	now, err := time.Parse(time.RFC3339, "2012-09-28T12:34:56Z")
	require.NoError(t, err)

	c, err := ControlFromString("60")
	require.NoError(t, err)
	require.NoError(t, c.Validate())

	timer := NewTimer(chess.ColorWhite, c, TimerOptions{
		Now: func() time.Time { return now },
	})

	assert.Equal(t, Clock{White: 60 * time.Second, Black: 60 * time.Second}, timer.Clock())

	now = now.Add(3 * time.Second)
	assert.Equal(t, Clock{White: 57 * time.Second, Black: 60 * time.Second}, timer.Clock())
	assert.False(t, timer.Outcome().IsFinished())
	timer.Update()
	assert.Equal(t, Clock{White: 57 * time.Second, Black: 60 * time.Second}, timer.Clock())
	assert.False(t, timer.Outcome().IsFinished())

	now = now.Add(5 * time.Second)
	assert.Equal(t, Clock{White: 52 * time.Second, Black: 60 * time.Second}, timer.Clock())
	assert.False(t, timer.Outcome().IsFinished())
	timer.Update()
	assert.Equal(t, Clock{White: 52 * time.Second, Black: 60 * time.Second}, timer.Clock())
	assert.False(t, timer.Outcome().IsFinished())

	now = now.Add(54 * time.Second)
	assert.Equal(t, Clock{White: -2 * time.Second, Black: 60 * time.Second}, timer.Clock())
	assert.False(t, timer.Outcome().IsFinished())
	timer.Update()
	assert.Equal(t, Clock{White: -2 * time.Second, Black: 60 * time.Second}, timer.Clock())
	assert.True(t, timer.Outcome().IsFinished())
}

func TestSplit(t *testing.T) {
	now, err := time.Parse(time.RFC3339, "2012-09-28T12:34:56Z")
	require.NoError(t, err)

	c, err := ControlFromString("3/20+1:40+2|4/30+3:50+4")
	require.NoError(t, err)
	require.NoError(t, c.Validate())

	timer := NewTimer(chess.ColorWhite, c, TimerOptions{
		Now: func() time.Time { return now },
	})

	assert.Equal(t, Clock{White: 20 * time.Second, Black: 30 * time.Second}, timer.Clock())
	assert.Equal(t, UCITimeSpec{
		Wtime:     20 * time.Second,
		Btime:     30 * time.Second,
		Winc:      1 * time.Second,
		Binc:      3 * time.Second,
		MovesToGo: 3,
	}, timer.UCITimeSpec())

	timer.Flip()
	assert.Equal(t, Clock{White: 21 * time.Second, Black: 30 * time.Second}, timer.Clock())
	assert.Equal(t, UCITimeSpec{
		Wtime:     21 * time.Second,
		Btime:     30 * time.Second,
		Winc:      1 * time.Second,
		Binc:      3 * time.Second,
		MovesToGo: 4,
	}, timer.UCITimeSpec())

	timer.Flip()
	assert.Equal(t, Clock{White: 21 * time.Second, Black: 33 * time.Second}, timer.Clock())
	assert.Equal(t, UCITimeSpec{
		Wtime:     21 * time.Second,
		Btime:     33 * time.Second,
		Winc:      1 * time.Second,
		Binc:      3 * time.Second,
		MovesToGo: 2,
	}, timer.UCITimeSpec())

	timer.Flip()
	assert.Equal(t, Clock{White: 22 * time.Second, Black: 33 * time.Second}, timer.Clock())

	timer.Flip()
	assert.Equal(t, Clock{White: 22 * time.Second, Black: 36 * time.Second}, timer.Clock())
	assert.Equal(t, UCITimeSpec{
		Wtime:     22 * time.Second,
		Btime:     36 * time.Second,
		Winc:      1 * time.Second,
		Binc:      3 * time.Second,
		MovesToGo: 1,
	}, timer.UCITimeSpec())

	timer.Flip()
	assert.Equal(t, Clock{White: 63 * time.Second, Black: 36 * time.Second}, timer.Clock())
	assert.Equal(t, UCITimeSpec{
		Wtime:     63 * time.Second,
		Btime:     36 * time.Second,
		Winc:      2 * time.Second,
		Binc:      3 * time.Second,
		MovesToGo: 2,
	}, timer.UCITimeSpec())

	timer.Flip()
	assert.Equal(t, Clock{White: 63 * time.Second, Black: 39 * time.Second}, timer.Clock())
	assert.Equal(t, UCITimeSpec{
		Wtime:     63 * time.Second,
		Btime:     39 * time.Second,
		Winc:      2 * time.Second,
		Binc:      3 * time.Second,
		MovesToGo: 0,
	}, timer.UCITimeSpec())

	timer.Flip()
	assert.Equal(t, Clock{White: 65 * time.Second, Black: 39 * time.Second}, timer.Clock())
	assert.Equal(t, UCITimeSpec{
		Wtime:     65 * time.Second,
		Btime:     39 * time.Second,
		Winc:      2 * time.Second,
		Binc:      3 * time.Second,
		MovesToGo: 1,
	}, timer.UCITimeSpec())

	timer.Flip()
	assert.Equal(t, Clock{White: 65 * time.Second, Black: 92 * time.Second}, timer.Clock())
	assert.Equal(t, UCITimeSpec{
		Wtime:     65 * time.Second,
		Btime:     92 * time.Second,
		Winc:      2 * time.Second,
		Binc:      4 * time.Second,
		MovesToGo: 0,
	}, timer.UCITimeSpec())

	timer.Flip()
	assert.Equal(t, Clock{White: 67 * time.Second, Black: 92 * time.Second}, timer.Clock())
	assert.Equal(t, UCITimeSpec{
		Wtime:     67 * time.Second,
		Btime:     92 * time.Second,
		Winc:      2 * time.Second,
		Binc:      4 * time.Second,
		MovesToGo: 0,
	}, timer.UCITimeSpec())

	timer.Flip()
	assert.Equal(t, Clock{White: 67 * time.Second, Black: 96 * time.Second}, timer.Clock())
}

func TestPreflipTricky(t *testing.T) {
	now, err := time.Parse(time.RFC3339, "2012-09-28T12:34:56Z")
	require.NoError(t, err)

	c, err := ControlFromString("3/20+1:40+2|4/30+3:50+4")
	require.NoError(t, err)
	require.NoError(t, c.Validate())

	timer := NewTimer(chess.ColorWhite, c, TimerOptions{
		NumFlips: 9,
		Now: func() time.Time {
			now = now.Add(1 * time.Minute)
			return now
		},
	})
	assert.Equal(t, Clock{White: 67 * time.Second, Black: 32 * time.Second}, timer.Clock())
	assert.False(t, timer.Outcome().IsFinished())
}

func TestPreflip(t *testing.T) {
	now, err := time.Parse(time.RFC3339, "2012-09-28T12:34:56Z")
	require.NoError(t, err)

	c, err := ControlFromString("3/20+1:40+2|4/30+3:50+4")
	require.NoError(t, err)
	require.NoError(t, c.Validate())

	timer := NewTimer(chess.ColorWhite, c, TimerOptions{
		NumFlips: 9,
		Now:      func() time.Time { return now },
	})
	assert.Equal(t, Clock{White: 67 * time.Second, Black: 92 * time.Second}, timer.Clock())
	assert.False(t, timer.Outcome().IsFinished())

	timer = NewTimer(chess.ColorWhite, c, TimerOptions{
		NumFlips: 9,
		Outcome:  chess.MustDrawOutcome(chess.VerdictDrawAgreement),
		Now:      func() time.Time { return now },
	})
	assert.Equal(t, Clock{White: 67 * time.Second, Black: 92 * time.Second}, timer.Clock())
	assert.True(t, timer.Outcome().IsFinished())
	assert.Equal(t, chess.MustDrawOutcome(chess.VerdictDrawAgreement), timer.Outcome())

	timer.Flip()
	timer.Flip()
	timer.Flip()
	now = now.Add(1 * time.Millisecond)
	assert.Equal(t, Clock{White: 67 * time.Second, Black: 92 * time.Second}, timer.Clock())
}

func TestHugeInc(t *testing.T) {
	now, err := time.Parse(time.RFC3339, "2012-09-28T12:34:56Z")
	require.NoError(t, err)

	c, err := ControlFromString("1+1000")
	require.NoError(t, err)
	require.NoError(t, c.Validate())

	timer := NewTimer(chess.ColorWhite, c, TimerOptions{
		Now: func() time.Time { return now },
	})
	assert.Equal(t, Clock{White: time.Second, Black: time.Second}, timer.Clock())
	now = now.Add(2 * time.Second)
	timer.Flip()
	assert.Equal(t, Clock{White: -time.Second, Black: time.Second}, timer.Clock())
	assert.True(t, timer.Outcome().IsFinished())
}
