package agents

import (
	"encoding/gob"
	"fmt"
	"math/rand"
	"os"
	"time"

	"mnkagent/common"
	"mnkagent/game"
)

// EnhancedRLAgent implements the EnhancedAgent interface with reinforcement learning
type EnhancedRLAgent struct {
	// Basic agent properties
	options common.AgentOptions
	stats   common.AgentStats
	
	// Game environment reference
	environment common.Environment
	
	// Board dimensions
	m, n, k int
	
	// Knowledge base
	Knowledge *RLAgentKnowledge
	
	// State tracking
	prev struct {
		state  game.MNKState
		action game.MNKAction
		reward float64
	}
	message string
	
	// Performance tracking
	moveEvaluationTimes []time.Duration
	decisionReasons     map[string]int
}

// NewEnhancedRLAgent creates a new enhanced RL agent
func NewEnhancedRLAgent(options common.AgentOptions, knowledge *RLAgentKnowledge) *EnhancedRLAgent {
	agent := &EnhancedRLAgent{
		options: options,
		stats: common.AgentStats{},
		Knowledge: knowledge,
		moveEvaluationTimes: make([]time.Duration, 0, 100),
		decisionReasons: make(map[string]int),
	}
	
	// Initialize knowledge base if needed
	if knowledge.Values == nil {
		knowledge.Values = make(map[string]float64)
	}
	
	return agent
}

// GetID returns the agent's ID
func (agent *EnhancedRLAgent) GetID() int {
	return agent.options.ID
}

// FetchMessage returns the agent's status message
func (agent *EnhancedRLAgent) FetchMessage() string {
	message := agent.message
	agent.message = ""
	return message
}

// FetchMove determines the next move using the Q-learning algorithm
func (agent *EnhancedRLAgent) FetchMove(state common.State, possibleActions []common.Action) (common.Action, error) {
	// Track performance metrics
	startTime := time.Now()
	defer func() {
		agent.moveEvaluationTimes = append(agent.moveEvaluationTimes, time.Since(startTime))
		agent.stats.TotalMoves++
	}()
	
	// Cast state to MNKState
	s := state.(game.MNKState)
	var action game.MNKAction
	var qMax float64
	var reason string

	// Exploration vs. exploitation decision
	e := rand.Float64()
	if e < agent.options.ExplorationFactor {
		// Exploration: Choose a random move
		agent.message = fmt.Sprintf("Exploratory action (%f)", e)
		rndi := rand.Intn(len(possibleActions))
		action = possibleActions[rndi].GetParams().(game.MNKAction)
		agent.Knowledge.RandomDispersion[action.Y*agent.m+action.X]++
		qMax = agent.lookup(s, action)
		
		// Track decision reason
		reason = "exploration"
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
		
		// Track decision reason
		reason = "exploitation"
	}
	
	// Track the decision reason
	agent.decisionReasons[reason]++

	// Update Q-values if learning is enabled
	if agent.options.IsLearner {
		agent.learn(qMax)
	}

	// Save the current state and action for the next learning update
	agent.prev.state = s
	agent.prev.action = action
	agent.prev.reward = agent.value(s, action)

	return action, nil
}

// GameOver handles the end of the game
func (agent *EnhancedRLAgent) GameOver(state common.State) {
	s := state.(game.MNKState)
	
	// Update statistics
	agent.stats.GamesPlayed++
	agent.stats.AverageMoves = float64(agent.stats.TotalMoves) / float64(agent.stats.GamesPlayed)
	
	// Update game outcome statistics
	result := agent.environment.Evaluate()
	switch result {
	case agent.options.ID:
		agent.stats.GamesWon++
	case -1:
		agent.stats.GamesDraw++
	case 0:
		// Game was interrupted, don't count
	default:
		agent.stats.GamesLost++
	}

	if agent.options.IsLearner {
		// Final learning update using terminal state
		agent.learn(agent.lookup(s, game.MNKAction{X: -1, Y: -1}))
		
		// Update learning stats
		agent.stats.TrainingEpisodes++
		agent.stats.KnownStates = len(agent.Knowledge.Values)
		
		// Simple heuristic for learning progress based on known states
		// A more sophisticated implementation would use learning curve metrics
		estimatedMaxStates := agent.m * agent.n * 3 * 10 // rough approximation
		agent.stats.LearningProgress = float64(agent.stats.KnownStates) / float64(estimatedMaxStates)
		if agent.stats.LearningProgress > 1.0 {
			agent.stats.LearningProgress = 1.0
		}
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
func (agent *EnhancedRLAgent) GetSign() string {
	return agent.options.Sign
}

// GetOptions returns the agent's configuration options
func (agent *EnhancedRLAgent) GetOptions() common.AgentOptions {
	return agent.options
}

// SetOptions updates the agent's configuration
func (agent *EnhancedRLAgent) SetOptions(options common.AgentOptions) error {
	// Validate option values
	if options.LearningRate < 0 || options.LearningRate > 1 {
		return fmt.Errorf("invalid learning rate: %f (must be between 0 and 1)", options.LearningRate)
	}
	if options.DiscountFactor < 0 || options.DiscountFactor > 1 {
		return fmt.Errorf("invalid discount factor: %f (must be between 0 and 1)", options.DiscountFactor)
	}
	if options.ExplorationFactor < 0 || options.ExplorationFactor > 1 {
		return fmt.Errorf("invalid exploration factor: %f (must be between 0 and 1)", options.ExplorationFactor)
	}
	
	// Apply valid options
	agent.options = options
	return nil
}

// GetCapabilities returns the agent's supported capabilities
func (agent *EnhancedRLAgent) GetCapabilities() common.AgentCapabilities {
	return common.Learning | common.StateExport | common.StateImport | common.Explainable
}

// Supports checks if the agent supports a specific capability
func (agent *EnhancedRLAgent) Supports(capability common.AgentCapabilities) bool {
	return (agent.GetCapabilities() & capability) == capability
}

// GetStats returns the agent's performance statistics
func (agent *EnhancedRLAgent) GetStats() common.AgentStats {
	return agent.stats
}

// ResetStats clears the agent's statistics
func (agent *EnhancedRLAgent) ResetStats() {
	agent.stats = common.AgentStats{}
	agent.moveEvaluationTimes = make([]time.Duration, 0, 100)
	agent.decisionReasons = make(map[string]int)
}

// SaveState persists the agent's state to a file
func (agent *EnhancedRLAgent) SaveState(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create state file: %w", err)
	}
	defer file.Close()
	
	// Create state snapshot
	stateSnapshot := struct {
		Options     common.AgentOptions
		Stats       common.AgentStats
		Knowledge   RLAgentKnowledge
		BoardParams struct {
			M, N, K int
		}
	}{
		Options:   agent.options,
		Stats:     agent.stats,
		Knowledge: *agent.Knowledge,
	}
	stateSnapshot.BoardParams.M = agent.m
	stateSnapshot.BoardParams.N = agent.n
	stateSnapshot.BoardParams.K = agent.k
	
	// Encode state to file
	enc := gob.NewEncoder(file)
	err = enc.Encode(stateSnapshot)
	if err != nil {
		return fmt.Errorf("failed to encode agent state: %w", err)
	}
	
	return nil
}

// LoadState loads the agent's state from a file
func (agent *EnhancedRLAgent) LoadState(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open state file: %w", err)
	}
	defer file.Close()
	
	// Define state structure
	var stateSnapshot struct {
		Options     common.AgentOptions
		Stats       common.AgentStats
		Knowledge   RLAgentKnowledge
		BoardParams struct {
			M, N, K int
		}
	}
	
	// Decode state from file
	dec := gob.NewDecoder(file)
	err = dec.Decode(&stateSnapshot)
	if err != nil {
		return fmt.Errorf("failed to decode agent state: %w", err)
	}
	
	// Update agent with loaded state
	agent.options = stateSnapshot.Options
	agent.stats = stateSnapshot.Stats
	*agent.Knowledge = stateSnapshot.Knowledge
	agent.m = stateSnapshot.BoardParams.M
	agent.n = stateSnapshot.BoardParams.N
	agent.k = stateSnapshot.BoardParams.K
	
	return nil
}

// ExplainMove provides an explanation of why the agent chose a particular move
func (agent *EnhancedRLAgent) ExplainMove(state common.State, action common.Action) string {
	s := state.(game.MNKState)
	a := action.GetParams().(game.MNKAction)
	
	// Get Q-value for this state-action
	qValue := agent.lookup(s, a)
	
	// Get immediate reward
	reward := agent.value(s, a)
	
	// Count neighboring pieces
	neighbors := countNeighbors(s, a, agent.options.ID)
	
	explanation := fmt.Sprintf("Move (%d,%d) has Q-value: %.3f\n", a.X, a.Y, qValue)
	explanation += fmt.Sprintf("Immediate reward: %.1f\n", reward)
	
	// Add interpretation
	if qValue > 0.7 {
		explanation += "This move has a high chance of leading to a win.\n"
	} else if qValue > 0.3 {
		explanation += "This move has a moderate chance of success.\n"
	} else if qValue > 0 {
		explanation += "This move is slightly favorable.\n"
	} else if qValue > -0.3 {
		explanation += "This move is neutral or slightly unfavorable.\n"
	} else {
		explanation += "This move is likely to lead to a loss.\n"
	}
	
	// Add context about the board position
	explanation += fmt.Sprintf("Position has %d friendly neighbors and %d opponent neighbors.\n", 
		neighbors.friendly, neighbors.opponent)
	
	return explanation
}

// Initialize prepares the agent for a new set of games
func (agent *EnhancedRLAgent) Initialize(environment common.Environment) error {
	agent.environment = environment
	
	// Extract board dimensions
	switch env := environment.(type) {
	case *game.MNKBoard:
		agent.m = env.GetWidth()
		agent.n = env.GetHeight()
		agent.k = env.GetWinLength()
	case *game.MNKBitboard:
		agent.m = env.GetWidth()
		agent.n = env.GetHeight()
		agent.k = env.GetWinLength()
	default:
		return fmt.Errorf("unsupported environment type: %T", environment)
	}
	
	// Initialize random dispersion tracking if needed
	if agent.Knowledge.RandomDispersion == nil || len(agent.Knowledge.RandomDispersion) != agent.m*agent.n {
		oldDispersion := agent.Knowledge.RandomDispersion
		agent.Knowledge.RandomDispersion = make([]int, agent.m*agent.n)
		
		// Copy existing data if possible
		if oldDispersion != nil {
			copyLen := min(len(oldDispersion), agent.m*agent.n)
			copy(agent.Knowledge.RandomDispersion[:copyLen], oldDispersion[:copyLen])
		}
	}
	
	return nil
}

// Cleanup releases resources when agent is no longer needed
func (agent *EnhancedRLAgent) Cleanup() error {
	// Auto-save if model file is specified
	if agent.options.ModelFile != "" && agent.options.IsLearner {
		return agent.SaveState(agent.options.ModelFile)
	}
	return nil
}

// Helper functions

// learn updates Q-values based on the current state-action pair
func (agent *EnhancedRLAgent) learn(qMax float64) {
	// Ignore empty state (happens on first move)
	if len(agent.prev.state) == 0 {
		return
	}

	// Get marshalled state representation
	mState := marshallState(agent.options.ID, agent.prev.state, agent.prev.action)
	oldVal, exists := agent.Knowledge.Values[mState]

	// Apply Q-learning update formula: Q(s,a) = Q(s,a) + α * (r + γ * max(Q(s',a')) - Q(s,a))
	qValue := oldVal
	if exists {
		qValue = oldVal + (agent.options.LearningRate * 
			(agent.prev.reward + (agent.options.DiscountFactor * qMax) - oldVal))
	} else {
		qValue = agent.prev.reward
	}
	
	agent.Knowledge.Values[mState] = qValue
}

// lookup retrieves the Q-value for a state-action pair
func (agent *EnhancedRLAgent) lookup(state game.MNKState, action game.MNKAction) float64 {
	mState := marshallState(agent.options.ID, state, action)
	val, ok := agent.Knowledge.Values[mState]
	if !ok {
		val = agent.value(state, action)
		agent.Knowledge.Values[mState] = val
	}
	return val
}

// value calculates the immediate reward for a state-action pair
func (agent *EnhancedRLAgent) value(_ game.MNKState, action game.MNKAction) float64 {
	// Special case for terminal state evaluation
	if action == (game.MNKAction{X: -1, Y: -1}) {
		switch agent.environment.Evaluate() {
		case agent.options.ID: // Agent won
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
	switch agent.environment.EvaluateAction(agent.options.ID, action) {
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

// Helper to count neighbors (used for move explanation)
type neighborCount struct {
	friendly int
	opponent int
}

func countNeighbors(state game.MNKState, action game.MNKAction, playerID int) neighborCount {
	result := neighborCount{}
	x, y := action.X, action.Y
	
	// Check all 8 adjacent positions
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue // Skip center position
			}
			
			nx, ny := x+dx, y+dy
			
			// Check bounds
			if ny < 0 || ny >= len(state) || nx < 0 || nx >= len(state[0]) {
				continue
			}
			
			// Count by player
			if state[ny][nx] == playerID {
				result.friendly++
			} else if state[ny][nx] != 0 {
				result.opponent++
			}
		}
	}
	
	return result
}

// Helper to get the minimum of two values
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}