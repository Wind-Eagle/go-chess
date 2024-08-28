package uci

import (
	"cmp"
	"fmt"
	"math"
)

type ScoreBound int8

const (
	ScoreExact ScoreBound = iota
	ScoreLower
	ScoreUpper
)

type Score struct {
	mate bool
	val  int32
}

func (s Score) String() string {
	if s.mate {
		return fmt.Sprintf("#%v", s.val)
	} else {
		if s.val > 0 {
			return fmt.Sprintf("+%v.%02d", s.val/100, s.val%100)
		} else if s.val == 0 {
			return "0.00"
		} else {
			return fmt.Sprintf("-%v.%02d", (-s.val)/100, (-s.val)%100)
		}
	}
}

func (s Score) IsMate() bool {
	return s.mate
}

func (s Score) IsWinMate() bool {
	return s.mate && s.val > 0
}

func (s Score) IsLoseMate() bool {
	return s.mate && s.val <= 0
}

func (s Score) Mate() (int32, bool) {
	if s.mate {
		return s.val, true
	}
	return 0, false
}

func (s Score) Value() (float64, bool) {
	if s.mate {
		if s.val > 0 {
			return math.Inf(+1), false
		} else {
			return math.Inf(-1), false
		}
	}
	return float64(s.val) / 100.0, true
}

func (s Score) Centipawns() (int32, bool) {
	if s.mate {
		if s.val > 0 {
			return math.MaxInt32, false
		} else {
			return math.MinInt32, false
		}
	}
	return s.val, true
}

func (s Score) cmpKind() int {
	if s.mate {
		if s.val > 0 {
			return +1
		} else {
			return -1
		}
	}
	return 0
}

func (s Score) Compare(t Score) int {
	sc, tc := s.cmpKind(), t.cmpKind()
	if sc != tc {
		return cmp.Compare(sc, tc)
	}
	if sc == 0 {
		return cmp.Compare(s.val, t.val)
	} else {
		return cmp.Compare(t.val, s.val)
	}
}

func (s Score) Less(t Score) bool {
	return s.Compare(t) < 0
}

func (s Score) Greater(t Score) bool {
	return s.Compare(t) > 0
}

func (s Score) Backtrack() Score {
	if s.mate {
		if s.val > 0 {
			return Score{mate: true, val: -s.val}
		} else {
			return Score{mate: true, val: -s.val - 1}
		}
	}
	return Score{mate: false, val: -s.val}
}

func ScoreMate(in int32) Score {
	return Score{mate: true, val: in}
}

func ScoreCentipawns(val int32) Score {
	return Score{mate: false, val: val}
}

type BoundedScore struct {
	Score Score
	Bound ScoreBound
}
