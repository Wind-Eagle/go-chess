package chess

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBitboardIter(t *testing.T) {
	bb := BbEmpty.
		With2(FileA, Rank4).
		With2(FileE, Rank2).
		With2(FileF, Rank3)

	var res []Coord
	for !bb.IsEmpty() {
		res = append(res, bb.Next())
	}

	exp := []Coord{
		CoordFromParts(FileA, Rank4),
		CoordFromParts(FileF, Rank3),
		CoordFromParts(FileE, Rank2),
	}
	assert.Equal(t, exp, res)
}

func TestBitboardOps(t *testing.T) {
	ca := CoordFromParts(FileA, Rank4)
	cb := CoordFromParts(FileE, Rank2)
	cc := CoordFromParts(FileF, Rank3)

	bb1 := BbEmpty.With(ca).With(cb)
	bb2 := BbEmpty.With(cb).With(cc)
	assert.Equal(t, BbEmpty.With(cb), bb1&bb2)
	assert.Equal(t, BbEmpty.With(ca).With(cb).With(cc), bb1|bb2)
	assert.Equal(t, BbEmpty.With(ca).With(cc), bb1^bb2)

	assert.Equal(t, 62, (^bb1).Len())
	assert.Equal(t, 2, bb1.Len())
}

func TestBitboardFormat(t *testing.T) {
	bb := BbEmpty.
		With2(FileA, Rank4).
		With2(FileE, Rank2).
		With2(FileF, Rank3).
		With2(FileH, Rank8)
	assert.Equal(t, "00000001/00000000/00000000/00000000/10000000/00000100/00001000/00000000", bb.String())
}
