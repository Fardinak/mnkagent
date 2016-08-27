package main

import (
	"errors"
	"math"
)

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

func (s MNKState) GetMN() (int, int) {
	return len(s[0]), len(s)
}

func (s MNKState) GetBucket(pos Position2D, dir Direction) (b MNKBucket, err error) {
	// Step coefficients based on direction
	i, j := dir.GetCoefficients()

	for o := 0; o < k; o++ {
		b.bucket = append(b.bucket, s[pos.Y+o*j][pos.X+o*i])
	}

	return
}

func (s MNKState) GetAllBuckets() (bs []MNKBucket) {
	if len(s) == 0 {
		return
	}

	for pos, err := s.NextNonEmptyPosition(0, 0); err == nil; pos, err = s.NextNonEmptyPosition(pos.X, pos.Y) {
		// If need be, possible to automate like so:
		// var directions = []Direction{DIRECTION_E, DIRECTION_S, DIRECTION_SE, DIRECTION_SW}
		// for dir := range directions {
		//	var i, j = dir.GetCoefficients()
		//	if pos.X+k*i+1 >= 0 && pos.X+k*i <= m && pos.Y+k*j+1 >= 0 && pos.Y+k*j <= n {
		//		b, _ := s.GetBucket(pos, dir)
		//		bs = append(bs, b)
		//	}
		// }

		if pos.X+k <= m {
			b, _ := s.GetBucket(pos, DIRECTION_E)
			bs = append(bs, b)
		}

		if pos.Y+k <= n {
			b, _ := s.GetBucket(pos, DIRECTION_S)
			bs = append(bs, b)
		}

		if pos.X+k <= m && pos.Y+k <= n {
			b, _ := s.GetBucket(pos, DIRECTION_SE)
			bs = append(bs, b)
		}

		if pos.X-k+1 >= 0 && pos.Y+k <= n {
			b, _ := s.GetBucket(pos, DIRECTION_SW)
			bs = append(bs, b)
		}
	}

	// TODO: Create a bucket cache and define a trigger to add new moves
	return
}

func (s MNKState) NextNonEmptyPosition(offsetX, offsetY int) (Position2D, error) {
	m, n := s.GetMN()

	for i := offsetY; i < n; i++ {
		j := 0
		if i == offsetY && offsetX != 0 {
			// Don't get trapped on one cell
			j = offsetX + 1
		}
		for ; j < m; j++ {
			if s[i][j] > 0 {
				return Position2D{j, i}, nil
			}
		}
	}

	return Position2D{}, errors.New("environment: no more empty positions available")
}

func (s MNKState) EvaluateAction(action MNKAction) float64 {
	// TODO: Get Relevant Buckets and calculate action's relative position in each
	// Remember to use the bucket cache to provide meaningful reward
	return 0
}

type MNKBucket struct {
	bucket    []int
	Position  Position2D
	Direction Direction
}

func (b MNKBucket) Evaluate() (p MNKBucketScore) {
	var k = len(b.bucket)
	var count = [3]float64{}

	for i := 0; i < k; i++ {
		count[b.bucket[i]]++
	}

	p.XScore = count[1] / float64(k)
	p.OScore = count[2] / float64(k)

	if p.OScore > 0 {
		p.XScore = 0
	}
	if p.XScore > 0 {
		p.OScore = 0
	}

	return
}

func (b MNKBucket) EvaluateAction(agentID, action int) (MNKBucketScore, error) {
	if action < 0 || action >= len(b.bucket) {
		return MNKBucketScore{}, errors.New("environment: action not in range")
	}
	// Clone Bucket
	var bucket = MNKBucket{}
	copy(bucket.bucket, b.bucket)

	// Evaluate state+action bucket
	b.bucket[action] = agentID
	return b.Evaluate(), nil
}

func (b MNKBucket) EvaluateRelativeAction(agentID int, action Action) (MNKBucketScore, error) {
	act := action.GetParams().(MNKAction)
	pos := b.Position
	i, j := b.Direction.GetCoefficients()

	// Validate action
	var (
		X0, Y0 = pos.X, pos.Y
		X1, Y1 = pos.X + k*i, pos.Y + k*j

		dxa = act.X - X0
		dya = act.Y - Y0

		dxl = X1 - X0
		dyl = Y1 - Y0

		// Gradient check
		cross = dxa*dyl - dya*dxl
		err   = cross != 0
	)

	// Range check
	if math.Abs(float64(dxl)) >= math.Abs(float64(dyl)) {
		if dxl > 0 {
			if X0 > act.X || act.X > X1 {
				err = true
			}
		} else {
			if X1 > act.X || act.X > X0 {
				err = true
			}
		}
	} else {
		if dyl > 0 {
			if Y0 > act.Y || act.Y > Y1 {
				err = true
			}
		} else {
			if Y1 > act.Y || act.Y > Y0 {
				err = true
			}
		}
	}

	if err {
		return MNKBucketScore{}, errors.New("environment: action not in range")
	}

	// Calculate relative action position
	a := int(math.Max(math.Abs(float64(pos.X-act.X)), math.Abs(float64(pos.Y-act.Y))))

	return b.EvaluateAction(agentID, a)
}

type MNKBucketScore struct {
	XScore float64
	OScore float64
}

type MNKAction struct {
	// TODO: Use Position2D in here
	Y, X int
}

func (a MNKAction) GetParams() interface{} {
	return a
}

type Direction int

const (
	DIRECTION_N = iota
	DIRECTION_NE
	DIRECTION_E
	DIRECTION_SE
	DIRECTION_S
	DIRECTION_SW
	DIRECTION_W
	DIRECTION_NW
)

func (d Direction) GetCoefficients() (i, j int) {
	switch d {
	case DIRECTION_N:
		return 0, -1
	case DIRECTION_NE:
		return 1, -1
	case DIRECTION_E:
		return 1, 0
	case DIRECTION_SE:
		return 1, 1
	case DIRECTION_S:
		return 0, 1
	case DIRECTION_SW:
		return -1, 1
	case DIRECTION_W:
		return -1, 0
	case DIRECTION_NW:
		return -1, -1
	}
	return
}

type Position2D struct {
	X, Y int
}

type MNKProbability struct {
	State       []int
	Position    Position2D
	XLikelyhood float64
	OLikelyhood float64
}
type MNKProbabilitySet []MNKProbability
