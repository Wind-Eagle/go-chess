package chess

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type RawBoard struct {
	Cells       [CoordMax]Cell
	Side        Color
	Castling    CastlingRights
	EpSource    MaybeCoord
	MoveCounter uint8
	MoveNumber  uint32
}

func InitialRawBoard() RawBoard {
	return RawBoard{
		Cells: func() (res [CoordMax]Cell) {
			for c := range CoordMax {
				res[c] = CellEmpty
			}
			for c := range ColorMax {
				r, pr := homeRank(c), pawnHomeRank(c)
				for f := range FileMax {
					res[CoordFromParts(f, pr)] = CellFromParts(c, PiecePawn)
				}
				res[CoordFromParts(FileA, r)] = CellFromParts(c, PieceRook)
				res[CoordFromParts(FileB, r)] = CellFromParts(c, PieceKnight)
				res[CoordFromParts(FileC, r)] = CellFromParts(c, PieceBishop)
				res[CoordFromParts(FileD, r)] = CellFromParts(c, PieceQueen)
				res[CoordFromParts(FileE, r)] = CellFromParts(c, PieceKing)
				res[CoordFromParts(FileF, r)] = CellFromParts(c, PieceBishop)
				res[CoordFromParts(FileG, r)] = CellFromParts(c, PieceKnight)
				res[CoordFromParts(FileH, r)] = CellFromParts(c, PieceRook)
			}
			return
		}(),
		Side:        ColorWhite,
		Castling:    CastlingRightsFull,
		EpSource:    NoCoord,
		MoveCounter: 0,
		MoveNumber:  1,
	}
}

func parseCells(s string) ([CoordMax]Cell, error) {
	type resT = [CoordMax]Cell
	var res resT
	for c := range CoordMax {
		res[c] = CellEmpty
	}

	file, rank, pos := 0, 0, 0
	for i := range len(s) {
		b := s[i]
		if '1' <= b && b <= '8' {
			add := int(b - '0')
			if file+add > 8 {
				return resT{}, fmt.Errorf("too many items in rank %v", Rank(rank))
			}
			file += add
			pos += add
		} else if b == '/' {
			if file < 8 {
				return resT{}, fmt.Errorf("not enough items in rank %v", Rank(rank))
			}
			rank++
			file = 0
			if rank >= 8 {
				return resT{}, fmt.Errorf("too many ranks")
			}
		} else {
			if file >= 8 {
				return resT{}, fmt.Errorf("too many items in rank %v", Rank(rank))
			}
			var err error
			res[pos], err = CellFromByte(b)
			if err != nil {
				return resT{}, fmt.Errorf("unexpected char %q", b)
			}
			file++
			pos++
		}
	}
	if file < 8 {
		return resT{}, fmt.Errorf("not enough items in rank %v", Rank(rank))
	}
	if rank < 7 {
		return resT{}, fmt.Errorf("too few ranks")
	}
	if file != 8 || rank != 7 || pos != 64 {
		panic("must not happen")
	}

	return res, nil
}

func parseEpSource(s string, side Color) (MaybeCoord, error) {
	res, err := MaybeCoordFromString(s)
	if err != nil {
		return MaybeCoord(0), fmt.Errorf("bad enpassant: %w", err)
	}
	c, ok := res.TryGet()
	if !ok {
		return NoCoord, nil
	}
	if c.Rank() != enpassantDstRank(side) {
		return MaybeCoord(0), fmt.Errorf("invalid enpassant rank %v", c.Rank())
	}
	return SomeCoord(CoordFromParts(c.File(), enpassantSrcRank(side))), nil
}

func RawBoardFromFEN(fen string) (RawBoard, error) {
	if !isASCII(fen) {
		return RawBoard{}, fmt.Errorf("non-ASCII data in FEN")
	}

	fen = strings.Trim(fen, " ")
	spl := strings.Split(fen, " ")

	if len(spl) < 1 {
		return RawBoard{}, fmt.Errorf("board not specified")
	}
	cells, err := parseCells(spl[0])
	if err != nil {
		return RawBoard{}, fmt.Errorf("bad board: %w", err)
	}

	if len(spl) < 2 {
		return RawBoard{}, fmt.Errorf("no move side")
	}
	side, err := ColorFromString(spl[1])
	if err != nil {
		return RawBoard{}, fmt.Errorf("bad move side: %w", err)
	}

	if len(spl) < 3 {
		return RawBoard{}, fmt.Errorf("no castling")
	}
	castling, err := CastlingRightsFromString(spl[2])
	if err != nil {
		return RawBoard{}, fmt.Errorf("bad castling: %w", err)
	}

	if len(spl) < 4 {
		return RawBoard{}, fmt.Errorf("no enpassant")
	}
	epSrc, err := parseEpSource(spl[3], side)
	if err != nil {
		return RawBoard{}, err
	}

	res := RawBoard{
		Cells:       cells,
		Side:        side,
		Castling:    castling,
		EpSource:    epSrc,
		MoveCounter: 0,
		MoveNumber:  1,
	}

	if len(spl) < 5 {
		return res, nil
	}
	moveCounter, err := strconv.ParseUint(spl[4], 10, 32)
	if err != nil {
		return RawBoard{}, fmt.Errorf("bad move counter: %w", err)
	}
	res.MoveCounter = uint8(min(moveCounter, math.MaxUint8))

	if len(spl) < 6 {
		return res, nil
	}
	moveNumber, err := strconv.ParseUint(spl[5], 10, 32)
	if err != nil {
		return RawBoard{}, fmt.Errorf("bad move number: %w", err)
	}
	res.MoveNumber = uint32(moveNumber)

	if len(spl) < 7 {
		return res, nil
	}
	return RawBoard{}, fmt.Errorf("extra data in FEN")
}

func (b *RawBoard) Get(c Coord) Cell {
	return b.Cells[c]
}

func (b *RawBoard) Get2(f File, r Rank) Cell {
	return b.Cells[CoordFromParts(f, r)]
}

func (b *RawBoard) Put(c Coord, ce Cell) {
	b.Cells[c] = ce
}

func (b *RawBoard) Put2(f File, r Rank, ce Cell) {
	b.Cells[CoordFromParts(f, r)] = ce
}

func (b *RawBoard) ZHash() ZHash {
	res := ZHash{}
	if b.Side == ColorWhite {
		res.XorEq(zobristMoveSide)
	}
	if ep, ok := b.EpSource.TryGet(); ok {
		res.XorEq(zobristEnpassant[ep])
	}
	res.XorEq(zobristCastling[b.Castling])
	for i, cell := range b.Cells {
		if cell.IsOccupied() {
			res.XorEq(zobristCells[cell][i])
		}
	}
	return res
}

func (b *RawBoard) EpDest() MaybeCoord {
	if ep, ok := b.EpSource.TryGet(); ok {
		return SomeCoord(CoordFromParts(ep.File(), enpassantDstRank(b.Side)))
	}
	return NoCoord
}

func fmtCells(cells [CoordMax]Cell) string {
	var b strings.Builder
	for r := range RankMax {
		if r != 0 {
			_ = b.WriteByte('/')
		}
		empty := 0
		for f := range FileMax {
			cell := cells[CoordFromParts(f, r)]
			if cell.IsFree() {
				empty++
				continue
			}
			if empty != 0 {
				_ = b.WriteByte(byte('0' + empty))
				empty = 0
			}
			_ = b.WriteByte(cell.ToByte())
		}
		if empty != 0 {
			_ = b.WriteByte(byte('0' + empty))
		}
	}
	return b.String()
}

func (b RawBoard) FEN() string {
	return fmt.Sprintf(
		"%v %v %v %v %v %v",
		fmtCells(b.Cells), b.Side, b.Castling, b.EpDest(), b.MoveCounter, b.MoveNumber,
	)
}

func (b RawBoard) String() string {
	return b.FEN()
}

type PrettyStyle uint8

const (
	PrettyStyleASCII PrettyStyle = iota
	PrettyStyleFancy
)

type styleTable struct {
	horzFrame  rune
	vertFrame  rune
	angleFrame rune
	indicator  [ColorMax]rune
	cell       [CellMax]rune
}

var (
	asciiStyleTable = &styleTable{
		horzFrame:  '-',
		vertFrame:  '|',
		angleFrame: '+',
		indicator:  [ColorMax]rune{ColorWhite: 'W', ColorBlack: 'B'},
		cell: func() (res [CellMax]rune) {
			for c := range CellMax {
				res[c] = rune(c.ToByte())
			}
			return
		}(),
	}

	fancyStyleTable = &styleTable{
		horzFrame:  '─',
		vertFrame:  '│',
		angleFrame: '┼',
		indicator:  [ColorMax]rune{ColorWhite: '○', ColorBlack: '●'},
		cell: func() (res [CellMax]rune) {
			for c := range CellMax {
				res[c] = c.ToRune()
			}
			return
		}(),
	}
)

func doPretty(b *RawBoard, tab *styleTable) string {
	var s strings.Builder
	for r := range RankMax {
		_ = s.WriteByte(r.ToByte())
		_, _ = s.WriteRune(tab.vertFrame)
		for f := range FileMax {
			_, _ = s.WriteRune(tab.cell[b.Get2(f, r)])
		}
		_ = s.WriteByte('\n')
	}
	_, _ = s.WriteRune(tab.horzFrame)
	_, _ = s.WriteRune(tab.angleFrame)
	for range FileMax {
		_, _ = s.WriteRune(tab.horzFrame)
	}
	_ = s.WriteByte('\n')
	_, _ = s.WriteRune(tab.indicator[b.Side])
	_, _ = s.WriteRune(tab.vertFrame)
	for f := range FileMax {
		_ = s.WriteByte(f.ToByte())
	}
	_ = s.WriteByte('\n')
	return s.String()
}

func (b *RawBoard) Pretty(style PrettyStyle) string {
	switch style {
	case PrettyStyleASCII:
		return doPretty(b, asciiStyleTable)
	case PrettyStyleFancy:
		return doPretty(b, fancyStyleTable)
	default:
		panic("invalid pretty style")
	}
}

type Board struct {
	r       RawBoard
	hash    ZHash
	bbCell  [CellMax]Bitboard
	bbColor [ColorMax]Bitboard
	bbAll   Bitboard
}

func NewBoard(r RawBoard) (*Board, error) {
	// Check for out-of-bounds values
	for c := range CoordMax {
		if !r.Cells[c].IsValid() {
			return nil, fmt.Errorf("cell %v is out-of-bounds", c)
		}
	}
	if !r.Side.IsValid() {
		return nil, fmt.Errorf("side is out-of-bounds")
	}
	if !r.Castling.IsValid() {
		return nil, fmt.Errorf("castling is out-of-bounds")
	}
	if !r.EpSource.IsValid() {
		return nil, fmt.Errorf("enpassant source is out-of-bounds")
	}

	// Check enpassant
	if p, ok := r.EpSource.TryGet(); ok {
		// Check InvalidEnpassant
		if p.Rank() != enpassantSrcRank(r.Side) {
			return nil, fmt.Errorf("invalid enpassant coord %v", p)
		}

		// Reset enpassant if either there is no pawn or the cell on the pawn's path is occupied
		pp := p.Add(pawnForwardDelta(r.Side))
		if r.Get(p) != CellFromParts(r.Side.Inv(), PiecePawn) || r.Get(pp).IsOccupied() {
			r.EpSource = NoCoord
		}
	}

	// Reset bad castling flags
	for c := range ColorMax {
		rank := homeRank(c)
		if r.Get2(FileE, rank) != CellFromParts(c, PieceKing) {
			r.Castling.Unset(c, CastlingQueenside)
			r.Castling.Unset(c, CastlingKingside)
		}
		if r.Get2(FileA, rank) != CellFromParts(c, PieceRook) {
			r.Castling.Unset(c, CastlingQueenside)
		}
		if r.Get2(FileH, rank) != CellFromParts(c, PieceRook) {
			r.Castling.Unset(c, CastlingKingside)
		}
	}

	// Calculate bitboards
	var (
		bbColor [ColorMax]Bitboard
		bbCell  [CellMax]Bitboard
	)
	for i, cell := range r.Cells {
		coord := Coord(i)
		if color, ok := cell.Color(); ok {
			bbColor[color].Set(coord)
			bbCell[cell].Set(coord)
		}
	}

	// Check TooManyPieces, NoKing, TooManyKings
	for c := range ColorMax {
		if bbColor[c].Len() > 16 {
			return nil, fmt.Errorf("too many pieces of color %v", c.LongString())
		}
		king := bbCell[CellFromParts(c, PieceKing)]
		if king.IsEmpty() {
			return nil, fmt.Errorf("no king of color %v", c.LongString())
		}
		if king.Len() > 1 {
			return nil, fmt.Errorf("too many kings of color %v", c.LongString())
		}
	}

	// Check InvalidPawn
	bbPawn := bbCell[CellFromParts(ColorWhite, PiecePawn)] | bbCell[CellFromParts(ColorBlack, PiecePawn)]
	const bbBadPawn Bitboard = 0xff000000000000ff
	if !(bbPawn & bbBadPawn).IsEmpty() {
		return nil, fmt.Errorf("invalid pawn position %v", (bbPawn & bbBadPawn).GetFirst())
	}

	// Check OpponentKingAttacked and ImpossibleCheck
	res := &Board{
		r:       r,
		hash:    r.ZHash(),
		bbCell:  bbCell,
		bbColor: bbColor,
		bbAll:   bbColor[ColorWhite] | bbColor[ColorBlack],
	}
	if res.isOpponentKingAttacked() {
		return nil, fmt.Errorf("opponent king is attacked")
	}
	if res.Checkers().Len() > 2 {
		return nil, fmt.Errorf("too many pieces attack the king simultaneously")
	}

	return res, nil
}

func InitialBoard() *Board {
	b, err := NewBoard(InitialRawBoard())
	if err != nil {
		panic(fmt.Sprintf("cannot create initial board: %v", err))
	}
	return b
}

func BoardFromFEN(fen string) (*Board, error) {
	r, err := RawBoardFromFEN(fen)
	if err != nil {
		return nil, fmt.Errorf("parse board: %w", err)
	}
	b, err := NewBoard(r)
	if err != nil {
		return nil, fmt.Errorf("create board: %w", err)
	}
	return b, nil
}

func (b *Board) Raw() RawBoard {
	return b.r
}

func (b *Board) Get(c Coord) Cell {
	return b.r.Get(c)
}

func (b *Board) Get2(f File, r Rank) Cell {
	return b.r.Get2(f, r)
}

func (b *Board) Side() Color {
	return b.r.Side
}

func (b *Board) Castling() CastlingRights {
	return b.r.Castling
}

func (b *Board) EpSource() MaybeCoord {
	return b.r.EpSource
}

func (b *Board) MoveCounter() uint8 {
	return b.r.MoveCounter
}

func (b *Board) MoveNumber() uint32 {
	return b.r.MoveNumber
}

func (b *Board) ZHash() ZHash {
	return b.hash
}

func (b *Board) EpDest() MaybeCoord {
	return b.r.EpDest()
}

func (b *Board) Pretty(style PrettyStyle) string {
	return b.r.Pretty(style)
}

func (b *Board) FEN() string {
	return b.r.FEN()
}

func (b *Board) String() string {
	return b.FEN()
}

func (b *Board) BbColor(c Color) Bitboard {
	return b.bbColor[c]
}

func (b *Board) BbCell(c Cell) Bitboard {
	return b.bbCell[c]
}

func (b *Board) BbPiece(c Color, p Piece) Bitboard {
	return b.bbCell[CellFromParts(c, p)]
}

func (b *Board) bbPieceDiag(c Color) Bitboard {
	return b.BbPiece(c, PieceBishop) | b.BbPiece(c, PieceQueen)
}

func (b *Board) bbPieceLine(c Color) Bitboard {
	return b.BbPiece(c, PieceRook) | b.BbPiece(c, PieceQueen)
}

func (b *Board) KingPos(c Color) Coord {
	return b.BbPiece(c, PieceKing).GetFirst()
}

func (b *Board) MakeLegalMove(mv Move) Undo {
	u := doMakeMove(b, mv)
	return Undo{u: u, mv: mv}
}

func (b *Board) MakeSemilegalMove(mv Move) (Undo, bool) {
	u := doMakeMove(b, mv)
	if b.isOpponentKingAttacked() {
		doUnmakeMove(b, mv, u)
		return Undo{}, false
	}
	return Undo{u: u, mv: mv}, true
}

func (b *Board) MakeMove(mv Move) (Undo, error) {
	if err := mv.Validate(b); err != nil {
		return Undo{}, err
	}
	return b.MakeLegalMove(mv), nil
}

func (b *Board) UnmakeMove(u Undo) {
	doUnmakeMove(b, u.mv, u.u)
}

func (b *Board) MakeUCIMove(um UCIMove) (Undo, error) {
	mv, err := MoveFromUCIMove(um, b)
	if err != nil {
		return Undo{}, fmt.Errorf("parse move: %w", err)
	}
	u, err := b.MakeMove(mv)
	if err != nil {
		return Undo{}, fmt.Errorf("make move: %w", err)
	}
	return u, nil
}

func (b *Board) MakeMoveUCI(s string) (Undo, error) {
	mv, err := MoveFromUCI(s, b)
	if err != nil {
		return Undo{}, fmt.Errorf("parse move: %w", err)
	}
	u, err := b.MakeMove(mv)
	if err != nil {
		return Undo{}, fmt.Errorf("make move: %w", err)
	}
	return u, nil
}

func (b *Board) MakeMoveSAN(s string) (Undo, error) {
	mv, err := LegalMoveFromSAN(s, b)
	if err != nil {
		return Undo{}, fmt.Errorf("parse move: %w", err)
	}
	u := b.MakeLegalMove(mv)
	return u, nil
}

func (b *Board) isOpponentKingAttacked() bool {
	c := b.r.Side
	return b.IsCellAttacked(b.KingPos(c.Inv()), c)
}

func (b *Board) IsCheck() bool {
	c := b.r.Side
	return b.IsCellAttacked(b.KingPos(c), c.Inv())
}

func (b *Board) Checkers() Bitboard {
	c := b.r.Side
	return b.CellAttackers(b.KingPos(c), c.Inv())
}

func (b *Board) IsInsufficientMaterial() bool {
	allWithoutKings :=
		b.bbAll ^ (b.BbPiece(ColorWhite, PieceKing) | b.BbPiece(ColorBlack, PieceKing))

	// If we have pieces on both white and black squares, then no draw occurs. This cutoff
	// optimizes the function in most positions.
	if !(allWithoutKings & BbLight).IsEmpty() && !(allWithoutKings & BbDark).IsEmpty() {
		return false
	}

	// Two kings only
	if allWithoutKings.IsEmpty() {
		return true
	}

	// King vs king + knight
	knights := b.BbPiece(ColorWhite, PieceKnight) | b.BbPiece(ColorBlack, PieceKnight)
	if allWithoutKings == knights && knights.Len() == 1 {
		return true
	}

	// Kings and bishops of the same cell color. Note that we checked above that all the pieces
	// have the same cell color, so we just need to ensure that all the pieces are bishops.
	bishops := b.BbPiece(ColorWhite, PieceBishop) | b.BbPiece(ColorBlack, PieceBishop)
	return allWithoutKings == bishops
}

func (b *Board) CalcOutcome() Outcome {
	// First, we verify for checkmate or stalemate, as force outcome take precedence over
	// non-force ones.
	if !b.HasLegalMoves() {
		if b.IsCheck() {
			return Outcome{Verdict: VerdictCheckmate, Side: b.r.Side.Inv()}
		} else {
			return Outcome{Verdict: VerdictStalemate}
		}
	}

	// Check for insufficient material
	if b.IsInsufficientMaterial() {
		return Outcome{Verdict: VerdictInsufficientMaterial}
	}

	// Check for 50/75 move rule. Note that check for 50 move rule must
	// come after all other ones, because it is non-strict.
	if b.r.MoveCounter >= 150 {
		return Outcome{Verdict: VerdictMoves75}
	}
	if b.r.MoveCounter >= 100 {
		return Outcome{Verdict: VerdictMoves50}
	}

	return Outcome{Verdict: VerdictRunning}
}

func (b *Board) Eq(o *Board) bool {
	return b.r == o.r
}

func (b *Board) Clone() *Board {
	bCopy := *b
	return &bCopy
}
