package common

// AgentStats tracks performance statistics for an agent
type AgentStats struct {
	// Game performance metrics
	GamesPlayed  int
	GamesWon     int
	GamesLost    int
	GamesDraw    int
	TotalMoves   int
	AverageMoves float64
	
	// Learning metrics (for learning agents)
	TrainingEpisodes int
	LearningProgress float64 // 0.0 to 1.0 indicating training progress
	KnownStates      int
}

// AgentOptions provides configuration for agents
type AgentOptions struct {
	// Basic agent configuration
	ID        int
	Name      string
	Sign      string
	IsLearner bool
	
	// Learning parameters (for learning agents)
	LearningRate      float64 // Alpha: step size
	DiscountFactor    float64 // Gamma: future reward discount
	ExplorationFactor float64 // Epsilon: exploration vs exploitation
	ModelFile         string  // Path to save/load the agent model
}

// AgentCapabilities defines special features an agent can support
type AgentCapabilities int

const (
	None AgentCapabilities = 0
	Learning AgentCapabilities = 1 << iota
	StateImport
	StateExport
	InterfaceHints
	Explainable
	DynamicDifficulty
	ParallelEvaluation
)

// EnhancedAgent extends the basic Agent interface with additional
// capabilities like configuration, state management, and statistics
type EnhancedAgent interface {
	Agent // Embed the original Agent interface
	
	// Configuration capabilities
	GetOptions() AgentOptions
	SetOptions(options AgentOptions) error
	
	// Capabilities information
	GetCapabilities() AgentCapabilities
	Supports(capability AgentCapabilities) bool
	
	// Statistics and performance tracking
	GetStats() AgentStats
	ResetStats()
	
	// State persistence and loading
	SaveState(path string) error
	LoadState(path string) error
	
	// Decision explanation (for explainable agents)
	ExplainMove(state State, action Action) string
	
	// Initialize prepares the agent for a new set of games
	Initialize(environment Environment) error
	
	// Cleanup releases resources when agent is no longer needed
	Cleanup() error
}