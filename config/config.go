// Package config provides configuration management for the mnkagent project
package config

import (
	"flag"
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

// Config contains all application configuration
type Config struct {
	Game GameConfig
	RL   RLConfig
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

	// RL flags
	flag.StringVar(&config.RL.ModelFile, "rl-model", "rl.kw", "RL trained model file location")
	flag.BoolVar(&config.RL.ModelStatusMode, "rl-model-status", false, "Show RL model status and exit")
	flag.BoolVar(&config.RL.NoLearn, "rl-no-learn", false, "Turn off learning for RL in normal mode and don't save model to disk")
	flag.UintVar(&config.RL.TrainingMode, "rl-train", 0, "Train RL for n iterations")

	flag.Parse()

	// Apply Gomoku settings if requested
	if config.Game.Gomoku {
		config.Game.M = 19
		config.Game.N = 19
		config.Game.K = 5
	}

	return config
}