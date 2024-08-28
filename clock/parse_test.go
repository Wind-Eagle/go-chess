package clock

import (
	"math/rand/v2"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDurationFormat(t *testing.T) {
	for _, v := range []struct {
		src int64
		res string
	}{
		{src: 0, res: "0"},
		{src: 1_000_000_000, res: "1"},
		{src: 250_000_000_000, res: "250"},
		{src: -1_000_000_000, res: "-1"},
		{src: -250_000_000_000, res: "-250"},
		{src: 1_010_000_000, res: "1.01"},
		{src: 1_000_000_001, res: "1.000000001"},
		{src: 1_300_000_000, res: "1.3"},
		{src: 901_234_500_000, res: "901.2345"},
		{src: -42_010_000_000, res: "-42.01"},
		{src: -1_000_300_000_000, res: "-1000.3"},
		{src: -1_234_500_000, res: "-1.2345"},
		{src: 9223372036854775807, res: "9223372036.854775807"},
		{src: -9223372036854775807, res: "-9223372036.854775807"},
		{src: -9223372036854775808, res: "-9223372036.854775808"},
	} {
		assert.Equal(t, v.res, formatDuration(time.Duration(v.src)*time.Nanosecond))
	}
}

func TestDurationParse(t *testing.T) {
	const billi = 1_000_000_000
	for _, v := range []struct {
		src string
		res int64
		err string
	}{
		{src: "0", res: 0},
		{src: "-0", res: 0},
		{src: "+0", res: 0},
		{src: "0.0000000", res: 0},
		{src: "+0.0000000", res: 0},
		{src: "-0.0000000", res: 0},
		{src: "0.0000000009", res: 0},
		{src: "1", res: 1 * billi},
		{src: "1.", res: 1 * billi},
		{src: "-1", res: -1 * billi},
		{src: "+1", res: 1 * billi},
		{src: "+1.0000000000000000000000000000000000000000000000000000000000000001", res: 1 * billi},
		{src: "-1.000", res: -1 * billi},
		{src: "250", res: 250 * billi},
		{src: "-2500", res: -2500 * billi},
		{src: "9223372036", res: 9223372036 * billi},
		{src: "9223372037", err: "out of range"},
		{src: "-9223372036", res: -9223372036 * billi},
		{src: "-9223372037", err: "out of range"},
		{src: "1.01", res: 1_010_000_000},
		{src: "1.0100000001", res: 1_010_000_000},
		{src: "1.000000001", res: 1_000_000_001},
		{src: "1.01000000000000000000000u", err: "fractional part contains non-digits"},
		{src: "1.3", res: 1_300_000_000},
		{src: "901.2345", res: 901_234_500_000},
		{src: ".1", err: "parse integer part: strconv.ParseInt: parsing \"\": invalid syntax"},
		{src: "-42.01", res: -42_010_000_000},
		{src: "-1000.3", res: -1_000_300_000_000},
		{src: "-1.2345", res: -1_234_500_000},
		{src: "-1.234500", res: -1_234_500_000},
		{src: "9223372036.854775807", res: 9223372036854775807},
		{src: "9223372036.854775808", err: "out of range"},
		{src: "-9223372036.854775807", res: -9223372036854775807},
		{src: "-9223372036.854775808", res: -9223372036854775808},
		{src: "-9223372036.854775809", err: "out of range"},
	} {
		res, err := parseDuration(v.src)
		if v.err != "" {
			assert.EqualError(t, err, v.err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, v.res, res.Nanoseconds())
		}
	}
}

func TestStress(t *testing.T) {
	rng := rand.New(rand.NewPCG(0x3141592653589793, 0x2384626433832795))
	for range 50_000 {
		i := int64(0)
		for range rng.IntN(18) + 1 {
			i *= 10
			if rng.IntN(3) != 0 {
				i += rng.Int64N(9) + 1
			}
		}
		s := formatDuration(time.Duration(i) * time.Nanosecond)
		parsed, err := parseDuration(s)
		require.NoError(t, err)
		require.Equal(t, i, parsed.Nanoseconds())
	}
}
