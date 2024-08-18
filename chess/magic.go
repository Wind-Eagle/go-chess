package chess

import (
	"fmt"
)

type magicOffsets struct {
	le    [CoordMax]int
	ri    [CoordMax]int
	total int
}

type magic interface {
	Deltas() []CoordDelta
	Magics() []uint64

	BuildMask(c Coord) Bitboard
	BuildPostMask(c Coord) Bitboard
	InitOffsets() magicOffsets
}

type magicRook struct{}

func (r magicRook) Deltas() []CoordDelta {
	return rookDeltas
}

func (r magicRook) Magics() []uint64 {
	return rookMagics[:]
}

func (r magicRook) BuildMask(c Coord) Bitboard {
	return ((BbFile(c.File()) & ^bbFileFrame) |
		(BbRank(c.Rank()) & ^bbRankFrame)) &
		^BitboardFromCoord(c)
}

func (r magicRook) BuildPostMask(c Coord) Bitboard {
	return BbFile(c.File()) ^ BbRank(c.Rank())
}

func (r magicRook) InitOffsets() (offs magicOffsets) {
	offs.total = 0
	for c1 := range CoordMax {
		c2 := c1 ^ Coord(9)
		if c1 > c2 {
			continue
		}
		maxLen := max(r.BuildMask(c1).Len(), r.BuildMask(c2).Len())
		offs.le[c1] = offs.total
		offs.le[c2] = offs.total
		offs.total += 1 << maxLen
		offs.ri[c1] = offs.total
		offs.ri[c2] = offs.total
	}
	return
}

type magicBishop struct{}

func (b magicBishop) Deltas() []CoordDelta {
	return bishopDeltas
}

func (b magicBishop) Magics() []uint64 {
	return bishopMagics[:]
}

func (b magicBishop) BuildMask(c Coord) Bitboard {
	return (BbDiag(c.Diag()) ^ BbAntidiag(c.Antidiag())) &
		^bbDiagFrame
}

func (b magicBishop) BuildPostMask(c Coord) Bitboard {
	return BbDiag(c.Diag()) ^ BbAntidiag(c.Antidiag())
}

func (b magicBishop) InitOffsets() (offs magicOffsets) {
	offs.total = 0
	// We consider 16 groups of shared bishop cells. A single group contains four cells with
	// coordinates `start + j * offset` for all `j = 0..3`.
	var (
		starts  = [16]int{0, 1, 32, 33, 2, 10, 18, 26, 34, 42, 50, 58, 6, 7, 38, 39}
		offsets = [16]int{8, 8, 8, 8, 1, 1, 1, 1, 1, 1, 1, 1, 8, 8, 8, 8}
	)
	for i := range 16 {
		var items [4]Coord
		maxLen := 0
		for j := range 4 {
			items[j] = Coord(starts[i] + j*offsets[i])
			maxLen = max(maxLen, b.BuildMask(items[j]).Len())
		}
		le := offs.total
		offs.total += 1 << maxLen
		ri := offs.total
		for j := range 4 {
			offs.le[items[j]] = le
			offs.ri[items[j]] = ri
		}
	}
	return
}

var (
	_ magic = magicRook{}
	_ magic = magicBishop{}
)

type magicEntry struct {
	shift    int
	index    int
	mask     Bitboard
	postMask Bitboard
}

func magicGenEntries[M magic](m M, lookup []Bitboard) (res [CoordMax]magicEntry) {
	magics := m.Magics()
	deltas := m.Deltas()
	offs := m.InitOffsets()
	if len(lookup) != offs.total {
		panic(fmt.Sprintf("invalid lookup table size: %v expected", offs.total))
	}
	for i := range lookup {
		lookup[i] = BbEmpty
	}
	for c := range CoordMax {
		mask := m.BuildMask(c)
		magic := magics[c]
		shift := mask.Len()
		submaskCnt := uint64(1) << shift
		res[c] = magicEntry{
			shift:    64 - m.BuildMask(c).Len(),
			index:    offs.le[c],
			mask:     mask,
			postMask: m.BuildPostMask(c),
		}
		for submask := range submaskCnt {
			occupied := mask.DepositBits(submask)
			idx := int((uint64(occupied) * magic) >> (64 - shift))
			target := &lookup[idx+res[c].index]
			for _, d := range deltas {
				p := c
				for {
					np, ok := p.Shift(d).TryGet()
					if !ok {
						break
					}
					target.Set(np)
					if occupied.Has(np) {
						break
					}
					p = np
				}
			}
			target.Unset(c)
		}
	}
	return
}

var (
	magicRookEntries   [CoordMax]magicEntry
	magicRookLookup    [65536]Bitboard
	magicBishopEntries [CoordMax]magicEntry
	magicBishopLookup  [1792]Bitboard
)

func init() {
	magicRookEntries = magicGenEntries(magicRook{}, magicRookLookup[:])
	magicBishopEntries = magicGenEntries(magicBishop{}, magicBishopLookup[:])
}

func rookAttacks(c Coord, occupied Bitboard) Bitboard {
	entry, magic := magicRookEntries[c], rookMagics[c]
	idx := int((uint64(occupied&entry.mask) * magic) >> entry.shift)
	return magicRookLookup[idx+entry.index] & entry.postMask
}

func bishopAttacks(c Coord, occupied Bitboard) Bitboard {
	entry, magic := magicBishopEntries[c], bishopMagics[c]
	idx := int((uint64(occupied&entry.mask) * magic) >> entry.shift)
	return magicBishopLookup[idx+entry.index] & entry.postMask
}
