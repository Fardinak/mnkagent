package main

import "testing"

var testBoard = MNKState{
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 1, 1, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 2, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 1, 2, 2, 2, 1, 1, 2, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

var (
	testM = 19
	testN = 19
	testK = 5

	testNonEmptyCellCount = 25
	testBucketCount       = 95  // Total Buckets Count: (m-k+1)*n + (n-k+1)*m + Max(0, Min(m, n)-k+1)*2 + (m-k)(m-k+1) + (n-k)(n-k+1)
	testBucketStateCount  = 162 // Total States Count: 2*3^(k-1)
)

func TestAnalyzeMNKState(t *testing.T) {
	probabilitySet := AnalyzeMNKState(testBoard, testM, testN, testK)

	// Assert set length (bucket count)
	if len(probabilitySet) != testBucketCount {
		t.Errorf("AnalyzeMNKState(): Expected %d buckets, actual %d", testBucketCount, len(probabilitySet))
		for i, p := range probabilitySet {
			t.Logf("%d %v", i+1, p.State)
		}
	}

	// Assert learned states count
	if len(rlKnowledge.Values) != testBucketStateCount {
		t.Logf("WARNING: AnalyzeMNKState(): Expected %d learned states, actual %d", testBucketStateCount, len(rlKnowledge.Values))
	}
}

func TestFindNonEmptyPosition(t *testing.T) {
	nonEmptyCells := make([][2]int, 0, testM*testN)

	cx, cy, err := FindNonEmptyPosition(testBoard, 0, 0)
	for err == nil {
		nonEmptyCells = append(nonEmptyCells, [2]int{cx, cy})
		cx, cy, err = FindNonEmptyPosition(testBoard, cx, cy)
	}

	if len(nonEmptyCells) != testNonEmptyCellCount {
		t.Errorf("FindNonEmptyPosition(): Expected %d buckets, actual %d", testNonEmptyCellCount, len(nonEmptyCells))
		for i, p := range nonEmptyCells {
			t.Logf("%d (%d, %d)", i+1, p[0], p[1])
		}
		t.Log(err)
	}
}

func BenchmarkAnalyzeMNKState(b *testing.B) {
	for n := 0; n < b.N; n++ {
		AnalyzeMNKState(testBoard, testM, testN, testK)
	}
}
