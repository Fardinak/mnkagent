package main

import "errors"

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

func AnalyzeMNKState(state MNKState, m, n, k int) (ps MNKProbabilitySet) {
	if len(state) == 0 {
		return
	}

	for cx, cy, err := FindNonEmptyPosition(state, 0, 0); err == nil; cx, cy, err = FindNonEmptyPosition(state, cx, cy) {
		if cx+k <= m {
			p := MNKProbability{}
			p.Position = Position2D{cx, cy}
			p.State = make([]int, 0, k)
			for o := 0; o < k; o++ {
				p.State = append(p.State, state[cy][cx+o])
			}
			ps = append(ps, p)
		}

		if cy+k <= n {
			p := MNKProbability{}
			p.Position = Position2D{cx, cy}
			p.State = make([]int, 0, k)
			for o := 0; o < k; o++ {
				p.State = append(p.State, state[cy+o][cx])
			}
			ps = append(ps, p)
		}

		if cx+k <= m && cy+k <= n {
			p := MNKProbability{}
			p.Position = Position2D{cx, cy}
			p.State = make([]int, 0, k)
			for o := 0; o < k; o++ {
				p.State = append(p.State, state[cy+o][cx+o])
			}
			ps = append(ps, p)
		}

		if cx-k+1 >= 0 && cy+k <= n {
			p := MNKProbability{}
			p.Position = Position2D{cx, cy}
			p.State = make([]int, 0, k)
			for o := 0; o < k; o++ {
				p.State = append(p.State, state[cy+o][cx-o])
			}
			ps = append(ps, p)
		}
	}

	return
}

func FindNonEmptyPosition(state MNKState, offsetX, offsetY int) (x, y int, err error) {
	var (
		n = len(state)
		m = len(state[0])
	)

	for i := offsetY; i < n; i++ {
		j := 0
		if i == offsetY && offsetX != 0 {
			// Don't get trapped on one cell
			j = offsetX + 1
		}
		for ; j < m; j++ {
			if state[i][j] > 0 {
				return j, i, nil
			}
		}
	}

	return x, y, errors.New("classifier: no more empty positions available")
}
