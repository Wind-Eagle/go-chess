package chess

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBetweenBishop(t *testing.T) {
	b4 := CoordFromParts(FileB, Rank4)
	e7 := CoordFromParts(FileE, Rank7)
	res := BbEmpty.With2(FileC, Rank5).With2(FileD, Rank6)
	assert.Equal(t, res, betweenBishopStrict(b4, e7))
	assert.Equal(t, res, betweenBishopStrict(e7, b4))

	f3 := CoordFromParts(FileF, Rank3)
	c6 := CoordFromParts(FileC, Rank6)
	res = BbEmpty.With2(FileE, Rank4).With2(FileD, Rank5)
	assert.Equal(t, res, betweenBishopStrict(f3, c6))
	assert.Equal(t, res, betweenBishopStrict(c6, f3))
}

func TestBetweenRook(t *testing.T) {
	b4 := CoordFromParts(FileB, Rank4)
	e4 := CoordFromParts(FileE, Rank4)
	res := BbEmpty.With2(FileC, Rank4).With2(FileD, Rank4)
	assert.Equal(t, res, betweenRookStrict(b4, e4))
	assert.Equal(t, res, betweenRookStrict(e4, b4))

	d3 := CoordFromParts(FileD, Rank3)
	d6 := CoordFromParts(FileD, Rank6)
	res = BbEmpty.With2(FileD, Rank4).With2(FileD, Rank5)
	assert.Equal(t, res, betweenRookStrict(d3, d6))
	assert.Equal(t, res, betweenRookStrict(d6, d3))
}
