package agents

import (
	"encoding/gob"
	"fmt"
	"math/rand"
	"os"

	"mnkagent/common"
	"mnkagent/game"
)

// RLAgentKnowledge stores the Q-learning knowledge base
type RLAgentKnowledge struct {
	Values           map[string]float64
	Iterations       uint
	RandomDispersion []int
}

// SaveToFile writes the knowledge map to the given path
func (k *RLAgentKnowledge) SaveToFile(path string) (bool, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return false, fmt.Errorf("could not open writable knowledge file: %w", err)
	}
	defer file.Close()

	enc := gob.NewEncoder(file)
	err = enc.Encode(k)
	if err != nil {
		return false, fmt.Errorf("encoding of knowledge failed: %w", err)
	}

	return true, nil
}

// LoadFromFile reads the knowledge from the given path
func (k *RLAgentKnowledge) LoadFromFile(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("could not open readable knowledge file: %w", err)
	}
	defer file.Close()

	dec := gob.NewDecoder(file)
	err = dec.Decode(k)
	if err != nil {
		return false, fmt.Errorf("decoding of knowledge failed: %w", err)
	}

	return true, nil
}

// RLAgent implements a reinforcement learning agent
type RLAgent struct {
	// Agent identification
	id   int
	sign string

	// Game definition
	m, n, k       int
	environment   common.Environment

	// RL parameters
	Learning          bool
	LearningRate      float64 // alpha
	DiscountFactor    float64 // gamma
	ExplorationFactor float64 // epsilon

	// Knowledge base
	Knowledge *RLAgentKnowledge
	
	// State tracking
	prev struct {
		state  game.MNKState
		action game.MNKAction
		reward float64
	}
	message string
}

// NewRLAgent creates a new reinforcement learning agent
func NewRLAgent(id int, sign string, m, n, k int, environment common.Environment, knowledge *RLAgentKnowledge, learn bool) *RLAgent {
	agent := &RLAgent{
		id:               id,
		sign:             sign,
		m:                m,
		n:                n,
		k:                k,
		environment:      environment,
		Learning:         learn,
		LearningRate:     0.2,  // Default alpha
		DiscountFactor:   0.8,  // Default gamma
		ExplorationFactor: 0.25, // Default epsilon
		Knowledge:        knowledge,
	}

	// Initialize knowledge base if needed
	if knowledge.Values == nil {
		knowledge.Values = make(map[string]float64)
	}
	
	// Initialize random dispersion tracking if needed
	if knowledge.RandomDispersion == nil || len(knowledge.RandomDispersion) != m*n {
		oldDispersion := knowledge.RandomDispersion
		knowledge.RandomDispersion = make([]int, m*n)
		
		// Copy existing data if possible
		if oldDispersion != nil {
			copyLen := len(oldDispersion)
			if copyLen > m*n {
				copyLen = m*n
			}
			copy(knowledge.RandomDispersion, oldDispersion[:copyLen])
		}
	}

	return agent
}

// GetID returns the agent's ID
func (agent *RLAgent) GetID() int {
	return agent.id
}

// FetchMessage returns the agent's status message
func (agent *RLAgent) FetchMessage() string {
	message := agent.message
	agent.message = ""
	return message
}

// FetchMove determines the next move using the Q-learning algorithm
func (agent *RLAgent) FetchMove(state common.State, possibleActions []common.Action) (common.Action, error) {
	// Cast state to MNKState
	s := state.(game.MNKState)
	var action game.MNKAction
	var qMax float64

	// Exploration vs. exploitation decision
	e := rand.Float64()
	if e < agent.ExplorationFactor {
		// Exploration: Choose a random move
		agent.message = fmt.Sprintf("Exploratory action (%f)", e)
		rndi := rand.Intn(len(possibleActions))
		action = possibleActions[rndi].GetParams().(game.MNKAction)
		agent.Knowledge.RandomDispersion[action.Y*agent.m+action.X]++
		qMax = agent.lookup(s, action)
	} else {
		// Exploitation: Choose the best move
		agent.message = fmt.Sprintf("Greedy action (%f)", e)
		
		// Find the move with the highest expected value
		var first = true
		for i := range s {
			for j := range s[i] {
				if s[i][j] == 0 {
					a := game.MNKAction{Y: i, X: j}
					v := agent.lookup(s, a)

					if v > qMax || first {
						qMax = v
						action = a
						first = false
					}
				}
			}
		}
	}

	// Update Q-values if learning is enabled
	if agent.Learning {
		agent.learn(qMax)
	}

	// Save the current state and action for the next learning update
	agent.prev.state = s
	agent.prev.action = action
	agent.prev.reward = agent.value(s, action)

	return action, nil
}

// GameOver handles the end of the game
func (agent *RLAgent) GameOver(state common.State) {
	s := state.(game.MNKState)

	if agent.Learning {
		// Final learning update using terminal state
		agent.learn(agent.lookup(s, game.MNKAction{X: -1, Y: -1}))
	}

	// Reset state for next game
	agent.prev.state = game.MNKState{}
	agent.prev.action = game.MNKAction{}
	agent.prev.reward = 0
	agent.message = ""

	// Increment iteration counter
	agent.Knowledge.Iterations++
}

// GetSign returns the character representing this player on the board
func (agent *RLAgent) GetSign() string {
	return agent.sign
}

// learn updates Q-values based on the current state-action pair
func (agent *RLAgent) learn(qMax float64) {
	// Ignore empty state (happens on first move)
	if len(agent.prev.state) == 0 {
		return
	}

	// Get marshalled state representation
	mState := marshallState(agent.id, agent.prev.state, agent.prev.action)
	oldVal, exists := agent.Knowledge.Values[mState]

	// Apply Q-learning update formula: Q(s,a) = Q(s,a) + α * (r + γ * max(Q(s',a')) - Q(s,a))
	qValue := oldVal
	if exists {
		qValue = oldVal + (agent.LearningRate * 
			(agent.prev.reward + (agent.DiscountFactor * qMax) - oldVal))
	} else {
		qValue = agent.prev.reward
	}
	
	agent.Knowledge.Values[mState] = qValue
}

// lookup retrieves the Q-value for a state-action pair
func (agent *RLAgent) lookup(state game.MNKState, action game.MNKAction) float64 {
	mState := marshallState(agent.id, state, action)
	val, ok := agent.Knowledge.Values[mState]
	if !ok {
		val = agent.value(state, action)
		agent.Knowledge.Values[mState] = val
	}
	return val
}

// value calculates the immediate reward for a state-action pair
func (agent *RLAgent) value(_ game.MNKState, action game.MNKAction) float64 {
	// Special case for terminal state evaluation
	if action == (game.MNKAction{X: -1, Y: -1}) {
		switch agent.environment.Evaluate() {
		case agent.id: // Agent won
			return 1
		case 0: // Game continues
			return 0
		case -1: // Draw
			return -0.5
		default: // Agent lost
			return -1
		}
	}

	// Evaluate potential action
	switch agent.environment.EvaluateAction(agent.id, action) {
	case 1: // Would win
		return 1
	case 0: // Game continues
		return 0
	case -1: // Would end in draw
		return -0.5
	default: // Should never happen
		return 0
	}
}

// marshallState converts a board state and action to a string representation
func marshallState(agentID int, state game.MNKState, action game.MNKAction) string {
	var result string
	
	for i := range state {
		for j := range state[i] {
			// Include the action in the state representation
			if i == action.Y && j == action.X {
				result += "X"
				continue
			}

			switch state[i][j] {
			case 0:
				result += "-"
			case agentID:
				result += "X"
			default:
				result += "O"
			}
		}
	}
	
	return result
}