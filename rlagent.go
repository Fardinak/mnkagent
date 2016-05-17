package main

import (
	"encoding/gob"
	"fmt"
	"math/rand"
	"os"
)

type RLAgent struct {
	// Agent PlayerID
	id int

	// View settings
	Sign string

	// Game definition
	m, n, k int

	// RL parameters
	Learning          bool
	LearningRate      float64 //alpha
	ExplorationFactor float64 //epsilon
	MaxReward         float64
	LossValue         float64

	// States stash
	values    map[string]float64
	prevState [][]int
	prevScore float64
	message   string
}

type RLAgentKnowledge struct {
	Values     map[string]float64
	Iterations uint
}

var rlKnowledge RLAgentKnowledge

func NewRLAgent(id int, sign string, m, n, k int, learn bool) (agent *RLAgent) {
	agent = new(RLAgent)
	agent.id = id
	agent.Sign = sign

	agent.m = m
	agent.n = n
	agent.k = k

	// Default values
	agent.Learning = learn
	agent.LearningRate = 0.2
	agent.ExplorationFactor = 0.25
	agent.MaxReward = 3
	agent.LossValue = -1

	// Initiate stash
	if rlKnowledge.Iterations == 0 {
		rlKnowledge.Values = make(map[string]float64)
	}
	agent.values = rlKnowledge.Values

	return
}

func (agent *RLAgent) FetchMessage() (message string) {
	message = agent.message
	agent.message = "-"
	return
}

func (agent *RLAgent) FetchMove(state [][]int) (pos int, err error) {
	var moveValue float64

	var e = rand.Float64()
	if e < agent.ExplorationFactor {
		agent.message = fmt.Sprintf("Exploratory action (%f)", e)

		// Choose a random move
		var emptyCells []int
		for i := range state {
			for j := range state[i] {
				if state[i][j] == 0 {
					emptyCells = append(emptyCells, i*agent.m+j+1)
				}
			}
		}

		pos = emptyCells[rand.Intn(len(emptyCells))]
		var i = (pos - 1) / agent.m
		var j = (pos - 1) % agent.m
		state[i][j] = agent.id
		moveValue = agent.lookup(state)
		state[i][j] = 0

	} else {
		agent.message = fmt.Sprintf("Greedy action (%f)", e)

		// Choose a greedy move
		var maxVal float64 = -10000
		var maxPos = 0

		for i := range state {
			for j := range state[i] {
				if state[i][j] == 0 {
					// REVIEW: Is tempState = copyState(state) too costly?
					state[i][j] = agent.id
					value := agent.lookup(state)
					state[i][j] = 0

					if value > maxVal {
						maxVal = value
						maxPos = i*agent.m + j + 1
					}
				}
			}
		}
		pos = maxPos
		moveValue = maxVal
	}

	agent.learn(moveValue)

	agent.prevState = copyState(state)
	agent.prevState[(pos-1)/agent.m][(pos-1)%agent.m] = agent.id
	agent.prevScore = moveValue

	return
}

func (agent *RLAgent) GameOver(state [][]int) {
	agent.learn(agent.value(state))
	rlKnowledge.Iterations++
}

func (agent *RLAgent) GetSign() string {
	return agent.Sign
}

// learn calculates new value for given state if agent is in learning mode
func (agent *RLAgent) learn(value float64) {
	if agent.Learning {
		agent.values[marshallState(agent.prevState, agent.id)] += agent.LearningRate *
			(value - agent.prevScore)
	}
}

// Return score for a certain state
func (agent *RLAgent) lookup(state [][]int) float64 {
	var mState = marshallState(state, agent.id) // Marshalled state
	val, ok := agent.values[mState]
	if !ok {
		val = agent.value(state)
		agent.values[mState] = val
	}
	return val
}

// value function returns given state's value
func (agent *RLAgent) value(state [][]int) float64 {
	switch evaluate(state) {
	case agent.id: // Agent won
		return 1
	case 0: // Game goes on
		return 0.5
	case -1: // Draw
		return 0
	default: // Agent lost
		return agent.LossValue
	}
}

// Generate and store all winning states and assign values
func (agent *RLAgent) enumerateStates(state [][]int, idx int, player int) {
	if idx >= agent.m*agent.n {
		// If last_to_act is agent add state
	} else if evaluate(state) > 0 {
		var i, j = idx / 3, idx % 3
		for val := 0; val < 3; val++ {
			state[i][j] = val
		}
	}
}

// storeKnowledge writes the knowledge map to given path
func (k *RLAgentKnowledge) saveToFile(path string) bool {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("[error] Could not open writable knowledge file on disk!")
		fmt.Println(err)
		return false
	}
	defer file.Close()

	enc := gob.NewEncoder(file)
	err = enc.Encode(rlKnowledge)
	if err != nil {
		fmt.Println("[error] Encoding of knowledge failed!")
		fmt.Println(err)
		return false
	}

	return true
}

// retrieveKnowledge reads the knowledge from given path to knowledge map
func (k *RLAgentKnowledge) loadFromFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("[error] Could not open readable knowledge file on disk!")
		fmt.Println(err)
		return false
	}
	defer file.Close()

	dec := gob.NewDecoder(file)
	err = dec.Decode(&rlKnowledge)
	if err != nil {
		fmt.Println("[error] Decoding of knowledge failed!")
		fmt.Println(err)
		return false
	}

	return true
}

func marshallState(state [][]int, agentID int) (m string) {
	for i := range state {
		for j := range state[i] {
			// Regulate based on current agent's ID so all agents are one!
			if state[i][j] > 0 {
				if state[i][j] == agentID {
					m += "X"
				} else {
					m += "O"
				}
			} else {
				m += "-"
			}
		}
	}
	return
}

func copyState(state [][]int) (c [][]int) {
	c = make([][]int, len(state))
	for i := range state {
		c[i] = make([]int, len(state[i]))
		copy(c[i], state[i])
	}
	return
}
