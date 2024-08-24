//go:build largetest

package chess_test

// Testing positions for Perft are taken from https://www.chessprogramming.org/Perft_Results.

import (
	"testing"

	"github.com/alex65536/go-chess/chess"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runPerft(b *chess.Board, d int, moves []int, captures []int) {
	var buf [256]chess.Move
	ms := b.GenLegalMoves(chess.MoveGenCapture, buf[:0])
	captures[d] += len(ms)
	ms = b.GenLegalMoves(chess.MoveGenSimple, ms)
	moves[d] += len(ms)
	if d == len(moves)-1 {
		return
	}
	for _, m := range ms {
		u := b.MakeLegalMove(m)
		runPerft(b, d+1, moves, captures)
		b.UnmakeMove(u)
	}
}

func doTestPerft(t *testing.T, fen string, moves, captures []int) {
	require.Equal(t, len(moves), len(captures))

	b, err := chess.BoardFromFEN(fen)
	require.NoError(t, err)

	resMoves := make([]int, len(moves))
	resCaptures := make([]int, len(captures))

	runPerft(b, 0, resMoves, resCaptures)

	assert.Equal(t, moves, resMoves)
	assert.Equal(t, captures, resCaptures)
}

func TestPerftInitial(t *testing.T) {
	t.Parallel()
	doTestPerft(
		t,
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		[]int{20, 400, 8902, 197281, 4865609, 119060324},
		[]int{0, 0, 34, 1576, 82719, 2812008},
	)
}

func TestPerftKiwipete(t *testing.T) {
	t.Parallel()
	doTestPerft(
		t,
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq -",
		[]int{48, 2039, 97862, 4085603, 193690690},
		[]int{8, 351, 17102, 757163, 35043416},
	)
}

func TestPerftPosition3(t *testing.T) {
	t.Parallel()
	doTestPerft(
		t,
		"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - -",
		[]int{14, 191, 2812, 43238, 674624, 11030083, 178633661},
		[]int{1, 14, 209, 3348, 52051, 940350, 14519036},
	)
}

func TestPerftPosition4(t *testing.T) {
	t.Parallel()
	doTestPerft(
		t,
		"r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
		[]int{6, 264, 9467, 422333, 15833292},
		[]int{0, 87, 1021, 131393, 2046173},
	)
}

func TestPerftPosition5(t *testing.T) {
	t.Parallel()
	doTestPerft(
		t,
		"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
		[]int{44, 1486, 62379, 2103487, 89941194},
		[]int{6, 222, 8517, 296153, 12320378},
	)
}

func TestPerftPosition6(t *testing.T) {
	t.Parallel()
	doTestPerft(
		t,
		"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
		[]int{46, 2079, 89890, 3894594, 164075551},
		[]int{4, 203, 9470, 440388, 19528068},
	)
}
