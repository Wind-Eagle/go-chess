//go:build codegen

package chess

import (
	"fmt"
	"math/rand/v2"
	"os"
)

func isValidMagic[M magic](m M, c Coord, magic uint64) bool {
	mask := m.BuildMask(c)
	shift := mask.Len()
	submaskCnt := uint64(1) << shift
	used := make([]bool, submaskCnt)
	for submask := range submaskCnt {
		occupied := mask.DepositBits(submask)
		idx := int((uint64(occupied) * magic) >> (64 - shift))
		if used[idx] {
			return false
		}
		used[idx] = true
	}
	return true
}

func genMagicCandidate(rng rand.Source) uint64 {
	res := uint64(0)
	for range CoordMax {
		res <<= 1
		if rng.Uint64()%8 == 0 {
			res |= 1
		}
	}
	return res
}

func genMagics[M magic](m M, rng rand.Source) (res [CoordMax]uint64) {
	for c := range CoordMax {
		for {
			res[c] = genMagicCandidate(rng)
			if isValidMagic(m, c, res[c]) {
				break
			}
		}
	}
	return
}

func genRookMagics(rng rand.Source) [CoordMax]uint64 {
	return genMagics(magicRook{}, rng)
}

func genBishopMagics(rng rand.Source) [CoordMax]uint64 {
	return genMagics(magicBishop{}, rng)
}

func Internal_CodegenMain() {
	if len(os.Args) != 2 || os.Args[1] != "run_chess_codegen" {
		panic("bad args passed to codegen")
	}

	f, err := os.Create("magic_values.go")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	makeRng := func() rand.Source {
		return rand.NewPCG(0x3141592653589793, 0x2384626433832795)
	}

	printTab := func(name string, value [CoordMax]uint64) {
		fmt.Fprintf(f, "\t%s = [CoordMax]uint64{\n", name)
		for _, v := range value {
			fmt.Fprintf(f, "\t\t0x%016x,\n", v)
		}
		fmt.Fprintln(f, "\t}")
	}

	fmt.Fprintln(f, "// Auto-generated file, DO NOT EDIT!")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "package chess\n\nvar (")
	printTab("rookMagics", genRookMagics(makeRng()))
	fmt.Fprintln(f)
	printTab("bishopMagics", genBishopMagics(makeRng()))
	fmt.Fprintln(f, ")")
}
