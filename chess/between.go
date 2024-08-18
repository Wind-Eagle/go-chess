package chess

func fillBetweenBishop[M ~func(Coord) Bitboard](res []Bitboard, mask M) {
	if len(res) != int(CoordMax) {
		panic("expected len(res) == int(CoordMax)")
	}
	for c := range CoordMax {
		val := BbDiag(c.Diag()) | BbAntidiag(c.Antidiag())
		res[c] = val & mask(c)
	}
}

func fillBetweenRook[M ~func(Coord) Bitboard](res []Bitboard, mask M) {
	if len(res) != int(CoordMax) {
		panic("expected len(res) == int(CoordMax)")
	}
	for c := range CoordMax {
		val := BbFile(c.File()) | BbRank(c.Rank())
		res[c] = val & mask(c)
	}
}

func maskNe(c Coord) Bitboard {
	return ^BitboardFromCoord(c)
}

func maskLt(c Coord) Bitboard {
	return BitboardFromCoord(c) - 1
}

func maskGt(c Coord) Bitboard {
	return maskNe(c) & ^maskLt(c)
}

var (
	bishopLt [CoordMax]Bitboard
	bishopGt [CoordMax]Bitboard
	bishopNe [CoordMax]Bitboard
	rookLt   [CoordMax]Bitboard
	rookGt   [CoordMax]Bitboard
	rookNe   [CoordMax]Bitboard
)

func init() {
	fillBetweenBishop(bishopLt[:], maskLt)
	fillBetweenBishop(bishopGt[:], maskGt)
	fillBetweenBishop(bishopNe[:], maskNe)
	fillBetweenRook(rookLt[:], maskLt)
	fillBetweenRook(rookGt[:], maskGt)
	fillBetweenRook(rookNe[:], maskNe)
}

func sort2(src, dst Coord) (Coord, Coord) {
	if src < dst {
		return src, dst
	} else {
		return dst, src
	}
}

func betweenBishopStrict(src, dst Coord) Bitboard {
	src, dst = sort2(src, dst)
	return bishopGt[src] & bishopLt[dst]
}

func betweenRookStrict(src, dst Coord) Bitboard {
	src, dst = sort2(src, dst)
	return rookGt[src] & rookLt[dst]
}

func isBishopMoveValid(src, dst Coord) bool {
	return bishopNe[src].Has(dst)
}

func isRookMoveValid(src, dst Coord) bool {
	return rookNe[src].Has(dst)
}
