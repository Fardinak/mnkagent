package main

import "testing"

var _ Environment = (*MNKBoard)(nil)

var b, _ = NewMNKBoard(3, 3, 3)
var agentID = 1

var ActionEvaluationTable = []struct {
	board    [][]int
	action   [2]int
	expected int
}{
	// NOTE: The following covers all winning states (expected = 1), but not otherwise
	// Top-Left to Bottom-Right
	{[][]int{{0, 0, 0}, {0, 1, 0}, {0, 0, 1}}, [2]int{0, 0}, 1},
	{[][]int{{1, 0, 0}, {0, 0, 0}, {0, 0, 1}}, [2]int{1, 1}, 1},
	{[][]int{{1, 0, 0}, {0, 1, 0}, {0, 0, 0}}, [2]int{2, 2}, 1},
	// Top-Right to Bottom-Left
	{[][]int{{0, 0, 0}, {0, 1, 0}, {1, 0, 0}}, [2]int{2, 0}, 1},
	{[][]int{{0, 0, 1}, {0, 0, 0}, {1, 0, 0}}, [2]int{1, 1}, 1},
	{[][]int{{0, 0, 1}, {0, 1, 0}, {0, 0, 0}}, [2]int{0, 2}, 1},
	// Row One
	{[][]int{{0, 1, 1}, {0, 0, 0}, {0, 0, 0}}, [2]int{0, 0}, 1},
	{[][]int{{1, 0, 1}, {0, 0, 0}, {0, 0, 0}}, [2]int{1, 0}, 1},
	{[][]int{{1, 1, 0}, {0, 0, 0}, {0, 0, 0}}, [2]int{2, 0}, 1},
	// Row Two
	{[][]int{{0, 0, 0}, {0, 1, 1}, {0, 0, 0}}, [2]int{0, 1}, 1},
	{[][]int{{0, 0, 0}, {1, 0, 1}, {0, 0, 0}}, [2]int{1, 1}, 1},
	{[][]int{{0, 0, 0}, {1, 1, 0}, {0, 0, 0}}, [2]int{2, 1}, 1},
	// Row Three
	{[][]int{{0, 0, 0}, {0, 0, 0}, {0, 1, 1}}, [2]int{0, 2}, 1},
	{[][]int{{0, 0, 0}, {0, 0, 0}, {1, 0, 1}}, [2]int{1, 2}, 1},
	{[][]int{{0, 0, 0}, {0, 0, 0}, {1, 1, 0}}, [2]int{2, 2}, 1},
	// Col One
	{[][]int{{0, 0, 0}, {1, 0, 0}, {1, 0, 0}}, [2]int{0, 0}, 1},
	{[][]int{{1, 0, 0}, {0, 0, 0}, {1, 0, 0}}, [2]int{0, 1}, 1},
	{[][]int{{1, 0, 0}, {1, 0, 0}, {0, 0, 0}}, [2]int{0, 2}, 1},
	// Col Two
	{[][]int{{0, 0, 0}, {0, 1, 0}, {0, 1, 0}}, [2]int{1, 0}, 1},
	{[][]int{{0, 1, 0}, {0, 0, 0}, {0, 1, 0}}, [2]int{1, 1}, 1},
	{[][]int{{0, 1, 0}, {0, 1, 0}, {0, 0, 0}}, [2]int{1, 2}, 1},
	// Col Three
	{[][]int{{0, 0, 0}, {0, 0, 1}, {0, 0, 1}}, [2]int{2, 0}, 1},
	{[][]int{{0, 0, 1}, {0, 0, 0}, {0, 0, 1}}, [2]int{2, 1}, 1},
	{[][]int{{0, 0, 1}, {0, 0, 1}, {0, 0, 0}}, [2]int{2, 2}, 1},

	// Expected 0
	{[][]int{{0, 0, 0}, {0, 1, 0}, {0, 0, 1}}, [2]int{0, 1}, 0},
	{[][]int{{1, 0, 0}, {0, 0, 0}, {0, 0, 1}}, [2]int{2, 1}, 0},
	{[][]int{{1, 0, 0}, {0, 1, 0}, {0, 0, 0}}, [2]int{2, 0}, 0},
	{[][]int{{0, 0, 0}, {0, 1, 0}, {1, 0, 0}}, [2]int{1, 0}, 0},
	{[][]int{{0, 0, 1}, {0, 0, 0}, {1, 0, 0}}, [2]int{0, 0}, 0},
	{[][]int{{0, 0, 1}, {0, 1, 0}, {0, 0, 0}}, [2]int{2, 2}, 0},
	{[][]int{{0, 1, 1}, {0, 0, 0}, {0, 0, 0}}, [2]int{2, 1}, 0},
	{[][]int{{1, 0, 1}, {0, 0, 0}, {0, 0, 0}}, [2]int{1, 2}, 0},
	{[][]int{{1, 1, 0}, {0, 0, 0}, {0, 0, 0}}, [2]int{2, 2}, 0},
}

func TestEvaluateAction(t *testing.T) {
	for _, a := range ActionEvaluationTable {
		b.board = a.board

		r := b.EvaluateAction(agentID, MNKAction{X: a.action[0], Y: a.action[1]})

		if r != a.expected {
			t.Errorf("EvaluateAction(%d %d): Expected %d for state(%s), "+
				"actual %d", a.action[0], a.action[1], a.expected, b.GetState(agentID), r)
		} else {
			t.Logf("EvaluateAction(%d %d): As expected %d for state(%s)",
				a.action[0], a.action[1], r, b.GetState(agentID))
		}
	}
}
