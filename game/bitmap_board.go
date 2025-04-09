package game

import (
	"errors"
	"fmt"
	
	"mnkagent/common"
)

// Direction constants for win checking
const (
	Horizontal int = iota
	Vertical
	DiagonalTLBR // Top-Left to Bottom-Right
	DiagonalTRBL // Top-Right to Bottom-Left
)

// BitboardState represents the board state using bitmaps
type BitboardState struct {
	// One bitmap per player
	PlayerBits []uint64
	Width      int
	Height     int
}

// Clone creates a deep copy of the state
func (s BitboardState) Clone() BitboardState {
	clone := BitboardState{
		PlayerBits: make([]uint64, len(s.PlayerBits)),
		Width:      s.Width,
		Height:     s.Height,
	}
	
	copy(clone.PlayerBits, s.PlayerBits)
	return clone
}

// MNKBitboard is a high-performance implementation of the MNK board
type MNKBitboard struct {
	m, n, k    int
	board      BitboardState
	moveCount  int
	lastMove   struct {
		playerID int
		x, y     int
	}
}

// NewMNKBitboard creates a new optimized MNK board
func NewMNKBitboard(m, n, k int) (*MNKBitboard, error) {
	// Validate parameters
	if k > m && k > n {
		return nil, errors.New("environment: k exceeds both m and n")
	}
	
	// Verify the board fits in our bitmap representation
	if m*n > 64 {
		return nil, fmt.Errorf("environment: board size %dx%d exceeds maximum supported size (64 cells)", m, n)
	}
	
	b := &MNKBitboard{
		m: m,
		n: n,
		k: k,
	}
	
	// Initialize the board
	b.Reset()
	
	return b, nil
}

// GetState returns the current bitmap state
func (b *MNKBitboard) GetState() common.State {
	return b.board.Clone()
}

// GetPotentialActions returns all valid moves for the given agent
func (b *MNKBitboard) GetPotentialActions(agentID int) []common.Action {
	var actions []common.Action
	
	// Calculate a bitmap with all occupied cells
	occupied := uint64(0)
	for _, playerBits := range b.board.PlayerBits {
		occupied |= playerBits
	}
	
	// Empty cells are potential moves
	for y := 0; y < b.n; y++ {
		for x := 0; x < b.m; x++ {
			pos := y*b.m + x
			if (occupied & (1 << pos)) == 0 {
				actions = append(actions, MNKAction{X: x, Y: y})
			}
		}
	}
	
	return actions
}

// Act executes the given action for the specified agent
func (b *MNKBitboard) Act(agentID int, action common.Action) (float64, error) {
	a := action.GetParams().(MNKAction)
	
	// Validate action
	if a.X < 0 || a.X >= b.m || a.Y < 0 || a.Y >= b.n {
		return 0, errors.New("environment: move out of range")
	}
	
	// Calculate bit position
	pos := a.Y*b.m + a.X
	posBit := uint64(1) << pos
	
	// Check if the cell is already occupied
	occupied := uint64(0)
	for _, playerBits := range b.board.PlayerBits {
		occupied |= playerBits
	}
	
	if (occupied & posBit) != 0 {
		return 0, errors.New("environment: invalid move")
	}
	
	// Update player bitmap and move count
	b.board.PlayerBits[agentID] |= posBit
	b.moveCount++
	
	// Record the last move for efficient evaluation
	b.lastMove.playerID = agentID
	b.lastMove.x = a.X
	b.lastMove.y = a.Y
	
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

// Evaluate checks if the game has ended and who has won
func (b *MNKBitboard) Evaluate() int {
	// If no moves have been made, game is ongoing
	if b.moveCount == 0 {
		return 0
	}
	
	// Check if the last player won
	if b.checkWin(b.lastMove.playerID, b.lastMove.x, b.lastMove.y) {
		return b.lastMove.playerID
	}
	
	// Check if the board is full (draw)
	if b.moveCount == b.m * b.n {
		return -1
	}
	
	// Game continues
	return 0
}

// EvaluateAction checks if the given action would result in a win
func (b *MNKBitboard) EvaluateAction(agentID int, action common.Action) int {
	a := action.GetParams().(MNKAction)
	
	// Apply the move temporarily
	pos := a.Y*b.m + a.X
	posBit := uint64(1) << pos
	oldBits := b.board.PlayerBits[agentID]
	b.board.PlayerBits[agentID] |= posBit
	
	// Check if this move would win
	result := 0
	if b.checkWin(agentID, a.X, a.Y) {
		result = 1
	} else if b.moveCount+1 == b.m*b.n {
		// Draw if board would be full
		result = -1
	}
	
	// Undo the temporary move
	b.board.PlayerBits[agentID] = oldBits
	
	return result
}

// Reset initializes the board to an empty state
func (b *MNKBitboard) Reset() {
	// Initialize with 3 players (0=empty, 1=player1, 2=player2)
	b.board = BitboardState{
		PlayerBits: make([]uint64, 3),
		Width:      b.m,
		Height:     b.n,
	}
	
	// Reset counters
	b.moveCount = 0
	b.lastMove.playerID = 0
	b.lastMove.x = -1
	b.lastMove.y = -1
}

// GetWidth returns the board width (m)
func (b *MNKBitboard) GetWidth() int {
	return b.m
}

// GetHeight returns the board height (n)
func (b *MNKBitboard) GetHeight() int {
	return b.n
}

// GetWinLength returns the winning sequence length (k)
func (b *MNKBitboard) GetWinLength() int {
	return b.k
}

// checkWin efficiently checks if the player has won by placing at position (x,y)
func (b *MNKBitboard) checkWin(playerID, x, y int) bool {
	return b.countInDirection(playerID, x, y, Horizontal) >= b.k ||
		b.countInDirection(playerID, x, y, Vertical) >= b.k ||
		b.countInDirection(playerID, x, y, DiagonalTLBR) >= b.k ||
		b.countInDirection(playerID, x, y, DiagonalTRBL) >= b.k
}

// countInDirection counts how many consecutive marks a player has in a given direction
func (b *MNKBitboard) countInDirection(playerID, x, y, direction int) int {
	// Get player's bitboard
	playerBits := b.board.PlayerBits[playerID]
	
	// Direction deltas
	var dx1, dy1, dx2, dy2 int
	
	switch direction {
	case Horizontal:
		dx1, dy1 = -1, 0
		dx2, dy2 = 1, 0
	case Vertical:
		dx1, dy1 = 0, -1
		dx2, dy2 = 0, 1
	case DiagonalTLBR:
		dx1, dy1 = -1, -1
		dx2, dy2 = 1, 1
	case DiagonalTRBL:
		dx1, dy1 = 1, -1
		dx2, dy2 = -1, 1
	}
	
	// Count in the first direction
	count := 1 // Start with 1 for the current position
	for i := 1; i < b.k; i++ {
		nx, ny := x + i*dx1, y + i*dy1
		
		// Check bounds
		if nx < 0 || nx >= b.m || ny < 0 || ny >= b.n {
			break
		}
		
		// Check if this position has the player's mark
		pos := ny*b.m + nx
		if (playerBits & (1 << pos)) != 0 {
			count++
		} else {
			break
		}
	}
	
	// Count in the second direction
	for i := 1; i < b.k; i++ {
		nx, ny := x + i*dx2, y + i*dy2
		
		// Check bounds
		if nx < 0 || nx >= b.m || ny < 0 || ny >= b.n {
			break
		}
		
		// Check if this position has the player's mark
		pos := ny*b.m + nx
		if (playerBits & (1 << pos)) != 0 {
			count++
		} else {
			break
		}
	}
	
	return count
}