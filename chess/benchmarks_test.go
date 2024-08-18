package chess_test

import (
	"testing"

	"github.com/alex65536/go-chess/chess"
)

var names = [10]string{
	"initial",
	"sicilian",
	"middle",
	"openPosition",
	"queen",
	"pawnMove",
	"pawnAttack",
	"pawnPromote",
	"cydonia",
	"max",
}

var boards = [10]string{
	"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
	"r1b1k2r/2qnbppp/p2ppn2/1p4B1/3NPPP1/2N2Q2/PPP4P/2KR1B1R w kq - 0 11",
	"1rq1r1k1/1p3ppp/pB3n2/3ppP2/Pbb1P3/1PN2B2/2P2QPP/R1R4K w - - 1 21",
	"4r1k1/3R1ppp/8/5P2/p7/6PP/4pK2/1rN1B3 w - - 4 43",
	"6K1/8/8/1k3q2/3Q4/8/8/8 w - - 0 1",
	"4k3/pppppppp/8/8/8/8/PPPPPPPP/4K3 w - - 0 1",
	"4k3/8/8/pppppppp/PPPPPPPP/8/8/4K3 w - - 0 1",
	"8/PPPPPPPP/8/2k1K3/8/8/pppppppp/8 w - - 0 1",
	"5K2/1N1N1N2/8/1N1N1N2/1n1n1n2/8/1n1n1n2/5k2 w - - 0 1",
	"3Q4/1Q4Q1/4Q3/2Q4R/Q4Q2/3Q4/NR4Q1/kN1BB1K1 w - - 0 1",
}

func BenchmarkGenMoves(b *testing.B) {
	for i := range boards {
		fen := boards[i]
		b.Run(names[i], func(b *testing.B) {
			board, err := chess.BoardFromFEN(fen)
			if err != nil {
				b.Fatalf("cannot parse board: %v", err)
			}
			var buf [256]chess.Move
			b.ResetTimer()
			for range b.N {
				board.GenSemilegalMoves(chess.MoveGenAll, buf[:0])
			}
		})
	}
}

func BenchmarkGenMovesLegal(b *testing.B) {
	for i := range boards {
		fen := boards[i]
		b.Run(names[i], func(b *testing.B) {
			board, err := chess.BoardFromFEN(fen)
			if err != nil {
				b.Fatalf("cannot parse board: %v", err)
			}
			var buf [256]chess.Move
			b.ResetTimer()
			for range b.N {
				board.GenLegalMoves(chess.MoveGenAll, buf[:0])
			}
		})
	}
}

func BenchmarkMakeMove(b *testing.B) {
	for i := range boards {
		fen := boards[i]
		b.Run(names[i], func(b *testing.B) {
			board, err := chess.BoardFromFEN(fen)
			if err != nil {
				b.Fatalf("cannot parse board: %v", err)
			}
			moves := board.GenSemilegalMoves(chess.MoveGenAll, nil)
			b.ResetTimer()
			for range b.N {
				for _, mv := range moves {
					u := board.MakeLegalMove(mv)
					board.UnmakeMove(u)
				}
			}
		})
	}
}

func BenchmarkMakeMoveChecked(b *testing.B) {
	for i := range boards {
		fen := boards[i]
		b.Run(names[i], func(b *testing.B) {
			board, err := chess.BoardFromFEN(fen)
			if err != nil {
				b.Fatalf("cannot parse board: %v", err)
			}
			moves := board.GenSemilegalMoves(chess.MoveGenAll, nil)
			b.ResetTimer()
			for range b.N {
				for _, mv := range moves {
					u, err := board.MakeMove(mv)
					if err == nil {
						board.UnmakeMove(u)
					}
				}
			}
		})
	}
}

func BenchmarkMoveSemiValidate(b *testing.B) {
	for i := range boards {
		fen := boards[i]
		b.Run(names[i], func(b *testing.B) {
			board, err := chess.BoardFromFEN(fen)
			if err != nil {
				b.Fatalf("cannot parse board: %v", err)
			}
			moves := board.GenSemilegalMoves(chess.MoveGenAll, nil)
			b.ResetTimer()
			for range b.N {
				for _, mv := range moves {
					_ = mv.SemiValidate(board)
				}
			}
		})
	}
}

func BenchmarkIsAttacked(b *testing.B) {
	for i := range boards {
		fen := boards[i]
		b.Run(names[i], func(b *testing.B) {
			board, err := chess.BoardFromFEN(fen)
			if err != nil {
				b.Fatalf("cannot parse board: %v", err)
			}
			b.ResetTimer()
			for range b.N {
				for color := range chess.ColorMax {
					for coord := range chess.CoordMax {
						board.IsCellAttacked(coord, color)
					}
				}
			}
		})
	}
}

func BenchmarkKingAttack(b *testing.B) {
	for i := range boards {
		fen := boards[i]
		b.Run(names[i], func(b *testing.B) {
			board, err := chess.BoardFromFEN(fen)
			if err != nil {
				b.Fatalf("cannot parse board: %v", err)
			}
			b.ResetTimer()
			for range b.N {
				board.IsCheck()
			}
		})
	}
}

func BenchmarkHasLegalMoves(b *testing.B) {
	for i := range boards {
		fen := boards[i]
		b.Run(names[i], func(b *testing.B) {
			board, err := chess.BoardFromFEN(fen)
			if err != nil {
				b.Fatalf("cannot parse board: %v", err)
			}
			b.ResetTimer()
			for range b.N {
				board.HasLegalMoves()
			}
		})
	}
}
