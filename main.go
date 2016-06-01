// The implementation of an m,n,k-game with swappable Agents
package main

// TODO: Tryout Gomoku: 19,19,5-game

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// Flags
var (
	// Game flags
	// TODO: dx, dy, inarow
	noDisplay bool

	// RL flags
	rlModelFile       string
	rlModelStatusMode bool
	rlNoLearn         bool
	rlTrainingMode    uint
)

const dimensions int = 3 // had to be int for compatibility
const inarow int = 3

// Signs
const (
	X = "\033[36;1mX\033[0m"
	O = "\033[31;1mO\033[0m"
)

var players = [3]Agent{
	nil, // No one
	NewHumanAgent(1, X),
	NewRLAgent(2, O, dimensions, dimensions, inarow, !rlNoLearn),
}
var rounds int
var board [][]int
var flags = make(map[string]bool)

type Agent interface {
	// FetchMessage returns agent's messages, if any
	FetchMessage() string

	// FetchMove returns the agent's move based on given state
	FetchMove([][]int) (int, error)

	// GameOver states that the game is over and that the latest state should be saved
	GameOver([][]int)

	// GetSign returns the agent's sign (X|O)
	GetSign() string
}

func init() {
	// Game flags
	flag.BoolVar(&noDisplay, "no-display", false, "Do now show board and "+
		"stats in training mode")

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
		log := train(rlTrainingMode)
		printStats(log)
		return
	}

	fmt.Printf("? > How many rounds shall we play? ")
	_, err := fmt.Scanln(&rounds)
	if err != nil {
		fmt.Println("\n[error] Shit happened!")
		panic(err)
	}
	fmt.Println("Great! Have fun.")

	log := play(rounds)
	printStats(log)
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

	players[1] = NewRLAgent(1, X, dimensions, dimensions, inarow, true)
	players[2] = NewRLAgent(2, O, dimensions, dimensions, inarow, true)

	var c uint
	var turn int
	for c, turn = 1, 1; c <= rounds; c++ {
		// Start a new round and get the winner's id
		turn = newRound(turn, !noDisplay) // Previous round's winner starts the game
		log[turn]++                       // Keep scores
		if turn == 0 {                    // If it was a draw, next player starts the game
			turn = getNextPlayer(turn)
		}

		if !noDisplay {
			fmt.Print("___________________________________\n\n")
		}

		if !rlNoLearn && c%1000 == 0 { // Store knowledge every 1K rounds
			rlKnowledge.saveToFile(rlModelFile)
		}
	}

	if !rlNoLearn {
		rlKnowledge.saveToFile(rlModelFile)
	}

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
	players[2] = NewRLAgent(2, O, dimensions, dimensions, inarow, !rlNoLearn)

	for c, turn := 1, 1; c <= rounds; c++ {
		// Start a new round and get the winner's id
		turn = newRound(turn, true) // Previous round's winner starts the game
		log[turn]++                 // Keep scores
		if turn == 0 {              // If it was a draw, next player starts the game
			turn = getNextPlayer(turn)
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
	initBoard(dimensions, dimensions)

	// Set runtime flags
	flags["first_run"] = true

	if visual {
		// Draw a new board
		display(board)
	}

	// Who starts the game if not specified
	if turn == 0 {
		turn = 1
	}

	// Start the game
	for {
		pos, err := players[turn].FetchMove(board)
		if err != nil {
			fmt.Println("\n\n[error] Shit happened!")
			panic(err)
		}

		if !move(turn, pos) {
			fmt.Print("Invalid move!                      ")
		} else {
			if visual {
				// Clear previous messages
				fmt.Print("                                   \r")
				fmt.Printf("Agent %s: %s / Agent %s: %s",
					players[1].GetSign(), players[1].FetchMessage(),
					players[2].GetSign(), players[2].FetchMessage())

				display(board)
			}

			var result = evaluate(board)
			if result == 0 { // The game goes on
				turn = getNextPlayer(turn)

			} else if result == -1 { // Draw
				if visual {
					// Clear prompt
					fmt.Print("\n                         \r")
					fmt.Println("It's a DRAW!")
				}

				players[1].GameOver(board)
				players[2].GameOver(board)
				return 0

			} else { // Someone won
				if visual {
					fmt.Print("                                   \n")
					fmt.Printf("We have a WINNER! Congratulations %s\n",
						players[result].GetSign())
				}

				players[1].GameOver(board)
				players[2].GameOver(board)
				return result
			}
		}
	}
}

// display draws the board on the terminal
// TODO: Move this to Human Agent
func display(board [][]int) {
	// TODO: Support dimensions
	var mark string

	if flags["first_run"] {
		flags["first_run"] = false
	} else {
		// Reset to app's 0x0 position: \r and seven lines up
		fmt.Print("\r\033[F\033[F\033[F\033[F\033[F\033[F\033[F")
	}

	for i := range board {
		if i == 0 {
			// Top
			fmt.Print("\u2554\u2550\u2550\u2550\u2550\u2550\u2564\u2550\u2550" +
				"\u2550\u2550\u2550\u2564\u2550\u2550\u2550\u2550\u2550\u2557\n")
		} else {
			// Middle
			fmt.Print("\u2551\u2500\u2500\u2500\u2500\u2500\u253c\u2500\u2500" +
				"\u2500\u2500\u2500\u253c\u2500\u2500\u2500\u2500\u2500\u2551\n")
		}

		fmt.Print("\u2551")
		for j := range board[i] {
			if j != 0 {
				fmt.Print("\u2502")
			}

			if board[i][j] == 0 {
				mark = "\033[37m" + strconv.Itoa(i*len(board[i])+j+1) + "\033[0m"
			} else {
				mark = players[board[i][j]].GetSign()
			}
			fmt.Printf("  %s  ", mark)
		}
		fmt.Print("\u2551\n")

		if i+1 == len(board) {
			// Bottom
			fmt.Print("\u255a\u2550\u2550\u2550\u2550\u2550\u2567\u2550\u2550" +
				"\u2550\u2550\u2550\u2567\u2550\u2550\u2550\u2550\u2550\u255d\n")
		}
	}
}

// printStats prints out statistics of given game log
func printStats(log []int) {
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
}

// move registers a move on the board
func move(player int, pos int) bool {
	if pos > dimensions*dimensions {
		return false
	}

	var i = (pos - 1) / dimensions
	var j = (pos - 1) % dimensions

	if board[i][j] != 0 {
		return false
	}

	board[i][j] = player
	return true
}

// Evaluates the board and returns
//   -1: Draw
//    0: Game continues
//   >1: Winner's id
func evaluate(board [][]int) int {
	var b = board
	var d = dimensions
	var i, j int

	// REVIEW: There must be a better solution to this

	for i = 0; i < d-1 && b[i][i] == b[i+1][i+1]; i++ { // Check i,i
		if i >= inarow-2 && b[i][i] != 0 {
			return b[i][i]
		}
	}
	for i = 0; i < d-1 && b[i][d-i-1] == b[i+1][d-i-2]; i++ { // Check i,d-i
		if i >= inarow-2 && b[i][d-i-1] != 0 {
			return b[i][d-i-1]
		}
	}
	for i = 0; i < d; i++ {
		for j = 0; j < d-1 && b[i][j] == b[i][j+1]; j++ { // Check i,j
			if j >= inarow-2 && b[i][j] != 0 {
				return b[i][j]
			}
		}
	}
	for i = 0; i < d; i++ {
		for j = 0; j < d-1 && b[j][i] == b[j+1][i]; j++ { // Check j,i
			if j >= inarow-2 && b[j][i] != 0 {
				return b[j][i]
			}
		}
	}

	// Check if there is any empty room
	for i = range board {
		for j = range board[i] {
			if board[i][j] == 0 {
				return 0
			}
		}
	}

	// It's a draw if none has retuned
	return -1
}

// getNextPlayer returns the next player's id
func getNextPlayer(current int) int {
	if current < len(players)-1 {
		return current + 1
	}
	return 1
}

// initBoard makes a 2D slice, limited to dimensions
func initBoard(m, n int) {
	board = make([][]int, m)
	for i := range board {
		board[i] = make([]int, n)
	}
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
	_, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	return
}
