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
	for i, c := 0, 1; i < b.m; i, c = i+1, 1 {
		for j := 0; j < b.n-1; j++ {
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
	for j, c := 0, 1; j < b.n; j, c = j+1, 1 {
		for i := 0; i < b.m-1; i++ {
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
	for o, c := 0, 1; o <= b.n-b.k; o, c = o+1, 1 {
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
	for o, c := 1, 1; o <= b.m-b.k; o, c = o+1, 1 {
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
	for o, c := 0, 1; o <= b.n-b.k; o, c = o+1, 1 {
		for i := 0; i < b.m-o-1 && i < b.n-o-1; i++ {
			if b.board[i][b.n-o-i-1] == b.board[i+1][b.n-o-i-2] {
				c++
				if c >= b.k && b.board[i][b.n-o-i-1] > 0 {
					return b.board[i][b.n-o-i-1]
				}
			} else {
				c = 1
			}
		}
	}

	// TR-BL lower
	for o, c := 1, 1; o <= b.m-b.k; o, c = o+1, 1 {
		for i := 0; i < b.m-o-1 && i < b.n-o-1; i++ {
			if b.board[i+o][b.n-i-1] == b.board[i+o+1][b.n-i-2] {
				c++
				if c >= b.k && b.board[i+o][b.n-i-1] > 0 {
					return b.board[i+o][b.n-i-1]
				}
			} else {
				c = 1
			}
		}
	}

	// Continuity check
	for i := 0; i < b.m; i++ {
		for j := 0; j < b.n; j++ {
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

	var (
		row      bool = b.k <= b.m
		col      bool = b.k <= b.n
		diagonal bool = row && col

		doneT, doneB, doneL, doneR     bool
		doneTL, doneTR, doneBL, doneBR bool
		doneOrthogonal, doneDiagonal   bool = false, !diagonal

		countRow, countCol, countTLBR, countTRBL int
	)

	// REVIEW: Benchmark agains multiple loops i.e. Evaluate()
	for o := 0; !doneDiagonal || !doneOrthogonal; o++ {
		if !doneDiagonal {
			// To bottom-right
			if !doneBR && a.Y+o < b.n-1 && a.X+o < b.m-1 {
				if b.board[a.Y+o+1][a.X+o+1] == agentID {
					countTLBR++
				} // REVIEW: else if x > n - k && y > m - k { doneBR = true }
			} else {
				doneBR = true
			}

			// To top-left
			if !doneTL && a.Y-o > 0 && a.X-o > 0 {
				if b.board[a.Y-o-1][a.X-o-1] == agentID {
					countTLBR++
				}
			} else {
				doneTL = true
			}

			// To bottom-left
			if !doneBL && a.Y+o < b.n-1 && a.X-o > 0 {
				if b.board[a.Y+o+1][a.X-o-1] == agentID {
					countTRBL++
				}
			} else {
				doneBL = true
			}

			// To top-right
			if !doneTR && a.Y-o > 0 && a.X+o < b.m-1 {
				if b.board[a.Y-o-1][a.X+o+1] == agentID {
					countTRBL++
				}
			} else {
				doneTR = true
			}

			doneDiagonal = doneTL && doneTR && doneBL && doneBR

			if countTLBR >= b.k-1 || countTRBL >= b.k-1 {
				// Win
				return 1
			}
		}

		// To bottom
		if !doneB && col && a.Y+o < b.n-1 {
			if b.board[a.Y+o+1][a.X] == agentID {
				countCol++
			}
		} else {
			doneB = true
		}

		// To top
		if !doneT && col && a.Y-o > 0 {
			if b.board[a.Y-o-1][a.X] == agentID {
				countCol++
			}
		} else {
			doneT = true
		}

		// To right
		if !doneR && row && a.X+o < b.m-1 {
			if b.board[a.Y][a.X+o+1] == agentID {
				countRow++
			}
		} else {
			doneR = true
		}

		// To left
		if !doneL && row && a.X-o > 0 {
			if b.board[a.Y][a.X-o-1] == agentID {
				countRow++
			}
		} else {
			doneL = true
		}

		doneOrthogonal = doneT && doneB && doneL && doneR

		if countRow >= b.k-1 || countCol >= b.k-1 {
			// Win
			return 1
		}
	}

	// Continuity check
	for i := 0; i < b.m; i++ {
		for j := 0; j < b.n; j++ {
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
