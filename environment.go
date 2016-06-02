package main

import "errors"

type Environment interface {
	// GetState returns the current state of the environment
	// It requires the agentID to provide pov
	GetState(int) interface{}

	// GetActions returns an array of possible actions
	GetActions(int) []interface{}

	// Act performs the given action and returns designated reward where -1 <= r <= 1
	Act(int, interface{}) (float64, error)

	// Evaluate returns a number representing the state of the episode
	// 1: Agent Succeeded, 0: Episode Incomplete, -1: Agent Failed
	Evaluate(int) int

	// EvaluateAction returns possible reward for given action without modifying the environment
	EvaluateAction(int, interface{}) int
}

type MNKBoard struct {
	board [][]int

	M, N, K int
}

type MNKAction struct {
	X, Y int
}

func (b *MNKBoard) GetState(agentID int) (s string) {
	for i := range b.board {
		for j := range b.board[i] {
			// Regulate based on current agent's ID so all agents are one!
			if b.board[i][j] > 0 {
				if b.board[i][j] == agentID {
					s += "X"
				} else {
					s += "O"
				}
			} else {
				s += "-"
			}
		}
	}
	return
}

func (b *MNKBoard) GetActions(agentID int) (a []MNKAction) {
	for i := range b.board {
		for j := range b.board[i] {
			if b.board[i][j] == 0 {
				a = append(a, MNKAction{X: j, Y: i})
			}
		}
	}
	return
}

func (b *MNKBoard) Act(agentID int, a MNKAction) (r float64, err error) {
	if a.X < 0 || a.X >= b.M || a.Y < 0 || a.Y >= b.N {
		return r, errors.New("environment: move out of range")
	}

	if board[a.Y][a.X] != 0 {
		return r, errors.New("environment: invalid move")
	}

	board[a.Y][a.X] = agentID
	switch b.Evaluate(agentID) {
	case agentID:
		return 1, nil
	case 0:
		return 0, nil
	case -1:
		return -0.5, nil
	default:
		return -1, nil
	}
}

func (b *MNKBoard) Evaluate(agentID int) int {
	return evaluate(b.board)
}

func (b *MNKBoard) EvaluateAction(agentID int, a MNKAction) int {
	return evaluate(b.board)
}
