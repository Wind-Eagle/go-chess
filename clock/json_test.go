package clock

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON(t *testing.T) {
	mustControl := func(s string) Control {
		c, err := ControlFromString(s)
		if err != nil {
			panic(err)
		}
		return c
	}

	for _, v := range []struct {
		j string
		o any
	}{
		{
			j: `"40/5+1.23"`,
			o: mustControl("40/5+1.23"),
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
