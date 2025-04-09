package game

import (
	"fmt"
	
	"mnkagent/common"
)

// BoardType represents different board implementations
type BoardType int

// Board implementation constants
const (
	Auto BoardType = iota
	Original
	Bitmap
)

// CreateBoard creates a board with the specified implementation type
func CreateBoard(boardType BoardType, m, n, k int) (common.Environment, error) {
	switch boardType {
	case Auto:
		// Automatically choose the most efficient implementation
		// Use bitmap representation for boards that fit in 64 bits
		if m*n <= 64 {
			return CreateBoard(Bitmap, m, n, k)
		}
		return CreateBoard(Original, m, n, k)
	
	case Original:
		return NewMNKBoard(m, n, k)
	
	case Bitmap:
		if m*n > 64 {
			return nil, fmt.Errorf("bitmap board only supports up to 64 cells, but board size is %dx%d = %d cells", 
				m, n, m*n)
		}
		return NewMNKBitboard(m, n, k)
	
	default:
		return nil, fmt.Errorf("unknown board type: %d", boardType)
	}
}

// ConvertMNKStateToString creates a string representation of the board state
func ConvertMNKStateToString(state common.State) string {
	switch s := state.(type) {
	case MNKState:
		var result string
		for i := range s {
			if i > 0 {
				result += "\n"
			}
			for j := range s[i] {
				switch s[i][j] {
				case 0:
					result += "." // Empty
				case 1:
					result += "X" // Player 1
				case 2:
					result += "O" // Player 2
				default:
					result += "?" // Unknown
				}
			}
		}
		return result
	
	case BitboardState:
		// Create a 2D array to represent the state
		board := make([][]int, s.Height)
		for i := range board {
			board[i] = make([]int, s.Width)
		}
		
		// Fill the board from player bitmaps
		for playerID := 1; playerID < len(s.PlayerBits); playerID++ {
			bits := s.PlayerBits[playerID]
			for pos := 0; bits > 0; pos++ {
				if (bits & 1) != 0 {
					y := pos / s.Width
					x := pos % s.Width
					board[y][x] = playerID
				}
				bits >>= 1
			}
		}
		
		// Convert to string
		var result string
		for i := range board {
			if i > 0 {
				result += "\n"
			}
			for j := range board[i] {
				switch board[i][j] {
				case 0:
					result += "." // Empty
				case 1:
					result += "X" // Player 1
				case 2:
					result += "O" // Player 2
				default:
					result += "?" // Unknown
				}
			}
		}
		return result
	
	default:
		return fmt.Sprintf("<Unsupported state type: %T>", state)
	}
}