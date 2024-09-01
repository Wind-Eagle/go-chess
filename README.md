# Chess Library for Golang

This library implements the basic chess primitives in Golang, allowing you to create various
chess-related applications.

## What Is Implemented

* Rules of chess, move validation, move generation, etc.
* FEN support
* Moves in UCI and SAN format
* Running UCI engines
* Time control

## What Is Not Implemented

* PGN (but see `chess.(*Game).Styled` if you want to export the game to PGN)
* UCI `"copyprotection"`, `"registration"` and `"register"` commands
* WinBoard/XBoard engines
* Support for opening book formats
* Endgame tablebases
* Chess960 and other chess variants

## Design Goals

The main design goal is to make the APIs simple to use and hard to misuse. Also, some care is taken
about performance.

Also, the project tends to be dependency-free, though, it uses [testify][testify] for testing.

The implementation borrows heavily from my chess implementations on other programming languages,
like [SoFCheck in C++][sofcheck] and [owlchess in Rust][owlchess].

## Usage

See the [examples][examples] and the [documentation][docs].

## Contributing

Pull requests are welcome ;) There is no formal process on how to submit pull requests, it may be
establish once there are many enough of them.

Also, reporting issues is appreciated.

## Move Generator Performance

Benchmarks taken on my AMD Ryzen 7 5700U CPU.

```
BenchmarkGenMoves/initial             9135909   132.2 ns/op
BenchmarkGenMoves/sicilian            5885708   204.1 ns/op
BenchmarkGenMoves/middle              5832038   204.9 ns/op
BenchmarkGenMoves/openPosition        8305612   144.5 ns/op
BenchmarkGenMoves/queen               7848086   152.8 ns/op
BenchmarkGenMoves/pawnMove            12137179  97.91 ns/op
BenchmarkGenMoves/pawnAttack          12487778  96.04 ns/op
BenchmarkGenMoves/pawnPromote         7141045   167.8 ns/op
BenchmarkGenMoves/cydonia             6376789   188.1 ns/op
BenchmarkGenMoves/max                 1353169   888.0 ns/op
BenchmarkGenMovesLegal/initial        3354638   358.9 ns/op
BenchmarkGenMovesLegal/sicilian       1920674   629.8 ns/op
BenchmarkGenMovesLegal/middle         1856020   662.3 ns/op
BenchmarkGenMovesLegal/openPosition   2683543   451.5 ns/op
BenchmarkGenMovesLegal/queen          2545749   470.2 ns/op
BenchmarkGenMovesLegal/pawnMove       4211289   283.7 ns/op
BenchmarkGenMovesLegal/pawnAttack     4142245   304.2 ns/op
BenchmarkGenMovesLegal/pawnPromote    2157934   555.5 ns/op
BenchmarkGenMovesLegal/cydonia        1969893   610.5 ns/op
BenchmarkGenMovesLegal/max            400219    2996 ns/op
BenchmarkMakeMove/initial             1693815   708.1 ns/op
BenchmarkMakeMove/sicilian            843788    1423 ns/op
BenchmarkMakeMove/middle              799842    1495 ns/op
BenchmarkMakeMove/openPosition        1000000   1004 ns/op
BenchmarkMakeMove/queen               1000000   1094 ns/op
BenchmarkMakeMove/pawnMove            1864294   643.6 ns/op
BenchmarkMakeMove/pawnAttack          1779667   674.3 ns/op
BenchmarkMakeMove/pawnPromote         844542    1414 ns/op
BenchmarkMakeMove/cydonia             853094    1408 ns/op
BenchmarkMakeMove/max                 161667    7407 ns/op
BenchmarkMakeMoveChecked/initial      1264999   946.8 ns/op
BenchmarkMakeMoveChecked/sicilian     603042    1977 ns/op
BenchmarkMakeMoveChecked/middle       574980    2076 ns/op
BenchmarkMakeMoveChecked/openPosition 904882    1317 ns/op
BenchmarkMakeMoveChecked/queen        805620    1485 ns/op
BenchmarkMakeMoveChecked/pawnMove     1396708   861.9 ns/op
BenchmarkMakeMoveChecked/pawnAttack   1337276   898.9 ns/op
BenchmarkMakeMoveChecked/pawnPromote  640267    1866 ns/op
BenchmarkMakeMoveChecked/cydonia      621738    1919 ns/op
BenchmarkMakeMoveChecked/max          113653    10555 ns/op
BenchmarkKingAttack/initial           135330769 8.797 ns/op
BenchmarkKingAttack/sicilian          136473920 8.788 ns/op
BenchmarkKingAttack/middle            136460737 8.794 ns/op
BenchmarkKingAttack/openPosition      136149042 8.794 ns/op
BenchmarkKingAttack/queen             136318644 8.794 ns/op
BenchmarkKingAttack/pawnMove          136359507 8.831 ns/op
BenchmarkKingAttack/pawnAttack        136371324 8.801 ns/op
BenchmarkKingAttack/pawnPromote       136274674 8.814 ns/op
BenchmarkKingAttack/cydonia           136332756 9.053 ns/op
BenchmarkKingAttack/max               136371939 8.848 ns/op
```

The performance is ~3 times slower than [owlchess][owlchess] in Rust, though note that Go compiler
applies less optimization and generates the slower code overall.

If you plan to implement a chess engine in Go, I advice you against of doing so. Better use more
performant languages such as Rust, C++ or Zig.

## License

The code is distributed under the terms of MIT License.

[sofcheck]: https://github.com/alex65536/sofcheck
[owlchess]: https://github.com/alex65536/owlchess
[examples]: examples
[docs]: https://pkg.go.dev/github.com/alex65536/go-chess
[testify]: https://github.com/stretchr/testify
