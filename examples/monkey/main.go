// An example which creates a chess game consisting of random moves and print it to stdout.
package main

import (
	"fmt"
	"math/rand/v2"

	"github.com/alex65536/go-chess/chess"
)

func main() {
	// Create a new chess game. Note that *chess.Game, unlike *clock.Game doesn't have any clock
	// and doesn't automatically try to finish the game according to the rules of chess.
	game := chess.NewGame()

	// Create a buffer for the generated moves. This helps us to prevent the move generator from
	// doing allocations. In the current example, this optimization might be pretty useless, but it
	// may be helpful in more performance-critical contexts.
	var buf [256]chess.Move

	// While the game is not finished, add more moves to it.
	for !game.IsFinished() {
		// Generate all the possible legal moves. The moves are generated into buf, and the
		// returned slice moves is the prefix of buf.
		moves := game.CurBoard().GenLegalMoves(chess.MoveGenAll, buf[:0])

		// Add a random legal move to the game. Since we know that the move is legal and don't
		// have to check it, we may use infallible PushLegalMove instead of PushMove, which
		// validates the move before adding.
		game.PushLegalMove(moves[rand.IntN(len(moves))])

		// Calculate the result of the game and whether it has to be finished. See "enginematch"
		// example to know move about outcome filters.
		game.SetAutoOutcome(chess.VerdictFilterStrict)
	}

	// Format the game as PGN move text.
	gameStr, err := game.Styled(chess.GameStyle{
		Move: chess.MoveStyleSAN,
		MoveNumber: chess.MoveNumberStyle{
			Enabled: true,
		},
		Outcome: chess.GameOutcomeShow,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(gameStr)
}
