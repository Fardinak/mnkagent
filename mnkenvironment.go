package main

import "errors"

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

	// Initialize the board
	b.Reset()

	return
}

func (b *MNKBoard) GetState() State {
	return b.board.Clone()
}

func (b *MNKBoard) GetPotentialActions(agentID int) (a []Action) {
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

func (b *MNKBoard) Act(agentID int, action Action) (r float64, err error) {
	a := action.GetParams().(MNKAction)
	if a.X < 0 || a.X >= b.m || a.Y < 0 || a.Y >= b.n {
		return 0, errors.New("environment: move out of range")
	}

	if b.board[a.Y][a.X] != 0 {
		return 0, errors.New("environment: invalid move")
	}

	b.board[a.Y][a.X] = agentID
	switch b.EvaluateAction(agentID, action) {
	case 1: // Won
		return 1, nil
	case 0: // Continue
		return 0, nil
	case -1: // Draw
		return -0.5, nil
	}

	// Never happens
	return 0, nil
}

func (b *MNKBoard) Evaluate() int {
	// Rows
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

	// Columns
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

	// TL-BR upper
	for o, c := 0, 1; o <= b.m-b.k; o, c = o+1, 1 {
		for i := 0; i < b.m-o-1 && i < b.n-o-1; i++ {
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

	// TL-BR lower
	for o, c := 1, 1; o <= b.n-b.k; o, c = o+1, 1 {
		for i := 0; i < b.m-o-1 && i < b.n-o-1; i++ {
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

	// TR-BL upper
	for o, c := 0, 1; o <= b.m-b.k; o, c = o+1, 1 {
		for i := 0; i < b.m-o-1 && i < b.n-o-1; i++ {
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

	// TR-BL lower
	for o, c := 1, 1; o <= b.n-b.k; o, c = o+1, 1 {
		for i := 0; i < b.m-o-1 && i < b.n-o-1; i++ {
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

	// Continuity check
	for i := 0; i < b.n; i++ {
		for j := 0; j < b.m; j++ {
			if b.board[i][j] == 0 {
				return 0
			}
		}
	}

	// Draw
	return -1
}

func (b *MNKBoard) EvaluateAction(agentID int, action Action) int {
	a := action.GetParams().(MNKAction)

	// Row
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

	// Column
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

	// TL-BR
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

	// TR-BL
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

	// Continuity check
	for i := 0; i < b.n; i++ {
		for j := 0; j < b.m; j++ {
			if b.board[i][j] == 0 {
				return 0
			}
		}
	}

	// Draw
	return -1
}

func (b *MNKBoard) Reset() {
	b.board = make([][]int, n)
	for i := range b.board {
		b.board[i] = make([]int, m)
	}
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
	Y, X int
}

func (a MNKAction) GetParams() interface{} {
	return a
}
