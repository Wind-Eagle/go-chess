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

type UCIMove struct {
	Kind    UCIMoveKind
	Src     Coord
	Dst     Coord
	Promote Piece
}

func (m UCIMove) ToMove(b *Board) (Move, error) {
	switch m.Kind {
	case UCIMoveNull:
		return NullMove(), nil
	case UCIMoveSimple, UCIMovePromote:
		if !m.Src.IsValid() {
			return Move{}, fmt.Errorf("invalid uci move src")
		}
		if !m.Dst.IsValid() {
			return Move{}, fmt.Errorf("invalid uci move dst")
		}

		srcCell := b.Get(m.Src)
		c := b.r.Side
		if !srcCell.HasColor(c) {
			return Move{}, fmt.Errorf("bad uci move src")
		}

		var kind MoveKind
		if m.Kind == UCIMovePromote {
			if !m.Promote.IsValid() {
				return Move{}, fmt.Errorf("invalid promote piece")
			}
			var ok bool
			kind, ok = MoveKindFromPromote(m.Promote)
			if !ok {
				return Move{}, fmt.Errorf("bad promote piece %v", m.Promote)
			}
		} else {
			piece, ok := srcCell.Piece()
			if !ok {
				panic("must not happen")
			}
			if piece == PiecePawn {
				if m.Src.Rank() == pawnHomeRank(c) && m.Dst.Rank() == pawnDoubleDstRank(c) {
					kind = MovePawnDouble
				} else if m.Src.File() != m.Dst.File() && b.Get(m.Dst).IsFree() {
					kind = MoveEnpassant
				} else {
					kind = MoveSimple
				}
			} else if piece == PieceKing {
				rank := homeRank(c)
				if m.Src == CoordFromParts(FileE, rank) {
					if m.Dst == CoordFromParts(FileC, rank) {
						kind = MoveCastlingQueenside
					} else if m.Dst == CoordFromParts(FileG, rank) {
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

		res := NewMoveUnchecked(kind, srcCell, m.Src, m.Dst)
		if !res.IsWellFormed() {
			return Move{}, errMoveNotWellFormed
		}
		return res, nil
	default:
		return Move{}, fmt.Errorf("bad uci move kind")
	}
}

func (m UCIMove) String() string {
	switch m.Kind {
	case UCIMoveNull:
		return "0000"
	case UCIMoveSimple, UCIMovePromote:
		var b strings.Builder
		_, _ = b.WriteString(m.Src.String())
		_, _ = b.WriteString(m.Dst.String())
		if m.Kind == UCIMovePromote {
			_ = b.WriteByte(m.Promote.ToByte())
		}
		return b.String()
	default:
		return "????"
	}
}

func UCIMoveFromString(s string) (UCIMove, error) {
	if s == "0000" {
		return UCIMove{Kind: UCIMoveNull}, nil
	}
	if len(s) != 4 && len(s) != 5 {
		return UCIMove{}, fmt.Errorf("bad string length")
	}
	m := UCIMove{Kind: UCIMoveSimple}
	var err error
	m.Src, err = CoordFromString(s[0:2])
	if err != nil {
		return UCIMove{}, fmt.Errorf("bad src: %w", err)
	}
	m.Dst, err = CoordFromString(s[2:4])
	if err != nil {
		return UCIMove{}, fmt.Errorf("bad dst: %w", err)
	}
	if len(s) == 5 {
		m.Kind = UCIMovePromote
		m.Promote, err = PieceFromByte(s[4])
		if err != nil {
			return UCIMove{}, fmt.Errorf("bad promote: %w", err)
		}
	}
	return m, nil
}
