// Package common provides shared interfaces and types used across the mnkagent project
package common

// Environment represents a game environment that can be interacted with
type Environment interface {
	// GetState returns the current state of the environment
	GetState() State

	// GetPotentialActions returns an array of possible actions for the given agent ID
	GetPotentialActions(agentID int) []Action

	// Act performs the given action for the specified agent and returns the reward
	// Reward values are between -1 and 1 inclusive
	Act(agentID int, action Action) (float64, error)

	// Evaluate returns the winning agent's ID, -1 for a draw, or 0 if the game should continue
	Evaluate() int

	// EvaluateAction returns 1 if action would result in a win for given agent,
	// -1 for a draw and zero if the game should continue
	EvaluateAction(agentID int, action Action) int

	// Reset restarts the environment
	Reset()
}

// State represents the state of an environment
type State interface{}

// Action represents an action that can be taken in an environment
type Action interface {
	// GetParams returns the parameters of the action
	GetParams() interface{}
}

// Agent represents a player that can interact with an environment
type Agent interface {
	// GetID returns the agent's ID
	GetID() int

	// FetchMessage returns agent's status message, if any
	FetchMessage() string

	// FetchMove returns the agent's move based on given state and set of actions
	FetchMove(state State, possibleActions []Action) (Action, error)

	// GameOver signals that the game is over with the given final state
	GameOver(state State)

	// GetSign returns the agent's display character
	GetSign() string
}