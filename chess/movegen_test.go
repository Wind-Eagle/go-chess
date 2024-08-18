package chess

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCellAttackers(t *testing.T) {
	b, err := BoardFromFEN("3R3B/8/3R4/1NP1Q3/3p4/1NP5/5B2/3R1K1k w - - 0 1")
	require.NoError(t, err)

	attackers := BbEmpty.
		With2(FileD, Rank6).
		With2(FileB, Rank5).
		With2(FileE, Rank5).
		With2(FileB, Rank3).
		With2(FileC, Rank3).
		With2(FileF, Rank2).
		With2(FileD, Rank1)

	assert.True(t, b.IsCellAttacked(CoordFromParts(FileD, Rank4), ColorWhite))
	assert.Equal(t, attackers, b.CellAttackers(CoordFromParts(FileD, Rank4), ColorWhite))

	assert.False(t, b.IsCellAttacked(CoordFromParts(FileD, Rank4), ColorBlack))
	assert.Equal(t, BbEmpty, b.CellAttackers(CoordFromParts(FileD, Rank4), ColorBlack))

	b, err = BoardFromFEN("8/8/8/2KPk3/8/8/8/8 w - - 0 1")
	require.NoError(t, err)

	assert.True(t, b.IsCellAttacked(CoordFromParts(FileD, Rank5), ColorWhite))
	assert.Equal(t, BbEmpty.With2(FileC, Rank5), b.CellAttackers(CoordFromParts(FileD, Rank5), ColorWhite))

	assert.True(t, b.IsCellAttacked(CoordFromParts(FileD, Rank5), ColorBlack))
	assert.Equal(t, BbEmpty.With2(FileE, Rank5), b.CellAttackers(CoordFromParts(FileD, Rank5), ColorBlack))
}

func movesToStr(ms []Move) []string {
	res := make([]string, len(ms))
	for i := range ms {
		res[i] = ms[i].String()
	}
	return res
}

func assertMovesLegal(t *testing.T, b *Board, ms []Move) {
	for _, m := range ms {
		err := m.Validate(b)
		assert.NoError(t, err)
	}
}

func TestSANCandidates(t *testing.T) {
	b, err := BoardFromFEN("3R3B/B7/1B1R4/1N2Q3/RQ1p4/1N6/5B2/3R1K1k w - - 0 1")
	require.NoError(t, err)

	for _, v := range []struct {
		p     Piece
		c     Coord
		moves []string
	}{
		{p: PieceKnight, c: CoordFromParts(FileD, Rank4), moves: []string{"b5d4", "b3d4"}},
		{p: PieceBishop, c: CoordFromParts(FileD, Rank4), moves: []string{"b6d4", "f2d4"}},
		{p: PieceRook, c: CoordFromParts(FileD, Rank4), moves: []string{"d1d4", "d6d4"}},
		{p: PieceQueen, c: CoordFromParts(FileD, Rank4), moves: []string{"b4d4", "e5d4"}},
		{p: PieceKing, c: CoordFromParts(FileD, Rank4), moves: []string{}},
		{p: PieceKing, c: CoordFromParts(FileE, Rank2), moves: []string{"f1e2"}},
	} {
		moves := b.sanCandidates(v.p, v.c, nil)
		assertMovesLegal(t, b, moves)
		assert.ElementsMatch(t, v.moves, movesToStr(moves))
	}
}

func TestSANPawnCaptureCandidates(t *testing.T) {
	b, err := BoardFromFEN("7n/6P1/8/2PpP2r/2P1P1P1/7q/6P1/3k1K2 w - d6 0 1")
	require.NoError(t, err)

	for _, v := range []struct {
		f1    File
		f2    File
		isP   bool
		p     Piece
		moves []string
	}{
		{f1: FileC, f2: FileD, moves: []string{"c4d5", "c5d6"}},
		{f1: FileE, f2: FileD, moves: []string{"e4d5", "e5d6"}},
		{f1: FileE, f2: FileD, isP: true, p: PieceKnight, moves: []string{}},
		{f1: FileG, f2: FileH, moves: []string{"g2h3", "g4h5"}},
		{f1: FileG, f2: FileH, isP: true, p: PieceBishop, moves: []string{"g7h8b"}},
	} {
		moves := b.sanPawnCaptureCandidates(v.f1, v.f2, v.isP, v.p, nil)
		assertMovesLegal(t, b, moves)
		assert.ElementsMatch(t, v.moves, movesToStr(moves))
	}
}

func TestHasLegalMoves(t *testing.T) {
	for _, v := range []struct {
		fen string
		has bool
	}{
		{fen: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", has: true},
		{fen: "rn1qkbnr/ppp2B1p/3p2p1/4N3/4P3/2N5/PPPP1PPP/R1BbK2R b KQkq - 0 6", has: true},
		{fen: "r1bqkb1r/pppp1Qpp/2n2n2/4p3/2B1P3/8/PPPP1PPP/RNB1K1NR b KQkq - 0 4", has: false},
		{fen: "K7/8/2n5/2n2p1p/5P1P/8/8/5k2 w - - 0 1", has: false},
	} {
		b, err := BoardFromFEN(v.fen)
		require.NoError(t, err)
		assert.Equal(t, v.has, b.HasLegalMoves())
	}
}

func TestMoveGenSimple(t *testing.T) {
	for _, v := range []struct {
		fen                                 string
		allSemi, simpleSemi, captureSemi    int
		allLegal, simpleLegal, captureLegal int
	}{
		{
			fen:          "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			allSemi:      20,
			simpleSemi:   20,
			captureSemi:  0,
			allLegal:     20,
			simpleLegal:  20,
			captureLegal: 0,
		},
		{
			fen:          "r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3",
			allSemi:      27,
			simpleSemi:   26,
			captureSemi:  1,
			allLegal:     27,
			simpleLegal:  26,
			captureLegal: 1,
		},
		{
			fen:          "6kb/R7/8/4P3/8/1p6/1K6/2r4r w - - 0 1",
			allSemi:      23,
			simpleSemi:   21,
			captureSemi:  2,
			allLegal:     16,
			simpleLegal:  15,
			captureLegal: 1,
		},
		{
			fen:          "5k2/8/8/8/8/8/8/4K2R w K - 0 1",
			allSemi:      15,
			simpleSemi:   15,
			captureSemi:  0,
			allLegal:     15,
			simpleLegal:  15,
			captureLegal: 0,
		},
		{
			fen:          "r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQkq - 0 1",
			allSemi:      25,
			simpleSemi:   25,
			captureSemi:  0,
			allLegal:     25,
			simpleLegal:  25,
			captureLegal: 0,
		},

		{
			fen:          "r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQkq - 0 1",
			allSemi:      25,
			simpleSemi:   25,
			captureSemi:  0,
			allLegal:     25,
			simpleLegal:  25,
			captureLegal: 0,
		},
		{
			fen:          "1b1n1r2/1P2P3/8/5PpP/8/8/1k1K4/8 w - g6 0 1",
			allSemi:      24,
			simpleSemi:   14,
			captureSemi:  10,
			allLegal:     21,
			simpleLegal:  11,
			captureLegal: 10,
		},
	} {
		b, err := BoardFromFEN(v.fen)
		require.NoError(t, err)

		assert.Len(t, b.GenSemilegalMoves(MoveGenAll, nil), v.allSemi)
		assert.Len(t, b.GenSemilegalMoves(MoveGenSimple, nil), v.simpleSemi)
		assert.Len(t, b.GenSemilegalMoves(MoveGenCapture, nil), v.captureSemi)

		assert.Len(t, b.GenLegalMoves(MoveGenAll, nil), v.allLegal)
		assert.Len(t, b.GenLegalMoves(MoveGenSimple, nil), v.simpleLegal)
		assert.Len(t, b.GenLegalMoves(MoveGenCapture, nil), v.captureLegal)
	}
}
