package clock

import (
	"testing"
	"time"

	"github.com/alex65536/go-chess/chess"
	"github.com/alex65536/go-chess/util/maybe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGame(t *testing.T) {
	now, err := time.Parse(time.RFC3339, "2012-09-28T12:34:56Z")
	require.NoError(t, err)

	c, err := ControlFromString("3/60+1")
	require.NoError(t, err)

	g := NewGame(chess.NewGame(), maybe.Some(c), GameOptions{
		Now: func() time.Time { return now },
	})
	clk, ok := g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 60 * time.Second, Black: 60 * time.Second, WhiteTicking: true}, clk)
	now = now.Add(time.Second)
	clk, ok = g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 59 * time.Second, Black: 60 * time.Second, WhiteTicking: true}, clk)

	cg := chess.NewGame()
	cg.SetOutcome(chess.MustDrawOutcome(chess.VerdictDrawAgreement))
	g = NewGame(cg, maybe.Some(c), GameOptions{
		Now: func() time.Time { return now },
	})
	clk, ok = g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 60 * time.Second, Black: 60 * time.Second}, clk)
	assert.Equal(t, chess.MustDrawOutcome(chess.VerdictDrawAgreement), g.Outcome())
	now = now.Add(time.Second)
	clk, ok = g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 60 * time.Second, Black: 60 * time.Second}, clk)

	cg, err = chess.GameFromUCIList(chess.InitialBoard(), "g2g4 e7e5 f2f3 d8h4")
	require.NoError(t, err)
	cg.SetOutcome(chess.RunningOutcome())
	g = NewGame(cg, maybe.Some(c), GameOptions{
		Now: func() time.Time { return now },
	})
	clk, ok = g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 62 * time.Second, Black: 61 * time.Second}, clk)
	assert.Equal(t, chess.MustWinOutcome(chess.VerdictCheckmate, chess.ColorBlack), g.Outcome())
	assert.True(t, g.IsFinished())

	cg, err = chess.GameFromUCIList(chess.InitialBoard(), "e2e4 e7e5 g1f3 b8c6")
	require.NoError(t, err)
	g = NewGame(cg, maybe.Some(c), GameOptions{
		Now: func() time.Time { return now },
	})
	clk, ok = g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 62 * time.Second, Black: 62 * time.Second, WhiteTicking: true}, clk)
	assert.Equal(t, chess.RunningOutcome(), g.Outcome())
	assert.False(t, g.IsFinished())
}

func TestGameSimple(t *testing.T) {
	now, err := time.Parse(time.RFC3339, "2012-09-28T12:34:56Z")
	require.NoError(t, err)

	c, err := ControlFromString("3/60+1")
	require.NoError(t, err)

	g := NewGame(chess.NewGame(), maybe.Some(c), GameOptions{
		Now: func() time.Time { return now },
	})
	clk, ok := g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 60 * time.Second, Black: 60 * time.Second, WhiteTicking: true}, clk)
	assert.True(t, g.HasTimer())

	mv, err := chess.MoveFromUCI("g2g4", g.CurBoard())
	require.NoError(t, err)
	now = now.Add(5 * time.Second)
	require.NoError(t, g.Push(mv))
	clk, ok = g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 56 * time.Second, Black: 60 * time.Second, BlackTicking: true}, clk)

	assert.Equal(t, chess.ColorBlack, g.CurSide())
	d, ok := g.Deadline()
	assert.True(t, ok)
	assert.Equal(t, now.Add(60*time.Second), d)
	assert.Equal(t, maybe.Some(UCITimeSpec{
		Wtime:     56 * time.Second,
		Btime:     60 * time.Second,
		Winc:      time.Second,
		Binc:      time.Second,
		MovesToGo: 3,
	}), maybe.Pack(g.UCITimeSpec()))

	mv, err = chess.MoveFromUCI("e7e5", g.CurBoard())
	require.NoError(t, err)
	now = now.Add(6 * time.Second)
	require.NoError(t, g.Push(mv))
	clk, ok = g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 56 * time.Second, Black: 55 * time.Second, WhiteTicking: true}, clk)

	mv, err = chess.MoveFromUCI("f2f3", g.CurBoard())
	require.NoError(t, err)
	now = now.Add(2 * time.Second)
	require.NoError(t, g.Push(mv))
	clk, ok = g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 55 * time.Second, Black: 55 * time.Second, BlackTicking: true}, clk)

	mv, err = chess.MoveFromUCI("d8h4", g.CurBoard())
	require.NoError(t, err)
	now = now.Add(2 * time.Second)
	require.NoError(t, g.Push(mv))
	clk, ok = g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 55 * time.Second, Black: 53 * time.Second}, clk)
	assert.True(t, g.IsFinished())
	assert.Equal(t, 4, g.Inner().Len())
	assert.True(t, g.IsFinished())
	assert.Equal(t, "rnb1kbnr/pppp1ppp/8/4p3/6Pq/5P2/PPPPP2P/RNBQKBNR w KQkq - 1 3", g.CurBoard().FEN())

	now = now.Add(10 * time.Second)
	clk, ok = g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 55 * time.Second, Black: 53 * time.Second}, clk)

	require.EqualError(t, g.Push(mv), "game already finished")
}

func TestGameTimeForfeit(t *testing.T) {
	now, err := time.Parse(time.RFC3339, "2012-09-28T12:34:56Z")
	require.NoError(t, err)

	c, err := ControlFromString("3/60+1")
	require.NoError(t, err)

	g := NewGame(chess.NewGame(), maybe.Some(c), GameOptions{
		Now: func() time.Time { return now },
	})
	clk, ok := g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 60 * time.Second, Black: 60 * time.Second, WhiteTicking: true}, clk)

	mv, err := chess.MoveFromUCI("e2e4", g.CurBoard())
	require.NoError(t, err)
	require.NoError(t, g.Push(mv))

	mv, err = chess.MoveFromUCI("e7e5", g.CurBoard())
	now = now.Add(62 * time.Second)
	require.NoError(t, err)
	require.NoError(t, g.Push(mv))

	assert.Equal(t, 1, g.Inner().Len())
	assert.Equal(t, "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1", g.CurBoard().FEN())
	assert.Equal(t, chess.MustWinOutcome(chess.VerdictTimeForfeit, chess.ColorWhite), g.Outcome())

	clk, ok = g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 61 * time.Second, Black: -2 * time.Second}, clk)
}

func TestGameTimeForfeitOnMate(t *testing.T) {
	now, err := time.Parse(time.RFC3339, "2012-09-28T12:34:56Z")
	require.NoError(t, err)

	c, err := ControlFromString("3/60+1")
	require.NoError(t, err)

	g := NewGame(chess.NewGame(), maybe.Some(c), GameOptions{
		Now: func() time.Time { return now },
	})

	for _, s := range []string{"g2g4", "e7e5", "f2f3"} {
		mv, err := chess.MoveFromUCI(s, g.CurBoard())
		require.NoError(t, err)
		require.NoError(t, g.Push(mv))
	}

	mv, err := chess.MoveFromUCI("d8h4", g.CurBoard())
	require.NoError(t, err)
	now = now.Add(65 * time.Second)
	require.NoError(t, g.Push(mv))

	assert.Equal(t, 3, g.Inner().Len())
	assert.Equal(t, "rnbqkbnr/pppp1ppp/8/4p3/6P1/5P2/PPPPP2P/RNBQKBNR b KQkq - 0 2", g.CurBoard().FEN())
	assert.Equal(t, chess.MustWinOutcome(chess.VerdictTimeForfeit, chess.ColorWhite), g.Outcome())

	clk, ok := g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 62 * time.Second, Black: -4 * time.Second}, clk)
}

func TestGameStop(t *testing.T) {
	now, err := time.Parse(time.RFC3339, "2012-09-28T12:34:56Z")
	require.NoError(t, err)

	c, err := ControlFromString("3/60+1")
	require.NoError(t, err)

	g := NewGame(chess.NewGame(), maybe.Some(c), GameOptions{
		Now: func() time.Time { return now },
	})

	require.EqualError(t, g.Finish(chess.RunningOutcome()), "outcome must finish the game")
	assert.False(t, g.IsFinished())

	now = now.Add(10 * time.Second)
	mv, err := chess.MoveFromUCI("e2e4", g.CurBoard())
	require.NoError(t, err)
	require.NoError(t, g.Push(mv))

	now = now.Add(5 * time.Second)
	require.NoError(t, g.Finish(chess.MustWinOutcome(chess.VerdictResign, chess.ColorWhite)))

	assert.True(t, g.IsFinished())
	assert.Equal(t, chess.MustWinOutcome(chess.VerdictResign, chess.ColorWhite), g.Outcome())

	clk, ok := g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 51 * time.Second, Black: 55 * time.Second}, clk)

	now = now.Add(5 * time.Second)
	clk, ok = g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 51 * time.Second, Black: 55 * time.Second}, clk)

	require.EqualError(t, g.Finish(chess.MustDrawOutcome(chess.VerdictDrawAgreement)), "game already finished")
	assert.Equal(t, chess.MustWinOutcome(chess.VerdictResign, chess.ColorWhite), g.Outcome())
}

func TestGameRacyStop(t *testing.T) {
	now, err := time.Parse(time.RFC3339, "2012-09-28T12:34:56Z")
	require.NoError(t, err)

	c, err := ControlFromString("3/60+1")
	require.NoError(t, err)

	g := NewGame(chess.NewGame(), maybe.Some(c), GameOptions{
		Now: func() time.Time { return now },
	})

	now = now.Add(10 * time.Second)
	mv, err := chess.MoveFromUCI("e2e4", g.CurBoard())
	require.NoError(t, err)
	require.NoError(t, g.Push(mv))

	now = now.Add(64 * time.Second)
	require.NoError(t, g.Finish(chess.MustDrawOutcome(chess.VerdictDrawAgreement)))
	assert.Equal(t, chess.MustWinOutcome(chess.VerdictTimeForfeit, chess.ColorWhite), g.Outcome())

	clk, ok := g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 51 * time.Second, Black: -4 * time.Second}, clk)

	now = now.Add(5 * time.Second)
	clk, ok = g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 51 * time.Second, Black: -4 * time.Second}, clk)
}

func TestGameUpdate(t *testing.T) {
	now, err := time.Parse(time.RFC3339, "2012-09-28T12:34:56Z")
	require.NoError(t, err)

	c, err := ControlFromString("3/60+1")
	require.NoError(t, err)

	g := NewGame(chess.NewGame(), maybe.Some(c), GameOptions{
		Now: func() time.Time { return now },
	})

	now = now.Add(10 * time.Second)
	g.UpdateTimer()
	assert.Equal(t, chess.RunningOutcome(), g.Outcome())

	clk, ok := g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 50 * time.Second, Black: 60 * time.Second, WhiteTicking: true}, clk)

	now = now.Add(50 * time.Second)
	g.UpdateTimer()
	assert.Equal(t, chess.MustWinOutcome(chess.VerdictTimeForfeit, chess.ColorBlack), g.Outcome())

	clk, ok = g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 0, Black: 60 * time.Second}, clk)

	now = now.Add(2 * time.Second)
	g.UpdateTimer()
	assert.Equal(t, chess.MustWinOutcome(chess.VerdictTimeForfeit, chess.ColorBlack), g.Outcome())

	clk, ok = g.Clock()
	assert.True(t, ok)
	assert.Equal(t, Clock{White: 0, Black: 60 * time.Second}, clk)
}

func TestInfiniteGame(t *testing.T) {
	g := NewGame(chess.NewGame(), maybe.None[Control](), GameOptions{})

	_, ok := g.UCITimeSpec()
	assert.False(t, ok)
	_, ok = g.Clock()
	assert.False(t, ok)
	_, ok = g.Deadline()
	assert.False(t, ok)

	mv, err := chess.MoveFromUCI("e2e4", g.CurBoard())
	require.NoError(t, err)
	require.NoError(t, g.Push(mv))

	g.UpdateTimer()
	assert.False(t, g.IsFinished())

	require.NoError(t, g.Finish(chess.MustDrawOutcome(chess.VerdictDrawAgreement)))
	assert.Equal(t, chess.MustDrawOutcome(chess.VerdictDrawAgreement), g.Outcome())
}

func TestInfiniteMate(t *testing.T) {
	g := NewGame(chess.NewGame(), maybe.None[Control](), GameOptions{})

	for _, s := range []string{"g2g4", "e7e5", "f2f3", "d8h4"} {
		mv, err := chess.MoveFromUCI(s, g.CurBoard())
		require.NoError(t, err)
		require.NoError(t, g.Push(mv))
	}

	assert.Equal(t, chess.MustWinOutcome(chess.VerdictCheckmate, chess.ColorBlack), g.Outcome())
}
