package chess

var (
	kingAttackTab   [CoordMax]Bitboard
	knightAttackTab [CoordMax]Bitboard
	pawnAttackTab   [ColorMax][CoordMax]Bitboard
)

func fillAttackTab(res []Bitboard, deltas []CoordDelta) {
	if len(res) != int(CoordMax) {
		panic("expected len(res) == int(CoordMax)")
	}
	for c := range CoordMax {
		bb := BbEmpty
		for _, d := range deltas {
			if nc, ok := c.Shift(d).TryGet(); ok {
				bb.Set(nc)
			}
		}
		res[c] = bb
	}
}

func init() {
	fillAttackTab(pawnAttackTab[ColorWhite][:], whitePawnDeltas)
	fillAttackTab(pawnAttackTab[ColorBlack][:], blackPawnDeltas)
	fillAttackTab(kingAttackTab[:], kingDeltas)
	fillAttackTab(knightAttackTab[:], knightDeltas)
}

func pawnAttacks(color Color, coord Coord) Bitboard {
	return pawnAttackTab[color][coord]
}

func kingAttacks(c Coord) Bitboard {
	return kingAttackTab[c]
}

func knightAttacks(c Coord) Bitboard {
	return knightAttackTab[c]
}
