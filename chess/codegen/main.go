//go:build codegen

package main

import (
	"github.com/alex65536/go-chess/chess"
)

func main() {
	chess.Internal_CodegenMain()
}
