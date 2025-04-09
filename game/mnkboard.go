// Package game provides game implementations for the mnkagent project
package game

import (
	"errors"

	"mnkagent/common"
)

// MNKState represents the state of an m,n,k-game
type MNKState [][]int

// Clone creates a deep copy of the state
func (s MNKState) Clone() MNKState {
	sp := make([][]int, len(s))
	for i := range s {
		sp[i] = make([]int, len(s[i]))
		copy(sp[i], s[i])
	}
	return sp
}

// MNKAction represents an action in an m,n,k-game
type MNKAction struct {
	Y, X int
}

// GetParams returns the parameters of the action
func (a MNKAction) GetParams() interface{} {
	return a
}

// MNKBoard implements the Environment interface for m,n,k-games
type MNKBoard struct {
	m, n, k int
	board   MNKState
}

// NewMNKBoard creates a new m,n,k-game board
func NewMNKBoard(m, n, k int) (*MNKBoard, error) {
	if k > m && k > n {
		return nil, errors.New("environment: k exceeds both m and n")
	}

	b := new(MNKBoard)
	b.m = m
	b.n = n
	b.k = k

	// Initialize the board
	b.Reset()

	return b, nil
}

// GetState returns the current state of the board
func (b *MNKBoard) GetState() common.State {
	return b.board.Clone()
}

// GetPotentialActions returns all valid moves for the given agent
func (b *MNKBoard) GetPotentialActions(agentID int) []common.Action {
	var actions []common.Action

	for i := range b.board {
		for j := range b.board[i] {
			if b.board[i][j] == 0 {
				actions = append(actions, MNKAction{
					X: j,
					Y: i,
				})
			}
		}
	}

	return actions
}

// Act executes the given action for the specified agent
func (b *MNKBoard) Act(agentID int, action common.Action) (float64, error) {
	a := action.GetParams().(MNKAction)

	// Validate action
	if a.X < 0 || a.X >= b.m || a.Y < 0 || a.Y >= b.n {
		return 0, errors.New("environment: move out of range")
	}

	if b.board[a.Y][a.X] != 0 {
		return 0, errors.New("environment: invalid move")
	}

	// Execute action
	b.board[a.Y][a.X] = agentID

	// Return reward based on game state
	switch b.EvaluateAction(agentID, action) {
	case 1: // Won
		return 1, nil
	case 0: // Continue
		return 0, nil
	case -1: // Draw
		return -0.5, nil
	default: // Should never happen
		return 0, nil
	}
}

// Evaluate determines if the game has ended and who has won
func (b *MNKBoard) Evaluate() int {
	// Check rows
	for i, c := 0, 1; i < b.n; i, c = i+1, 1 {
		for j := 0; j < b.m-1; j++ {
			if b.board[i][j] == b.board[i][j+1] {
				c++
				if c >= b.k && b.board[i][j] > 0 {
					return b.board[i][j]
				}
			} else {
				c = 1
			}
		}
	}

	// Check columns
	for j, c := 0, 1; j < b.m; j, c = j+1, 1 {
		for i := 0; i < b.n-1; i++ {
			if b.board[i][j] == b.board[i+1][j] {
				c++
				if c >= b.k && b.board[i][j] > 0 {
					return b.board[i][j]
				}
			} else {
				c = 1
			}
		}
	}

	// Check TL-BR diagonals (upper half)
	for o, c := 0, 1; o <= b.m-b.k; o, c = o+1, 1 {
		for i := 0; i < b.m-o-1 && i < b.n-1; i++ {
			if b.board[i][o+i] == b.board[i+1][o+i+1] {
				c++
				if c >= b.k && b.board[i][o+i] > 0 {
					return b.board[i][o+i]
				}
			} else {
				c = 1
			}
		}
	}

	// Check TL-BR diagonals (lower half)
	for o, c := 1, 1; o <= b.n-b.k; o, c = o+1, 1 {
		for i := 0; i < b.m-1 && i < b.n-o-1; i++ {
			if b.board[o+i][i] == b.board[o+i+1][i+1] {
				c++
				if c >= b.k && b.board[o+i][i] > 0 {
					return b.board[o+i][i]
				}
			} else {
				c = 1
			}
		}
	}

	// Check TR-BL diagonals (upper half)
	for o, c := 0, 1; o <= b.m-b.k; o, c = o+1, 1 {
		for i := 0; i < b.m-o-1 && i < b.n-1; i++ {
			if b.board[i][b.m-o-i-1] == b.board[i+1][b.m-o-i-2] {
				c++
				if c >= b.k && b.board[i][b.m-o-i-1] > 0 {
					return b.board[i][b.m-o-i-1]
				}
			} else {
				c = 1
			}
		}
	}

	// Check TR-BL diagonals (lower half)
	for o, c := 1, 1; o <= b.n-b.k; o, c = o+1, 1 {
		for i := 0; i < b.m-1 && i < b.n-o-1; i++ {
			if b.board[i+o][b.m-i-1] == b.board[i+o+1][b.m-i-2] {
				c++
				if c >= b.k && b.board[i+o][b.m-i-1] > 0 {
					return b.board[i+o][b.m-i-1]
				}
			} else {
				c = 1
			}
		}
	}

	// Check if board is full (draw)
	for i := 0; i < b.n; i++ {
		for j := 0; j < b.m; j++ {
			if b.board[i][j] == 0 {
				return 0 // Game continues
			}
		}
	}

	// Draw
	return -1
}

// EvaluateAction checks if the given action would result in a win
func (b *MNKBoard) EvaluateAction(agentID int, action common.Action) int {
	a := action.GetParams().(MNKAction)

	// Check row
	for i, c, d := 1, 1, 0; i < b.k && d < 6; i++ {
		if a.X+i < b.m && b.board[a.Y][a.X+i] == agentID {
			c++
		} else {
			d |= 2
		}

		if a.X-i >= 0 && b.board[a.Y][a.X-i] == agentID {
			c++
		} else {
			d |= 4
		}

		if c >= b.k {
			return 1
		}
	}

	// Check column
	for i, c, d := 1, 1, 0; i < b.k && d < 6; i++ {
		if a.Y+i < b.n && b.board[a.Y+i][a.X] == agentID {
			c++
		} else {
			d |= 2
		}

		if a.Y-i >= 0 && b.board[a.Y-i][a.X] == agentID {
			c++
		} else {
			d |= 4
		}

		if c >= b.k {
			return 1
		}
	}

	// Check TL-BR diagonal
	for i, c, d := 1, 1, 0; i < b.k && d < 6; i++ {
		if a.X+i < b.m && a.Y+i < b.n && b.board[a.Y+i][a.X+i] == agentID {
			c++
		} else {
			d |= 2
		}

		if a.X-i >= 0 && a.Y-i >= 0 && b.board[a.Y-i][a.X-i] == agentID {
			c++
		} else {
			d |= 4
		}

		if c >= b.k {
			return 1
		}
	}

	// Check TR-BL diagonal
	for i, c, d := 1, 1, 0; i < b.k && d < 6; i++ {
		if a.X-i >= 0 && a.Y+i < b.n && b.board[a.Y+i][a.X-i] == agentID {
			c++
		} else {
			d |= 2
		}

		if a.X+i < b.m && a.Y-i >= 0 && b.board[a.Y-i][a.X+i] == agentID {
			c++
		} else {
			d |= 4
		}

		if c >= b.k {
			return 1
		}
	}

	// Check if board is full (draw)
	for i := 0; i < b.n; i++ {
		for j := 0; j < b.m; j++ {
			if b.board[i][j] == 0 {
				return 0 // Game continues
			}
		}
	}

	// Draw
	return -1
}

// Reset initializes the board to an empty state
func (b *MNKBoard) Reset() {
	b.board = make([][]int, b.n)
	for i := range b.board {
		b.board[i] = make([]int, b.m)
	}
}