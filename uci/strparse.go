package uci

import (
	"fmt"
	"strings"
)

type coderOptions struct {
	SanitizeUTF8                bool
	AllowBadSubstringsInOptions bool
}

func caseFold(s string) string {
	var b strings.Builder
	for i := range len(s) {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c = c - 'A' + 'a'
		}
		_ = b.WriteByte(c)
	}
	return b.String()
}

func isSpace(b byte) bool {
	const mask uint64 = (1 << '\t') | (1 << '\n') | (1 << '\v') | (1 << '\f') | (1 << '\r') | (1 << ' ')
	return b <= ' ' && ((uint64(1)<<b)&mask) != 0
}

func isGoodUntrimmedString(s string, o coderOptions) bool {
	for i := range len(s) {
		b := s[i]
		if b == '\n' || b == '\r' {
			return false
		}
		if isSpace(b) {
			continue
		}
		if b < 0x20 || b == 0x7f {
			return false
		}
		if o.SanitizeUTF8 && b >= 0x80 {
			return false
		}
	}
	return true
}

func isGoodString(s string, o coderOptions) bool {
	if len(s) == 0 {
		return true
	}
	if !isGoodUntrimmedString(s, o) {
		return false
	}
	return !isSpace(s[0]) && !isSpace(s[len(s)-1])
}

type subrange struct {
	l int
	r int
}

type tokenizer struct {
	s      string
	tokens []subrange
	pos    int
}

func newTokenizer(s string, o coderOptions) (*tokenizer, error) {
	if !isGoodUntrimmedString(s, o) {
		return nil, fmt.Errorf("bad string")
	}
	var tokens []subrange
	r := 0
	for r < len(s) && isSpace(s[r]) {
		r++
	}
	for r < len(s) {
		l := r
		for r < len(s) && !isSpace(s[r]) {
			r++
		}
		tokens = append(tokens, subrange{l: l, r: r})
		for r < len(s) && isSpace(s[r]) {
			r++
		}
	}
	return &tokenizer{
		s:      s,
		tokens: tokens,
		pos:    0,
	}, nil
}

func (t *tokenizer) Undo() bool {
	if t.pos == 0 {
		return false
	}
	t.pos--
	return true
}

func (t *tokenizer) Next() (string, bool) {
	if t.pos >= len(t.tokens) {
		return "", false
	}
	sub := t.tokens[t.pos]
	t.pos++
	return t.s[sub.l:sub.r], true
}

func (t *tokenizer) More() bool {
	return t.pos < len(t.tokens)
}

func (t *tokenizer) NextUntil(stop func(s string) bool) string {
	if t.pos >= len(t.tokens) {
		return ""
	}
	l := t.tokens[t.pos].l
	r := l
	for t.pos < len(t.tokens) {
		sub := t.tokens[t.pos]
		if stop(t.s[sub.l:sub.r]) {
			break
		}
		r = sub.r
		t.pos++
	}
	if l == r {
		return ""
	}
	return t.s[l:r]
}

func (t *tokenizer) NextUntilEnd() string {
	return t.NextUntil(func(_ string) bool { return false })
}
