package chess

import (
	"fmt"
	"strings"
)

type sanStyleTable struct {
	pieces      [PieceMax]rune
	promoteSign string
}

var (
	sanStyleASCII = &sanStyleTable{
		pieces: [PieceMax]rune{
			PiecePawn:   'P',
			PieceKing:   'K',
			PieceKnight: 'N',
			PieceBishop: 'B',
			PieceRook:   'R',
			PieceQueen:  'Q',
		},
		promoteSign: "=",
	}

	sanStyleFancy = &sanStyleTable{
		pieces: [PieceMax]rune{
			PiecePawn:   '♙',
			PieceKing:   '♔',
			PieceKnight: '♘',
			PieceBishop: '♗',
			PieceRook:   '♖',
			PieceQueen:  '♕',
		},
		promoteSign: "",
	}
)

func sanResolveAmbiguity(m Move, candidates []Move) (needsFile bool, needsRank bool) {
	simAny, simFile, simRank := false, false, false
	for _, cm := range candidates {
		if cm == m {
			continue
		}
		simAny = true
		if m.src.File() == cm.src.File() {
			simFile = true
		}
		if m.src.Rank() == cm.src.Rank() {
			simRank = true
		}
	}
	needsFile = simAny && (simRank || !simFile)
	needsRank = simAny && simFile
	return
}

func sanSelectMove(file File, rank Rank, hasFile, hasRank bool, candidates []Move) (Move, error) {
	srcs := BbFull
	if hasFile {
		srcs &= BbFile(file)
	}
	if hasRank {
		srcs &= BbRank(rank)
	}
	found := false
	var m Move
	for _, cm := range candidates {
		if !srcs.Has(cm.src) {
			continue
		}
		if found {
			return Move{}, fmt.Errorf("ambiguous move: %v and %v are candidates", m, cm)
		}
		found = true
		m = cm
	}
	if !found {
		return Move{}, fmt.Errorf("no such move")
	}
	return m, nil
}

type sanCheckMark uint8

const (
	sanNoCheck sanCheckMark = iota
	sanCheck
	sanCheckmate
)

type sanMoveKind uint8

const (
	sanMoveUCI sanMoveKind = iota
	sanMoveCastling
	sanMoveSimple
)

type sanMoveFlags uint8

const (
	sanMoveIsCapture sanMoveFlags = 1 << iota
	sanMoveIsShortCapture
	sanMoveHasPromote
	sanMoveHasFile
	sanMoveHasRank
)

type sanMove struct {
	kind  sanMoveKind
	check sanCheckMark

	// sanMoveUCI
	uci UCIMove

	// sanMoveCastling
	castling CastlingSide

	// sanMoveSimple
	flags   sanMoveFlags
	piece   Piece
	dst     Coord // When sanMoveIsShortCapture, rank is always set to Rank8
	promote Piece
	file    File
	rank    Rank
}

func sanMoveFromMoveWithoutCheck(m Move, b *Board) sanMove {
	if m.kind == MoveNull {
		return sanMove{kind: sanMoveUCI, uci: UCIMove{Kind: UCIMoveNull}}
	}
	if s, ok := m.kind.CastlingSide(); ok {
		return sanMove{kind: sanMoveCastling, castling: s}
	}

	piece, ok := m.srcCell.Piece()
	if !ok {
		panic("must not happen")
	}
	san := sanMove{
		kind:  sanMoveSimple,
		flags: 0,
		piece: piece,
		dst:   m.dst,
	}
	if promote, ok := m.kind.Promote(); ok {
		san.flags |= sanMoveHasPromote
		san.promote = promote
	}

	switch m.kind {
	case MovePawnDouble, MoveEnpassant, MoveSimple,
		MovePromoteKnight, MovePromoteBishop, MovePromoteRook, MovePromoteQueen:
		isCapture := m.kind == MoveEnpassant || b.Get(m.dst).IsOccupied()
		if isCapture {
			san.flags |= sanMoveIsCapture
		}
		if piece == PiecePawn {
			if isCapture {
				san.flags |= sanMoveHasFile
				san.file = m.src.File()
			}
		} else {
			var buf [8]Move
			moves := b.sanCandidates(piece, m.dst, buf[:0])
			needsFile, needsRank := sanResolveAmbiguity(m, moves)
			if needsFile {
				san.flags |= sanMoveHasFile
				san.file = m.src.File()
			}
			if needsRank {
				san.flags |= sanMoveHasRank
				san.rank = m.src.Rank()
			}
		}
	case MoveNull, MoveCastlingQueenside, MoveCastlingKingside:
		panic("must not happen")
	default:
		panic("invalid move kind")
	}

	return san
}

func sanMoveFromMove(m Move, b *Board) (sanMove, error) {
	newB := *b
	if _, err := newB.MakeMove(m); err != nil {
		return sanMove{}, fmt.Errorf("bad move: %w", err)
	}
	san := sanMoveFromMoveWithoutCheck(m, b)
	san.check = sanNoCheck
	if newB.IsCheck() {
		if newB.HasLegalMoves() {
			san.check = sanCheck
		} else {
			san.check = sanCheckmate
		}
	}
	return san, nil
}

func sanMoveFromStringWithoutCheck(s string) (sanMove, error) {
	if s == "" {
		return sanMove{}, fmt.Errorf("empty san string")
	}
	if s == "O-O-O" || s == "0-0-0" {
		return sanMove{kind: sanMoveCastling, castling: CastlingQueenside}, nil
	}
	if s == "O-O" || s == "0-0" {
		return sanMove{kind: sanMoveCastling, castling: CastlingKingside}, nil
	}
	u, err := UCIMoveFromString(s)
	if err == nil {
		return sanMove{kind: sanMoveUCI, uci: u}, nil
	}

	san := sanMove{
		kind:  sanMoveSimple,
		flags: 0,
	}

	switch s[0] {
	case 'K', 'N', 'B', 'R', 'Q':
		switch s[0] {
		case 'K':
			san.piece = PieceKing
		case 'N':
			san.piece = PieceKnight
		case 'B':
			san.piece = PieceBishop
		case 'R':
			san.piece = PieceRook
		case 'Q':
			san.piece = PieceQueen
		default:
			panic("must not happen")
		}
		s = s[1:]
		if len(s) < 2 {
			return sanMove{}, fmt.Errorf("san move too short")
		}
		san.dst, err = CoordFromString(s[len(s)-2:])
		if err != nil {
			return sanMove{}, fmt.Errorf("bad san dst: %w", err)
		}
		s = s[:len(s)-2]
		if len(s) != 0 && 'a' <= s[0] && s[0] <= 'h' {
			san.flags |= sanMoveHasFile
			san.file, _ = FileFromByte(s[0])
			s = s[1:]
		}
		if len(s) != 0 && '1' <= s[0] && s[0] <= '8' {
			san.flags |= sanMoveHasRank
			san.rank, _ = RankFromByte(s[0])
			s = s[1:]
		}
		if len(s) != 0 && (s[0] == ':' || s[0] == 'x') {
			san.flags |= sanMoveIsCapture
			s = s[1:]
		}
		if len(s) != 0 {
			return sanMove{}, fmt.Errorf("extra data in san move")
		}
	default:
		// Pawn move
		hasPromote := true
		switch s[len(s)-1] {
		case 'N':
			san.promote = PieceKnight
		case 'B':
			san.promote = PieceBishop
		case 'R':
			san.promote = PieceRook
		case 'Q':
			san.promote = PieceQueen
		default:
			hasPromote = false
		}
		if hasPromote {
			s = s[:len(s)-1]
			if len(s) != 0 && s[len(s)-1] == '=' {
				s = s[:len(s)-1]
			}
			san.flags |= sanMoveHasPromote
		}
		if len(s) < 2 {
			return sanMove{}, fmt.Errorf("san pawn move is too short")
		}
		if len(s) == 2 && 'a' <= s[0] && s[0] <= 'h' && 'a' <= s[1] && s[1] <= 'h' {
			// Short capture
			san.flags |= sanMoveIsCapture | sanMoveIsShortCapture | sanMoveHasFile
			san.file, _ = FileFromByte(s[0])
			dstFile, _ := FileFromByte(s[1])
			san.dst = CoordFromParts(dstFile, Rank8)
		} else {
			san.dst, err = CoordFromString(s[len(s)-2:])
			if err != nil {
				return sanMove{}, fmt.Errorf("bad san pawn dst: %w", err)
			}
			s = s[:len(s)-2]
			switch len(s) {
			case 0:
				// Simple move, do nothing
			case 1:
				return sanMove{}, fmt.Errorf("bad san pawn move")
			case 2:
				if !('a' <= s[0] && s[0] <= 'h' && (s[1] == ':' || s[1] == 'x')) {
					return sanMove{}, fmt.Errorf("bad san pawn move")
				}
				san.flags |= sanMoveIsCapture | sanMoveHasFile
				san.file, _ = FileFromByte(s[0])
			default:
				return sanMove{}, fmt.Errorf("san pawn move too long")
			}
		}
	}

	return san, nil
}

func sanMoveFromString(s string) (sanMove, error) {
	check := sanNoCheck
	if strings.HasSuffix(s, "++") {
		s = s[:len(s)-2]
		check = sanCheckmate
	} else if strings.HasSuffix(s, "#") {
		s = s[:len(s)-1]
		check = sanCheckmate
	} else if strings.HasSuffix(s, "+") {
		s = s[:len(s)-1]
		check = sanCheck
	}
	m, err := sanMoveFromStringWithoutCheck(s)
	if err != nil {
		return sanMove{}, err
	}
	m.check = check
	return m, nil
}

func (m sanMove) hasFlag(f sanMoveFlags) bool {
	return (m.flags & f) != 0
}

func (m sanMove) styledWithoutCheck(tab *sanStyleTable) (string, error) {
	switch m.kind {
	case sanMoveUCI:
		return m.uci.String(), nil
	case sanMoveCastling:
		switch m.castling {
		case CastlingQueenside:
			return "O-O-O", nil
		case CastlingKingside:
			return "O-O", nil
		default:
			return "", fmt.Errorf("invalid san move")
		}
	case sanMoveSimple:
		var b strings.Builder
		switch m.piece {
		case PiecePawn:
			if m.hasFlag(sanMoveIsCapture) {
				const allowedFlags = sanMoveIsCapture | sanMoveIsShortCapture | sanMoveHasFile | sanMoveHasPromote
				if m.flags != (m.flags&allowedFlags) || !m.hasFlag(sanMoveHasFile) {
					return "", fmt.Errorf("invalid san move")
				}
				if m.hasFlag(sanMoveIsShortCapture) {
					_ = b.WriteByte(m.file.ToByte())
					_ = b.WriteByte(m.dst.File().ToByte())
				} else {
					_ = b.WriteByte(m.file.ToByte())
					_ = b.WriteByte('x')
					_, _ = b.WriteString(m.dst.String())
				}
			} else {
				const allowedFlags = sanMoveHasPromote
				if m.flags != (m.flags & allowedFlags) {
					return "", fmt.Errorf("invalid san move")
				}
				_, _ = b.WriteString(m.dst.String())
			}
			if m.hasFlag(sanMoveHasPromote) {
				if _, ok := MoveKindFromPromote(m.promote); !ok {
					return "", fmt.Errorf("invalid san move")
				}
				_, _ = b.WriteString(tab.promoteSign)
				_, _ = b.WriteRune(tab.pieces[m.promote])
			}
		case PieceKing, PieceKnight, PieceBishop, PieceRook, PieceQueen:
			const allowedFlags = sanMoveIsCapture | sanMoveHasFile | sanMoveHasRank
			if m.flags != (m.flags & allowedFlags) {
				return "", fmt.Errorf("invalid san move")
			}
			_, _ = b.WriteRune(tab.pieces[m.piece])
			if m.hasFlag(sanMoveHasFile) {
				_ = b.WriteByte(m.file.ToByte())
			}
			if m.hasFlag(sanMoveHasRank) {
				_ = b.WriteByte(m.rank.ToByte())
			}
			if m.hasFlag(sanMoveIsCapture) {
				_ = b.WriteByte('x')
			}
			_, _ = b.WriteString(m.dst.String())
		default:
			return "", fmt.Errorf("invalid san move")
		}
		return b.String(), nil
	default:
		return "", fmt.Errorf("invalid san move")
	}
}

func (m sanMove) styled(tab *sanStyleTable) (string, error) {
	s, err := m.styledWithoutCheck(tab)
	if err != nil {
		return "", err
	}
	switch m.check {
	case sanNoCheck:
		// Do nothing
	case sanCheck:
		s += "+"
	case sanCheckmate:
		s += "#"
	default:
		return "", fmt.Errorf("invalid san move")
	}
	return s, nil
}

func (m sanMove) toLegalMoveImpl(b *Board) (move Move, needValidate bool, err error) {
	switch m.kind {
	case sanMoveUCI:
		res, err := m.uci.ToMove(b)
		if err != nil {
			return Move{}, false, fmt.Errorf("parse uci move: %w", err)
		}
		return res, true, nil
	case sanMoveCastling:
		if !m.castling.IsValid() {
			return Move{}, false, fmt.Errorf("invalid san move")
		}
		return MoveFromCastling(b.r.Side, m.castling), true, nil
	case sanMoveSimple:
		var buf [8]Move
		switch m.piece {
		case PiecePawn:
			pawn := CellFromParts(b.r.Side, PiecePawn)
			kind := MoveSimple
			if m.hasFlag(sanMoveHasPromote) {
				var ok bool
				kind, ok = MoveKindFromPromote(m.promote)
				if !ok {
					return Move{}, false, fmt.Errorf("invalid san move")
				}
			}
			if m.hasFlag(sanMoveIsCapture) {
				const allowedFlags = sanMoveIsCapture | sanMoveIsShortCapture | sanMoveHasFile | sanMoveHasPromote
				if m.flags != (m.flags&allowedFlags) || !m.hasFlag(sanMoveHasFile) {
					return Move{}, false, fmt.Errorf("invalid san move")
				}
				if m.hasFlag(sanMoveIsShortCapture) {
					isPromote := m.hasFlag(sanMoveHasPromote)
					moves := b.sanPawnCaptureCandidates(m.file, m.dst.File(), isPromote, m.promote, buf[:0])
					res, err := sanSelectMove(File(0), Rank(0), false, false, moves)
					if err != nil {
						return Move{}, false, err
					}
					return res, false, nil
				} else {
					if m.dst.Rank() == homeRank(b.r.Side) {
						return Move{}, false, errMoveNotWellFormed
					}
					if ep, ok := b.r.EpDest().TryGet(); ok {
						if m.dst == ep {
							if kind != MoveSimple {
								return Move{}, false, errMoveNotWellFormed
							}
							kind = MoveEnpassant
						}
					}
					if kind != MoveEnpassant && b.Get(m.dst).IsFree() {
						return Move{}, false, fmt.Errorf("capture is expected")
					}
					src := CoordFromParts(m.file, m.dst.Rank()).Add(-pawnForwardDelta(b.r.Side))
					res, err := NewMove(kind, pawn, src, m.dst)
					if err != nil {
						return Move{}, false, errMoveNotWellFormed
					}
					return res, true, nil
				}
			} else {
				const allowedFlags = sanMoveHasPromote
				if m.flags != (m.flags & allowedFlags) {
					return Move{}, false, fmt.Errorf("invalid san move")
				}
				if m.dst.Rank() == homeRank(b.r.Side) {
					return Move{}, false, errMoveNotWellFormed
				}
				src := m.dst.Add(-pawnForwardDelta(b.r.Side))
				if !b.Get(src).IsOccupied() {
					if kind != MoveSimple {
						return Move{}, false, errMoveNotWellFormed
					}
					src = CoordFromParts(m.dst.File(), pawnHomeRank(b.r.Side))
					kind = MovePawnDouble
				}
				res, err := NewMove(kind, pawn, src, m.dst)
				if err != nil {
					return Move{}, false, errMoveNotWellFormed
				}
				return res, true, nil
			}
		case PieceKing, PieceKnight, PieceBishop, PieceRook, PieceQueen:
			const allowedFlags = sanMoveIsCapture | sanMoveHasFile | sanMoveHasRank
			if m.flags != (m.flags & allowedFlags) {
				return Move{}, false, fmt.Errorf("invalid san move")
			}
			if m.hasFlag(sanMoveIsCapture) && b.Get(m.dst).IsFree() {
				return Move{}, false, fmt.Errorf("capture is expected")
			}
			moves := b.sanCandidates(m.piece, m.dst, buf[:0])
			hasFile, hasRank := m.hasFlag(sanMoveHasFile), m.hasFlag(sanMoveHasRank)
			res, err := sanSelectMove(m.file, m.rank, hasFile, hasRank, moves)
			if err != nil {
				return Move{}, false, err
			}
			return res, false, err
		default:
			return Move{}, false, fmt.Errorf("invalid san move")
		}
	default:
		return Move{}, false, fmt.Errorf("invalid san move")
	}
}

func (m sanMove) toLegalMove(b *Board) (Move, error) {
	move, needValidate, err := m.toLegalMoveImpl(b)
	if err != nil {
		return Move{}, err
	}
	if needValidate {
		err = move.Validate(b)
		if err != nil {
			return Move{}, err
		}
	}
	return move, nil
}
