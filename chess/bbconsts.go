package chess

var bbDiag = [15]Bitboard{
	0x0000000000000001,
	0x0000000000000102,
	0x0000000000010204,
	0x0000000001020408,
	0x0000000102040810,
	0x0000010204081020,
	0x0001020408102040,
	0x0102040810204080,
	0x0204081020408000,
	0x0408102040800000,
	0x0810204080000000,
	0x1020408000000000,
	0x2040800000000000,
	0x4080000000000000,
	0x8000000000000000,
}

var bbAntidiag = [15]Bitboard{
	0x0100000000000000,
	0x0201000000000000,
	0x0402010000000000,
	0x0804020100000000,
	0x1008040201000000,
	0x2010080402010000,
	0x4020100804020100,
	0x8040201008040201,
	0x0080402010080402,
	0x0000804020100804,
	0x0000008040201008,
	0x0000000080402010,
	0x0000000000804020,
	0x0000000000008040,
	0x0000000000000080,
}

func BbDiag(d int) Bitboard {
	return bbDiag[d]
}

func BbAntidiag(d int) Bitboard {
	return bbAntidiag[d]
}

func BbRank(r Rank) Bitboard {
	return Bitboard(0xff) << (int(r) << 3)
}

func BbFile(f File) Bitboard {
	return Bitboard(0x0101010101010101) << int(f)
}

const (
	BbLight Bitboard = 0xaa55aa55aa55aa55
	BbDark  Bitboard = 0x55aa55aa55aa55aa
)

const (
	bbFileFrame Bitboard = 0xff000000000000ff
	bbRankFrame Bitboard = 0x8181818181818181
	bbDiagFrame Bitboard = 0xff818181818181ff
)

func bbCastlingPass(c Color, s CastlingSide) Bitboard {
	x := chooseByCastlingSide(s, Bitboard(0x0e), Bitboard(0x60))
	return x << castlingOffset(c)
}

func bbCastlingSrcs(c Color, s CastlingSide) Bitboard {
	x := chooseByCastlingSide(s, Bitboard(0x11), Bitboard(0x90))
	return x << castlingOffset(c)
}

const bbCastlingAllSrcs Bitboard = 0x91 | (0x91 << 56)
