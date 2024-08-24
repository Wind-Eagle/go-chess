package chess

import (
	"fmt"
	"strings"
)

type UCIMoveKind uint8

const (
	UCIMoveNull UCIMoveKind = iota
	UCIMoveSimple
	UCIMovePromote
	UCIMoveKindMax
)

func (k UCIMoveKind) IsValid() bool {
	return k < UCIMoveKindMax
}

type UCIMove struct {
	kind    UCIMoveKind
	src     Coord
	dst     Coord
	promote Piece
}

func NullUCIMove() UCIMove {
	return UCIMove{kind: UCIMoveNull}
}

func SimpleUCIMove(src, dst Coord) UCIMove {
	return UCIMove{kind: UCIMoveSimple, src: src, dst: dst}
}

func PromoteUCIMove(src, dst Coord, promote Piece) UCIMove {
	return UCIMove{kind: UCIMovePromote, src: src, dst: dst, promote: promote}
}

func (m UCIMove) Kind() UCIMoveKind      { return m.kind }
func (m UCIMove) Src() Coord             { return m.src }
func (m UCIMove) Dst() Coord             { return m.dst }
func (m UCIMove) Promote() (Piece, bool) { return m.promote, m.kind == UCIMovePromote }

func (m UCIMove) ToMove(b *Board) (Move, error) {
	switch m.kind {
	case UCIMoveNull:
		return NullMove(), nil
	case UCIMoveSimple, UCIMovePromote:
		if !m.src.IsValid() {
			return Move{}, fmt.Errorf("invalid uci move src")
		}
		if !m.dst.IsValid() {
			return Move{}, fmt.Errorf("invalid uci move dst")
		}

		srcCell := b.Get(m.src)
		c := b.r.Side
		if !srcCell.HasColor(c) {
			return Move{}, fmt.Errorf("bad uci move src")
		}

		var kind MoveKind
		if m.kind == UCIMovePromote {
			if !m.promote.IsValid() {
				return Move{}, fmt.Errorf("invalid promote piece")
			}
			var ok bool
			kind, ok = MoveKindFromPromote(m.promote)
			if !ok {
				return Move{}, fmt.Errorf("bad promote piece %v", m.promote)
			}
		} else {
			piece, ok := srcCell.Piece()
			if !ok {
				panic("must not happen")
			}
			if piece == PiecePawn {
				if m.src.Rank() == pawnHomeRank(c) && m.dst.Rank() == pawnDoubleDstRank(c) {
					kind = MovePawnDouble
				} else if m.src.File() != m.dst.File() && b.Get(m.dst).IsFree() {
					kind = MoveEnpassant
				} else {
					kind = MoveSimple
				}
			} else if piece == PieceKing {
				rank := homeRank(c)
				if m.src == CoordFromParts(FileE, rank) {
					if m.dst == CoordFromParts(FileC, rank) {
						kind = MoveCastlingQueenside
					} else if m.dst == CoordFromParts(FileG, rank) {
						kind = MoveCastlingKingside
					} else {
						kind = MoveSimple
					}
				} else {
					kind = MoveSimple
				}
			} else {
				kind = MoveSimple
			}
		}

		res := NewMoveUnchecked(kind, srcCell, m.src, m.dst)
		if !res.IsWellFormed() {
			return Move{}, errMoveNotWellFormed
		}
		return res, nil
	default:
		return Move{}, fmt.Errorf("bad uci move kind")
	}
}

func (m UCIMove) String() string {
	switch m.kind {
	case UCIMoveNull:
		return "0000"
	case UCIMoveSimple, UCIMovePromote:
		var b strings.Builder
		_, _ = b.WriteString(m.src.String())
		_, _ = b.WriteString(m.dst.String())
		if m.kind == UCIMovePromote {
			_ = b.WriteByte(m.promote.ToByte())
		}
		return b.String()
	default:
		return "????"
	}
}

func UCIMoveFromString(s string) (UCIMove, error) {
	if s == "0000" {
		return NullUCIMove(), nil
	}
	if len(s) != 4 && len(s) != 5 {
		return UCIMove{}, fmt.Errorf("bad string length")
	}
	m := UCIMove{kind: UCIMoveSimple}
	var err error
	m.src, err = CoordFromString(s[0:2])
	if err != nil {
		return UCIMove{}, fmt.Errorf("bad src: %w", err)
	}
	m.dst, err = CoordFromString(s[2:4])
	if err != nil {
		return UCIMove{}, fmt.Errorf("bad dst: %w", err)
	}
	if len(s) == 5 {
		m.kind = UCIMovePromote
		m.promote, err = PieceFromByte(s[4])
		if err != nil {
			return UCIMove{}, fmt.Errorf("bad promote: %w", err)
		}
	}
	return m, nil
}
