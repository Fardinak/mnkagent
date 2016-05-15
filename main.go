// The implementation of an m,n,k-game with swappable Agents
package main

import (
	"fmt"
	"strconv"
)

// TODO: Tryout Gomoku: 19,19,5-game
const dimensions int = 3 // had to be int for compatibility
const inarow int = 3

var players = []Agent{
	nil, // No one
	NewHumanAgent("\033[36;1mX\033[0m"),
	NewHumanAgent("\033[31;1mO\033[0m"),
}
var rounds int
var board [][]int
var flags = make(map[string]bool)

type Agent interface {
	FetchMove() (int, error)
	GetSign() string
}

func main() {
	// Make a 2D slice, limited to dimensions
	initBoard()

	fmt.Println("Tic-Tac-Toe v1")
	fmt.Printf("_ > How many rounds shall we play? ")
	_, err := fmt.Scanln(&rounds)
	if err != nil {
		fmt.Println("\n[error] Shit happened!")
		panic(err)
	}
	fmt.Println("Great! Have fun.")

	var log = make([]int, 3)
	for c, turn := 1, 1; c <= rounds; c++ {
		// Start a new round and get the winner's id
		turn = newRound(turn) // Previous round's winner starts the game
		log[turn]++
		if turn == 0 { // If it was a draw, next player starts the game
			turn = getNextPlayer(turn)
		}

		if rounds > 1 {
			fmt.Print("___________________________________\n\n")
		}
	}

	fmt.Printf("Stats: %s/%s/Draw = %d/%d/%d\nOverall winner: %s\n",
		players[1].GetSign(), players[2].GetSign(), log[1], log[2], log[0],
		players[max(log)].GetSign())
}

// Get the key of the maximum array item
func max(arr []int) (key int) {
	var m int
	for i := range arr {
		if arr[i] > m {
			m = arr[i]
			key = i
		}
	}
	return
}

// Initialize empty board ([[0 0 0] [0 0 0] [0 0 0]])
func initBoard() {
	board = make([][]int, dimensions)
	for i := range board {
		board[i] = make([]int, dimensions)
	}
}

// newRound starts a new round
func newRound(turn int) int {
	// Reset board
	initBoard()

	// Set flags
	flags["first_run"] = true

	// Draw a new board
	display(board)

	// Who starts the game if not specified
	if turn == 0 {
		turn = 1
	}

	// Start the game
	for {
		pos, err := players[turn].FetchMove()
		if err != nil {
			fmt.Println("\n\n[error] Shit happened!")
			panic(err)
		}

		if !move(turn, pos) {
			fmt.Print("Invalid move!")
		} else {
			fmt.Print("             ")
			display(board)
			var result = evaluate(board)
			if result == 0 { // The game goes on
				turn = getNextPlayer(turn)

			} else if result == -1 { // Draw
				fmt.Print("\n                         \r")
				fmt.Println("It's a DRAW!")
				return 0

			} else { // Someone won
				fmt.Printf("\nWe have a WINNER! Congratulations %s\n",
					players[result].GetSign())
				return result
			}
		}
	}
}

// display draws the board on the terminal
func display(board [][]int) {
	var mark string

	if flags["first_run"] {
		flags["first_run"] = false
	} else {
		// Reset to app's 0x0 position: \r and seven lines up
		fmt.Print("\r\033[F\033[F\033[F\033[F\033[F\033[F\033[F")
	}

	for i := range board {
		if i == 0 {
			fmt.Print("\u2554\u2550\u2550\u2550\u2550\u2550\u2564\u2550\u2550" +
				"\u2550\u2550\u2550\u2564\u2550\u2550\u2550\u2550\u2550\u2557\n")
		} else {
			fmt.Print("\u2551\u2500\u2500\u2500\u2500\u2500\u253c\u2500\u2500" +
				"\u2500\u2500\u2500\u253c\u2500\u2500\u2500\u2500\u2500\u2551\n")
		}

		fmt.Print("\u2551")
		for j := range board[i] {
			if j != 0 {
				fmt.Print("\u2502")
			}

			if board[i][j] == 0 {
				mark = "\033[37m" + strconv.Itoa(i*dimensions+j+1) + "\033[0m"
			} else {
				mark = players[board[i][j]].GetSign()
			}
			fmt.Printf("  %s  ", mark)
		}
		fmt.Print("\u2551\n")

		if i+1 == dimensions {
			fmt.Print("\u255a\u2550\u2550\u2550\u2550\u2550\u2567\u2550\u2550" +
				"\u2550\u2550\u2550\u2567\u2550\u2550\u2550\u2550\u2550\u255d\n")
		}
	}
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
