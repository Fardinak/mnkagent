// Package config provides configuration management for the mnkagent project
package config

import (
	"flag"
	"fmt"
)

// GameConfig contains game-related configuration
type GameConfig struct {
	M         int  // Board width
	N         int  // Board height
	K         int  // Number of marks in a row needed to win
	NoDisplay bool // Don't show board in training mode
	Gomoku    bool // Use Gomoku settings (19x19 board, 5 in a row)
	Rounds    int  // Number of rounds to play
}

// RLConfig contains reinforcement learning configuration
type RLConfig struct {
	ModelFile       string // File path for the RL model
	ModelStatusMode bool   // Display model status and exit
	NoLearn         bool   // Disable learning
	TrainingMode    uint   // Number of training iterations
}

// DQNConfig contains Deep Q-Network configuration
type DQNConfig struct {
	ModelFile       string // File path for the DQN model
	BatchSize       int    // Batch size for training
	UpdateFrequency int    // How often to update the network
	ReplaySize      int    // Size of experience replay buffer
	HiddenSize      int    // Size of hidden layer in neural network
	NoLearn         bool   // Disable learning
}

// Config contains all application configuration
type Config struct {
	Game     GameConfig
	RL       RLConfig
	DQN      DQNConfig
	AgentType string    // Type of agent to use ("rl" or "dqn")
	NoLearn   bool      // Global flag to disable learning for all agent types
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate board dimensions
	if c.Game.M <= 0 {
		return fmt.Errorf("invalid board width (m): %d - must be positive", c.Game.M)
	}
	if c.Game.N <= 0 {
		return fmt.Errorf("invalid board height (n): %d - must be positive", c.Game.N)
	}
	if c.Game.K <= 0 {
		return fmt.Errorf("invalid win condition (k): %d - must be positive", c.Game.K)
	}
	
	// Validate k in relation to board dimensions
	if c.Game.K > c.Game.M && c.Game.K > c.Game.N {
		return fmt.Errorf("win condition (k=%d) exceeds both board width (m=%d) and height (n=%d)", 
			c.Game.K, c.Game.M, c.Game.N)
	}
	
	// Validate agent type
	if c.AgentType != "rl" && c.AgentType != "dqn" {
		return fmt.Errorf("invalid agent type: %s - must be 'rl' or 'dqn'", c.AgentType)
	}
	
	// Validate RL configuration if RL agent is selected
	if c.AgentType == "rl" {
		// Validate model file path if not in no-learn mode
		if !c.RL.NoLearn && c.RL.ModelFile == "" {
			return fmt.Errorf("RL model file path cannot be empty when learning is enabled")
		}
	}
	
	// Validate DQN configuration if DQN agent is selected
	if c.AgentType == "dqn" {
		// Validate model file path if not in no-learn mode
		if !c.DQN.NoLearn && c.DQN.ModelFile == "" {
			return fmt.Errorf("DQN model file path cannot be empty when learning is enabled")
		}
		
		// Validate batch size
		if c.DQN.BatchSize <= 0 {
			return fmt.Errorf("invalid batch size: %d - must be positive", c.DQN.BatchSize)
		}
		
		// Validate update frequency
		if c.DQN.UpdateFrequency <= 0 {
			return fmt.Errorf("invalid update frequency: %d - must be positive", c.DQN.UpdateFrequency)
		}
		
		// Validate replay buffer size
		if c.DQN.ReplaySize <= 0 {
			return fmt.Errorf("invalid replay buffer size: %d - must be positive", c.DQN.ReplaySize)
		}
		
		// Validate hidden layer size
		if c.DQN.HiddenSize <= 0 {
			return fmt.Errorf("invalid hidden layer size: %d - must be positive", c.DQN.HiddenSize)
		}
	}
	
	return nil
}

// LoadFromArgs parses command-line arguments into configuration
func LoadFromArgs() *Config {
	config := &Config{}

	// Game flags
	flag.IntVar(&config.Game.M, "m", 3, "Board dimension across the horizontal (x) axis")
	flag.IntVar(&config.Game.N, "n", 3, "Board dimension across the vertical (y) axis")
	flag.IntVar(&config.Game.K, "k", 3, "Number of marks in a row needed to win")
	flag.BoolVar(&config.Game.NoDisplay, "no-display", false, "Do not show board and stats in training mode")
	flag.BoolVar(&config.Game.Gomoku, "gomoku", false, "Shortcut for a 19,19,5 game (overrides m, n and k)")

	// Agent type selection
	flag.StringVar(&config.AgentType, "agent", "rl", "Type of agent to use (rl or dqn)")

	// Global learning flag
	flag.BoolVar(&config.NoLearn, "no-learn", false, "Disable learning for all agent types")

	// RL flags
	flag.StringVar(&config.RL.ModelFile, "rl-model", "rl.kw", "RL trained model file location")
	flag.BoolVar(&config.RL.ModelStatusMode, "rl-model-status", false, "Show RL model status and exit")
	flag.BoolVar(&config.RL.NoLearn, "rl-no-learn", false, "Turn off learning for RL in normal mode and don't save model to disk")
	flag.UintVar(&config.RL.TrainingMode, "rl-train", 0, "Train RL for n iterations")

	// DQN flags
	flag.StringVar(&config.DQN.ModelFile, "dq-model", "dqn.kw", "DQN trained model file location")
	flag.IntVar(&config.DQN.BatchSize, "dq-batch-size", 32, "Batch size for DQN training")
	flag.IntVar(&config.DQN.UpdateFrequency, "dq-update-freq", 4, "How often to update the DQN network")
	flag.IntVar(&config.DQN.ReplaySize, "dq-replay-size", 10000, "Size of experience replay buffer")
	flag.IntVar(&config.DQN.HiddenSize, "dq-hidden-size", 128, "Size of hidden layer in neural network")
	flag.BoolVar(&config.DQN.NoLearn, "dq-no-learn", false, "Turn off learning for DQN in normal mode")

	flag.Parse()

	// Apply Gomoku settings if requested
	if config.Game.Gomoku {
		config.Game.M = 19
		config.Game.N = 19
		config.Game.K = 5
	}

	return config
}