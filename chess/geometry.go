package chess

func chooseByColor[T any](c Color, w, b T) T {
	if c == ColorWhite {
		return w
	} else {
		return b
	}
}

func chooseByCastlingSide[T any](s CastlingSide, q, k T) T {
	if s == CastlingQueenside {
		return q
	} else {
		return k
	}
}

func homeRank(c Color) Rank {
	return chooseByColor(c, Rank1, Rank8)
}

func pawnHomeRank(c Color) Rank {
	return chooseByColor(c, Rank2, Rank7)
}

func pawnDoubleDstRank(c Color) Rank {
	return chooseByColor(c, Rank4, Rank5)
}

func promoteSrcRank(c Color) Rank {
	return chooseByColor(c, Rank7, Rank2)
}

func promoteDstRank(c Color) Rank {
	return chooseByColor(c, Rank8, Rank1)
}

func enpassantSrcRank(c Color) Rank {
	return chooseByColor(c, Rank5, Rank4)
}

func enpassantDstRank(c Color) Rank {
	return chooseByColor(c, Rank6, Rank3)
}

func pawnForwardDelta(c Color) int8 {
	return chooseByColor(c, int8(-8), 8)
}

func pawnLeftDelta(c Color) int8 {
	return chooseByColor(c, int8(-9), 7)
}

func pawnRightDelta(c Color) int8 {
	return chooseByColor(c, int8(-7), 9)
}

func castlingDstFile(s CastlingSide) File {
	return chooseByCastlingSide(s, FileC, FileG)
}

func castlingOffset(c Color) uint8 {
	return chooseByColor(c, uint8(56), 0)
}

var (
	whitePawnDeltas = []CoordDelta{
		{File: -1, Rank: -1},
		{File: 1, Rank: -1},
	}
	blackPawnDeltas = []CoordDelta{
		{File: -1, Rank: 1},
		{File: 1, Rank: 1},
	}
	kingDeltas = []CoordDelta{
		{File: -1, Rank: -1},
		{File: -1, Rank: 0},
		{File: -1, Rank: 1},
		{File: 0, Rank: -1},
		{File: 0, Rank: 1},
		{File: 1, Rank: -1},
		{File: 1, Rank: 0},
		{File: 1, Rank: 1},
	}
	knightDeltas = []CoordDelta{
		{File: -2, Rank: -1},
		{File: -2, Rank: 1},
		{File: 2, Rank: -1},
		{File: 2, Rank: 1},
		{File: -1, Rank: -2},
		{File: -1, Rank: 2},
		{File: 1, Rank: -2},
		{File: 1, Rank: 2},
	}
	rookDeltas = []CoordDelta{
		{File: 0, Rank: 1},
		{File: 0, Rank: -1},
		{File: -1, Rank: 0},
		{File: 1, Rank: 0},
	}
	bishopDeltas = []CoordDelta{
		{File: -1, Rank: -1},
		{File: -1, Rank: 1},
		{File: 1, Rank: -1},
		{File: 1, Rank: 1},
	}
)
