package chess

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBoardInitial(t *testing.T) {
	const iniFen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

	assert.Equal(t, iniFen, InitialRawBoard().String())
	assert.Equal(t, iniFen, InitialBoard().String())

	raw, err := RawBoardFromFEN(iniFen)
	require.NoError(t, err)
	assert.Equal(t, InitialRawBoard(), raw)

	b, err := BoardFromFEN(iniFen)
	require.NoError(t, err)
	assert.Equal(t, InitialRawBoard(), b.Raw())
}

func TestBoardMidgame(t *testing.T) {
	const fen = "1rq1r1k1/1p3ppp/pB3n2/3ppP2/Pbb1P3/1PN2B2/2P2QPP/R1R4K w - - 1 21"

	b, err := BoardFromFEN(fen)
	require.NoError(t, err)

	assert.Equal(t, CoordFromParts(FileH, Rank1), b.KingPos(ColorWhite))
	assert.Equal(t, CoordFromParts(FileG, Rank8), b.KingPos(ColorBlack))

	r := b.Raw()
	assert.Equal(t, CellFromParts(ColorBlack, PieceBishop), r.Get2(FileB, Rank4))
	assert.Equal(t, CellFromParts(ColorWhite, PieceQueen), r.Get2(FileF, Rank2))
	assert.Equal(t, ColorWhite, r.Side)
	assert.Equal(t, CastlingRightsEmpty, r.Castling)
	assert.Equal(t, NoCoord, r.EpSource)
	assert.Equal(t, uint8(1), r.MoveCounter)
	assert.Equal(t, uint32(21), r.MoveNumber)

	assert.Equal(t, CellFromParts(ColorBlack, PieceBishop), b.Get2(FileB, Rank4))
	assert.Equal(t, CellFromParts(ColorWhite, PieceQueen), b.Get2(FileF, Rank2))
	assert.Equal(t, ColorWhite, b.Side())
	assert.Equal(t, CastlingRightsEmpty, b.Castling())
	assert.Equal(t, NoCoord, b.EpSource())
	assert.Equal(t, uint8(1), b.MoveCounter())
	assert.Equal(t, uint32(21), b.MoveNumber())
}

func TestBoardFixes(t *testing.T) {
	const fen = "r1bq1b1r/ppppkppp/2n2n2/4p3/2B1P3/5N2/PPPP1PPP/RNBQK1R1 w KQkq c6 6 5"

	r, err := RawBoardFromFEN(fen)
	require.NoError(t, err)

	assert.Equal(t, CastlingRightsFull, r.Castling)
	assert.Equal(t, SomeCoord(CoordFromParts(FileC, Rank5)), r.EpSource)
	assert.Equal(t, SomeCoord(CoordFromParts(FileC, Rank6)), r.EpDest())
	assert.Equal(t, fen, r.FEN())

	b, err := NewBoard(r)
	require.NoError(t, err)

	assert.Equal(t, CastlingRightsEmpty.With(ColorWhite, CastlingQueenside), b.Castling())
	assert.Equal(t, NoCoord, b.EpSource())
	assert.Equal(t, NoCoord, b.EpDest())
}

func TestBoardIncomplete(t *testing.T) {
	_, err := RawBoardFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR")
	assert.EqualError(t, err, "no move side")

	_, err = RawBoardFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w")
	assert.EqualError(t, err, "no castling")

	_, err = RawBoardFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq")
	assert.EqualError(t, err, "no enpassant")

	r, err := RawBoardFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -")
	require.NoError(t, err)
	assert.Equal(t, uint8(0), r.MoveCounter)
	assert.Equal(t, uint32(1), r.MoveNumber)

	r, err = RawBoardFromFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 10")
	require.NoError(t, err)
	assert.Equal(t, uint8(10), r.MoveCounter)
	assert.Equal(t, uint32(1), r.MoveNumber)
}

func TestBoardMakeMoveExtra(t *testing.T) {
	const fen = "rn1qkbnr/ppp2Bp1/3p3p/4N3/4P3/2N5/PPPP1PPP/R1BbK2R b KQkq - 0 6"
	b, err := BoardFromFEN(fen)
	require.NoError(t, err)

	mv, err := SemilegalMoveFromUCI("e8f7", b)
	require.NoError(t, err)

	_, ok := b.MakeSemilegalMove(mv)
	assert.False(t, ok)
	assert.Equal(t, fen, b.FEN())

	mv, err = SemilegalMoveFromUCI("e8e7", b)
	require.NoError(t, err)

	u, ok := b.MakeSemilegalMove(mv)
	assert.True(t, ok)
	assert.Equal(t, "rn1q1bnr/ppp1kBp1/3p3p/4N3/4P3/2N5/PPPP1PPP/R1BbK2R w KQ - 1 7", b.FEN())
	b.UnmakeMove(u)

	_, err = b.MakeUCIMove(UCIMove{
		Kind: UCIMoveSimple,
		Src:  CoordFromParts(FileE, Rank8),
		Dst:  CoordFromParts(FileE, Rank7),
	})
	require.NoError(t, err)
	assert.Equal(t, "rn1q1bnr/ppp1kBp1/3p3p/4N3/4P3/2N5/PPPP1PPP/R1BbK2R w KQ - 1 7", b.FEN())
}

func TestBoardIsCheck(t *testing.T) {
	b, err := BoardFromFEN("8/1r3K2/8/4nk2/8/8/8/8 w - - 1 2")
	require.NoError(t, err)
	assert.True(t, b.IsCheck())
	assert.Equal(t, BbEmpty.With2(FileB, Rank7).With2(FileE, Rank5), b.Checkers())
}

func TestBoardPretty(t *testing.T) {
	b, err := BoardFromFEN("rnbqkbnr/ppp2ppp/3p4/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 0 3")
	require.NoError(t, err)
	assert.Equal(
		t,
		strings.Join(
			[]string{
				"8|rnbqkbnr",
				"7|ppp..ppp",
				"6|...p....",
				"5|....p...",
				"4|....P...",
				"3|.....N..",
				"2|PPPP.PPP",
				"1|RNBQKB.R",
				"-+--------",
				"W|abcdefgh",
				"",
			},
			"\n",
		),
		b.Pretty(PrettyStyleASCII),
	)
	assert.Equal(
		t,
		strings.Join(
			[]string{
				"8│♜♞♝♛♚♝♞♜",
				"7│♟♟♟..♟♟♟",
				"6│...♟....",
				"5│....♟...",
				"4│....♙...",
				"3│.....♘..",
				"2│♙♙♙♙.♙♙♙",
				"1│♖♘♗♕♔♗.♖",
				"─┼────────",
				"○│abcdefgh",
				"",
			},
			"\n",
		),
		b.Pretty(PrettyStyleFancy),
	)

	b, err = BoardFromFEN("rnbqkbnr/ppp2ppp/3p4/4p3/2B1P3/5N2/PPPP1PPP/RNBQK2R b KQkq - 1 3")
	require.NoError(t, err)
	assert.Equal(
		t,
		strings.Join(
			[]string{
				"8|rnbqkbnr",
				"7|ppp..ppp",
				"6|...p....",
				"5|....p...",
				"4|..B.P...",
				"3|.....N..",
				"2|PPPP.PPP",
				"1|RNBQK..R",
				"-+--------",
				"B|abcdefgh",
				"",
			},
			"\n",
		),
		b.Pretty(PrettyStyleASCII),
	)
	assert.Equal(
		t,
		strings.Join(
			[]string{
				"8│♜♞♝♛♚♝♞♜",
				"7│♟♟♟..♟♟♟",
				"6│...♟....",
				"5│....♟...",
				"4│..♗.♙...",
				"3│.....♘..",
				"2│♙♙♙♙.♙♙♙",
				"1│♖♘♗♕♔..♖",
				"─┼────────",
				"●│abcdefgh",
				"",
			},
			"\n",
		),
		b.Pretty(PrettyStyleFancy),
	)
}

func TestBoardOutcome(t *testing.T) {
	for _, v := range []struct {
		fen string
		out Outcome
	}{
		{
			fen: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			out: Outcome{Verdict: VerdictRunning},
		},
		{
			fen: "rn1qkbnr/ppp2B1p/3p2p1/4N3/4P3/2N5/PPPP1PPP/R1BbK2R b KQkq - 0 6",
			out: Outcome{Verdict: VerdictRunning},
		},
		{
			fen: "rn1q1bnr/ppp1kB1p/3p2p1/3NN3/4P3/8/PPPP1PPP/R1BbK2R b KQ - 2 7",
			out: Outcome{Verdict: VerdictCheckmate, Side: ColorWhite},
		},
		{
			fen: "7K/8/5n2/5n2/8/8/7k/8 w - - 0 1",
			out: Outcome{Verdict: VerdictStalemate},
		},
		{
			fen: "7K/8/5n2/8/8/8/7k/8 w - - 0 1",
			out: Outcome{Verdict: VerdictInsufficientMaterial},
		},
		{
			fen: "7K/8/5b2/8/8/8/7k/8 w - - 0 1",
			out: Outcome{Verdict: VerdictInsufficientMaterial},
		},
		{
			fen: "2K4k/8/8/8/B1B5/1B1B4/B1B5/1B1B4 w - - 0 1",
			out: Outcome{Verdict: VerdictInsufficientMaterial},
		},
		{
			fen: "8/3k4/8/8/8/8/1K6/8 w - - 0 1",
			out: Outcome{Verdict: VerdictInsufficientMaterial},
		},
		{
			fen: "8/3k4/8/6B1/5B1B/4B1B1/1K3B1B/4B1B1 w - - 0 1",
			out: Outcome{Verdict: VerdictInsufficientMaterial},
		},
		{
			fen: "BBK4k/8/8/8/8/8/8/8 w - - 0 1",
			out: Outcome{Verdict: VerdictRunning},
		},
		{
			fen: "NNK4k/8/8/8/8/8/8/8 w - - 0 1",
			out: Outcome{Verdict: VerdictRunning},
		},
		{
			fen: "NNK4k/8/8/8/8/8/8/8 w - - 99 80",
			out: Outcome{Verdict: VerdictRunning},
		},
		{
			fen: "NNK4k/8/8/8/8/8/8/8 w - - 100 80",
			out: Outcome{Verdict: VerdictMoves50},
		},
		{
			fen: "NNK4k/8/8/8/8/8/8/8 w - - 149 90",
			out: Outcome{Verdict: VerdictMoves50},
		},
		{
			fen: "NNK4k/8/8/8/8/8/8/8 w - - 150 90",
			out: Outcome{Verdict: VerdictMoves75},
		},
	} {
		b, err := BoardFromFEN(v.fen)
		require.NoError(t, err)
		hasLegalMoves := v.out.Verdict != VerdictCheckmate && v.out.Verdict != VerdictStalemate
		assert.Equal(t, hasLegalMoves, b.HasLegalMoves())
		assert.Equal(t, v.out, b.CalcOutcome())
	}
}
