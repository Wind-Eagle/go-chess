package chess

import (
	"math/rand/v2"
)

type ZHash [1]uint64

func (z ZHash) Xor(o ZHash) ZHash {
	return ZHash{z[0] ^ o[0]}
}

func (z *ZHash) XorEq(o ZHash) {
	*z = z.Xor(o)
}

func randZobrist() ZHash {
	return ZHash{rand.Uint64()}
}

var (
	zobristCells         [CellMax][CoordMax]ZHash
	zobristMoveSide      ZHash
	zobristCastling      [CastlingRightsMax]ZHash
	zobristEnpassant     [CoordMax]ZHash
	zobristCastlingDelta [ColorMax][CastlingSideMax]ZHash
)

func init() {
	for i := range CellMax {
		if i == CellEmpty {
			for j := range CoordMax {
				zobristCells[i][j] = ZHash{}
			}
			continue
		}
		for j := range CoordMax {
			zobristCells[i][j] = randZobrist()
		}
	}

	zobristMoveSide = randZobrist()

	for i := range zobristCastling {
		zobristCastling[i] = randZobrist()
	}

	for i := range CoordMax {
		zobristEnpassant[i] = randZobrist()
	}

	for c := range ColorMax {
		rook := CellFromParts(c, PieceRook)
		king := CellFromParts(c, PieceKing)
		rank := homeRank(c)
		r := func(f File) Coord { return CoordFromParts(f, rank) }
		zobristCastlingDelta[c][CastlingKingside] =
			zobristCells[king][r(FileE)].
				Xor(zobristCells[king][r(FileG)]).
				Xor(zobristCells[rook][r(FileH)]).
				Xor(zobristCells[rook][r(FileF)])
		zobristCastlingDelta[c][CastlingQueenside] =
			zobristCells[king][r(FileE)].
				Xor(zobristCells[king][r(FileC)]).
				Xor(zobristCells[rook][r(FileA)]).
				Xor(zobristCells[rook][r(FileD)])
	}
}
