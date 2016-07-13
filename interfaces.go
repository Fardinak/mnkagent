package main

// Environment interface
type Environment interface {
	// GetState returns the current state of the environment
	GetState() State

	// GetPotentialActions returns an array of possible actions
	GetPotentialActions(int) []Action

	// Act performs the given action and returns designated reward where -1 <= r <= 1
	Act(int, Action) (float64, error)

	// Evaluate returns a the winning agent's id, or -1 for a draw, otherwise
	// zero if game should go on
	Evaluate() int

	// EvaluateAction returns 1 if action would result in a win for given agent,
	// -1 for a draw and zero otherwise
	EvaluateAction(int, Action) int

	// Reset restarts the environment
	Reset()
}

// Environment State interface
type State interface{}

// Environment Action interface
type Action interface {
	GetParams() interface{}
}

// Agent interface
type Agent interface {
	// FetchMessage returns agent's messages, if any
	FetchMessage() string

	// FetchMove returns the agent's move based on given state and set of actions
	FetchMove(State, []Action) (Action, error)

	// GameOver states that the game is over and that the latest state should be saved
	GameOver(State)

	// GetSign returns the agent's sign (X|O)
	GetSign() string
}
