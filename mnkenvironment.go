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

func (b *MNKBoard) GetState(agentID int) State {
	// TODO: Remove POV from GetState
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

func (b *MNKBoard) GetWorld() State {
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
	switch b.Evaluate() {
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

func (b *MNKBoard) Evaluate() int {
	// REVIEW: Using State Strings and Regular Expressions
	// pattern := regexp.MustCompile(fmt.Sprintf("((X[-OX|]{%d}){%d}X)|((X[-OX|]{%d}){%d}X)|((X[-OX|]{%d}){%d}X)|(\|[-OX]*X{%d}[-OX]*\|)", b.n, b.k-1, b.n+1, b.k-1, b.n-1, b.k-1, b.k))
	// result := pattern.MatchString(state.String())

	var (
		nRange int = b.n - b.k + 2
		mRange int = b.m - b.k + 2

		countRow   int = 1
		countCol   int = 1
		countTLBRU int = 1
		countTLBRL int = 1
		countTRBLU int = 1
		countTRBLL int = 1
	)

	var max int = b.m
	if b.n > max {
		max = b.n
	}

	// REVIEW: This is a horror show, isn't it?!
	for o := 0; o < max; o++ {
		for i, oi := 0, o; i < max; i, oi = i+1, oi+1 { // oi: offset + i
			// Row (o: Y, i: X)
			if o < b.m && i < nRange && b.board[o][i] == b.board[o][i+1] {
				countRow++
				if countRow >= b.k {
					return b.board[o][i]
				}
			} else {
				countRow = 1
			}

			// Col (o: X, i: Y)
			if o < b.n && i < mRange && b.board[i][o] == b.board[i+1][o] {
				countCol++
				if countCol >= b.k {
					return b.board[i][o]
				}
			} else {
				countCol = 1
			}

			// Top-Left to Bottom-Right upper half (o: Offset, i: Y, o+i: X)
			if i < mRange && oi < nRange && b.board[i][oi] == b.board[i+1][oi+1] {
				countTLBRU++
				if countTLBRU >= b.k {
					return b.board[i][oi]
				}
			} else {
				countTLBRU = 1
			}

			// Top-Left to Bottom-Right lower half (o: Offset, i: Y, o+i: X)
			if oi < mRange && i < nRange && b.board[oi][i] == b.board[oi+1][i+1] { // o > 0 feels optional!
				countTLBRL++
				if countTLBRL >= b.k {
					return b.board[oi][i]
				}
			} else {
				countTLBRL = 1
			}

			// Top-Right to Bottom-Left upper half (o: Offset, o+i: Y, b.n-o-i: X)
			if i < mRange && b.n-oi > 1 && b.board[i][b.n-oi-1] == b.board[i+1][b.n-oi-2] {
				countTRBLU++
				if countTRBLU >= b.k {
					return b.board[i][b.n-oi]
				}
			} else {
				countTRBLU = 1
			}

			// Top-Right to Bottom-Left lower half (o: Offset, o-i: Y, b.n-o-i: X)
			if oi < mRange && i < b.n && b.board[oi][b.n-i-1] == b.board[oi+1][b.n-i-2] { // o > 0 feels optional!
				countTRBLL++
				if countTRBLL >= b.k {
					return b.board[oi][b.n-i]
				}
			} else {
				countTRBLL = 1
			}
		}

		// Reset counters
		countRow = 1
		countCol = 1
		countTLBRU = 1
		countTLBRL = 1
		countTRBLU = 1
		countTRBLL = 1
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

	// REVIEW: This is a horror show, isn't it? Really! We could just use multiple loops
	for o := 0; !doneDiagonal || !doneOrthogonal; o++ {
		if !doneDiagonal {
			// To bottom-right
			if !doneBR && a.Y+o < b.n-1 && a.X+o < b.m-1 && b.board[a.Y+o+1][a.X+o+1] == agentID {
				countTLBR++
			} else {
				doneBR = true
			}

			// To top-left
			if !doneTL && a.Y-o > 0 && a.X-o > 0 && b.board[a.Y-o-1][a.X-o-1] == agentID {
				countTLBR++
			} else {
				doneTL = true
			}

			// To bottom-left
			if !doneBL && a.Y+o < b.n-1 && a.X-o > 0 && b.board[a.Y+o+1][a.X-o-1] == agentID {
				countTRBL++
			} else {
				doneBL = true
			}

			// To top-right
			if !doneTR && a.Y-o > 0 && a.X+o < b.m-1 && b.board[a.Y-o-1][a.X+o+1] == agentID {
				countTRBL++
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
		if !doneB && col && a.Y+o < b.n-1 && b.board[a.Y+o+1][a.X] == agentID {
			countCol++
		} else {
			doneB = true
		}

		// To top
		if !doneT && col && a.Y-o > 0 && b.board[a.Y-o-1][a.X] == agentID {
			countCol++
		} else {
			doneT = true
		}

		// To right
		if !doneR && row && a.X+o < b.m-1 && b.board[a.Y][a.X+o+1] == agentID {
			countRow++
		} else {
			doneR = true
		}

		// To left
		if !doneL && row && a.X-o > 0 && b.board[a.Y][a.X-o-1] == agentID {
			countRow++
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
