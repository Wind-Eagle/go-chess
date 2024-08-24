package chess

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGameSimple(t *testing.T) {
	const fen = "rnbqk1nr/ppp1bppp/3p4/4p3/2B1P3/5N2/PPPP1PPP/RNBQ1RK1 b kq - 3 4"
	g, err := NewGameWithFEN(fen)
	require.NoError(t, err)
	assert.Equal(t, fen, g.CurPos().FEN())
	assert.Equal(t, fen, g.StartPos().FEN())
	assert.Equal(t, 0, g.Len())

	require.NoError(t, g.PushMoveUCI("g8f6"))
	assert.Equal(t, "rnbqk2r/ppp1bppp/3p1n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQ1RK1 w kq - 4 5", g.CurPos().FEN())
	assert.Equal(t, fen, g.StartPos().FEN())
	assert.Equal(t, 1, g.Len())

	require.NoError(t, g.PushMoveSAN("d4"))
	assert.Equal(t, "rnbqk2r/ppp1bppp/3p1n2/4p3/2BPP3/5N2/PPP2PPP/RNBQ1RK1 b kq d3 0 5", g.CurPos().FEN())
	assert.Equal(t, fen, g.StartPos().FEN())
	assert.Equal(t, 2, g.Len())

	m, ok := g.Pop()
	require.True(t, ok)
	require.Equal(t, "d2d4", m.UCI())
	assert.Equal(t, "rnbqk2r/ppp1bppp/3p1n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQ1RK1 w kq - 4 5", g.CurPos().FEN())
	assert.Equal(t, fen, g.StartPos().FEN())
	assert.Equal(t, 1, g.Len())

	require.NoError(t, g.PushMove(m))
	assert.Equal(t, "rnbqk2r/ppp1bppp/3p1n2/4p3/2BPP3/5N2/PPP2PPP/RNBQ1RK1 b kq d3 0 5", g.CurPos().FEN())
	assert.Equal(t, fen, g.StartPos().FEN())
	assert.Equal(t, 2, g.Len())

	assert.Equal(t, "g8f6", g.MoveAt(0).UCI())
	assert.Equal(t, "d2d4", g.MoveAt(1).UCI())
	assert.Equal(t, "g8f6 d2d4", g.UCIList())
}

func TestGameRepeat(t *testing.T) {
	g := NewGame()
	assert.Equal(t, RunningOutcome(), g.Outcome())

	_, err := g.PushUCIList("g1f3 b8c6 f3g1 c6b8 g1f3 b8c6 f3g1 c6b8")
	require.NoError(t, err)
	assert.Equal(t, RunningOutcome(), g.Outcome())
	assert.Equal(t, MustDrawOutcome(VerdictRepeat3), g.CalcOutcome())

	g.SetAutoOutcome(VerdictFilterStrict)
	assert.Equal(t, RunningOutcome(), g.Outcome())

	_, err = g.PushUCIList("g1f3 b8c6 f3g1 c6b8 g1f3 b8c6 f3g1 c6b8")
	require.NoError(t, err)
	assert.Equal(t, RunningOutcome(), g.Outcome())
	assert.Equal(t, MustDrawOutcome(VerdictRepeat5), g.CalcOutcome())

	g.SetAutoOutcome(VerdictFilterStrict)
	assert.True(t, g.IsFinished())
	assert.Equal(t, MustDrawOutcome(VerdictRepeat5), g.Outcome())

	_, ok := g.Pop()
	assert.True(t, ok)
	assert.False(t, g.IsFinished())
	assert.Equal(t, RunningOutcome(), g.Outcome())

	g.SetAutoOutcome(VerdictFilterRelaxed)
	assert.True(t, g.IsFinished())
	assert.Equal(t, MustDrawOutcome(VerdictRepeat3), g.Outcome())

	assert.Equal(t, "g1f3 b8c6 f3g1 c6b8 g1f3 b8c6 f3g1 c6b8 g1f3 b8c6 f3g1 c6b8 g1f3 b8c6 f3g1", g.UCIList())
}

func TestGameCheckmate(t *testing.T) {
	g := NewGame()
	_, err := g.PushUCIList("g2g4 e7e5 f2f4 d8h4")
	require.NoError(t, err)

	outcome := MustWinOutcome(VerdictCheckmate, ColorBlack)
	assert.Equal(t, outcome, g.SetAutoOutcome(VerdictFilterForce))
	assert.True(t, g.IsFinished())
	assert.Equal(t, outcome, g.Outcome())

	s, err := g.Styled(GameStyle{
		Move:       MoveStyleSAN,
		MoveNumber: MoveNumberStyle{Enabled: true},
		Outcome:    GameOutcomeShow,
	})
	require.NoError(t, err)
	assert.Equal(t, "1. g4 e5 2. f4 Qh4# 0-1", s)
}

func TestGamePushUCIList(t *testing.T) {
	g := NewGame()

	n, err := g.PushUCIList("e2e4 e7e5 g1f3")
	require.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, 3, g.Len())

	n, err = g.PushUCIList("d7d6 f1c4 c8g4 b1c3 h7h6 f3e5 g4d1 c4f7 d1h5 f7h5")
	assert.Equal(t, 8, n)
	assert.Equal(t, 11, g.Len())
	assert.EqualError(t, err, "push uci move #9: make move: move is not legal")
	assert.Equal(t, "rn1qkbnr/ppp2Bp1/3p3p/4N3/4P3/2N5/PPPP1PPP/R1BbK2R b KQkq - 0 6", g.CurPos().FEN())

	n, err = g.PushUCIList("d8h7")
	assert.Equal(t, 0, n)
	assert.Equal(t, 11, g.Len())
	assert.EqualError(t, err, "push uci move #1: parse move: convert uci: move is not well-formed")
}

func TestGameOutcome(t *testing.T) {
	g, err := NewGameWithFEN("8/1Q2K3/8/3qk3/8/8/8/8 w - - 0 1")
	require.NoError(t, err)

	n, err := g.PushUCIList("b7d5 e5d5")
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, MustDrawOutcome(VerdictInsufficientMaterial), g.CalcOutcome())

	g.SetAutoOutcome(VerdictFilterForce)
	assert.Equal(t, RunningOutcome(), g.Outcome())

	g.SetAutoOutcome(VerdictFilterStrict)
	assert.Equal(t, MustDrawOutcome(VerdictInsufficientMaterial), g.Outcome())
}

func TestGameOutcome2(t *testing.T) {
	g, err := NewGameWithFEN("8/R7/2r5/8/5k1K/8/8/8 w - - 98 1")
	require.NoError(t, err)
	_, err = g.PushUCIList("a7a8 c6c5 a8a7 c5c6")
	require.NoError(t, err)
	assert.Equal(t, "8/R7/2r5/8/5k1K/8/8/8 w - - 102 3", g.CurPos().FEN())
	assert.Equal(t, MustDrawOutcome(VerdictMoves50), g.CalcOutcome())
}

func TestGameOutcomePriority(t *testing.T) {
	g, err := NewGameWithFEN("8/R7/2r5/8/5k1K/8/8/8 w - - 90 1")
	require.NoError(t, err)
	_, err = g.PushUCIList("a7a8 c6c5 a8a7 c5c6 a7a8 c6c5 a8a7 c5c6 a7a8 c6c5 a8a7 c5c6 a7a8 c6c5 a8a7 c5c6")
	require.NoError(t, err)
	assert.Equal(t, "8/R7/2r5/8/5k1K/8/8/8 w - - 106 9", g.CurPos().FEN())
	assert.Equal(t, MustDrawOutcome(VerdictRepeat5), g.CalcOutcome())
	g.SetAutoOutcome(VerdictFilterStrict)
	assert.True(t, g.IsFinished())

	g, err = NewGameWithFEN("8/R7/2r5/8/5k1K/8/8/8 w - - 142 1")
	require.NoError(t, err)
	_, err = g.PushUCIList("a7a8 c6c5 a8a7 c5c6 a7a8 c6c5 a8a7 c5c6")
	require.NoError(t, err)
	assert.Equal(t, "8/R7/2r5/8/5k1K/8/8/8 w - - 150 5", g.CurPos().FEN())
	assert.Equal(t, MustDrawOutcome(VerdictMoves75), g.CalcOutcome())
	g.SetAutoOutcome(VerdictFilterStrict)
	assert.True(t, g.IsFinished())
}

func TestWalker(t *testing.T) {
	g := NewGame()
	for _, m := range []string{
		"e4", "e5", "Nf3", "d6", "Bc4", "Bg4", "Nc3", "g6", "Nxe5", "Bxd1", "Bxf7", "Ke7", "Nd5#",
	} {
		require.NoError(t, g.PushMoveSAN(m))
	}

	w := g.Walk()
	assert.Equal(t, 13, w.Len())
	assert.Equal(t, 13, w.Pos())
	assert.False(t, w.Next())
	assert.Equal(t, 13, w.Pos())

	w.First()
	assert.Equal(t, 0, w.Pos())
	assert.False(t, w.Prev())
	assert.Equal(t, 0, w.Pos())

	w.Last()
	assert.Equal(t, 13, w.Pos())
	assert.False(t, w.Next())
	assert.Equal(t, 13, w.Pos())

	assert.True(t, w.Prev())
	assert.True(t, w.Prev())
	assert.True(t, w.Prev())
	assert.Equal(t, "rn1qkbnr/ppp2p1p/3p2p1/4N3/2B1P3/2N5/PPPP1PPP/R1BbK2R w KQkq - 0 6", w.Board().FEN())
	assert.Equal(t, 10, w.Pos())

	assert.True(t, w.Next())
	assert.Equal(t, "rn1qkbnr/ppp2B1p/3p2p1/4N3/4P3/2N5/PPPP1PPP/R1BbK2R b KQkq - 0 6", w.Board().FEN())
	assert.Equal(t, 11, w.Pos())

	assert.True(t, w.Next())
	assert.Equal(t, "rn1q1bnr/ppp1kB1p/3p2p1/4N3/4P3/2N5/PPPP1PPP/R1BbK2R w KQ - 1 7", w.Board().FEN())
	assert.Equal(t, 12, w.Pos())

	assert.True(t, w.Jump(3))
	assert.Equal(t, "rnbqkbnr/pppp1ppp/8/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2", w.Board().FEN())
	assert.Equal(t, 3, w.Pos())

	assert.True(t, w.Jump(11))
	assert.Equal(t, "rn1qkbnr/ppp2B1p/3p2p1/4N3/4P3/2N5/PPPP1PPP/R1BbK2R b KQkq - 0 6", w.Board().FEN())
	assert.Equal(t, 11, w.Pos())
}

func TestGameStyled(t *testing.T) {
	for _, v := range []struct {
		fen   string
		src   string
		out   Outcome
		style GameStyle
		res   string
	}{
		{
			src: "e2e4 e7e5 g1f3 d7d6 f1b5",
			style: GameStyle{
				Move:       MoveStyleSAN,
				MoveNumber: MoveNumberStyle{Enabled: false},
				Outcome:    GameOutcomeHide,
			},
			res: "e4 e5 Nf3 d6 Bb5+",
		},
		{
			src: "e2e4 e7e5 g1f3 d7d6 f1b5",
			style: GameStyle{
				Move:       MoveStyleSAN,
				MoveNumber: MoveNumberStyle{Enabled: true},
				Outcome:    GameOutcomeHide,
			},
			res: "1. e4 e5 2. Nf3 d6 3. Bb5+",
		},
		{
			src: "e2e4 e7e5 g1f3 d7d6 f1b5",
			style: GameStyle{
				Move: MoveStyleSAN,
				MoveNumber: MoveNumberStyle{
					Enabled:         true,
					Custom:          true,
					CustomStartFrom: 42,
				},
				Outcome: GameOutcomeHide,
			},
			res: "42. e4 e5 43. Nf3 d6 44. Bb5+",
		},
		{
			src: "e2e4 e7e5 g1f3 d7d6 f1b5",
			style: GameStyle{
				Move:       MoveStyleFancySAN,
				MoveNumber: MoveNumberStyle{Enabled: true},
				Outcome:    GameOutcomeHide,
			},
			res: "1. e4 e5 2. ♘f3 d6 3. ♗b5+",
		},
		{
			src: "e2e4 e7e5 g1f3 d7d6 f1b5",
			style: GameStyle{
				Move:       MoveStyleUCI,
				MoveNumber: MoveNumberStyle{Enabled: true},
				Outcome:    GameOutcomeHide,
			},
			res: "1. e2e4 e7e5 2. g1f3 d7d6 3. f1b5",
		},
		{
			src: "e2e4 e7e5 g1f3 d7d6 f1b5 c7c6",
			style: GameStyle{
				Move:       MoveStyleSAN,
				MoveNumber: MoveNumberStyle{Enabled: false},
				Outcome:    GameOutcomeHide,
			},
			res: "e4 e5 Nf3 d6 Bb5+ c6",
		},
		{
			src: "e2e4 e7e5 g1f3 d7d6 f1b5 c7c6",
			style: GameStyle{
				Move:       MoveStyleSAN,
				MoveNumber: MoveNumberStyle{Enabled: true},
				Outcome:    GameOutcomeHide,
			},
			res: "1. e4 e5 2. Nf3 d6 3. Bb5+ c6",
		},
		{
			src: "e2e4 e7e5 g1f3 d7d6 f1b5 c7c6",
			style: GameStyle{
				Move: MoveStyleSAN,
				MoveNumber: MoveNumberStyle{
					Enabled:         true,
					Custom:          true,
					CustomStartFrom: 42,
				},
				Outcome: GameOutcomeHide,
			},
			res: "42. e4 e5 43. Nf3 d6 44. Bb5+ c6",
		},
		{
			fen: "5K2/4P3/8/8/8/8/6p1/7k b - - 0 12",
			src: "g2g1q e7e8q",
			style: GameStyle{
				Move:       MoveStyleSAN,
				MoveNumber: MoveNumberStyle{Enabled: false},
				Outcome:    GameOutcomeHide,
			},
			res: "g1=Q e8=Q",
		},
		{
			fen: "5K2/4P3/8/8/8/8/6p1/7k b - - 0 12",
			src: "g2g1q e7e8q",
			style: GameStyle{
				Move:       MoveStyleSAN,
				MoveNumber: MoveNumberStyle{Enabled: true},
				Outcome:    GameOutcomeHide,
			},
			res: "12... g1=Q 13. e8=Q",
		},
		{
			fen: "5K2/4P3/8/8/8/8/6p1/7k b - - 0 12",
			src: "g2g1q e7e8q",
			style: GameStyle{
				Move: MoveStyleSAN,
				MoveNumber: MoveNumberStyle{
					Enabled:         true,
					Custom:          true,
					CustomStartFrom: 42,
				},
				Outcome: GameOutcomeHide,
			},
			res: "42... g1=Q 43. e8=Q",
		},
		{
			fen: "5K2/4P3/8/8/8/8/6p1/7k b - - 0 12",
			src: "g2g1q e7e8q g1c5",
			style: GameStyle{
				Move:       MoveStyleSAN,
				MoveNumber: MoveNumberStyle{Enabled: false},
				Outcome:    GameOutcomeHide,
			},
			res: "g1=Q e8=Q Qc5+",
		},
		{
			fen: "5K2/4P3/8/8/8/8/6p1/7k b - - 0 12",
			src: "g2g1q e7e8q g1c5",
			style: GameStyle{
				Move:       MoveStyleSAN,
				MoveNumber: MoveNumberStyle{Enabled: true},
				Outcome:    GameOutcomeHide,
			},
			res: "12... g1=Q 13. e8=Q Qc5+",
		},
		{
			fen: "5K2/4P3/8/8/8/8/6p1/7k b - - 0 12",
			src: "g2g1q e7e8q g1c5",
			style: GameStyle{
				Move: MoveStyleSAN,
				MoveNumber: MoveNumberStyle{
					Enabled:         true,
					Custom:          true,
					CustomStartFrom: 42,
				},
				Outcome: GameOutcomeHide,
			},
			res: "42... g1=Q 43. e8=Q Qc5+",
		},
		{
			src: "e2e4 e7e5 g1f3 d7d6 f1b5",
			out: MustDrawOutcome(VerdictDrawAgreement),
			style: GameStyle{
				Move:       MoveStyleSAN,
				MoveNumber: MoveNumberStyle{Enabled: true},
				Outcome:    GameOutcomeShow,
			},
			res: "1. e4 e5 2. Nf3 d6 3. Bb5+ 1/2-1/2",
		},
		{
			src: "e2e4 e7e5 g1f3 d7d6 f1b5",
			out: MustDrawOutcome(VerdictDrawAgreement),
			style: GameStyle{
				Move:       MoveStyleSAN,
				MoveNumber: MoveNumberStyle{Enabled: true},
				Outcome:    GameOutcomeFinishedOnly,
			},
			res: "1. e4 e5 2. Nf3 d6 3. Bb5+ 1/2-1/2",
		},
		{
			src: "e2e4 e7e5 g1f3 d7d6 f1b5",
			out: MustDrawOutcome(VerdictDrawAgreement),
			style: GameStyle{
				Move:       MoveStyleSAN,
				MoveNumber: MoveNumberStyle{Enabled: true},
				Outcome:    GameOutcomeHide,
			},
			res: "1. e4 e5 2. Nf3 d6 3. Bb5+",
		},
		{
			src: "e2e4 e7e5 g1f3 d7d6 f1b5",
			style: GameStyle{
				Move:       MoveStyleSAN,
				MoveNumber: MoveNumberStyle{Enabled: true},
				Outcome:    GameOutcomeFinishedOnly,
			},
			res: "1. e4 e5 2. Nf3 d6 3. Bb5+",
		},
	} {
		b := InitialBoard()
		if v.fen != "" {
			var err error
			b, err = BoardFromFEN(v.fen)
			require.NoError(t, err)
		}

		g, err := GameFromUCIList(b, v.src)
		require.NoError(t, err)
		g.SetOutcome(v.out)

		s, err := g.Styled(v.style)
		require.NoError(t, err)
		assert.Equal(t, v.res, s)
	}
}

func TestGameMisc(t *testing.T) {
	for _, v := range []struct {
		fen  string
		src  string
		clen int
		lfen string
	}{
		{
			src:  "e2e4 e7e5 g1f3 d7d6 f1b5",
			clen: 5,
			lfen: "rnbqkbnr/ppp2ppp/3p4/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R b KQkq - 1 3",
		},
		{
			src:  "e2e4 e7e5 g1f3 d7d6 f1b5 c7c6",
			clen: 6,
			lfen: "rnbqkbnr/pp3ppp/2pp4/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 4",
		},
		{
			fen:  "5K2/4P3/8/8/8/8/6p1/7k b - - 0 12",
			src:  "g2g1q e7e8q",
			clen: 2,
			lfen: "4QK2/8/8/8/8/8/8/6qk b - - 0 13",
		},
		{
			fen:  "5K2/4P3/8/8/8/8/6p1/7k b - - 0 12",
			src:  "g2g1q e7e8q g1c5",
			clen: 3,
			lfen: "4QK2/8/8/2q5/8/8/8/7k w - - 1 14",
		},
	} {
		b := InitialBoard()
		if v.fen != "" {
			var err error
			b, err = BoardFromFEN(v.fen)
			require.NoError(t, err)
		}

		g, err := GameFromUCIList(b, v.src)
		require.NoError(t, err)
		assert.Equal(t, v.clen, g.Len())
		assert.Equal(t, v.lfen, g.CurPos().FEN())
	}
}

func TestGamePushExtra(t *testing.T) {
	g := NewGame()

	mv, err := LegalMoveFromSAN("e4", g.CurBoard())
	require.NoError(t, err)

	g.PushLegalMove(mv)

	err = g.PushUCIMove(SimpleUCIMove(
		CoordFromParts(FileE, Rank7),
		CoordFromParts(FileE, Rank5),
	))
	require.NoError(t, err)

	_, err = g.PushUCIList("f1c4 f8c5 c4f7")
	require.NoError(t, err)

	mv, err = SemilegalMoveFromUCI("d7d5", g.CurBoard())
	require.NoError(t, err)

	require.False(t, g.PushSemilegalMove(mv))
	assert.Equal(t, "rnbqk1nr/pppp1Bpp/8/2b1p3/4P3/8/PPPP1PPP/RNBQK1NR b KQkq - 0 3", g.CurPos().FEN())
	assert.Equal(t, 5, g.Len())

	mv, err = SemilegalMoveFromUCI("e8f7", g.CurBoard())
	require.NoError(t, err)

	require.True(t, g.PushSemilegalMove(mv))
	assert.Equal(t, "rnbq2nr/pppp1kpp/8/2b1p3/4P3/8/PPPP1PPP/RNBQK1NR w KQ - 0 4", g.CurPos().FEN())
	assert.Equal(t, 6, g.Len())
}
