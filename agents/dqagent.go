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

// DQNAgent implements Deep Q-Network reinforcement learning agent
type DQNAgent struct {
	// Basic agent properties
	options common.AgentOptions
	stats   common.AgentStats
	
	// Game environment reference
	environment common.Environment
	
	// Board dimensions
	m, n, k int
	
	// Knowledge base
	Knowledge *RLAgentKnowledge
	
	// Neural network for value approximation
	ValueNetwork *NeuralNetwork
	
	// Experience replay buffer
	ReplayBuffer *ExperienceBuffer
	
	// State tracking
	prev struct {
		state  game.MNKState
		action game.MNKAction
		reward float64
		nextState game.MNKState
		terminal bool
	}
	message string
	
	// Performance tracking
	moveEvaluationTimes []time.Duration
	decisionReasons     map[string]int
	
	// Neural network and replay configuration
	batchSize int
	updateFrequency int
	stepCounter int
}

// NewDQNAgent creates a new Deep Q-Network agent
func NewDQNAgent(options common.AgentOptions, knowledge *RLAgentKnowledge) *DQNAgent {
	agent := &DQNAgent{
		options: options,
		stats: common.AgentStats{},
		Knowledge: knowledge,
		moveEvaluationTimes: make([]time.Duration, 0, 100),
		decisionReasons: make(map[string]int),
		batchSize: 32,              // Default batch size for training
		updateFrequency: 4,         // Update network every 4 steps
	}
	
	// Initialize knowledge base if needed
	if knowledge.Values == nil {
		knowledge.Values = make(map[string]float64)
	}
	
	// Initialize experience replay buffer with capacity of 10000
	agent.ReplayBuffer = NewExperienceBuffer(10000)
	
	return agent
}

// GetID returns the agent's ID
func (agent *DQNAgent) GetID() int {
	return agent.options.ID
}

// FetchMessage returns the agent's status message
func (agent *DQNAgent) FetchMessage() string {
	message := agent.message
	agent.message = ""
	return message
}

// FetchMove determines the next move using the Q-learning algorithm
func (agent *DQNAgent) FetchMove(state common.State, possibleActions []common.Action) (common.Action, error) {
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

	// Get the immediate reward for this state-action pair
	immediateReward := agent.value(s, action)
	
	// Save the current state and action for the next learning update
	agent.prev.state = s
	agent.prev.action = action
	agent.prev.reward = immediateReward
	agent.prev.terminal = false // Will be updated in GameOver if needed

	return action, nil
}

// GameOver handles the end of the game
func (agent *DQNAgent) GameOver(state common.State) {
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
		// Mark the current state as terminal for experience replay
		agent.prev.nextState = s
		agent.prev.terminal = true
		
		// Final learning update using terminal state
		agent.learn(agent.lookup(s, game.MNKAction{X: -1, Y: -1}))
		
		// Add final experience to replay buffer if available
		if agent.ReplayBuffer != nil {
			// Calculate terminal state reward
			terminalReward := 0.0
			switch result {
			case agent.options.ID: // Agent won
				terminalReward = 1.0
			case -1: // Draw
				terminalReward = -0.5
			case 0: // Game interrupted
				terminalReward = 0.0
			default: // Agent lost
				terminalReward = -1.0
			}
			
			// Add terminal experience
			terminalExp := Experience{
				State:     agent.prev.state,
				Action:    agent.prev.action,
				Reward:    terminalReward,
				NextState: s,
				Terminal:  true,
			}
			agent.ReplayBuffer.Add(terminalExp)
			
			// Train on a batch if enough experiences are available
			if agent.ValueNetwork != nil && agent.ReplayBuffer.Size >= agent.batchSize {
				batch := agent.ReplayBuffer.Sample(agent.batchSize)
				agent.trainOnBatch(batch)
			}
		}
		
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
	agent.prev.nextState = game.MNKState{}
	agent.prev.terminal = false
	agent.message = ""

	// Increment iteration counter
	agent.Knowledge.Iterations++
}

// GetSign returns the character representing this player on the board
func (agent *DQNAgent) GetSign() string {
	return agent.options.Sign
}

// GetOptions returns the agent's configuration options
func (agent *DQNAgent) GetOptions() common.AgentOptions {
	return agent.options
}

// SetOptions updates the agent's configuration
func (agent *DQNAgent) SetOptions(options common.AgentOptions) error {
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
func (agent *DQNAgent) GetCapabilities() common.AgentCapabilities {
	return common.Learning | common.StateExport | common.StateImport | common.Explainable
}

// Supports checks if the agent supports a specific capability
func (agent *DQNAgent) Supports(capability common.AgentCapabilities) bool {
	return (agent.GetCapabilities() & capability) == capability
}

// GetStats returns the agent's performance statistics
func (agent *DQNAgent) GetStats() common.AgentStats {
	return agent.stats
}

// ResetStats clears the agent's statistics
func (agent *DQNAgent) ResetStats() {
	agent.stats = common.AgentStats{}
	agent.moveEvaluationTimes = make([]time.Duration, 0, 100)
	agent.decisionReasons = make(map[string]int)
}

// SaveState persists the agent's state to a file
func (agent *DQNAgent) SaveState(path string) error {
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
func (agent *DQNAgent) LoadState(path string) error {
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
func (agent *DQNAgent) ExplainMove(state common.State, action common.Action) string {
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
func (agent *DQNAgent) Initialize(environment common.Environment) error {
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
	
	// Initialize neural network if not already initialized
	if agent.ValueNetwork == nil {
		// Input size: flattened board state (m*n*3 for three possible states: empty, player, opponent)
		inputSize := agent.m * agent.n * 3
		
		// Create a neural network with one hidden layer
		hiddenSize := 128 // This can be adjusted based on board size and complexity
		outputSize := 1   // Single output representing the value of the state
		
		agent.ValueNetwork = NewNeuralNetwork(inputSize, hiddenSize, outputSize, agent.options.LearningRate)
	}
	
	// Reset step counter
	agent.stepCounter = 0
	
	return nil
}

// SetBatchSize sets the batch size for training
func (agent *DQNAgent) SetBatchSize(size int) {
	agent.batchSize = size
}

// SetUpdateFrequency sets how often to update the network
func (agent *DQNAgent) SetUpdateFrequency(freq int) {
	agent.updateFrequency = freq
}

// Cleanup releases resources when agent is no longer needed
func (agent *DQNAgent) Cleanup() error {
	// Auto-save if model file is specified
	if agent.options.ModelFile != "" && agent.options.IsLearner {
		return agent.SaveState(agent.options.ModelFile)
	}
	return nil
}

// Helper functions

// learn updates Q-values based on the current state-action pair
func (agent *DQNAgent) learn(qMax float64) {
	// Ignore empty state (happens on first move)
	if len(agent.prev.state) == 0 {
		return
	}

	// Store experience in replay buffer if it's available
	if agent.ReplayBuffer != nil && len(agent.prev.nextState) > 0 {
		experience := Experience{
			State:     agent.prev.state,
			Action:    agent.prev.action,
			Reward:    agent.prev.reward,
			NextState: agent.prev.nextState,
			Terminal:  agent.prev.terminal,
		}
		agent.ReplayBuffer.Add(experience)
	}

	// Increment step counter
	agent.stepCounter++

	// Train neural network periodically if it's available
	if agent.ValueNetwork != nil && agent.ReplayBuffer != nil && 
	   agent.stepCounter % agent.updateFrequency == 0 && 
	   agent.ReplayBuffer.Size >= agent.batchSize {
		// Sample batch from replay buffer
		batch := agent.ReplayBuffer.Sample(agent.batchSize)
		agent.trainOnBatch(batch)
	}

	// Also perform traditional Q-learning update
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
func (agent *DQNAgent) lookup(state game.MNKState, action game.MNKAction) float64 {
	// If neural network is available, use it for value approximation
	if agent.ValueNetwork != nil {
		// Convert state to neural network input format
		inputs := agent.boardToInput(state, action)
		
		// Get prediction from neural network
		outputs, err := agent.ValueNetwork.Predict(inputs)
		if err == nil && len(outputs) > 0 {
			// Neural network outputs value between 0-1, rescale to [-1,1]
			return outputs[0]*2 - 1
		}
		// Fall back to table lookup if neural network fails
	}
	
	// Traditional table-based lookup
	mState := marshallState(agent.options.ID, state, action)
	val, ok := agent.Knowledge.Values[mState]
	if !ok {
		val = agent.value(state, action)
		agent.Knowledge.Values[mState] = val
	}
	return val
}

// value calculates the immediate reward for a state-action pair
func (agent *DQNAgent) value(_ game.MNKState, action game.MNKAction) float64 {
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

// trainOnBatch trains the neural network on a batch of experiences
func (agent *DQNAgent) trainOnBatch(batch []Experience) {
	// Skip if batch is empty
	if len(batch) == 0 {
		return
	}

	// Process each experience in the batch
	for _, experience := range batch {
		// Get current state-action value
		currentInputs := agent.boardToInput(experience.State, experience.Action)
		
		// Calculate target value
		targetValue := experience.Reward
		
		// If not terminal state, add discounted future value
		if !experience.Terminal {
			// Find max Q value for next state
			var maxQ float64
			var first = true
			
			// Check all possible actions from the next state
			for i := range experience.NextState {
				for j := range experience.NextState[i] {
					if experience.NextState[i][j] == 0 {
						nextAction := game.MNKAction{Y: i, X: j}
						q := agent.lookup(experience.NextState, nextAction)
						
						if q > maxQ || first {
							maxQ = q
							first = false
						}
					}
				}
			}
			
			// Add discounted future value
			targetValue += agent.options.DiscountFactor * maxQ
		}
		
		// Scale target value from [-1,1] to [0,1] for neural network
		targetValue = (targetValue + 1) / 2
		targets := []float64{targetValue}
		
		// Train the neural network
		_ = agent.ValueNetwork.Train(currentInputs, targets)
	}
}

// boardToInput converts a board state to neural network input format
func (agent *DQNAgent) boardToInput(state game.MNKState, action game.MNKAction) []float64 {
	// Create input vector with one-hot encoding for each cell
	// For each cell: [1,0,0] = empty, [0,1,0] = player, [0,0,1] = opponent
	inputSize := agent.m * agent.n * 3
	inputs := make([]float64, inputSize)
	
	// Apply the action to the state to get the resulting state
	stateCopy := make(game.MNKState, len(state))
	for i := range state {
		stateCopy[i] = make([]int, len(state[i]))
		copy(stateCopy[i], state[i])
	}
	
	// Apply action if it's a valid one
	if action.X >= 0 && action.Y >= 0 && 
	   action.Y < len(stateCopy) && action.X < len(stateCopy[0]) && 
	   stateCopy[action.Y][action.X] == 0 {
		stateCopy[action.Y][action.X] = agent.options.ID
	}
	
	// Convert the state to input format
	for i := 0; i < agent.n; i++ {
		for j := 0; j < agent.m; j++ {
			baseIdx := (i*agent.m + j) * 3
			if i < len(stateCopy) && j < len(stateCopy[i]) {
				switch stateCopy[i][j] {
				case 0: // Empty cell
					inputs[baseIdx] = 1.0
				case agent.options.ID: // Agent's piece
					inputs[baseIdx+1] = 1.0
				default: // Opponent's piece
					inputs[baseIdx+2] = 1.0
				}
			} else {
				// Default to empty if out of bounds
				inputs[baseIdx] = 1.0
			}
		}
	}
	
	return inputs
}