package chess

import (
	"errors"
	"fmt"
	"math"
)

var (
	errMoveNotWellFormed = errors.New("move is not well-formed")
	errMoveNotSemiLegal  = errors.New("move is not semi-legal")
	errMoveNotLegal      = errors.New("move is not legal")
)

type MoveKind uint8

const (
	MoveNull MoveKind = iota
	MoveSimple
	MoveCastlingQueenside
	MoveCastlingKingside
	MovePawnDouble
	MoveEnpassant
	MovePromoteKnight
	MovePromoteBishop
	MovePromoteRook
	MovePromoteQueen
	MoveKindMax
)

type MoveStyle uint8

const (
	MoveStyleSAN MoveStyle = iota
	MoveStyleFancySAN
	MoveStyleUCI
)

func MoveKindFromCastlingSide(s CastlingSide) MoveKind {
	switch s {
	case CastlingQueenside:
		return MoveCastlingQueenside
	case CastlingKingside:
		return MoveCastlingKingside
	default:
		panic("bad castling side")
	}
}

func (k MoveKind) IsValid() bool {
	return k < MoveKindMax
}

func (k MoveKind) CastlingSide() (CastlingSide, bool) {
	switch k {
	case MoveCastlingQueenside:
		return CastlingQueenside, true
	case MoveCastlingKingside:
		return CastlingKingside, true
	default:
		return CastlingSide(0), false
	}
}

func MoveKindFromPromote(p Piece) (MoveKind, bool) {
	switch p {
	case PieceKnight:
		return MovePromoteKnight, true
	case PieceBishop:
		return MovePromoteBishop, true
	case PieceRook:
		return MovePromoteRook, true
	case PieceQueen:
		return MovePromoteQueen, true
	default:
		return MoveKind(0), false
	}
}

func (k MoveKind) Promote() (Piece, bool) {
	switch k {
	case MovePromoteKnight:
		return PieceKnight, true
	case MovePromoteBishop:
		return PieceBishop, true
	case MovePromoteRook:
		return PieceRook, true
	case MovePromoteQueen:
		return PieceQueen, true
	default:
		return Piece(0), false
	}
}

func (k MoveKind) MatchesPiece(p Piece) bool {
	switch k {
	case MoveNull:
		return false
	case MoveSimple:
		return true
	case MoveCastlingQueenside, MoveCastlingKingside:
		return p == PieceKing
	case MovePawnDouble, MoveEnpassant:
		return p == PiecePawn
	case MovePromoteKnight, MovePromoteBishop, MovePromoteRook, MovePromoteQueen:
		return p == PiecePawn
	default:
		return false
	}
}

type Move struct {
	kind    MoveKind
	srcCell Cell
	src     Coord
	dst     Coord
}

func NullMove() Move {
	return Move{kind: MoveNull, srcCell: CellEmpty, src: Coord(0), dst: Coord(0)}
}

func MoveFromCastling(c Color, s CastlingSide) Move {
	rank := homeRank(c)
	return Move{
		kind:    MoveKindFromCastlingSide(s),
		srcCell: CellFromParts(c, PieceKing),
		src:     CoordFromParts(FileE, rank),
		dst:     CoordFromParts(castlingDstFile(s), rank),
	}
}

func NewMoveUnchecked(kind MoveKind, srcCell Cell, src, dst Coord) Move {
	return Move{kind: kind, srcCell: srcCell, src: src, dst: dst}
}

func MoveFromUCI(s string, b *Board) (Move, error) {
	u, err := UCIMoveFromString(s)
	if err != nil {
		return Move{}, fmt.Errorf("parse uci: %w", err)
	}
	m, err := u.ToMove(b)
	if err != nil {
		return Move{}, fmt.Errorf("convert uci: %w", err)
	}
	return m, nil
}

func SemilegalMoveFromUCI(s string, b *Board) (Move, error) {
	m, err := MoveFromUCI(s, b)
	if err != nil {
		return Move{}, err
	}
	err = m.SemiValidate(b)
	if err != nil {
		return Move{}, fmt.Errorf("bad uci move: %w", err)
	}
	return m, nil
}

func LegalMoveFromUCI(s string, b *Board) (Move, error) {
	m, err := MoveFromUCI(s, b)
	if err != nil {
		return Move{}, err
	}
	err = m.Validate(b)
	if err != nil {
		return Move{}, fmt.Errorf("bad uci move: %w", err)
	}
	return m, nil
}

func LegalMoveFromSAN(s string, b *Board) (Move, error) {
	san, err := sanMoveFromString(s)
	if err != nil {
		return Move{}, fmt.Errorf("parse san: %w", err)
	}
	m, err := san.toLegalMove(b)
	if err != nil {
		return Move{}, fmt.Errorf("convert san: %w", err)
	}
	return m, nil
}

func (m Move) SemiValidate(b *Board) error {
	if !doIsMoveSemilegal(b, m) {
		return errMoveNotSemiLegal
	}
	return nil
}

func (m Move) IsLegalWhenSemilegal(b *Board) bool {
	return doIsLegal(b, newIsLegalState(b), m)
}

func (m Move) Validate(b *Board) error {
	if err := m.SemiValidate(b); err != nil {
		return err
	}
	if !m.IsLegalWhenSemilegal(b) {
		return errMoveNotLegal
	}
	return nil
}

func NewMove(kind MoveKind, srcCell Cell, src, dst Coord) (Move, error) {
	mv := Move{kind: kind, srcCell: srcCell, src: src, dst: dst}
	if !mv.IsWellFormed() {
		return Move{}, errMoveNotWellFormed
	}
	return mv, nil
}

func (m Move) IsWellFormed() bool {
	if !m.kind.IsValid() || !m.srcCell.IsValid() || !m.src.IsValid() || !m.dst.IsValid() {
		return false
	}

	if m.kind == MoveNull {
		return m == NullMove()
	}

	// Valid only for null moves, but they have already been considered.
	if m.srcCell == CellEmpty || m.src == m.dst {
		return false
	}

	color, _ := m.srcCell.Color()
	piece, _ := m.srcCell.Piece()

	if !m.kind.MatchesPiece(piece) {
		return false
	}

	switch m.kind {
	case MoveSimple:
		switch piece {
		case PiecePawn:
			srcFile, srcRank := m.src.File(), m.src.Rank()
			dstFile, dstRank := m.dst.File(), m.dst.Rank()
			if abs(int(srcFile)-int(dstFile)) > 1 ||
				srcRank == Rank1 || srcRank == Rank8 ||
				dstRank == Rank1 || dstRank == Rank8 {
				return false
			}
			switch color {
			case ColorWhite:
				return srcRank == dstRank+1
			case ColorBlack:
				return srcRank+1 == dstRank
			default:
				panic("must not happen")
			}
		case PieceKing:
			return kingAttacks(m.src).Has(m.dst)
		case PieceKnight:
			return knightAttacks(m.src).Has(m.dst)
		case PieceBishop:
			return isBishopMoveValid(m.src, m.dst)
		case PieceRook:
			return isRookMoveValid(m.src, m.dst)
		case PieceQueen:
			return isBishopMoveValid(m.src, m.dst) || isRookMoveValid(m.src, m.dst)
		default:
			panic("must not happen")
		}
	case MoveCastlingQueenside:
		rank := homeRank(color)
		return m.src == CoordFromParts(FileE, rank) && m.dst == CoordFromParts(FileC, rank)
	case MoveCastlingKingside:
		rank := homeRank(color)
		return m.src == CoordFromParts(FileE, rank) && m.dst == CoordFromParts(FileG, rank)
	case MovePawnDouble:
		return m.src.File() == m.dst.File() &&
			m.src.Rank() == pawnHomeRank(color) &&
			m.dst.Rank() == pawnDoubleDstRank(color)
	case MoveEnpassant:
		return m.src.Rank() == enpassantSrcRank(color) &&
			m.dst.Rank() == enpassantDstRank(color) &&
			abs(int(m.src.File())-int(m.dst.File())) == 1
	case MovePromoteKnight, MovePromoteBishop, MovePromoteRook, MovePromoteQueen:
		return m.src.Rank() == promoteSrcRank(color) &&
			m.dst.Rank() == promoteDstRank(color) &&
			abs(int(m.src.File())-int(m.dst.File())) <= 1
	case MoveNull:
		panic("must not happen")
	default:
		return false
	}
}

func (m Move) Kind() MoveKind {
	return m.kind
}

func (m Move) Src() Coord {
	return m.src
}

func (m Move) Dst() Coord {
	return m.dst
}

func (m Move) SrcCell() Cell {
	return m.srcCell
}

func MoveFromUCIMove(u UCIMove, b *Board) (Move, error) {
	return u.ToMove(b)
}

func SemilegalMoveFromUCIMove(u UCIMove, b *Board) (Move, error) {
	m, err := MoveFromUCIMove(u, b)
	if err != nil {
		return Move{}, err
	}
	err = m.SemiValidate(b)
	if err != nil {
		return Move{}, fmt.Errorf("bad uci move: %w", err)
	}
	return m, nil
}

func LegalMoveFromUCIMove(u UCIMove, b *Board) (Move, error) {
	m, err := MoveFromUCIMove(u, b)
	if err != nil {
		return Move{}, err
	}
	err = m.Validate(b)
	if err != nil {
		return Move{}, fmt.Errorf("bad uci move: %w", err)
	}
	return m, nil
}

func (m Move) UCIMove() UCIMove {
	if m.kind == MoveNull {
		return UCIMove{Kind: UCIMoveNull}
	}
	if p, ok := m.kind.Promote(); ok {
		return UCIMove{Kind: UCIMovePromote, Src: m.src, Dst: m.dst, Promote: p}
	}
	return UCIMove{Kind: UCIMoveSimple, Src: m.src, Dst: m.dst}
}

func (m Move) String() string {
	return m.UCI()
}

func (m Move) UCI() string {
	return m.UCIMove().String()
}

func (m Move) sanImpl(b *Board, tab *sanStyleTable) (string, error) {
	san, err := sanMoveFromMove(m, b)
	if err != nil {
		return "", fmt.Errorf("convert move to san: %w", err)
	}
	s, err := san.styled(tab)
	if err != nil {
		return "", fmt.Errorf("convert san to string: %w", err)
	}
	return s, nil
}

func (m Move) Styled(b *Board, style MoveStyle) (string, error) {
	switch style {
	case MoveStyleUCI:
		return m.UCI(), nil
	case MoveStyleSAN:
		return m.sanImpl(b, sanStyleASCII)
	case MoveStyleFancySAN:
		return m.sanImpl(b, sanStyleFancy)
	default:
		panic("bad move style")
	}
}

type rawUndo struct {
	hash        ZHash
	dstCell     Cell
	castling    CastlingRights
	epSource    MaybeCoord
	moveCounter uint8
	moveNumber  uint32
}

type Undo struct {
	u  rawUndo
	mv Move
}

func (u Undo) Move() Move {
	return u.mv
}

func updateCastling(b *Board, bbDiff Bitboard) {
	if (bbDiff & bbCastlingAllSrcs).IsEmpty() {
		return
	}

	c := b.r.Castling
	if !(bbDiff & bbCastlingSrcs(ColorWhite, CastlingQueenside)).IsEmpty() {
		c.Unset(ColorWhite, CastlingQueenside)
	}
	if !(bbDiff & bbCastlingSrcs(ColorWhite, CastlingKingside)).IsEmpty() {
		c.Unset(ColorWhite, CastlingKingside)
	}
	if !(bbDiff & bbCastlingSrcs(ColorBlack, CastlingQueenside)).IsEmpty() {
		c.Unset(ColorBlack, CastlingQueenside)
	}
	if !(bbDiff & bbCastlingSrcs(ColorBlack, CastlingKingside)).IsEmpty() {
		c.Unset(ColorBlack, CastlingKingside)
	}

	if c != b.r.Castling {
		b.hash.XorEq(zobristCastling[b.r.Castling])
		b.r.Castling = c
		b.hash.XorEq(zobristCastling[b.r.Castling])
	}
}

func doMakePawnDouble(b *Board, c Color, mv Move, bbDiff Bitboard, inv bool) {
	pawn := CellFromParts(c, PiecePawn)
	if inv {
		b.r.Put(mv.src, pawn)
		b.r.Put(mv.dst, CellEmpty)
	} else {
		b.r.Put(mv.src, CellEmpty)
		b.r.Put(mv.dst, pawn)
		b.hash.XorEq(zobristCells[pawn][mv.src].Xor(zobristCells[pawn][mv.dst]))
	}
	b.bbColor[c] ^= bbDiff
	b.bbCell[pawn] ^= bbDiff
	if !inv {
		b.r.EpSource = SomeCoord(mv.dst)
		b.hash.XorEq(zobristEnpassant[mv.dst])
	}
}

func doMakeEnpassant(b *Board, c Color, mv Move, bbDiff Bitboard, inv bool) {
	aux := mv.dst.Add(-pawnForwardDelta(c))
	bbAux := BitboardFromCoord(aux)
	ourPawn, theirPawn := CellFromParts(c, PiecePawn), CellFromParts(c.Inv(), PiecePawn)
	if inv {
		b.r.Put(mv.src, ourPawn)
		b.r.Put(mv.dst, CellEmpty)
		b.r.Put(aux, theirPawn)
	} else {
		b.r.Put(mv.src, CellEmpty)
		b.r.Put(mv.dst, ourPawn)
		b.r.Put(aux, CellEmpty)
		b.hash.XorEq(
			zobristCells[ourPawn][mv.src].
				Xor(zobristCells[ourPawn][mv.dst]).
				Xor(zobristCells[theirPawn][aux]),
		)
	}
	b.bbColor[c] ^= bbDiff
	b.bbCell[ourPawn] ^= bbDiff
	b.bbColor[c.Inv()] ^= bbAux
	b.bbCell[theirPawn] ^= bbAux
}

func doMakeCastlingQueenside(b *Board, c Color, inv bool) {
	king := CellFromParts(c, PieceKing)
	rook := CellFromParts(c, PieceRook)
	rank := homeRank(c)
	if inv {
		b.r.Put2(FileA, rank, rook)
		b.r.Put2(FileC, rank, CellEmpty)
		b.r.Put2(FileD, rank, CellEmpty)
		b.r.Put2(FileE, rank, king)
	} else {
		b.r.Put2(FileA, rank, CellEmpty)
		b.r.Put2(FileC, rank, king)
		b.r.Put2(FileD, rank, rook)
		b.r.Put2(FileE, rank, CellEmpty)
		b.hash.XorEq(zobristCastlingDelta[c][CastlingQueenside])
	}
	off := castlingOffset(c)
	b.bbColor[c] ^= Bitboard(0x1d) << off
	b.bbCell[rook] ^= Bitboard(0x09) << off
	b.bbCell[king] ^= Bitboard(0x14) << off
	if !inv {
		b.hash.XorEq(zobristCastling[b.r.Castling])
		b.r.Castling.UnsetColor(c)
		b.hash.XorEq(zobristCastling[b.r.Castling])
	}
}

func doMakeCastlingKingside(b *Board, c Color, inv bool) {
	king := CellFromParts(c, PieceKing)
	rook := CellFromParts(c, PieceRook)
	rank := homeRank(c)
	if inv {
		b.r.Put2(FileE, rank, king)
		b.r.Put2(FileF, rank, CellEmpty)
		b.r.Put2(FileG, rank, CellEmpty)
		b.r.Put2(FileH, rank, rook)
	} else {
		b.r.Put2(FileE, rank, CellEmpty)
		b.r.Put2(FileF, rank, rook)
		b.r.Put2(FileG, rank, king)
		b.r.Put2(FileH, rank, CellEmpty)
		b.hash.XorEq(zobristCastlingDelta[c][CastlingKingside])
	}
	off := castlingOffset(c)
	b.bbColor[c] ^= Bitboard(0xf0) << off
	b.bbCell[rook] ^= Bitboard(0xa0) << off
	b.bbCell[king] ^= Bitboard(0x50) << off
	if !inv {
		b.hash.XorEq(zobristCastling[b.r.Castling])
		b.r.Castling.UnsetColor(c)
		b.hash.XorEq(zobristCastling[b.r.Castling])
	}
}

func doMakeMove(b *Board, mv Move) rawUndo {
	srcCell, dstCell := mv.srcCell, b.Get(mv.dst)
	undo := rawUndo{
		hash:        b.hash,
		dstCell:     dstCell,
		castling:    b.r.Castling,
		epSource:    b.r.EpSource,
		moveCounter: b.r.MoveCounter,
		moveNumber:  b.r.MoveNumber,
	}
	bbSrc, bbDst := BitboardFromCoord(mv.src), BitboardFromCoord(mv.dst)
	bbDiff := bbSrc | bbDst
	c := b.r.Side
	pawn := CellFromParts(c, PiecePawn)
	if p, ok := b.r.EpSource.TryGet(); ok {
		b.hash.XorEq(zobristEnpassant[p])
		b.r.EpSource = NoCoord
	}
	switch mv.kind {
	case MoveSimple:
		b.r.Put(mv.src, CellEmpty)
		b.r.Put(mv.dst, srcCell)
		b.hash.XorEq(
			zobristCells[srcCell][mv.src].
				Xor(zobristCells[srcCell][mv.dst]).
				Xor(zobristCells[dstCell][mv.dst]),
		)
		b.bbColor[c] ^= bbDiff
		b.bbCell[srcCell] ^= bbDiff
		b.bbColor[c.Inv()] &= ^bbDst
		b.bbCell[dstCell] &= ^bbDst
		if srcCell != pawn {
			updateCastling(b, bbDiff)
		}
	case MovePawnDouble:
		doMakePawnDouble(b, c, mv, bbDiff, false)
	case MovePromoteKnight, MovePromoteBishop, MovePromoteRook, MovePromoteQueen:
		promotePiece, _ := mv.kind.Promote()
		promote := CellFromParts(c, promotePiece)
		b.r.Put(mv.src, CellEmpty)
		b.r.Put(mv.dst, promote)
		b.hash.XorEq(
			zobristCells[srcCell][mv.src].
				Xor(zobristCells[promote][mv.dst]).
				Xor(zobristCells[dstCell][mv.dst]),
		)
		b.bbColor[c] ^= bbDiff
		b.bbCell[srcCell] ^= bbSrc
		b.bbCell[promote] ^= bbDst
		b.bbColor[c.Inv()] &= ^bbDst
		b.bbCell[dstCell] &= ^bbDst
		updateCastling(b, bbDiff)
	case MoveCastlingQueenside:
		doMakeCastlingQueenside(b, c, false)
	case MoveCastlingKingside:
		doMakeCastlingKingside(b, c, false)
	case MoveNull:
		// Do nothing
	case MoveEnpassant:
		doMakeEnpassant(b, c, mv, bbDiff, false)
	default:
		panic("bad move kind")
	}

	if dstCell != CellEmpty || srcCell == pawn {
		b.r.MoveCounter = 0
	} else if b.r.MoveCounter != math.MaxUint8 {
		b.r.MoveCounter++
	}
	b.r.Side = c.Inv()
	b.hash.XorEq(zobristMoveSide)
	if c == ColorBlack && b.r.MoveNumber != math.MaxUint32 {
		b.r.MoveNumber++
	}
	b.bbAll = b.bbColor[ColorWhite] | b.bbColor[ColorBlack]

	return undo
}

func doUnmakeMove(b *Board, mv Move, u rawUndo) {
	bbSrc, bbDst := BitboardFromCoord(mv.src), BitboardFromCoord(mv.dst)
	bbDiff := bbSrc | bbDst
	srcCell, dstCell := b.Get(mv.dst), u.dstCell
	c := b.r.Side.Inv()

	switch mv.kind {
	case MoveSimple:
		b.r.Put(mv.src, srcCell)
		b.r.Put(mv.dst, dstCell)
		b.bbColor[c] ^= bbDiff
		b.bbCell[srcCell] ^= bbDiff
		if dstCell.IsOccupied() {
			b.bbColor[c.Inv()] |= bbDst
			b.bbCell[dstCell] |= bbDst
		}
	case MovePawnDouble:
		doMakePawnDouble(b, c, mv, bbDiff, true)
	case MovePromoteKnight, MovePromoteBishop, MovePromoteRook, MovePromoteQueen:
		pawn := CellFromParts(c, PiecePawn)
		b.r.Put(mv.src, pawn)
		b.r.Put(mv.dst, dstCell)
		b.bbColor[c] ^= bbDiff
		b.bbCell[pawn] ^= bbSrc
		b.bbCell[srcCell] ^= bbDst
		if dstCell.IsOccupied() {
			b.bbColor[c.Inv()] |= bbDst
			b.bbCell[dstCell] |= bbDst
		}
	case MoveCastlingQueenside:
		doMakeCastlingQueenside(b, c, true)
	case MoveCastlingKingside:
		doMakeCastlingKingside(b, c, true)
	case MoveNull:
		// Do nothing
	case MoveEnpassant:
		doMakeEnpassant(b, c, mv, bbDiff, true)
	default:
		panic("bad move kind")
	}

	b.hash = u.hash
	b.r.Castling = u.castling
	b.r.EpSource = u.epSource
	b.r.MoveCounter = u.moveCounter
	b.r.MoveNumber = u.moveNumber
	b.r.Side = c
	b.bbAll = b.bbColor[ColorWhite] | b.bbColor[ColorBlack]
}

func isQueenSemilegal(src, dst Coord, bbAll Bitboard) bool {
	if isBishopMoveValid(src, dst) {
		return (betweenBishopStrict(src, dst) & bbAll).IsEmpty()
	} else {
		return (betweenRookStrict(src, dst) & bbAll).IsEmpty()
	}
}

func doIsMoveSemilegal(b *Board, mv Move) bool {
	c := b.r.Side
	dstCell := b.Get(mv.dst)

	if mv.kind == MoveNull ||
		b.Get(mv.src) != mv.srcCell ||
		!mv.srcCell.HasColor(c) ||
		dstCell.HasColor(c) {
		return false
	}

	piece, ok := mv.srcCell.Piece()
	if !ok {
		panic("must not happen")
	}

	switch piece {
	case PiecePawn:
		switch mv.kind {
		case MovePawnDouble:
			tmpCell := b.Get(mv.src.Add(pawnForwardDelta(c)))
			return tmpCell.IsFree() && dstCell.IsFree()
		case MoveEnpassant:
			if p, ok := b.r.EpSource.TryGet(); ok {
				return mv.dst == p.Add(pawnForwardDelta(c))
			} else {
				return false
			}
		case MoveSimple, MovePromoteKnight, MovePromoteBishop, MovePromoteRook, MovePromoteQueen:
			return (mv.dst.File() == mv.src.File()) == dstCell.IsFree()
		default:
			panic("must not happen")
		}
	case PieceKing:
		switch mv.kind {
		case MoveCastlingQueenside:
			return b.r.Castling.Has(c, CastlingQueenside) &&
				(b.bbAll & bbCastlingPass(c, CastlingQueenside)).IsEmpty() &&
				!b.IsCellAttacked(mv.src, c.Inv()) &&
				!b.IsCellAttacked(mv.src.Add(-1), c.Inv())
		case MoveCastlingKingside:
			return b.r.Castling.Has(c, CastlingKingside) &&
				(b.bbAll & bbCastlingPass(c, CastlingKingside)).IsEmpty() &&
				!b.IsCellAttacked(mv.src, c.Inv()) &&
				!b.IsCellAttacked(mv.src.Add(1), c.Inv())
		case MoveSimple:
			return true
		default:
			panic("must not happen")
		}
	case PieceKnight:
		return true
	case PieceBishop:
		return (betweenBishopStrict(mv.src, mv.dst) & b.bbAll).IsEmpty()
	case PieceRook:
		return (betweenRookStrict(mv.src, mv.dst) & b.bbAll).IsEmpty()
	case PieceQueen:
		return isQueenSemilegal(mv.src, mv.dst, b.bbAll)
	default:
		panic("must not happen")
	}
}
