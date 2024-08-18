package chess

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSANSimpleGame(t *testing.T) {
	b := InitialBoard()
	for _, v := range []struct {
		mv  string
		fen string
	}{
		{
			mv:  "e4",
			fen: "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
		},
		{
			mv:  "Nc6",
			fen: "r1bqkbnr/pppppppp/2n5/8/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 1 2",
		},
		{
			mv:  "Nf3",
			fen: "r1bqkbnr/pppppppp/2n5/8/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 2 2",
		},
		{
			mv:  "e5",
			fen: "r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq e6 0 3",
		},
		{
			mv:  "Bb5",
			fen: "r1bqkbnr/pppp1ppp/2n5/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R b KQkq - 1 3",
		},
		{
			mv:  "Nf6",
			fen: "r1bqkb1r/pppp1ppp/2n2n2/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 2 4",
		},
		{
			mv:  "O-O",
			fen: "r1bqkb1r/pppp1ppp/2n2n2/1B2p3/4P3/5N2/PPPP1PPP/RNBQ1RK1 b kq - 3 4",
		},
		{
			mv:  "Nxe4",
			fen: "r1bqkb1r/pppp1ppp/2n5/1B2p3/4n3/5N2/PPPP1PPP/RNBQ1RK1 w kq - 0 5",
		},
		{
			mv:  "Re1",
			fen: "r1bqkb1r/pppp1ppp/2n5/1B2p3/4n3/5N2/PPPP1PPP/RNBQR1K1 b kq - 1 5",
		},
		{
			mv:  "Qh4",
			fen: "r1b1kb1r/pppp1ppp/2n5/1B2p3/4n2q/5N2/PPPP1PPP/RNBQR1K1 w kq - 2 6",
		},
		{
			mv:  "Kh1",
			fen: "r1b1kb1r/pppp1ppp/2n5/1B2p3/4n2q/5N2/PPPP1PPP/RNBQR2K b kq - 3 6",
		},
	} {
		m, err := LegalMoveFromSAN(v.mv, b)
		require.NoError(t, err)

		san, err := m.Styled(b, MoveStyleSAN)
		require.NoError(t, err)
		assert.Equal(t, v.mv, san)

		_, err = b.MakeMove(m)
		require.NoError(t, err)

		assert.Equal(t, v.fen, b.FEN())

		b2, err := NewBoard(b.r)
		require.NoError(t, err)
		require.Equal(t, *b, *b2)
	}
}

func TestSAN(t *testing.T) {
	for _, v := range []struct {
		fen string
		san string
		uci string
		err string
		cvt string
	}{
		{
			fen: "8/8/1p6/2P5/1p5k/2P5/7K/8 w - - 0 1",
			san: "cb",
			err: "convert san: ambiguous move: c5b6 and c3b4 are candidates",
		},
		{
			fen: "8/8/1p6/2P5/1p5k/2P5/7K/8 w - - 0 1",
			san: "cxb4",
			uci: "c3b4",
		},
		{
			fen: "8/8/1p6/2P5/1p5k/2P5/7K/8 w - - 0 1",
			san: "cxb6",
			uci: "c5b6",
		},
		{
			fen: "8/8/1p6/2P5/1p5k/2P5/7K/8 w - - 0 1",
			san: "cd",
			err: "convert san: no such move",
		},
		{
			fen: "k5K1/8/5q2/6n1/8/2P5/5q2/8 b - - 0 1",
			san: "Qe5",
			uci: "f6e5",
		},
		{
			fen: "k5K1/8/5q2/6n1/8/2P5/5q2/8 b - - 0 1",
			san: "Qd4",
			err: "convert san: ambiguous move: f6d4 and f2d4 are candidates",
		},
		{
			fen: "k5K1/8/5q2/6n1/8/2P5/5q2/8 b - - 0 1",
			san: "Qxc3",
			uci: "f6c3",
		},
		{
			fen: "k5K1/8/5q2/6n1/8/2P5/5q2/8 b - - 0 1",
			san: "Qb2",
			uci: "f2b2",
		},
		{
			fen: "k5K1/8/5q2/6n1/8/2P5/5q2/8 b - - 0 1",
			san: "Qa1",
			err: "convert san: no such move",
		},
		{
			fen: "k5K1/8/5q2/6n1/8/2P5/5q2/8 b - - 0 1",
			san: "Qg5",
			err: "convert san: no such move",
		},
		{
			fen: "k5K1/8/5q2/6n1/8/2P5/5q2/8 b - - 0 1",
			san: "Qe3",
			uci: "f2e3",
		},
		{
			fen: "k5K1/8/5q2/6n1/8/2P5/5q2/8 b - - 0 1",
			san: "Q2d4",
			uci: "f2d4",
		},
		{
			fen: "k5K1/8/5q2/6n1/8/2P5/5q2/8 b - - 0 1",
			san: "Q6d4",
			uci: "f6d4",
		},
		{
			fen: "k5K1/8/5q2/6n1/8/2P5/5q2/8 b - - 0 1",
			san: "Qf6d4",
			uci: "f6d4",
			cvt: "Q6d4",
		},
		{
			fen: "k5K1/8/5q2/6n1/8/2P5/5q2/8 b - - 0 1",
			san: "Qfe5",
			uci: "f6e5",
			cvt: "Qe5",
		},
		{
			fen: "k5K1/8/5q2/6n1/8/2P5/5q2/8 b - - 0 1",
			san: "Q6e5",
			uci: "f6e5",
			cvt: "Qe5",
		},
		{
			fen: "k5K1/8/5q2/6n1/8/2P5/5q2/8 b - - 0 1",
			san: "Qf6e5",
			uci: "f6e5",
			cvt: "Qe5",
		},
		{
			fen: "k5K1/8/5q2/6n1/8/2P5/5q2/8 b - - 0 1",
			san: "Qge5",
			err: "convert san: no such move",
		},
		{
			fen: "k5K1/8/5q2/6n1/8/2P5/5q2/8 b - - 0 1",
			san: "Q5e5",
			err: "convert san: no such move",
		},
		{
			fen: "k5K1/8/5q2/6n1/8/2P5/5q2/8 b - - 0 1",
			san: "Qg5e5",
			err: "convert san: no such move",
		},
		{
			fen: "k5K1/8/5q2/6n1/8/2P5/5q2/8 b - - 0 1",
			san: "Qfd4",
			err: "convert san: ambiguous move: f6d4 and f2d4 are candidates",
		},
		{
			fen: "k5K1/8/5q2/6n1/8/2P5/5q2/8 b - - 0 1",
			san: "Ka7",
			uci: "a8a7",
		},
		{
			fen: "k5K1/8/5q2/6n1/8/2P5/5q2/8 b - - 0 1",
			san: "Kaa7",
			uci: "a8a7",
			cvt: "Ka7",
		},
		{
			fen: "k5K1/8/p4q2/1P4n1/8/2P5/5q2/8 b - - 0 1",
			san: "Qxe5",
			err: "convert san: capture is expected",
		},
		{
			fen: "k5K1/8/p4q2/1P4n1/8/2P5/5q2/8 b - - 0 1",
			san: "Qe5",
			uci: "f6e5",
		},
		{
			fen: "k5K1/8/p4q2/1P4n1/8/2P5/5q2/8 b - - 0 1",
			san: "Qc3",
			uci: "f6c3",
			cvt: "Qxc3",
		},
		{
			fen: "k5K1/8/p4q2/1P4n1/8/2P5/5q2/8 b - - 0 1",
			san: "Qxc3",
			uci: "f6c3",
		},
		{
			fen: "k5K1/8/p4q2/1P4n1/8/2P5/5q2/8 b - - 0 1",
			san: "axb5",
			uci: "a6b5",
		},
		{
			fen: "k5K1/8/p4q2/1P4n1/8/2P5/5q2/8 b - - 0 1",
			san: "b5",
			err: "convert san: move is not semi-legal",
		},
		{
			fen: "k5K1/8/p4q2/1P4n1/8/2P5/5q2/8 b - - 0 1",
			san: "axa5",
			err: "convert san: capture is expected",
		},
		{
			fen: "k5K1/8/p4q2/1P4n1/8/2P5/5q2/8 b - - 0 1",
			san: "a5",
			uci: "a6a5",
		},
		{
			fen: "k5K1/8/p4q2/1P4n1/8/2P5/5q2/8 b - - 0 1",
			san: "bxa8",
			err: "convert san: move is not well-formed",
		},
		{
			fen: "k5K1/8/p4q2/1P4n1/8/2P5/5q2/8 b - - 0 1",
			san: "c8",
			err: "convert san: move is not well-formed",
		},
		{
			fen: "8/8/8/4p3/3P4/8/3P4/5K1k w - - 0 1",
			san: "de",
			uci: "d4e5",
			cvt: "dxe5",
		},
		{
			fen: "8/8/8/2PpP3/8/8/5k1K/8 w - d6 0 1",
			san: "cd",
			uci: "c5d6",
			cvt: "cxd6",
		},
		{
			fen: "8/8/8/2PpP3/8/8/5k1K/8 w - d6 0 1",
			san: "ed",
			uci: "e5d6",
			cvt: "exd6",
		},
		{
			fen: "8/8/8/3pP3/2P5/8/5k1K/8 w - d6 0 1",
			san: "cd",
			uci: "c4d5",
			cvt: "cxd5",
		},
		{
			fen: "8/8/8/3pP3/2P5/8/5k1K/8 w - d6 0 1",
			san: "ed",
			uci: "e5d6",
			cvt: "exd6",
		},
		{
			fen: "2n2n1n/3P2P1/8/8/8/8/3K1k2/8 w - - 0 1",
			san: "d8N",
			uci: "d7d8n",
			cvt: "d8=N",
		},
		{
			fen: "2n2n1n/3P2P1/8/8/8/8/3K1k2/8 w - - 0 1",
			san: "dcB",
			uci: "d7c8b",
			cvt: "dxc8=B",
		},
		{
			fen: "2n2n1n/3P2P1/8/8/8/8/3K1k2/8 w - - 0 1",
			san: "gf=R",
			uci: "g7f8r",
			cvt: "gxf8=R+",
		},
		{
			fen: "2n2n1n/3P2P1/8/8/8/8/3K1k2/8 w - - 0 1",
			san: "gh=Q",
			uci: "g7h8q",
			cvt: "gxh8=Q",
		},
		{
			fen: "8/8/8/8/3p3k/2P5/1PP4K/8 w - - 0 1",
			san: "b3",
			uci: "b2b3",
		},
		{
			fen: "8/8/8/8/3p3k/2P5/1PP4K/8 w - - 0 1",
			san: "b4",
			uci: "b2b4",
		},
		{
			fen: "8/8/8/8/3p3k/2P5/1PP4K/8 w - - 0 1",
			san: "c4",
			uci: "c3c4",
		},
		{
			fen: "8/8/8/8/3p3k/2P5/1PP4K/8 w - - 0 1",
			san: "cd",
			uci: "c3d4",
			cvt: "cxd4",
		},
		{
			fen: "4k3/6K1/8/2N5/8/8/8/N7 w - - 0 1",
			san: "Nab3",
			uci: "a1b3",
		},
		{
			fen: "4k3/6K1/8/N7/8/8/8/N7 w - - 0 1",
			san: "N1b3",
			uci: "a1b3",
		},
		{
			fen: "4k3/6K1/8/8/8/8/8/N1N5 w - - 0 1",
			san: "Nab3",
			uci: "a1b3",
		},
		{
			fen: "4k3/6K1/8/N1N5/8/8/8/N1N5 w - - 0 1",
			san: "Na1b3",
			uci: "a1b3",
		},
		{
			fen: "5k2/8/5K2/8/3R3R/8/8/b7 w - - 0 1",
			san: "Rf4",
			uci: "h4f4",
		},
		{
			fen: "4k3/6K1/8/2N5/8/1r6/8/N7 w - - 0 1",
			san: "Naxb3",
			uci: "a1b3",
		},
		{
			fen: "4k3/6K1/8/N7/8/1r6/8/N7 w - - 0 1",
			san: "N1xb3",
			uci: "a1b3",
		},
		{
			fen: "4k3/6K1/8/8/8/1r6/8/N1N5 w - - 0 1",
			san: "Naxb3",
			uci: "a1b3",
		},
		{
			fen: "4k3/6K1/8/N1N5/8/1r6/8/N1N5 w - - 0 1",
			san: "Na1xb3",
			uci: "a1b3",
		},
		{
			fen: "1r5k/8/8/8/8/6p1/r7/5K2 b - - 0 1",
			san: "g2+",
			uci: "g3g2",
		},
		{
			fen: "1r5k/8/8/8/8/6p1/r7/5K2 b - - 0 1",
			san: "g2",
			uci: "g3g2",
			cvt: "g2+",
		},
		{
			fen: "1r5k/8/8/8/8/6p1/r7/5K2 b - - 0 1",
			san: "Rb1#",
			uci: "b8b1",
		},
		{
			fen: "1r5k/8/8/8/8/6p1/r7/5K2 b - - 0 1",
			san: "Rb1",
			uci: "b8b1",
			cvt: "Rb1#",
		},
		{
			fen: "1r5k/8/8/8/8/6p1/r7/5K2 b - - 0 1",
			san: "Rb1+",
			uci: "b8b1",
			cvt: "Rb1#",
		},
		{
			fen: "1r5k/8/8/8/8/6p1/r7/5K2 b - - 0 1",
			san: "Rb1++",
			uci: "b8b1",
			cvt: "Rb1#",
		},
		{
			fen: "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
			san: "b8c6",
			uci: "b8c6",
			cvt: "Nc6",
		},
		{
			fen: "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
			san: "b8d7",
			err: "convert san: move is not semi-legal",
		},
		{
			fen: "r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQkq - 0 1",
			san: "O-O",
			uci: "e1g1",
			cvt: "O-O",
		},
		{
			fen: "r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R w KQkq - 0 1",
			san: "0-0-0",
			uci: "e1c1",
			cvt: "O-O-O",
		},
		{
			fen: "r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R b KQkq - 0 1",
			san: "0-0",
			uci: "e8g8",
			cvt: "O-O",
		},
		{
			fen: "r3k2r/pppppppp/8/8/8/8/PPPPPPPP/R3K2R b KQkq - 0 1",
			san: "O-O-O",
			uci: "e8c8",
			cvt: "O-O-O",
		},
	} {
		b, err := BoardFromFEN(v.fen)
		require.NoError(t, err)

		m, err := LegalMoveFromSAN(v.san, b)
		if v.err != "" {
			assert.EqualError(t, err, v.err)
			continue
		}
		require.NoError(t, err)

		err = m.Validate(b)
		require.NoError(t, err)

		um, err := LegalMoveFromUCI(v.uci, b)
		require.NoError(t, err)

		assert.Equal(t, um, m)

		if v.cvt == "" {
			v.cvt = v.san
		}
		san, err := m.Styled(b, MoveStyleSAN)
		require.NoError(t, err)
		assert.Equal(t, v.cvt, san)

		if v.cvt != v.san {
			m2, err := LegalMoveFromSAN(v.cvt, b)
			require.NoError(t, err)
			assert.Equal(t, m2, m)
		}
	}
}

func TestSANStyled(t *testing.T) {
	b, err := BoardFromFEN("8/2P5/8/8/8/8/4k1K1/8 w - - 0 1")
	require.NoError(t, err)

	for _, v := range []struct {
		uci   string
		style MoveStyle
		res   string
	}{
		{
			uci:   "g2h2",
			style: MoveStyleFancySAN,
			res:   "♔h2",
		},
		{
			uci:   "g2h2",
			style: MoveStyleSAN,
			res:   "Kh2",
		},
		{
			uci:   "c7c8b",
			style: MoveStyleFancySAN,
			res:   "c8♗",
		},
		{
			uci:   "c7c8b",
			style: MoveStyleSAN,
			res:   "c8=B",
		},
	} {
		m, err := LegalMoveFromUCI(v.uci, b)
		require.NoError(t, err)
		s, err := m.Styled(b, v.style)
		require.NoError(t, err)
		assert.Equal(t, v.res, s)
	}
}
