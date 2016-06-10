package main

import "errors"

type Environment interface {
	// GetState returns the current state of the environment
	// It requires the agentID to provide pov
	GetState(int) State

	// GetActions returns an array of possible actions
	GetActions(int) []Action

	// Act performs the given action and returns designated reward where -1 <= r <= 1
	Act(int, Action) (float64, error)

	// Evaluate returns a number representing the state of the episode
	// 1: Agent Succeeded, 0: Episode Incomplete, -1: Agent Failed
	Evaluate(int) float64

	// EvaluateAction returns possible reward for given action without modifying the environment
	EvaluateAction(int, Action) int

	// Reset restarts the environment
	Reset()
}

type State interface{}

type Action interface {
	GetParams() interface{}
}

type MNKBoard struct {
	m, n, k int
	board   MNKState
}

func NewMNKBoard(m, n, k int) (b *MNKBoard, err error) {
	if k > m && k > n {
		return nil, errors.New("environment: k exceeds both m and n")
	}

	b = new(MNKBoard)

	b.m = m
	b.n = n
	b.k = k

	b.board = make([][]int, n)
	for i := range b.board {
		b.board[i] = make([]int, m)
	}

	return
}

func (b *MNKBoard) GetState(agentID int) State {
	var s MNKState = b.board.Clone()

	// Create populate the board with 0: empty, 1: agent's, -1: rival's
	for i := range s {
		for j := range s[i] {
			// Regulate based on given agent ID
			if s[i][j] > 0 {
				if s[i][j] == agentID {
					s[i][j] = 1
				} else {
					s[i][j] = -1
				}
			}
		}
	}
	return s
}

func (b *MNKBoard) GetActions(agentID int) (a []Action) {
	for i := range b.board {
		for j := range b.board[i] {
			if b.board[i][j] == 0 {
				a = append(a, MNKAction{
					X: j,
					Y: i,
				})
			}
		}
	}
	return
}

func (b *MNKBoard) Act(agentID int, a Action) (r float64, err error) {
	aa := a.GetParams().(MNKAction)
	if aa.X < 0 || aa.X >= b.m || aa.Y < 0 || aa.Y >= b.n {
		return r, errors.New("environment: move out of range")
	}

	if board[aa.Y][aa.X] != 0 {
		return r, errors.New("environment: invalid move")
	}

	board[aa.Y][aa.X] = agentID
	return b.Evaluate(agentID), nil
}

func (b *MNKBoard) Evaluate(agentID int) float64 {
	panic("environment.Evaluate: Not Implemented.")
}

func (b *MNKBoard) EvaluateAction(agentID int, a Action) int {
	aa := a.GetParams().(MNKAction)

	var (
		row      bool = b.k <= b.m
		col      bool = b.k <= b.n
		diagonal bool = row && col

		doneT, doneB, doneL, doneR     bool
		doneTL, doneTR, doneBL, doneBR bool
		doneOrthogonal, doneDiagonal   bool = false, !diagonal

		countRow, countCol, countTLBR, countTRBL int
	)

	for o := 0; !doneDiagonal || !doneOrthogonal; o++ {
		if !doneDiagonal {
			// To bottom-right
			if !doneBR && aa.Y+o < b.n-1 && aa.X+o < b.m-1 && b.board[aa.Y+o+1][aa.X+o+1] == agentID {
				countTLBR++
			} else {
				doneBR = true
			}

			// To top-left
			if !doneTL && aa.Y-o > 0 && aa.X-o > 0 && b.board[aa.Y-o-1][aa.X-o-1] == agentID {
				countTLBR++
			} else {
				doneTL = true
			}

			// To bottom-left
			if !doneBL && aa.Y+o < b.n-1 && aa.X-o > 0 && b.board[aa.Y+o+1][aa.X-o-1] == agentID {
				countTRBL++
			} else {
				doneBL = true
			}

			// To top-right
			if !doneTR && aa.Y-o > 0 && aa.X+o < b.m-1 && b.board[aa.Y-o-1][aa.X+o+1] == agentID {
				countTRBL++
			} else {
				doneTR = true
			}

			doneDiagonal = doneTL && doneTR && doneBL && doneBR

			if countTLBR >= b.k-1 || countTRBL >= b.k-1 {
				return 1
			}
		}

		// To bottom
		if !doneB && col && aa.Y+o < b.n-1 && b.board[aa.Y+o+1][aa.X] == agentID {
			countCol++
		} else {
			doneB = true
		}

		// To top
		if !doneT && col && aa.Y-o > 0 && b.board[aa.Y-o-1][aa.X] == agentID {
			countCol++
		} else {
			doneT = true
		}

		// To right
		if !doneR && row && aa.X+o < b.m-1 && b.board[aa.Y][aa.X+o+1] == agentID {
			countRow++
		} else {
			doneR = true
		}

		// To left
		if !doneL && row && aa.X-o > 0 && b.board[aa.Y][aa.X-o-1] == agentID {
			countRow++
		} else {
			doneL = true
		}

		doneOrthogonal = doneT && doneB && doneL && doneR

		if countRow >= b.k-1 || countCol >= b.k-1 {
			return 1
		}
	}

	return 0
}

func (b *MNKBoard) Reset() {
	b, _ = NewMNKBoard(b.m, b.n, b.k)
}

type MNKState [][]int

func (s MNKState) Clone() (sp MNKState) {
	sp = make([][]int, len(s))
	for i := range s {
		sp[i] = make([]int, len(s[i]))
		copy(sp[i], s[i])
	}
	return
}

type MNKAction struct {
	X, Y int
}

func (a MNKAction) GetParams() interface{} {
	return a
}
