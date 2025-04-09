// The implementation of an m,n,k-game with swappable Agents
package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"mnkagent/agents"
	"mnkagent/common"
	"mnkagent/config"
	"mnkagent/game"
	"mnkagent/ui"
)

// Signs for players
const (
	X = "\033[36;1mX\033[0m"
	O = "\033[31;1mO\033[0m"
)

func main() {
	fmt.Println("MNK Agent v2")

	// Load configuration
	cfg := config.LoadFromArgs()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Initialize random seed
	rand.Seed(time.Now().UTC().UnixNano())

	// Create game board with optimal implementation
	board, err := game.CreateBoard(game.Auto, cfg.Game.M, cfg.Game.N, cfg.Game.K)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Initialize RL knowledge
	rlKnowledge := &agents.RLAgentKnowledge{}
	readKnowledgeOK, err := rlKnowledge.LoadFromFile(cfg.RL.ModelFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Notice: No existing model file found at %s. A new one will be created.\n", cfg.RL.ModelFile)
		} else {
			fmt.Printf("Warning: Could not load RL model: %v\n", err)
		}
	}

	// If model status mode is enabled, show stats and exit
	if cfg.RL.ModelStatusMode {
		if !readKnowledgeOK {
			fmt.Println("No readable model file found")
			return
		}

		fmt.Println("Reinforcement learning model report")
		fmt.Printf("Iterations: %d\n", rlKnowledge.Iterations)
		fmt.Printf("Learned states: %d\n", len(rlKnowledge.Values))
		
		var max, min float64
		for _, v := range rlKnowledge.Values {
			if v > max {
				max = v
			} else if v < min {
				min = v
			}
		}
		
		fmt.Printf("Maximum value: %f\n", max)
		fmt.Printf("Minimum value: %f\n", min)
		return
	}

	// Create agent map
	agentMap := make(map[int]common.Agent)
	agentMap[0] = nil // No one (used for empty spaces)

	// Create display handler
	display := ui.NewDisplay(ui.DisplayConfig{
		NoDisplay: cfg.Game.NoDisplay,
		M:         cfg.Game.M,
		N:         cfg.Game.N,
		FirstRun:  true,
	}, map[int]string{
		1: X,
		2: O,
	})

	// Setup for training mode
	if cfg.RL.TrainingMode > 0 {
		// Register SIGINT handler for clean termination
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		terminateFlag := false
		
		go func() {
			<-sigint
			terminateFlag = true
			signal.Reset(os.Interrupt)
		}()
		defer close(sigint)

		// Setup agents for training based on selected agent type
		switch cfg.AgentType {
		case "dqn":
			// Create DQN agents for training (both agents must be same type)
			dqnOptions1 := common.AgentOptions{
				ID:               1,
				Sign:             X,
				IsLearner:        true,
				LearningRate:     0.01,
				DiscountFactor:   0.95,
				ExplorationFactor: 0.25,
				ModelFile:        cfg.DQN.ModelFile,
			}
			
			dqnOptions2 := common.AgentOptions{
				ID:               2,
				Sign:             O,
				IsLearner:        true,
				LearningRate:     0.01,
				DiscountFactor:   0.95,
				ExplorationFactor: 0.25,
				ModelFile:        cfg.DQN.ModelFile,
			}
			
			p1 := agents.NewDQNAgent(dqnOptions1, rlKnowledge)
			p1.SetBatchSize(cfg.DQN.BatchSize)
			p1.SetUpdateFrequency(cfg.DQN.UpdateFrequency)
			p1.ReplayBuffer = agents.NewExperienceBuffer(cfg.DQN.ReplaySize)
			
			p2 := agents.NewDQNAgent(dqnOptions2, rlKnowledge)
			p2.SetBatchSize(cfg.DQN.BatchSize)
			p2.SetUpdateFrequency(cfg.DQN.UpdateFrequency)
			p2.ReplayBuffer = agents.NewExperienceBuffer(cfg.DQN.ReplaySize)
			
			agentMap[1] = p1
			agentMap[2] = p2
			
		default: // "rl" is the default
			// Setup RL agents for training
			p1 := agents.NewRLAgent(1, X, cfg.Game.M, cfg.Game.N, cfg.Game.K, board, rlKnowledge, true)
			p1.LearningRate = 0.2
			p1.DiscountFactor = 0.8
			p1.ExplorationFactor = 0.25
			
			p2 := agents.NewRLAgent(2, O, cfg.Game.M, cfg.Game.N, cfg.Game.K, board, rlKnowledge, true)
			p2.LearningRate = 0.2
			p2.DiscountFactor = 0.8
			p2.ExplorationFactor = 0.25

			agentMap[1] = p1
			agentMap[2] = p2
		}

		// Start training
		log := train(cfg, board, agentMap, display, rlKnowledge, &terminateFlag)
		display.ShowStats(log, agentMap, true, rlKnowledge.RandomDispersion)
		return
	}

	// Setup for normal play mode
	agentMap[1] = agents.NewHumanAgent(1, X, cfg.Game.M, cfg.Game.N)
	
	// Create agent based on selected type
	switch cfg.AgentType {
	case "dqn":
		// Create DQN agent options
		dqnOptions := common.AgentOptions{
			ID:               2,
			Sign:             O,
			IsLearner:        !cfg.NoLearn, // Use global NoLearn flag
			LearningRate:     0.01,
			DiscountFactor:   0.95,
			ExplorationFactor: 0.1,
			ModelFile:        cfg.DQN.ModelFile,
		}
		// Create DQN agent
		dqnAgent := agents.NewDQNAgent(dqnOptions, rlKnowledge)
		// Set batch size and update frequency from config
		dqnAgent.SetBatchSize(cfg.DQN.BatchSize)
		dqnAgent.SetUpdateFrequency(cfg.DQN.UpdateFrequency)
		// Initialize replay buffer with specified size
		dqnAgent.ReplayBuffer = agents.NewExperienceBuffer(cfg.DQN.ReplaySize)
		agentMap[2] = dqnAgent
	default: // "rl" is the default
		agentMap[2] = agents.NewRLAgent(2, O, cfg.Game.M, cfg.Game.N, cfg.Game.K, board, rlKnowledge, !cfg.NoLearn) // Use global NoLearn flag
	}

	// Ask for number of rounds
	fmt.Printf("? > How many rounds shall we play? ")
	_, err = fmt.Scanln(&cfg.Game.Rounds)
	if err != nil {
		fmt.Println("\n[error] Invalid input!")
		return
	}
	fmt.Println("Great! Have fun.")

	// Start the game
	log := play(cfg, board, agentMap, display, rlKnowledge)
	display.ShowStats(log, agentMap, false, nil)
}

// train runs the training process for the specified number of rounds
func train(cfg *config.Config, board common.Environment, agents map[int]common.Agent, display *ui.Display, knowledge *agents.RLAgentKnowledge, terminateFlag *bool) []int {
	log := make([]int, 3)
	fmt.Println("Commencing training...")

	// Verify model file is accessible and directory exists
	modelDir := filepath.Dir(cfg.RL.ModelFile)
	if modelDir != "." {
		if err := os.MkdirAll(modelDir, 0755); err != nil {
			fmt.Printf("Failed to create directory for model file: %v\n", err)
			return log
		}
	}
	
	file, err := os.OpenFile(cfg.RL.ModelFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Model file not accessible: %v\n", err)
		return log
	}
	defer file.Close()

	// Game state tracking
	var turn int = 1
	var progress int
	termW, _ := ui.GetTerminalSize()

	// Training loop
	for c := uint(1); c <= cfg.RL.TrainingMode; c++ {
		// Update progress bar on percentage changes or first iteration
		pTick := c*100%cfg.RL.TrainingMode == 0
		if pTick || c == 1 {
			// Refresh terminal width
			termW, _ = ui.GetTerminalSize()
			
			// Track progress percentage
			progress = int(c * 100 / cfg.RL.TrainingMode)
			
			// Display progress bar
			display.ClearPrompt()
			display.ShowProgressBar(progress, termW, "Training...", false)
		}

		// Start a new round and get the winner's ID
		prevTurn := turn
		turn = newRound(board, agents, display, turn, !cfg.Game.NoDisplay)
		log[turn]++ // Update score
		
		// If it was a draw, next player starts
		if turn == 0 {
			turn = getNextPlayer(prevTurn, len(agents)-1)
		}

		if !cfg.Game.NoDisplay {
			display.ShowSeparator()
		}

		// Check for termination signal
		if *terminateFlag {
			display.ShowProgressBar(progress, termW, "Terminated.", false)
			fmt.Println() // Add newline after progress bar
			
			// Save the model
			if !cfg.RL.NoLearn {
				knowledge.SaveToFile(cfg.RL.ModelFile)
			}
			return log
		}

		// Periodically save the model
		if !cfg.RL.NoLearn && pTick {
			knowledge.SaveToFile(cfg.RL.ModelFile)
		}
	}

	// Final progress bar
	display.ShowProgressBar(100, termW, "Training completed", true)
	fmt.Println() // Add newline after progress bar

	return log
}

// play starts a game between a human and the RL agent
func play(cfg *config.Config, board common.Environment, agents map[int]common.Agent, display *ui.Display, knowledge *agents.RLAgentKnowledge) []int {
	log := make([]int, 3)

	// Verify model file is accessible if learning is enabled
	if !cfg.RL.NoLearn {
		// Ensure directory exists
		modelDir := filepath.Dir(cfg.RL.ModelFile)
		if modelDir != "." {
			if err := os.MkdirAll(modelDir, 0755); err != nil {
				fmt.Printf("Failed to create directory for model file: %v\n", err)
				// Continue anyway, but warn user
			}
		}
		
		// Try to open the file
		file, err := os.OpenFile(cfg.RL.ModelFile, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			fmt.Printf("Model file not accessible: %v\nLearning will continue but progress won't be saved.\n", err)
		} else {
			defer file.Close()
		}
	}

	// Game loop
	turn := 1
	for c := 1; c <= cfg.Game.Rounds; c++ {
		// Start a new round and get the winner's ID
		prevTurn := turn
		turn = newRound(board, agents, display, turn, true)
		log[turn]++ // Update score
		
		// If it was a draw, next player starts
		if turn == 0 {
			turn = getNextPlayer(prevTurn, len(agents)-1)
		}

		display.ShowSeparator()

		// Save the model after each round if learning is enabled
		if !cfg.RL.NoLearn {
			// Save model directly using the knowledge instance
			knowledge.SaveToFile(cfg.RL.ModelFile)
		}
	}

	return log
}

// newRound starts a new game round
func newRound(board common.Environment, agents map[int]common.Agent, display *ui.Display, turn int, visual bool) int {
	// Reset the board
	board.Reset()
	
	// Reset display
	display.ResetFirstRun()
	
	// Draw the initial board
	if visual {
		display.ShowBoard(board.GetState())
	}

	// Set starting player if not specified
	if turn == 0 {
		turn = 1
	}

	// Game loop
	for {
		// Get current player's move
		possibleActions := board.GetPotentialActions(turn)
		
		// Validate we have available actions
		if len(possibleActions) == 0 {
			display.ClearPrompt()
			fmt.Printf("ERROR: No valid moves available for player %s (ID: %d)\n", agents[turn].GetSign(), turn)
			return 0 // Draw
		}
		
		// Get agent's move with better error context
		action, err := agents[turn].FetchMove(board.GetState(), possibleActions)
		if err != nil {
			display.ClearPrompt()
			fmt.Printf("Error getting move from agent %s (ID: %d): %v\n", agents[turn].GetSign(), turn, err)
			
			// For human agents, we'll retry. For AI agents, this is potentially a critical error
			// Check if this is a human agent by ID (ID 1 is human by convention)
			if turn != 1 { // Non-human agent
				fmt.Println("Critical AI error - ending game")
				return 0 // Force a draw to end the game
			}
			continue
		}

		// Execute the move
		_, err = board.Act(turn, action)
		if err != nil {
			// Show error and try again
			display.ClearPrompt()
			fmt.Print(err)
			continue
		}

		// Update display
		if visual {
			display.ShowMessages([]common.Agent{agents[1], agents[2]})
			display.ShowBoard(board.GetState())
		}

		// Check game state
		result := board.EvaluateAction(turn, action)

		// Clear prompt if game ended
		if visual && result != 0 {
			display.ClearPrompt()
		}

		// Handle game outcome
		if result == 0 {
			// Game continues
			turn = getNextPlayer(turn, len(agents)-1)
		} else if result == -1 {
			// Draw
			if visual {
				fmt.Println("It's a DRAW!")
			}

			// Notify agents of game end
			for id, agent := range agents {
				if id > 0 && agent != nil {
					agent.GameOver(board.GetState())
				}
			}
			
			return 0
		} else {
			// Current player won
			if visual {
				fmt.Printf("We have a WINNER! Congratulations %s\n", agents[turn].GetSign())
			}

			// Notify agents of game end
			for id, agent := range agents {
				if id > 0 && agent != nil {
					agent.GameOver(board.GetState())
				}
			}
			
			return turn
		}
	}
}

// getNextPlayer returns the ID of the next player
func getNextPlayer(current, maxPlayers int) int {
	if current < maxPlayers {
		return current + 1
	}
	return 1
}