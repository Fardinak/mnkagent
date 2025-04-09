// Package agents contains implementations of player agents
package human

import (
	"fmt"

	"mnkagent/common"
	"mnkagent/game"
)

// HumanAgent represents a human player
type HumanAgent struct {
	id   int
	sign string
	m, n int // Board dimensions for position calculation
}

// NewHumanAgent creates a new human agent
func NewHumanAgent(id int, sign string, m, n int) *HumanAgent {
	return &HumanAgent{
		id:   id,
		sign: sign,
		m:    m,
		n:    n,
	}
}

// GetID returns the agent's ID
func (agent *HumanAgent) GetID() int {
	return agent.id
}

// FetchMessage returns an empty string (humans don't automatically send messages)
func (agent *HumanAgent) FetchMessage() string {
	return ""
}

// FetchMove prompts the human player for their move
func (agent *HumanAgent) FetchMove(state common.State, _ []common.Action) (common.Action, error) {
	fmt.Print("\n\033[2K\r")
	fmt.Printf("%s > Your move? ", agent.sign)

	var pos int
	_, err := fmt.Scanln(&pos)

	fmt.Print("\r\033[F\033[F")

	if err != nil {
		return nil, err
	}

	// Convert 1-based position to 0-based x,y coordinates
	return game.MNKAction{Y: (pos - 1) / agent.m, X: (pos - 1) % agent.m}, nil
}

// GameOver does nothing for human players
func (agent *HumanAgent) GameOver(_ common.State) {
	// No action needed for human players
}

// GetSign returns the character representing this player on the board
func (agent *HumanAgent) GetSign() string {
	return agent.sign
}