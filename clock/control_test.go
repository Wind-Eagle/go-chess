package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestControlItemParse(t *testing.T) {
	res, err := ControlItemFromString("40/900+1", true)
	require.NoError(t, err)
	assert.Equal(t, ControlItem{Time: 15 * time.Minute, Moves: 40, Inc: time.Second}, res)

	_, err = ControlItemFromString("bad", true)
	assert.Error(t, err)

	_, err = ControlItemFromString("-1/900", true)
	assert.EqualError(t, err, "validate: negative moves")

	_, err = ControlItemFromString("-1", true)
	assert.EqualError(t, err, "validate: negative time")

	_, err = ControlItemFromString("1+-1", true)
	assert.EqualError(t, err, "validate: negative inc")

	_, err = ControlItemFromString("900", false)
	assert.EqualError(t, err, "validate: number of moves must be specified for non-final controls")
}

func TestControlSideFormat(t *testing.T) {
	for _, v := range []struct {
		src ControlSide
		res string
	}{
		{
			src: []ControlItem{
				{Time: 15 * time.Minute, Moves: 40},
			},
			res: "40/900",
		},
		{
			src: []ControlItem{
				{Time: 3 * time.Second, Moves: 20},
				{Time: 640 * time.Millisecond},
			},
			res: "20/3:0.64",
		},
		{
			src: []ControlItem{
				{Time: 640 * time.Millisecond, Moves: 10},
				{Time: 640 * time.Millisecond, Inc: 12 * time.Millisecond, Moves: 10},
				{Time: 640 * time.Millisecond, Inc: 12 * time.Millisecond},
			},
			res: "10/0.64:10/0.64+0.012:0.64+0.012",
		},
	} {
		err := v.src.Validate()
		require.NoError(t, err)
		s := v.src.String()
		assert.Equal(t, v.res, s)
	}
}

func TestControlSideParse(t *testing.T) {
	for _, v := range []struct {
		src string
		res ControlSide
		err string
	}{
		{
			src: "40/900",
			res: []ControlItem{
				{Time: 15 * time.Minute, Moves: 40},
			},
		},
		{
			src: "20/3:0.64",
			res: []ControlItem{
				{Time: 3 * time.Second, Moves: 20},
				{Time: 640 * time.Millisecond},
			},
		},
		{
			src: "10/0.64:10/0.64+0.012:0.64+0.012",
			res: []ControlItem{
				{Time: 640 * time.Millisecond, Moves: 10},
				{Time: 640 * time.Millisecond, Inc: 12 * time.Millisecond, Moves: 10},
				{Time: 640 * time.Millisecond, Inc: 12 * time.Millisecond},
			},
		},
		{
			src: ":20/3",
			err: "parse section #1: parse time: parse integer part: strconv.ParseInt: parsing \"\": invalid syntax",
		},
		{
			src: "",
			err: "empty string",
		},
		{
			src: "/3",
			err: "parse section #1: parse moves: strconv.ParseInt: parsing \"\": invalid syntax",
		},
		{
			src: "42:20/3+",
			err: "parse section #2: parse inc: parse integer part: strconv.ParseInt: parsing \"\": invalid syntax",
		},
		{
			src: "42:42:3+",
			err: "parse section #3: parse inc: parse integer part: strconv.ParseInt: parsing \"\": invalid syntax",
		},
		{
			src: "0.01/3",
			err: "parse section #1: parse moves: strconv.ParseInt: parsing \"0.01\": invalid syntax",
		},
		{
			src: "-1",
			err: "validate: section #1: negative time",
		},
		{
			src: "20/0+1:120",
			err: "validate: initial time must be positive",
		},
	} {
		c, err := ControlSideFromString(v.src)
		if v.err != "" {
			assert.EqualError(t, err, v.err)
		} else {
			require.NoError(t, err)
			assert.NoError(t, c.Validate())
			assert.Equal(t, v.res, c)
			assert.True(t, v.res.Eq(c))
		}
	}
}

func TestControlFormatAndParse(t *testing.T) {
	for _, v := range []struct {
		src Control
		res string
	}{
		{
			src: Control{
				White: []ControlItem{
					{Time: 3 * time.Second, Moves: 20},
					{Time: 640 * time.Millisecond},
				},
				Black: []ControlItem{
					{Time: 3 * time.Second, Moves: 20},
					{Time: 640 * time.Millisecond},
				},
			},
			res: "20/3:0.64",
		},
		{
			src: Control{
				White: []ControlItem{
					{Time: 3 * time.Second, Moves: 20},
					{Time: 640 * time.Millisecond},
				},
				Black: []ControlItem{
					{Time: 3001 * time.Millisecond, Moves: 20},
					{Time: 640 * time.Millisecond},
				},
			},
			res: "20/3:0.64|20/3.001:0.64",
		},
	} {
		err := v.src.Validate()
		require.NoError(t, err)
		s := v.src.String()
		assert.Equal(t, v.res, s)
		parsed, err := ControlFromString(v.res)
		require.NoError(t, err)
		assert.Equal(t, v.src, parsed)
		assert.True(t, v.src.Eq(parsed))
	}
}
