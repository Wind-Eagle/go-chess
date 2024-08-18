package chess

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMoveStyle(t *testing.T) {
	b := InitialBoard()

	mv, err := MoveFromUCI("g1f3", b)
	require.NoError(t, err)

	s, err := mv.Styled(b, MoveStyleUCI)
	require.NoError(t, err)
	assert.Equal(t, "g1f3", s)

	s, err = mv.Styled(b, MoveStyleSAN)
	require.NoError(t, err)
	assert.Equal(t, "Nf3", s)

	s, err = mv.Styled(b, MoveStyleFancySAN)
	require.NoError(t, err)
	assert.Equal(t, "♘f3", s)

	s = mv.UCI()
	require.NoError(t, err)
	assert.Equal(t, "g1f3", s)
}

func TestMoveSimple(t *testing.T) {
	b := InitialBoard()
	for _, v := range []struct {
		mv   string
		fen  string
		kind MoveKind
	}{
		{
			mv:   "e2e4",
			fen:  "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
			kind: MovePawnDouble,
		},
		{
			mv:   "b8c6",
			fen:  "r1bqkbnr/pppppppp/2n5/8/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 1 2",
			kind: MoveSimple,
		},
		{
			mv:   "g1f3",
			fen:  "r1bqkbnr/pppppppp/2n5/8/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 2 2",
			kind: MoveSimple,
		},
		{
			mv:   "e7e5",
			fen:  "r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq e6 0 3",
			kind: MovePawnDouble,
		},
		{
			mv:   "f1b5",
			fen:  "r1bqkbnr/pppp1ppp/2n5/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R b KQkq - 1 3",
			kind: MoveSimple,
		},
		{
			mv:   "g8f6",
			fen:  "r1bqkb1r/pppp1ppp/2n2n2/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 2 4",
			kind: MoveSimple,
		},
		{
			mv:   "e1g1",
			fen:  "r1bqkb1r/pppp1ppp/2n2n2/1B2p3/4P3/5N2/PPPP1PPP/RNBQ1RK1 b kq - 3 4",
			kind: MoveCastlingKingside,
		},
		{
			mv:   "f6e4",
			fen:  "r1bqkb1r/pppp1ppp/2n5/1B2p3/4n3/5N2/PPPP1PPP/RNBQ1RK1 w kq - 0 5",
			kind: MoveSimple,
		},
	} {
		m, err := LegalMoveFromUCI(v.mv, b)
		require.NoError(t, err)
		_, err = b.MakeMove(m)
		require.NoError(t, err)
		assert.Equal(t, v.fen, b.FEN())
		nb, err := NewBoard(b.Raw())
		require.NoError(t, err)
		require.Equal(t, b.Raw(), nb.Raw())
	}
}

func TestMoveWithUndo(t *testing.T) {
	for _, v := range []struct {
		mv   string
		ofen string
		nfen string
	}{
		{
			mv:   "c7c8q",
			ofen: "1b1b1K2/2P5/8/8/7k/8/8/8 w - - 0 1",
			nfen: "1bQb1K2/8/8/8/7k/8/8/8 b - - 0 1",
		},
		{
			mv:   "c7b8n",
			ofen: "1b1b1K2/2P5/8/8/7k/8/8/8 w - - 0 1",
			nfen: "1N1b1K2/8/8/8/7k/8/8/8 b - - 0 1",
		},
		{
			mv:   "c7d8r",
			ofen: "1b1b1K2/2P5/8/8/7k/8/8/8 w - - 0 1",
			nfen: "1b1R1K2/8/8/8/7k/8/8/8 b - - 0 1",
		},
		{
			mv:   "c7d8n",
			ofen: "1b1b1K2/2P5/8/8/7k/8/8/8 w - - 0 1",
			nfen: "1b1N1K2/8/8/8/7k/8/8/8 b - - 0 1",
		},
		{
			mv:   "g2g3",
			ofen: "3K4/3p4/8/3PpP2/8/5p2/6P1/2k5 w - e6 0 1",
			nfen: "3K4/3p4/8/3PpP2/8/5pP1/8/2k5 b - - 0 1",
		},
		{
			mv:   "g2g4",
			ofen: "3K4/3p4/8/3PpP2/8/5p2/6P1/2k5 w - e6 0 1",
			nfen: "3K4/3p4/8/3PpP2/6P1/5p2/8/2k5 b - g3 0 1",
		},
		{
			mv:   "g2f3",
			ofen: "3K4/3p4/8/3PpP2/8/5p2/6P1/2k5 w - e6 0 1",
			nfen: "3K4/3p4/8/3PpP2/8/5P2/8/2k5 b - - 0 1",
		},
		{
			mv:   "d5e6",
			ofen: "3K4/3p4/8/3PpP2/8/5p2/6P1/2k5 w - e6 0 1",
			nfen: "3K4/3p4/4P3/5P2/8/5p2/6P1/2k5 b - - 0 1",
		},
		{
			mv:   "f5e6",
			ofen: "3K4/3p4/8/3PpP2/8/5p2/6P1/2k5 w - e6 0 1",
			nfen: "3K4/3p4/4P3/3P4/8/5p2/6P1/2k5 b - - 0 1",
		},
		{
			mv:   "e1g1",
			ofen: "r1bqk2r/ppp2ppp/2np1n2/1Bb1p3/4P3/2PP1N2/PP3PPP/RNBQK2R w KQkq - 0 6",
			nfen: "r1bqk2r/ppp2ppp/2np1n2/1Bb1p3/4P3/2PP1N2/PP3PPP/RNBQ1RK1 b kq - 1 6",
		},
		{
			mv:   "f3e5",
			ofen: "r1bqk2r/ppp2ppp/2np1n2/1Bb1p3/4P3/2PP1N2/PP3PPP/RNBQK2R w KQkq - 0 6",
			nfen: "r1bqk2r/ppp2ppp/2np1n2/1Bb1N3/4P3/2PP4/PP3PPP/RNBQK2R b KQkq - 0 6",
		},
		{
			mv:   "b2b4",
			ofen: "r1bqk2r/ppp2ppp/2np1n2/1Bb1p3/4P3/2PP1N2/PP3PPP/RNBQK2R w KQkq - 0 6",
			nfen: "r1bqk2r/ppp2ppp/2np1n2/1Bb1p3/1P2P3/2PP1N2/P4PPP/RNBQK2R b KQkq b3 0 6",
		},
		{
			mv:   "c3c4",
			ofen: "r1bqk2r/ppp2ppp/2np1n2/1Bb1p3/4P3/2PP1N2/PP3PPP/RNBQK2R w KQkq - 0 6",
			nfen: "r1bqk2r/ppp2ppp/2np1n2/1Bb1p3/2P1P3/3P1N2/PP3PPP/RNBQK2R b KQkq - 0 6",
		},
		{
			mv:   "e8g8",
			ofen: "r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R b KQkq - 0 1",
			nfen: "r4rk1/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQ - 1 2",
		},
		{
			mv:   "e8c8",
			ofen: "r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R b KQkq - 0 1",
			nfen: "2kr3r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQ - 1 2",
		},
	} {
		b, err := BoardFromFEN(v.ofen)
		require.NoError(t, err)
		b2 := b.Clone()
		u, err := b.MakeMoveUCI(v.mv)
		require.NoError(t, err)
		assert.Equal(t, v.nfen, b.FEN())
		nb, err := NewBoard(b.Raw())
		require.NoError(t, err)
		assert.Equal(t, b.Raw(), nb.Raw())
		b.UnmakeMove(u)
		assert.Equal(t, *b2, *b)
	}
}

func TestMoveSemiValidate(t *testing.T) {
	b, err := BoardFromFEN("r1bqk2r/ppp2ppp/2np1n2/1Bb1p3/4P3/2PP1N2/PP3PPP/RNBQK2R w KQkq - 0 6")
	require.NoError(t, err)

	m, err := MoveFromUCI("e1c1", b)
	require.NoError(t, err)
	assert.EqualError(t, m.SemiValidate(b), "move is not semi-legal")

	m, err = MoveFromUCI("b5e8", b)
	require.NoError(t, err)
	assert.EqualError(t, m.SemiValidate(b), "move is not semi-legal")

	_, err = MoveFromUCI("a3a4", b)
	assert.EqualError(t, err, "convert uci: bad uci move src")

	_, err = MoveFromUCI("b5h5", b)
	assert.EqualError(t, err, "convert uci: move is not well-formed")

	m, err = MoveFromUCI("e1d1", b)
	require.NoError(t, err)
	assert.EqualError(t, m.SemiValidate(b), "move is not semi-legal")

	_, err = MoveFromUCI("c3c5", b)
	assert.EqualError(t, err, "convert uci: move is not well-formed")
}
