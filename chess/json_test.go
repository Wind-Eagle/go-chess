package chess

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON(t *testing.T) {
	mustBoard := func(fen string) *Board {
		b, err := BoardFromFEN(fen)
		if err != nil {
			panic(err)
		}
		return b
	}

	for _, v := range []struct {
		j string
		o any
	}{
		{
			j: `"r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3"`,
			o: mustBoard("r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3").Raw(),
		},
		{
			j: `"r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3"`,
			o: mustBoard("r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3"),
		},
		{
			j: `"e2e4"`,
			o: SimpleUCIMove(CoordFromParts(FileE, Rank2), CoordFromParts(FileE, Rank4)),
		},
		{
			j: `"w"`,
			o: ColorWhite,
		},
		{
			j: `"*"`,
			o: StatusRunning,
		},
		{
			j: `"1/2-1/2"`,
			o: StatusDraw,
		},
		{
			j: `"1-0"`,
			o: StatusWhiteWins,
		},
		{
			j: `"0-1"`,
			o: StatusBlackWins,
		},
		{
			j: `"b6"`,
			o: CoordFromParts(FileB, Rank6),
		},
		{
			j: `"b6"`,
			o: SomeCoord(CoordFromParts(FileB, Rank6)),
		},
		{
			j: `"-"`,
			o: NoCoord,
		},
		{
			j: `"Q"`,
			o: CellFromParts(ColorWhite, PieceQueen),
		},
		{
			j: `"Kq"`,
			o: CastlingRightsEmpty.With(ColorWhite, CastlingKingside).With(ColorBlack, CastlingQueenside),
		},
	} {
		val := reflect.New(reflect.TypeOf(v.o))
		err := json.Unmarshal([]byte(v.j), val.Interface())
		require.NoError(t, err)
		assert.Equal(t, v.o, val.Elem().Interface())

		j2, err := json.Marshal(v.o)
		require.NoError(t, err)
		assert.Equal(t, v.j, string(j2))
	}
}
