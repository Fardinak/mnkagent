package common

import (
	"math/rand"
	
	"mnkagent/game"
)

// Experience represents a single training example (s, a, r, s')
type Experience struct {
	State       game.MNKState
	Action      game.MNKAction
	Reward      float64
	NextState   game.MNKState
	Terminal    bool
}

// ExperienceBuffer implements an experience replay buffer
type ExperienceBuffer struct {
	Buffer    []Experience
	Capacity  int
	Size      int
	Index     int
}

// NewExperienceBuffer creates a new experience replay buffer
func NewExperienceBuffer(capacity int) *ExperienceBuffer {
	return &ExperienceBuffer{
		Buffer:   make([]Experience, capacity),
		Capacity: capacity,
		Size:     0,
		Index:    0,
	}
}

// Add adds a new experience to the buffer
func (eb *ExperienceBuffer) Add(exp Experience) {
	// Store in buffer using circular array approach
	eb.Buffer[eb.Index] = exp
	eb.Index = (eb.Index + 1) % eb.Capacity
	
	// Update size counter
	if eb.Size < eb.Capacity {
		eb.Size++
	}
}

// Sample randomly samples a batch of experiences from the buffer
func (eb *ExperienceBuffer) Sample(batchSize int) []Experience {
	// Adjust batch size if buffer has fewer items
	if batchSize > eb.Size {
		batchSize = eb.Size
	}
	
	// No sampling possible from empty buffer
	if batchSize == 0 {
		return nil
	}
	
	// Sample without replacement
	samples := make([]Experience, batchSize)
	indices := make([]int, eb.Size)
	
	// Initialize indices
	for i := 0; i < eb.Size; i++ {
		indices[i] = i
	}
	
	// Sample random indices
	for i := 0; i < batchSize; i++ {
		// Choose random index from remaining indices
		r := rand.Intn(len(indices) - i)
		
		// Get the experience at this index
		if indices[r] < eb.Size {
			samples[i] = eb.Buffer[indices[r]]
		}
		
		// Remove the selected index by swapping with the last index
		indices[r], indices[len(indices)-i-1] = indices[len(indices)-i-1], indices[r]
	}
	
	return samples
}

// Clear empties the buffer
func (eb *ExperienceBuffer) Clear() {
	eb.Size = 0
	eb.Index = 0
}