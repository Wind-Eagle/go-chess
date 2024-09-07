package uci

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON(t *testing.T) {
	for _, v := range []struct {
		j string
		o any
	}{
		{
			j: `{"mate":true,"v":1}`,
			o: ScoreMate(1),
		},
		{
			j: `{"mate":false,"v":123}`,
			o: ScoreCentipawns(123),
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
