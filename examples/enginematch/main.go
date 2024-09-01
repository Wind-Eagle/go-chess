// An example on how to run a match between chess engines.
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/alex65536/go-chess/chess"
	"github.com/alex65536/go-chess/clock"
	"github.com/alex65536/go-chess/uci"
	"github.com/alex65536/go-chess/util/maybe"
)

// Run "ucinewgame" command and wait for its completion with a timeout.
func uciNewGame(ctx context.Context, engine *uci.Engine, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := engine.UCINewGame(ctx, true); err != nil {
		return fmt.Errorf("ucinewgame: %w", err)
	}
	return nil
}

func main() {
	// Create a time control. "40/60+0.5" means "60 seconds for each 40 moves, with a 0.5 second
	// increment for each move".
	control, err := clock.ControlFromString("40/60+0.5")
	if err != nil {
		panic(err)
	}

	// Create a context in which the match is being run.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the engine for White. Name is the name of the executable file. The UCI library will
	// wait with a timeout until the engine is initialized. The default timeout value is stored in
	// Options.InitTimeout.
	whiteEngine, err := uci.NewEasyEngine(ctx, uci.EasyEngineOptions{
		Name:            "stockfish",
		WaitInitialized: true,
	})
	if err != nil {
		panic(err)
	}
	// Do not forget to close the engine in the end. Closing ensures that all the resources are
	// freed and the engine process is finished.
	defer whiteEngine.Close()

	// Create the engine for Black, using the same pattern.
	blackEngine, err := uci.NewEasyEngine(ctx, uci.EasyEngineOptions{
		Name:            "stockfish",
		WaitInitialized: true,
	})
	if err != nil {
		panic(err)
	}
	defer blackEngine.Close()

	// Store the engines in an array, so that it would be easier to refer to them later.
	engines := [chess.ColorMax]*uci.Engine{whiteEngine, blackEngine}

	// Send "ucinewgame" for each of the engines and wait until they are ready.
	for c := range chess.ColorMax {
		if err := uciNewGame(ctx, engines[c], 3*time.Second); err != nil {
			panic(fmt.Errorf("ucinewgame %v: %w", c.LongString(), err))
		}
	}

	// Create the new chess game. Note that *clock.Game will manage our clocks automatically. Also,
	// it will autmatically finish the game when it's required by the rules of chess.
	//
	// OutcomeFilter allows you to configure the conditions when the game is terminated. Force
	// filter considers only checkmate and stalemate. Relaxed filter considers more conditions,
	// including the threefold repetition and 50-move rule. Strict filter adheres to FIDE rules and
	// requires the five-fold repetition and 75-move rule for draw instead.
	game := clock.NewGame(
		chess.NewGame(),
		maybe.Some(control),
		clock.GameOptions{
			OutcomeFilter: maybe.Some(chess.VerdictFilterRelaxed),
		},
	)

	// Print the current position in a nice way.
	fmt.Printf("%v\n\n", game.CurBoard().Pretty(chess.PrettyStyleFancy))

	// Create the comments array that will be used later to print the game. We don't want any
	// comments before the first move, so the first item is nil.
	comments := [][]string{nil}

	// Let the next engine make a move until the game terminates.
	for !game.IsFinished() {
		// Get the side to move and the engine associated with it.
		curSide := game.CurSide()
		engine := engines[curSide]

		if err := func() error {
			// Get the deadline for the current engine until it forfeits on time. This function
			// returns false when either the game is finished or there is no time control, which
			// cannot happen in our case.
			deadline, ok := game.Deadline()
			if !ok {
				panic("no time control")
			}

			// Create a context to stop the search when the time is out. Also, we add a small
			// margin of 1ms to allow the engine stop the search gracefully if it is a bit late.
			ctx, cancel := context.WithDeadline(ctx, deadline.Add(1*time.Millisecond))
			defer cancel()

			// Transfer the current position to the engine.
			if err := engine.SetPosition(ctx, game.Inner()); err != nil {
				return fmt.Errorf("set position: %w", err)
			}

			// Start the search. The last argument is nil because we don't want receive the updates
			// about the search progress (depth, PV, current score, etc.)
			search, err := engine.Go(ctx, uci.GoOptions{
				TimeSpec: maybe.Pack(game.UCITimeSpec()),
			}, nil)
			if err != nil {
				return fmt.Errorf("go: %w", err)
			}

			// Wait until the search is finished.
			if err := search.Wait(ctx); err != nil {
				return fmt.Errorf("wait: %w", err)
			}

			// Get the best move. Note that the function returns an error if the engine didn't
			// return the best move or it violates the rules of chess.
			bestMove, err := search.BestMove()
			if err != nil {
				return fmt.Errorf("best move: %w", err)
			}

			// Convert the best move to SAN (Standard Algebraic Notation) to view it later. We
			// should do it before adding the move, as you must pass the current position to the
			// Styled() method.
			bestMoveSAN, err := bestMove.Styled(game.Inner().CurBoard(), chess.MoveStyleSAN)
			if err != nil {
				return fmt.Errorf("convert best move: %w", err)
			}

			// Add the move to the game. This will automatically switch sides and finish the game
			// if necessary.
			if err := game.Push(bestMove); err != nil {
				return fmt.Errorf("add move: %w", err)
			}

			// Get the current clock for the two engines to be printed onto the console. In our
			// case, clock exists so panic if it's not present.
			clock, ok := game.Clock()
			if !ok {
				panic("no clock")
			}

			// Get the position score from the engine and push it into the comments. If there is no
			// score, then skip the comments for the current move.
			maybeScore := search.Status().Score
			if score, ok := maybeScore.TryGet(); ok {
				comments = append(comments, []string{score.String()})
			} else {
				comments = append(comments, nil)
			}

			// Move the cursor to the beginning using escape sequences.
			fmt.Print("\r\033[12A")

			// Print the current position.
			fmt.Printf("%v\n", game.CurBoard().Pretty(chess.PrettyStyleFancy))

			// Erase the current line using escape sequences.
			fmt.Print("\033[2K")

			// Print some information (last player, its move, score, clock)
			fmt.Printf(
				"%v> %v (score: %v, clock: %v - %v)\n",
				curSide.LongString(),
				bestMoveSAN,
				maybeScore.GetOr(uci.ScoreCentipawns(0)),
				clock.White,
				clock.Black,
			)

			return nil
		}(); err != nil {
			// Error communicating with an engine. Probably it doesn't respond because of a bug.
			// Finish the game and indicate that the buggy engine has lost.
			fmt.Printf("engine error: %v\n", err)
			_ = game.Finish(chess.MustWinOutcome(chess.VerdictEngineError, curSide.Inv()))
		}
	}

	// Print the result of the game.
	fmt.Println()
	fmt.Println(game.Outcome())
	fmt.Println()

	// Print the entire game as PGN. Options indicate that the moves are printed in SAN, the move
	// numbers are enabled and there is the game outcome (either "0-1", "1-0" or "1/2-1/2") in the
	// end.
	styledMoves, err := game.Inner().StyledExt(chess.GameStyle{
		Move: chess.MoveStyleSAN,
		MoveNumber: chess.MoveNumberStyle{
			Enabled: true,
		},
		Outcome: chess.GameOutcomeShow,
	}, chess.GameAnnotations{Comments: comments})
	if err != nil {
		panic(err)
	}
	fmt.Println(styledMoves)
}
