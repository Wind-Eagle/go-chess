package chess

type isLegalState struct {
	c    Color
	king Coord
}

func newIsLegalState(b *Board) isLegalState {
	c := b.r.Side
	return isLegalState{c: c, king: b.KingPos(c)}
}

func doIsCellAttackedMasked(b *Board, c Color, coord Coord, bbAll, bbMask Bitboard) bool {
	// Near attacks
	if !(b.BbPiece(c, PiecePawn) & pawnAttacks(c.Inv(), coord) & bbMask).IsEmpty() ||
		!(b.BbPiece(c, PieceKing) & kingAttacks(coord) & bbMask).IsEmpty() ||
		!(b.BbPiece(c, PieceKnight) & knightAttacks(coord) & bbMask).IsEmpty() {
		return true
	}

	// Far attacks
	return !(b.bbPieceDiag(c) & bishopAttacks(coord, bbAll) & bbMask).IsEmpty() ||
		!(b.bbPieceLine(c) & rookAttacks(coord, bbAll) & bbMask).IsEmpty()

}

func doIsLegal(b *Board, s isLegalState, mv Move) bool {
	c, inv := s.c, s.c.Inv()
	bbSrc, bbDst := BitboardFromCoord(mv.src), BitboardFromCoord(mv.dst)

	if mv.srcCell == CellFromParts(c, PieceKing) {
		return !doIsCellAttackedMasked(b, inv, mv.dst, b.bbAll^bbSrc, BbFull)
	}

	bbAll := (b.bbAll ^ bbSrc) | bbDst
	bbMask := ^bbDst
	if mv.kind == MoveEnpassant {
		bbTmp := pawnAdvanceForward(inv, bbDst)
		bbAll ^= bbTmp
		bbMask ^= bbTmp
	}
	return !doIsCellAttackedMasked(b, inv, s.king, bbAll, bbMask)
}
