package game

import (
	"testing"
)

// Prepare test cases for different board sizes
var benchmarkSizes = []struct {
	m, n, k int
	name    string
}{
	{3, 3, 3, "TicTacToe"},
	{9, 9, 5, "Gomoku-Small"},
	{15, 15, 5, "Gomoku-Medium"},
	{19, 19, 5, "Gomoku-Full"},
}

// Generate a standard test pattern for each board
func generateTestPattern(m, n int) [][]int {
	// Create a diagonal pattern with some common game scenarios
	board := make([][]int, n)
	for i := range board {
		board[i] = make([]int, m)
	}

	// Add a diagonal of player 1
	for i := 0; i < min(m, n, 5); i++ {
		board[i][i] = 1
	}

	// Add some player 2 marks
	for i := 0; i < min(m, n-1, 4); i++ {
		board[i][i+1] = 2
	}

	return board
}

// Helper function to get minimum of three values
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// Benchmark original board implementation (EvaluateAction)
func BenchmarkOriginalBoard_EvaluateAction(b *testing.B) {
	for _, size := range benchmarkSizes {
		b.Run(size.name, func(b *testing.B) {
			board, _ := NewMNKBoard(size.m, size.n, size.k)
			testPattern := generateTestPattern(size.m, size.n)
			
			// Apply test pattern to board
			for y := range testPattern {
				for x := range testPattern[y] {
					if testPattern[y][x] > 0 {
						board.board[y][x] = testPattern[y][x]
					}
				}
			}
			
			// Define a test action
			action := MNKAction{X: 5 % size.m, Y: 5 % size.n}
			agentID := 1
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				board.EvaluateAction(agentID, action)
			}
		})
	}
}

// Benchmark original board implementation (Evaluate)
func BenchmarkOriginalBoard_Evaluate(b *testing.B) {
	for _, size := range benchmarkSizes {
		b.Run(size.name, func(b *testing.B) {
			board, _ := NewMNKBoard(size.m, size.n, size.k)
			testPattern := generateTestPattern(size.m, size.n)
			
			// Apply test pattern to board
			for y := range testPattern {
				for x := range testPattern[y] {
					if testPattern[y][x] > 0 {
						board.board[y][x] = testPattern[y][x]
					}
				}
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				board.Evaluate()
			}
		})
	}
}

// Benchmark bitmap board implementation (EvaluateAction)
func BenchmarkBitmapBoard_EvaluateAction(b *testing.B) {
	for _, size := range benchmarkSizes {
		// Skip large boards that don't fit in 64 bits
		if size.m * size.n > 64 {
			continue
		}
		
		b.Run(size.name, func(b *testing.B) {
			board, _ := NewMNKBitboard(size.m, size.n, size.k)
			testPattern := generateTestPattern(size.m, size.n)
			
			// Apply test pattern to bitmap board
			for y := range testPattern {
				for x := range testPattern[y] {
					if testPattern[y][x] > 0 {
						pos := y*size.m + x
						board.board.PlayerBits[testPattern[y][x]] |= 1 << pos
						board.moveCount++
					}
				}
			}
			
			// Define a test action
			action := MNKAction{X: 5 % size.m, Y: 5 % size.n}
			agentID := 1
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				board.EvaluateAction(agentID, action)
			}
		})
	}
}

// Benchmark bitmap board implementation (Evaluate)
func BenchmarkBitmapBoard_Evaluate(b *testing.B) {
	for _, size := range benchmarkSizes {
		// Skip large boards that don't fit in 64 bits
		if size.m * size.n > 64 {
			continue
		}
		
		b.Run(size.name, func(b *testing.B) {
			board, _ := NewMNKBitboard(size.m, size.n, size.k)
			testPattern := generateTestPattern(size.m, size.n)
			
			// Apply test pattern to bitmap board
			for y := range testPattern {
				for x := range testPattern[y] {
					if testPattern[y][x] > 0 {
						pos := y*size.m + x
						board.board.PlayerBits[testPattern[y][x]] |= 1 << pos
						board.moveCount++
						board.lastMove.playerID = testPattern[y][x]
						board.lastMove.x = x
						board.lastMove.y = y
					}
				}
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				board.Evaluate()
			}
		})
	}
}