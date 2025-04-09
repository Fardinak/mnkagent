// Package agents provides player agents for the game
package agents

import (
	agentscommon "mnkagent/agents/common"
	"mnkagent/agents/dqn"
	"mnkagent/agents/human"
	"mnkagent/agents/rl"
	gamecommon "mnkagent/common"
)

// Export RL Knowledge type
var _ *rl.RLAgentKnowledge // for godoc

// NewRLAgent creates a new reinforcement learning agent
func NewRLAgent(id int, sign string, m, n, k int, environment gamecommon.Environment, knowledge *rl.RLAgentKnowledge, learn bool) *rl.RLAgent {
	return rl.NewRLAgent(id, sign, m, n, k, environment, knowledge, learn)
}

// NewDQNAgent creates a new Deep Q-Network agent
func NewDQNAgent(options gamecommon.AgentOptions, knowledge *rl.RLAgentKnowledge) *dqn.DQNAgent {
	return dqn.NewDQNAgent(options, knowledge)
}

// NewHumanAgent creates a new human-controlled agent
func NewHumanAgent(id int, sign string, m, n int) *human.HumanAgent {
	return human.NewHumanAgent(id, sign, m, n)
}

// NewExperienceBuffer creates a new experience replay buffer
func NewExperienceBuffer(capacity int) *agentscommon.ExperienceBuffer {
	return agentscommon.NewExperienceBuffer(capacity)
}

// NewNeuralNetwork creates a new neural network
func NewNeuralNetwork(inputSize, hiddenSize, outputSize int, learningRate float64) *agentscommon.NeuralNetwork {
	return agentscommon.NewNeuralNetwork(inputSize, hiddenSize, outputSize, learningRate)
}