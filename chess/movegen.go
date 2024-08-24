package chess

import (
	"github.com/alex65536/go-chess/util/maybe"
)

func (b *Board) IsCellAttacked(coord Coord, c Color) bool {
	// Near attacks
	if !(b.BbPiece(c, PiecePawn) & pawnAttacks(c.Inv(), coord)).IsEmpty() ||
		!(b.BbPiece(c, PieceKing) & kingAttacks(coord)).IsEmpty() ||
		!(b.BbPiece(c, PieceKnight) & knightAttacks(coord)).IsEmpty() {
		return true
	}

	// Far attacks
	return !(b.bbPieceDiag(c) & bishopAttacks(coord, b.bbAll)).IsEmpty() ||
		!(b.bbPieceLine(c) & rookAttacks(coord, b.bbAll)).IsEmpty()
}

func (b *Board) CellAttackers(coord Coord, c Color) Bitboard {
	return (b.BbPiece(c, PiecePawn) & pawnAttacks(c.Inv(), coord)) |
		(b.BbPiece(c, PieceKing) & kingAttacks(coord)) |
		(b.BbPiece(c, PieceKnight) & knightAttacks(coord)) |
		(b.bbPieceDiag(c) & bishopAttacks(coord, b.bbAll)) |
		(b.bbPieceLine(c) & rookAttacks(coord, b.bbAll))
}

type genMode uint8

const (
	genModeSimple genMode = 1 << iota
	genModeCapture
	genModeSimplePromote
	genModeCastling
	genModeMax

	genModeAll = genModeMax - 1
)

func doGenMoves[F ~func(m Move) bool](b *Board, mode genMode, f F) bool {
	c := b.r.Side

	// We generate moves in this order in hope to speed up HasLegalMoves().

	if submode := mode & (genModeSimple | genModeCapture); submode != 0 {
		var bbDstMsk Bitboard
		switch submode {
		case genModeSimple:
			bbDstMsk = ^b.bbAll
		case genModeCapture:
			bbDstMsk = b.bbColor[c.Inv()]
		case genModeSimple | genModeCapture:
			bbDstMsk = ^b.bbColor[c]
		default:
			panic("must not happen")
		}

		// King
		{
			king := CellFromParts(c, PieceKing)
			bbSrc := b.BbCell(king)
			for !bbSrc.IsEmpty() {
				src := bbSrc.Next()
				bbDst := kingAttacks(src) & bbDstMsk
				for !bbDst.IsEmpty() {
					dst := bbDst.Next()
					if !f(NewMoveUnchecked(MoveSimple, king, src, dst)) {
						return false
					}
				}
			}
		}

		// Queen
		{
			queen := CellFromParts(c, PieceQueen)
			bbSrc := b.BbCell(queen)
			for !bbSrc.IsEmpty() {
				src := bbSrc.Next()
				bbDst := (rookAttacks(src, b.bbAll) | bishopAttacks(src, b.bbAll)) & bbDstMsk
				for !bbDst.IsEmpty() {
					dst := bbDst.Next()
					if !f(NewMoveUnchecked(MoveSimple, queen, src, dst)) {
						return false
					}
				}
			}
		}

		// Rook
		{
			rook := CellFromParts(c, PieceRook)
			bbSrc := b.BbCell(rook)
			for !bbSrc.IsEmpty() {
				src := bbSrc.Next()
				bbDst := rookAttacks(src, b.bbAll) & bbDstMsk
				for !bbDst.IsEmpty() {
					dst := bbDst.Next()
					if !f(NewMoveUnchecked(MoveSimple, rook, src, dst)) {
						return false
					}
				}
			}
		}

		// Bishop
		{
			bishop := CellFromParts(c, PieceBishop)
			bbSrc := b.BbCell(bishop)
			for !bbSrc.IsEmpty() {
				src := bbSrc.Next()
				bbDst := bishopAttacks(src, b.bbAll) & bbDstMsk
				for !bbDst.IsEmpty() {
					dst := bbDst.Next()
					if !f(NewMoveUnchecked(MoveSimple, bishop, src, dst)) {
						return false
					}
				}
			}
		}

		// Knight
		{
			knight := CellFromParts(c, PieceKnight)
			bbSrc := b.BbCell(knight)
			for !bbSrc.IsEmpty() {
				src := bbSrc.Next()
				bbDst := knightAttacks(src) & bbDstMsk
				for !bbDst.IsEmpty() {
					dst := bbDst.Next()
					if !f(NewMoveUnchecked(MoveSimple, knight, src, dst)) {
						return false
					}
				}
			}
		}
	}

	// Pawn
	{
		pawn := CellFromParts(c, PiecePawn)
		bbPawn := b.BbCell(pawn)
		bbPromote := BbRank(promoteSrcRank(c))

		addPromote := func(src, dst Coord) bool {
			return f(NewMoveUnchecked(MovePromoteKnight, pawn, src, dst)) &&
				f(NewMoveUnchecked(MovePromoteBishop, pawn, src, dst)) &&
				f(NewMoveUnchecked(MovePromoteRook, pawn, src, dst)) &&
				f(NewMoveUnchecked(MovePromoteQueen, pawn, src, dst))
		}

		if (mode & (genModeSimple | genModeSimplePromote)) != 0 {
			bbDouble := BbRank(pawnHomeRank(c))
			if (mode & genModeSimple) != 0 {
				// Simple move
				bb := pawnAdvanceForward(c, bbPawn & ^bbPromote) & ^b.bbAll
				for !bb.IsEmpty() {
					dst := bb.Next()
					if !f(NewMoveUnchecked(MoveSimple, pawn, dst.Add(-pawnForwardDelta(c)), dst)) {
						return false
					}
				}

				// Double move
				bbTmp := pawnAdvanceForward(c, bbPawn&bbDouble) & ^b.bbAll
				bb = pawnAdvanceForward(c, bbTmp) & ^b.bbAll
				for !bb.IsEmpty() {
					dst := bb.Next()
					src := dst.Add(-2 * pawnForwardDelta(c))
					if !f(NewMoveUnchecked(MovePawnDouble, pawn, src, dst)) {
						return false
					}
				}
			}
			if (mode & genModeSimplePromote) != 0 {
				// Simple promote
				bb := pawnAdvanceForward(c, bbPawn&bbPromote) & ^b.bbAll
				for !bb.IsEmpty() {
					dst := bb.Next()
					if !addPromote(dst.Add(-pawnForwardDelta(c)), dst) {
						return false
					}
				}
			}
		}
		if (mode & genModeCapture) != 0 {
			bbAllow := b.bbColor[c.Inv()]
			dl, dr := -pawnLeftDelta(c), -pawnRightDelta(c)

			// Capture
			bbPawnMasked := bbPawn & ^bbPromote
			bb := pawnAdvanceLeft(c, bbPawnMasked) & bbAllow
			for !bb.IsEmpty() {
				dst := bb.Next()
				if !f(NewMoveUnchecked(MoveSimple, pawn, dst.Add(dl), dst)) {
					return false
				}
			}
			bb = pawnAdvanceRight(c, bbPawnMasked) & bbAllow
			for !bb.IsEmpty() {
				dst := bb.Next()
				if !f(NewMoveUnchecked(MoveSimple, pawn, dst.Add(dr), dst)) {
					return false
				}
			}

			// Capture promote
			bbPawnMasked = bbPawn & bbPromote
			bb = pawnAdvanceLeft(c, bbPawnMasked) & bbAllow
			for !bb.IsEmpty() {
				dst := bb.Next()
				if !addPromote(dst.Add(dl), dst) {
					return false
				}
			}
			bb = pawnAdvanceRight(c, bbPawnMasked) & bbAllow
			for !bb.IsEmpty() {
				dst := bb.Next()
				if !addPromote(dst.Add(dr), dst) {
					return false
				}
			}

			// Enpassant
			if ep, ok := b.r.EpSource.TryGet(); ok {
				file := ep.File()
				dst := ep.Add(pawnForwardDelta(c))
				// We assume that the cell behind the pawn that made double move is empty, so don't check it
				lp, rp := ep.Add(-1), ep.Add(1)
				if file != FileA && b.Get(lp) == pawn &&
					!f(NewMoveUnchecked(MoveEnpassant, pawn, lp, dst)) {
					return false
				}
				if file != FileH && b.Get(rp) == pawn &&
					!f(NewMoveUnchecked(MoveEnpassant, pawn, rp, dst)) {
					return false
				}
			}
		}
	}

	// Castling
	if (mode&genModeCastling) != 0 && b.r.Castling.HasColor(c) {
		rank := homeRank(c)
		inv := c.Inv()
		src := CoordFromParts(FileE, rank)
		king := CellFromParts(c, PieceKing)

		if b.r.Castling.Has(c, CastlingQueenside) {
			tmp, dst := CoordFromParts(FileD, rank), CoordFromParts(FileC, rank)
			if (bbCastlingPass(c, CastlingQueenside) & b.bbAll).IsEmpty() &&
				!b.IsCellAttacked(src, inv) &&
				!b.IsCellAttacked(tmp, inv) &&
				!f(NewMoveUnchecked(MoveCastlingQueenside, king, src, dst)) {
				return false
			}
		}

		if b.r.Castling.Has(c, CastlingKingside) {
			tmp, dst := CoordFromParts(FileF, rank), CoordFromParts(FileG, rank)
			if (bbCastlingPass(c, CastlingKingside) & b.bbAll).IsEmpty() &&
				!b.IsCellAttacked(src, inv) &&
				!b.IsCellAttacked(tmp, inv) &&
				!f(NewMoveUnchecked(MoveCastlingKingside, king, src, dst)) {
				return false
			}
		}
	}

	return true
}

func (b *Board) sanCandidates(piece Piece, dst Coord, res []Move) []Move {
	c := b.r.Side

	if b.Get(dst).HasColor(c) {
		return res
	}

	var bbMask Bitboard
	switch piece {
	case PiecePawn:
		panic("pawns are not supported here")
	case PieceKing:
		bbMask = kingAttacks(dst)
	case PieceKnight:
		bbMask = knightAttacks(dst)
	case PieceBishop:
		bbMask = bishopAttacks(dst, b.bbAll)
	case PieceRook:
		bbMask = rookAttacks(dst, b.bbAll)
	case PieceQueen:
		bbMask = bishopAttacks(dst, b.bbAll) | rookAttacks(dst, b.bbAll)
	default:
		panic("must not happen")
	}

	cell := CellFromParts(c, piece)
	bb := bbMask & b.BbCell(cell)
	for !bb.IsEmpty() {
		src := bb.Next()
		res = append(res, NewMoveUnchecked(MoveSimple, cell, src, dst))
	}

	return filterLegalMoves(b, res)
}

func (b *Board) sanPawnCaptureCandidates(src, dst File, promote maybe.Maybe[Piece], res []Move) []Move {
	c := b.r.Side
	bbPromote := BbRank(promoteSrcRank(c))
	var (
		bbMask Bitboard
		kind   MoveKind
	)
	if p, ok := promote.TryGet(); ok {
		bbMask = bbPromote
		var ok bool
		kind, ok = MoveKindFromPromote(p)
		if !ok {
			panic("piece is not supported for promote")
		}
	} else {
		bbMask = ^bbPromote
		kind = MoveSimple
	}

	pawn := CellFromParts(c, PiecePawn)
	bbPawns := b.BbCell(pawn) & bbMask & BbFile(src)
	bbAllow := b.bbColor[c.Inv()]

	if src == dst+1 {
		dl := -pawnLeftDelta(c)
		bb := pawnAdvanceLeft(c, bbPawns) & bbAllow
		for !bb.IsEmpty() {
			dst := bb.Next()
			res = append(res, NewMoveUnchecked(kind, pawn, dst.Add(dl), dst))
		}
	}

	if dst == src+1 {
		dr := -pawnRightDelta(c)
		bb := pawnAdvanceRight(c, bbPawns) & bbAllow
		for !bb.IsEmpty() {
			dst := bb.Next()
			res = append(res, NewMoveUnchecked(kind, pawn, dst.Add(dr), dst))
		}
	}

	if ep, ok := b.r.EpSource.TryGet(); ok && promote.IsNone() && ep.File() == dst {
		dstCoord := ep.Add(pawnForwardDelta(c))
		// We assume that the cell behind the pawn that made double move is empty, so don't check it
		lp, rp := ep.Add(-1), ep.Add(1)
		if src+1 == dst && b.Get(lp) == pawn {
			res = append(res, NewMoveUnchecked(MoveEnpassant, pawn, lp, dstCoord))
		}
		if src == dst+1 && b.Get(rp) == pawn {
			res = append(res, NewMoveUnchecked(MoveEnpassant, pawn, rp, dstCoord))
		}
	}

	return filterLegalMoves(b, res)
}

type MoveGenPreset uint8

const (
	MoveGenAll MoveGenPreset = iota
	MoveGenCapture
	MoveGenSimple
	MoveGenSimpleNoPromote
	MoveGenSimplePromote
	MoveGenPresetMax
)

var moveGenPresets = []genMode{
	MoveGenAll:             genModeAll,
	MoveGenCapture:         genModeCapture,
	MoveGenSimple:          genModeSimple | genModeSimplePromote | genModeCastling,
	MoveGenSimpleNoPromote: genModeSimple | genModeCastling,
	MoveGenSimplePromote:   genModeSimplePromote,
}

func (b *Board) GenSemilegalMoves(preset MoveGenPreset, res []Move) []Move {
	doGenMoves(b, moveGenPresets[preset], func(m Move) bool {
		res = append(res, m)
		return true
	})
	return res
}

func (b *Board) GenLegalMoves(preset MoveGenPreset, res []Move) []Move {
	res = b.GenSemilegalMoves(preset, res)
	return filterLegalMoves(b, res)
}

func filterLegalMoves(b *Board, res []Move) []Move {
	s := newIsLegalState(b)
	cnt := 0
	for i := range res {
		if doIsLegal(b, s, res[i]) {
			res[i], res[cnt] = res[cnt], res[i]
			cnt++
		}
	}
	return res[:cnt]
}

func (b *Board) HasLegalMoves() bool {
	s := newIsLegalState(b)
	ok := false
	doGenMoves(b, genModeAll, func(m Move) bool {
		if !doIsLegal(b, s, m) {
			return true
		}
		ok = true
		return false
	})
	return ok
}
