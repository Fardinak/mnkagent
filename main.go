// The implementation of an m,n,k-game with swappable Agents
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

// Flags
var (
	// Game flags
	m         int
	n         int
	k         int
	noDisplay bool
	gomoku    bool

	// RL flags
	rlModelFile       string
	rlModelStatusMode bool
	rlNoLearn         bool
	rlTrainingMode    uint
)

// Signal channel
var sigint chan os.Signal

// Signs
const (
	X = "\033[36;1mX\033[0m"
	O = "\033[31;1mO\033[0m"
)

var players = [3]Agent{
	nil, // No one
	NewHumanAgent(1, X),
	NewRLAgent(2, O, m, n, k, !rlNoLearn),
}
var rounds int
var board *MNKBoard
var flags = make(map[string]bool)

func init() {
	// Game flags
	flag.IntVar(&m, "m", 3, "Board dimention across the horizontal (x) axis")
	flag.IntVar(&n, "n", 3, "Board dimention across the vertical (y) axis")
	flag.IntVar(&k, "k", 3, "Number of marks in a row")
	flag.BoolVar(&noDisplay, "no-display", false, "Do now show board and "+
		"stats in training mode")
	flag.BoolVar(&gomoku, "gomoku", false, "Shortcut for a 19,19,5 game (overrides m, n and k)")

	// RL flags
	flag.StringVar(&rlModelFile, "rl-model", "rl.kw", "RL trained model file "+
		"location")
	flag.BoolVar(&rlModelStatusMode, "rl-model-status", false, "RL trained "+
		"model status")
	flag.BoolVar(&rlNoLearn, "rl-no-learn", false, "Turn off learning for RL "+
		"in normal mode and don't save model to disk")
	flag.UintVar(&rlTrainingMode, "rl-train", 0, "Train RL for n iterations")

	flag.Parse()
}

func main() {
	fmt.Println("Tic-Tac-Toe v1")

	if gomoku {
		m = 19
		n = 19
		k = 5
	}

	var err error
	board, err = NewMNKBoard(m, n, k)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rand.Seed(time.Now().UTC().UnixNano())
	readKnowledgeOK := rlKnowledge.loadFromFile(rlModelFile)

	if rlModelStatusMode {
		if !readKnowledgeOK {
			return
		}

		fmt.Println("Reinforcement learning model report")
		fmt.Printf("Iterations: %d\n", rlKnowledge.Iterations)
		fmt.Printf("Learned states: %d\n", len(rlKnowledge.Values))
		var max float64 = 0
		var min float64 = 0
		for _, v := range rlKnowledge.Values {
			if v > max {
				max = v
			} else if v < min {
				min = v
			}
		}
		fmt.Printf("Maximum value: %f\n", max)
		fmt.Printf("Minimum value: %f\n", min)
		return
	}

	if rlTrainingMode > 0 {
		// Register SIGINT handler
		sigint = make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		go func(c <-chan os.Signal) {
			<-c
			flags["terminate"] = true
			signal.Reset(os.Interrupt)
		}(sigint)
		defer close(sigint)

		// Start training loop
		log := train(rlTrainingMode)
		printStats(log, true)
		return
	}

	fmt.Printf("? > How many rounds shall we play? ")
	_, err = fmt.Scanln(&rounds)
	if err != nil {
		fmt.Println("\n[error] Shit happened!")
		panic(err)
	}
	fmt.Println("Great! Have fun.")

	log := play(rounds)
	printStats(log, false)
}

// train initiates training for given rounds
func train(rounds uint) (log []int) {
	log = make([]int, 3)

	fmt.Println("Commencing training...")

	if err := fileAccessible(rlModelFile); err != nil {
		fmt.Println("Model file not accessible")
		fmt.Println(err)
		return
	}

	players[1] = NewRLAgent(1, X, m, n, k, true)
	players[2] = NewRLAgent(2, O, m, n, k, true)

	var (
		// For the game
		c    uint
		turn int

		// For the progress bar
		termW         int
		cleanupLine   string
		displayH      int = n*2 + 3
		displayBottom string
		displayTop    string
		progress      int
		progressbar   string
		color         string = "\033[41;3m"
		colorDone     string = "\033[46;3m"
	)

	for i := 0; i <= displayH; i++ {
		displayBottom += "\n"
		displayTop += "\033[F"
	}

	for c, turn = 1, 1; c <= rounds; c++ {
		pTick := c*100%rounds == 0
		if pTick || c == 1 {
			// Get terminal width
			termW, _ = getTermSize()

			// Generate cleanup line for terminal width
			cleanupLine = ""
			for i := 0; i < termW; i++ {
				cleanupLine += " "
			}
			cleanupLine += "\r"

			// Track progress
			progress = int(c * 100 / rounds)
			progressbar = generateProgressBar(progress, termW, color, "Training...")
		}

		if pTick || !noDisplay {
			// Clear the progress bar
			fmt.Print(cleanupLine)
		}

		// Start a new round and get the winner's id
		turn = newRound(turn, !noDisplay) // Previous round's winner starts the game
		log[turn]++                       // Keep scores
		if turn == 0 {                    // If it was a draw, next player starts the game
			turn = getNextPlayer(turn)
		}

		if !noDisplay {
			// Print separator and cleanup progress bar
			fmt.Printf("___________________________________\n%s\n", cleanupLine)
		}

		if flags["terminate"] {
			fmt.Print("\r", generateProgressBar(progress, termW, color, "Terminated."), "\n")
			if !rlNoLearn {
				rlKnowledge.saveToFile(rlModelFile)
			}
			return
		}

		if pTick || !noDisplay {
			if !noDisplay && c != rounds {
				// If not 100%, leave room for next board display
				fmt.Print(displayBottom)
			}

			// Print progress bar
			fmt.Print(progressbar)

			if !noDisplay && c != rounds {
				// If not 100%, go to display 0x0
				fmt.Print(displayTop)
			}
		}

		if !rlNoLearn && pTick {
			// Store knowledge every 1/100 of rounds
			rlKnowledge.saveToFile(rlModelFile)
		}
	}

	// Progress bar final touch
	fmt.Print(generateProgressBar(100, termW, colorDone, "Training completed"), "\n")

	return
}

// play initiates game between Human Agent and RL Agent for given rounds
func play(rounds int) (log []int) {
	log = make([]int, 3)

	if err := fileAccessible(rlModelFile); err != nil {
		fmt.Println("Model file not accessible")
		fmt.Println(err)
	}

	players[1] = NewHumanAgent(1, X)
	players[2] = NewRLAgent(2, O, m, n, k, !rlNoLearn)

	for c, turn := 1, 1; c <= rounds; c++ {
		// Start a new round and get the winner's id
		pTurn := turn
		turn = newRound(turn, true) // Previous round's winner starts the game
		log[turn]++                 // Keep scores
		if turn == 0 {              // If it was a draw, next player starts the game
			turn = getNextPlayer(pTurn)
		}

		fmt.Print("___________________________________\n\n")

		if !rlNoLearn {
			rlKnowledge.saveToFile(rlModelFile)
		}
	}
	return
}

// newRound starts a new round
func newRound(turn int, visual bool) int {
	// Reset board
	board.Reset()

	// Set runtime flags
	flags["first_run"] = true

	if visual {
		// Draw a new board
		display(board.GetWorld())
	}

	// Who starts the game if not specified
	if turn == 0 {
		turn = 1
	}

	// Start the game
	for {
		action, err := players[turn].FetchMove(
			board.GetState(turn),
			board.GetPotentialActions(turn))
		if err != nil {
			panic(err)
		}

		_, err = board.Act(turn, action)
		if err != nil {
			// Clear prompt
			fmt.Print("\033[2K\r", err)
		} else {
			if visual {
				// Clear previous messages
				fmt.Printf("\033[2K\rAgent %s: %s / Agent %s: %s",
					players[1].GetSign(), players[1].FetchMessage(),
					players[2].GetSign(), players[2].FetchMessage())

				display(board.GetWorld())
			}

			var result = board.EvaluateAction(turn, action)

			if visual && result != 0 { // Game ended
				// Clear prompt
				fmt.Print("\033[2K\n\033[2K\r")
			}

			if result == 0 { // The game goes on
				turn = getNextPlayer(turn)

			} else if result == -1 { // Draw
				if visual {
					fmt.Println("It's a DRAW!")
				}

				players[1].GameOver(board.GetState(1))
				players[2].GameOver(board.GetState(2))
				return 0

			} else { // Current player won
				if visual {
					fmt.Printf("We have a WINNER! Congratulations %s\n",
						players[turn].GetSign())
				}

				players[1].GameOver(board.GetState(1))
				players[2].GameOver(board.GetState(2))
				return turn
			}
		}
	}
}

// display draws the board on the terminal
func display(board State) {
	var b MNKState = board.(MNKState)
	var mark string

	if flags["first_run"] {
		flags["first_run"] = false
	} else {
		// Reset to app's 0x0 position
		reset := "\r"
		for i := 0; i < n*2+1; i++ {
			reset += "\033[F"
		}
		fmt.Print(reset)
	}

	for i := 0; i < n; i++ {
		line := ""
		if i == 0 {
			// Top
			line = "\u2554"
			for j := 0; j < m; j++ {
				line += "\u2550\u2550\u2550\u2550\u2550"
				if j < m-1 {
					line += "\u2564"
				} else {
					line += "\u2557"
				}
			}
		} else {
			// Middle
			line = "\u2551"
			for j := 0; j < m; j++ {
				line += "\u2500\u2500\u2500\u2500\u2500"
				if j < m-1 {
					line += "\u253c"
				} else {
					line += "\u2551"
				}
			}
		}
		fmt.Println(line)

		line = "\u2551"
		for j := 0; j < m; j++ {
			if j != 0 {
				line += "\u2502"
			}

			index := i*m + j + 1
			padding := [2]string{"", ""}

			if b[i][j] == 0 {
				mark = fmt.Sprintf("\033[37m%d\033[0m", index)

				if index < 10 {
					padding = [2]string{"  ", "  "}
				} else if index < 100 {
					padding = [2]string{" ", "  "}
				} else if index < 1000 {
					padding = [2]string{" ", " "}
				}

			} else {
				mark = players[b[i][j]].GetSign()
				padding = [2]string{"  ", "  "}
			}

			line += padding[0]
			line += mark
			line += padding[1]
		}
		line += "\u2551"
		fmt.Println(line)

		if i+1 == len(b) {
			// Bottom
			line = "\u255a"
			for j := 0; j < m; j++ {
				line += "\u2550\u2550\u2550\u2550\u2550"
				if j < m-1 {
					line += "\u2567"
				} else {
					line += "\u255d"
				}
			}
			fmt.Println(line)
		}
	}
}

// printStats prints out statistics of given game log
func printStats(log []int, rmd bool) {
	var winnerSign string
	winner := max(log)
	if winner == 0 {
		winnerSign = "DRAW"
	} else {
		winnerSign = players[winner].GetSign()
	}
	fmt.Printf("Stats: %s/%s/Draw = %d/%d/%d\nOverall winner: %s\n",
		players[1].GetSign(), players[2].GetSign(), log[1], log[2], log[0],
		winnerSign)

	if rmd {
		fmt.Println("Random move dispersion:")
		for i := 0; i < 9; i++ {
			fmt.Printf("%d: %d\n", i+1, rlKnowledge.randomDispersion[i])
		}
	}
}

// getNextPlayer returns the next player's id
func getNextPlayer(current int) int {
	if current < len(players)-1 {
		return current + 1
	}
	return 1
}

// Get the key of the maximum array item
func max(arr []int) (key int) {
	var max int
	for i := range arr {
		if arr[i] > max {
			max = arr[i]
			key = i
		}
	}
	return
}

// fileAccessible returns true if given path is writable
func fileAccessible(path string) (err error) {
	var f *os.File
	f, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	defer f.Close()
	return
}

// Get terminal size
func getTermSize() (w int, h int) {
	// TODO: Use a terminal ui library like termbox-go
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	d, _ := cmd.Output()
	fmt.Sscan(string(d), &h, &w)
	return
}

// generateProgressBar constructs a progress bar with given information
func generateProgressBar(progress, width int, color string, msg string) string {
	var reset = []byte("\033[0m")
	var ellipsis = []byte("...")
	var fProgress = fmt.Sprintf("%d%%", progress)
	var pb = make([]byte, width, 2*width+20)
	var m = make([]byte, len(msg), len(msg)+len(ellipsis))
	var p = make([]byte, len(fProgress))

	// Generate progress indicator string
	copy(p, fProgress)

	// Truncate message if necessary
	// TODO: Either escape special commands like color, or take care not to truncate them
	copy(m, msg)
	if len(m) > width-8 {
		m = m[:width-5-len(ellipsis)]
		copy(m[len(m)-len(ellipsis):], ellipsis)
	}

	// Fill buffer with empty space
	for i := 0; i < width; i++ {
		copy(pb[i:], " ")
	}

	// Progress indicator position
	pIndex := width*progress/100 + len(color)
	// Message position
	mOffset := (width-8)/2 - len(m)/2
	if mOffset < 1 {
		mOffset = 1
	}

	// Buffer message
	copy(pb[mOffset:], m)

	// Buffer progress indicator
	copy(pb[len(pb)-len(p)-1:], p)

	// Extend buffer to support color, reset and BOL
	pb = pb[0 : len(pb)+len(color)+len(reset)+len("\r")]

	// Add progress color
	copy(pb[len(color):], pb[:])
	copy(pb[:len(color)], color)

	// Reset progress color
	copy(pb[pIndex+len(reset):], pb[pIndex:])
	copy(pb[pIndex:pIndex+len(reset)], reset)

	// Set cursor to BOL
	copy(pb[len(pb)-len("\r"):], "\r")

	return string(pb)
}
