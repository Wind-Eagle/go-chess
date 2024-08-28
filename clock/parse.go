package clock

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

const nsecInSec = 1_000_000_000

func formatDuration(t time.Duration) string {
	if n := t.Nanoseconds(); n%nsecInSec == 0 {
		return strconv.FormatInt(n/nsecInSec, 10)
	}
	var n uint64
	var b strings.Builder
	if t < 0 {
		n = uint64(-t.Nanoseconds())
		_ = b.WriteByte('-')
	} else {
		n = uint64(t.Nanoseconds())
	}
	_, _ = b.WriteString(strconv.FormatUint(n/nsecInSec, 10))
	_ = b.WriteByte('.')
	rem := n % nsecInSec
	var buf [9]byte
	for i := 8; i >= 0; i-- {
		buf[i] = byte(rem%10) + '0'
		rem /= 10
	}
	size := 9
	for size != 0 && buf[size-1] == '0' {
		size--
	}
	_, _ = b.Write(buf[:size])
	return b.String()
}

func parseDuration(s string) (time.Duration, error) {
	dot := strings.IndexByte(s, '.')
	pre, suf := "", ""
	if dot == -1 {
		pre = s
	} else {
		pre = s[:dot]
		suf = s[dot+1:]
	}
	secs, err := strconv.ParseInt(pre, 10, 64)
	if err != nil {
		return time.Duration(0), fmt.Errorf("parse integer part: %w", err)
	}
	if secs > math.MaxInt64/nsecInSec || secs < -math.MaxInt64/nsecInSec {
		return time.Duration(0), fmt.Errorf("out of range")
	}
	secs *= nsecInSec
	for i := range len(suf) {
		if !('0' <= suf[i] && suf[i] <= '9') {
			return time.Duration(0), fmt.Errorf("fractional part contains non-digits")
		}
	}
	if len(suf) > 9 {
		suf = suf[:9]
	}
	mul := int64(1)
	for range 9 - len(suf) {
		mul *= 10
	}
	var nsecs int64
	if len(suf) > 0 {
		nsecs, err = strconv.ParseInt(suf, 10, 64)
		if err != nil {
			return time.Duration(0), fmt.Errorf("parse fractional part: %w", err)
		}
	}
	nsecs *= mul
	if secs < 0 {
		nsecs = secs - nsecs
		if nsecs > 0 {
			return time.Duration(0), fmt.Errorf("out of range")
		}
	} else {
		nsecs += secs
		if nsecs < 0 {
			return time.Duration(0), fmt.Errorf("out of range")
		}
	}
	return time.Duration(nsecs) * time.Nanosecond, nil
}
