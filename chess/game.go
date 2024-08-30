package chess

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

type RepeatTable struct {
	mp map[RawBoard]int
}

func NewRepeatTable() *RepeatTable {
	return &RepeatTable{
		mp: make(map[RawBoard]int),
	}
}

func moldRepeatTableKey(b *Board) RawBoard {
	raw := b.r
	raw.MoveCounter = 0
	raw.MoveNumber = 0
	return raw
}

func (r *RepeatTable) Push(b *Board) {
	k := moldRepeatTableKey(b)
	r.mp[k]++
}

func (r *RepeatTable) Pop(b *Board) {
	k := moldRepeatTableKey(b)
	val := r.mp[k] - 1
	if val == 0 {
		delete(r.mp, k)
	} else {
		r.mp[k] = val
	}
}

func (r *RepeatTable) Count(b *Board) int {
	k := moldRepeatTableKey(b)
	return r.mp[k]
}

func (r *RepeatTable) Clone() *RepeatTable {
	return &RepeatTable{
		mp: maps.Clone(r.mp),
	}
}

type MoveNumberStyle struct {
	Enabled         bool
	Custom          bool
	CustomStartFrom int
}

type GameOutcomeStyle uint8

const (
	GameOutcomeHide GameOutcomeStyle = iota
	GameOutcomeFinishedOnly
	GameOutcomeShow
)

type GameStyle struct {
	Move       MoveStyle
	MoveNumber MoveNumberStyle
	Outcome    GameOutcomeStyle
}

type GameAnnotations struct {
	Comments [][]string
}

type Game struct {
	start   RawBoard
	board   *Board
	repeat  *RepeatTable
	stack   []Undo
	outcome Outcome
}

func (g *Game) Clone() *Game {
	return &Game{
		start:   g.start,
		board:   g.board.Clone(),
		repeat:  g.repeat.Clone(),
		stack:   slices.Clone(g.stack),
		outcome: g.outcome,
	}
}

func NewGameWithPosition(b *Board) *Game {
	g := &Game{
		start:   b.r,
		board:   b.Clone(),
		repeat:  NewRepeatTable(),
		stack:   nil,
		outcome: RunningOutcome(),
	}
	g.repeat.Push(g.board)
	return g
}

func NewGame() *Game {
	return NewGameWithPosition(InitialBoard())
}

func GameFromUCIList(b *Board, ucis string) (*Game, error) {
	g := NewGameWithPosition(b)
	if _, err := g.PushUCIList(ucis); err != nil {
		return nil, err
	}
	return g, nil
}

func NewGameWithFEN(fen string) (*Game, error) {
	b, err := BoardFromFEN(fen)
	if err != nil {
		return nil, fmt.Errorf("parse fen: %w", err)
	}
	return NewGameWithPosition(b), nil
}

func (g *Game) StartPos() RawBoard {
	return g.start
}

func (g *Game) CurPos() RawBoard {
	return g.board.r
}

// The board updates automatically after calls to Push...() or Pop().
//
// Do not mutate the returned board.
func (g *Game) CurBoard() *Board {
	return g.board
}

func (g *Game) Len() int {
	return len(g.stack)
}

func (g *Game) IsEmpty() bool {
	return len(g.stack) == 0
}

func (g *Game) MoveAt(index int) Move {
	return g.stack[index].Move()
}

func (g *Game) Outcome() Outcome {
	return g.outcome
}

func (g *Game) IsFinished() bool {
	return g.outcome.IsFinished()
}

func (g *Game) ClearOutcome() {
	g.outcome = RunningOutcome()
}

func (g *Game) SetOutcome(o Outcome) {
	g.outcome = o
}

func (g *Game) CalcOutcome() Outcome {
	outcome := g.board.CalcOutcome()
	if outcome.IsFinished() && outcome.Verdict().Passes(VerdictFilterStrict) {
		return outcome
	}
	rep := g.repeat.Count(g.board)
	if rep >= 5 {
		return MustDrawOutcome(VerdictRepeat5)
	}
	if rep >= 3 {
		return MustDrawOutcome(VerdictRepeat3)
	}
	return outcome
}

func (g *Game) SetAutoOutcome(filter VerdictFilter) Outcome {
	if !g.outcome.IsFinished() {
		outcome := g.CalcOutcome()
		if outcome.IsFinished() && outcome.Passes(filter) {
			g.outcome = outcome
		}
	}
	return g.outcome
}

func (g *Game) doFinishPush(u Undo) {
	g.repeat.Push(g.board)
	g.stack = append(g.stack, u)
}

func (g *Game) PushLegalMove(mv Move) {
	u := g.board.MakeLegalMove(mv)
	g.doFinishPush(u)
}

func (g *Game) PushSemilegalMove(mv Move) bool {
	u, ok := g.board.MakeSemilegalMove(mv)
	if !ok {
		return false
	}
	g.doFinishPush(u)
	return true
}

func (g *Game) PushMove(mv Move) error {
	u, err := g.board.MakeMove(mv)
	if err != nil {
		return err
	}
	g.doFinishPush(u)
	return nil
}

func (g *Game) PushUCIMove(um UCIMove) error {
	u, err := g.board.MakeUCIMove(um)
	if err != nil {
		return err
	}
	g.doFinishPush(u)
	return nil
}

func (g *Game) PushMoveUCI(s string) error {
	u, err := g.board.MakeMoveUCI(s)
	if err != nil {
		return err
	}
	g.doFinishPush(u)
	return nil
}

func (g *Game) PushMoveSAN(s string) error {
	u, err := g.board.MakeMoveSAN(s)
	if err != nil {
		return err
	}
	g.doFinishPush(u)
	return nil
}

func (g *Game) PushUCIList(ucis string) (int, error) {
	count := 0
	for _, u := range strings.Fields(ucis) {
		err := g.PushMoveUCI(u)
		if err != nil {
			return count, fmt.Errorf("push uci move #%v: %w", count+1, err)
		}
		count++
	}
	return count, nil
}

func (g *Game) Pop() (Move, bool) {
	if len(g.stack) == 0 {
		return Move{}, false
	}
	u := g.stack[len(g.stack)-1]
	g.stack = g.stack[:len(g.stack)-1]
	g.repeat.Pop(g.board)
	g.ClearOutcome()
	g.board.UnmakeMove(u)
	return u.Move(), true
}

func (g *Game) UCIList() string {
	var b strings.Builder
	for i, u := range g.stack {
		if i != 0 {
			_ = b.WriteByte(' ')
		}
		_, _ = b.WriteString(u.mv.UCI())
	}
	return b.String()
}

func (g *Game) Walk() Walker {
	return Walker{
		board: g.board.Clone(),
		stack: g.stack,
		pos:   len(g.stack),
	}
}

func (g *Game) Styled(style GameStyle) (string, error) {
	return g.StyledExt(style, GameAnnotations{})
}

func doAddComments(b *strings.Builder, cs []string) {
	for i, c := range cs {
		if i != 0 {
			_ = b.WriteByte(' ')
		}
		_, _ = fmt.Fprintf(b, "{%v}", strings.ReplaceAll(c, "}", ""))
	}
}

func (g *Game) StyledExt(style GameStyle, ga GameAnnotations) (string, error) {
	var b strings.Builder

	first := true
	if len(ga.Comments) > 0 && len(ga.Comments[0]) != 0 {
		if !first {
			_ = b.WriteByte(' ')
		}
		first = false
		doAddComments(&b, ga.Comments[0])
	}

	if len(g.stack) != 0 {
		w := g.Walk()
		w.First()
		var moveNumber int
		if style.MoveNumber.Custom {
			moveNumber = style.MoveNumber.CustomStartFrom
		} else {
			moveNumber = int(w.Board().MoveNumber())
		}
		mustNumber := true
		for i, u := range g.stack {
			if !first {
				_ = b.WriteByte(' ')
			}
			first = false
			if style.MoveNumber.Enabled {
				if w.Board().Side() == ColorWhite {
					_, _ = fmt.Fprintf(&b, "%v. ", moveNumber)
				} else if mustNumber {
					_, _ = fmt.Fprintf(&b, "%v... ", moveNumber)
				}
			}
			mustNumber = false
			s, err := u.Move().Styled(w.Board(), style.Move)
			if err != nil {
				return "", fmt.Errorf("style move #%d: %w", i+1, err)
			}
			_, _ = b.WriteString(s)
			if len(ga.Comments) > i+1 && len(ga.Comments[i+1]) != 0 {
				_ = b.WriteByte(' ')
				doAddComments(&b, ga.Comments[i+1])
				mustNumber = true
			}
			if w.Board().Side() == ColorBlack {
				moveNumber++
			}
			_ = w.Next()
		}
	}

	if style.Outcome == GameOutcomeShow ||
		(style.Outcome == GameOutcomeFinishedOnly && g.outcome.IsFinished()) {
		if !first {
			_ = b.WriteByte(' ')
		}
		first = false // nolint:ineffassign
		_, _ = b.WriteString(g.outcome.Status().String())
	}

	return b.String(), nil
}

func (g *Game) Eq(o *Game) bool {
	return g.start == o.start &&
		g.outcome == o.outcome &&
		slices.EqualFunc(g.stack, o.stack, func(a, b Undo) bool {
			return a.Move() == b.Move()
		})
}

type Walker struct {
	board *Board
	stack []Undo
	pos   int
}

func (w *Walker) Len() int {
	return len(w.stack)
}

func (w *Walker) IsEmpty() bool {
	return len(w.stack) == 0
}

func (w *Walker) Pos() int {
	return w.pos
}

// The board updates automatically after call to Next(), Prev(), First(), Last() or Jump().
//
// Do not mutate the returned board.
func (w *Walker) Board() *Board {
	return w.board
}

func (w *Walker) doPrev() {
	w.pos--
	w.board.UnmakeMove(w.stack[w.pos])
}

func (w *Walker) doNext() {
	_ = w.board.MakeLegalMove(w.stack[w.pos].Move())
	w.pos++
}

func (w *Walker) doJump(pos int) {
	for w.pos > pos {
		w.doPrev()
	}
	for w.pos < pos {
		w.doNext()
	}
}

func (w *Walker) Next() bool {
	if w.pos == len(w.stack) {
		return false
	}
	w.doNext()
	return true
}

func (w *Walker) Prev() bool {
	if w.pos == 0 {
		return false
	}
	w.doPrev()
	return true
}

func (w *Walker) Jump(pos int) bool {
	if w.pos < 0 || w.pos > len(w.stack) {
		return false
	}
	w.doJump(pos)
	return true
}

func (w *Walker) First() {
	w.doJump(0)
}

func (w *Walker) Last() {
	w.doJump(len(w.stack))
}
