package chess

func pawnAdvanceForward(c Color, b Bitboard) Bitboard {
	if c == ColorWhite {
		return b >> 8
	} else {
		return b << 8
	}
}

func pawnAdvanceLeft(c Color, b Bitboard) Bitboard {
	b &= ^BbFile(FileA)
	if c == ColorWhite {
		return b >> 9
	} else {
		return b << 7
	}
}

func pawnAdvanceRight(c Color, b Bitboard) Bitboard {
	b &= ^BbFile(FileH)
	if c == ColorWhite {
		return b >> 7
	} else {
		return b << 9
	}
}
